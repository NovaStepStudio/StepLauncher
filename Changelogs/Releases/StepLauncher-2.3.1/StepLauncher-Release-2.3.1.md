# Release 2.3.1 — StepLauncher

- **Fecha**: 2026-08-15
- **Versión**: 2.3.1

## Resumen

Esta actualización pone el foco en la **estabilidad del lanzamiento** y en convertir a las **instancias** en un sistema completo y cómodo: panel propio, varias versiones por instancia, modloaders por instancia, capturas y ajustes independientes. Además estrena un **nuevo diálogo de bienvenida** en el primer inicio (con modelo 3D de jugador), un **directorio del launcher configurable** (Normal, Minecraft, Portable o Custom), un **verificador de integridad manual**, la vista previa completa de personalización y una tanda importante de correcciones — entre ellas, la más crítica: **vanilla y NeoForge/Forge moderno vuelven a arrancar** (classpath/module path corregido) y el soporte de **instaladores legacy de Forge (1.8.9 y anteriores)** queda reparado.

Se tocaron más de 30 cambios registrados entre backend Go y frontend Vue.

---

## Funcionalidades nuevas

### 1. Sistema de instancias completo
- **Panel propio** de instancias con grilla de tarjetas, favoritos, menú de acciones (clonar, verificar, capturas) y estado de descarga en vivo.
- **Vista de detalle** por instancia: héroe con icono/banner, chips del modloader, secciones de versiones instaladas / modloader / integridad / capturas y panel de descarga integrado (Releases/Snapshots/Antiguas).
- **Varias versiones por instancia** y **modloaders por instancia** (Fabric, Quilt, LegacyFabric, Forge y NeoForge) con instalación, detección del instalado y registro en el resumen.
- **Ajustes propios** por instancia: RAM, Java, pantalla completa, resolución, GC y GPU, con librerías compartidas entre instancias (carpeta `shared`).
- **Capturas por instancia** (panel de Fotos con soporte por instancia), **botón "Abrir carpeta"**, clonación con confirmación por nombre, **verificación de integridad** y ocultar el launcher al jugar, igual que en el flujo global.
- **Descarga completa real de la versión**: ya no baja solo el client jar; descarga librerías, assets e índices y activa la versión automáticamente al terminar.
- La descarga de una instancia **sigue visible aunque cierres el modal**: barra fina en la tarjeta, overlay temporal en el detalle y recuperación del estado al volver a abrir.

### 2. Bienvenida nueva en el primer inicio
- Diálogo de bienvenida en **3 pasos**: bienvenida → **personalizar el launcher** (7 paletas de acento en vivo, animaciones/desenfoque/sombras) → **crear la primera cuenta**.
- **Modelo 3D de jugador** (Skin3D) con skins aleatorias y el formulario de cuenta integrado; sin salidas laterales: solo avanzar o configurar.
- **La carpeta del launcher pasa a ser el primer paso obligatorio**: sin carpeta configurada no hay escapatoria de la bienvenida (selector Normal/Minecraft/Portable/Custom con detección de `.minecraft` existente).
- Nuevo indicador de carga: el pollo de Minecraft (`chicken_jockey_run.gif`) en todas las pantallas de espera.

### 3. Directorio del launcher configurable
- Sección **"Directorio del launcher"** en Ajustes con 4 modos: **Normal** (por defecto), **Minecraft** (usa tu `.minecraft` oficial), **Portable** (junto al ejecutable) y **Custom** (ruta a elección con "Examinar").
- Cambiar de directorio **valida y reinicia** la app sin sobrescribir tu configuración existente (solo se copia si el destino no tiene ninguna).
- Nueva opción **"Carpeta de juego separada"** desactivable: al desactivarla, el juego usa el propio directorio del launcher (forzado en modo Minecraft).
- **Archivos de Minecraft protegidos** contra sobrescritura (`launcher_profiles.json`, `options.txt`, `servers.dat`, etc.): ya no se pisotean al lanzar.

### 4. Verificador de integridad manual
- Botón **"Verificar integridad de archivos"** en Ajustes con selector de sector: **Todo / Global / Instancias**.
- Verificación por fases en segundo plano (existencia, reintentos y hash SHA1 + tamaño) con **progreso en vivo (X/Y archivos)** y reparación automática de archivos corruptos o faltantes.

### 5. Vista previa completa de personalización
- Modal a pantalla completa que **replica la interfaz real con tus colores** en vivo: cambia un selector y la vista previa se repinta al momento, con barra lateral de fondos (imagen/vídeo/dinámico) y ejemplos de colores.

