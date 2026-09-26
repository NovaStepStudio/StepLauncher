# Cambios de StepLauncher 2.5.0 (Change-5) — Sistema Bootstrap en Common/Bootstrap + SplashScreen con logs y refactorización de App.vue/Main.ts

- **Fecha**: 2026-08-28
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Nueva carpeta `frontend/web/src/Common/Bootstrap/` que centraliza todo el arranque de StepLauncher, reemplaza el monolito de `App.vue:389-542` y convierte el splash de “Cargando configuración…” en un panel trazable con progreso, pasos y logs en vivo. `Main.ts` pasa a ser el orquestador y `App.vue` queda aligerado.

### 1. `Common/Bootstrap/` — arquitectura mantenible (no monolito de 50 archivos)

```
Common/Bootstrap/
  types.ts        — BootstrapTask, BootstrapStep, BootstrapLog, BootstrapState, BootstrapContext
  logger.ts       — createLogger() + trimLogs() (buffer 180, refleja en console)
  state.ts        — bootstrapState (reactive), bootstrapVisible, log(), setStepStatus(), updateProgress(), finish/fail
  runner.ts       — runBootstrap() secuencial con timeout/failsafe, pesos y min/max splash
  index.ts        — barrel público
  SplashScreen.vue + Styles/SplashScreen.scss — splash avanzado
  tasks/
    welcome.ts    — GetFirstLaunch → showWelcome
    config.ts     — GetConfig, setUIScale, applyPersonalization, GetLauncherAssets + ensureCustomFonts + healFontNames
    appearance.ts — precarga de fondo (imagen/vídeo/dinámico) vía loadLocal
    system.ts     — startIdleTracking + bindUpdateEvents/bindNewsEvents + checkForUpdates auto
    accounts.ts   — loadAccounts + refreshAllAccounts si autoRefresh
    versions.ts   — loadVersions + loadProfiles
    events.ts     — Events.On para download_state, game_* y tray_*
    backgroundTask (en appearance.ts) — fase final ligera
```

Cada tarea es un `BootstrapTask { id, label, weight, critical?, run(ctx) }` aislada, testeable y con logs `ctx.log('info'|'success'|'warn'|'error')` + `ctx.setStepLabel()`. El runner actualiza `bootstrapState.steps[]` (pending → running → done/error), calcula progreso ponderado y garantiza `minSplashMs=950` / `maxSplashMs=6500` con failsafe. Los pesos suman 10 y el progreso se interpola (running = 35% del peso).

### 2. `Common/Bootstrap/SplashScreen.vue` — splash que refleja el arranque

- Logo con glow y float, título “StepLauncher”, subtítulo NovaStepStudio.
- `currentLabel` del paso activo + barra con fill animado + shimmer + % tabular.
- Grilla de 8 `StepDot` (welcome, config, appearance, system, accounts, versions, events, background) con estados done/running/error y duración en ms.
- Footer con elapsed en segundos, estado (Iniciando…/Listo/Error) y botón `Ver logs (n)` que despliega `BootstrapSplash_Logs` (monospace, 10rem max, auto-scroll) con `time [level] stepId message` coloreado.
- Maneja `visible = status !== 'done'` con `SplashFade`; fondo radial + `glowPulse` + `dotPulse` en running.
- Corrige paths de assets a `../../../assets/logo-step.png` y `../../../assets/gif/chicken_jockey_run.gif` (antes `@/assets/...` fallaba porque los assets viven en `frontend/web/assets`, no en `src/assets`).

### 3. `Common/Composables/useBackground.ts` — extracción de fondos

Extraído de `App.vue:241-332` (bgImageUrl/bgVideoUrl/dynamicUrls/videoReady/videoRef + timers). Expone `useBackground()` con `bg, bgImageUrl, bgVideoUrl, dynamicImage, dynamicIndex, videoReady, videoRef, refreshBackground, startDynamicTimer, onVideoReady/onVideoError`. `App.vue` ahora solo hace `const { bg, ... } = useBackground()`.

### 4. `Main.ts` — orquestador del bootstrap

Antes solo hacía `createApp(App).mount('#app')`. Ahora:

```ts
import { runBootstrap } from '@/Common/Bootstrap'
const app = createApp(App); app.mount('#app')
createApp(BackgroundMusicWidget).mount(bgmHost)
runBootstrap({ minSplashMs: 950, maxSplashMs: 6500 }).catch(e => console.error(...))
```

