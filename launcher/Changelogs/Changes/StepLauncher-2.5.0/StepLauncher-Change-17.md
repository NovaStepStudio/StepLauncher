# Cambios de StepLauncher 2.5.0 (Step-17) — Acerca de renovado: sin internos, con créditos, cuentas y legal

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Sección "Acerca de" renovada (`Settings/Sections/About.vue`, `Settings/Styles/About.scss`)

- Antes: enlaces desactualizados (documentación a un dominio viejo del estudio), lista de librerías de terceros cargada en el launcher con modal propio y una sección de "Componentes internos del launcher" que nombraba módulos del backend (`internal/Core/...`, `internal/Tray`, etc.).
- Ahora: sin nombres de componentes internos y sin listas pesadas con modal. Los créditos completos viven en la web (`https://steplauncher.pages.dev/credits`) y el launcher solo muestra un resumen por categorías (Tipografías, Librerías UI/UX, Nativo) con botones que abren la web vía `Browser.OpenURL`.
- Hero actualizado: descripción actual (versiones, modloaders, instancias y cuentas), botones "Ir al repositorio" y "Web oficial", sello de licencia GPL-3.0.
- Recursos actualizados: repositorio, web oficial, NovaStepStudio y Créditos a Terceros. Se eliminó el enlace a la documentación vieja.
- Grupo nuevo "Cuentas": resumen offline/online y texturas/sesiones, botón "Gestionar cuentas" que abre el gestor interno (`accountsOpen` de `Common/Overlays/Store`, z-index 100 por encima de ajustes) y botón "Preguntas frecuentes" a la web (`/faq`).
- Grupo nuevo "Legal": botones a la política de privacidad (`/privacy`) y términos (`/terms`) de la web. Los textos legales no se duplican en el launcher.
- Estilos nuevos en `About.scss` (`SsHeroActions`, `SsHeroLicense`, `SsCreditCats`, `SsCreditCat`, `SsCreditActions`, `SsCreditNote`) siguiendo el patrón de las hojas existentes; el resto reutiliza clases ya definidas (`SsLinks`, `SsLink`, `SsCredit`).

## Por qué

A pedido del owner: el "Acerca de" estaba ultra desactualizado, no debe nombrar componentes internos y los créditos deben ser como la página de Créditos a Terceros del website, solo con botones que redireccionen a la web en vez de cargar las librerías.

## API afectada

Sin cambios en el backend Go ni bindings. Solo frontend: se usa el flag reactivo `accountsOpen` ya existente y `Browser.OpenURL` del runtime de Wails.

## Comportamiento anterior/nuevo

- Antes: ver créditos desplegaba listas internas/externas con modal dentro del launcher.
- Ahora: ver créditos abre `https://steplauncher.pages.dev/credits` en el navegador; gestionar cuentas abre el modal de cuentas; lo legal abre `/privacy` y `/terms` de la web.

## Cómo verificar

- `bun run build` en `launcher/frontend` (type-check + Vite) y `go build ./...` en `launcher/` — pendientes: el owner pidió no compilar en esta tanda, así que la verificación queda a cargo de la próxima sesión o del owner antes de la release.
