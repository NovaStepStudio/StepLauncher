# Cambios de StepLauncher 2.5.0 (Step-16) — Botones del welcome sin gradiente y splash de carga dark sólido

- **Fecha**:   2026-09-24
- **Versión**: 2.5.0
- **Estado**:  implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

### 1. Botones del welcome en sólido, sin gradientes (`Welcome/Styles/Welcome.scss`)

- `.WelcomeModal_BtnPrimary`: el gradiente brillante (`button-primary 92% + blanco 8% → 62% + negro`) quedó en sólido `var(--background-button-primary)` con borde `#2e2e2e`; el hover sube a `brightness(1.35)` en vez de repetir el gradiente.
- `.WelcomeModal_PreviewPlay` (vista previa del botón jugar en el onboarding): sólido `var(--background-play-button)` con el mismo borde.
- `.WelcomeModal_Card` y su hover (tarjetas clicables de opciones): sólidos `#101010` / `#161616` en vez del velo blanco-negro en diagonal.
- `.WelcomeModal_PreviewThumb`: sólido `#1a1a1c`.
- No se tocaron los gradientes que no son botones: viñeta del fondo y relleno de la barra de progreso.

### 2. Pantalla de carga ("Cargando configuración") dark y opaca (`web/index.html`, `Common/Bootstrap/Styles/SplashScreen.scss`)

- El fondo radial con tinte azulado `#1e293b` pasó a sólido `#0a0a0a`, tanto en el splash estático del `index.html` como en el `SplashScreen.vue` del bootstrap.
- El panel de pasos (`BootstrapSplash_Steps`) pasó de `rgba(0, 0, 0, 0.25)` con blur a sólido `#101010` con borde `#1e1e1e`.

## Por qué

A pedido del owner: los botones con gradiente del welcome se veían horribles y la pantalla donde carga la config se veía clara/transparente; ahora todo es dark mode sólido.

## API afectada

Solo estilos del frontend. Sin cambios en el backend Go ni bindings.

## Comportamiento anterior/nuevo

- Antes: botones del welcome con brillo en diagonal y splash con fondo azulado en degradado y panel de pasos translúcido.
- Ahora: botones en color plano (jugar `#1f1f1f`, primario `#19191a` con los defaults actuales) y splash negro sólido `#0a0a0a` con panel `#101010`.

## Cómo verificar

- `bun run build` en `launcher/frontend` (pasa `vue-tsc --build` y `vite build`).
