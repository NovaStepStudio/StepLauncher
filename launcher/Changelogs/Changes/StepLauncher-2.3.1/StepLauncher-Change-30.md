# Changes/StepLauncher-2.3.1/StepLauncher-Change-30.md

- **Fecha**: 2026-08-14
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (bun run type-check OK, go build OK).

## Qué cambió

### 1. "Abrir panel" abre el modal de instalación (no un overlay)

El botón "Panel" del overlay de descarga en el detalle de instancia abría el panel global "Descargas activas" (overlay con z-index 90, que quedaba oculto bajo el modal de instancias, z-index 100: parecía que no se abría). Ahora:

- `frontend/web/src/Modals/InstanceDetailView.vue`: el botón "Panel" emite `download` (abre `InstanceDownloadModal`, el mismo modal que abre el botón de descargar del detalle y de la tarjeta). Se elimina el import de `showDownloadsPanel`.
- `frontend/web/src/Modals/InstanceDownloadModal.vue`: `syncFromBackend()` adopta además la **instalación de modloader en curso** (sin descarga de versión activa): si `loaderDlOf(name)` está en `resolving|downloading|installing`, pasa a `phase='installing'` + `loaderPhase='running'`, carga el mensaje/progreso y el loader correspondiente, y sigue recibiendo los eventos `modloader_*` para actualizarse en vivo (el filtro de sesión no bloquea porque el modal no tiene sesión propia).
- Se elimina `panelVisible`/`showDownloadsPanel`/`hideDownloadsPanel` de `Stores/Downloads.ts`.

### 2. Adiós al panel "Descargas activas" (overlay): solo widget

- Eliminados `frontend/web/src/Modals/DownloadsPanel.vue` y `Styles/Modals/DownloadsPanel.scss`.
- `frontend/web/src/App.vue`: `openWidget()` ya no abre el panel con varias descargas; el widget siempre abre el destino según la descarga más reciente: `kind: 'version'` → modal de instalación global, resto (`instance`/`loader`) → panel de instancias (que muestra el progreso por tarjeta y en el detalle).

### 3. El botón "Panel" no desaparece al instalar un modloader

- `InstanceDetailView.vue`: la rama de instalación de modloader (`instLdr`) del overlay ahora también muestra las acciones con el botón "Panel" (antes solo existía en la rama de descarga `instDl`, por lo que el botón "desaparecía" al instalar un modloader).

### 4. RAM de instancia: el campo dice GB y ahora respeta GB

`frontend/web/src/Modals/InstanceSettingsModal.vue`:

- El campo "RAM máxima (GB)" mostraba y guardaba el valor en **MB** (al poner 4, el launcher lanzaba `-Xmx4M`). Ahora al guardar se envía `maxRam * 1024` (MB, como espera `buildJVMArgs` con `-Xmx%dM`) y al cargar se divide entre 1024.
- Migración de valores guardados por el bug anterior: si el valor almacenado es `< 1024` se interpreta como GB (lo que el usuario pretendía con la etiqueta vieja), así el "4" guardado se muestra como 4 GB y al guardar se persiste como 4096 MB.

### 5. Borrado de instancia: card eliminada al instante y sin estado fantasma

- La card ya se eliminaba de la vista (refresco optimista + recarga). Se añade limpieza de los mapeos internos `dlToInstance` y `loaderSessionToInstance` en `deleteInstance` (`Stores/Instances.ts`), para que eventos tardíos de descarga/modloader no vuelvan a crear estado para la instancia borrada.

## Por qué

El usuario reportó que al pulsar "Abrir panel" durante la descarga de una instancia "el panel no se abría": en realidad se abría un overlay con z-index inferior al del modal de instancias (invisible). Además pidió que ese botón abriera el mismo modal de descarga/instalación que el botón de descargar, que el panel "Descargas activas" fuera un widget (o nada) en lugar de un overlay inútil, que el botón siguiera visible al instalar un modloader, y reportó que configurar 4 (GB) lanzaba Minecraft con 4 MB.

## API afectada

- `Stores/Downloads.ts`: eliminados `panelVisible`, `showDownloadsPanel`, `hideDownloadsPanel` (sin uso restante).
- Sin cambios en bindings de Wails ni en el backend.

## Comportamiento anterior/nuevo

- **Antes**: "Abrir panel" abría un overlay invisible (z-index 90 < 100) y desaparecía durante la instalación de un modloader; existía el panel "Descargas activas" con overlay; la RAM de instancia se guardaba en MB con etiqueta GB (`-Xmx4M` con un "4"); al borrar una instancia la card sí desaparecía.
- **Ahora**: "Panel" abre el modal de instalación de la instancia (con progreso, pausa/cancelar y adopción del estado del modloader en curso) y está presente también al instalar un modloader; el overlay "Descargas activas" se eliminó y el widget de descarga cubre todo; 4 en el campo RAM = 4096 MB (`-Xmx4096M`), con migración de los valores viejos; el borrado limpia los mapeos internos.

## Cómo verificar

- `bun run type-check` en `frontend/`: OK.
- Manual: iniciar la descarga de una instancia → en el detalle pulsar "Panel": se abre el modal de instalación con el progreso; instalar un modloader → el botón "Panel" sigue visible y abre el modal mostrando el progreso del loader; con varias descargas activas, el widget lleva al panel de instancias; en Configuración de instancia poner RAM 4 y lanzar: `-Xmx4096M`; borrar una instancia → la card desaparece al instante y no reaparece aunque queden eventos de descarga.