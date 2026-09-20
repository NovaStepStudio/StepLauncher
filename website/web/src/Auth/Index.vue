// Página única de acceso: entrar y crear cuenta en el mismo lugar,
// con pestañas. `?tab=register` abre directo la pestaña de registro.
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IconLogin, IconUserPlus, IconEye, IconEyeOff, IconAlertCircle } from '@tabler/icons-vue';
import AccountHero from '@/Auth/Components/AccountHero.vue';
import SliderTabs from '@/Auth/Components/SliderTabs.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { validarIdentificador, validarEmail, validarUsername, validarPassword } from '@/Auth/validation';
import { ApiError } from '@/Auth/Api';

type Tab = 'login' | 'register';

const ruta = useRoute();
const router = useRouter();
const { login, register, ocupado } = useAuth();

const pestanas = [
    { key: 'login', label: 'Entrar', icon: IconLogin },
    { key: 'register', label: 'Crear cuenta', icon: IconUserPlus },
] as const;

// --- Pestaña activa (sincronizada con ?tab=) ---------------------------------
const tab = ref<Tab>(ruta.query.tab === 'register' ? 'register' : 'login');

watch(
    () => ruta.query.tab,
    (nueva) => {
        tab.value = nueva === 'register' ? 'register' : 'login';
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

async function enviarLogin() {
    errorIdentificador.value = validarIdentificador(identifier.value);
    errorClave.value = password.value ? '' : 'Ingresá tu contraseña.';
    errorLogin.value = '';
    if (errorIdentificador.value || errorClave.value) return;
    try {
        await login(identifier.value, password.value);
        await router.push(destino());
    } catch (err) {
        errorLogin.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
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

async function enviarRegistro() {
    errorEmail.value = validarEmail(email.value);
    errorUsername.value = validarUsername(username.value);
    errorPassword.value = validarPassword(nuevaClave.value);
    errorConfirm.value = confirm.value === nuevaClave.value ? '' : 'Las contraseñas no coinciden.';
    errorRegister.value = '';
    if (errorEmail.value || errorUsername.value || errorPassword.value || errorConfirm.value) return;
    try {
        await register(email.value, username.value, nuevaClave.value);
        await router.push(destino());
    } catch (err) {
        errorRegister.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    }
}
</script>

<template>
    <div class="Auth">
        <AccountHero
            titulo="Mi cuenta"
            subtitulo="Entrá o creá tu cuenta."
            descripcion="Un solo lugar para iniciar sesión o registrarte gratis con tu email, tu usuario y tu contraseña."
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
                    <button class="BtnPrimary" type="submit" :disabled="ocupado">
                        <IconLogin stroke="2" />
                        {{ ocupado ? 'Entrando...' : 'Entrar' }}
                    </button>
                </form>
                <form v-else class="Form" @submit.prevent="enviarRegistro" novalidate>
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
        }
    }
}
</style>
