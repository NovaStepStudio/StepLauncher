# StepLauncher-Change-14: Bordes como colores independientes registrados en la Config

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

`--border-modal-style` y `--border-style` ya no se derivan de los fondos con
`color-mix(...)`; ahora son **colores independientes** que viajan en la
configuración (`personalization.colors`), como el resto de colores del tema.

**1) Variables en `index.css` (valores por defecto independientes):**

```css
--border-modal-style: 1px solid #494949;
--border-style: 1px solid rgba(37, 37, 37, 0.3);
```

**2) Config (backend y frontend):** se reincorpora `colors.border` a
`ThemeColors` (se había retirado en Change-13) junto a `colors.borderModal`:

- `internal/Config/Config.go`: campo `Border` en `ThemeColors`, default
  `"rgba(37, 37, 37, 0.3)"`, sanitize y registro en el historial de recents.
- `stores/ui.ts`: interfaz `ThemeColors.border`, normalización con su default
  y `applyPersonalization` escribe ambas variables directamente desde la
  config:

```ts
applyRootVar('--border-modal-style', `1px solid ${c.borderModal}`);
applyRootVar('--border-style', `1px solid ${c.border}`);
```

- `frontend/wailsjs/go/models.ts`: clase `ThemeColors` con `border` (los
  bindings se regeneran igual en el próximo `wails build`).

**3) UI de configuración:** vuelve la fila "Bordes generales" en
Personalización → Colores (ref `colorBorder`, `buildPersonalization`,
`trackRecents`, carga en `onMounted` y preview en vivo vía
`onPreviewColor('border', ...)`), junto a la ya existente "Bordes de modales":
así los dos bordes se configuran como cualquier otro color.

**Archivos tocados:**
- `web/index.css`
- `web/src/stores/ui.ts`
- `web/src/Layouts/Sections/Settings/PersonalizationSettings.vue`
- `internal/Config/Config.go`
- `frontend/wailsjs/go/models.ts`

## Por qué

- Los bordes no deben depender de los colores de fondo: cada borde es un color
  propio editable (independiente), persistido en la config como el resto.

## API afectada

- `Personalization.colors` vuelve a tener `border` (junto a `borderModal`).
  Las configs existentes sin `border` se sanean solas al default en el primer
  guardado.

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- En Configuración → Colores hay dos filas de bordes; al cambiar cualquiera,
  `launcher_config.json` guarda el nuevo color y los elementos con
  `--border-style` / `--border-modal-style` lo reflejan (la puesta en la
  interfaz la hace el usuario, como acordado).
