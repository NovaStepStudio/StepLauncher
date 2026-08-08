# AGENTS.md

Launcher de Minecraft (Wails v2 + Vue 3 + TS). Frontend en `frontend/web/src`, backend Go en `internal/` con bindings finos en `app.go` que delegan a `internal/Handlers` (motor en `internal/Handlers/Engine`, integrado en el programa).

## Versiones y documentación

- Go **1.26.4** (go.mod) y Wails **v2.13.0** (go.mod).
- Cualquier modificación que involucre a Wails (bindings, runtime, ventanas, eventos) requiere leer antes la documentación oficial: https://v3.wails.io/quick-start/why-wails/

## Comandos (Windows / PowerShell)

- **bun es el gestor obligatorio**: usar `bun install` y `bun run build` dentro de `frontend/` (el build incluye el `vue-tsc` type-check).
- Verificación Go: `go build ./...` en la raíz.
- Build completo: `wails build` en la raíz → `build/bin/StepLauncher.exe`; regenera los bindings `frontend/wailsjs` y embebe el dist.
- Dev en vivo: `wails dev`.

## Arquitectura (lo que no es obvio)

- **Config**: `internal/Config` (`Manager` con mutex + `sanitize`). Persiste en `%APPDATA%\.StepLauncher\launcher_config.json`. `UpdatePersonalization` reemplaza todo el objeto preservando `UIScale` y el historial de colores.
- **Assets personalizables**: `internal/Core/Assets` gestiona `launcher_assets.json` (rutas relativas `launcher/...`; `sanitize` normaliza y garantiza placeholders).
- **El sistema de instancias (`internal/Core/Launcher` y el resto de `internal/Core/`) es OPCIONAL**: es una implementación adicional, NO es el núcleo del launcher. Nada debe asumirlo como requisito ni el flujo principal depender de él.

## Uso de IA y agentes (obligatorio)

StepLauncher es un proyecto grande (más de 80 archivos Go interconectados, algunos de más de 1000 líneas) y sus sistemas están fuertemente acoplados (handlers, bindings de Wails, config, personalización, instancias, descargas, cuentas, noticias). Debido a ese tamaño y acoplamiento, los agentes de IA se utilizan como herramienta de análisis, auditoría y asistencia durante el desarrollo: pueden analizar el código, rastrear dependencias, diagnosticar bugs y colaborar en la implementación de cambios. La IA es una herramienta de desarrollo y análisis, no una autoridad absoluta: el resultado de su trabajo debe verificarse igual que cualquier otro cambio.

### Cómo trabaja un agente en este repositorio

- Puede analizar y modificar código cuando la tarea lo requiere; no está limitado a leer.
- Antes de modificar, debe comprender la arquitectura existente (ver sección anterior) y consultar `Changelogs/` si el área afectada tiene historial (ver sección "Registro de errores y cambios").
- Puede usar `Changelogs/` como contexto histórico: bugs ya diagnosticados, causas raíz conocidas, soluciones aplicadas, reglas aprendidas y regresiones posibles.
- No debe asumir que una implementación aislada representa toda la arquitectura: existen muchos subsistemas conectados (p. ej. `internal/Core/` es opcional, pero otros sistemas forman parte del flujo principal).
- Antes de modificar APIs, handlers, bindings o sistemas compartidos, debe rastrear las conexiones entre componentes: revisar callers/callees y dependencias antes de cambiar funciones públicas o estructuras compartidas.
- Debe considerar que el backend tiene más de 80 archivos Go interdependientes: un cambio local puede afectar módulos lejanos.
- No debe inventar comportamiento, APIs, archivos, dependencias ni reglas que no existan en el repositorio.
- Sin evidencia suficiente, debe inspeccionar el código antes de afirmar una causa.
- Debe distinguir siempre entre **hechos observados**, **hipótesis** y **conclusiones verificadas**.

### Flujo de trabajo obligatorio de un agente

1. Inspeccionar el repositorio (código, estructura, dependencias).
2. Consultar documentación e historial cuando corresponda (`AGENTS.md`, `Changelogs/`, documentación oficial de Wails).
3. Analizar dependencias y conexiones del área afectada.
4. Implementar el cambio siguiendo las reglas de este archivo.
5. Verificarlo (build y type-check de Go y frontend, según corresponda).
6. Documentar errores o cambios relevantes en `Changelogs/` (ver sección "Registro de errores y cambios").
7. Reportar incertidumbres si existen (hipótesis sin verificar, comportamiento no confirmado).

Usar IA no elimina la obligación de verificar el resultado: todo cambio debe comprobarse antes de darse por terminado.

## Concurrencia y deadlocks (obligatorio)

