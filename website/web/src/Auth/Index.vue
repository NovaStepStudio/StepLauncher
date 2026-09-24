// Página única de acceso: entrar, crear cuenta y recuperar contraseña.
// El registro queda pendiente de confirmación (plantilla Supabase Confirm sign up);
// el login sin confirmar ofrece reenviar; el recupero usa Reset password.
// `?tab=` sincroniza la pestaña (login | register | recovery).
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IconLogin, IconUserPlus, IconKey, IconEye, IconEyeOff, IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import AccountHero from '@/Auth/Components/AccountHero.vue';
import SliderTabs from '@/Auth/Components/SliderTabs.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { validarIdentificador, validarEmail, validarUsername, validarPassword } from '@/Auth/validation';
import { ApiError, reenviarConfirmacion, pedirRecupero } from '@/Auth/Api';

type Tab = 'login' | 'register' | 'recovery';

const ruta = useRoute();
const router = useRouter();
const { login, register, ocupado } = useAuth();

const pestanas = [
    { key: 'login', label: 'Entrar', icon: IconLogin },
    { key: 'register', label: 'Crear cuenta', icon: IconUserPlus },
    { key: 'recovery', label: 'Recuperar', icon: IconKey },
] as const;

function tabDe(query: unknown): Tab {
    if (query === 'register' || query === 'recovery') return query;
    return 'login';
}

// --- Pestaña activa (sincronizada con ?tab=) ---------------------------------
const tab = ref<Tab>(tabDe(ruta.query.tab));

watch(
    () => ruta.query.tab,
    (nueva) => {
        tab.value = tabDe(nueva);
    },
);

watch(tab, (nueva) => {
    router.replace({ path: '/auth', query: { ...ruta.query, tab: nueva } });
});

function destino(): string {
    const next = ruta.query.next;
    return typeof next === 'string' && next.startsWith('/') ? next : '/dashboard';
}

// --- Formulario de entrar ------------------------------------------------------
const identifier = ref('');
const password = ref('');
const verClave = ref(false);
const errorIdentificador = ref('');
const errorClave = ref('');
const errorLogin = ref('');
const sinConfirmar = ref(false);
const emailReenvio = ref('');
const errorReenvio = ref('');
const avisoReenvio = ref('');
const enviandoReenvio = ref(false);

async function enviarLogin() {
    errorIdentificador.value = validarIdentificador(identifier.value);
    errorClave.value = password.value ? '' : 'Ingresá tu contraseña.';
    errorLogin.value = '';
    sinConfirmar.value = false;
    avisoReenvio.value = '';
    if (errorIdentificador.value || errorClave.value) return;
    try {
        await login(identifier.value, password.value);
        await router.push(destino());
    } catch (err) {
        if (err instanceof ApiError && err.code === 'email_not_confirmed') {
            sinConfirmar.value = true;
            const posible = identifier.value.trim().toLowerCase();
            emailReenvio.value = posible.includes('@') ? posible : '';
            errorLogin.value = err.message;
            return;
        }
        errorLogin.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    }
}

async function reenviarDesdeLogin() {
    errorReenvio.value = validarEmail(emailReenvio.value);
    avisoReenvio.value = '';
    if (errorReenvio.value) return;
    enviandoReenvio.value = true;
    try {
        await reenviarConfirmacion({ email: emailReenvio.value.trim().toLowerCase(), type: 'signup' });
        avisoReenvio.value = 'Listo: si ese correo está pendiente, te llega un nuevo enlace.';
    } catch (err) {
        errorReenvio.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviandoReenvio.value = false;
    }
}

// --- Formulario de registro -----------------------------------------------------
const email = ref('');
const username = ref('');
const nuevaClave = ref('');
const confirm = ref('');
const errorEmail = ref('');
const errorUsername = ref('');
const errorPassword = ref('');
const errorConfirm = ref('');
const errorRegister = ref('');
const pendienteEmail = ref('');
const avisoReenvioRegistro = ref('');
const errorReenvioRegistro = ref('');
const enviandoReenvioRegistro = ref(false);

