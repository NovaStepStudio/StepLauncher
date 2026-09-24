// Vuelta de los correos Supabase (SITE_URL + /auth/callback).
// Lee SOLA la URL y actúa: el enlace de recupero trae la sesión en el hash y
// muestra directo la contraseña nueva; el de confirmación entra solo al panel.
// El usuario no pega nada salvo que el enlace llegue incompleto (solo el email).
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IconAlertCircle, IconCircleCheck, IconKey, IconLogin } from '@tabler/icons-vue';
import AccountHero from '@/Auth/Components/AccountHero.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { validarCodigo, validarEmail, validarPassword } from '@/Auth/validation';
import {
    ApiError,
    confirmarCodigo,
    fijarNuevaPassword,
    fijarPasswordRecupero,
    miPerfil,
    reenviarConfirmacion,
    type ApiSession,
    type ConfirmType,
} from '@/Auth/Api';

type Fase = 'trabajando' | 'clave-recupero' | 'clave-otp' | 'pide-email' | 'manual' | 'lista';

const ruta = useRoute();
const router = useRouter();
const { confirmarSesion } = useAuth();

const fase = ref<Fase>('trabajando');
const estado = ref('Leyendo tu enlace...');
const errorGeneral = ref('');
const aviso = ref('');

const tipo = ref<ConfirmType>('signup');
const email = ref('');
const codigoUrl = ref('');
const sesionRecupero = ref('');
// Respaldo OTP (recupero por código): email + token ya verificados.
const otpEmail = ref('');
const otpToken = ref('');

const errorEmail = ref('');
const enviando = ref(false);

const nueva = ref('');
const repite = ref('');
const errorClave = ref('');
const errorClaveApi = ref('');
const avisoClave = ref('');
const guardandoClave = ref(false);

const titulo = computed(() => {
    if (tipo.value === 'recovery') return 'Restablecer contraseña';
    if (tipo.value === 'email_change') return 'Confirmar correo nuevo';
    return 'Confirmar tu cuenta';
});

function queryStr(nombre: string): string {
    const valor = ruta.query[nombre];
    if (Array.isArray(valor)) return typeof valor[0] === 'string' ? valor[0].trim() : '';
    return typeof valor === 'string' ? valor.trim() : '';
}

function tipoDe(valor: string): ConfirmType | null {
    if (valor === 'signup' || valor === 'email_change' || valor === 'recovery') return valor;
    return null;
}

