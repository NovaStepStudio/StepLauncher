// Genera el par de claves RSA del servicio Yggdrasil y lo deja listo
// para `wrangler secret put` y para el `.dev.vars` local.
//
// Uso (dentro de `yggdrasil/`, con Bun):
//   bun run gen:claves                  # genera en `.keys/` (no pisa sin --force)
//   bun run gen:claves -- --force       # regenera aunque ya existan
//   bun run gen:claves -- --write-dev-vars  # además escribe la privada en `.dev.vars`
//   bun run gen:claves -- --reempaquetar     # no genera: rearma los .txt desde el .pem que ya tenés
//
// Lo que genera (todo queda en `.keys/`, gitignored, nunca se commitea):
//   ygg-private.pem                    # privada PKCS8 (multilínea, NO pegarla en prompts)
//   ygg-public.pem                     # pública SPKI → solo referencia (el Worker la deriva solo)
//   YGG_SIGN_PRIVATE_KEY.devvars.txt   # `YGG_SIGN_PRIVATE_KEY="..."` en UNA línea, lista para `.dev.vars`
//   YGG_SIGN_PRIVATE_KEY.secret.txt    # valor solo en UNA línea, listo para `wrangler secret put`
//
// Por qué en una línea: los prompts multilínea cortan el pegado y el secreto
// llega incompleto. El Worker reconvierte `\n` a saltos reales al leer
// (ver `src/lib/crypto.ts` y `src/routes/meta.ts`).
//
// La privada NUNCA se imprime por consola: solo rutas y huella SHA-256.

import { generateKeyPairSync, createHash } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, writeFileSync, chmodSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

// Raíz del subproyecto (este script vive en `yggdrasil/scripts/`).
const DIR_RAIZ = dirname(dirname(fileURLToPath(import.meta.url)));

function ayuda() {
  console.log(`Generador de claves RSA para StepLauncher Yggdrasil.
Uso: bun run gen:claves [--opciones]

Opciones:
  --out-dir <carpeta>   Destino de los .pem (defecto: .keys)
  --bits <2048|4096>    Tamaño de la clave (defecto: 2048, mínimo: 2048)
  --force               Regenera aunque ya existan archivos
  --write-dev-vars      Escribe la privada en .dev.vars (lo crea desde el ejemplo si falta)
  --reempaquetar        No genera: rearma los .txt de una línea desde el .pem existente
  --help                Muestra esta ayuda

Después de generar:
  1. Local: completa .dev.vars (o usa --write-dev-vars) y corre \`bun run dev\`.
  2. Producción: sube el secreto con wrangler (ver comandos al final).
  3. Verifica con: curl http://localhost:8787/  (debe traer signaturePublickey)
  4. Borra .keys/ cuando ya no lo necesites.`);
}

function leerArgs(argv) {
  const o = {
    outDir: ".keys",
    bits: 2048,
    force: false,
    writeDevVars: false,
    reempaquetar: false,
    help: false,
  };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === "--help" || a === "-h") o.help = true;
    else if (a === "--force") o.force = true;
    else if (a === "--write-dev-vars") o.writeDevVars = true;
    else if (a === "--reempaquetar") o.reempaquetar = true;
    else if (a === "--out-dir") o.outDir = argv[++i] ?? o.outDir;
    else if (a === "--bits") o.bits = Number(argv[++i] ?? o.bits);
    else {
      console.error(`Opción desconocida: ${a}\n`);
      ayuda();
      process.exit(2);
    }
  }
  return o;
}

/** Inserta o reemplaza YGG_SIGN_PRIVATE_KEY en el contenido de `.dev.vars`. */
function ponerEnDevVars(contenido, lineaEscapada) {
  const linea = `YGG_SIGN_PRIVATE_KEY="${lineaEscapada}"`;
  const re = /^YGG_SIGN_PRIVATE_KEY=.*$/m;
  if (re.test(contenido)) return contenido.replace(re, linea);
  const base = contenido.endsWith("\n") ? contenido : contenido + "\n";
  return base + linea + "\n";
}

/**
 * Escribe los dos archivos de una línea desde la privada en PEM:
 * - `YGG_SIGN_PRIVATE_KEY.devvars.txt`: `YGG_SIGN_PRIVATE_KEY="..."` (para `.dev.vars`)
 * - `YGG_SIGN_PRIVATE_KEY.secret.txt`: solo el valor (para `wrangler secret put`)
 * Devuelve la versión escapada por si hay que escribirla en `.dev.vars`.
 */
function empaquetarUnaLinea(destino, privateKey) {
  const escapada = privateKey.replace(/\r?\n/g, "\\n").replace(/\\n$/, "");
  const rutaDevVars = join(destino, "YGG_SIGN_PRIVATE_KEY.devvars.txt");
  const rutaSecreto = join(destino, "YGG_SIGN_PRIVATE_KEY.secret.txt");
  writeFileSync(rutaDevVars, `YGG_SIGN_PRIVATE_KEY="${escapada}"\n`, "utf8");
  writeFileSync(rutaSecreto, escapada, "utf8");
  return escapada;
}

