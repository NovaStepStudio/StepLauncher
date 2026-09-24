<div align="center">

<a href="https://steplauncher.pages.dev">
  <img src="../branding/banner-002.png" alt="StepLauncher — Banner oficial | Powered by NovaCore Engine" width="100%">
</a>

<img src="../branding/appicon-neon.png" alt="StepLauncher" width="110">

# StepLauncher API

**Backend a medida de StepLauncher — cuenta del jugador, comunidad y servicios del launcher, creado por NovaStepStudio**

Cloudflare Workers + **Hono** + **Supabase** &nbsp;·&nbsp; rápida (edge), versionada y segura por defecto

> 🌐 **La consumen:** la app de escritorio ([`launcher/`](../launcher/)) y la web pública ([`website/`](../website/) en [steplauncher.pages.dev](https://steplauncher.pages.dev)) — registro, sesiones, perfil, cosméticos, amigos y notificaciones.

[![API v1](https://img.shields.io/badge/API-v1-31b3ff?style=for-the-badge&logo=cloudflare&logoColor=white)](docs/api/v1/README.md)
[![Hono](https://img.shields.io/badge/Hono-4.6-E36002?style=for-the-badge&logo=hono&logoColor=white)](https://hono.dev)
[![Supabase](https://img.shields.io/badge/Supabase-Postgres+Auth-3FCF8E?style=for-the-badge&logo=supabase&logoColor=white)](https://supabase.com)
[![Zod](https://img.shields.io/badge/Zod-3.23-3068B7?style=for-the-badge&logo=zod&logoColor=white)](https://zod.dev)
[![Workers](https://img.shields.io/badge/Workers-Edge-F68230?style=for-the-badge&logo=cloudflareworkers&logoColor=white)](https://workers.cloudflare.com)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![bun](https://img.shields.io/badge/bun-f9f1e1?style=for-the-badge&logo=bun&logoColor=black)](https://bun.sh)
[![Licencia](https://img.shields.io/badge/Licencia-GPL--3.0-a42e2e?style=for-the-badge&label=Licencia)](https://github.com/NovaStepStudio/StepLauncher/blob/main/LICENSE.md)

</div>

---

## 📖 Sobre

La API es el centro de la **cuenta del jugador**: identidad (Supabase Auth), perfil, presencia, avatar y banner, skins y capas, cosméticos, amistades y bloqueos, notificaciones y búsqueda de jugadores. Todo endpoint público cuelga de `/vN` (`/v1` hoy) y **nunca se rompe**: lo incompatible va a una versión nueva copiando la anterior.

Reglas IA del subproyecto: [`AGENTS.md`](AGENTS.md). Memoria obligatoria: [`docs/`](docs/) (si no está en `docs/`, no existe).

## ✨ Qué incluye (dominios v1)

| Dominio | Rutas | Qué hace |
|---|---|---|
| 🩺 Salud | `GET /v1/health` | Sonda de salud y versión (rate-limit `public`) |
| 🔑 Auth | `POST /v1/auth/register\|login\|refresh\|logout\|resend\|recover\|confirm\|reset-password\|recovery-password` | Registro con confirmación por correo, login, renovación y cierre de sesión, reenvío y recupero (rate-limit `sensitive`) |
| 👤 Cuenta | `GET\|PATCH /v1/accounts/me` + presencia, email, contraseña, avatar, banner, cosméticos, privacidad | Perfil propio, `is_online`, cambio de email/contraseña, subida de avatar (2 MB) y banner (8 MB, máx. 1920×1080), equipar cosméticos |
| 📁 Archivos | `POST /v1/files/skin\|cape`, `GET /v1/files`, `GET /v1/files/:id/url`, `DELETE /v1/files/:id` | Skins y capas privadas (1 MB, firma mágica verificada, URLs firmadas de 1 h, cuota de 10/día) |
| 🔔 Notificaciones | `GET /v1/notifications`, `PATCH …/:id/read`, `POST …/read-all`, `DELETE …/:id` | Bandeja propia paginada (`?limit=` defecto 20, máx. 50) |
| 🤝 Amigos | `/v1/friends` (solicitudes, amistades, `blocks`) | Enviar/aceptar/rechazar/retirar solicitudes, romper amistad, lista negra |
| 🔎 Usuarios | `GET /v1/users/search` | Búsqueda pública de jugadores (máx. 20) |
| 🌍 Comunidad | `GET /v1/community/profiles/:identifier` | Previsualización pública del perfil + relación (solo cosméticos equipados) |

Contrato HTTP exacto por dominio: [`docs/api/v1/`](docs/api/v1/README.md) (`health.md`, `auth.md`, `cuentas.md`, `archivos.md`, `notificaciones.md`, `amigos.md`, `usuarios.md`, `comunidad.md`).

## 🔐 Seguridad (no negociable en endpoints de cuenta)

Todo endpoint que lea o modifique datos del usuario cumple, por orden:

1. `rateLimit(preset)` — `account` (30/min) por defecto, `sensitive` (20/10 min) en login/registro/recupero.
2. `requireAuth()` — el dueño sale de `c.get("user").id` (Bearer validado en Supabase). **Prohibido** aceptar `userId`/`email` del cliente como identidad.
3. `validateJson` / `validateQuery` con Zod (`.strict()` donde aplique).
4. `supabaseForUser()` (RLS activo). `supabaseAdmin()` solo si es imprescindible, documentando el porqué, jamás devolviendo su resultado crudo.
5. Errores genéricos en `snake_case` (`unauthorized`, `not_found`, `rate_limited`) con `requestId`, sin stack/SQL/secretos. El 401 no distingue expirado de inválido.
6. Tablas de usuario con RLS `auth.uid() = user_id` (USING + WITH CHECK) y prueba cruzada A-vs-B antes de mergear.

Checklist completa: [`docs/arquitectura/seguridad.md`](docs/arquitectura/seguridad.md).

## 📦 Respuestas (sobre único)

Éxito (`200`, `201` en creaciones) con `ok()`:

```json
{ "success": true, "data": { "...": "..." } }
```

Error con `fail()` (el launcher y la web gestionan por `code`, nunca por `message`):

```json
{ "success": false, "error": { "code": "unauthorized", "message": "Autenticación requerida.", "requestId": "…" } }
```

Los 400 de validación añaden `details` por campo y el 429 trae `Retry-After`. Tabla completa de códigos: [`docs/api/v1/README.md`](docs/api/v1/README.md#códigos-de-error-estables).

## 🧱 Stack tecnológico

| Capa | Tecnología |
|------|------------|
| Runtime | Cloudflare Workers (edge) + Hono 4.6, `nodejs_compat` |
| Datos y Auth | Supabase (`supabase-js` 2.45): Postgres + Auth + RLS + Storage (`avatars`/`banners` públicos sin listado, `skins`/`capes` privados) |
| Contratos | Zod 3.23 por versión (v2 **copia** de v1, nunca reutiliza) |
| Tubería global | `request-id` → `security-headers` → CORS estricto → rutas → 404 → errores (todo con el mismo sobre) |
| SQL | Paquete versionado `db/v1/install.sql` (uno solo, transaccional) + `verify.sql`, generado con `bun run db:build` |
| Gestión | bun + `tsc --noEmit` (**prohibido npm**) |

## 📁 Estructura

```
api/
├── src/
│   ├── index.ts              # Entrada: tubería global + router versionado (/v1)
│   ├── env.ts                # getEnv(): lectura y validación (prohibido c.env directo)
│   ├── lib/                  # respond.ts (ok/fail), supabase.ts (user/admin), images.ts
│   ├── middleware/           # auth, cors, rate-limit, request-id, security-headers, upload-quota, validate
│   ├── routes/v1/            # health, auth, accounts, files, notifications, friends, users, community + index.ts
│   ├── schemas/v1/           # account, auth, community, cosmetics, files, friends, notifications, privacy
│   └── types/                # app.ts (Bindings, AuthUser, AppEnv)
├── docs/                     # memoria obligatoria (ver docs/README.md)
│   ├── arquitectura/         # decisiones (ADRs), estructura, versionado, seguridad, secretos-y-entorno
│   ├── oauth/                # sesiones, tokens, PKCE, launcher desktop
│   ├── cuentas/              # modelo de cuenta, perfil y vida social
│   └── api/v1/               # contrato HTTP exacto v1 (v2 será carpeta nueva)
├── db/
│   ├── v1/                   # 01–07 (partes) + install.sql (GENERADO) + verify.sql
│   └── migrations/           # 0001–0008 congelado como historial (no se toca)
├── scripts/                  # build-db-bundle.ts (bun run db:build)
├── wrangler.jsonc            # solo config NO secreta (vars: API_ENV, ALLOWED_ORIGINS, SITE_URL)
├── .dev.vars.example         # plantilla local (el real .dev.vars está en .gitignore)
├── AGENTS.md                 # reglas IA de la api
└── README.md                 # estás aquí
```

## 🗄️ Base de datos (un solo pegar-y-ejecutar)

1. Supabase → SQL Editor → pegar `db/v1/install.sql` **completo** → Run (transaccional, idempotente: crea todo o converge sin duplicar ni borrar).
2. Pegar `db/v1/verify.sql` → Run y comparar con los esperados.
3. Evolución: editar solo las partes `01–07`, regenerar con `bun run db:build`, **jamás** `install.sql` a mano. Lo incompatible exige `db/v2/` nuevo.

Detalle: [`db/v1/README.md`](db/v1/README.md) y [`db/README.md`](db/README.md).

## 🔐 Variables y secretos

| Qué | Dónde | Ejemplo |
|---|---|---|
| Config NO secreta | `vars` en `wrangler.jsonc` | `API_ENV`, `ALLOWED_ORIGINS`, `SITE_URL` |
| Secreto en producción | `wrangler secret put` | `SUPABASE_SERVICE_ROLE_KEY` |
| Secreto en local | `.dev.vars` (gitignored, copiar de `.dev.vars.example`) | `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY` |

Si ves un token en un `.ts`, es un bug crítico. Revisión obligatoria antes de cada `commit`/`pull`: `git diff --cached` + buscar `service_role|secret|token\s*=\s*['"]` en `src` y `wrangler.jsonc`. Si aparece un valor real: parar, mover a secreto y **rotarlo**. Detalle: [`docs/arquitectura/secretos-y-entorno.md`](docs/arquitectura/secretos-y-entorno.md).

## 🚀 Desarrollo

```powershell
cd api
bun install                       # bun obligatorio, prohibido npm
copy .dev.vars.example .dev.vars  # rellenar variables locales (nunca commitear secretos)
bun run dev                       # wrangler dev (probar / y /v1/health)
bun run typecheck                 # tsc --noEmit (verificación obligatoria antes de cada PR)
bun run db:build                  # regenera db/v1/install.sql tras tocar 01–07
bun run deploy                    # despliegue a Workers
```

Probar:

```bash
curl http://localhost:8787/
curl http://localhost:8787/v1/health
```

## 🤝 Añadir un endpoint (chuleta)

1. Esquema Zod en `src/schemas/v1/<dominio>.ts` (si hay entrada).
2. Ruta en `src/routes/v1/<dominio>.ts`: `rateLimit` → `requireAuth` (si es cuenta) → `validate*` → handler con `ok`/`fail`.
3. Montar en `src/routes/v1/index.ts` si es dominio nuevo.
4. Docs en `docs/api/v1/<dominio>.md` (+ `docs/cuentas/` u `docs/oauth/` si aplica) **en el mismo cambio**.
5. `bun run typecheck` + revisión de secretos. Si rompe el contrato de otro endpoint, va a versión nueva, no se "ajusta" la vieja.

---

<div align="center">

**NovaStepStudio** — Santiago Stepnicka

[🌐 Página principal](https://steplauncher.pages.dev) · [GitHub](https://github.com/NovaStepStudio) · [Repositorio](https://github.com/NovaStepStudio/StepLauncher)

<sub>© 2026 NovaStepStudio — Powered by NovaCore Engine · No afiliado a Mojang Studios ni a Microsoft.</sub>

</div>