- **Nunca bloquear el hilo principal ni los bindings de Wails**: las funciones expuestas en `app.go`/bindings deben ser rápidas. Cualquier trabajo lento (I/O, red, procesamiento) DEBE ejecutarse en una goroutine propia y devolver inmediatamente (callback/evento o `chan`), nunca por un `sync.WaitGroup` o `chan` que se espere a la espera de la misma goroutine.
- **Orden de adquisición y tiempos de bloqueo**: usar siempre `sync.RWMutex` en vez de `Lock()` global cuando el recurso se lee mucho. Cuando se necesiten varios mutex, respetar SIEMPRE el mismo orden de adquisición en todo el código para evitar interbloqueos. Mantener el `Lock()` el mínimo tiempo posible y NUNCA llamar a funciones que a su vez adquieran el mismo mutex (evita `deadlock` por reentrada y `sync` `self-contained`).
- **No bloquear el generador de bindings de Wails**: el generador de bindings (`wails generate bindings` / `wails build`/`wails dev`) no debe verse afectado por locks cuyo ciclo de vida dependa de la UI o del runtime del frontend. Mantener el bloqueo dentro del `Manager` y nunca exponer una operación que se quede esperando indefinidamente (sin `defer m.Unlock()`, sin `close(chan)` olvidado, etc.).
- **Siempre `defer`** para desbloquear mutex y cerrar canales tras su uso. Antes de bloquearse con `<-chan` o con `Wait`, verificar que hay una vía garantizada de cierre o `close` del canal iniciada por otra goroutine.
- Cualquier refactor que toque `internal/Config`, `internal/Core/` o los handlers DEBE revisar mutex, canales y `WaitGroup` para no introducir deadlocks (especialmente en el flujo que dispara el generador de bindings).

### Caso documentado: self-deadlock de RWMutex en Config (agosto 2026)

- **Qué pasó**: `wails build`/`wails dev` quedaban colgados en "Generating bindings" con la app a 0% CPU. `Save()` en `internal/Config/Config.go` hacía `m.mu.Lock()` y, **sin liberar**, llamaba a `m.logf()`, que intenta `m.mu.RLock()` sobre el mismo mutex. `sync.RWMutex` **no es reentrante**: un `RLock()` mientras la misma goroutine posee el `Lock()` se bloquea para siempre (espera circular: el writer que debería liberar es la misma goroutine parqueada).
- **Cadena que lo disparaba**: `main()` → `NewApp()` → `Handlers.NewApp()` → `Config.NewManager()` → `load()` → `Save()`. Todo corre en `main()` **antes** de `wails.Run()`, que es donde Wails genera los bindings → el proceso nunca llegaba a generarlos. Cualquier deadlock en la construcción de `NewApp()` (no solo en Config) tiene este mismo síntoma.
- **Regla aprendida**: ninguna función que adquiera un lock puede llamar a otra que adquiera el **mismo** lock. Auditarlo así: por cada `m.mu.Lock()`, listar las llamadas que hace el cuerpo; si alguna vuelve a tocar `m.mu` (aunque sea `RLock`, y aunque sea con `defer`), romper el lock antes o diferir la llamada (p. ej. `m.logf(...)` tras el `Unlock()`, o capturar el mensaje dentro del lock y loguearlo fuera). Los métodos de `Config` usan ese patrón: mutar bajo `Lock()`, `Unlock()` y loguear después.
- **Cómo detectarlo si vuelve a pasar**: si el build/dev cuelga en bindings con CPU 0% y sin goroutines ejecutables, revisar la cadena de construcción (main → NewApp → Handlers → NewManager/...) buscando `Lock()` seguido de llamadas que lockeen (los `logf`/loggers son los sospechosos habituales). El deadlock se reproduce también ejecutando el binario con `WailsGenerateBindings=true`.

## Registro de errores y cambios: carpeta `Changelogs/` (obligatorio)

