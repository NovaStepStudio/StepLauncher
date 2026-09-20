# StepLauncher-Change-4: Avatar = cabeza de la skin (head) en las cuentas

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El avatar de las cuentas AuthLib ahora es **solo la cabeza (head) de la skin
del jugador**, recortada de la textura en el backend y mostrada en la lista de
cuentas (`AccountsView`), en lugar de la inicial del nombre.

**Backend (Go):**
- `internal/Core/Auth/Authlib.go`:
  - Refactor de la descarga de imágenes: `fetchImageBytes` (bytes +
    Content-Type) compartida por `FetchImageDataURL` y la nueva
    `FetchSkinWithHead`.
  - `FetchSkinWithHead(ctx, url)`: descarga la skin **una sola vez** y devuelve
    la data URL de la skin completa y la del avatar.
  - `extractHeadPNG`: recorta la cabeza de la skin:
    - Cara frontal (head): área (8,8)-(16,16) de 8x8 px.
    - Sombrero (capa 2, overlay): área (40,8)-(48,16) superpuesta encima,
      solo en skins 64x64 y respetando los píxeles transparentes.
    - Escalado 8x (64x64) con vecino más cercano: queda pixelado, como en
      Minecraft.
    - Si la skin no se puede procesar, devuelve la head vacía sin error.
- `internal/Core/Accounts/Types.go`: nuevo campo `AvatarDataURL
  (json:"avatarDataUrl")` en `AccountAssets`.
- `internal/Core/Accounts/Manager.go` (`AccountAssets`): usa
  `FetchSkinWithHead`; el avatar llega al frontend por el evento
  `account_assets`.

**Frontend (Vue 3 + TS):**
- `stores/accounts.ts`:
  - Nuevo estado reactivo `accountAvatars` (id → data URL del avatar).
  - `ensureAssetsListener`: listener único del evento `account_assets`
    (se registra bajo demanda, sin duplicados).
  - `fetchAccountAvatar(id)`: pide la skin al backend solo si el avatar no
    está cacheado (evita peticiones repetidas).
  - `loadAccounts()` dispara la carga de avatares de las cuentas AuthLib en
    segundo plano (también tras login/refresh vía los listeners de App.vue).
- `Modals/AccountsView.vue`: el `.Acct_Avatar` muestra `<img>` con la cabeza
  de la skin cuando está disponible y la inicial como fallback; el estilo
  mantiene el círculo con `image-rendering: pixelated`.

## Por qué

El evento `account_assets` y el binding `GetAccountAssets` existían pero **no
se consumían en el frontend**: nunca se mostraba ninguna imagen de la cuenta.
Además se quería el avatar clásico de Minecraft: la cabeza de la skin, no la
skin completa. Ahora cada cuenta AuthLib pinta su cabeza automáticamente.

## API afectada

- `AccountAssets` (modelo Go / `frontend/wailsjs/go/models.ts`): campo nuevo
  `avatarDataUrl` (solo lectura; no rompe bindings existentes).
- Bindings: ninguno nuevo (`GetAccountAssets` ya existía).

## Cómo verificar

- `go build ./...` → OK; `bun run build` en `frontend/` → OK (type-check).
- En la app: abrir Cuentas con una cuenta AuthLib que tenga skin: el avatar
  pasa de la inicial a la cabeza pixelada de su skin en unos segundos. Sin
  skin, se mantiene la inicial.
- Skins 64x64 con sombrero: la capa del sombrero se superpone a la cara;
  skins 64x32: solo la cara.