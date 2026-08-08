# Release 2.3.0 — StepLauncher

- **Fecha**: 2026-08-07
- **Versión**: 2.3.0

## Resumen

Esta actualización es la más completa hasta la fecha: incorpora un **sistema de cuentas** con login real (offline y authlib-injector/Yggdrasil), **presencia en Discord**, **actualizador automático** desde GitHub, **historial y diagnóstico de crashes** de Minecraft, **soporte moderno de Forge y NeoForge (1.21+)**, un **instalador de versiones y modloaders** renovado por completo, descargas a fondo reconstruidas y una personalización visual muy ampliada. Además corrige decenas de bugs de estabilidad (bloqueos, congelamientos, crashes de lanzamiento, problemas de skins y colores) y reorganiza toda la parte técnica para hacerla más robusta y mantenible.

Se tocaron más de 80 archivos entre backend Go y frontend Vue: unos 28 nuevos y el resto modificados.

---

## Funcionalidades nuevas

### 1. Sistema de cuentas (offline y authlib)
- Nuevo gestor de cuentas persistido en `launcher_accounts.json` con soporte para dos tipos:
  - **Offline** (sin conexión): nick validado (3-16 caracteres, letras, números y guion bajo) y UUID generado de forma estándar.
  - **AuthLib (Yggdrasil)**: login real contra servidores compatibles con `authlib-injector` (servidores no-premium y modificados), con servidor de autenticación configurable, usuario y contraseña.
- **Panel de cuentas** completo: lista, crear, editar, eliminar, seleccionar cuenta activa y ver tu avatar.
- **Renovación de sesiones** (manual y automática) para mantener la sesión vigente sin volver a iniciarla, con **login cancelable** (modal de progreso con cancelación real en segundo plano).
- La cuenta seleccionada se inyecta automáticamente al lanzar Minecraft.
- `authlib-injector.jar` se descarga y verifica solo (checksum SHA-256) con fuente de respaldo (BMCLAPI) si el servidor oficial no responde.

### 2. Avatares y skins
- Los avatares de las cuentas ahora son la **cabeza real de la skin** (cara + sombrero, generada desde el backend) en lugar de un color plano.
- **Caché de skins y avatares en disco** (se evitan descargas HTTPS repetidas); la tarjeta de usuario muestra el avatar de la cuenta activa.
- Mejoras del renderizado de skins en el navegador (Canvas API), incluida la corrección del brazo *slim* en skins de alta resolución.

### 3. Presencia en Discord (Rich Presence)
- El launcher **se comunica directamente con el cliente de Discord** (IPC propio: named pipes en Windows, sockets Unix en Linux/macOS con soporte de Snap/Flatpak) y muestra el estado:
  - "Navegando por el menú" → "Lanzando Minecraft" → "Jugando Minecraft" (con tiempo de juego).
- Reconexión automática si Discord se abre o se reinicia y **opción de activar/desactivar** en Ajustes > Comportamiento (activado por defecto).

### 4. Actualizador automático desde GitHub
- El launcher consulta la **última release de GitHub** (`NovaStepStudio/StepLauncher`) y avisa cuando hay versión nueva (modal con notas de release y botón "Actualizar").
- En **Windows** descarga y lanza de forma desacoplada el `StepLauncher-Updater.exe` (se cierra el launcher y el updater completa el reemplazo); en **Linux/macOS** abre la página de la release.
- La actualización nunca es obligatoria: opción "Buscar actualizaciones al iniciar" y botón manual "Buscar actualización" en Ajustes > Acerca de.

### 5. Lanzamiento moderno de Forge y NeoForge (1.21+)
- Se soporta el formato moderno de **FML/NeoForge (JPMS)**: se genera `launcher.properties` con el secreto aleatorio `fml.client.secret` (token `${launcher_properties}`), y el classpath se entrega por `--module-path` cuando el template o el plan de ejecución lo requiere (manteniendo el classpath clásico para vanilla/Forge antiguo/Fabric).
- **Runtime de Java oficial** si falta el necesario (reutilizando el mismo motor de descargas), se prioriza `javaw.exe` en Windows y se inyectan los flags correctos (`-Djava.library.path`, `-javaagent` de authlib-injector apuntando a la ruta del jar).
- Mapas de versión **NeoForge → Minecraft** corregidos (p. ej. `21.1.x` → `1.21.1`, no `1.21`) y filtrado exacto de compatibilidad de loaders por versión.

