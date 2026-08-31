# Guía Maestra para Agentes de IA - StepLauncher

StepLauncher es un launcher de Minecraft construido sobre **Wails v3** (backend en Go) y **Vue 3 + TS** (frontend). Debido a su tamaño (más de 80 archivos acoplados), esta guía establece las directivas y el flujo maestro de desarrollo.

---

## Estructura Tecnológica y Versiones
- **Backend**: Go **1.26.4** (go.mod) + Wails **v3**.
- **Frontend**: Vue 3 + TS + Vite en `frontend/web/src`.
- **Gestor Frontend**: **Bun** (obligatorio usar `bun install` y `bun run build`).
- **Arquitectura Frontend**: Organización por **dominios** (feature-first): cada dominio (`Instances/`, `Accounts/`, `Settings/`, `Downloads/`…) es una carpeta que contiene sus componentes, su `Store.ts`, sus `Styles/` y composables propios. Lo transversal (reutilizable entre dominios) vive en `Common/` (componentes, stores, composables, estilos base y por tipo de componente). Los diálogos pequeños se abren a través del sistema global de overlays (`Common/Overlays/`), nunca importados dentro de otros overlays.
- **Verificación**: `go build ./...` en la raíz. Build completo: `wails build` (compila, genera bindings en `frontend/wailsjs` y empaqueta el binario).

## Flujo Maestro de Trabajo
Toda tarea ejecutada por un agente de IA debe seguir este orden estricto de razonamiento:
1. **Investigar antes de actuar**: Leer el código fuente y sus consumidores.
2. **Desarrollar y Validar**: Compilar con `go build ./...` en la raíz y ejecutar el type-check y build del frontend con `bun run build`.
3. **Registrar Cambios**: Documentar todo en la carpeta `Changelogs/` siguiendo el sistema de changelogs descrito en su `README.md`.

---

## Principios y Reglas Críticas

- **Concurrencia**: Prohibido bloquear el hilo principal o los bindings de Wails. No mantengas locks de `sync.RWMutex` en operaciones I/O o llamadas a callbacks. Evita a toda costa self-deadlocks.
- **Estilos de UI**: Los componentes de Vue no deben contener estilos inline complejos. Las hojas SCSS residen en `Styles/` de su dominio (o `Common/Styles/Components/` por tipo si son compartidas, `Common/Styles/base/` si son globales) e importarse con `@use`. Respeta el sistema de variables de personalización en `var(--color-*)`.
- **Imports Alias**: Usa siempre alias `@wailsjs/...` y `@/...` en lugar de rutas relativas largas en el frontend.
- **Git y Destrucción**: No ejecutes comandos destructivos de git (`reset`, `clean`, `checkout .`) para limpiar errores. Tampoco borres archivos sin permiso explícito del usuario. Las carpetas de `build/<plataforma>/` existentes (`windows`, `darwin`, `linux`, `android`) son requeridas por los includes del `Taskfile.yml` raíz: no se eliminan. El proyecto es **desktop-only (Windows/macOS/Linux)**: el soporte iOS se eliminó en 2.5.0 por decisión del owner, así que `build/ios/` NO existe y su include NO está en el `Taskfile.yml` — no hay que restaurarlo ni añadirlo (ver `Changelogs/Errors/StepLauncher-2.5.0/StepLauncher-Error-1.md`).
- **Comentarios y Mensajes**: Toda la comunicación y comentarios de código deben redactarse en **español**.