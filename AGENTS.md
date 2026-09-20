# Guía raíz para IAs — StepLauncher-Wails3 (monorepo)

> **StepLauncher-Wails3 es un monorepo con subproyectos independientes.** Cada subproyecto se trabaja **individualmente**: un cambio en la web no modifica el launcher, y un cambio en el launcher no toca la web ni ningún otro subproyecto. Este archivo es el **enrutador**: dice qué carpeta corresponde a cada petición. Las reglas de detalle viven en el `AGENTS.md` de cada subproyecto.

Toda comunicación y comentarios de código deben redactarse en **español**.

---

## 1. Enrutador: qué carpeta toca cada petición (léelo primero, siempre)

Antes de leer código o modificar nada, clasifica la petición del usuario en una sola fila y quédate en esa carpeta:

| Si piden… | Trabaja SOLO en… | Lee primero… | Verificación obligatoria | ¿Changelog? |
|---|---|---|---|---|
| App de escritorio: jugar, descargar versiones, lanzar Minecraft, Java, modloaders, instancias, cuentas, música del launcher, personalización, bindings Wails, backend Go, crash, logs, updater, tray | `launcher/` | `launcher/AGENTS.md` y, si es bug/cambio, `launcher/Changelogs/README.md` | `go build ./...` dentro de `launcher/` + `bun run build` dentro de `launcher/frontend` | Sí, en `launcher/Changelogs/` |
| Web pública `steplauncher.pages.dev`: landing, descargas, changelog web, About, privacidad, términos, estilos o router de la web | `website/` | `website/AGENTS.md` | `bun run build` dentro de `website/` | No (la web no tiene `Changelogs/`) |
| API / backend futuro | `api/` (hoy vacía y **reservada**) | Nada más: pide permiso explícito antes de crear o modificar algo | Según lo que se autorice | No, avisar al owner |
| CI, releases, plantillas de issues/PR | `.github/` | El workflow o plantilla afectada | Validar el YAML afectado, sin builds | No |
| Docs del monorepo (este `README.md` / `AGENTS.md` raíz, `LICENSE.md`) | Raíz `/` | Este archivo + el `README.md` raíz | Ninguna (solo markdown) | No |

Si la petición mezcla dos filas (p. ej. "cambia el launcher y la web"), trátala como **dos tareas separadas**, una por subproyecto, y avísalo.

## 2. Regla de oro: aislamiento total entre subproyectos

1. **Un cambio = un subproyecto.** Solo crean, modifican o borran archivos dentro de la carpeta asignada por el enrutador (§1).
2. **Prohibido contaminar**: no toques, "de paso", ni el subproyecto vecino ni la raíz. Nada de "ya que estoy, ajusto también…".
3. **Leer fuera ≠ modificar fuera.** Puedes leer otro subproyecto para entender contexto (p. ej. qué versión muestra la web), pero sin modificarlo.
4. **Sin código compartido.** `launcher/` y `website/` no comparten imports, alias, bindings ni estilos. No crees dependencias cruzadas (`website/` nunca importa `@wailsjs` ni bindings de Wails; `launcher/` nunca importa de `website/`).
5. **Si algo parece exigir un cambio cruzado, PARA.** Avisa al usuario, pide permiso explícito y divídelo en dos cambios independientes.
6. **No pidas ni hagas reestructuraciones del monorepo** (mover carpetas, fusionar `package.json`/`go.mod`, unificar builds) sin orden explícita del owner.

## 3. Flujo maestro de trabajo

1. **Enrutar** (§1): una fila, una carpeta.
2. **Investigar dentro de esa carpeta**: lee su `AGENTS.md`, el código fuente y sus consumidores. Solo si es `launcher/`, consulta también `launcher/Changelogs/` antes de diagnosticar (causas raíz y reglas ya documentadas no se re-descubren).
3. **Desarrollar y validar** con los comandos del subproyecto (§4), nunca desde la raíz.
4. **Registrar**: solo `launcher/` genera entradas en `launcher/Changelogs/` siguiendo su `README.md`. `website/`, `api/` y `.github/` no generan changelogs.

## 4. Verificación por subproyecto (no mezclar)

No hay `go.mod` ni `package.json` en la raíz: **todo se ejecuta dentro del subproyecto**.