### 6. Instalador de versiones y modloaders rediseñado
- Lista de versiones de Mojang con **filtros**: Releases / Snapshots / Antiguas, con grupos y marcas de versión "más reciente" / "recomendada".
- **6 modloaders** con selector visual: Vanilla, Fabric, Forge, NeoForge, Quilt y Legacy Fabric; los **incompatibles con la versión elegida se deshabilitan** mostrando la nota.
- Modo recomendado (automático) o selección manual de versión.
- **Instalación con barras de progreso reales**: porcentaje global, velocidad, secciones, archivos activos, cola y log; botones **Pausar / Reanudar / Cancelar**, descarga del loader en segundo plano y **recuperación de descargas activas** al abrir el launcher de nuevo.
- Los jars de librerías usan repositorios Maven centralizados con fallback por grupo (`maven.minecraftforge.net`).

### 7. Descargas rediseñadas (motor completo)
- En la interfaz: **anillo de progreso SVG** con porcentaje y check al completar, detalles técnicos colapsables (estadísticas, sección, archivos activos, cola, log) y **widget flotante** "Descargando…" cuando el modal está cerrado.
- Detrás: progreso real por sección (librerías / objetos / Java…), **velocidad instantánea suavizada**, cancelación limpia (sin locks rotos) y distinción clara entre descargar / **verificar (SHA1)** / completado.
- Opción **"Verificación de integridad (SHA1)"** en Ajustes > Comportamiento (activada por defecto; puede desactivarse).

### 8. Crashes de Minecraft: diagnóstico completo
- Modal de crash rediseñado en **dos pestañas**: "Datos del error" y "Códigos de errores", con el **contenido del reporte/crashlog** en monoespaciada y botón **Copiar** (ya no solo rutas de archivo).
- **Códigos de salida oficiales de Minecraft** (tabla real `net.minecraft.ExitCodes`): categorías explicadas en castellano (crash con/sin reporte, crash temprano, watchdog, cierre por shutdown…), con versión, código de salida, uptime, PID y jugador.
- Nuevo **historial de crashes persistente** con la ruta de los 3 logs (launcher, juego y JVM) para consultar incidentes antiguos.

### 9. Gestión de versiones instaladas
- El launcher recuerda la **última versión seleccionada** (`selectedVersion` persistida) y los modos **versión/perfil se vuelven excluyentes**: elegir versión limpia el perfil activo; elegir perfil muestra su versión.
- El panel de versiones tiene identidad propia (acento violeta) con **badges de tipo por color** (release verde, snapshot dorada, beta naranja, alpha roja).
- Perfiles: formulario completo (icono importado opcional, **resolución/pantalla completa precargadas con las de la ventana actual**), compat con perfiles oficiales de Forge (`lastVersionId`).

### 10. Navegación y novedades visuales
- Nueva barra superior (TopOptions) y barra lateral (Sidebar) con secciones (Inicio, Fotos, Instancias, Mods, Skins).
- **Selector de color totalmente rediseñado** (ColorField): arrastre con *pointer lock* (no se corta en el borde de la ventana), **rueda del ratón** y **flechas del teclado** para el tono, previsualización en vivo durante el arrastre y **guardado al soltar/cerrar** (evita escrituras continuas en disco), soporte de colores cortos `#RGB`/`#RGBA`.
- **Personalización ampliada**: nuevos colores editables (progreso, botón Jugar, botones primarios, borde/borde modal, error, éxito, etiquetas, aviso), **sombra de texto**, tamaños de tipografía y el nuevo sistema de **fuentes personalizadas** (importas archivos con tu letra y se aplican a la interfaz).
- **Previsualización real de cada color** con una réplica en miniatura del modal junto al selector.
- Toggle "**Ocultar el launcher al abrir Minecraft**": la ventana se oculta durante la partida y reaparece solo cuando el juego termina/crashea.

### 11. Pantallas de juego y más
- **Visor de capturas** (Fotos) con **zoom y desplazamiento** reales (1-4× con rueda) y actualización automática al volver del juego.
- El **modal de instalación** permanece en estado instalando hasta que el loader termina realmente.
- Cierre automático de modales por inactividad y **apagado limpio**: al cerrar la app se cancelan descargas y se detienen los juegos.

