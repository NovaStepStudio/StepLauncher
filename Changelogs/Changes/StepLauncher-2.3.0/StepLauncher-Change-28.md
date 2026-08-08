# StepLauncher-Change-28: Modal de instalación con spinner en "verificando" + visor de capturas con zoom/pan reales

- **Fecha**: 2026-08-06
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

### 1. Modal de instalación (`frontend/web/src/Modals/InstallationModal.vue`)

- El anillo de progreso ya no muestra un porcentaje engañoso durante los pasos
  de proceso: `ringMode` computed distingue `percent` (descargando/pausado/
  pendiente), `busy` (verificando/re-downloading) y `done` (completed).
- En modo `busy` el centro del círculo es un spinner giratorio (en vez del
  `%`), y el arco exterior gira en bucle (`stroke-dasharray: 42 285`) en lugar
  de quedarse congelado en 100%.
- En modo `done` el centro muestra un `IconCheck`.
- El badge de paso (`InstallationModal_StepBadge`) se movió fuera de la fila de
  estado a un bloque bajo el círculo (`InstallationModal_Step`), con tipografía
  más grande (0.92rem) y colores por estado (`accent`/`warn`/`ok`/`muted`).
- Nuevo texto de ayuda `stepHint` bajo el badge según el paso: "Comprobando la
  integridad de los archivos descargados…", "Extrayendo las bibliotecas
  nativas…", "Reparando los archivos que fallaron la verificación…".

### 2. Visor de capturas (`frontend/web/src/Modals/ScreenshotsModal.vue`)

- El visor dejó de usar `transform: scale()` (la imagen tapaba los botones y
  no había forma de moverse). Ahora el escenario (`Shots_PreviewStage`) ocupa
  toda la pantalla bajo la barra superior y scrollea; la imagen crece en
  píxeles explícitos (`scaledSize` = naturales × zoom) dentro de un wrapper con
  `margin: auto` que la centra sin zoom y permite scroll con zoom.
- Rueda del ratón = zoom (paso 0.1, límite 1–4); arrastrar = moverse (solo con
  zoom > 1, cursores grab/grabbing). Botones y barra superior con `z-index`
  por encima, clicables.
- Escape cierra el visor; flechas cambian de captura; `+`/`-` hacen zoom.
- Las cards son más grandes (minmax 13.5rem) y el subtítulo (tamaño · fecha)
  usa `--text-secondary` con `opacity: 0.85`.

## Por qué

- En "verificando" el progreso llegaba al 100% antes de acabar y el único
  cambio era un label pequeño: el usuario pedía que el contenido del círculo
  cambiase (spinner/check) y el estado se viera claramente bajo el círculo.
- El zoom del visor estaba roto y dejaba al usuario atrapado sin saber cómo
  salir ni cómo moverse por la imagen ampliada.

## API afectada

- Sin cambios de bindings ni de backend: solo `InstallationModal.vue`, su SCSS
  y `ScreenshotsModal.vue`.

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (frontend) → OK.
- Instalar una versión y observar "verificando": el círculo debe mostrar
  spinner girando (nunca 100% estático) y el badge/hint bajo el círculo.
- Abrir Fotos con una captura: zoom con la rueda, arrastrar para moverse,
  botones clicables por encima, Escape cierra.