### 6. Descargas más resilientes
- Timeouts y reintentos ampliados con **reducción adaptativa de trabajadores** ante fallos masivos y reintento final de tareas fallidas.
- **Pre-escaneo de archivos existentes** antes de descargar: no se vuelve a descargar lo que ya está (y los totales mostrados son reales).
- Widget flotante que muestra **"N descargas activas"** y cubre también las instalaciones de modloaders ("Instalando Forge en <instancia>…").
- Caché mejorada: los instaladores y sus logs se limpian por TTL (24 h), escritura atómica y limpieza inmediata al arrancar.

### 7. Y además…
- **Modal de crash rediseñado**: badge de categoría, grilla de 9 datos (versión, fecha, instancia, jugador, PID, tiempo en juego, estado, RAM máxima, Java), motivo a ancho completo, botón **"Copiar log"** y pestaña con las rutas de los 3 logs + botón **"Abrir"** directo.
- **Opción "Lanzar Minecraft al terminar una instalación"** en Ajustes: al terminar de instalar una versión o modloader, el juego arranca solo.
- **Reinstalación de versiones ya instaladas**: badges "REINSTALAR" con aviso en ambos modales.
- **ESC por capas**: un solo atajo global que cierra los diálogos en el orden correcto (aviso → panel → ventana).
- **RAM mínima siempre 512 MB** en todos los flujos de lanzamiento.
- **Rich Presence con tope de reintentos**: con Discord cerrado ya no hay cientos de intentos de conexión ni spam de logs; se reanuda solo al activarlo.

---

## Errores de la auditoría resueltos en esta versión

Todos los incidentes mencionados abajo fueron detectados durante el desarrollo, registrados en la auditoría interna del proyecto y quedaron corregidos para esta release:

- [StepLauncher-Error-1: El botón Jugar decía "Descargando…" durante la extracción de nativos](../../Errors/StepLauncher-2.3.1/StepLauncher-Error-1.md) — corregido.
- [StepLauncher-Error-2: Instaladores legacy de Forge (1.8.9 y anteriores) fallaban al instalar](../../Errors/StepLauncher-2.3.1/StepLauncher-Error-2.md) — corregido.
- [StepLauncher-Error-3: Instalación de modloaders legacy: perfil incompleto, instalador no siempre ejecutado y logs no guardados](../../Errors/StepLauncher-2.3.1/StepLauncher-Error-3.md) — corregido.

---

## Correcciones y bugs resueltos

### Lanzamiento de Minecraft (lo más importante)
- **Vanilla y NeoForge/Forge moderno vuelven a arrancar**: el classpath/module path se construye ahora correctamente (`--module-path` + `-cp` deduplicado). Vanilla 1.21.1 y NeoForge 21.1 fallaban al instante (exit 1 sin mensaje visible).
- **Deduplicación del classpath**: al combinar librerías de versión base y modloader ya no hay jars repetidos (el `BootstrapLauncher` moría con "Duplicate key").
- **Split package de NeoForge corregido**: el client jar se copia al nombre de archivo correcto (`ensureClientJarCopy`) y las librerías parent+child se fusionan dando prioridad al modloader (otro origen de `ResolutionException`).
- **RAM mínima siempre 512 MB**: eliminado el fallback de "la mitad de la máxima" (podía generar `-Xms2048M` absurdos).
- **Java 8 oficial (jre-legacy) para versiones antiguas**: si la versión no declara Java y el del sistema es moderno (≥ 17), se conmuta automáticamente al runtime oficial adecuado (Forge 1.12.2 ya no crashea con Java 25).
- **Forge antiguo con fallback `-universal.jar`**: si el jar exacto da 404 en Maven, se reintenta con el universal.
- **Fases reales en el botón Jugar**: ahora distingue "Descargando… → Extrayendo nativos… → Lanzando…" y el globo de estado se limpia al arrancar el juego.

### Modloaders
- **Forge/NeoForge modernos instalados con el instalador oficial** (`java -jar <instalador> --installClient`): los jars parcheados solo los generan los procesadores del instalador, así que ahora se ejecuta de verdad, con resolución de Java por prioridad (runtime oficial → Java del launcher → Java del sistema).
- **Instaladores legacy de Forge (1.8.9 y anteriores) corregidos**: se usa el perfil embebido (`install_profile.versionInfo`), el instalador se ejecuta siempre (tolerando el no-op de los legacy) y el jar universal se extrae a su coordenada Maven correcta.
- **NeoForge 26.x/2026 resuelto**: el mapeo a la versión de Minecraft ya no antepone un prefijo incorrecto y soporta versiones de 4 partes (vuelve a listar sus versiones reales).
- **Orden de versiones de los modloaders corregido**: Forge/NeoForge/Quilt se ordenan por versión real (antes lexicográfico, con las viejas arriba).
- **Estado del modloader sin archivos en disco**: detección derivada del disco y caché en memoria (adiós a carpetas de estado fantasma) y **logs de los instaladores conservados** en `cache/modloader-logs`.
- **Registro del modloader en el metadata de la instancia**: ya no solo consta la versión base de Minecraft.

