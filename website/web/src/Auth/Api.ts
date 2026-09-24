// Cliente HTTP de la API StepLauncher (v1). Base desde `VITE_API_BASE_URL`
// con fallback local. Errores visibles genéricos en español; sin internals.

// Base sin barra final. No es secreta: es la URL pública de la API.
function leerBase(): string {
    const cruda = (import.meta.env.VITE_API_BASE_URL ?? '').trim().replace(/\/+$/, '');
    return cruda || 'https://steplauncher.stepnicka012.workers.dev';
}

export const API_BASE = leerBase();

export interface ApiSession {
    accessToken: string;
    refreshToken: string;
    expiresAt: number | null;
    tokenType: 'bearer';
}

export interface AuthUser {
    id: string;
    email: string | null;
    username: string | null;
}

export interface AuthPayload {
    user: AuthUser;
    session: ApiSession | null;
    emailConfirmationRequired?: boolean;
    message?: string;
}

export type ConfirmType = 'signup' | 'email_change' | 'recovery';

export interface ConfirmResult {
    user?: AuthUser;
    session?: ApiSession | null;
    recoveryVerified?: boolean;
    message?: string;
}

export interface Profile {
    id: string;
    email: string | null;
    username: string | null;
    displayName: string | null;
    avatarUrl: string | null;
    bio: string | null;
    lastMcVersion: string | null;
    mcUuid: string | null;
    bannerUrl: string | null;
    isOnline: boolean;
    createdAt: string | null;
    lastSessionAt: string | null;
}

export interface FileUpload {
    id: string;
    kind: string;
    sizeBytes: number;
    mime: string;
    createdAt: string;
}

export interface FilesPage {
    uploads: FileUpload[];
    limit: number;
    offset: number;
    dailyQuota: { limit: number; window: string; kinds: string[] };
}

// Error visible: mensaje genérico en español, apto para mostrar en la UI.
export class ApiError extends Error {
    readonly code: string;
    readonly status: number;

    constructor(code: string, message: string, status: number) {
        super(message);
        this.name = 'ApiError';
        this.code = code;
        this.status = status;
    }
}

interface ErrorBody {
    code?: unknown;
    message?: unknown;
    details?: unknown;
}

// Mapeo de códigos estables de la API a mensajes genéricos en español.
// Nada de internals: sin URLs, tokens, requestIds ni SQL en lo visible.
function mensajePara(code: string, details: unknown): string {
    switch (code) {
        case 'username_taken':
            return 'Ese nombre de usuario ya está en uso.';
        case 'email_taken':
            return 'Ese correo ya está registrado.';
        case 'weak_password':
            return 'Elegí una contraseña de al menos 8 caracteres.';
        case 'invalid_credentials':
            return 'Credenciales inválidas. Revisá tus datos e intentá de nuevo.';
        case 'email_not_confirmed':
            return 'Tenés que confirmar tu correo antes de entrar. Revisá tu bandeja.';
        case 'invalid_code':
            return 'Código inválido o expirado. Pedí un nuevo enlace.';
        case 'recovery_expired':
            return 'Ese enlace venció o ya se usó. Pedí uno nuevo desde Recuperar.';
        case 'invalid_refresh':
            return 'Tu sesión expiró. Iniciá sesión de nuevo.';
        case 'current_password_incorrect':
            return 'La contraseña actual no es correcta.';
        case 'same_email':
            return 'Ya estás usando ese correo.';
        case 'file_too_large':
            return 'El archivo supera el tamaño máximo permitido.';
        case 'invalid_file_type':
            return 'Ese tipo de archivo no está permitido.';
        case 'invalid_dimensions':
            return 'La imagen supera los 1920×1080 permitidos.';
        case 'user_not_found':
            return 'Jugador no encontrado o no disponible.';
        case 'cannot_add_self':
            return 'No podés añadirte a vos mismo.';
        case 'already_friends':
            return 'Ya son amigos.';
        case 'already_pending':
            return 'Ya hay una solicitud pendiente.';
        case 'profile_not_found':
            return 'Ese perfil no existe o es privado.';
        case 'user_blocked':
            return 'Desbloqueá a ese jugador primero.';
        case 'cannot_block':
            return 'No se puede bloquear a ese jugador.';
        case 'mc_uuid_taken':
            return 'Ese UUID de Minecraft ya está enlazado.';
        case 'upload_quota_exceeded':
            return 'Llegaste al límite diario de subidas. Probá de nuevo mañana.';
        case 'upload_failed':
            return 'No se pudo guardar el archivo. Probá más tarde.';
        case 'rate_limited':
            return 'Demasiados intentos. Esperá un momento y probá de nuevo.';
        case 'unauthorized':
            return 'Tu sesión no es válida. Iniciá sesión de nuevo.';
        case 'not_found':
            return 'No se encontró lo que buscabas.';
        case 'validation_error':
            return mensajeDeValidacion(details);
        default:
            return 'Servicio no disponible. Probá más tarde.';
    }
}