| Subproyecto | Dónde ejecutar | Comandos |
|---|---|---|
| `launcher/` backend | `launcher/` | `go build ./...` |
| `launcher/` frontend | `launcher/frontend` | `bun install` (solo si faltan deps) · `bun run build` (type-check + Vite, obligatorio) |
| `launcher/` app completa | `launcher/` | `wails3 dev` (dev) · `wails3 build` (binario en `bin/`) |
| `website/` | `website/` | `bun install` · `bun run build` (type-check + Vite, obligatorio) · `bun run dev` (dev) · `bun run deploy` (solo al desplegar a Pages) |
| `api/` | — | Nada (vacía y reservada) |

Gestor frontend **siempre Bun** (`bun install`, `bun run build`). **Prohibido npm** en todo el monorepo (más lento y riesgo de malware en paquetes; Bun valida al descargar).

## 5. Reglas globales (valen en todo el monorepo)

- **Español** en comunicación y comentarios de código.
- **Git y destrucción**: prohibidos `reset`, `clean`, `checkout .` para "limpiar errores". No borrar archivos sin permiso explícito del usuario.
- **SCSS obligatorio**: nada de CSS plano suelto ni estilos inline complejos. Cada componente usa `<style scoped lang="scss">`, hojas junto a su dominio/componente, importadas con `@use`, siguiendo el patrón de las hojas existentes. Variables `var(--color-*)`, alias `@/...` (y `@wailsjs/...` solo en `launcher/`).
- **Secrets**: nunca hardcodeados. En `website/` van en `.env` (local) + `wrangler.jsonc` (producción) con la misma variable (ver `website/AGENTS.md`).
- **Concurrencia Go/Wails** (solo `launcher/`): prohibido bloquear el hilo principal o los bindings; no mantener `sync.RWMutex` en I/O o callbacks; evitar self-deadlocks (ver `launcher/AGENTS.md`).
- **Desktop-only** (solo `launcher/`): Windows/macOS/Linux. `build/ios/` no existe ni debe restaurarse; las carpetas `build/<plataforma>/` existentes son requeridas por los includes del `Taskfile.yml` y no se eliminan.

## 6. Dónde vive cada cosa (chuleta para no perderse)

- `launcher/`: `main.go`, `go.mod`, `Taskfile.yml`, `internal/` (Config, Core, Handlers/Engine, Music, Services, Tray, Updater…), `frontend/` (`web/src/` por dominios: Accounts, Instances, Settings, Music… + `bindings/` generados), `resources/` (banners/capturas), `build/` (config Wails por plataforma), `bin/` (binarios), `Changelogs/` (memoria histórica del launcher), `README.md` + `AGENTS.md` propios.
- `website/`: `web/src/` (Home, Download, Changelog, About, Common…), `web/public/`, `vite.config.ts`, `wrangler.jsonc`, `dist/` (generado, no se edita a mano), `AGENTS.md` propio.
- `api/`: vacía, reservada para futuro. No tocar.
- `.github/`: `workflows/` (prerelease, publish-release, release), plantillas, `release.yml`, `settings.yml`.
- Raíz `/`: solo índice del monorepo (`README.md`, este `AGENTS.md`, `LICENSE.md`, `CODE_OF_CONDUCT.md`, `.gitignore`).

## 7. Errores típicos de IA en este repo (no cometer)

- Compilar o instalar desde la raíz (`go build ./...` o `bun install` en `/`): siempre dentro del subproyecto.
- Tocar `launcher/` cuando pidieron `website/` (o viceversa).
- Usar bindings Wails o `@wailsjs` en `website/`.
- Crear CSS plano, estilos inline complejos o inventar un sistema de estilos nuevo en vez de seguir el existente.
- Hardcodear tokens/sitekeys o exponer endpoints internos en la UI.
- Tocar `build/<plataforma>/`, sus includes del `Taskfile.yml`, o "restaurar" `build/ios/`.
- Documentar en `Changelogs/` cosas de `website/`/`api/`, o inventar entradas ficticias: los changelogs registran hechos reales del `launcher/` (ver `launcher/Changelogs/README.md`).
- Fragmentar una tarea en varios MD: **una tarea = una entrada** (solo en `launcher/Changelogs/`).