// Parsea el fragmento del enlace (viene como clave=valor unidos con &).
function parsearHash(hash: string): Record<string, string> {
    const limpio = hash.replace(/^#\/?/, '');
    const partes: Record<string, string> = {};
    for (const par of limpio.split('&')) {
        const i = par.indexOf('=');
        if (i <= 0) continue;
        try {
            partes[par.slice(0, i)] = decodeURIComponent(par.slice(i + 1));
        } catch {
            partes[par.slice(0, i)] = par.slice(i + 1);
        }
    }
    return partes;
}

// Saca los tokens de la barra para que no queden en el historial.
function limpiarUrl(): void {
    void router.replace({ path: ruta.path, query: ruta.query });
}

function segundosAExpira(hash: Record<string, string>): number | null {
    const directo = Number(hash['expires_at']);
    if (Number.isFinite(directo) && directo > 0) return directo;
    const segundos = Number(hash['expires_in']);
    if (Number.isFinite(segundos) && segundos > 0) return Math.floor(Date.now() / 1000) + segundos;
    return null;
}

async function entrarConSesion(sesion: ApiSession): Promise<void> {
    const perfil = await miPerfil(sesion.accessToken);
    await confirmarSesion({ id: perfil.id, email: perfil.email, username: perfil.username }, sesion);
    await router.push('/dashboard');
}

// Enlace de confirmación/cambio de email con código: confirma y entra solo.
async function confirmarConCodigo(mail: string, token: string, kind: ConfirmType): Promise<void> {
    const res = await confirmarCodigo({ email: mail, token, type: kind });
    if (res.session && res.user) {
        await entrarConSesion(res.session);
        return;
    }
    fase.value = 'lista';
    aviso.value = 'Correo confirmado. Ya puedes entrar.';
}

async function arrancar(): Promise<void> {
    const hash = parsearHash(ruta.hash || '');
    const acceso = (hash['access_token'] || '').trim();
    const tipoHash = tipoDe(hash['type'] || '');
    const mailQuery = queryStr('email').toLowerCase();
    const tipoQuery = tipoDe(queryStr('type'));
    const tokenQuery = queryStr('token_hash') || queryStr('token') || queryStr('code');
    const kind: ConfirmType = tipoHash ?? tipoQuery ?? 'signup';
    tipo.value = kind;

    // 1. El enlace trae la sesión en el hash: nada que pegar.
    if (acceso) {
        limpiarUrl();
        if (kind === 'recovery') {
            sesionRecupero.value = acceso;
            fase.value = 'clave-recupero';
            return;
        }
        estado.value = 'Confirmando tu cuenta...';
        try {
            await entrarConSesion({
                accessToken: acceso,
                refreshToken: (hash['refresh_token'] || '').trim(),
                expiresAt: segundosAExpira(hash),
                tokenType: 'bearer',
            });
        } catch (err) {
            fase.value = 'lista';
            errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
        }
        return;
    }

    // 2. El enlace trae código + email: confirma solo.
    if (tokenQuery && mailQuery && !validarEmail(mailQuery)) {
        limpiarUrl();
        estado.value = 'Confirmando tu cuenta...';
        try {
            if (kind === 'recovery') {
                const res = await confirmarCodigo({ email: mailQuery, token: tokenQuery, type: 'recovery' });
                if (res.recoveryVerified) {
                    otpEmail.value = mailQuery;
                    otpToken.value = tokenQuery;
                    fase.value = 'clave-otp';
                    return;
                }
            } else {
                await confirmarConCodigo(mailQuery, tokenQuery, kind);
                return;
            }
        } catch (err) {
            fase.value = 'lista';
            errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
            return;
        }
    }

    // 3. El enlace trae código pero sin email: se pide SOLO el email.
    if (tokenQuery) {
        limpiarUrl();
        email.value = '';
        codigoUrl.value = tokenQuery;
        fase.value = 'pide-email';
        return;
    }

    // 4. Llegada directa (sin enlace): form manual mínimo.
    if (mailQuery && !validarEmail(mailQuery)) email.value = mailQuery;
    fase.value = 'manual';
}

async function confirmarSoloEmail() {
    errorEmail.value = validarEmail(email.value);
    errorGeneral.value = '';
    if (errorEmail.value) return;
    enviando.value = true;
    estado.value = 'Confirmando tu cuenta...';
    try {
        const mail = email.value.trim().toLowerCase();
        if (tipo.value === 'recovery') {
            const res = await confirmarCodigo({ email: mail, token: codigoUrl.value, type: 'recovery' });
            if (res.recoveryVerified) {
                otpEmail.value = mail;
                otpToken.value = codigoUrl.value;
                fase.value = 'clave-otp';
                return;
            }
        } else {
            await confirmarConCodigo(mail, codigoUrl.value, tipo.value);
        }
    } catch (err) {
        errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviando.value = false;
    }
}

async function reenviar() {
    errorGeneral.value = '';
    aviso.value = '';
    if (tipo.value === 'recovery') {
        await router.push({ path: '/auth', query: { tab: 'recovery' } });
        return;
    }
    const mail = (email.value || otpEmail.value).trim().toLowerCase();
    if (validarEmail(mail)) {
        errorGeneral.value = 'Volvé a pedir el enlace desde Entrar o Crear cuenta.';
        return;
    }
    enviando.value = true;
    try {
        await reenviarConfirmacion({ email: mail, type: tipo.value });
        aviso.value = 'Listo: si ese correo está pendiente, te llega un nuevo enlace.';
    } catch (err) {
        errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviando.value = false;
    }
}

function claveValida(): boolean {
    const mala = validarPassword(nueva.value);
    const coincide = repite.value === nueva.value ? '' : 'Las contraseñas nuevas no coinciden.';
    errorClave.value = mala || coincide;
    return !errorClave.value;
}

async function guardarClaveRecupero() {
    if (!claveValida()) return;
    errorClaveApi.value = '';
    avisoClave.value = '';
    guardandoClave.value = true;
    try {
        await fijarPasswordRecupero(sesionRecupero.value, nueva.value);
        avisoClave.value = 'Contraseña actualizada. Te avisamos por correo.';
        nueva.value = '';
        repite.value = '';
    } catch (err) {
        errorClaveApi.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        guardandoClave.value = false;
    }
}

async function guardarClaveOtp() {
    if (!claveValida()) return;
    errorClaveApi.value = '';
    avisoClave.value = '';
    guardandoClave.value = true;
    try {
        await fijarNuevaPassword({ email: otpEmail.value, token: otpToken.value, newPassword: nueva.value });
        avisoClave.value = 'Contraseña actualizada. Te avisamos por correo.';
        nueva.value = '';
        repite.value = '';
    } catch (err) {
        errorClaveApi.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        guardandoClave.value = false;
    }
}

onMounted(() => {
    void arrancar();
});
</script>

<template>
    <div class="Callback">
        <AccountHero
            titulo="Confirmar correo"
            :subtitulo="titulo"
            descripcion="Abriste el enlace de tu correo: terminamos solos, sin que pegues nada."
            insignia="Verificación"
        />
        <div class="FormWrap" v-reveal>
            <div class="Card sl-stagger">
                <template v-if="fase === 'trabajando'">
                    <p class="State" role="status">{{ estado }}</p>
                </template>
                <form v-else-if="fase === 'clave-recupero'" class="Form" @submit.prevent="guardarClaveRecupero" novalidate>
                    <p class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        Enlace válido: elegí tu contraseña nueva.
                    </p>
                    <div class="Field">
                        <label for="cb-new">Contraseña nueva</label>
                        <input
                            id="cb-new"
                            v-model="nueva"
                            type="password"
                            autocomplete="new-password"
                            placeholder="Mínimo 8 caracteres"
                        >
                    </div>
                    <div class="Field">
                        <label for="cb-rep">Repetí la nueva</label>
                        <input
                            id="cb-rep"
                            v-model="repite"
                            type="password"
                            autocomplete="new-password"
                            placeholder="Una vez más, por las dudas"
                        >
                        <small v-if="errorClave" class="FieldError">{{ errorClave }}</small>
                    </div>
                    <p v-if="errorClaveApi" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorClaveApi }}
                    </p>
                    <p v-if="avisoClave" class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        {{ avisoClave }}
                    </p>
                    <button v-if="!avisoClave" class="BtnPrimary" type="submit" :disabled="guardandoClave">
                        <IconKey stroke="2" />
                        {{ guardandoClave ? 'Guardando...' : 'Guardar contraseña' }}
                    </button>
                    <p v-if="errorClaveApi && !avisoClave" class="Switch">
                        <RouterLink to="/auth?tab=recovery">Pedir un nuevo enlace</RouterLink>
                    </p>
                    <p v-else class="Switch">
                        <RouterLink to="/auth?tab=login">
                            <span class="LinkInline">
                                <IconLogin stroke="2" />
                                Ir a entrar
                            </span>
                        </RouterLink>
                    </p>
                </form>
                <form v-else-if="fase === 'clave-otp'" class="Form" @submit.prevent="guardarClaveOtp" novalidate>
                    <p class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        Código válido: elegí tu contraseña nueva.
                    </p>
                    <div class="Field">
                        <label for="cb-otp-new">Contraseña nueva</label>
                        <input
                            id="cb-otp-new"
                            v-model="nueva"
                            type="password"
                            autocomplete="new-password"
                            placeholder="Mínimo 8 caracteres"
                        >
                    </div>
                    <div class="Field">
                        <label for="cb-otp-rep">Repetí la nueva</label>
                        <input
                            id="cb-otp-rep"
                            v-model="repite"
                            type="password"
                            autocomplete="new-password"
                            placeholder="Una vez más, por las dudas"
                        >
                        <small v-if="errorClave" class="FieldError">{{ errorClave }}</small>
                    </div>
                    <p v-if="errorClaveApi" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorClaveApi }}
                    </p>
                    <p v-if="avisoClave" class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        {{ avisoClave }}
                    </p>
                    <button v-if="!avisoClave" class="BtnPrimary" type="submit" :disabled="guardandoClave">
                        <IconKey stroke="2" />
                        {{ guardandoClave ? 'Guardando...' : 'Guardar contraseña' }}
                    </button>
                    <p v-else class="Switch">
                        <RouterLink to="/auth?tab=login">
                            <span class="LinkInline">
                                <IconLogin stroke="2" />
                                Ir a entrar
                            </span>
                        </RouterLink>
                    </p>
                </form>
                <form v-else-if="fase === 'pide-email'" class="Form" @submit.prevent="confirmarSoloEmail" novalidate>
                    <p class="State">Casi listo: confirmá cuál es tu correo para terminar.</p>
                    <div class="Field">
                        <label for="cb-email">Tu correo</label>
                        <input
                            id="cb-email"
                            v-model="email"
                            type="email"
                            autocomplete="email"
                            placeholder="correo@ejemplo.com"
                            :aria-invalid="!!errorEmail"
                        >
                        <small v-if="errorEmail" class="FieldError">{{ errorEmail }}</small>
                    </div>
                    <p v-if="errorGeneral" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorGeneral }}
                    </p>
                    <button class="BtnPrimary" type="submit" :disabled="enviando">
                        {{ enviando ? 'Confirmando...' : 'Confirmar' }}
                    </button>
                </form>
                <div v-else class="Form">
                    <p v-if="aviso" class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        {{ aviso }}
                    </p>
                    <p v-if="errorGeneral" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorGeneral }}
                    </p>
                    <button v-if="errorGeneral" type="button" class="GhostBtn" :disabled="enviando" @click="reenviar">
                        Pedir un nuevo enlace
                    </button>
                    <p class="Switch">
                        <RouterLink to="/auth?tab=login">Ir a entrar</RouterLink>
                    </p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Callback{
    display: flex;
    flex-direction: column;
    background: #000;
}
.FormWrap{
    width: 100%;
    display: flex;
    justify-content: center;
    padding: 2.5rem 1.5rem 4rem 1.5rem;
    .Card{
        display: flex;
        flex-direction: column;
        gap: 1.2rem;
        width: 100%;
        max-width: 26rem;
        padding: 1.8rem;
        border-radius: .9rem;
        border: 1px solid #ffffff18;
        background: #ffffff08;
        .State{
            margin: 0;
            text-align: center;
            font-size: .85rem;
            opacity: .55;
        }
        .Form{
            display: flex;
            flex-direction: column;
            gap: 1rem;
            .Field{
                display: flex;
                flex-direction: column;
                gap: .45rem;
                label{
                    font-size: .75rem;
                    font-weight: 600;
                    font-family: 'Lexend';
                    opacity: .75;
                }
                input{
                    width: 100%;
                    padding: .65rem .85rem;
                    border-radius: .55rem;
                    border: 1px solid #ffffff25;
                    background: #00000080;
                    color: #fff;
                    font-size: .85rem;
                    font-family: inherit;
                    outline: none;
                    transition: border-color 150ms;
                    &::placeholder{
                        color: #ffffff45;
                    }
                    &:focus{
                        border-color: #ffffff60;
                    }
                }
                .FieldError{
                    font-size: .72rem;
                    color: #ff9d9d;
                }
            }
            .FormError{
                display: flex;
                align-items: center;
                gap: .5rem;
                margin: 0;
                padding: .65rem .85rem;
                border-radius: .55rem;
                border: 1px solid #ff9d9d45;
                background: #ff6b6b14;
                font-size: .8rem;
                line-height: 1.5;
                svg{
                    flex-shrink: 0;
                    width: 1.1rem;
                    height: 1.1rem;
                    color: #ff9d9d;
                }
            }
            .Ok{
                display: flex;
                align-items: center;
                gap: .5rem;
                margin: 0;
                padding: .65rem .85rem;
                border-radius: .55rem;
                border: 1px solid #9dffb045;
                background: #9dffb014;
                font-size: .8rem;
                line-height: 1.5;
                color: #9dffb0;
                svg{
                    flex-shrink: 0;
                    width: 1.1rem;
                    height: 1.1rem;
                }
            }
            .BtnPrimary{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .45rem;
                padding: .7rem 1.3rem;
                border-radius: .55rem;
                border: 1px solid #fff;
                background: #fff;
                color: #000;
                font-size: .85rem;
                font-weight: 700;
                font-family: inherit;
                cursor: pointer;
                transition: filter 150ms, transform 150ms, opacity 150ms;
                svg{
                    width: 1.1rem;
                    height: 1.1rem;
                }
                &:hover:not(:disabled){
                    filter: brightness(.85);
                    transform: translateY(-1px);
                }
                &:disabled{
                    opacity: .6;
                    cursor: wait;
                }
            }
            .GhostBtn{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .4rem;
                padding: .6rem 1.1rem;
                border-radius: .55rem;
                border: 1px solid #ffffff25;
                background: #ffffff08;
                color: #fff;
                font-size: .8rem;
                font-weight: 600;
                font-family: inherit;
                cursor: pointer;
                transition: background 150ms, opacity 150ms;
                &:hover:not(:disabled){
                    background: #ffffff14;
                }
                &:disabled{
                    opacity: .5;
                    cursor: wait;
                }
            }
            .Switch{
                margin: 0;
                text-align: center;
                font-size: .8rem;
                opacity: .75;
                a{
                    color: #fff;
                    font-weight: 600;
                    text-underline-offset: 3px;
                    &:hover{
                        opacity: .75;
                    }
                }
                .LinkInline{
                    display: inline-flex;
                    align-items: center;
                    gap: .3rem;
                    svg{
                        width: 1rem;
                        height: 1rem;
                    }
                }
            }
        }
    }
}
@media (max-width: 600px){
    .FormWrap{
        padding: 2rem 1rem 3rem 1rem;
        .Card{
            min-width: 0;
            padding: 1.2rem;
            .Form{
                .BtnPrimary,
                .GhostBtn{
                    width: 100%;
                    min-height: 2.75rem;
                }
            }
        }
    }
}
</style>
