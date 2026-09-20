# StepLauncher-Change-3: SkinPlayer adaptado a Canvas API del navegador

- **Fecha**: 2026-08-03
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El composable `Composables/SkinPlayer` dependía del paquete de Node `canvas`
(`createCanvas`, `loadImage`, `Image`) y de `Buffer`, que no existen en el
navegador (WebView2/Chromium de Wails). Se adaptó por completo a la Canvas API
del navegador para que compile (`vue-tsc`) y funcione en el frontend.

**Frontend (Vue 3 + TS):**
- `utils.ts`:
  - Nuevo tipo `SkinBitmapSource` (`ImageBitmap | HTMLImageElement | HTMLCanvasElement`)
    que sustituye al `Image`/`Buffer` de `canvas`.
  - `loadSkin(source: Blob | string)` → `ImageBitmap` mediante `fetch` +
    `createImageBitmap` (reemplaza a `loadImage` de Node).
  - `checkSlim` usa `document.createElement('canvas')` y
    `getContext('2d', { willReadFrequently: true })` en vez de `createCanvas`.
  - Nuevo helper `canvasToBlob(canvas)` que envuelve `canvas.toBlob()` en una
    `Promise<Blob>` (reemplaza a `canvas.toBuffer()`).
- `renderers/head.ts` y `renderers/full-body.ts`: usan `document.createElement('canvas')`
  y devuelven `Promise<Blob>` PNG en lugar de `Buffer`.
- `functions/render-head.ts` y `functions/render-full-body.ts`: aceptan
  `SkinBitmapSource` y devuelven `Promise<Blob>`; los cálculos de escalas son
  idénticos.
- `index.ts`: reexporta con `export type` los tipos (requisito de
  `verbatimModuleSyntax`) y añade `loadSkin`/`canvasToBlob` a la API pública.
- Todos los imports de tipos usan `import type` (el tsconfig del proyecto tiene
  `verbatimModuleSyntax` activado).
- Corrección de errata heredada: la capa del brazo slim en
  `getFullBodyModern` usaba `44, 36` sin multiplicar por `inputScale`
  (rompía skins HD con `inputScale > 1`); ahora usa `44 * inputScale`.

## Por qué

El paquete `canvas` de Node no se ejecuta en el navegador y `vue-tsc` fallaba
con `TS2307 (Cannot find module 'canvas')`, `TS2591 (Buffer)` y `TS1484`
(imports de tipos con `verbatimModuleSyntax`), bloqueando `bun run build`.

## API afectada

- Firma de las funciones exportadas por `Composables/SkinPlayer`:
  - `renderHead(skin, options?)` → `Promise<Blob>` (antes `Promise<Buffer>`).
  - `renderFullBody(skin, options?)` → `Promise<Blob>` (antes `Promise<Buffer>`).
  - `checkSlim(skin)` acepta `SkinBitmapSource` (antes `Buffer | Image`).
  - Nuevos: `loadSkin(Blob | string)` y `canvasToBlob(canvas)`.
  - El consumidor debe cargar la textura con `loadSkin` o pasar directamente
    una `ImageBitmap`/`HTMLImageElement`/`HTMLCanvasElement`.

## Cómo verificar

- `bun run build` en `frontend/` (incluye el type-check de `vue-tsc`): debe
  compilar sin errores.
- En la app: cargar un skin con `loadSkin` y renderizar `renderHead`/
  `renderFullBody` con `scale > 1`; el resultado se muestra con
  `URL.createObjectURL(blob)`.
