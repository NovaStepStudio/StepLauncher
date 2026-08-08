# StepLauncher-Change-5: Cache de skins en disco + avatar en la UserCard

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

**Cache de skins en disco:** la skin y la cabeza (avatar) de cada cuenta
AuthLib se guardan en `workDir/cache/avatars/<id>.json` y se sirven desde
disco sin repetir HTTPS. En cada refresh de tokens (login, renovación manual
o automática) el cache se **reestablece**: se elimina el archivo viejo y se
descarga y guarda la nueva.

**Avatar en la UserCard:** la tarjeta de usuario (arriba a la derecha) muestra
la cabeza de la skin de la cuenta activa, en lugar de la imagen de relleno
`avatar_not_found.png`.

**UI:** el nombre del usuario (y el subtítulo) de la UserCard se corta con
`...` si es demasiado largo; el avatar de `AccountsView` queda sin
box-shadow.

**Backend (Go), `internal/Core/Accounts/Manager.go`:**
- `AccountAssets`: si existe `workDir/cache/avatars/<id>.json` se emiten las
  data URLs cacheadas directamente (evento con `cached: true`) **sin hacer
  ninguna petición HTTPS**; si no, descarga la skin, extrae la cabeza y
  persiste el cache (`writeSkinCache`).
- `clearSkinCache(id)`: elimina el archivo de cache de una cuenta (json y tmp).
- `LoginAuthLib`: al iniciar sesión se reestablece el cache (la próxima
  petición baja y guarda la skin nueva).
- `RefreshAuthLib`: al validar o renovar la sesión, elimina el cache viejo y
  relanza la descarga de la skin nueva en segundo plano (`go m.AccountAssets`).
- `ResolveForLaunch`: al renovar el token automáticamente (durante el
  lanzamiento) elimina el cache, sin descargar en el camino crítico.
- Los tokens guardados en el JSON de cache son data URLs (la skin completa +
  la cabeza), por lo que el re-lector no necesita red ni decodificación.

**Frontend (Vue 3 + TS):**
- `App.vue` (UserCard): el avatar muestra `accountAvatars[selectedAccountId]`
  (cabeza de la skin de la cuenta activa) y `avatar_not_found.png` como
  fallback.
- `Styles/App.scss` (`.UserCard .Username`): `white-space: nowrap` +
  `text-overflow: ellipsis` + `min-width: 0` en `h1` y `p` para truncar
  nombres largos.
- `Modals/AccountsView.vue`: se elimina el `box-shadow` del `.Acct_Avatar`
  (queda solo el círculo con la imagen o la inicial).

## Por qué

- Evitar repetir la descarga HTTPS de la skin en cada apertura de la vista de
  cuentas o cada arranque: la skin no cambia entre sesiones salvo que el
  usuario la cambie en el servidor, y ese cambio se recoge al refrescar la
  sesión (el cache se reestablece).
- La UserCard quedaba con la imagen de relleno aunque la cuenta tuviera skin:
  ahora muestra la cabeza del jugador activo.
- Nombres largos desbordaban la tarjeta; ahora se cortan con `...`.

## API afectada

- Ningún binding nuevo (el evento `account_assets` ya existía; ahora el
  payload incluye `cached: true` cuando viene del disco).
- Modelo `AccountAssets` sin cambios en esta iteración.

## Cómo verificar

- `go build ./...` → OK.
- Con una cuenta AuthLib: la primera petición de assets crea
  `%APPDATA%\.StepLauncher\cache\avatars\<id>.json`; al reabrir la app la
  skin se sirve de la cache (en el log: "servidos desde cache") sin red.
- Al renovar sesión (botón renovar / login nuevo): el cache se elimina y se
  descarga la skin de nuevo.
- La UserCard muestra la cabeza de la cuenta activa; un username largo se
  corta con `...`; en AccountsView el avatar no tiene sombra.
