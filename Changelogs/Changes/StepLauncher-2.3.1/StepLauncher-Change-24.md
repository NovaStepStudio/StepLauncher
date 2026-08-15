# Changes/StepLauncher-2.3.1/StepLauncher-Change-24.md

- **Fecha**: 2026-08-13
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

### 1. Diálogo de bienvenida para el primer inicio (onboarding)

- **Backend** (`internal/Config/Config.go`, `internal/Handlers/App.go`): nuevo flag `firstLaunch` en la config. Nace en `true` para instalaciones nuevas (el archivo de config no existía); los usuarios existentes quedan con `false` porque el campo ausente se deserializa como falso. Nuevos bindings `GetFirstLaunch()` y `SetFirstLaunchDone()` (con setter en el Manager que persiste `firstLaunch: false`).
- **Frontend** (`frontend/web/src/Modals/WelcomeModal.vue` + `Styles/Modals/WelcomeModal.scss`): modal de dos pantallas:
  1. Bienvenida con la imagen `assets/decorations/having-fun.webp` (los estilos usan únicamente las variables del launcher `var(--color-*)`, sin colores extraídos de la imagen).
  2. Configuración de la primera cuenta: cuenta local (solo nombre) o en línea (AuthLib: servidor, usuario y contraseña), reutilizando `createAccount` y `loginAuthlib` del store `Accounts`.
- **App.vue**: tras ocultar el splash, consulta `GetFirstLaunch()` y abre el modal; el cierre (botón X, ESC u "Omitir") marca el onboarding como completado. El modal también se cierra con el sistema idle.

### 2. Modelo 3D de bienvenida con la librería Skin3D

- Renombrados los archivos de `Composables/Skin3D/` a mayúscula inicial: `skin3d.ts → Skin3D.ts`, `model.ts → Model.ts`, `animation.ts → Animation.ts`, `nametag.ts → Nametag.ts` (actualizados sus imports internos y los de `index.ts`).
- Nuevo `Composables/Skin3D/RandomSkin.ts`: importa **como archivos** todos los assets de `web/assets/extra/` (skin-001..004 y cape-001..003) para que Vite los incluya en el build final, y expone `randomSkinCombo()` que elige una skin y una capa al azar (12 combinaciones posibles).
- Nuevo `Composables/Skin3D/UseRandomPlayer.ts`: `createRandomPlayer(canvas)` crea un `Render` con skin+capa aleatoria, animación idle y auto-rotación; reutilizable para bienvenida o inicio de sesión.
- **Dependencias**: se añadieron `three`, `skinview-utils` y `@types/three` (dev) a `frontend/package.json`; la librería las necesita y sin ellas el build de Vite fallaba al resolver `skinview-utils`.

### 3. Indicador de carga: `chicken_jockey_run.gif` en lugar de `loading.gif`

- Reemplazado `assets/gif/loading.gif` por `assets/gif/chicken_jockey_run.gif` en: splash de `App.vue`, carga de fondo de vídeo (`BgLoading`), `LoginProgressModal`, `NewsModal` y `UpdateModal`. Ajustados los tamaños de los spinners (1rem/1.1rem → 2rem aprox.) para que el gif se aprecie.

## Por qué

Primera experiencia del launcher: recibir al usuario con una bienvenida que además le permite dejar lista su primera cuenta sin buscar la sección de cuentas, y un toque visual propio (gallina corriendo) en las transiciones/cargas.

## API afectada

- Go: nuevos métodos expuestos `GetFirstLaunch()` y `SetFirstLaunchDone()` (config `firstLaunch`).
- Frontend: nuevos composables exportados desde `Composables/Skin3D/Index.ts` (`randomSkinCombo`, `createRandomPlayer`).

## Comportamiento anterior/nuevo

- **Antes**: al abrir el launcher por primera vez no había ninguna guía; el usuario debía descubrir dónde añadir una cuenta. El indicador de carga era `loading.gif`.
- **Ahora**: en el primer inicio aparece el diálogo de bienvenida con el 3D de un jugador aleatorio (skin+capa de la carpeta extra) y el formulario de primera cuenta; al completarlo o omitirlo no vuelve a mostrarse. Todas las cargas de pantalla usan `chicken_jockey_run.gif`.

## Cómo verificar

- `go build ./...` en la raíz: OK.
- `bun run build` en `frontend/`: type-check y vite build OK; los assets `skin-001..004.png` y `cape-001..003.png` aparecen hasheados en `dist/assets/img/`.
- Manual: borrar el archivo de config del launcher para simular primer inicio → debe aparecer el diálogo de bienvenida; crear la cuenta local y comprobar que no reaparece al relanzar.

## Corrección posterior (2026-08-13): la bienvenida no aparecía

Al probar borrando la carpeta de config no se mostraba la bienvenida. Causas encontradas y corregidas:

1. **Bug principal — bindings Wails**: los métodos `GetFirstLaunch`/`SetFirstLaunchDone` se habían añadido solo a `Handlers.App`, pero el objeto expuesto a Wails es el `App` raíz (`app.go`), que no los delegaba. En el frontend `window.go.main.App.GetFirstLaunch` era `undefined` y el `?.()` fallaba en silencio (`undefined === true` → nunca se mostraba). Se añadieron los delegados en `app.go`.
2. **Carrera con el failsafe del splash**: la comprobación de primer inicio se hacía al final de `onMounted`, después de llamadas de red (`loadAccounts`, `loadVersions`, `loadProfiles`, etc.). Si tardaban más de los 6 s del failsafe, el splash se ocultaba antes y la bienvenida se perdía. La comprobación ahora se ejecuta al inicio de `onMounted`, antes de cualquier await de red; `hideSplash` (tanto por flujo normal como por failsafe) decide si mostrar la bienvenida.
3. **Refuerzo backend**: `Handlers.App` recuerda si el archivo de config no existía al arrancar (`firstLaunchPending`), de modo que `GetFirstLaunch()` devuelve `true` también cuando el flag faltaba (p. ej. config creada por un binario anterior sin el flag), hasta que se complete/omita el onboarding.
