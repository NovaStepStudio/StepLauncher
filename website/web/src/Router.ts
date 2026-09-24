import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router';
import HomeIndex from './Home/Index.vue';
import DownloadIndex from './Download/Index.vue';
import AboutIndex from './About/Index.vue';
import FaqIndex from './Faq/Index.vue';
import ChangelogIndex from './Changelog/Index.vue';
import BrandingIndex from './Branding/Index.vue';
import PrivacyIndex from './Privacy/Index.vue';
import TermsIndex from './Terms/Index.vue';
import AuthIndex from './Auth/Index.vue';
import AuthCallbackIndex from './Auth/Callback/Index.vue';
import DashboardIndex from './Auth/Dashboard/Index.vue';
import CuentaIndex from './Community/Index.vue';
import { useAuth } from './Auth/Composables/useAuth';
import { applySeo, DEFAULT_IMAGE, SITE_URL, type SeoData } from '@/Common/Composables/useSeo';

declare module 'vue-router' {
    interface RouteMeta {
        seo?: SeoData;
        requiresAuth?: boolean;
        guest?: boolean;
    }
}

// SEO base por dominio: título y descripción que coinciden con lo visible.
// La URL canónica se resuelve en seoForRoute para incluir :id cuando hay.
const SEO_HOME: SeoData = {
    title: 'StepLauncher - Launcher de Minecraft Java',
    description: 'Launcher moderno, rápido y multiplataforma para Minecraft: Java Edition. Gestioná versiones, modloaders, instancias y cuentas con Yggdrasil desde una interfaz limpia.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/`,
};
const SEO_DOWNLOAD: SeoData = {
    title: 'Descargar StepLauncher - Windows, Linux y macOS',
    description: 'Descargá la última versión estable de StepLauncher directo de GitHub para Windows, Linux y macOS. Si te gusta probar lo nuevo, también tenés betas y alphas.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/download`,
};
const SEO_ABOUT: SeoData = {
    title: 'Acerca de StepLauncher - Proyecto open source de NovaStepStudio',
    description: 'StepLauncher nació para jugar Minecraft: Java Edition sin vueltas: rápido, personalizable y open source con licencia GPL-3.0, hecho con Wails + Vue por NovaStepStudio.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/about`,
};
const SEO_FAQ: SeoData = {
    title: 'Preguntas frecuentes - StepLauncher',
    description: 'Respuestas cortas sobre StepLauncher: descargas oficiales, mods, cuentas con Yggdrasil, modo offline y online, seguridad y datos.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/faq`,
};
const SEO_CHANGELOG: SeoData = {
    title: 'Historial de cambios - StepLauncher',
    description: 'Todas las versiones de StepLauncher: estables, betas y alphas con sus notas, fechas y descargas desde GitHub.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/changelog`,
};
const SEO_BRANDING: SeoData = {
    title: 'Branding oficial - StepLauncher',
    description: 'El kit oficial de marca de StepLauncher: banner e iconos en alta resolución, listos para descargar.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/branding`,
};
const SEO_PRIVACY: SeoData = {
    title: 'Política de privacidad - StepLauncher',
    description: 'Qué datos guarda StepLauncher, para qué los usa y cómo pedir que se borren.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/privacy`,
};
const SEO_TERMS: SeoData = {
    title: 'Términos y condiciones - StepLauncher',
    description: 'Las reglas para usar StepLauncher y tu cuenta: licencia, cuentas, uso aceptable y responsabilidad.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/terms`,
};
const SEO_LOGIN: SeoData = {
    title: 'Entrar a StepLauncher',
    description: 'Entrá a tu cuenta de StepLauncher para gestionar tu perfil, cosméticos, amigos y notificaciones.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/auth`,
    robots: 'noindex, nofollow',
};
const SEO_REGISTER: SeoData = {
    title: 'Crear cuenta en StepLauncher',
    description: 'Creá tu cuenta gratis de StepLauncher para personalizar tu perfil con skins y capas en 3D, sumar amigos y más.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/auth`,
    robots: 'noindex, nofollow',
};
const SEO_RECOVERY: SeoData = {
    title: 'Recuperar contraseña - StepLauncher',
    description: 'Pedí un enlace a tu correo para elegir una contraseña nueva de StepLauncher.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/auth`,
    robots: 'noindex, nofollow',
};
const SEO_CALLBACK: SeoData = {
    title: 'Confirmar correo - StepLauncher',
    description: 'Confirmá tu correo o restablecé tu contraseña de StepLauncher desde el enlace que te llegó.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/auth/callback`,
    robots: 'noindex, nofollow',
};
const SEO_DASHBOARD: SeoData = {
    title: 'Mi panel - StepLauncher',
    description: 'Tu panel de StepLauncher: perfil, seguridad, cosméticos en 3D, amigos y notificaciones.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/dashboard`,
    robots: 'noindex, nofollow',
};
const SEO_COMMUNITY: SeoData = {
    title: 'Comunidad StepLauncher - Buscar jugadores',
    description: 'Buscá jugadores de la comunidad StepLauncher y mirá sus perfiles, cosméticos y datos de Minecraft.',
    image: DEFAULT_IMAGE,
    url: `${SITE_URL}/community/account`,
    robots: 'noindex, nofollow',
};

