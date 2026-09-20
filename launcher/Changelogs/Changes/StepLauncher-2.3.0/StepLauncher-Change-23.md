# StepLauncher-Change-23: Rediseño del panel de versiones (identidad propia frente al panel de cuentas)

- **Fecha**: 2026-08-05
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El panel de versiones y perfiles deja de parecer un clon del panel de cuentas y
adquiere una **identidad visual propia** (acento violeta en lugar del azul) y
mejora la legibilidad de los tipos de versión.

- **Acento violeta** (`$vers: #a974ff`) en `frontend/web/src/Modals/VersionsView.vue`,
  que sustituye al azul heredado (`$accent`) en:
  - selección de versión en la sidebar `.Vers_Item.active` (borde, gradiente y
    barra izquierda),
  - botones laterales `.Vers_SideBtn.active` y su `inset` de marcado,
  - botón editar `.Vers_Tool` (hover) y badge `.Vers_ItemBadge`,
  - cabecera del modal en `frontend/web/src/Modals/VersionsModal.vue`:
    banda de gradiente violeta (`::before`), motivo decorativo de "pilas de
    versiones" (`::after`, tres barras apiladas a la derecha) e icono con tinte
    violeta; fondo del diálogo con tinte violeta (`rgba(18, 10, 30, 0.42)`).
- **Tipos de versión con color propio**: nueva función `typeHue(type)` que
  aplica la clase `Vers_TypeHue_IsRelease` (verde), `IsSnapshot` (dorado),
  `IsBeta` (naranja), `IsAlpha` (rojo) o `IsOther` (gris) al badge de cada
  versión, con estilos scoped en `VersionsView.vue`. El badge se distingue así
  de los paneles de cuentas, que conservan el azul.

## Por qué

Visualmente los paneles de cuentas y de versiones eran casi idénticos, lo que
podía confundir al usuario; además no había forma rápida de saber el tipo de
versión (release/snapshot/beta/alpha) de un vistazo.

## API afectada

- Ninguna (cambios puramente visuales; `typeHue()` es local al componente).

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- Abrir el modal de versiones: el panel ahora se ve violeta y distinto del de
  cuentas (azul); cada tipo de versión muestra su badge con su color; el perfil
  seleccionado conserva el marcado lateral violeta.
- Cambiar color de personalización: los acentos del launcher (sidebar, botones)
  siguen respetando las variables CSS; el violeto del panel de versiones es
  intencionalmente fijo para mantener su identidad.