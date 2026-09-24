# Uso del launcher — cómo debe consumir este servicio

El launcher StepLauncher ya habla Yggdrasil genérico (`TypeAuthLib` con
`AuthServerURL`, `LoginAuthLib`, `RefreshAuthLib`, `ResolveForLaunch` en
`launcher/internal/Core/Accounts/`): **no hay que reinventar el login**, solo
apuntar al nuevo Worker y lanzar con authlib-injector. (Este archivo solo
documenta el consumo; no modifica `launcher/`.)

## Configuración (una vez)

- `AuthServerURL = https://<tu-worker-yggdrasil>/` (raíz, con `/` final).
  Ejemplo local: `http://localhost:8787/`. Vale también con sufijo
  `/authserver`, pero se recomienda la raíz (así el pre-verify y el
  descubrimiento ALI van directo).
- **Nombre de jugador**: StepLauncher permite 3–20 con `._-`, pero Minecraft
  en `online-mode` exige 3–16 con `[A-Za-z0-9_]`. Las cuentas con nombre
  incompatible entran al login pero el juego/servidor las rechaza: el
  usuario debe cambiar su `username` en la web antes de jugar online.
- Descubrimiento: `GET <AuthServerURL>` → `{ meta, skinDomains,
  signaturePublickey }` + cabecera `X-Authlib-Injector-API-Location`.
  Guardar `serverName` para mostrarlo y cachear la pública para verificar
  texturas sin pedirla cada vez.

## Crear cuenta Yggdrasil en el launcher

1. Pedir **email o usuario de StepLauncher + contraseña** (la misma de
   `api/`). Avisar que la cuenta debe tener el **correo confirmado**.
2. Llamar `POST <AuthServerURL>/authserver/authenticate` con
   `{ username, password, clientToken?, requestUser: true,
   agent: { name: "Minecraft", version: 1 } }`.
   - `clientToken`: generar un UUID propio por instalación y **reutilizarlo**
     siempre (como hace el launcher vanilla).
3. Guardar `{ accessToken, clientToken, uuid: selectedProfile.id,
   username: selectedProfile.name }` en el almacén seguro del SO (nunca en
   logs). Marcar `sessionValid = true`.
4. Si `403` → credenciales o sin confirmar: mostrar error genérico y ofrecer
   ir a la web a confirmar/reenviar (flujo de `api/`, no de aquí).

## Jugar (lanzamiento)

1. Antes de lanzar: `POST …/authserver/validate` con los tokens guardados.
   Si `403`, intentar `POST …/authserver/refresh` **una vez**; si sigue
   fallando, pedir login (sin bucles).
2. Descargar/cachear `authlib-injector.jar` y pasar al juego:
   `-javaagent:<ruta>/authlib-injector.jar=<AuthServerURL>`
   más `-Dauthlibinjector.yggdrasil.prefetched=<base64 de GET />`
   (evita un viaje de red al arrancar y caídas por red).
3. Argumentos del juego: `--uuid <selectedProfile.id sin guiones>
   --accessToken <accessToken> --userType mojang --username <name>`.
4. Al cerrar sesión en el launcher: `POST …/authserver/invalidate` y borrar
   tokens locales. Para "cerrar todo": `POST …/authserver/signout` con
   usuario+contraseña.

## Servidores (para quien hostea)

- `online-mode=true` en `server.properties` y arrancar con
  `java -javaagent:authlib-injector.jar=<AuthServerURL> -jar server.jar nogui`.
- El servidor llamará a `…/session/minecraft/hasJoined` de este Worker por
  cada entrada: debe ser alcanzable por HTTPS desde el host del servidor.
- Mezclar premium Mojang y StepLauncher en un mismo servidor exige
  `FallbackAPIServers`/passthrough (fuera del MVP: aquí solo StepLauncher).

## Reglas del launcher (no negociables)

- Guardar tokens en almacenamiento seguro del SO; renovar antes de `expiresAt`.
- Ante `401/403` del juego o del Worker: un refresh y luego login (no reintentos
  infinitos).
- `refreshToken` de `api/` (Supabase) y `accessToken` Yggdrasil **no se
  mezclan**: el primero abre la cuenta, el segundo abre el juego.
- Sin `selectedProfile` no se juega: es cuenta sin `profiles` (caso raro).
  Mandar al usuario a la web a completar su perfil.
- La skin/capa se sube desde la web (`api/`); este servicio solo la lee y la
  sirve. Si el avatar no carga, el jugador aún no subió skin (defecto Steve).