// Sufijos del panel según ?tab= para que la pestaña coincida con el título.
const PANEL_TITLES: Record<string, string> = {
    perfil: 'Mi panel · Perfil - StepLauncher',
    seguridad: 'Mi panel · Seguridad - StepLauncher',
    cosmeticos: 'Mi panel · Cosméticos - StepLauncher',
    amigos: 'Mi panel · Amigos - StepLauncher',
    notificaciones: 'Mi panel · Notificaciones - StepLauncher',
};

// Resuelve el SEO final según ruta, query (?tab=) y params (:id).
function seoForRoute(to: RouteLocationNormalized): SeoData {
    if (to.name === 'auth') {
        if (to.query.tab === 'register') return SEO_REGISTER;
        if (to.query.tab === 'recovery') return SEO_RECOVERY;
        return SEO_LOGIN;
    }
    if (to.name === 'auth-callback') {
        return SEO_CALLBACK;
    }
    if (to.name === 'dashboard') {
        const tab = typeof to.query.tab === 'string' ? to.query.tab : '';
        const title = PANEL_TITLES[tab] ?? SEO_DASHBOARD.title;
        return { ...SEO_DASHBOARD, title };
    }
    if (to.name === 'cuenta') {
        const raw = to.params.id;
        const id = (Array.isArray(raw) ? (raw[0] ?? '') : (raw ?? '')).trim().replace(/^@/, '');
        if (!id) return SEO_COMMUNITY;
        return {
            title: `Perfil de @${id} en StepLauncher`,
            description: `Mirá el perfil de @${id} en la comunidad StepLauncher: cosméticos, datos de Minecraft y más.`,
            image: DEFAULT_IMAGE,
            url: `${SITE_URL}${to.path}`,
            type: 'profile',
            robots: 'noindex, nofollow',
        };
    }
    return to.meta.seo ?? SEO_HOME;
}

const Router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', name: 'home', component: HomeIndex, meta: { seo: SEO_HOME } },
        { path: '/download', name: 'download', component: DownloadIndex, meta: { seo: SEO_DOWNLOAD } },
        { path: '/about', name: 'about', component: AboutIndex, meta: { seo: SEO_ABOUT } },
        { path: '/faq', name: 'faq', component: FaqIndex, meta: { seo: SEO_FAQ } },
        { path: '/changelog', name: 'changelog', component: ChangelogIndex, meta: { seo: SEO_CHANGELOG } },
        { path: '/branding', name: 'branding', component: BrandingIndex, meta: { seo: SEO_BRANDING } },
        { path: '/privacy', name: 'privacy', component: PrivacyIndex, meta: { seo: SEO_PRIVACY } },
        { path: '/terms', name: 'terms', component: TermsIndex, meta: { seo: SEO_TERMS } },
        { path: '/login', redirect: { path: '/auth', query: { tab: 'login' } } },
        { path: '/register', redirect: { path: '/auth', query: { tab: 'register' } } },
        { path: '/auth', name: 'auth', component: AuthIndex, meta: { guest: true, seo: SEO_LOGIN } },
        { path: '/auth/callback', name: 'auth-callback', component: AuthCallbackIndex, meta: { seo: SEO_CALLBACK } },
        { path: '/dashboard', name: 'dashboard', component: DashboardIndex, meta: { requiresAuth: true, seo: SEO_DASHBOARD } },
        { path: '/community/account/:id?', name: 'cuenta', component: CuentaIndex, meta: { requiresAuth: true, seo: SEO_COMMUNITY } },
        {
            path: '/jugador/:id',
            redirect: (to) => {
                const id = Array.isArray(to.params.id) ? (to.params.id[0] ?? '') : (to.params.id ?? '');
                return { path: `/community/account/${id}` };
            },
        },
    ],
    scrollBehavior() {
        return { top: 0 };
    },
});

// Guard de cuenta: el panel exige sesión (restaura la guardada primero);
// login/registro redirigen al panel si ya hay sesión.
Router.beforeEach(async (to) => {
    const auth = useAuth();
    await auth.esperarLista();
    if (to.meta.requiresAuth && !auth.autenticado.value) {
        return { path: '/login', query: { next: to.fullPath } };
    }
    if (to.meta.guest && auth.autenticado.value) {
        return { path: '/dashboard' };
    }
    return true;
});

// En cada navegación se actualizan título, descripción, canónica y
// OpenGraph/Twitter para que coincidan con el dominio visible.
Router.afterEach((to) => {
    applySeo(seoForRoute(to));
});

export default Router;
