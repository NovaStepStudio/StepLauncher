# Guía para IAs — `api/` (StepLauncher API)

> API a medida para **StepLauncher** (proyecto maestro): todo gira alrededor de la
> cuenta del jugador. Runtime **Cloudflare Workers** + **Hono**, datos y Auth en
> **Supabase**. Toda comunicación y comentarios de código en **español**.

## 0. Alcance (aislamiento del monorepo)

- Un cambio aquí **solo** toca `api/`. Prohibido modificar `launcher/`, `website/`
  o la raíz "de paso". Leer fuera para contexto sí, modificar fuera no.
- Gestor **siempre Bun** (`bun install`, `bun run dev`). **Prohibido npm**.
- Todo se ejecuta dentro de `api/` (aquí no hay `go build`; eso es de `launcher/`).

## 1. Código mantenible que no rompa otros endpoints

- Un dominio = un archivo en `src/routes/vN/` + esquemas en `src/schemas/vN/`
  + una línea `route()` en `src/routes/vN/index.ts`. Sin lógica en `index.ts`.
- Respuestas solo con `ok()` / `fail()` de `src/lib/respond.ts` (sobre único).
- Variables solo vía `getEnv()` (`src/env.ts`); prohibido `c.env` directo en rutas.
- Prohibido compartir esquemas entre versiones: v2 **copia** de v1, nunca reutiliza.
- Antes de cada PR: `bun run typecheck`. Si rompe el contrato de otro endpoint,
  el cambio va a versión nueva (ver §3), no se "ajusta" la vieja.

## 2. Seguridad mínima en endpoints de cuenta (obligatoria, no negociable)

Todo endpoint que lea o modifique datos del usuario debe cumplir, por orden:

1. `rateLimit(preset)` — `account` por defecto, `sensitive` en login/registro/recuperación.
2. `requireAuth()` — el dueño sale de `c.get("user").id` (Bearer validado en Supabase).
   **Prohibido** aceptar `userId`/`email` del cliente como identidad.
3. `validateJson` / `validateQuery` con Zod (`.strict()` donde aplique).
4. Acceso a datos con `supabaseForUser()` (RLS activo). `supabaseAdmin()` solo si es
   imprescindible, documentando el porqué, y jamás devolviendo su resultado crudo con
   campos sensibles.
5. Errores genéricos (`unauthorized`, `not_found`, `rate_limited`) sin stack/SQL/secretos,
   con `requestId`. 401 sin distinguir expirado de inválido.
6. Tablas de usuario con RLS `auth.uid() = user_id` (USING + WITH CHECK) y prueba
   cruzada A-vs-B antes de mergear.

Checklist completa: `docs/arquitectura/seguridad.md`. Si un punto falla, no se mergea.

## 3. Versionado v1, v2, v3… (nunca romper)

- Toda ruta pública cuelga de `/vN`. Solo `/` (índice) vive fuera de versión.
- Compatible (campo opcional nuevo, endpoint nuevo) → puede entrar en la misma versión.
- Incompatible (quitar/renombrar, cambiar semántica, exigir auth nueva) → **versión nueva**
  copiando la carpeta anterior (`src/routes/`, `src/schemas/`, `docs/api/`).
- Deprecar = marcar `Deprecated` en `docs/api/vX/` + ADR + aviso; mantener hasta que el
  launcher viejo deje de usarlo. Detalle: `docs/arquitectura/versionado.md`.

## 4. Documentación en `docs/` (esencial, sin excepciones)

`docs/` es la memoria del proyecto: un dev debe entender decisiones y contratos sin
preguntar. Si no está en `docs/`, no existe.

```text
docs/arquitectura/  → porqués: decisiones.md (ADRs), estructura.md, versionado.md,
                      seguridad.md, secretos-y-entorno.md
docs/oauth/          → TODO lo de OAuth: flujos, tokens, PKCE, sesiones, launcher desktop
docs/cuentas/        → TODO lo de cuentas: modelo, perfil, tabla profiles
docs/api/v1/...     → contrato HTTP exacto v1 (health.md, cuentas.md). v2 será carpeta nueva.
```

Reglas:

- Endpoint nuevo/changed → actualizar `docs/api/vX/<dominio>.md` **en el mismo cambio**.
- Decisión (librería, tabla, flujo, secreto) → ADR en `docs/arquitectura/decisiones.md`.
- Duda de ubicación: OAuth → `docs/oauth/`; cuentas → `docs/cuentas/`; cómo versionar
  o dónde poner un archivo → `docs/arquitectura/`; método/ruta/códigos → `docs/api/vX/`.
- Índice y obligación de documentar: `docs/README.md`.

## 5. Código abierto: disciplina de secretos antes de commit/pull

- **Nada incrustado en código**: ni tokens, ni `service_role`, ni URLs con credenciales.
  Config no secreta → `vars` en `wrangler.jsonc`. Secreto → `wrangler secret put`
  (producción) / `.dev.vars` local (gitignored). Plantilla sin valores → `.dev.vars.example`.
- Revisión obligatoria antes de cada `commit`/`pull`:
  `git diff --cached` + buscar `service_role|secret|token\s*=\s*['"]` en `src` y
  `wrangler.jsonc`. Si aparece un valor real: parar, mover a secreto y **rotarlo**
  (borrar del código no basta, queda en historial).
- Decisión seria = documentada (ADR) + typecheck verde + docs actualizadas. Sin eso, no hay merge.

Detalle: `docs/arquitectura/secretos-y-entorno.md`.

## 6. Verificación obligatoria

```bash
# Dentro de api/
bun install        # solo si faltan deps
bun run typecheck  # tsc --noEmit, obligatorio antes de cada PR
bun run dev        # wrangler dev (probar / y /v1/health)
```

## 7. Añadir un endpoint (chuleta)

1. Esquema Zod en `src/schemas/v1/<dominio>.ts` (si hay entrada).
2. Ruta en `src/routes/v1/<dominio>.ts`: `rateLimit` → `requireAuth` (si es cuenta)
   → `validate*` → handler con `ok`/`fail`.
3. Montar en `src/routes/v1/index.ts` si es dominio nuevo.
4. Docs en `docs/api/v1/<dominio>.md` (+ `docs/cuentas/` u `docs/oauth/` si aplica).
5. `bun run typecheck` + revisión de secretos (§5).