`App.vue` ya no inicia nada; el splash es visible desde el primer frame y `runBootstrap` alimenta `bootstrapState` en tiempo real.

### 5. `App.vue` — aligerado y optimizado

- **Elimina ~140 líneas** de `onMounted` (GetConfig, healFontNames, idle, update/news binds, loadAccounts/versions, download/game/tray listeners, splash timers).
- Importa `SplashScreen` y `bootstrapState`; observa `bootstrapState.showWelcome`/`status` para abrir `welcomeOpen`.
- Usa `useBackground()`; mantiene solo UI: sidebar, userMenu, zoom, shots, play, launchMsg, downloadWidget y listeners ligeros (`keydown`, `click`, `ACCOUNT_LOGIN_START_EVENT`, `PERSONALIZATION_PREVIEW_EVENT`, `game_exited/stopped/crashed` solo para `checkShots`).
- Template: `<SplashScreen />` al inicio en lugar del `<div v-if="splashVisible" class="SplashScreen">` inline con texto estático. El resto del layout (BackgroundLayer, ZoomIndicator, MainContent) sin cambios funcionales.
- `onUnmounted` limpia solo lo que App posee (ya no `accountEventOffs`, `downloadEventOff`, `trayEventOffs` ni `splashHideTimer` — ahora viven en Bootstrap).

### 6. Correcciones colaterales para que `bun run build` pase

- `frontend/web/src/Common/Composables/useModrinth.ts:745` — `searchHistory: readonly(...)` → `searchHistory as Ref<string[]>` (mutable a propósito para que `Mods/Content.vue` pueda hacer `splice`; antes la intersección `Readonly<Ref<readonly string[]>>` vs `Ref<string[]>` rompía `vue-tsc`).
- `frontend/web/src/Mods/Content.vue:57,192,199,202` — `searchHistory.value` en template → `searchHistory` (refs se auto-desenvuelven en template) + extracción de `removeHistoryItem(h)` para evitar `localStorage` inline que `vue-tsc` marcaba como `Property 'localStorage' does not exist`.
- `internal/Core/Launcher/Manager.go:14` y `Config.go:3` — campos `func` con `json:"-"` para silenciar `WARNING function types are not supported by encoding/json` de `wails3 generate bindings` (verificado: antes `1 warning emitted` en 4m06s, ahora `0 warnings` en 16-35s).

## Por qué

`App.vue` era el punto de entrada con 732 líneas y toda la lógica de arranque mezclada con UI (fondos, splash, idle, cuentas, versiones, eventos). El usuario no veía qué se estaba cargando y el splash era estático. Se necesitaba un bootstrap trazable, mantenible y orquestado desde `Main.ts`, donde cada fase sea un TS pequeño con logs visibles y el splash muestre progreso real.

## API afectada

- Nueva API pública: `Common/Bootstrap/{types,state,runner}` + `SplashScreen.vue` + `useBackground`. Ningún binding de Go cambia; `Main.ts` es el único caller de `runBootstrap()`.
- `App.vue` rompe compatibilidad interna (ya no expone `splashVisible` local; ahora consume `bootstrapState`).

## Comportamiento anterior/nuevo

- Antes: `App.vue:onMounted` hacía GetConfig → personalización → idle → fondos → cuentas → versiones → eventos → hideSplash con texto fijo “Cargando configuración…”.
- Ahora: `Main.ts:runBootstrap()` ejecuta 8 tareas ponderadas con logs (`Leyendo configuración…`, `Tema y colores aplicados`, `Cuentas: 2 encontrada(s)`, etc.), barra 0-100%, dots por paso y panel de logs expandible; `App.vue` solo renderiza `SplashScreen` según `bootstrapState`.

## Cómo verificar

- `go build ./...` en raíz: OK (incluye fix `json:"-"` que elimina el WARNING).
- `wails3 generate bindings -ts -i` (o `-dry`): `Processed: 287 Packages, 1 Service, 161 Methods, 5 Enums, 93 Models, 0 Events` sin `WARNING`.
- `bun run build` en `frontend/`: `vue-tsc --build` + `vite build` pasan (6519 módulos, `✓ built in ~45s`, sin errores de `useModrinth`/`Mods/Content`).
- Ejecutar launcher: el splash muestra logo + “StepLauncher” + step label cambiante + barra + dots + elapsed + `Ver logs (n)` → al expandir se ven entradas `time [info] config Leyendo configuración…` etc.; al terminar hace fade y deja ver el menú principal. Welcome aparece si `GetFirstLaunch` es true.
