# Cambios de StepLauncher 2.5.0 (Change-6) — Reorganización de Configuración para usuario promedio: Red, Descargas, Integridad y Almacenamiento + lenguaje simplificado

- **Fecha**: 2026-08-28
- **Versión**: 2.5.0 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

Reorganización completa del modal de Configuración para que un usuario promedio entienda cada opción sin jerga técnica, aplicando el feedback del owner sobre la distinción **automático vs manual** en integridad y la separación **Red vs Descargas**.

### 1. Cuatro secciones nuevas extraídas de `General.vue` (que pasó de 964 a 398 líneas) — renombradas a inglés por petición del owner

**`Settings/Sections/Network.vue` — Red** (antes `Red.vue`)
- Extraído de `General.vue:594-625`: `proxyEnabled`, `proxyHost`, `proxyPort`, `proxyUser`, `proxyPass` + `authVerify`.
- Lógica: `GetConfig` → `SetProxy` / `SetAuthVerify` (`Network.vue:6`).
- Descripciones simplificadas para promedio: “Comprobar cuenta al iniciar sesión — Verifica que tu cuenta sea válida antes de entrar” y “Usar proxy solo para Minecraft — Si tu internet necesita proxy, el juego lo usará. No afecta al launcher” + tip “Solo cambia esto si sabes lo que es un proxy”.
- Estilos: `@use '../Styles/General.scss'` (reusa `Ss*`). Sin emojis en textos.

**`Settings/Sections/Downloads.vue` — Descargas** (antes `Descargas.vue`)
- Extraído de `General.vue:559-584`: `concurrentDownloads`, `maxMbps`.
- Lógica: `GetConfig` → `SetConcurrentDownloads` / `SetMaxMbps` (`Downloads.vue:6`).
- Descripciones simplificadas: “Archivos a la vez — Cuántos archivos se bajan al mismo tiempo. Más es más rápido, pero usa más internet.” y “Límite de velocidad — Hasta qué velocidad puede descargar. 0 es sin límite”.
- Tip contextual sin emoji: “¿Los archivos se bajan mal? Activa la verificación automática en Integridad.” (enlace entre Descargas e Integridad sin duplicar el toggle).
- No incluye `verifyIntegrity` (queda solo en Integrity para evitar duplicación confusa).

**`Settings/Sections/Integrity.vue` — Integridad** (antes `Integridad.vue`, diseño propuesto por el owner)
- Extraído de `General.vue:55-142` + `General.vue:892-941`: `verifyIntegrity`, `integrityScope`, `integrityStatus`, polling, `StartIntegrityCheck`.
- **Estructura exacta del mock del owner (sin emojis):**
  ```
  Verificación automática
  ────────────────────────
  Revisar archivos al descargar
    (Si está activo, comprueba cada archivo al bajarlo. Más lento, pero evita que el juego falle.)

  Verificación manual
  ────────────────────
  Qué quieres revisar: Todo / Juego / Mundos
  [ Revisar ahora ]  + barra + fase
  ```
- `Todo/Juego/Mundos` mapean a `todo/global/instances` (`Integrity.vue:124`).
- Fases simplificadas: “Buscando qué tiene que estar”, “Bajando lo que falta”, “Comprobando que todo esté bien” en lugar de “Recorriendo JSON de versiones / Verificando SHA1”.
- Tips de contraste: “Esto es automático. No hace nada ahora, solo cuando descargas algo.” vs “Esto sí hace algo ahora.”
- Lógica: `GetConfig` para `verifyIntegrity`+`integritySector`, `SetVerifyIntegrity`, `SetIntegritySector`, `StartIntegrityCheck`, `IntegrityStatus` polling cada 500ms (`Integrity.vue:58`).