async function enviarRegistro() {
    errorEmail.value = validarEmail(email.value);
    errorUsername.value = validarUsername(username.value);
    errorPassword.value = validarPassword(nuevaClave.value);
    errorConfirm.value = confirm.value === nuevaClave.value ? '' : 'Las contraseñas no coinciden.';
    errorRegister.value = '';
    if (errorEmail.value || errorUsername.value || errorPassword.value || errorConfirm.value) return;
    try {
        const res = await register(email.value, username.value, nuevaClave.value);
        if (res.pendiente) {
            pendienteEmail.value = email.value.trim().toLowerCase();
            return;
        }
        await router.push(destino());
    } catch (err) {
        errorRegister.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    }
}

async function reenviarRegistro() {
    errorReenvioRegistro.value = '';
    avisoReenvioRegistro.value = '';
    if (!pendienteEmail.value) return;
    enviandoReenvioRegistro.value = true;
    try {
        await reenviarConfirmacion({ email: pendienteEmail.value, type: 'signup' });
        avisoReenvioRegistro.value = 'Listo: revisá tu bandeja (y el spam).';
    } catch (err) {
        errorReenvioRegistro.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviandoReenvioRegistro.value = false;
    }
}

// --- Formulario de recupero -------------------------------------------------------
const emailRecupero = ref('');
const errorRecupero = ref('');
const avisoRecupero = ref('');
const errorRecuperoApi = ref('');
const enviandoRecupero = ref(false);

async function enviarRecupero() {
    errorRecupero.value = validarEmail(emailRecupero.value);
    errorRecuperoApi.value = '';
    avisoRecupero.value = '';
    if (errorRecupero.value) return;
    enviandoRecupero.value = true;
    try {
        await pedirRecupero(emailRecupero.value.trim().toLowerCase());
        avisoRecupero.value = 'Listo: si ese correo está registrado, te llega un enlace para restablecer.';
        emailRecupero.value = '';
    } catch (err) {
        errorRecuperoApi.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviandoRecupero.value = false;
    }
}
</script>

