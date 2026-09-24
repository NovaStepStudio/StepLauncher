# StepLauncher Yggdrasil

Servidor **Yggdrasil** a medida para StepLauncher: permite entrar a servidores
Minecraft Java en `online-mode` con la **cuenta de StepLauncher**, mediante
**authlib-injector** en cliente y servidor.

> Subproyecto separado del monorepo: **no** vive en `api/`. La API principal
> (`api/`) sigue intacta; este servicio solo **lee** su misma base de usuarios
> (Supabase Auth + `profiles`). Ver `docs/integracion-api.md`.

- Runtime: Cloudflare Workers + **Hono** + **Supabase** (`supabase-js`).
- Protocolo: Yggdrasil Mojang + extensiones authlib-injector (metadatos, ALI).
- Usuarios: los de StepLauncher (email confirmado + `profiles.username` +
  `profiles.mc_uuid`). Aquí **no** se registran cuentas.
- Tokens: Yggdrasil propios (`accessToken`/`clientToken`, 15 días, máx. 10 por
  usuario), guardados en `ygg_tokens`. Las sesiones de entrada (`join`) viven
  30 s en `ygg_joins`.
- Texturas: se leen de los buckets `skins`/`capes` de `api/` (la subida vive
  en la web), se sirven por `/textures/:hash` y se firman con la RSA propia.

## Qué implementa (MVP)

| Método | Ruta | Qué hace |
|---|---|---|
| `GET` | `/` | Metadatos authlib-injector (`meta`, `skinDomains`, `signaturePublickey`) + ALI |
| `POST` | `/authserver/authenticate` | Login con email/usuario + contraseña → tokens + perfil |
| `POST` | `/authserver/refresh` | Rota `accessToken` (y selecciona perfil si procede) |
| `POST` | `/authserver/validate` | `204` si el token vale, `403` si no |
| `POST` | `/authserver/invalidate` | Revoca un token (`204` siempre) |
| `POST` | `/authserver/signout` | Revoca todos los tokens del usuario (`204`) |
| `POST` | `/sessionserver/session/minecraft/join` | El cliente anuncia `serverId` antes de entrar |
| `GET` | `/sessionserver/session/minecraft/hasJoined` | El servidor verifica la sesión (`200` perfil o `204`) |
| `GET` | `/sessionserver/session/minecraft/profile/:uuid` | Perfil completo + `textures` firmadas |
| `GET` | `/textures/:hash` | PNG de skin/capa (SHA-256, `image/png`, caché inmutable) |
| `GET` | `/skins/MinecraftSkins/:usuario.png` | Respaldo legacy para juegos viejos |
| `GET` | `/skins/:nombre` · `/cloaks/:nombre` | Skin/capa por nombre (estilo Ely.by) |
| `GET` | `/signature-verification-key.pem` | Clave pública para verificar firmas |
| `GET` | `/sessionserver/blockedservers` | `[]` (lo pide el juego al abrir multijugador) |
| `POST` | `/api/profiles/minecraft` | Búsqueda por nombres (máx. 10 por llamada) |

Contrato exacto: `docs/contrato.md`. Base completa: `docs/base-yggdrasil.md`.
Uso del launcher: `docs/uso-launcher.md`. Montar un servidor: `docs/servidores.md`.

## Estructura

```text
yggdrasil/
  src/
    index.ts              # Tubería global + GET / + routers
    env.ts                # getEnv(): única vía a variables
    lib/crypto.ts         # UUID sin guiones, tokens aleatorios, firma RSA
    lib/ygg.ts            # Formas user/profile/textures del protocolo
    lib/textures.ts       # Skins/capas desde buckets de api/ + firma + hash
    lib/store.ts          # Supabase: usuarios, perfiles, tokens, joins
    middleware/           # request-id, security-headers, cors, rate-limit, validate
    routes/               # authserver, sessionserver, profiles, textures, meta
    schemas/ygg.ts        # Zod estricto (propio, no reutiliza api/)
    types/app.ts          # Bindings y entorno Hono
  docs/                   # base, integración, launcher, contrato, decisiones
  db/                     # SQL propio (ygg_tokens, ygg_joins)
  scripts/                # generar-claves.js (par RSA local, gitignored)
  wrangler.jsonc          # Sin secretos (solo vars no sensibles)
  .dev.vars.example       # Plantilla local (el real .dev.vars no se commitea)
```

## Desarrollo

```powershell
cd yggdrasil
bun install                       # bun obligatorio, prohibido npm
copy .dev.vars.example .dev.vars  # rellenar (SUPABASE_*, YGG_SIGN_PRIVATE_KEY…)
bun run gen:claves                # genera el par RSA en .keys/ (gitignored)
bun run gen:claves -- --write-dev-vars  # opcional: escribe la privada en .dev.vars
bun run dev                       # wrangler dev
bun run typecheck                 # tsc --noEmit (obligatorio antes de cada PR)
```

Claves: el script deja `ygg-private.pem`, `ygg-public.pem` y dos archivos de
**una sola línea** (el `.pem` multilínea se corta al pegarlo en prompts):
`YGG_SIGN_PRIVATE_KEY.devvars.txt` (para `.dev.vars`) y
`YGG_SIGN_PRIVATE_KEY.secret.txt` (solo el valor, para producción).
A producción se sube con
`Get-Content '.keys\YGG_SIGN_PRIVATE_KEY.secret.txt' -Raw | wrangler secret put YGG_SIGN_PRIVATE_KEY`.
Si ya tenés el `.pem` pero te faltan los `.txt`: `bun run gen:claves -- --reempaquetar`.
La pública no se configura: el Worker la deriva solo y la expone en `GET /`.

Probar:

```bash
curl http://localhost:8787/
curl -X POST http://localhost:8787/authserver/authenticate \
  -H 'Content-Type: application/json' \
  -d '{"username":"Steve_01","password":"secreta123"}'
```

## Variables y secretos

| Qué | Dónde | Ejemplo |
|---|---|---|
| Config no secreta | `vars` en `wrangler.jsonc` | `SERVER_NAME`, `SKIN_DOMAINS` |
| Secreto en producción | `wrangler secret put` | `SUPABASE_SERVICE_ROLE_KEY`, `YGG_SIGN_PRIVATE_KEY` |
| Secreto en local | `.dev.vars` (gitignored) | Los mismos de arriba |

La clave RSA firma `textures` (`SHA1withRSA`); la pública se expone en `GET /`
como `signaturePublickey`. Si se rota, los clientes re-descargan metadatos.

> Nota git: la raíz ignora `*.jsonc` salvo `api/` y `website/`. Para versionar
> `yggdrasil/wrangler.jsonc` hay que añadir `!yggdrasil/wrangler.jsonc` al
> `.gitignore` raíz (cambio cruzado: pedir permiso explícito al owner).
