# Cambios de StepLauncher 2.5.0 (Change-12) — Concurrencia y configuración de red del backend

- **Fecha**: 2026-09-18
- **Versión**: 2.5.0
- **Estado**: implementado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Cola de descargas no bloqueante

`internal/Core/Downloader/Queue.go` mueve la espera por cupo de concurrencia a la goroutine de trabajo. Iniciar una descarga ya no espera a que termine otra cuando la cola alcanzó su límite. La capacidad se actualiza sin reemplazar la cola activa, por lo que las tareas en curso conservan su seguimiento.

### 2. Estado compartido sincronizado

`internal/Handlers/Engine/engineconfig/Config.go` protege la configuración del motor con `sync.RWMutex` y resuelve los directorios antes de publicar el estado. `internal/Core/Launcher/Types.go` expone una instantánea de juego utilizada por eventos y respuestas para evitar leer campos mientras el proceso actualiza su salida.

### 3. Persistencia serializada por instancia

`internal/Core/Launcher/Instance/` incorpora un mutex por instancia para las operaciones de lectura-modificación-escritura de metadata y configuración. Las descargas finalizadas, ediciones, eliminaciones de versión, cierres de juego y backups automáticos no pisan cambios entre sí.

### 4. Red aplicada en tiempo de ejecución

`internal/Core/Downloader/Download.go` construye clientes HTTP aislados con límite de velocidad y proxy opcional. `internal/Handlers/Engine/` vuelve a configurar ambos gestores de descargas al cambiar el proxy o el límite de ancho de banda.

### 5. Cambio de directorio seguro

`internal/Handlers/Engine/Directory.go` considera las descargas de instancias, incluidas las pausadas, antes de permitir cambiar el directorio de trabajo.

## Comportamiento anterior y nuevo

- Antes, una llamada para iniciar una descarga podía esperar por un cupo ocupado y bloquear su binding.
- Ahora, la llamada registra la descarga de inmediato y la tarea espera en segundo plano.
- Antes, los ajustes de proxy y límite de velocidad se persistían sin actualizar los clientes HTTP activos.
- Ahora, las siguientes solicitudes de ambos gestores usan la nueva configuración.
- Antes, actualizaciones concurrentes de una instancia podían sobrescribir metadata o configuración reciente.
- Ahora, cada instancia serializa esas actualizaciones persistentes.

## Cómo verificar

- Iniciar más descargas que el límite configurado y comprobar que cada llamada devuelve su identificador sin esperar a que finalicen las anteriores.
- Modificar el límite de ancho de banda o el proxy y comenzar una nueva descarga para comprobar que usa el cliente actualizado.
- Finalizar descargas de versiones mientras se edita la misma instancia y comprobar que metadata y configuración conservan ambos cambios.
- Intentar cambiar el directorio de trabajo con una descarga de instancia activa o pausada y comprobar que se rechaza.
