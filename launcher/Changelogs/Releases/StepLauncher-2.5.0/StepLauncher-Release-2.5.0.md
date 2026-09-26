# Actualización StepLauncher-2.5.0

- **Fecha**:   2026-09-26
- **Versión**: 2.5.0
- **Estado**:  publicada

## De dónde viene la 2.5.0: 2.3.1 → 2.4.1 → salto a Wails v3

La 2.5.0 no empieza de cero: culmina el camino de las dos versiones anteriores y por eso se publica como final, no como beta.

- **2.3.1** ([changelog](../StepLauncher-2.3.1/StepLauncher-Release-2.3.1.md)): convirtió las instancias en un sistema completo (panel propio, varias versiones y modloaders por instancia, capturas y ajustes independientes), estrenó la bienvenida en 3 pasos, el directorio del launcher configurable (Normal/Minecraft/Portable/Custom), el verificador de integridad manual y descargas resilientes. Su corrección más crítica: vanilla y Forge/NeoForge moderno volvieron a arrancar.
- **2.4.1** (línea que quedó en desarrollo): llevó la base Wails v2 al límite — soporte de versiones de terceros (BatMod: paths Maven normalizados y logs log4j2-XML legibles, [Change-1](../../Changes/StepLauncher-2.4.1/StepLauncher-Change-1.md)), configuración avanzada por instancia (sin GC, argumentos JVM/juego propios, logs detallados, verificación propia bloqueante, backups comprimidos, [Change-2](../../Changes/StepLauncher-2.4.1/StepLauncher-Change-2.md)) y una tanda de Java/instancias ([Bug-1](../../Bugs/StepLauncher-2.4.1/StepLauncher-Bug-1.md), [Bug-2](../../Bugs/StepLauncher-2.4.1/StepLauncher-Bug-2.md)). Ahí se chocó con el techo de Wails v2.
- **2.5.0**: reconstrucción total sobre **Wails v3** (backend por servicios, bandeja nativa, build por plataforma con `Taskfile.yml`, frontend con Bun). Todo lo de 2.3.1 y 2.4.1 sigue funcionando sobre la base nueva, y lo que sigue abajo es lo que la 2.5.0 añade y corrige encima.

## Funcionalidades nuevas

### 1. Base nueva: migración completa a Wails v3
- Backend Go y frontend Vue reconstruidos sobre **Wails v3** (antes v2): arranque más rápido, bandeja del sistema nativa y servicios por dominio en vez de un monolito.
- Gestor frontend **Bun** e infraestructura de build por plataforma (`Taskfile.yml` + `build/config.yml`): instalador NSIS en Windows, `.deb`/`.rpm`/`.AppImage` en Linux y `.app`/`.dmg` en macOS.
- El proyecto es **solo escritorio (Windows/macOS/Linux)**: se eliminó el soporte iOS.

### 2. Panel de Mods de Modrinth
- Explora y descarga **mods, shaders, texturas y modpacks** directo de Modrinth, solo contenido de cliente.
- Los **modpacks (.mrpack)** crean su instancia solos: descargan su Minecraft, instalan su loader y quedan listos para jugar; la instancia aparece al 100% y bloqueada mientras se crea.
- Pestaña **Instalado** con iconos reales de cada archivo (Fabric, Quilt, Forge, NeoForge), mover contenido entre el juego global y tus instancias, y abrir su ubicación.
- Se corrigió el **404 de Forge** con versiones cortas del manifiesto y si un loader falla, la instalación **sigue igual** con aviso (ya no se pierde todo).

### 3. Música reconstruida
- Reproducción de fondo **fiable** (se quedaba muda en silencio: ahora suena siempre) con control de **volumen en Ajustes** y aviso visible si una pista falla.
- Panel de música nuevo con tu **biblioteca real**: selector visual de pistas, playlists sin escribir rutas, modos de reproducción y carátula como fondo.
- Índice persistente: bibliotecas grandes abren al instante y **sin picos de RAM**.

### 4. Bandeja del sistema (tray) renovada
- Un solo icono: clic para mostrar/ocultar, doble clic para enfocar, menú con clic derecho.
- Lanza tus **últimas sesiones e instancias** desde la bandeja, cambia de **cuenta** y usa acciones rápidas sin abrir la ventana.

### 5. Arranque trazable y Ajustes reorganizados
- Nueva pantalla de carga con **progreso real por pasos y logs en vivo**: si algo tarda, ves qué es.
- Ajustes divididos en secciones claras (**Red, Descargas, Integridad, Almacenamiento…**) con lenguaje sin jerga, textos técnicos precisos donde importan (−Xmx, hash, JVM, Yggdrasil) y cada sección abre desde arriba.
- **Acerca de** renovado: banner oficial, créditos a terceros en la web, accesos a cuentas y avisos legales.

### 6. Primer arranque inteligente
- Nueva paleta oscura por defecto, **RAM inicial = la mitad de tu equipo** (entre 2 y 8 GB), Rich Presence, auto-refresh de cuentas y fondos desde tus capturas.

### 7. Actualizador directo a GitHub
- Sin servidores intermedios: el launcher consulta las releases de GitHub directamente.
- En **Windows** descarga el instalador (`*-installer.exe`), **cierra el launcher** y el instalador completa la actualización.
- En **Linux/macOS** no hay actualizador automático: el diálogo te avisa de la nueva versión y te lleva a GitHub para instalarla manualmente.

### 8. Red y rendimiento
- El **proxy es solo para Minecraft**: todo el launcher va en directo (HTTP/1.1), así ningún proxy rompe manifiestos, noticias o actualizaciones.
- La lista de versiones **funciona sin internet** usando la copia guardada; el interruptor de proxy ya se puede configurar sin punto muerto.
- Descargas no bloqueantes, configuración protegida y persistencia por instancia sin pisarse.

## Correcciones (con enlaces a la auditoría)

- [StepLauncher-Error-1: `wails3 dev`/`wails3 build` rotos por la eliminación de build/ios](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md) fue corregido en esta versión.
- [StepLauncher-Error-2: el tray nunca aparecía](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-2.md) fue corregido en esta versión.
- [StepLauncher-Error-3: dos iconos de tray en la bandeja](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-3.md) fue corregido en esta versión.
- [StepLauncher-Error-4: verificación antes de lanzar ignorada y fondos de galería duplicados](../../Errors/StepLauncher-2.5.0/StepLauncher-Error-4.md) fue corregido en esta versión.
- [StepLauncher-Bug-1: acumulación infinita de fondos de galería](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-1.md) fue corregido en esta versión.
- [StepLauncher-Bug-2: widget de descargas roto con textos largos y pico de RAM por carátulas](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-2.md) fue corregido en esta versión.
- [StepLauncher-Bug-3: manifiesto roto sin red + interruptor de proxy imposible de activar](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-3.md) fue corregido en esta versión.
- [StepLauncher-Bug-4: proxy manual solo para Minecraft + transporte directo](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-4.md) fue corregido en esta versión.
- [StepLauncher-Bug-5: botones secundarios pegados al fondo, hover imperceptible](../../Bugs/StepLauncher-2.5.0/StepLauncher-Bug-5.md) — arreglo aplicado en esta versión (entrada en corrección al cierre de la release).

## Notas para el usuario

- Al primer arranque, cuentas, configuración, historial y fondos se migran solos; la verificación de integridad viene activada por defecto.
- Si vienes de una beta 2.5.0, el launcher te ofrecerá la final con el botón "Actualizar": en Windows se descarga el instalador y se cierra el launcher; en Linux/macOS se abre la release para instalación manual.