/** Muestra los pasos finales (PowerShell) sin exponer la privada. */
function mostrarPasos(args, destino, rutaPublica, yaEnDevVars) {
  console.log(`Listo en ${destino} (todo en UNA línea, sin renglones):
  devvars : ${join(destino, "YGG_SIGN_PRIVATE_KEY.devvars.txt")} (para .dev.vars)
  secreto : ${join(destino, "YGG_SIGN_PRIVATE_KEY.secret.txt")} (para wrangler)

Siguientes pasos (PowerShell, dentro de yggdrasil/):
  1. Local:  ${yaEnDevVars ? "ya hecho con --write-dev-vars." : "pega el contenido de YGG_SIGN_PRIVATE_KEY.devvars.txt en tu .dev.vars"}
  2. Produc: Get-Content '${args.outDir}\\YGG_SIGN_PRIVATE_KEY.secret.txt' -Raw | wrangler secret put YGG_SIGN_PRIVATE_KEY
     (una sola línea: ya no se corta el pegado)
  3. Verificá: bun run dev  y luego  curl http://localhost:8787/
     (el campo signaturePublickey debe coincidir con ${rutaPublica})
  4. Cuando termines, borra la carpeta ${args.outDir}.`);
}

function main() {
  const args = leerArgs(process.argv.slice(2));
  if (args.help) {
    ayuda();
    return;
  }

  const destino = resolve(DIR_RAIZ, args.outDir);
  const rutaPrivada = join(destino, "ygg-private.pem");
  const rutaPublica = join(destino, "ygg-public.pem");

  // Reempaquetar: no genera nada, solo rearma los .txt desde el .pem que ya tenés.
  if (args.reempaquetar) {
    if (!existsSync(rutaPrivada)) {
      console.error(`No hay ${rutaPrivada}: primero genera con \`bun run gen:claves\`.`);
      process.exit(1);
    }
    const privateKey = readFileSync(rutaPrivada, "utf8");
    if (!privateKey.includes("BEGIN PRIVATE KEY")) {
      console.error("El archivo no parece una privada PKCS8 válida.");
      process.exit(1);
    }
    const escapada = empaquetarUnaLinea(destino, privateKey);
    if (args.writeDevVars) {
      const rutaEjemplo = resolve(DIR_RAIZ, ".dev.vars.example");
      const rutaLocal = resolve(DIR_RAIZ, ".dev.vars");
      let base = "";
      if (existsSync(rutaLocal)) base = readFileSync(rutaLocal, "utf8");
      else if (existsSync(rutaEjemplo)) base = readFileSync(rutaEjemplo, "utf8");
      writeFileSync(rutaLocal, ponerEnDevVars(base, escapada), "utf8");
      console.log("`.dev.vars` actualizado con la YGG_SIGN_PRIVATE_KEY de una línea.");
    }
    let huella = "(sin pública para comparar)";
    if (existsSync(rutaPublica)) {
      huella = createHash("sha256").update(readFileSync(rutaPublica, "utf8")).digest("hex");
    }
    console.log(`Reempaquetado OK. Huella SHA-256 de la pública: ${huella}\n`);
    mostrarPasos(args, destino, rutaPublica, args.writeDevVars);
    return;
  }

  if (!Number.isInteger(args.bits) || args.bits < 2048) {
    console.error("Bits inválidos: usa 2048 o más (recomendado 2048).");
    process.exit(2);
  }

  if (!args.force && (existsSync(rutaPrivada) || existsSync(rutaPublica))) {
    console.error(
      `Ya hay claves en ${destino}. Usa --force para regenerar (ojo: rota la firma y los clientes deberán re-descargar metadatos), o --reempaquetar para rearmar los .txt sin cambiar la clave.`,
    );
    process.exit(1);
  }

  const { privateKey, publicKey } = generateKeyPairSync("rsa", {
    modulusLength: args.bits,
    publicKeyEncoding: { type: "spki", format: "pem" },
    privateKeyEncoding: { type: "pkcs8", format: "pem" },
  });

  mkdirSync(destino, { recursive: true });
  writeFileSync(rutaPrivada, privateKey, "utf8");
  writeFileSync(rutaPublica, publicKey, "utf8");
  try {
    chmodSync(rutaPrivada, 0o600);
  } catch {
    // En Windows no hay chmod POSIX: el resguardo real es el .gitignore.
  }

  const escapada = empaquetarUnaLinea(destino, privateKey);

  if (args.writeDevVars) {
    const rutaEjemplo = resolve(DIR_RAIZ, ".dev.vars.example");
    const rutaLocal = resolve(DIR_RAIZ, ".dev.vars");
    let base = "";
    if (existsSync(rutaLocal)) base = readFileSync(rutaLocal, "utf8");
    else if (existsSync(rutaEjemplo)) base = readFileSync(rutaEjemplo, "utf8");
    writeFileSync(rutaLocal, ponerEnDevVars(base, escapada), "utf8");
    console.log("`.dev.vars` actualizado con la nueva YGG_SIGN_PRIVATE_KEY.");
  }

  const huella = createHash("sha256").update(publicKey).digest("hex");
  console.log(`Claves generadas en ${destino}:
  privada : ${rutaPrivada} (no se muestra por seguridad)
  pública : ${rutaPublica}
  huella SHA-256 de la pública: ${huella}
`);
  mostrarPasos(args, destino, rutaPublica, args.writeDevVars);
}

main();