**`Settings/Sections/Storage.vue` — Almacenamiento** (antes `Almacenamiento.vue`)
- Extraído de `General.vue:166-258` + `627-730`: `dirInfo`, `dirMode`, `customPath`, `separateGameDir` + `cacheInfo`.
- Directorio: `GetDirectorySettings`, `GetSeparateGameDir`, `PickDirectory`, `SetDirectoryMode`, `RestartApp`, `SetSeparateGameDir` (`Almacenamiento.vue:6`).
- Caché: `GetCacheInfo`, `ClearAllCache`, `RefreshManifests` (`Almacenamiento.vue:10`).
- Descripciones simplificadas: “Dónde se guarda todo — Aquí se guardan tus mundos, versiones y cuentas.”, “Tipo de carpeta — Normal es lo recomendado. Minecraft usa tu carpeta oficial.”, “Tu carpeta — Elige dónde quieres guardar todo.”, “Estás usando: {{workDir}}”, “Guardar mundos aparte — Si está activo, tus mundos van en carpeta "game" separada. Más ordenado.”, “Qué hay guardado — {{cacheDetail}}”, “Limpiar — Borra archivos temporales para liberar espacio. No borra tus mundos.” + botones “Actualizar lista / Borrar”.
- Mantiene banner “¿Usar tu Minecraft?” y validación `dirChanged` + `RestartApp`.

### 2. `Settings/Sections/General.vue` — aligerado y simplificado

- **Eliminados**: grupo `Internet`, `Cache`, `Directorio del launcher`, `Integridad` + lógica asociada (`concurrentDownloads`, `maxMbps`, `authVerify`, `proxy*`, `dirInfo`, `cacheInfo`, `integrity*`, `dir*`, `saveDownloads`, `saveProxy`, `saveAuthVerify`, `saveMbps`, `saveVerifyIntegrity`, `saveDirectory`, `clearCache`, etc.) y su segundo `onMounted` duplicado.
- **Quedan 6 grupos simplificados** con lenguaje para promedio:
  - `Tamaño` (zoom): “Tamaño de todo — Haz el launcher más grande o pequeño. También puedes usar Ctrl + y Ctrl -.”
  - `Al jugar` (hideLauncher, richPresence, launchAfterInstall): “Esconder el launcher al jugar”, “Mostrar en Discord — Tus amigos ven que estás jugando.”, “Abrir el juego al terminar de instalar — Cuando termina una instalación, empieza a jugar sin tocar nada.”
  - `Efectos` (animations, blur, shadows, textShadow): “Movimiento — Animaciones suaves”, “Fondo borroso — Desenfoque detrás de las ventanas”, “Sombras”, “Letras con brillo — Un resplandor suave en el texto” + “Cuánto brillo”.
  - `Si no estás usando el launcher` (idle): “Cerrar ventanas si no lo usas — Si te alejas un rato, cierra las ventanas abiertas”, “Cuánto esperar”, “Revisar que todo siga igual — De vez en cuando comprueba que tus colores y letras sigan como los dejaste.”
  - `Restablecer` y `Actualizaciones`: textos acortados — “Volver todo como al inicio — Borra tus ajustes”, “Buscar al abrir — Mira si hay versión nueva”, “Probar ahora — [Buscar]”.
- Imports reducidos a `GetConfig`, `SetHideLauncher`, `SetRichPresenceEnabled`, `SetUIScale`, `UpdatePersonalization`, `ResetConfig`, `SetCheckForUpdatesOnStart`, `SetLaunchAfterInstall` (`General.vue:6`).
- De 964 a 398 líneas, sin `onActivated` de caché/integridad (ahora viven en sus secciones).

### 3. `Settings/Sections/Minecraft.vue` — lenguaje para promedio