// La API devuelve `details` por campo (mensajes Zod, sin datos sensibles):
// se muestra el primero si existe, si no un genérico.
function mensajeDeValidacion(details: unknown): string {
    if (Array.isArray(details)) {
        const primero = details.find((d): d is { message?: unknown } => typeof d === 'object' && d !== null);
        const msg = primero && typeof primero.message === 'string' ? primero.message.trim() : '';
        if (msg) return msg;
    } else if (typeof details === 'object' && details !== null) {
        for (const valor of Object.values(details)) {
            if (typeof valor === 'string' && valor.trim()) return valor;
            if (Array.isArray(valor)) {
                const msg = valor.find((v): v is string => typeof v === 'string' && v.trim().length > 0);
                if (msg) return msg;
            }
        }
    }
    return 'Revisá los datos ingresados.';
}

function soloDesarrollo(...args: unknown[]) {
    if (import.meta.env.DEV) {
        // eslint-disable-next-line no-console
        console.debug('[api]', ...args);
    }
}

async function pedir<T>(ruta: string, init: RequestInit, token?: string): Promise<T> {
    let respuesta: Response;
    try {
        const headers = new Headers(init.headers);
        if (!(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
        if (token) headers.set('Authorization', `Bearer ${token}`);
        respuesta = await fetch(`${API_BASE}${ruta}`, { ...init, headers });
    } catch (err) {
        soloDesarrollo('fallo de red:', err);
        throw new ApiError('network_error', 'No se pudo conectar con el servicio. Revisá tu conexión.', 0);
    }

    let cuerpo: unknown = null;
    try {
        cuerpo = await respuesta.json();
    } catch (err) {
        soloDesarrollo('respuesta no JSON:', err);
    }

    if (typeof cuerpo === 'object' && cuerpo !== null && 'success' in cuerpo) {
        const sobre = cuerpo as { success: boolean; data?: T; error?: ErrorBody };
        if (sobre.success) return sobre.data as T;
        const code = typeof sobre.error?.code === 'string' ? sobre.error.code : 'internal_error';
        soloDesarrollo(ruta, '->', code);
        throw new ApiError(code, mensajePara(code, sobre.error?.details), respuesta.status);
    }

    soloDesarrollo(ruta, '-> respuesta inesperada:', respuesta.status);
    throw new ApiError('internal_error', mensajePara('internal_error', null), respuesta.status);
}

function json<T>(ruta: string, metodo: string, cuerpo?: unknown, token?: string): Promise<T> {
    return pedir<T>(ruta, { method: metodo, body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo) }, token);
}

function multipart<T>(ruta: string, campo: string, archivo: File, token: string): Promise<T> {
    const forma = new FormData();
    forma.append(campo, archivo);
    return pedir<T>(ruta, { method: 'POST', body: forma }, token);
}

// --- Auth: registro, login, sesiones ----------------------------------------

export function registrar(input: { email: string; username: string; password: string }): Promise<AuthPayload> {
    return json<AuthPayload>('/v1/auth/register', 'POST', input);
}

export function entrar(input: { identifier: string; password: string }): Promise<AuthPayload> {
    return json<AuthPayload>('/v1/auth/login', 'POST', input);
}

export function renovar(refreshToken: string): Promise<{ session: ApiSession }> {
    return json<{ session: ApiSession }>('/v1/auth/refresh', 'POST', { refreshToken });
}

export function salir(token: string): Promise<{ loggedOut: boolean }> {
    return json<{ loggedOut: boolean }>('/v1/auth/logout', 'POST', undefined, token);
}

// --- Verificación y recupero (plantillas Supabase) ------------------------------
// La API responde éxito genérico en resend/recover para no enumerar correos.

export function reenviarConfirmacion(input: { email: string; type?: ConfirmType }): Promise<{ resent: boolean; message: string }> {
    return json('/v1/auth/resend', 'POST', { type: 'signup', ...input });
}

export function pedirRecupero(email: string): Promise<{ recoverySent: boolean; message: string }> {
    return json('/v1/auth/recover', 'POST', { email });
}

export function confirmarCodigo(input: { email: string; token: string; type: ConfirmType }): Promise<ConfirmResult> {
    return json<ConfirmResult>('/v1/auth/confirm', 'POST', input);
}

export function fijarNuevaPassword(input: { email: string; token: string; newPassword: string }): Promise<{ passwordReset: boolean }> {
    return json('/v1/auth/reset-password', 'POST', input);
}

// Fijar contraseña con la sesión del enlace de recupero (Bearer del hash,
// sin email ni código: la sesión ya prueba el correo). No guarda sesión.
export function fijarPasswordRecupero(token: string, newPassword: string): Promise<{ passwordReset: boolean }> {
    return json('/v1/auth/recovery-password', 'POST', { newPassword }, token);
}

// --- Cuenta propia ------------------------------------------------------------

export function miPerfil(token: string): Promise<Profile> {
    return json<Profile>('/v1/accounts/me', 'GET', undefined, token);
}

export function actualizarPerfil(
    token: string,
    input: { username?: string; displayName?: string; bio?: string; lastMcVersion?: string },
): Promise<Profile> {
    return json<Profile>('/v1/accounts/me', 'PATCH', input, token);
}

export function pedirCambioEmail(token: string, newEmail: string): Promise<{ emailChangeRequested: boolean; message: string }> {
    return json('/v1/accounts/me/email-change', 'POST', { newEmail }, token);
}

export function cambiarContrasena(token: string, input: { currentPassword: string; newPassword: string }): Promise<{ passwordChanged: boolean }> {
    return json('/v1/accounts/me/password-change', 'POST', input, token);
}

export function subirAvatar(token: string, archivo: File): Promise<{ avatarUrl: string }> {
    return multipart('/v1/accounts/me/avatar', 'avatar', archivo, token);
}

export function subirBanner(token: string, archivo: File): Promise<{ bannerUrl: string }> {
    return multipart('/v1/accounts/me/banner', 'banner', archivo, token);
}

// --- Privacidad propia -----------------------------------------------------------

export interface Privacy {
    searchable: boolean;
    allowEmailSearch: boolean;
    receiveFriendRequests: boolean;
    receiveNotifications: boolean;
}

export function miPrivacidad(token: string): Promise<Privacy> {
    return json<Privacy>('/v1/accounts/me/privacy', 'GET', undefined, token);
}

export function actualizarPrivacidad(token: string, input: Partial<Privacy>): Promise<Privacy> {
    return json<Privacy>('/v1/accounts/me/privacy', 'PATCH', input, token);
}

// Presencia: solo automático (al entrar/salir), nunca manual desde la UI.
export function cambiarPresencia(token: string, isOnline: boolean): Promise<{ isOnline: boolean }> {
    return json('/v1/accounts/me/presence', 'PATCH', { isOnline }, token);
}

// --- Archivos: skins y capas (PNG, privados) -----------------------------------

export type KindArchivo = 'skin' | 'cape';

export function listarArchivos(token: string, limit = 20, offset = 0): Promise<FilesPage> {
    return json<FilesPage>(`/v1/files?limit=${limit}&offset=${offset}`, 'GET', undefined, token);
}

export function subirArchivo(token: string, kind: KindArchivo, archivo: File): Promise<FileUpload> {
    return multipart(`/v1/files/${kind}`, 'file', archivo, token);
}

export function urlArchivo(token: string, id: string): Promise<{ url: string; expiresIn: number }> {
    return json(`/v1/files/${encodeURIComponent(id)}/url`, 'GET', undefined, token);
}

export function borrarArchivo(token: string, id: string): Promise<{ deleted: boolean }> {
    return pedir(`/v1/files/${encodeURIComponent(id)}`, { method: 'DELETE' }, token);
}

// --- Notificaciones: bandeja propia --------------------------------------------

export interface NotificationItem {
    id: string;
    title: string;
    body: string;
    read: boolean;
    createdAt: string;
}

export interface NotificationsPage {
    notifications: NotificationItem[];
    unreadCount: number;
    limit: number;
    offset: number;
}

export function listarNotificaciones(token: string, limit = 10, offset = 0, unreadOnly = false): Promise<NotificationsPage> {
    return json<NotificationsPage>(
        `/v1/notifications?limit=${limit}&offset=${offset}&unreadOnly=${unreadOnly ? 'true' : 'false'}`,
        'GET',
        undefined,
        token,
    );
}

export function marcarLeida(token: string, id: string): Promise<NotificationItem> {
    return json(`/v1/notifications/${encodeURIComponent(id)}/read`, 'PATCH', undefined, token);
}

export function marcarTodasLeidas(token: string): Promise<{ markedRead: number }> {
    return json('/v1/notifications/read-all', 'POST', undefined, token);
}

export function borrarNotificacion(token: string, id: string): Promise<{ deleted: boolean }> {
    return pedir(`/v1/notifications/${encodeURIComponent(id)}`, { method: 'DELETE' }, token);
}

// --- Amigos: búsqueda, solicitudes y lista -----------------------------------------

export interface Jugador {
    username: string;
    displayName: string;
    avatarUrl: string | null;
    mcUuid: string | null;
    isOnline: boolean;
}

export interface MiniPerfil {
    userId: string;
    username: string;
    displayName: string;
    avatarUrl: string | null;
}

export function buscarJugadores(token: string, q: string, limit = 10): Promise<{ users: Jugador[] }> {
    return json(`/v1/users/search?q=${encodeURIComponent(q)}&limit=${limit}`, 'GET', undefined, token);
}

export interface Solicitud {
    id: string;
    direction: 'sent' | 'incoming';
    status: string;
    createdAt: string;
    respondedAt: string | null;
    user: MiniPerfil | null;
}

export function enviarSolicitud(token: string, identifier: string): Promise<{ id: string; status: string; toUser: MiniPerfil | null }> {
    return json('/v1/friends/requests', 'POST', { identifier }, token);
}

export function listarSolicitudes(token: string, type: 'incoming' | 'sent' | 'all' = 'incoming'): Promise<{ requests: Solicitud[] }> {
    return json(`/v1/friends/requests?type=${type}`, 'GET', undefined, token);
}

export function aceptarSolicitud(token: string, id: string): Promise<{ id: string; status: string; friend: MiniPerfil | null }> {
    return json(`/v1/friends/requests/${encodeURIComponent(id)}/accept`, 'POST', undefined, token);
}

export function rechazarSolicitud(token: string, id: string): Promise<{ id: string; status: string }> {
    return json(`/v1/friends/requests/${encodeURIComponent(id)}/decline`, 'POST', undefined, token);
}

export function cancelarSolicitud(token: string, id: string): Promise<{ id: string; status: string }> {
    return json(`/v1/friends/requests/${encodeURIComponent(id)}/cancel`, 'POST', undefined, token);
}

export interface Amigo {
    userId: string;
    username: string;
    displayName: string;
    avatarUrl: string | null;
    mcUuid: string | null;
    friendsSince: string;
    isOnline: boolean;
}

export function listarAmigos(token: string): Promise<{ friends: Amigo[]; count: number }> {
    return json('/v1/friends', 'GET', undefined, token);
}

export function romperAmistad(token: string, friendId: string): Promise<{ removed: boolean }> {
    return pedir(`/v1/friends/${encodeURIComponent(friendId)}`, { method: 'DELETE' }, token);
}

export interface Bloqueado {
    user: MiniPerfil | { userId: string };
    blockedSince: string;
}

export function listarBloqueos(token: string): Promise<{ blocked: Bloqueado[] }> {
    return json('/v1/friends/blocks', 'GET', undefined, token);
}

export function bloquear(token: string, identifier: string): Promise<{ blocked: boolean }> {
    return json('/v1/friends/blocks', 'POST', { identifier }, token);
}

export function desbloquear(token: string, userId: string): Promise<{ unblocked: boolean }> {
    return pedir(`/v1/friends/blocks/${encodeURIComponent(userId)}`, { method: 'DELETE' }, token);
}

// --- Comunidad: previsualización de perfiles ---------------------------------------

export interface PerfilPublico {
    userId: string;
    username: string;
    displayName: string;
    avatarUrl: string | null;
    bannerUrl: string | null;
    bio: string | null;
    memberSince: string;
    isOnline: boolean;
}

export interface CosmeticoEquipado {
    id: string;
    slug: string;
    name: string;
    kind: string;
    imageUrl: string | null;
}

export type EstadoRelacion = 'self' | 'friends' | 'pending_sent' | 'pending_received' | 'none';

export interface PreviewComunidad {
    blocked: false;
    profile: PerfilPublico;
    minecraft: { uuid: string | null; lastVersion: string | null };
    equippedCosmetics: CosmeticoEquipado[];
    relationship: { status: EstadoRelacion; requestId: string | null; friendsSince: string | null };
}

export interface AvisoBloqueo {
    blocked: true;
    reason: 'blocked_you' | 'blocked_by_you';
    message: string;
}

export function verPerfil(token: string, identifier: string): Promise<PreviewComunidad | AvisoBloqueo> {
    return json(`/v1/community/profiles/${encodeURIComponent(identifier)}`, 'GET', undefined, token);
}
