# StepLauncher-Change-39: Renombrado a mayúsculas en `src/` y build dividido en chunks

## Fecha
Agosto 2026

## Release
StepLauncher-2.3.0 — se mencionó por primera vez en esta release.

## Contexto
El usuario pidió que **todo el contenido de la carpeta `frontend/web/src`** empiece con mayúscula (carpetas y archivos). Quedan excluidas de esta regla: `index.html`, `index.css`, `env.d.ts` (nombres canónicos del toolchain), `node_modules/`, `wailsjs/` (generado por Wails) y `dist/`. La carpeta raíz `src` **conserva su nombre en minúsculas** (no se renombra a `Src`). `assets/` y todos sus archivos **no se tocan** (siguen en minúsculas).

## Qué se hizo
- `web/src/Renamos all contenido:
  - `stores/` → `Stores/` y cada archivo con inicial mayúscula: `Accounts.ts`, `Colorfield.ts`, `Downloads.ts`, `Fonts.ts`, `Idle.ts`, `Launcher.ts`, `Ui.ts`, `Update.ts`.
  - `Composables/SkinPlayer/`: `index.ts` → `Index.ts`, `utils.ts` → `Utils.ts`, `functions/` → `Functions/`, `functions/render-full-body.ts` → `RenderFullBody.ts`, `functions/render-head.ts` → `RenderHead.ts`, `structures/` → `Structures/`; `renderers/` → `Renderers/`, `renderers/full-body.ts` → `FullBody.ts`, `renderers/head.ts` → `Head.ts`, subcarpeta `Structures/`.
- `main.ts` → `Main.ts` (ampliado el archivo de `src/main.ts` a `src/Main.ts` en `index.html`).
- Actualización de todos los imports afectados en `.vue`, `.ts` y `.scss` (rutas del tipo `../stores/...` → `../Stores/...` y archivos con inicial mayúscula), siguiendo los errores del build para no tocar nada más.
- `vite.config.ts`: limpieza de la agrupación de chunks:
  - Se elimina el chunk manual de `axios` (dependency no existe, no debe referenciarse).
  - Se conservan: `settings`, `modals`, `skin-player`, `stores`, `composables`, `vendor-icons`, `vendor-vue` y `vendor` general.
  - `assetsInlineLimit: 0` + `assetFileNames` separado: `assets/css`, `assets/fonts`, `assets/img`, `assets/js`.
  - `emptyOutDir: true` (limpiaba dist de builds anteriores).

## Resultado del build verificable
- Chunks JS separados: `index-*.js`, `settings-*.js`, `modals-*.js`, `vendor-icons-*.js`.
- CSS separados: `index-*.css`, `settings-*.css`, `modals-*.css`.
- Imágenes y fuentes en `assets/img/` y `assets/fonts/`.
- `bun run build` con type-check OK (`vue-tsc --build` sin errores).

## Notas
- La renamentación se hizo cont en Windows: NTFS es case-insensitive, así que el cambio de caja de carpeta `src`→`Src` NO se debe hacer (se dejó `src` tal cual); solo se renombra el contenido interno.
- Si en CI/Linux el build diferencia mayúsculas, los imports ya apuntan exactamente a los nombres renombrados, por lo que es seguro.