### Instancias
- **RAM de instancia en GB corregida**: el campo "RAM máxima (GB)" guardaba megabytes (4 → `-Xmx4M`); ahora se convierte correctamente (con migración de valores antiguos).
- **Librerías de instancia movidas a `shared`**: al lanzar o al terminar un modloader se consolidan y la carpeta de la instancia se limpia.
- **Borrado limpio**: eliminar una instancia ya no deja descargas o instalaciones fantasma registradas.
- **Caché global fuera de `shared/cache`**: todo el caché vive en el directorio de trabajo.

### Descargas e interfaz
- **Widget de descarga sin espaciado fantasma** y sin "MB" duplicado; el botón "Abrir panel" del detalle abre el modal de instalación correcto (antes apuntaba a un overlay invisible).
- **Verificación con progreso en vivo (X/Y)** en ambos modales de instalación.
- **Crashes detectados siempre**: el listener de crash ya no depende de "ocultar el launcher"; con la opción desactivada el modal se abre igual y la ventana se restaura.
- **Detección de `.minecraft`** y selectores de carpeta mejorados (en la bienvenida y en Ajustes).

### Cuentas y sistema
- Los logs del launcher ya reportan la **versión 2.3.1** (decían 2.3.0) y el User-Agent de descarga de skins también.
- **Onboarding robustecido**: el modal de bienvenida ya no se pierde con el failsafe del splash ni cuando se configura todo antes de las llamadas de red.

---

## Cambios técnicos (para curiosos y desarrolladores)

- **Refactor integral del frontend por dominios** (feature-first): `Accounts/`, `Instances/`, `Versions/`, `Settings/`, `Downloads/`, `News/`, `Welcome/`, `Updates/`, `Crash/`, `Login/`, `Screenshots/`, `Launcher/` con estilos colocalizados; lo transversal vive en `Common/`.
- **Sistema global de overlays con Host**: los diálogos se abren por nombre desde un único punto y se eliminan todos los anidamientos de modales; nueva confirmación global por Promise (`ask()`).
- **Paneles reales**: Instancias y Capturas dejan de ser overlays oscuros (paneles a pantalla completa con `v-show`) y la visibilidad de los 9 modales principales se centraliza en el store.
- **Descargador resiliente**: timeout global 30→90 s, retries 3→5 con jitter, 8 workers con reducción adaptativa, pre-escaneo y IDs de descarga únicos por dominio (`ver-`/`inst-`/`dl-`).
- **Verificador de integridad por fases** en goroutine (sin bloquear bindings): índice → existencia → reintento → SHA1.
- **Directorio de trabajo desacoplado**: la preferencia se persiste fuera del workdir (`%APPDATA%\StepLauncher\directory.json`) con validación de bloqueo y reinicio.
- **Caché con TTL y escritura atómica** (tmp + rename) y reporte de tamaños reales por categoría.
- **Estilos SCSS separados** de los `.vue` (~5.300 líneas movidas a módulos) y variables de transición unificadas (de 59 a 13).

---

## Qué significa para el usuario

- **No necesitas hacer nada manual**: tu configuración, cuentas, historial y carpetas se conservan; los cambios de estructura (librerías a `shared`, caché a workdir) se aplican solos en el primer arranque.
- Si vienes de una versión anterior, el launcher te **notificará si hay actualización** y podrás actualizar con un clic (o desde Ajustes > Acerca de > "Buscar actualización").
- Si juegas con **instancias**, la descarga ahora es completa de verdad y puedes instalar modloaders y ver capturas por instancia.
- Si tu sistema usa **Java moderno** y juegas versiones antiguas (p. ej. Forge 1.12.2), el launcher descarga y usa el Java correcto automáticamente.

## Cómo actualizar a esta versión

- **Actualización automática**: la app te ofrecerá la nueva release desde GitHub (`NovaStepStudio/StepLauncher`) con el botón "Actualizar" (en Windows se lanza el `StepLauncher-Updater.exe`).
- **Actualización manual**: descarga la release desde la página de GitHub del proyecto y reemplaza el ejecutable; tus datos de configuración se conservan.
