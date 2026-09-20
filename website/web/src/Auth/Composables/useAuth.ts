// Sesión global (singleton). Tokens en `localStorage`, nunca en URLs/logs/UI.

import { computed, ref } from 'vue';
import {
    API_BASE,
    ApiError,
    cambiarPresencia,
    entrar,
    miPerfil,
    registrar,
    renovar,
    salir,
    type ApiSession,
    type AuthUser,
    type Profile,
} from '@/Auth/Api';

const CLAVE = 'steplauncher.auth.v1';
// Margen para renovar antes de que venza el access token.
const MARGEN_SEGUNDOS = 60;

const session = ref<ApiSession | null>(null);
const user = ref<AuthUser | null>(null);
const profile = ref<Profile | null>(null);
const listo = ref(false);
const ocupado = ref(false);

let inicio: Promise<void> | null = null;
let avisoCierreRegistrado = false;

// Marca presencia sin bloquear: el usuario nunca lo cambia a mano.
async function ponerEnLinea(token: string, enLinea: boolean): Promise<void> {
    try {
        await cambiarPresencia(token, enLinea);
    } catch {
        // Mejor esfuerzo: la sesión vale igual sin presencia.
    }
}

// Al cerrar la pestaña no se puede esperar respuesta: envío con keepalive.
function avisarCierre(): void {
    const token = session.value?.accessToken;
    if (!token) return;
    try {
        fetch(`${API_BASE}/v1/accounts/me/presence`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
            body: JSON.stringify({ isOnline: false }),
            keepalive: true,
        }).catch(() => undefined);
    } catch {
        // Sin red no hay nada que avisar.
    }
}

function registrarAvisoCierre(): void {
    if (avisoCierreRegistrado || typeof window === 'undefined') return;
    avisoCierreRegistrado = true;
    window.addEventListener('pagehide', avisarCierre);
}

interface SesionGuardada {
    accessToken: string;
    refreshToken: string;
    expiresAt: number | null;
}

function guardar() {
    try {
        if (!session.value) {
            localStorage.removeItem(CLAVE);
            return;
        }
        const plana: SesionGuardada = {
            accessToken: session.value.accessToken,
            refreshToken: session.value.refreshToken,
            expiresAt: session.value.expiresAt,
        };
        localStorage.setItem(CLAVE, JSON.stringify(plana));
    } catch {
        // Almacenamiento lleno o bloqueado: la sesión queda solo en memoria.
    }
}

function leerGuardada(): SesionGuardada | null {
    try {
        const crudo = localStorage.getItem(CLAVE);
        if (!crudo) return null;
        const datos = JSON.parse(crudo) as Partial<SesionGuardada>;
        if (typeof datos.accessToken !== 'string' || typeof datos.refreshToken !== 'string') return null;
        return {
            accessToken: datos.accessToken,
            refreshToken: datos.refreshToken,
            expiresAt: typeof datos.expiresAt === 'number' ? datos.expiresAt : null,
        };
    } catch {
        return null;
    }
}

function vencida(s: SesionGuardada | ApiSession): boolean {
    if (s.expiresAt === null || s.expiresAt === undefined) return false;
    const ahoraSeg = Math.floor(Date.now() / 1000);
    return s.expiresAt <= ahoraSeg + MARGEN_SEGUNDOS;
}

function limpiar() {
    session.value = null;
    user.value = null;
    profile.value = null;
    guardar();
}

function tomarSesion(nueva: ApiSession) {
    session.value = nueva;
    guardar();
}

async function cargarPerfil(): Promise<void> {
    if (!session.value) return;
    try {
        profile.value = await miPerfil(session.value.accessToken);
        if (profile.value.email !== null && profile.value.email !== undefined) {
            user.value = {
                id: profile.value.id,
                email: profile.value.email,
                username: profile.value.username,
            };
        }
    } catch {
        // Sin perfil no se bloquea la sesión: el dashboard reintenta al montar.
        profile.value = null;
    }
}

