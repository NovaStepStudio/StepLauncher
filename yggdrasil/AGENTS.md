# Guía para IAs — `yggdrasil/` (StepLauncher Yggdrasil)

> Servidor Yggdrasil a medida para **StepLauncher**: permite jugar Minecraft Java
> con la cuenta de StepLauncher en servidores `online-mode` mediante
> authlib-injector. Runtime **Cloudflare Workers** + **Hono**, usuarios en
> **Supabase** (la misma base que `api/`). Toda comunicación y comentarios en
> **español**.

## 0. Alcance (aislamiento del monorepo)

- Un cambio aquí **solo** toca `yggdrasil/`. Prohibido modificar `api/`,
  `launcher/`, `website/` o la raíz "de paso". Leer fuera para contexto sí,
  modificar fuera no.
- **Sin código compartido**: `yggdrasil/` nunca importa de `api/` ni de
  `website/`/`launcher/`. Si hay que reutilizar una idea, se copia y se adapta.
- Gestor **siempre Bun** (`bun install`, `bun run dev`). **Prohibido npm**.
- Todo se ejecuta dentro de `yggdrasil/`.

## 1. Compatibilidad Yggdrasil (no negociable)

- El protocolo manda: formatos exactos de authlib-injector / Mojang Yggdrasil
  (ver `docs/base-yggdrasil.md`). Nada de sobre `ok()`/`fail()` de `api/`.
- Errores con forma Yggdrasil: `{ "error": "...", "errorMessage": "...",
  "cause?": "..." }` y códigos HTTP del spec (403 token inválido, etc.).
- Rutas separadas por dominio Mojang, montadas en la raíz del Worker:
  `/authserver/*`, `/sessionserver/*`, `/api/profiles/*`, más `GET /` (metadatos).
- `GET /` devuelve `{ meta, skinDomains, signaturePublickey }` y cabecera
  `X-Authlib-Injector-API-Location` para descubrimiento (ALI).

## 2. De dónde salen los usuarios (fuente única)

- Los usuarios **no** se crean aquí: viven en **Supabase Auth + `profiles`**
  (la misma base que usa `api/`). Ver `docs/integracion-api.md`.
- `authenticate`/`signout` verifican email+contraseña contra Supabase
  (`signInWithPassword` en servidor, jamás en cliente). Solo cuentas con
  email confirmado entran al juego.
- El perfil Minecraft sale de `profiles`: `username` = nombre de jugador,
  `mc_uuid` = UUID del perfil. Sin `profiles` no hay `selectedProfile`.
- Los tokens Yggdrasil (`accessToken`/`clientToken`) son propios de este
  servicio (tabla `ygg_tokens`), con expiración de 15 días y límite de 10
  por usuario. No son JWT de Supabase y nunca se confunden con ellos.

## 3. Seguridad mínima

1. `rateLimit(preset)` — `sensitive` en `authenticate`/`signout`,
   `account` en el resto, `join` con ventana corta (6/30s por cuenta).
2. `validateJson` / `validateQuery` con Zod (`.strict()` donde aplique).
3. Supabase solo con `supabaseAdmin()` en servidor para Auth y lectura de
   `profiles`; documentar el porqué en cada uso. Jamás devolver filas crudas.
4. Secretos solo vía `getEnv()` (`src/env.ts`): `SUPABASE_*`,
   `YGG_SIGN_PRIVATE_KEY` (PEM RSA para firmar `textures`). Prohibido
   `c.env` directo en rutas y prohibido hardcodear claves.
5. Texturas: solo URLs del dominio propio + `skinDomains`; `Content-Type`
   `image/png`; firma `SHA1withRSA` verificable con `signaturePublickey`.
6. Errores genéricos al cliente en login (no enumerar usuarios), con
   `Retry-After` en 429.

## 4. Estructura (un dominio = un archivo)

```text
yggdrasil/
  src/
    index.ts              # Tubería global + GET / + montaje de routers
    env.ts                # getEnv(): única vía a variables
    lib/crypto.ts         # UUID sin guiones, tokens, firma RSA
    lib/ygg.ts            # Serialización user/profile/textures
    lib/textures.ts       # Skins/capas desde buckets de api/ + firma + hash
    lib/store.ts          # Acceso a Supabase (tokens, joins, perfiles)
    middleware/           # request-id, security-headers, cors, rate-limit, validate
    routes/               # authserver, sessionserver, profiles, textures, legacy, meta
    schemas/ygg.ts        # Zod estricto por endpoint (no se comparte con api/)
    types/app.ts          # Bindings, Variables, AppEnv (solo tipos)
  docs/                   # Memoria obligatoria (si no está aquí, no existe)
  db/                     # SQL propio: ygg_tokens, ygg_joins, ygg_textures + README
```

## 5. Documentación en `docs/` (obligatoria)

- `base-yggdrasil.md` — qué es Yggdrasil al completo (endpoints, flujos, firmas).
- `integracion-api.md` — cómo usa la API real / Supabase como fuente de usuarios.
- `uso-launcher.md` — cómo el launcher debe consumir este servicio.
- `contrato.md` — rutas, cuerpos, respuestas y errores exactos implementados.
- Decisión (tabla, clave, límite, flujo) → ADR en `docs/decisiones.md`.

## 6. Verificación obligatoria

```bash
# Dentro de yggdrasil/
bun install        # solo si faltan deps
bun run typecheck  # tsc --noEmit, obligatorio antes de cada PR
bun run dev        # wrangler dev (probar GET / y POST /authserver/authenticate)
```

Revisión de secretos antes de cada `commit`/`pull`: `git diff --cached` +
buscar `PRIVATE KEY|service_role|secret|token\s*=\s*['"]` en `src` y
`wrangler.jsonc`. Si aparece un valor real: parar, mover a secreto y rotarlo.
