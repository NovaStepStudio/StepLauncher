# StepLauncher-Change-20: El ColorField de "Progreso de descarga" ya no se centra en preview (se queda en su sitio, como el de Modal)

- **Fecha**: 2026-08-04
- **Versión**: 2.3.0 (productVersion de `wails.json`)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El campo de color "Progreso de descarga" de Configuración → Personalización →
Colores llevaba `preview` (introducido en `StepLauncher-Change-7`/`-8` y usado
por la réplica de `StepLauncher-Change-19`): al arrastrar el color, el selector
se centraba en la pantalla y ocultaba el panel de ajustes.

Ahora ese campo se comporta como el de "Modal" (sin `preview`): el selector es
**quieto** durante el arrastre, no se mueve al centro ni se oculta el panel. Se
ha retirado la prop `preview` y el handler `@preview` del `ColorField` de
progreso en `PersonalizationSettings.vue`, dejándolo idéntico a los campos sin
preview (`colorModal`, `colorBorderModal`, colores de tipografía).

El resto de campos con `preview` (sidebar, botones, bordes generales) conservan
su comportamiento de centrado y preview en vivo.

## Archivos tocados

- `frontend/web/src/Layouts/Sections/Settings/PersonalizationSettings.vue`

## Por qué

El centrado en el preview del campo de progreso resultaba molesto: el selector
saltaba al centro de la pantalla mientras se arrastraba. Ahora se queda quieto
en su sitio, exactamente igual que el ColorField de "Modal".

## Cómo verificar

- `bun run build` (dentro de `frontend/`) → OK (`vue-tsc --build` + `vite build`).
- En Settings → Personalización → Colores → "Progreso de descarga": al abrir el
  selector y arrastrar, el campo permanece en su lugar dentro del panel, sin
  centrarse ni ocultar la configuración.
- Sidebar, Botones y Bordes generales siguen con su preview con centrado.