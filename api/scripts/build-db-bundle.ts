// Genera db/v1/install.sql a partir de las partes 01–07.
// Fuente de verdad: los archivos db/v1/0*.sql. El bundle resultante es lo que
// se pega UNA vez en Supabase (SQL Editor). No editar install.sql a mano:
// se sobrescribe en cada ejecución.
//
// Uso (dentro de api/): bun run db:build

import { readdir, readFile, writeFile } from "node:fs/promises";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const v1Dir = join(here, "..", "db", "v1");
const outFile = join(v1Dir, "install.sql");

const header = `-- ============================================================================
-- StepLauncher API — Paquete v1 (instalador único, GENERADO, no editar a mano).
-- Generado con: bun run db:build  (fuente: db/v1/01_*.sql … 07_*.sql)
--
-- Uso: pegar este archivo COMPLETO en Supabase → SQL Editor → Run.
-- Sirve para instalación nueva Y para BD existentes (todo es idempotente:
-- IF NOT EXISTS / OR REPLACE / DROP IF EXISTS / ON CONFLICT DO NOTHING).
-- Se aplica en UNA transacción: o entra todo o no entra nada.
-- Después, ejecutar db/v1/verify.sql para comprobar el resultado.
-- v2 será db/v2/ con su propio install.sql (v1 nunca se reescribe a mano).
-- ============================================================================

begin;

-- Ventana suficiente para instalaciones grandes sin bloquear en exceso.
set local statement_timeout = '120s';
set local lock_timeout = '30s';

-- gen_random_uuid() de file_uploads/notifications/cosmetics lo exige.
create extension if not exists "pgcrypto";

-- Registro de versión aplicada (trazabilidad: qué paquete vio esta BD).
create table if not exists public.schema_version (
  version text primary key,
  applied_at timestamptz not null default now()
);
`;

const footer = `
-- Marca v1 como aplicada (reejecutable: no duplica).
insert into public.schema_version (version) values ('v1')
on conflict (version) do nothing;

commit;
`;

const files = (await readdir(v1Dir))
  .filter((f) => /^[0-9]{2}_.*\.sql$/.test(f))
  .sort();

if (files.length === 0) {
  console.error("Sin partes 01–07 en db/v1. Nada que generar.");
  process.exit(1);
}

let bundle = header;
for (const f of files) {
  const body = await readFile(join(v1Dir, f), "utf8");
  bundle += `\n-- ----------------------------------------------------------------------------\n`;
  bundle += `-- ${f}\n`;
  bundle += `-- ----------------------------------------------------------------------------\n`;
  bundle += body.trim() + "\n";
}
bundle += footer;

await writeFile(outFile, bundle, "utf8");
console.log(`Paquete v1 generado: db/v1/install.sql (${files.length} partes).`);
