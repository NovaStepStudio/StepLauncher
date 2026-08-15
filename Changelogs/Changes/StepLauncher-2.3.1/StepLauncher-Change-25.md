# Changes/StepLauncher-2.3.1/StepLauncher-Change-25.md

- **Fecha**: 2026-08-13
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (build OK).

## Qué cambió

### 1. Rediseño de la pantalla de bienvenida a 3 pasos

`frontend/web/src/Modals/WelcomeModal.vue` + `Styles/Modals/WelcomeModal.scss`:

1. **Bienvenida**: logo, badge, descripción, indicador de progreso de 3 pasos, dos tarjetas de acción ("Personalizar el launcher" — destacada como Recomendado — y "Configurar mi cuenta" — marcada Obligatoria) y fila de features del launcher (Jugar, Instancias, Skins, Personalizable).
2. **Configurar el launcher** (nuevo): selector de 7 paletas de acento (Default + 6 colores) que aplican `colors.buttonPrimary/tag/progress` en vivo vía `applyPersonalization` y persisten con `UpdatePersonalization`; 3 toggles propios (animaciones, desenfoque, sombras) que modifican y persisten las mismas flags del store `Ui`.
3. **Cuenta** (obligatoria): se eliminó el botón "Omitir" — ya no existe forma de saltar la cuenta dentro del onboarding (sin X, sin ESC, sin skip). El botón de cierre del formulario es "Volver" al paso anterior (`backStep`).
4. **Botones e inputs propios**: se eliminaron las clases compartidas `SsBtn`/`SsIn`/`SsInW` del welcome (y el `@use` de `Shared/Settings.scss`); todo usa clases exclusivas `WelcomeModal_Btn(_Primary/_Ghost)`, `WelcomeModal_Input`, `WelcomeModal_Toggle`, `WelcomeModal_Card*`, `WelcomeModal_Palette*`.

### 2. Animación de transición entre pasos (slide espejo + cambio de fondo)

- El contenido (panel derecho) se desliza con `<Transition name="WelcomeSlide" mode="out-in">`: el paso anterior sale hacia la izquierda (`translateX(-80px)`) y el nuevo entra desde la derecha (`translateX(80px)`), con easing `cubic-bezier(0.2,0.7,0.3,1)`.
- El stage (panel izquierdo) reemplaza el `::after` estático por una `<img>` con `<Transition name="WelcomeBg" mode="out-in">`: el fondo sale hacia la derecha y el nuevo entra desde la izquierda. El fondo cambia según el paso: `bg-welcome.webp` (bienvenida) → `bg-welcome-2.webp` (personalizar y cuenta). El gradiente/blur pasó a `.WelcomeModal_StageShade` (no animado, para que no parpadee).
- La gallina (`chicken_jockey_run.gif`) ahora vive dentro del stage (`.WelcomeModal_StageChicken`, con rebote `WelcomeChickenBounce`) mientras dura el cambio de paso; `goTo()` usa doble timer (350ms + 420ms) para que la gallina cubra toda la animación out-in.
- `.WelcomeModal_Content` fija `min-height: 27rem` para evitar saltos verticales entre pasos.

### 3. El menú principal se oculta con el welcome

`frontend/web/src/App.vue`: `showWelcome` se añadió al computed `mainMenuHidden`, igual que instancias/versiones/cuentas/noticias: al abrir la bienvenida el `MainContent` queda `opacity: 0`, `blur(8px)` y sin `pointer-events` (`Styles/App/App.scss` `.menuHidden`).

### 4. Vista previa del launcher en el panel izquierdo (paso "Configurar el launcher")

`WelcomeModal.vue` + `WelcomeModal.scss`: al entrar al paso de personalización, el stage muestra un mockup en miniatura del launcher (`.WelcomeModal_Preview*`): sidebar con items, versión seleccionada con badge, botón JUGAR, barra de progreso y mensaje de éxito. Todo construido exclusivamente con variables CSS (`var(--background-button-primary)`, `var(--color-tag)`, `var(--progress-color)`, `var(--color-success)`, `var(--background-modal-primary)`, `var(--border-style)`, etc.), por lo que al pulsar una paleta (o Default) el mockup se repinta al instante mostrando cómo quedará el launcher. La barra de progreso se re-anima al cambiar de paleta (`:key="activePaletteName"`) y el nombre de la paleta activa se muestra en la cabecera del mockup. Entra con fade + scale (`WelcomePreview`).

### 5. Paso de cuenta: fondo `bg-welcome-3.webp` + decoraciones

- El fondo del stage ahora se asigna por paso: `bg-welcome.webp` (bienvenida) → `bg-welcome-2.webp` (personalizar) → `bg-welcome-3.webp` (cuenta).
- En el paso de cuenta, el stage muestra las decoraciones `assets/decorations/having-fun.webp` y `assets/decorations/steve_and_alex.png` (ambas recortes de personajes de Minecraft con canal alfa). Disposición final: `having-fun.webp` centrada en el stage vía flex; `steve_and_alex.png` con `position: absolute`, centrada horizontalmente (`left: 50%` + `translateX(-50%)`) y anclada abajo (`bottom: 0.6rem`). Ninguna tiene bordes ni marcos ni `border-radius`, se renderizan a resolución nativa con suavizado normal (sin `image-rendering: pixelated`) y solo llevan un `drop-shadow` suave para integrarse con el fondo.

## Por qué

El usuario pidió una bienvenida más rica que no pierda la estética principal: que el primer destino sea "Personalizar el launcher" (no la cuenta directa), que la cuenta siga siendo obligatoria sin posibilidad de omitirla, que el contenido trasero se oculte como hacen los paneles de instancias/versiones y que las transiciones sean animadas con cambio de fondo y deslizamiento lateral.

## API afectada

- Frontend: se eliminó el uso de `SsBtn`/`SsIn` en el welcome (sin API de backend afectada). Reutiliza `applyPersonalization`/`personalization` de `Stores/Ui.ts` y `UpdatePersonalization` del binding existente.
- Assets: `assets/background/bg-welcome-2.webp` ahora se importa directamente desde el componente (antes no se usaba en ningún lado).

## Comportamiento anterior/nuevo

- **Antes**: bienvenida de 2 pasos (welcome → cuenta) con botón "Omitir"; fondo estático `bg-welcome.webp`; el menú principal quedaba visible detrás del modal.
- **Ahora**: bienvenida de 3 pasos con tarjetas de acción, personalización real (paletas + toggles persistidos), cuenta obligatoria sin omitir, transición slide espejo con cambio de fondo animado (`bg-welcome` → `bg-welcome-2` → `bg-welcome-3`), menú principal oculto, vista previa en vivo del launcher mientras se eligen colores y decoraciones `having-fun.webp` + `steve_and_alex.png` en el paso de cuenta.

## Cómo verificar

- `bun run build` en `frontend/`: type-check y vite build OK; `bg-welcome.webp` y `bg-welcome-2.webp` hasheados en `dist/assets/img/`.
- `go build ./...` en la raíz: OK.
- Manual: primer inicio → el menú principal desaparece; al pulsar "Personalizar el launcher" el contenido se desliza a la izquierda mientras el fondo cambia a `bg-welcome-2.webp` con la gallina rebotando; las paletas (incluida "Default") cambian los colores del launcher y persisten; el paso de cuenta no tiene forma de omitirse; al crear la cuenta se cierra y no reaparece.