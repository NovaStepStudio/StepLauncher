# Cambios de StepLauncher 2.5.0 (Change-15) — Config inicial completa, mitad de RAM, richPresence en launcher, auto-refresh activo y fondo desde capturas

- **Fecha**:   2026-09-24
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

### 1. Config inicial completa y paleta nueva oscuro moderno morado (`internal/Config/Config.go`, `frontend/web/src/Common/Stores/Ui.ts`, `Settings/Sections/Personalization.vue`, `Common/Styles/base/_variables.scss`)

- Los valores `#111` (modal, botones, play) ni siquiera pasaban el propio `sanitizeColorString` (exige `#RGB` de 4 dígitos o `#RRGGBB`): al sanear se caían al fallback y la primera config quedaba a medias.
- Nueva paleta por defecto dark mode (a pedido del owner, sin tonos morados): sidebar `#00000066`, modal `#0b0b0b`, botones `#2b2b3d`, bordes `#131313`, progreso `#5ed89a`, jugar `#1f1f1f`, primario `#19191a`, error `#ff6b6b`, éxito `#34d399`, tag `#a974ff`, aviso `#fbbf24`.
- `Default()` ahora inicializa `DynamicImages: []string{}` y `RecentColors: []string{}` (antes `nil` → `null` en el JSON) y `sanitize()` garantiza que nunca queden en `nil`.
- Fallbacks replicados en `Ui.ts`, `Personalization.vue` y `_variables.scss` para que backend y frontend arranquen con los mismos colores.

### 2. RAM inicial = mitad de la RAM del equipo, entre 2 y 8 GB (`internal/Config/Config.go`)

- Nueva función `defaultRAMGB()`: lee `platform.TotalRAMMB()`, calcula la mitad y la limita a mínimo 2 GB y máximo 8 GB (si no se detecta, 4 GB).
- `Default()` usa `defaultRAMGB()` en vez del fijo `2`, así el primer arranque y `Reset()` asignan la mitad real (p. ej. 16 GB → 8 GB, 8 GB → 4 GB).

### 3. `richPresence` movido dentro del bloque `launcher` como booleano simple (`internal/Config/Config.go`, `internal/Handlers/App.go`, `internal/Services/Account/Service.go`, `Settings/Sections/General.vue`)

- Antes: bloque propio `"richPresence": {"enabled": true}`. Ahora: `"launcher": { ..., "richPresence": true }` (`*bool` en Go con `nil` = true para distinguir "ausente" de `false`).
- `migrateRichPresence()` traslada el bloque antiguo (objeto o booleano) a `launcher.richPresence` sin pisar el valor nuevo si ya existe; al guardar desaparece el bloque viejo.
- `GetRichPresenceConfig()` ahora devuelve `bool` y `General.vue` lee `cfg.launcher?.richPresence ?? true`.

### 4. Renovación de sesiones activada desde el inicio (`internal/Core/Accounts/Manager.go`, `frontend/web/src/Accounts/Store.ts`)

- `NewManager` arranca con `AutoRefresh: true` y `Load()` migra los `launcher_accounts.json` antiguos sin la clave a `true` (respeta el `false` explícito).
- El store parte de `ref(true)` y ante respuesta ausente usa `ar !== false`.

### 5. Botón "Colocar como fondo" en el visor de Screenshots (`internal/Handlers/Background.go`, `internal/Services/Appearance/Service.go`, `Screenshots/Screenshots.vue`, `Screenshots/Styles/Screenshots.scss`)

- Nuevo binding `SetScreenshotAsBackground(relPath)`: valida que la ruta quede dentro del directorio del launcher, copia la captura a `cache/backgrounds/` vía `ImportBackground` y la fija como `background.type = "image"` (limpia autor/mod/url/vídeo).
- En el visor grande (cabecera `Shots_PreviewTools`) hay botón con `IconWallpaper` + mensaje de confirmación, con el mismo flujo de refresco que la galería de mods (`GetConfig` → `applyPersonalization` → `refreshBackground`).

## Por qué

A pedido del owner: la primera config salía incompleta y con grises feos, los 2 GB fijos se quedaban cortos en versiones nuevas, `richPresence` no debía ser bloque propio, la renovación de sesiones debía venir activa y las capturas debían poder ponerse de fondo como las de mods.

## API afectada

- Bindings regenerados (`wails3 generate bindings -ts -i`): `Config/models.ts` (`launcher.richPresence?: boolean | null`, sin bloque `richPresence`), `accountservice.ts` (`GetRichPresenceConfig(): boolean`), `appearanceservice.ts` (nuevo `SetScreenshotAsBackground`).
- Sin cambios en eventos ni en runtime.

## Comportamiento anterior/nuevo

- Antes: primera config con `maxRamGB: 2`, colores `#111`/`#0005`, `dynamicImages: null`, `recentColors: null`, bloque `richPresence` propio y `autoRefresh` apagado; las capturas no podían usarse de fondo.
- Ahora: primera config con mitad de RAM (2–8 GB), paleta dark válida, arreglos no nulos, `launcher.richPresence` booleano (con migración del bloque viejo), `autoRefresh` en true (con migración) y botón de fondo en el visor de capturas.

## Cómo verificar

- `go build ./...` en `launcher/` (pasa).
- `wails3 generate bindings -ts -i` en `launcher/` (320 paquetes, 9 servicios, 213 métodos).
- `bun run build` en `launcher/frontend` (pasa `vue-tsc --build` y `vite build`).
