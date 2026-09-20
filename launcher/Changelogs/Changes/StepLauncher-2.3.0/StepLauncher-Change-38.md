# StepLauncher-Change-38: Modal de crash rediseñado (pestañas Datos del error | Códigos de errores)

- **Fecha**: 2026-08-07
- **Versión**: 2.3.0 (en desarrollo)
- **Release**: StepLauncher-2.3.0 — en este release se menciona que fue añadido.
## Qué cambió

El modal de crash de Minecraft (`frontend/web/src/Modals/CrashModal.vue`) se rediseña con dos pestañas y un ancho mayor (entre el 65% y el 80% del viewport):

- **Pestaña «Datos del error»**: conserva la información general del crash (versión, código de salida, categoría, motivo, tiempo en juego, PID y jugador) más la explicación breve de la categoría.
- **Pestaña «Códigos de errores»**: muestra el código de error (código de salida) y la categoría como tarjetas, y debajo el **texto del log de error** en una caja monospace con scroll propio y botón «Copiar». **Se elimina por completo la visualización de rutas/directorios de archivos**: antes se mostraba «Reporte de crash: <ruta del archivo>», que era poco útil para el jugador.

## Por qué

La ruta del crash report en disco no aportaba información útil al jugador; lo que se necesita es el contenido del log para entender la causa (mods, librerías, JVM, etc.). Además se aprovecha el ancho extra para organizar la información en dos pestañas más claras.

## Qué se hizo

1. **Backend (contenido en lugar de ruta)**:
   - `internal/Core/Launcher/Types.go`: nuevo campo `CrashLogContent string` en `GameInstance`.
   - `internal/Core/Launcher/Events.go`: nuevo campo `CrashLogText string` (`json:"crashLogText,omitempty"`) en `GameEventData`, rellenado desde `inst.CrashLogContent`.
   - `internal/Core/Launcher/Launcher.go`: en `recordCrash` se completa `instance.CrashLogContent` con la nueva función `readCrashLogText`, que lee el contenido del crash report / JVM log (limitado a 32 KB y 400 líneas finales) y, si no hay archivo legible, usa las líneas de contexto del buffer de memoria del log del juego. La ruta (`CrashLog`) se sigue registrando internamente para el historial, pero ya no se expone visualmente en el modal.
2. **Frontend**:
   - `stores/launcher.ts`: `CrashInfo` gana `crashLogText?: string` y se mapea en `onGameCrash` desde `d.crashLogText`.
   - `CrashModal.vue`: reescrito con pestañas (`Datos del error` / `Códigos de errores`), tarjetas de código/categoría, caja de log con `white-space: pre-wrap`, scroll propio y botón copiar (con estado «Copiado» temporal). Eliminado el bloque que imprimía la ruta del log.

## API afectada

- Evento `game_crashed`: nuevo campo opcional `crashLogText` con el texto del log de error (sin rutas). Sin cambios en bindings Wails.

## Cómo verificar

- `go build ./...` OK en Windows.
- `bun run build` (type-check + vite) OK en `frontend/`.
- Al crashear una versión: el modal aparece con las dos pestañas; en «Datos del error» está la info completa; en «Códigos de errores» se ve el código de salida y el texto del log sin ninguna ruta de archivo; «Copiar» copia el log al portapapeles.