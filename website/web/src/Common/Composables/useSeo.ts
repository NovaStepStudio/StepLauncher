// ============================================================================
// Composable useSeo — OpenGraph y metadatos dinámicos por ruta.
// Cada dominio (/, /download, /about, /changelog, /auth, ...) muestra su
// propio título, descripción, canónica y OpenGraph para que lo compartido
// coincida con lo que se está viendo. Los crawlers sin JS leen el index.html
// (valores de inicio = home); el router actualiza el DOM en cada navegación.
// ============================================================================

export const SITE_URL = 'https://steplauncher.pages.dev';
export const SITE_NAME = 'StepLauncher';
export const SITE_LOCALE = 'es_AR';
export const DEFAULT_IMAGE = `${SITE_URL}/og/banner.png`;
export const IMAGE_WIDTH = '1920';
export const IMAGE_HEIGHT = '1080';

export interface SeoData {
    title: string;
    description: string;
    // Ruta absoluta (https://...) de la imagen. Si se omite usa el banner.
    image?: string;
    // URL canónica absoluta. Si se omite se usa SITE_URL + path.
    url?: string;
    // Tipo OpenGraph: website en páginas, profile en perfiles, etc.
    type?: string;
    // Robots: 'index, follow' en públicas, 'noindex, nofollow' en privadas.
    robots?: string;
}

// Fija o crea un <meta name="...">.
function setMetaByName(name: string, content: string): void {
    if (typeof document === 'undefined') return;
    let tag = document.head.querySelector<HTMLMetaElement>(`meta[name="${name}"]`);
    if (!tag) {
        tag = document.createElement('meta');
        tag.setAttribute('name', name);
        document.head.appendChild(tag);
    }
    tag.setAttribute('content', content);
}

// Fija o crea un <meta property="..."> (OpenGraph).
function setMetaByProperty(property: string, content: string): void {
    if (typeof document === 'undefined') return;
    let tag = document.head.querySelector<HTMLMetaElement>(`meta[property="${property}"]`);
    if (!tag) {
        tag = document.createElement('meta');
        tag.setAttribute('property', property);
        document.head.appendChild(tag);
    }
    tag.setAttribute('content', content);
}

// Fija la canónica (<link rel="canonical">).
function setCanonical(url: string): void {
    if (typeof document === 'undefined') return;
    let link = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]');
    if (!link) {
        link = document.createElement('link');
        link.setAttribute('rel', 'canonical');
        document.head.appendChild(link);
    }
    link.setAttribute('href', url);
}

// Aplica título, descripción, canónica, robots y OpenGraph/Twitter.
export function applySeo(data: SeoData): void {
    if (typeof document === 'undefined') return;
    const url = data.url ?? SITE_URL;
    const image = data.image ?? DEFAULT_IMAGE;
    const type = data.type ?? 'website';
    const robots = data.robots ?? 'index, follow';

    document.title = data.title;
    setMetaByName('description', data.description);
    setMetaByName('robots', robots);
    setMetaByName('googlebot', robots);
    setCanonical(url);

    setMetaByProperty('og:type', type);
    setMetaByProperty('og:title', data.title);
    setMetaByProperty('og:description', data.description);
    setMetaByProperty('og:image', image);
    setMetaByProperty('og:url', url);
    setMetaByProperty('og:site_name', SITE_NAME);
    setMetaByProperty('og:locale', SITE_LOCALE);
    setMetaByProperty('og:image:secure_url', image);
    setMetaByProperty('og:image:type', 'image/png');
    setMetaByProperty('og:image:width', IMAGE_WIDTH);
    setMetaByProperty('og:image:height', IMAGE_HEIGHT);
    setMetaByProperty('og:image:alt', data.title);

    setMetaByName('twitter:card', 'summary_large_image');
    setMetaByName('twitter:title', data.title);
    setMetaByName('twitter:description', data.description);
    setMetaByName('twitter:image', image);
    setMetaByName('twitter:image:alt', data.title);
}
