# StepLauncher API

API a medida para **StepLauncher** (proyecto maestro): gestión de cuentas y servicios
del launcher. Rápida (edge), mantenible y segura por defecto.

- Runtime: **Cloudflare Workers** + **Hono**
- Datos y Auth: **Supabase** (Postgres + Auth + RLS)
- Versionado: `/v1`, `/v2`, … (sin romper versiones anteriores)
- Documentación: `docs/` (esencial y obligatoria)

## Inicio rápido

```bash
# 1. Instalar (siempre Bun, prohibido npm en este monorepo)
bun install

# 2. Variables locales (nunca commitear secretos)
copy .dev.vars.example .dev.vars
# ...rellenar .dev.vars...

# 3. Desarrollo
bun run dev

# 4. Verificación obligatoria antes de cada PR
bun run typecheck
```

Probar:

```bash
curl http://localhost:8787/
curl http://localhost:8787/v1/health
```

## Estructura

```text
api/
  src/
    index.ts            # Entrada: CORS, headers, request-id, router versionado
    env.ts              # Lectura y validación de variables (sin secretos en código)
    lib/                # Utilidades puras: supabase, respuestas JSON
    middleware/         # Capas HTTP: auth, rate-limit, headers, validación
    routes/v1/          # Endpoints v1 (health, accounts...). v2 será carpeta nueva.
    schemas/v1/         # Contratos Zod por versión (nunca se reutilizan entre versiones)
    types/              # Tipos compartidos (Bindings, AppEnv, Usuario)
  docs/                 # Documentación esencial (decisiones, seguridad, OAuth, cuentas, api/vX)
  db/v1/                # SQL versionado v1: partes + install.sql (un solo Run) + verify.sql
  scripts/              # build-db-bundle.ts: genera db/v1/install.sql (bun run db:build)
  wrangler.jsonc        # Solo config NO secreta. Secretos vía `wrangler secret put`.
  .dev.vars.example     # Plantilla local. El real `.dev.vars` está en .gitignore.
```

## Reglas (resumen)

1. **Versionado obligatorio**: todo endpoint cuelga de `/vN`. Nada fuera de versión salvo `/`.
2. **Seguridad en cuentas**: todo endpoint de cuenta exige `auth()` + validación Zod + RLS en Supabase.
3. **Sin secretos en código**: si ves un token en un `.ts`, es un bug crítico.
4. **Docs o no existe**: cada endpoint y cada decisión van a `docs/`.

Detalle completo para IAs y devs: leer **`api/AGENTS.md`** antes de tocar nada.
