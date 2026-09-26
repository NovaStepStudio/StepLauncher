# Cambios de StepLauncher 2.5.0 (Step-18) — Acerca de con banner rectangular y hero rediseñado

- **Fecha**:   2026-09-25
- **Versión**: 2.5.0
- **Estado**:  implementado
- **Release**: StepLauncher-2.5.0 — en esta release se menciona que fue añadido.

## Qué cambió

### 1. Hero con el banner rectangular (`Settings/Sections/About.vue`, `Settings/Styles/About.scss`, `web/assets/banner.webp`)

- El "Acerca de" ahora abre con el banner rectangular oficial (`banner.webp`, "Tu Minecraft, sin rodeos") a ancho completo, con bordes redondeados y borde `var(--control-border)`.
- Sobre el banner va una franja inferior con degradado: logo, nombre, píldora de versión (`vX.Y.Z` del backend vía `useAppVersion`) y el lema "Tu Minecraft, sin rodeos.".
- Debajo: fila de sellos (Open source · GPL-3.0 · Hecho con Wails + Vue), texto de presentación y los botones Repositorio + Web oficial.
- Los logos (`logo-step.png`, `logo-wails.png`) y el banner se importan como módulos (`import ... from '../../../assets/...'`, patrón que ya usa `Welcome.vue`) en vez de rutas sueltas en el `src`.
- Se eliminaron las clases viejas del hero (`SsHero`, `SsHeroLogo`, `SsHeroTitle`, `SsHeroVersion`, `SsHeroLicense`); las nuevas (`SsAboutTop`, `SsBanner*`, `SsChips`, `SsLead`) siguen el patrón de las hojas existentes y respetan las variables de personalización (`var(--font-primary)`, `var(--font-size-primary)`, `var(--control-border)`).

### 2. Orden de grupos más útil

- El grupo "Cuentas" subió al segundo lugar (después del hero): es lo más accionable (gestionar cuentas, FAQ). Luego Recursos, Créditos y Legal. Sin cambios de contenido en esos grupos.

### 3. Presentación corregida tras la primera prueba visual

- El banner se mostraba recortado (`aspect-ratio: 3.4` + `object-fit: cover` cortaba costados y texto) y la franja de identidad superpuesta tapaba el "sin rodeos" del propio banner. Ahora la imagen se muestra completa (`width: 100%; height: auto`, sin recorte) y la identidad (logo + nombre + píldora de versión + lema) va debajo en una fila `SsIdentity` con colores del tema en vez de blanco fijo.
- Las acciones de "Cuentas" y "Créditos" pasaron de columna a fila (`flex-direction: row` con `wrap`): los botones quedan lado a lado y la nota ocupa toda la fila (`flex-basis: 100%`).

## Por qué

A pedido del owner: el owner agregó `banner.webp` a los assets para que pegue mejor en el "Acerca de", y el contenido anterior era mejorable visualmente.

## API afectada

Sin cambios en el backend Go ni bindings. Solo frontend.

## Comportamiento anterior/nuevo

- Antes: hero simple con logo grande centrado y versión en texto plano.
- Ahora: banner oficial completo sin recortes, con identidad debajo (logo + nombre + píldora de versión + lema), sellos, presentación y acciones.
- Revisión: la primera versión superponía la identidad sobre el banner y apilaba los botones; ahora no hay superposición y los botones van lado a lado.

## Cómo verificar

- Abrir Ajustes → Acerca de y comprobar banner, versión, botones y orden de grupos.
- `bun run build` en `launcher/frontend` y `go build ./...` en `launcher/` — pendientes: el owner pidió no compilar en esta tanda.