- `Hardware` → `Gráficos`: “Usar opciones de gráfica — Si lo apagas, el juego usa lo que trae por defecto”, “Aceleración — Usa tu gráfica para que vaya más fluido”, “Qué gráfica usar — Automático elige la mejor”, “Modo de la gráfica — Rendimiento va más rápido, Calidad se ve mejor”, tip “Si el juego no abre, prueba apagando esto.” (`Minecraft.vue:163`).
- `Memoria`: “Memoria para Minecraft — Cuánta memoria le das al juego. Tienes {{totalRAM}} GB en total.” + tip “usa la mitad” (`Minecraft.vue:222`).
- `Java`: “Qué Java usar — Déjalo en Automático si no sabes qué es esto” con opciones “Automático / Elegir de mi PC / El que trae Minecraft / Ruta manual”, “Elige uno — Busca los Java que tienes y elige”, “Dónde está Java — Pega la ruta a javaw.exe” (`Minecraft.vue:246`).
- `Ventana`: “Tamaño al abrir — Ancho y alto de la ventana”, “Pantalla completa — Que el juego ocupe toda la pantalla” (`Minecraft.vue:296`).
- `Argumentos` → `Avanzado` con disclaimer “Solo toca esto si sabes lo que haces” + placeholders `Ej: -XX:+UseG1GC` (`Minecraft.vue:328`).
- `Avanzado` → `Más opciones`: “Jugar sin internet — Te deja entrar aunque no tengas conexión”, “Modo para PCs viejos — Si tu PC tiene muchos años y el juego falla”, “Guardar más info si falla — Útil si necesitas reportar un error” (`Minecraft.vue:353`).

### 4. `App.vue` — registro de secciones

- Nuevos imports: `IconWorld`, `IconShieldCheck`, `IconDatabase`, `NetworkSettings`, `DownloadsSettings`, `IntegritySettings`, `StorageSettings` (`App.vue:14` / `App.vue:22`).
- Imports renombrados a inglés: `Red.vue`→`Network.vue`, `Descargas.vue`→`Downloads.vue`, `Integridad.vue`→`Integrity.vue`, `Almacenamiento.vue`→`Storage.vue` por petición del owner (archivos en inglés).
- `settingsSections` pasa de 5 a 9, en orden pensado para promedio: `General` → `Personalización` → `Minecraft` → `Red` → `Descargas` → `Integridad` → `Almacenamiento` → `Cuentas` → `Acerca de` (`App.vue:175`).
- Eliminados emojis de todos los textos/descripciones por petición (ej. `Descargas.vue:77` `Integridad` sin escudo).

### 5. `TableConfig.md` — tablas rotas corregidas

- 4 tablas de 9 columnas (`§3.4 Personalización`, `§3.5 Idle`, `§3.6 Rich Presence`, `§3.7 Directorio`) tenían separador con 8 columnas (`|---|---|---|---|---|---|---:|---|` → 9 `|`). Corregidas a `|---|---|---|---|---|---|---|---:|---|` (10 `|` / 9 cols, con `Tipo` right-aligned) (`TableConfig.md:116,159,168,174`). Verificado con `check.py` — 20/20 bloques consistentes.

### 6. `Settings/Sections/About.vue` — bibliotecas y créditos

- Nuevo grupo `Bibliotecas` con botón `Ver`/`Ocultar` que despliega un menú colapsable (`showCredits`, transición `SsCollapse`) y modal de detalle (`SsLibModalOverlay`, `useOverlayEscape`, transición `SsModalFade`) (`About.vue:152`).
- **Mayor aporte al proyecto — Wails 3** destacado con borde `color-success` y mensaje pedido por el owner: “Gracias Wails por crear un excelente framework. Esperamos que Wails 3 siga creciendo asi, muchas gracias Wails 3 sin ti no pudimos hacer StepLauncher.” (`About.vue:180`).
- **Librerías de terceros** (5): `music-metadata` (metadatos audio), `marked` (Markdown→HTML), `typescript` (tipado frontend), `sass-embedded` (SCSS→CSS), `vue` (framework UI) — cada una con icono, descripción y botón `Abrir sitio` a su URL (`About.vue:188` + `thirdPartyLibs`).
- **Componentes internos del launcher** (10) que merecen crédito: `Launcher` (motor), `Downloader` (verificación/reintentos), `ModLoader` (Fabric/Forge/NeoForge/Quilt/LegacyFabric), `Accounts` (offline/Authlib), `Assets` (fondos/música/fuentes), `Tray`, `RichPresence`, `Updater`, `Config`, `Platform` — con `path` `internal/...` (`About.vue:201` + `internalLibs`). Cada item abre el mismo modal con `path` y `desc`.
- Estilos nuevos en `About.vue` scoped: `SsCreditsCollapse`, `SsWailsHighlight`, `SsCreditsList`, `SsCreditItem`, `SsLibModalOverlay`, `SsLibModal` etc., sin emojis, respetando `var(--color-*)` y `Ss*`.
- Dependencias verificadas en `go.mod:6` (wails/v3, go-winio, x/sys) y `frontend/package.json:18` (marked, music-metadata, vue, typescript, sass-embedded, @wailsio/runtime).

