# Cambios de StepLauncher 2.4.1 (Change-2) — Configuración avanzada y personalización de instancias

- **Fecha**: 2026-08-16
- **Versión**: 2.4.1 (en desarrollo)
- **Estado**: implementado y verificado
- **Release**: en desarrollo — aún no mencionado en ninguna release.

## Qué cambió

Tanda de mejoras de instancia pedidas por el usuario: opción para desactivar el GC en el selector, argumentos JVM/game propios, logs detallados, verificación de integridad propia (que bloquea el uso de la instancia durante el proceso), RAM por defecto de 2 GB cuando la PC lo permite, backups automáticos comprimidos, y personalización con etiquetas y descripción de hasta 512 caracteres.

### 1. Desactivar el GC en el selector

En `internal/Core/Launcher/Helpers/System.go`, `GCFlags` ahora tiene un case explícito `"none"` (sin flags de GC, el JVM usa su recolector por defecto). En `Instances/Settings.vue` el select de presets incluye la opción "Desactivado" (`GC_PRESETS` + `'none'`), que se guarda como `gcPreset: "none"` y ya no se trata como "sin configurar".

### 2. Argumentos JVM y del juego propios por instancia

- `InstanceLaunchConfig` (interno) ya tenía `javaArgs`/`gameArgs`; el mapeo de `Instance/Launch.go` los aplicaba con semántica de reemplazo (los de la instancia sustituyen a los globales del launcher).
- `Instance/Manager.go` (`UpdateConfig`): antes estos campos NO se persistían al guardar la config; ahora se hace merge de `JavaArgs`/`GameArgs` (si `cfg.JavaArgs != nil`, lo que permite vaciarlos y volver a heredar los globales).
- `Instances/Settings.vue`: dos textareas (JVM y juego) que dividen la línea en tokens respetando comillas (misma lógica que `splitArgs` del backend), y se envían siempre (aunque vacíos).

### 3. Logs detallados por instancia

- `Core/Launcher/Advconfig.go`: `AdvancedConfig` gana el campo `DetailedLogs bool json:"detailedLogs"`.
- `Instance/Launch.go`: el flag `detailedLogs` de la instancia se mapea a `adv.DetailedLogs`.
- `Core/Launcher/Launcher.go` (`buildGameArgs`): si `DetailedLogs` está activo y no hay un `LogLevel` explícito, se añade `--log-level debug` al juego para un registro más verboso.
- `Instances/Settings.vue`: switch "Logs detallados".

### 4. Verificación de integridad propia de la instancia (bloqueante)

- Nuevo struct `InstanceVerifyProgress` (`Instance/Types.go`): `state` (`idle|verifying|done|error|cancelled`), `phase`, `version` en curso, `percent`, `found` (versiones comprobadas), `issues` y `error`.
- `Instance/Verify.go` reescrito: la lógica síncrona existente se extrajo a `VerifySingleVersion`/`verifyDownloaderConfig`/`verifyVersion`; la nueva `StartInstanceVerify` lanza la verificación completa de la instancia en una goroutine con cancelación (`CancelInstanceVerify`) y progreso consultable por polling (`InstanceVerifyStatus`). Durante el proceso la instancia queda bloqueada: `LaunchInstance`, `AddVersion`, `InstallModLoader`, `UpdateConfig`, `UpdateMetadata`, `Delete` y `Clone` rechazan la operación con el helper `assertUsable` (también se rechaza arrancar si hay una descarga activa o el juego corriendo).
- Bindings nuevos en `Handlers/Engine/Instance.go` + wrappers en `app.go`: `StartInstanceVerify`, `CancelInstanceVerify`, `InstanceVerifyStatus`, `CreateInstanceBackup`.
- `Instances/Store.ts`: `verifyStates` por instancia, `startInstanceVerify`/`cancelInstanceVerify`/`refreshInstanceVerify`, y `isInstanceBusy` ahora también considera el estado `verifying` (deshabilita Jugar/Descargar/Editar/Configurar en la UI).
- `Instances/Settings.vue`: botón "Verificar", barra de progreso con porcentaje y versión en curso, y botón cancelar (polling de 600 ms mientras está activo).
- `Instances/List.vue`: chip pulsante "Verificando…" en la tarjeta mientras dura.

### 5. RAM por defecto de 2 GB cuando la PC lo permite

