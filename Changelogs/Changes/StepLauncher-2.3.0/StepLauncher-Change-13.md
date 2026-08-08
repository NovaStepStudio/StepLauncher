# StepLauncher-Change-13: Eliminado "Bordes generales" y las variables `--color-*` duplicadas

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El color de la interfaz usaba dos juegos de variables CSS que valían lo mismo:
`--color-*` (duplicado) y `--background-*` / `--border-*` (el tema real). Se
elimina el juego duplicado y la personalización extra de borde general.

**1) Fuera `--color-*`:** se elimina el bloque de `index.css`:

```css
--color-sidebar: #0005;
--color-modal: #111;
--color-button: #111;
--color-border-modal: #494949;
--color-border: rgba(37, 37, 37, 0.3);
```

y su aplicación en `applyPersonalization()` (`stores/ui.ts`). Los consumidores
pasan a usar las variables existentes:

- `--color-modal` → `--background-modal-primray` (ColorField: panel, snippet,
  barra de alpha).
- `--color-border` → `--border-style` (bordes finos del ColorField).
- `--color-border-modal` → composición con `--background-modal-primray`
  (focus del snippet, item activo del menú de usuario en `App.vue`, hover de
  botones de cuentas en `AccountsView.vue`).

**2) Fuera "Bordes generales":** se elimina la fila de Personalización
("Bordes de tarjetas y elementos en general"), el ref `colorBorder`, su
inclusión en `buildPersonalization`, `trackRecents` y la carga en
`onMounted`. El borde general `--border-style` deja de sobrescribirse en
runtime y queda el de `index.css`, que deriva del color de la barra lateral
(`--background-sidebar`).

**3) Fuera `colors.border` de la config:** el campo `Border` se elimina del
struct `ThemeColors` en `internal/Config/Config.go` (defaults, sanitize y el
registro de recents) y de `ThemeColors`/`normalizePersonalization` en el
frontend. Las configs antiguas en disco con la clave `border` siguen
cargando (se ignora sin error). Se mantiene `colors.borderModal`
("Bordes de modales"), que sigue alimentando `--border-modal-style`.

**4) Fix del parseo de hex corto** (detallado en Error-7): `parseColor`
acepta `#RGB`/`#RGBA` para que los defaults `#0005`/`#111` no se corrompan a
`#000000` al tocar el ColorField.

**Archivos tocados:**
- `web/index.css`: bloque `--color-*` eliminado.
- `web/src/stores/ui.ts`: interfaz `ThemeColors` sin `border`,
  `normalizePersonalization` y `applyPersonalization` sin `--color-*`/`--border-style`.
- `web/src/Layouts/Sections/Settings/PersonalizationSettings.vue`: fila
  "Bordes generales", ref, build, recents y carga eliminados.
- `web/src/Layouts/Sections/Settings/ColorField.vue`: parseo de hex corto +
  estilos migrados a `--background-modal-primray`/`--border-style`.
- `web/src/App.vue` y `web/src/Modals/AccountsView.vue`: usos de
  `--color-border-modal` migrados.
- `internal/Config/Config.go`: campo `ThemeColors.Border` eliminado.
- `frontend/wailsjs/go/models.ts`: clase `ThemeColors` regenerada a mano
  (bindings se regeneran igual en el próximo `wails build`).

## Por qué

- El tema ya se expresaba con `--background-*` y `--border-*`; duplicarlo en
  `--color-*` era trabajo extra y fuente de divergencia. Un solo sistema de
  variables evita recrear valores que ya existen.
- El borde general estaba personalizable pero tenía un default que
  sobrescribía el diseño derivado de la barra lateral; con la fila fuera, el
  borde general sigue al color de la barra lateral, como en `index.css`.

## API afectada

- `Personalization.colors` pierde la clave `border` (JSON y bindings Go).
  `colors.borderModal` no cambia. Las configs existentes con `border` cargan
  sin error y se sanea a su valor por defecto en el siguiente guardado.

## Cómo verificar

- `go build ./...` → OK.
- `bun run build` (dentro de `frontend/`) → OK (type-check incluido).
- Abrir Configuración → Colores: ya no existe "Bordes generales"; abrir y
  cerrar los selectores con la config por defecto no cambia los colores
  (`#0005` sigue conservando su transparencia al persistir como `#00000054`;
  `#111` como `#111111`); los bordes del selector se ven con `--border-style`.