async function hacerRefresh(): Promise<boolean> {
    const actual = session.value;
    if (!actual) return false;
    try {
        const { session: nueva } = await renovar(actual.refreshToken);
        tomarSesion(nueva);
        return true;
    } catch {
        limpiar();
        return false;
    }
}

// Token vigente: renueva primero si está por vencer.
async function tokenVigente(): Promise<string | null> {
    const actual = session.value;
    if (!actual) return null;
    if (vencida(actual)) {
        const ok = await hacerRefresh();
        if (!ok) return null;
    }
    return session.value?.accessToken ?? null;
}

export function useAuth() {
    const autenticado = computed(() => session.value !== null);
    const nombre = computed(() => profile.value?.displayName || profile.value?.username || user.value?.username || '');

    // Inicialización única: restaura la sesión guardada y la valida.
    function init(): Promise<void> {
        if (inicio) return inicio;
        inicio = (async () => {
            registrarAvisoCierre();
            const guardada = leerGuardada();
            if (guardada) {
                tomarSesion({ ...guardada, tokenType: 'bearer' });
                if (vencida(guardada)) {
                    await hacerRefresh();
                }
                if (session.value) {
                    await cargarPerfil();
                    await ponerEnLinea(session.value.accessToken, true);
                }
            }
            listo.value = true;
        })();
        return inicio;
    }

    function esperarLista(): Promise<void> {
        return init();
    }

    async function login(identifier: string, password: string): Promise<void> {
        ocupado.value = true;
        try {
            const res = await entrar({ identifier: identifier.trim(), password });
            user.value = res.user;
            tomarSesion(res.session);
            await cargarPerfil();
            await ponerEnLinea(res.session.accessToken, true);
        } finally {
            ocupado.value = false;
        }
    }

    async function register(email: string, username: string, password: string): Promise<void> {
        ocupado.value = true;
        try {
            const res = await registrar({ email: email.trim().toLowerCase(), username: username.trim(), password });
            user.value = res.user;
            tomarSesion(res.session);
            await cargarPerfil();
            await ponerEnLinea(res.session.accessToken, true);
        } finally {
            ocupado.value = false;
        }
    }

    async function logout(): Promise<void> {
        const token = session.value?.accessToken;
        if (token) {
            await ponerEnLinea(token, false);
        }
        limpiar();
        if (token) {
            try {
                await salir(token);
            } catch {
                // Ya se borró lo local: revocar es mejor esfuerzo.
            }
        }
    }

    // Ejecuta una petición autenticada con reintento único ante 401:
    // renueva la sesión una vez y reintenta; si sigue fallando, cierra.
    async function conAuth<T>(llamada: (token: string) => Promise<T>): Promise<T> {
        const token = await tokenVigente();
        if (!token) {
            throw new ApiError('unauthorized', 'Tu sesión no es válida. Iniciá sesión de nuevo.', 401);
        }
        try {
            return await llamada(token);
        } catch (err) {
            if (err instanceof ApiError && err.status === 401) {
                const ok = await hacerRefresh();
                if (!ok) {
                    throw new ApiError('unauthorized', 'Tu sesión expiró. Iniciá sesión de nuevo.', 401);
                }
                const nuevo = session.value?.accessToken;
                if (!nuevo) {
                    throw new ApiError('unauthorized', 'Tu sesión expiró. Iniciá sesión de nuevo.', 401);
                }
                try {
                    return await llamada(nuevo);
                } catch (reintento) {
                    if (reintento instanceof ApiError && reintento.status === 401) limpiar();
                    throw reintento;
                }
            }
            throw err;
        }
    }

    async function refrescarPerfil(): Promise<void> {
        await conAuth((token) => miPerfil(token)).then((p) => {
            profile.value = p;
        });
    }

    return {
        session,
        user,
        profile,
        listo,
        ocupado,
        autenticado,
        nombre,
        init,
        esperarLista,
        login,
        register,
        logout,
        conAuth,
        cargarPerfil,
        refrescarPerfil,
    };
}
