import { ref, computed } from 'vue';
import { EventsOn } from '@wailsjs/runtime/runtime';

export type AccountType = 'offline' | 'authlib';

export const ACCOUNT_LOGIN_START_EVENT = 'account-login-start';

export interface AuthLoginResult {
    ok: boolean;
    error: string;
}

let loginResultPromise: Promise<AuthLoginResult> | null = null;

export interface AccountInfo {
    id: string;
    name: string;
    type: AccountType;
    username: string;
    uuid: string;
    authServerUrl?: string;
    serverName?: string;
    hasToken: boolean;
    sessionValid: boolean;
    createdAt: string;
    lastUsed?: string;
    customProperties?: Record<string, string>;
}

export interface CreateAccountReq {
    type: AccountType;
    name?: string;
    username: string;
    accessToken?: string;
    authServerUrl?: string;
    uuid?: string;
}

export interface AuthlibLoginReq {
    authServerUrl: string;
    username: string;
    password: string;
}

export const accounts = ref<AccountInfo[]>([]);
export const selectedAccountId = ref('');
export const autoRefresh = ref(false);

export const accountAvatars = ref<Record<string, string>>({});

let assetsOff: (() => void) | null = null;

function ensureAssetsListener() {
    if (assetsOff) return;
    assetsOff = EventsOn('account_assets', (data: any) => {
        let payload: any = data;
        if (typeof data === 'string') {
            try {
                payload = JSON.parse(data);
            } catch {
                payload = {};
            }
        }
        if (payload?.id && payload?.assets?.avatarDataUrl) {
            accountAvatars.value = { ...accountAvatars.value, [payload.id]: payload.assets.avatarDataUrl };
        }
    });
}

export function fetchAccountAvatar(id: string): void {
    if (!id) return;
    if (accountAvatars.value[id]) return;
    ensureAssetsListener();
    try {
        goApp()?.GetAccountAssets?.(id);
    } catch (e) {
        console.error('[Accounts] no se pudo pedir el avatar de', id, e);
    }
}

const goApp = () => (window as any)?.go?.main?.App;

export async function loadAccounts(): Promise<void> {
    try {
        const list = await goApp()?.ListAccounts?.();
        if (Array.isArray(list)) accounts.value = list;
    } catch { }
    try {
        const sel = await goApp()?.GetSelectedAccount?.();
        if (typeof sel === 'string') selectedAccountId.value = sel;
    } catch { }
    try {
        const ar = await goApp()?.GetAccountsAutoRefresh?.();
        autoRefresh.value = ar === true;
    } catch { }
    for (const a of accounts.value) {
        if (a.type === 'authlib') fetchAccountAvatar(a.id);
    }
}

export async function createAccount(req: CreateAccountReq): Promise<string> {
    try {
        const created = await goApp()?.CreateAccount?.(req);
        if (created?.id) selectedAccountId.value = created.id;
        await loadAccounts();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo crear la cuenta.';
    }
}

export async function updateAccount(id: string, req: CreateAccountReq): Promise<string> {
    try {
        await goApp()?.UpdateAccount?.(id, req);
        await loadAccounts();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo actualizar la cuenta.';
    }
}

export async function deleteAccount(id: string): Promise<string> {
    try {
        await goApp()?.DeleteAccount?.(id);
        await loadAccounts();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo eliminar la cuenta.';
    }
}

export async function setSelected(id: string): Promise<string> {
    try {
        await goApp()?.SetSelectedAccount?.(id);
        selectedAccountId.value = id;
        await loadAccounts();
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo seleccionar la cuenta.';
    }
}

export function loginAuthlib(req: AuthlibLoginReq): string | Promise<AuthLoginResult> {
    if (!req.authServerUrl.trim()) return 'Introduce la URL del servidor de autenticación.';
    if (!req.username.trim()) return 'Introduce tu email o nombre de usuario.';
    if (!req.password) return 'Introduce tu contraseña.';

    if (!loginResultPromise) {
        loginResultPromise = new Promise<AuthLoginResult>((resolve) => {
            const off = EventsOn('account_login', (data: any) => {
                off();
                let payload: any = data;
                if (typeof data === 'string') {
                    try {
                        payload = JSON.parse(data);
                    } catch {
                        payload = {};
                    }
                }
                console.log('[Accounts] login: evento account_login ->', payload);
                resolve({
                    ok: !!payload?.ok,
                    error: payload?.error ?? 'No se pudo iniciar sesión en el servidor.',
                });
            });
            window.setTimeout(() => {
                off();
                console.warn('[Accounts] login: el servidor no respondio en 35s (timeout)');
                resolve({ ok: false, error: 'El servidor de autenticación no respondió a tiempo.' });
            }, 35000);
        }).finally(() => {
            loginResultPromise = null;
        });
    }
    try {
        goApp()?.LoginAuthlib?.(req);
        console.log('[Accounts] login iniciado:', req.username, '->', req.authServerUrl);
    } catch (e) {
        console.error('[Accounts] login: fallo al invocar LoginAuthlib', e);
    }
    return loginResultPromise;
}

export function pendingLogin(): Promise<AuthLoginResult> | null {
    return loginResultPromise;
}

export async function cancelLogin(): Promise<string> {
    try {
        await goApp()?.CancelAuthlibLogin?.();
        console.log('[Accounts] cancelLogin solicitado');
        return '';
    } catch (e: any) {
        console.error('[Accounts] cancelLogin fallo', e);
        return e?.message ?? 'No se pudo cancelar el inicio de sesión.';
    }
}

export async function refreshAccount(id: string): Promise<string> {
    try {
        await goApp()?.RefreshAccount?.(id);
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo renovar la sesión.';
    }
}

export async function refreshAllAccounts(): Promise<string> {
    const res = await goApp()?.RefreshAllAccounts?.();
    return typeof res === 'number' ? String(res) : '';
}

export async function setAutoRefresh(v: boolean): Promise<string> {
    try {
        await goApp()?.SetAccountsAutoRefresh?.(v);
        autoRefresh.value = v;
        return '';
    } catch (e: any) {
        return e?.message ?? 'No se pudo cambiar la opción.';
    }
}

export const selectedAccount = computed<AccountInfo | null>(() => {
    return accounts.value.find((a) => a.id === selectedAccountId.value) ?? null;
});

export const selectedLabel = computed(() => {
    const acc = selectedAccount.value;
    if (!acc) return { name: 'Usuario', sub: 'Sin Cuenta' };
    return { name: acc.username, sub: acc.type === 'authlib' && acc.serverName ? acc.serverName : typeLabel(acc.type) };
});

export function typeLabel(t: AccountType): string {
    switch (t) {
        case 'offline': return 'Local';
        case 'authlib': return 'En línea';
        default: return 'Cuenta';
    }
}

export function serverLabel(a: AccountInfo | null | undefined): string {
    if (!a || a.type !== 'authlib') return '';
    return a.serverName ?? a.authServerUrl ?? '';
}

export function typeDescription(t: AccountType): string {
    switch (t) {
        case 'offline': return 'Juega solo con un nombre, sin iniciar sesión en ningún servidor.';
        case 'authlib': return 'Usa el usuario y contraseña de tu cuenta del servidor para jugar con tu skin y tu nombre.';
        default: return '';
    }
}