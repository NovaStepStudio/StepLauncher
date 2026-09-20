# StepLauncher-Change-19: Color de progreso propio y personalizable, preview con réplica del modal, detalles técnicos sin corte y cancelación al cerrar la app

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

Ronda de refinamiento de la descarga iniciada en `StepLauncher-Change-18`.

### 1. Color de la barra de progreso: variable PROPIA y personalizable

Antes `--progress-color` se derivaba del color de los botones (`c.buttons`), por
lo que no era un color independiente ni se podía editar. Ahora es un campo de
color registrado de verdad:

- **Backend** (`internal/Config/Config.go`): nuevo campo
  `ThemeColors.Progress` (`json:"progress"`), default `#5ed89a` en `Default()`,
  saneado en `sanitize()` (`sanitizeColor(&...Progress, "#5ed89a")`) y
  registrado en el historial de colores de `UpdatePersonalization`.
- **Frontend** (`stores/ui.ts`): `ThemeColors.progress` en la interfaz,
  `default '#5ed89a'` en `normalizePersonalization` y
  `applyRootVar('--progress-color', c.progress)` (ya no se lee de `buttons`).
- **CSS** (`frontend/web/index.css`): `--progress-color: #5ed89a;`.
- **Settings** (`PersonalizationSettings.vue`): nuevo `ref colorProgress`,
  cargado en `onMounted`, enviado en `buildPersonalization()`, incluido en
  `trackRecents()` y nueva fila "Progreso de descarga" con `ColorField`
  (`preview` + `example`).

### 2. Preview en vivo con réplica del modal de descarga

`ColorField.vue` gana la prop `example` (solo usada por el campo de progreso).
Al arrastrar el color el selector se centra en pantalla (modo preview) y, **al
lado**, aparece una réplica en miniatura del modal de descarga:

- Cabeza con icono teñido del color y "Descargando el juego · 1.20.1 · Fabric…"
- Aro SVG con el color elegido (`:stroke="previewStyle"`) al 64%.
- Línea "Descargado 1.9 MB de 2.8 MB · 4.2 MB/s · assets".
- Chip "Detalles técnicos".

La réplica usa las variables CSS reales del modal (`--background-modal-primary`,
`--border-modal-style`, etc.) y se repinta en tiempo real mientras se arrastra.
Se retiró el mini-ejemplo que se había colocado dentro del panel (el ejemplo
debe verse AL LADO del selector centrado).

### 3. Detalles técnicos sin corte ni doble scroll

`.InstallationModal_Details` dejó de tener `max-height`/`overflow-y` propios ni
padding de contenedor: ahora crece libremente y el scroll lo gestiona el cuerpo
del modal (`.InstallationModal_Body`, que ya era el scroll del conjunto). Se
elimina así el scroll anidado que cortaba las filas y el desborde que el
padding empujaba al contenedor padre. El registro también queda sin scroll
propio (10 líneas).

### 4. Cancelación garantizada al cerrar la aplicación

- `internal/Core/Downloader/Manager.go`:
  - `CancelActive()` cancela todas las descargas en
    `pending/downloading/paused/verifying/redownloading` (lee el estado bajo su
    lock y llama a `Cancel` por id).
  - `runDownload()` ahora comprueba al arrancar si el contexto ya fue
    cancelado (`dl.ctx.Err() != nil`): si la descarga fue cancelada mientras
    estaba en cola (p. ej. al cerrar la app), marca `StateCancelled` y sale sin
    arrancar (antes podía reactivarse al ejecutarse el job).
- `internal/Handlers/Engine/Engine.go`: `Engine.Shutdown()` llama a
  `accounts.CancelLogin()`, `StopAllGames()` y `CancelActive()` sobre los dos
  managers de descarga (`downloader` y `sharedDl`).
- `app.go`: `App.shutdown(ctx)` → `engine.Shutdown()`.
- `main.go`: se registra el hook `OnShutdown: app.shutdown`.

Con esto, al cerrar la ventana se abandona todo el trabajo en curso de forma
limpia, sin descargas dando vueltas ni errores de fondo tras el cierre.

## Archivos tocados

- `internal/Config/Config.go`
- `internal/Core/Downloader/Manager.go`
- `internal/Handlers/Engine/Engine.go`
- `app.go`, `main.go`
- `frontend/web/index.css`
- `frontend/web/src/stores/ui.ts`
- `frontend/web/src/Layouts/Sections/Settings/PersonalizationSettings.vue`
- `frontend/web/src/Layouts/Sections/Settings/ColorField.vue`
- `frontend/web/src/Styles/InstallationModal.scss`

## Por qué

- El color del progreso debía ser un color propio e independiente, editable en
  Configuración → Personalización → Colores (antes se derivaba de botones y no
  se podía cambiar).
- El preview del color debía mostrar AL LADO cómo queda el modal de descarga,
  no un bloque perdido dentro del selector.
- Los detalles técnicos se cortaban por `max-height`+`overflow-y` internos y el
  padding desbordaba el contenedor padre; con scroll único en el cuerpo se ven
  completos.
- Cerrar la app debía cancelar siempre las descargas para no dejar procesos ni
  rastro de errores.

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- En Settings → Personalización → Colores → "Progreso de descarga": al abrir el
  selector y arrastrar, el campo se centra y a su derecha aparece la réplica del
  modal coloreándose en vivo; al soltar se guarda y `--progress-color` cambia.
- Cerrar la app con una descarga activa ya no deja la descarga corriendo.
- CSS compilado contiene `--progress-color`, `.CfReplica`, y
  `.InstallationModal_Details` sin `max-height` interno.