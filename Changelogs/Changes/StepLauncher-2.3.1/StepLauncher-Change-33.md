# Changes/StepLauncher-2.3.1/StepLauncher-Change-33.md

- **Fecha**: 2026-08-15
- **Versión**: 2.3.1
- **Release**: StepLauncher-2.3.1 — en este release se menciona que fue añadido.
- **Estado**: implementado y verificado (solo documentación; sin código afectado).

## Qué cambió

Actualización integral del `README.md` para alinearlo con la release 2.3.1 y darle un aspecto más profesional:

### 1. Badges ampliados

- La versión del badge pasa de 2.3.0 a **2.3.1**.
- Se añaden badges de tecnología (Go 1.26.4, Wails v2.13.0, Vue 3.5, TypeScript, Vite 8, bun) y de sistemas soportados (Windows | Linux | macOS), además de los existentes (descargas, estrellas, último commit, licencia).

### 2. Capturas del launcher

- El **menú principal** (`resources/MainMenu.png`) se muestra como imagen destacada justo debajo de los badges, arriba del todo.
- El resto de capturas viven en la sección **"Galería de capturas"** al final del README (justo antes del pie de página): variante del menú principal, onboarding de bienvenida (3), pantalla de juego, noticias, gestor de descargas (3), instancias (2) y personalización — todas las imágenes de `resources/`.

### 3. Sección "Novedades en la 2.3.1"

- Resumen de los cambios más destacados de la release: onboarding con modelo 3D, fix de lanzamiento (classpath/module path, deduplicación, RAM mínima 512 MB), soporte legacy de modloaders, modal de crash rediseñado, navegación por capas con ESC y la opción de lanzar Minecraft al terminar una instalación.
- Enlace al registro completo en `Changelogs/`.

### 4. Estructura del repositorio actualizada

- El árbol de `frontend/web/src` refleja la arquitectura **feature-first por dominios** (`Common/`, `Accounts/`, `Downloads/`, `Instances/`, `Launcher/`, `Settings/`, `Updates/`, `Crash/`, `Welcome/`, `News/`, `Screenshots/`, `Widgets/`, `Login/`, `Versions/`), sustituyendo el esquema antiguo (`Modals/`, `Stores/`, `Layouts/`).
- Se añade `resources/` al árbol.

## Por qué

El usuario pidió mejorar el README para que suene más profesional, mostrar capturas del launcher (ubicadas en `resources/`), añadir badges y actualizarlo a la versión 2.3.1 (que estaba desactualizado en 2.3.0).

## API afectada

- Sin cambios en bindings de Wails, backend ni frontend (solo `README.md`).

## Comportamiento anterior/nuevo

- **Antes**: README con versión 2.3.0, sin galería de capturas, con pocos badges y con una estructura de frontend desactualizada (carpetas `Modals/`, `Stores/`, `Layouts/` que ya no existen).
- **Ahora**: README alineado con 2.3.1, con el menú principal destacado arriba del todo, galería de capturas al final desde `resources/`, badges de versión/tecnología/sistemas, sección de novedades de la release y árbol de repositorio fiel a la arquitectura por dominios.

## Cómo verificar

- Visual: abrir `README.md` en GitHub o en un visor de Markdown y comprobar que las 13 capturas de `resources/` renderizan, los badges cargan y la estructura refleja las carpetas reales de `frontend/web/src`.