---

## Errores de la auditoría resueltos en esta versión

Todos los incidentes mencionados abajo fueron detectados durante el desarrollo, registrados en la auditoría interna del proyecto y quedaron corregidos para esta release:

- [StepLauncher-Error-1: Self-deadlock de RWMutex en Config.Save() (build colgado)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-1.md) — corregido.
- [StepLauncher-Error-2: accessToken de Yggdrasil guardado sin los guiones del UUID](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-2.md) — corregido.
- [StepLauncher-Error-3: `launcher_accounts.json` reportado como "corrupto" al cargar](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-3.md) — corregido.
- [StepLauncher-Error-4: La barra de tono (HUE) no se podía mover](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-4.md) — corregido.
- [StepLauncher-Error-5: El panel de Ajustes no cargaba (Temporal Dead Zone en ColorField)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-5.md) — corregido.
- [StepLauncher-Error-6: Guardado continuo del color durante el arrastre](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-6.md) — corregido.
- [StepLauncher-Error-7: El ColorField corrompía los colores por defecto (`#0005`/`#111` → `#000000`)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-7.md) — corregido.
- [StepLauncher-Error-8: Crash al cancelar una descarga (doble Unlock del mutex)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-8.md) — corregido.
- [StepLauncher-Error-9: Stack overflow al lanzar Minecraft (recursión infinita en `SubstituteVars`)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-9.md) — corregido.
- [StepLauncher-Error-10: LWJGL no encuentra `lwjgl.dll` en versiones con natives en subcarpetas (26.2)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-10.md) — corregido.
- [StepLauncher-Error-11: El descargador instalaba natives de 32 bits (y de todas las plataformas) en sistemas de 64 bits](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-11.md) — corregido.
- [StepLauncher-Error-12: El panel de Fotos no cargaba la galería](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-12.md) — corregido.
- [StepLauncher-Error-13: Un perfil importado con `lastVersionId` lanzaba la versión base en vez del modloader](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-13.md) — corregido.
- [StepLauncher-Error-14: Instalar un modloader creaba estructura de instancias y dejaba la versión fuera de `versions/`](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-14.md) — corregido.
- [StepLauncher-Error-15: El mapeo de NeoForge a Minecraft era incorrecto (21.1.x → 1.21 en lugar de 1.21.1)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-15.md) — corregido.
- [StepLauncher-Error-16: Forge/NeoForge modernos fallaban al descargar el jar "client" (URL vacía → `.forge_patched_minecraft` no encontrado)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-16.md) — corregido.
- [StepLauncher-Error-17: Rich Presence no aparecía en Discord (named pipes de Windows)](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-17.md) — corregido.
- [StepLauncher-Error-18: Al instalar un modloader se creaba una carpeta "shared" con librerías](../../Errors/StepLauncher-2.3.0/StepLauncher-Error-18.md) — corregido.

---

## Correcciones y bugs resueltos

### Bloqueos y congelamientos
- **Deadlock en la configuración** que colgaba el build ("Generating bindings" con la CPU al 0%): un `RLock` reentrante sobre el mismo mutex dentro de `Save()`. Patrón cambiado: liberar el lock antes de loguear y nunca llamar a funciones que adquieran el mismo lock.
- **Stack overflow al lanzar Minecraft**: la sustitución de variables `${...}` en plantillas era recursiva e incondicional; pasa a ser iterativa, con tope de 8 pasadas y sin reprocesar variables sin valor.
- **Doble `Unlock()` al cancelar o pausar una descarga** (`sync: unlock of unlocked mutex`) — cada lock se libera una y solo una vez.
- **Pantalla de Ajustes que no cargaba** por una variable leída antes de su declaración (Temporal Dead Zone) en el selector de color.

### Lanzamiento de Minecraft
- **LWJGL no encontraba `lwjgl.dll`** en versiones nuevas (26.2): se lee la ruta que el template indica y se extrae a la subcarpeta correcta (p. ej. `natives/java`).
- **Natives equivocadas (32 bits / otras plataformas) en sistemas x64**: selección por coincidencia exacta del clasificador `natives-windows`, `natives-windows-arm64`, etc. según SO + arquitectura (función central única).
- **Perfiles de Forge importados lanzaban la base en vez del modloader**: se soporta el campo clásico `lastVersionId`, con normalización en carga/guardado y merge de resolución/pantalla completa.
- **Librerías de Forge/NeoForge indescargables (URL vacía)**: resolución Maven centralizada con repositorio/fallback por grupo.
- **NeoForge → Minecraft desincronizado**: mapeo reescrito con la regla exacta mientras el segundo segmento es `0`.

