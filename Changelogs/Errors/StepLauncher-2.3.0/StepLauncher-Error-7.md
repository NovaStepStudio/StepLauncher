# StepLauncher-Error-7: El ColorField corrompía los colores por defecto (`#0005`/`#111` → `#000000`)

- **Fecha**: 2026-08-04
- **Estado**: corregido
- **Fixed?**: Sí — corregido y mencionado en la release `StepLauncher-2.3.0`.
- **Versión afectada**: 2.3.0 (productVersion de `wails.json`)

---

## 1. Síntoma

Al abrir por primera vez el selector de color (ColorField) en Configuración →
Personalización → Colores, los colores de la configuración perdían toda la
estética:

```json
"colors": {
  "sidebar": "#0005",
  "modal": "#111",
  "buttons": "#111",
  "borderModal": "#494949",
  "border": "rgba(37, 37, 37, 0.3)"
}
```

pasaban a:

```json
"colors": {
  "sidebar": "#000000",
  "modal": "#000000",
  "buttons": "#000000",
  "borderModal": "#494949",
  "border": "rgba(37,37,37,0.30)"
}
```

Solo con abrir el selector (sin tocar nada) se guardaba el negro sólido
`#000000`, destruyendo la semitransparencia de `#0005` y el gris de `#111`.

## 2. Causa raíz

`parseColor()` en `ColorField.vue` solo reconocía hex de **6 y 8 dígitos**
(`#rrggbb` / `#rrggbbaa`) y `rgb()/rgba()`. Los valores por defecto de la
config usan **hex corto** de CSS: `#0005` (4 dígitos, RGBA) y `#111`
(3 dígitos). Al no parsearlos, `syncFromValue()` salía sin inicializar el
selector, que quedaba en su estado por defecto (negro `#000000`) y
`lastEmitted` sin fijar.

El cierre del selector (auto-cierre de 5s o toggling) llama a
`commitIfDirty()`, que compara `outputText()` (`#000000`) contra
`lastEmitted` (`''`) → considera que hay cambios → `commit()` → `save()` →
`UpdatePersonalization` → el negro se persistía en disco. Abrir y cerrar el
campo era suficiente para corromper el color.

## 3. Solución aplicada

- `parseColor()` ahora acepta también hex de 3 y 4 dígitos (`#RGB`, `#RGBA`):
  - `#RGB` se expande duplicando cada dígito (`#111` → `#111111`).
  - `#RGBA` calcula el alpha con la escala de 4 dígitos (dígito × 17) y expande
    los tres primeros (`#0005` → RGB negro con 33% de opacidad).
  - Se usa `slice()` en lugar de indexado `h[i]` porque el proyecto tiene
    `noUncheckedIndexedAccess` (TS lo rechaza con `string | undefined`).
- Con el parseo correcto, `syncFromValue()` fija `lastEmitted` al valor
  expandido (p. ej. `#00000054`), por lo que abrir/cerrar sin tocar nada ya no
  genera escrituras; si el usuario cambia el color, se guarda el valor
  expandido, semánticamente equivalente (la transparencia se conserva).

## 4. Verificación

- `go build ./...` → OK.
- `bun run build` (dentro de `frontend/`, type-check incluido) → OK.
- Traza mental del flujo: `#0005` se parsea a rgb(0,0,0) alpha 33 → el swatch
  y el panel muestran negro semitransparente (no negro sólido) → abrir y
  cerrar sin cambios no persiste nada → el JSON de config conserva la estética.

## 5. Regla aprendida

`parseColor` debe aceptar TODOS los formatos que la propia app produce o usa
como default (`#RGB`, `#RGBA`, `#rrggbb`, `#rrggbbaa`, `rgb()`, `rgba()`); si
un formato válido para CSS no se parsea, el selector arranca en negro y el
primer commit lo persiste. En TypeScript con `noUncheckedIndexedAccess`, no
indexar strings con `h[i]`: usar `slice()`.