- **La carpeta `Changelogs/` es el registro oficial de TODOS los errores y cambios del launcher** (backend Go, frontend Vue, config, builds). Antes de diagnosticar o modificar algo, consultarla: contiene el historial completo del repositorio (qué pasó, por qué, cómo se arregló).
- **Separación por tipo en subcarpetas en inglés**: `Changelogs/Errors/` (bugs y problemas de la app), `Changelogs/Changes/` (nuevas funcionalidades y mejoras) y `Changelogs/Releases/` (notas por versión). Los nombres de carpetas y archivos son en inglés, pero el contenido de los archivos de `Changelogs/` se escribe en **español**.
- **Cada versión tiene sus propias carpetas**: `Changelogs/Errors/StepLauncher-X.Y.Z/`, `Changelogs/Changes/StepLauncher-X.Y.Z/` y `Changelogs/Releases/StepLauncher-X.Y.Z/` (ver `Changelogs/README.md` para el flujo completo y el `news.json`).
- **Los `index.json`** (`Changelogs/Errors/index.json`, `Changelogs/Changes/index.json`, `Changelogs/Releases/index.json`) son el punto de entrada de todas las entradas y la base del centro de noticias: cada entrada nueva se registra en el suyo con **rutas SIEMPRE relativas a la ubicación del propio `index.json`** (p. ej. `./StepLauncher-2.3.0/news.json` dentro de `Releases/index.json`), nunca absolutas ni partiendo de `Changelogs/`. Errors y Changes se agrupan por versión (`versions` con `version` + rutas a los MD); Releases usa `latest` + `content`. Los tres índices se regeneran **SOLO al publicar la release** con `Changelogs/generate_indexes.ps1` (nunca al crear un error o cambio individual).
- Convención de nombres (numeración `N` secuencial por tipo Y por versión, **reinicia en 1 en cada carpeta de versión**): `StepLauncher-Error-N.md` en `Errors/StepLauncher-X.Y.Z/`, `StepLauncher-Change-N.md` en `Changes/StepLauncher-X.Y.Z/`, `StepLauncher-Release-X.Y.Z.md` + `news.json` en `Releases/StepLauncher-X.Y.Z/`.
- **Obligatorio documentar**: al corregir un error o terminar un cambio relevante, crear el archivo correspondiente en la carpeta de la versión en curso con su numeración siguiente. El `Changelogs/Errors/StepLauncher-2.3.0/StepLauncher-Error-1.md` documenta el self-deadlock de `sync.RWMutex` en `Config.Save()` (ver reglas en la sección de Concurrencia).
- Al terminar una corrección de bug: el error debe quedar documentado en `Changelogs/` ANTES de dar la tarea por terminada.

## Resolución de alias de import (vite.config.ts) (obligatorio)

- **Los `alias` de `frontend/vite.config.ts` definen TODOS los resolucivos de import permitidos en el frontend**: usarlos SIEMPRE en lugar de rutas relativas largas del estilo `../../../../wailsjs/go/main/models.ts`.
- **Solo se puede usar cualquier alias/path que esté explícitamente expuesto/definido en el bloque `resolve.alias` de `frontend/vite.config.ts`** (p. ej. `@wailsjs` → `./wailsjs`, `@` → `./web/src`). No inventar ni usar atajos que no estén declarados ahí; si un path no está, primero añadirlo al `vite.config.ts` y luego usarlo.
- **Añadir nuevos alias cuando algo se reutiliza mucho**: si se detecta un import o un fragmento de código que se repite bastante (p. ej. rutas hacia `wailsjs`, `src`, subcarpetas de `src`, utilidades, etc.), crear un nuevo alias en `resolve.alias` de `vite.config.ts` y usarlo en el frontend. El alias debe nombrarse con el prefijo `@` (p. ej. `@wailsjs`) y apuntar a la carpeta correcta con `fileURLToPath(new URL(...))` siguiendo el patrón existente.
- No romper los alias existentes: cuando se añada uno nuevo, respetar y conservar los que ya están definidos.

## Estilo (obligatorio)

- Todo panel, modal, tipografía o color nuevo DEBE respetar las variables CSS existentes. Se aplican en runtime desde `stores/ui.ts` → `applyPersonalization` y los componentes las consumen con `var(--...)` (p. ej. `--color-*`, `--background-*`, `--border-style`, `--border-modal-style`, `--font-primary`, `--font-secundary`). Conservar los nombres actuales aunque tengan erratas (`--background-modal-primary`, `--font-secundary`).
- Si un nuevo panel necesita estilo propio: registrar el panel (secciones en `App.vue`, `settingsSections`) y AÑADIR las variables de estilo necesarias junto con su registro; no inventar estilos sueltos fuera del sistema de variables.
- Seguir el estilo visual actual: SCSS con anidamiento, estilos `scoped`, clases prefijadas por sección (`.Ss*` en Ajustes, prefijo del componente para el resto).

## Convenciones

- **GIT**: el agente NO tiene permiso para ejecutar comandos de git que modifiquen el repositorio (commit, push, branch, tag, revert, etc.). Si una tarea requiere usar git, primero debe pedir permiso al usuario; este le dirá si puede o no.
- **ELIMINACIÓN**: el agente NO puede ejecutar comandos de eliminación de archivos o carpetas (del/rm/Remove-Item, etc.) sin pedir permiso primero. Están PROHIBIDAS las eliminaciones dentro de `internal/` y `frontend/` (incluido `frontend/web`); si hace falta eliminar un archivo concreto o una carpeta completa, debe pedir permiso obligatorio al usuario antes de ejecutar nada.
- Comentarios y mensajes en **español**.
- **No escribir tests** salvo que se pidan explícitamente.
- Bloques de `onMounted`/carga: `try/catch` independientes por paso; un fallo no debe matar los siguientes.
- Al diagnosticar bugs: no tratar los archivos de `%APPDATA%\.StepLauncher` como verdad; la app debe funcionar con cualquier estado en disco.
- El usuario edita los mismos archivos en paralelo con frecuencia: releer el archivo antes de cada edit.