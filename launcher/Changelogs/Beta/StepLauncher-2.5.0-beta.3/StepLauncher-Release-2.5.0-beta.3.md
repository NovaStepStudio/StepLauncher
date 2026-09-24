# Beta 3 — StepLauncher 2.5.0 (config inicial, dark mode y red)

- **Fecha**: 2026-09-24
- **Versión**: 2.5.0-beta.3
- **Estado**: beta

## Resumen

Tercera beta de la 2.5.0: **config inicial completa** (mitad de RAM automática, paleta dark mode, `richPresence` dentro de `launcher`, renovación de sesiones activa), **botón para poner capturas de fondo**, **welcome sin gradientes** con splash de carga negro sólido y **mejoras internas de descargas y red**. En la web se eliminó el sistema de cuentas (login/register/dashboard): queda como página pública.

> Esta beta consolida los **Changes 12, 13, 14, 15 y 16** de la rama 2.5.0. No hay errores ni bugs nuevos desde la beta anterior. Es para pruebas internas: puede tener regresiones menores.

---

## Funcionalidades nuevas

### 1. Primera config completa y dark mode de verdad
- Al iniciar por primera vez ya no faltan datos: `dynamicImages` y `recentColors` arrancan como listas vacías y los colores por defecto son dark mode válido (sidebar `#00000066`, modal `#0b0b0b`, botones `#2b2b3d`, bordes `#131313`, jugar `#1f1f1f`, primario `#19191a`).
- Los grises `#111` viejos ni siquiera pasaban el validador interno: ahora se corrigen solos al abrir el launcher.

### 2. RAM inicial = mitad de tu equipo
- En vez de 2 GB fijos, el launcher detecta tu RAM total y asigna la mitad (mínimo 2 GB, máximo 8 GB). Ej: 16 GB → 8 GB, 8 GB → 4 GB. También aplica al restablecer ajustes.

### 3. Discord y sesiones
- `richPresence` ahora vive dentro del bloque `launcher` como booleano simple (las configs viejas se migran solas).
- "Renovar sesiones al iniciar" viene activado desde el inicio, también para archivos de cuentas antiguos.

### 4. Capturas como fondo
- El visor de Screenshots tiene botón "Colocar como fondo", igual que la galería de mods: copia la captura a `cache/backgrounds/` y la aplica al instante.

### 5. Welcome y carga en negro sólido
- Botones y tarjetas del welcome sin gradientes (color plano) y pantalla de "Cargando configuración" en negro sólido `#0a0a0a` con panel `#101010`.

### 6. Descargas y red más sólidas (interno)
- Iniciar una descarga ya no se queda esperando si la cola está llena: se registra al momento y espera en segundo plano.
- El proxy y el límite de velocidad se aplican a las descargas nuevas sin reiniciar.
- Los cambios de metadata de una instancia no se pisan entre sí y no se puede cambiar el directorio de trabajo con descargas activas o pausadas.

### 7. Web pública sin cuentas
- Se eliminó el sistema de login/register/dashboard del sitio (antes conectado a la Accounts API en el Change-13): el sitio queda informativo con Inicio, Descarga, Historial y Acerca De.

---

## Errores de la auditoría resueltos en esta versión

Sin errores ni bugs nuevos en esta beta. Los 4 errores y 2 bugs de la beta anterior siguen corregidos (ver [Beta 2.5.0](./../StepLauncher-2.5.0/StepLauncher-Release-2.5.0.md)).

---

## Cambios técnicos (para curiosos y desarrolladores)

- `internal/Config`: `defaultRAMGB()` con `platform.TotalRAMMB()`, `LauncherConfig.RichPresence *bool` con `RichPresenceEnabled()` y `migrateRichPresence()` del bloque antiguo.
- `internal/Core/Accounts`: `AutoRefresh: true` por defecto + migración de archivos sin la clave (respeta el `false` explícito).
- Nuevo binding `SetScreenshotAsBackground` (`AppearanceService` → `Handlers/Background.go`) con validación de ruta dentro del workdir.
- Bindings regenerados (`wails3 generate bindings -ts -i`): `launcher.richPresence?: boolean | null`, `GetRichPresenceConfig(): boolean`, `SetScreenshotAsBackground`.
- Frontend: fallbacks de paleta unificados en `Ui.ts`, `Personalization.vue` y `_variables.scss`; splash estático y de Vue en `#0a0a0a`.

---

## Qué significa para el usuario

- **No necesitas hacer nada manual**: tu `launcher_config.json` se migra solo (colores inválidos → dark nuevo, bloque `richPresence` viejo → `launcher.richPresence`) y tu `launcher_accounts.json` sin la clave activa la renovación automática.
- **Si es tu primera vez**: el launcher arranca con la mitad de tu RAM y el tema oscuro nuevo.
- **Capturas**: abre Fotos, entra a una captura y usa "Colocar como fondo".

---

## Cómo actualizar a esta beta

- **Desde la beta anterior**: reemplaza el ejecutable por `StepLauncher-2.5.0-beta.3`; `directory.json` fuera del workdir conserva tu directorio.
- **Validación beta**: borrar `launcher_config.json` y abrir (debe crearse con mitad de RAM y colores dark), poner una captura de fondo, desactivar/reactivar Discord en Ajustes > General y comprobar la renovación de sesiones al iniciar.

## Notas beta

- Esta beta puede tener regresiones menores en la migración de configs muy antiguas. Reportar adjuntando `launcher_config.json` (sin datos sensibles) y el log de arranque.
