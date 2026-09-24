<div align="center">

<a href="https://steplauncher.pages.dev">
  <img src="../branding/banner-002.png" alt="StepLauncher — Banner oficial | Powered by NovaCore Engine" width="100%">
</a>

<img src="../branding/appicon-neon.png" alt="StepLauncher" width="110">

# StepLauncher Website

**Web pública e informativa de StepLauncher — presentación, descargas, historial, comunidad y cuenta del jugador, creada por NovaStepStudio**

Vue 3 + TypeScript + Vite &nbsp;·&nbsp; desplegada como sitio estático en Cloudflare Pages

> 🌐 **Web en producción:** [**steplauncher.pages.dev**](https://steplauncher.pages.dev) — landing, descargas por plataforma, changelog desde GitHub, comunidad, panel de cuenta, privacidad y términos.

[![Web](https://img.shields.io/badge/Web-steplauncher.pages.dev-31b3ff?style=for-the-badge&logo=cloudflare&logoColor=white)](https://steplauncher.pages.dev)
[![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?style=for-the-badge&logo=vuedotjs&logoColor=white)](https://vuejs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![Vite](https://img.shields.io/badge/Vite-8-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vitejs.dev)
[![bun](https://img.shields.io/badge/bun-f9f1e1?style=for-the-badge&logo=bun&logoColor=black)](https://bun.sh)
[![Licencia](https://img.shields.io/badge/Licencia-GPL--3.0-a42e2e?style=for-the-badge&label=Licencia)](https://github.com/NovaStepStudio/StepLauncher/blob/main/LICENSE.md)

</div>

---

## 📖 Sobre

La web es el escaparate público del proyecto: cuenta qué es StepLauncher, ofrece la descarga correcta según tu sistema, muestra el historial real de releases y aloja la cuenta del jugador (entrar, crear cuenta, recuperar contraseña, panel y comunidad). Todo el texto visible está en **español** y cada dominio actualiza su propio SEO (título, descripción, canónica y OpenGraph) para que lo compartido coincida con lo que se ve.

Reglas IA del subproyecto: [`AGENTS.md`](AGENTS.md).

## ✨ Qué incluye (por dominio)

### 🏠 Home (`web/src/Home/`) — `/`
- Hero, instancias, mods, música, rendimiento, personalización, open source y llamada a descargar.

### ⬇️ Download (`web/src/Download/`) — `/download`
- Hero + release **estable** y **prereleases** (beta/alpha) con assets agrupados por plataforma (Windows/macOS/Linux), tamaño, arquitectura y tipo (instalador/portable).
- Datos en vivo de GitHub (`useReleases`) con fallback al worker propio si la API falla.

### 📜 Changelog (`web/src/Changelog/`) — `/changelog`
- Historial completo con markdown original de cada release, filtros por canal (todas/estables/betas/alphas), buscador y paginación de 5 por página.

### 👤 Cuenta (`web/src/Auth/`) — `/auth`, `/auth/callback`, `/dashboard`
- Página única de acceso con pestañas sincronizadas por `?tab=` (entrar, crear cuenta, recuperar).
- Panel con pestañas por `?tab=` (perfil, seguridad, cosméticos con render 3D de skins/capas, amigos, notificaciones).
- Guard de ruta: el panel exige sesión (restaura la guardada antes de montar para no parpadear el header) y login/registro redirigen al panel si ya hay sesión.

### 🤝 Comunidad (`web/src/Community/`) — `/community/account/:id?`
- Buscador de jugadores y previsualización pública del perfil (portada, datos, Minecraft, cosméticos equipados en 3D, amistad/bloqueos). Ruta protegida.

### 🏢 About (`web/src/About/`) — `/about`
- Proyecto, cómo funciona, stack, ecosistema y estudio.

### 🎨 Branding (`web/src/Branding/`) — `/branding`
- Kit oficial de marca: banner e iconos en alta resolución con descarga directa (archivos en `web/public/branding/`, copia de `branding/` de la raíz).

### ⚖️ Legal — `/privacy`, `/terms`
- Política de privacidad y términos en español, sin letra chica.

### 🧩 Transversal (`web/src/Common/`)
- `Header` / `Footer`, composables `useSeo` (metadatos por ruta) y `useReleases` (releases + estrellas/forks), directiva `v-reveal` para aparición suave.

## 🧱 Stack tecnológico

| Capa | Tecnología |
|------|------------|
| UI | Vue 3.5, Vue Router 5, Tabler Icons, SCSS (`<style scoped lang="scss">`, sin CSS plano ni inline complejo) |
| 3D | `skinview3d` para skins y capas (prohibido reinventar renderer o usar 2D plano) |
| Markdown | `marked` para notas de releases |
| Datos | API de GitHub + worker propio (`VITE_API_BASE_URL`) + Supabase vía `api/` |
| Build | Vite 8 + `vue-tsc`, alias `@/...` (nunca rutas relativas largas) |
| Despliegue | Cloudflare Pages (`steplauncher`), SEO por ruta + `sitemap.xml` + `robots.txt` + `_redirects` |

## 📁 Estructura

```
website/
├── web/
│   ├── index.html        # SEO base (home), OG/Twitter con /og/banner.png, favicon
│   ├── index.css
│   ├── public/           # og/banner.png, branding/*.png, robots.txt, sitemap.xml, _redirects
│   ├── assets/           # logos, fondos, decoraciones, modloaders, gifs
│   └── src/
│       ├── Home/         # landing por secciones
│       ├── Download/     # estable + prereleases + pasos
│       ├── About/        # proyecto, stack, ecosistema, estudio
│       ├── Branding/     # kit oficial: banner e iconos con descarga
│       ├── Changelog/    # historial desde GitHub
│       ├── Auth/         # acceso, callback, panel (Dashboard), Api, validación, firmadas
│       ├── Community/    # buscador + perfil público
│       ├── Privacy/ Terms/  # legal
│       ├── Common/       # Header, Footer, Composables (useSeo, useReleases...), estilos
│       ├── App.vue / Main.ts / Router.ts
├── vite.config.ts        # root ./web, alias @ -> web/src
├── wrangler.jsonc        # config NO secreta de Pages
├── .env.example          # plantilla de VITE_API_BASE_URL
├── AGENTS.md             # reglas IA de la web
└── README.md             # estás aquí
```

> La web no tiene `Changelogs/`: el historial público vive en [/changelog](https://steplauncher.pages.dev/changelog) (datos de GitHub) y la memoria del launcher vive en `launcher/Changelogs/`.

## 🔐 Variables y secretos

Todo secreto o token va en `.env` (local) y en `wrangler.jsonc` (producción) con la **misma variable**. Nunca hardcodear secrets, sitekeys ni endpoints internos en el código; los errores visibles son mensajes genéricos en español.

```bash
# Desarrollo / producción (sin barra final)
VITE_API_BASE_URL="https://steplauncher.<tu-subdominio>.workers.dev"
```

## 🚀 Desarrollo

```powershell
cd website
bun install                       # bun obligatorio, prohibido npm
bun run build                     # type-check + build Vite (verificación obligatoria)
bun run dev                       # desarrollo local
bun run deploy                    # build + despliegue a Cloudflare Pages
```

> Todo lo nuevo o modificado debe funcionar en **móvil y tablet** (layouts fluidos, sin scroll horizontal, táctil usable) con los breakpoints que ya usa el proyecto. Si algo no puede ser responsive, se avisa y justifica.

---

<div align="center">

**NovaStepStudio** — Santiago Stepnicka

[🌐 Página principal](https://steplauncher.pages.dev) · [GitHub](https://github.com/NovaStepStudio) · [Repositorio](https://github.com/NovaStepStudio/StepLauncher)

<sub>© 2026 NovaStepStudio — Powered by NovaCore Engine · No afiliado a Mojang Studios ni a Microsoft.</sub>

</div>