## Por qué

- `General.vue` era un cajón de sastre inmantenible y confuso para un usuario promedio (9 grupos heterogéneos, 960 líneas, scroll > 4000px). El owner pidió distinguir **verificación automática (al descargar) vs manual (reparar ahora)** en un mismo sitio con contraste visual inmediato, y separar **Red (proxy/auth) vs Descargas (concurrencia/límite)** por cohesión conceptual.
- Las descripciones técnicas (“Comprueba el SHA1…”, “Enruta el tráfico…”, “Verifica SHA1 y tamaño…”) no aportan a un usuario promedio; se reemplazan por frases cortas y ejemplos cotidianos sin jerga, siguiendo la instrucción “sin dar info detallada como 'Hace esta funcion'”.

## API afectada

- Ningún binding nuevo. Se reutilizan: `GetConfig`, `SetProxy`, `SetAuthVerify`, `SetConcurrentDownloads`, `SetMaxMbps`, `SetVerifyIntegrity`, `SetIntegritySector`, `StartIntegrityCheck`, `IntegrityStatus`, `GetDirectorySettings`, `GetSeparateGameDir`, `PickDirectory`, `SetDirectoryMode`, `RestartApp`, `SetSeparateGameDir`, `GetCacheInfo`, `ClearAllCache`, `RefreshManifests`, `SetHideLauncher`, `SetRichPresenceEnabled`, `SetUIScale`, `UpdatePersonalization`, `ResetConfig`, `SetCheckForUpdatesOnStart`, `SetLaunchAfterInstall`, `DetectJavaInstallations`, `TotalRAMGB`, `SetMaxRAM`, `GetMinecraftConfig`, `UpdateMinecraftConfig` (todos ya existían).

## Comportamiento anterior/nuevo

- Antes: todo en `General` (Internet con proxy + descargas + auth, Cache, Directorio, Integridad con scope+botón mezclados, Comportamiento con verify toggle suelto). Descripciones técnicas. 5 secciones en sidebar.
- Ahora: `General` solo comportamiento + tamaño + efectos + inactividad; `Red` solo red; `Descargas` solo velocidad; `Integridad` con dos sub-secciones claramente separadas (automático con tip “no hace nada ahora” vs manual con “sí hace algo ahora” + scope Todo/Juego/Mundos + barra); `Almacenamiento` con directorio + caché. 9 secciones en sidebar ordenadas para flujo promedio. Descripciones cortas y amigables. Misma persistencia, mismos bindings.

## Cómo verificar

- `go build ./...` en raíz: OK (sin output).
- `bun run build` en `frontend`: `vue-tsc --build` + `vite build` OK (6532 módulos, `✓ built in ~54s`, sin errores de tipos).
- Abrir Configuración → verificar 9 secciones con iconos `IconHome` / `IconPalette` / `IconCpu` / `IconWorld` / `IconDownload` / `IconShieldCheck` / `IconDatabase` / `IconUsers` / `IconInfoCircle`; probar toggles en `Red` (proxy), `Descargas` (concurrent/maxMbps), `Integridad` (verificar automático checkbox + scope + Revisar ahora con barra), `Almacenamiento` (cambiar modo, ver caché, limpiar).
- `TableConfig.md`: abrir en preview markdown — 4 tablas de §3.4-§3.7 ya no aparecen rotas.