`Instance/Manager.go` (`Create`): si el equipo tiene al menos 2,5 GB de RAM total (`platform.TotalRAMMB() >= 2560`), la config nueva nace con `MaxRAM = 2048` y `MinRAM = 512` (en vez de sin RAM configurada).

### 6. Backups automáticos comprimidos

- Nuevo `Instance/Backup.go`: `CreateBackup(name)` genera `<instancesDir>/backups/<name>.zip` comprimiendo la carpeta de la instancia (excluye `logs/` y `*.lock`), con escritura atómica (`.tmp` + rename). `maybeAutoBackup(name, cfg)` decide según `backupSchedule`: `session` → siempre que termina una sesión de juego; `hours`/`days` → solo si pasó el intervalo desde `LastBackupAt`.
- `Instance/Launch.go`: tras el `Done()` de una sesión (goroutine post-cierre) se llama `maybeAutoBackup`.
- `InstanceLaunchConfig` gana `backupSchedule`, `backupInterval` y `lastBackupAt` (persistidos por `UpdateConfig`).
- `Instances/Settings.vue`: sección "Backups automáticos" (frecuencia: desactivados / al cerrar una sesión / cada X horas / cada X días) + botón "Crear backup ahora" (usa `CreateInstanceBackup` y muestra la ruta del zip).

### 7. Personalización: etiquetas y descripción

- `Instances/Form.vue`: campo de etiquetas con chips (añadir con Intro, borrar con Backspace o la X, máx. 8 etiquetas de 24 caracteres, espacios convertidos en guiones) y descripción con `maxlength="512"` + contador `X/512`.
- `Instance/Manager.go` (`List`): `InstanceInfo` gana `Tags` (antes no se exponía en la lista).
- `Instances/List.vue`: chips con las primeras 4 etiquetas en cada tarjeta (+N si hay más) y la búsqueda ahora también encuentra por etiqueta.
- Los GIF siguen soportados como icono/banner: el backend ya aceptaba `.gif` (`imageExts`) y `<img>` los anima.

## Por qué

El usuario pidió un control más fino por instancia (GC desactivado, args propios, logs detallados, verificación que inmovilice el uso), backups de seguridad automáticos y una personalización mayor de las tarjetas (etiquetas, descripción larga). Además, el widget de descargas se rompía con textos largos (Bug-2, documentado aparte).

## API afectada

- Nuevos campos en `InstanceLaunchConfig` (`backupSchedule`, `backupInterval`, `lastBackupAt`, `detailedLogs`).
- `AdvancedConfig.DetailedLogs` (Core).
- `InstanceInfo.Tags`.
- Bindings Wails nuevos (regenerados con `wails generate module`): `StartInstanceVerify`, `CancelInstanceVerify`, `InstanceVerifyStatus`, `CreateInstanceBackup`.

## Comportamiento anterior/nuevo

- GC: no se podía "desactivar" el GC personalizado (vacío o `auto` era lo máximo) → opción "Desactivado" explícita.
- Args de instancia: se guardaban en la UI pero `UpdateConfig` los descartaba → ahora se persisten y sustituyen a los globales.
- Logs detallados: opción de UI sin efecto real → ahora fuerza `--log-level debug` en el juego.
- Verificación: solo global y sin bloqueo → verificación por instancia que bloquea su uso durante el proceso.
- RAM: las instancias nuevas nacían sin RAM configurada → 2 GB por defecto en equipos con ≥ 2,5 GB.
- Backups: inexistentes → automáticos por sesión/horas/días y manuales, como `instances/backups/<nombre>.zip`.
- Tarjetas: solo título/nombre → etiquetas, descripción de hasta 512 caracteres y búsqueda por etiqueta.

## Cómo verificar

- `go build ./...` (raíz): compila sin errores.
- `wails generate module` + `bun run build` (frontend): type-check + build correctos.
- Crear una instancia con etiquetas y una descripción larga; comprobar que la tarjeta muestra las etiquetas y que se puede buscar por una de ellas.
- Configurar una instancia: GC "Desactivado", args JVM/game propios, logs detallados y frecuencia de backup; lanzar el juego y comprobar en el comando real que no hay flags de GC, que los args propios están, que aparece `--log-level debug` y que al cerrar la sesión se genera el zip en `instances/backups/`.
- Verificar la instancia: durante el proceso el botón Jugar y las acciones quedan deshabilitadas y aparece el chip "Verificando…"; al terminar se muestra el resumen de versiones y problemas.