### Instalación y estructura
- **Instalar un modloader ya no crea instancias**: la instalación va siempre al directorio de trabajo + `versions/` (el sistema de instancias sigue siendo opcional y no interfiere).
- **Carpeta `shared` fantasma eliminada**: las librerías viven siempre junto al launcher.
- **Historial de crashes**: rutas de los 3 logs relativas al workdir y escritura atómica.

### Cuentas y configuración
- `launcher_accounts.json` **corrupto al cargar**: se unificaron escritura/lectura en la raíz del archivo y se deserializa en variable fresca (adiós a la "cuenta fantasma").
- **Tokens de acceso con el UUID sin guiones**: normalizados a formato canónico en todos los puntos de escritura y migración.
- `extraData` pasa de array a **objeto con claves por dominio** (`assets`, `accounts`, `history`, `profiles`, `crashHistory`) con migración automática.
- **Colores cortos corruptos** (`#0005`, `#111`): el parseador acepta todos los formatos (`#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`).
- **Persistencia del color corregida**: no se escribe en disco por cada píxel arrastrado; solo al soltar, y cerrar el modal sin cambios ya no pisa valores.

### Interfaz y modales
- El modal de fotos **no cargaba la galería**: los modales que se montan una sola vez se refrescan al abrirse (no en `onMounted`) y las miniaturas usan un `Map` reactivo, además de suscribirse a los eventos de salida/crash del juego.
- El modal de crash muestra el **contenido del crashlog** (ya no solo rutas) con pestañas y botón "Copiar".
- Errantes del ColorField: tono inaccesible con ratón bajo pointer lock (anclaje de deltas rediseñado) y cierre limpio al perder el foco de la ventana.
- Integridad del layout: eliminados juegos duplicados de variables CSS que rompían el tema visual de algunos paneles.

---

## Cambios técnicos (para curiosos y desarrolladores)

- **Configuración del motor separada** en su propio paquete (`engineconfig`); la gestión de assets personalizables (fuentes, imágenes) queda centralizada en `internal/Core/Assets` con `launcher_assets.json`.
- **Motor de descargas reconstruido**: tareas por secciones, clasificadores nativos solo por arquitectura exacta, velocidades reales, cancelación controlada y cola compartida con todas las tareas (Java oficial, authlib-injector…).
- **Launcher**: refactor de construcción de argumentos base desde la config del launcher, parseo de args que respeta comillas, evento `game_prepare` para el progreso del botón "Jugar" y distinción clara de salidas (`game_stopped` vs crash).
- **Apagado limpio**: al cerrar la app (`OnShutdown`) se cancelan las descargas, se corta el login y se detienen los juegos, sin dejar subprocesos colgados.
- **Frontend**: gestor de paquetes **bun**, aliases centralizados en `vite.config.ts`, build dividido en chunks, assets con hash y limpieza del `dist`; nombres de archivos de `src` normalizados.
- **Dependencias Go**: `go-winio` (named pipes de Discord en Windows) y librerías transitivas actualizadas.
- Bindings de `frontend/wailsjs` regenerados con todas las APIs nuevas.

---

## Qué significa para el usuario

- **No necesitas hacer nada manual**: cuentas, config, historial y assets antiguos se migran y completan solos al primer arranque.
- Si vienes de una versión anterior, el launcher te **notificará si hay actualización** y podrás actualizar con un clic (o desde Ajustes > Acerca de > "Buscar actualización").
- La presencia en Discord empieza solo al abrir; puedes desactivarla en Ajustes > Comportamiento.
- La verificación de integridad (SHA1) viene activada por defecto; puede desactivarse si descargas muy grandes se vuelven lentas.

## Cómo actualizar a esta versión

- **Actualización automática**: la app te ofrecerá la nueva release desde GitHub (`NovaStepStudio/StepLauncher`) con el botón "Actualizar" (en Windows se lanza el `StepLauncher-Updater.exe`).
- **Actualización manual**: descarga la release desde la página de GitHub del proyecto y reemplaza el ejecutable; tus datos de configuración se conservan.