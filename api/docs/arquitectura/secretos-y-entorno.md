# Secretos y entorno

**Regla de oro (código abierto): ningún secreto va en el repo.** Ni tokens, ni
`service_role`, ni URLs con credenciales. Si ves un valor real en un `.ts`, es un
bug crítico: parar, moverlo a secreto y **rotarlo** (borrarlo del código no basta,
queda en historial).

## Dónde vive cada cosa

| Tipo | Dónde | Ejemplo |
|---|---|---|
| Config NO secreta | `vars` en `wrangler.jsonc` | `API_ENV`, `ALLOWED_ORIGINS` |
| Secreto en producción | `wrangler secret put <NOMBRE>` | `SUPABASE_SERVICE_ROLE_KEY` |
| Secreto en local | `.dev.vars` (gitignored, nunca commiteado) | los tres de Supabase |
| Plantilla sin valores | `.dev.vars.example` | copiar a `.dev.vars` y rellenar |

## Variables (ver `src/types/app.ts` y `src/env.ts`)

| Variable | Secreta | Qué es |
|---|---|---|
| `SUPABASE_URL` | No (pero va en variable por entorno) | URL del proyecto Supabase |
| `SUPABASE_ANON_KEY` | Pública limitada por RLS | Valida sesiones de usuario |
| `SUPABASE_SERVICE_ROLE_KEY` | **Sí, ultra** | Bypass RLS; solo servidor, jamás al cliente ni a logs |
| `API_ENV` | No | `local` \| `preview` \| `production` (defecto `production`; HSTS solo en producción) |
| `ALLOWED_ORIGINS` | No | Orígenes CORS separados por coma, sin wildcard (ej: `https://steplauncher.pages.dev,http://localhost:5173`) |

`getEnv()` falla con `server_misconfigured` (500 genérico, sin decir qué falta:
detallarlo sería enumeración para un atacante) si falta alguna de Supabase.

## Revisión obligatoria antes de cada `commit`/`pull`

```bash
git diff --cached
# Buscar en src y wrangler.jsonc:
# service_role|secret|token\s*=\s*['"]
```

- Decisión seria = ADR + `bun run typecheck` verde + docs actualizadas. Sin eso, no hay merge.
- El SQL tampoco lleva secretos ni datos reales (ver `db/v1/README.md`).