<template>
    <div class="Auth">
        <AccountHero
            titulo="Mi cuenta"
            subtitulo="Entrá o creá tu cuenta."
            descripcion="Un solo lugar para iniciar sesión, registrarte gratis o recuperar tu contraseña con un enlace a tu correo."
            insignia="Gratis para siempre"
        />
        <div class="FormWrap" v-reveal>
            <div class="Card sl-stagger">
                <SliderTabs v-model="tab" :tabs="pestanas" />
                <form v-if="tab === 'login'" class="Form" @submit.prevent="enviarLogin" novalidate>
                    <div class="Field">
                        <label for="auth-id">Correo o usuario</label>
                        <input
                            id="auth-id"
                            v-model="identifier"
                            type="text"
                            autocomplete="username"
                            placeholder="Steve_01 o correo@ejemplo.com"
                            :aria-invalid="!!errorIdentificador"
                        >
                        <small v-if="errorIdentificador" class="FieldError">{{ errorIdentificador }}</small>
                    </div>
                    <div class="Field">
                        <label for="auth-pass">Contraseña</label>
                        <div class="PassRow">
                            <input
                                id="auth-pass"
                                v-model="password"
                                :type="verClave ? 'text' : 'password'"
                                autocomplete="current-password"
                                placeholder="Tu contraseña"
                                :aria-invalid="!!errorClave"
                            >
                            <button type="button" class="IconBtn" :aria-label="verClave ? 'Ocultar contraseña' : 'Mostrar contraseña'" @click="verClave = !verClave">
                                <component :is="verClave ? IconEyeOff : IconEye" stroke="2" />
                            </button>
                        </div>
                        <small v-if="errorClave" class="FieldError">{{ errorClave }}</small>
                    </div>
                    <p v-if="errorLogin" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorLogin }}
                    </p>
                    <div v-if="sinConfirmar" class="Resend">
                        <div class="Field">
                            <label for="auth-resend">Correo para reenviar la confirmación</label>
                            <input
                                id="auth-resend"
                                v-model="emailReenvio"
                                type="email"
                                autocomplete="email"
                                placeholder="correo@ejemplo.com"
                                :aria-invalid="!!errorReenvio"
                            >
                            <small v-if="errorReenvio" class="FieldError">{{ errorReenvio }}</small>
                        </div>
                        <p v-if="avisoReenvio" class="Ok" role="status">
                            <IconCircleCheck stroke="2" />
                            {{ avisoReenvio }}
                        </p>
                        <button type="button" class="GhostBtn" :disabled="enviandoReenvio" @click="reenviarDesdeLogin">
                            {{ enviandoReenvio ? 'Enviando...' : 'Reenviar confirmación' }}
                        </button>
                    </div>
                    <button class="BtnPrimary" type="submit" :disabled="ocupado">
                        <IconLogin stroke="2" />
                        {{ ocupado ? 'Entrando...' : 'Entrar' }}
                    </button>
                    <p class="Switch">
                        ¿Olvidaste tu contraseña?
                        <button type="button" class="LinkBtn" @click="tab = 'recovery'">Recuperala acá</button>
                    </p>
                </form>
                <form v-else-if="tab === 'register'" class="Form" @submit.prevent="enviarRegistro" novalidate>
                    <template v-if="!pendienteEmail">
                        <div class="Field">
                            <label for="auth-email">Correo</label>
                            <input
                                id="auth-email"
                                v-model="email"
                                type="email"
                                autocomplete="email"
                                placeholder="correo@ejemplo.com"
                                :aria-invalid="!!errorEmail"
                            >
                            <small v-if="errorEmail" class="FieldError">{{ errorEmail }}</small>
                        </div>
                        <div class="Field">
                            <label for="auth-user">Nombre de usuario</label>
                            <input
                                id="auth-user"
                                v-model="username"
                                type="text"
                                autocomplete="username"
                                placeholder="Steve_01"
                                :aria-invalid="!!errorUsername"
                            >
                            <small v-if="errorUsername" class="FieldError">{{ errorUsername }}</small>
                            <small v-else class="Hint">De 3 a 20 caracteres: letras, números y . _ -</small>
                        </div>
                        <div class="Field">
                            <label for="auth-new">Contraseña</label>
                            <div class="PassRow">
                                <input
                                    id="auth-new"
                                    v-model="nuevaClave"
                                    :type="verClave ? 'text' : 'password'"
                                    autocomplete="new-password"
                                    placeholder="Mínimo 8 caracteres"
                                    :aria-invalid="!!errorPassword"
                                >
                                <button type="button" class="IconBtn" :aria-label="verClave ? 'Ocultar contraseña' : 'Mostrar contraseña'" @click="verClave = !verClave">
                                    <component :is="verClave ? IconEyeOff : IconEye" stroke="2" />
                                </button>
                            </div>
                            <small v-if="errorPassword" class="FieldError">{{ errorPassword }}</small>
                        </div>
                        <div class="Field">
                            <label for="auth-confirm">Repetí la contraseña</label>
                            <input
                                id="auth-confirm"
                                v-model="confirm"
                                :type="verClave ? 'text' : 'password'"
                                autocomplete="new-password"
                                placeholder="Una vez más, por las dudas"
                                :aria-invalid="!!errorConfirm"
                            >
                            <small v-if="errorConfirm" class="FieldError">{{ errorConfirm }}</small>
                        </div>
                        <p v-if="errorRegister" class="FormError" role="alert">
                            <IconAlertCircle stroke="2" />
                            {{ errorRegister }}
                        </p>
                        <button class="BtnPrimary" type="submit" :disabled="ocupado">
                            <IconUserPlus stroke="2" />
                            {{ ocupado ? 'Creando...' : 'Crear mi cuenta' }}
                        </button>
                    </template>
                    <template v-else>
                        <p class="Ok" role="status">
                            <IconCircleCheck stroke="2" />
                            Cuenta creada: te enviamos un correo a {{ pendienteEmail }} para confirmarla.
                        </p>
                        <p v-if="avisoReenvioRegistro" class="Ok" role="status">
                            <IconCircleCheck stroke="2" />
                            {{ avisoReenvioRegistro }}
                        </p>
                        <p v-if="errorReenvioRegistro" class="FormError" role="alert">
                            <IconAlertCircle stroke="2" />
                            {{ errorReenvioRegistro }}
                        </p>
                        <button type="button" class="GhostBtn" :disabled="enviandoReenvioRegistro" @click="reenviarRegistro">
                            {{ enviandoReenvioRegistro ? 'Enviando...' : 'Reenviar correo' }}
                        </button>
                        <p class="Switch">
                            ¿Ya la confirmaste?
                            <button type="button" class="LinkBtn" @click="tab = 'login'">Entrá acá</button>
                        </p>
                    </template>
                </form>
                <form v-else class="Form" @submit.prevent="enviarRecupero" novalidate>
                    <div class="Field">
                        <label for="auth-recover">Correo de tu cuenta</label>
                        <input
                            id="auth-recover"
                            v-model="emailRecupero"
                            type="email"
                            autocomplete="email"
                            placeholder="correo@ejemplo.com"
                            :aria-invalid="!!errorRecupero"
                        >
                        <small v-if="errorRecupero" class="FieldError">{{ errorRecupero }}</small>
                        <small v-else class="Hint">Abrí el enlace y elegí tu contraseña nueva ahí mismo.</small>
                    </div>
                    <p v-if="errorRecuperoApi" class="FormError" role="alert">
                        <IconAlertCircle stroke="2" />
                        {{ errorRecuperoApi }}
                    </p>
                    <p v-if="avisoRecupero" class="Ok" role="status">
                        <IconCircleCheck stroke="2" />
                        {{ avisoRecupero }}
                    </p>
                    <button class="BtnPrimary" type="submit" :disabled="enviandoRecupero">
                        <IconKey stroke="2" />
                        {{ enviandoRecupero ? 'Enviando...' : 'Enviar enlace' }}
                    </button>
                    <p class="Switch">
                        ¿Te acordaste?
                        <button type="button" class="LinkBtn" @click="tab = 'login'">Volver a entrar</button>
                    </p>
                </form>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Auth{
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
                .Hint{
                    font-size: .7rem;
                    opacity: .45;
                }
                .FieldError{
                    font-size: .72rem;
                    color: #ff9d9d;
                }
                .PassRow{
                    display: flex;
                    gap: .5rem;
                    .IconBtn{
                        display: flex;
                        justify-content: center;
                        align-items: center;
                        flex-shrink: 0;
                        width: 2.6rem;
                        border-radius: .55rem;
                        border: 1px solid #ffffff25;
                        background: #ffffff08;
                        color: #ffffffa6;
                        cursor: pointer;
                        transition: background 150ms, color 150ms;
                        svg{
                            width: 1.1rem;
                            height: 1.1rem;
                        }
                        &:hover{
                            background: #ffffff14;
                            color: #fff;
                        }
                    }
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
            .Resend{
                display: flex;
                flex-direction: column;
                gap: .8rem;
                padding: 1rem;
                border-radius: .55rem;
                border: 1px solid #ffffff18;
                background: #ffffff05;
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
            .Switch{
                margin: 0;
                text-align: center;
                font-size: .8rem;
                opacity: .75;
                .LinkBtn{
                    padding: 0;
                    border: 0;
                    background: transparent;
                    color: #fff;
                    font-size: inherit;
                    font-weight: 600;
                    font-family: inherit;
                    cursor: pointer;
                    text-decoration: underline;
                    text-underline-offset: 3px;
                    &:hover{
                        opacity: .75;
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
                .GhostBtn{
                    width: 100%;
                    min-height: 2.75rem;
                }
                .BtnPrimary{
                    width: 100%;
                    min-height: 2.75rem;
                }
            }
        }
    }
}
</style>
