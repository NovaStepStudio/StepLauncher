// Página de inicio de sesión: email O usuario + contraseña.
// Si ya hay sesión, el guard del router manda directo al panel.
<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { IconLogin, IconEye, IconEyeOff, IconAlertCircle } from '@tabler/icons-vue';
import AccountHero from '@/Auth/Components/AccountHero.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { validarIdentificador } from '@/Auth/validation';
import { ApiError } from '@/Auth/Api';

const ruta = useRoute();
const router = useRouter();
const { login, ocupado } = useAuth();

const identifier = ref('');
const password = ref('');
const verClave = ref(false);
const errorIdentificador = ref('');
const errorClave = ref('');
const errorGeneral = ref('');

function destino(): string {
    const next = ruta.query.next;
    return typeof next === 'string' && next.startsWith('/') ? next : '/dashboard';
}

async function enviar() {
    errorIdentificador.value = validarIdentificador(identifier.value);
    errorClave.value = password.value ? '' : 'Ingresá tu contraseña.';
    errorGeneral.value = '';
    if (errorIdentificador.value || errorClave.value) return;
    try {
        await login(identifier.value, password.value);
        await router.push(destino());
    } catch (err) {
        errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    }
}
</script>

<template>
    <div class="Login">
        <AccountHero
            titulo="Entrar"
            subtitulo="Qué bueno verte de nuevo."
            descripcion="Iniciá sesión con tu correo o tu nombre de usuario para gestionar tu cuenta."
            insignia="Sesión segura"
        />
        <div class="FormWrap" v-reveal>
            <form class="Card sl-stagger" @submit.prevent="enviar" novalidate>
                <div class="Field">
                    <label for="login-id">Correo o usuario</label>
                    <input
                        id="login-id"
                        v-model="identifier"
                        type="text"
                        autocomplete="username"
                        placeholder="Steve_01 o correo@ejemplo.com"
                        :aria-invalid="!!errorIdentificador"
                    >
                    <small v-if="errorIdentificador" class="FieldError">{{ errorIdentificador }}</small>
                </div>
                <div class="Field">
                    <label for="login-pass">Contraseña</label>
                    <div class="PassRow">
                        <input
                            id="login-pass"
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
                <p v-if="errorGeneral" class="FormError" role="alert">
                    <IconAlertCircle stroke="2" />
                    {{ errorGeneral }}
                </p>
                <button class="BtnPrimary" type="submit" :disabled="ocupado">
                    <IconLogin stroke="2" />
                    {{ ocupado ? 'Entrando...' : 'Entrar' }}
                </button>
                <p class="Switch">
                    ¿Todavía no tenés cuenta?
                    <RouterLink to="/register">Creá una gratis</RouterLink>
                </p>
            </form>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Login{
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
        gap: 1rem;
        width: 100%;
        max-width: 26rem;
        padding: 1.8rem;
        border-radius: .9rem;
        border: 1px solid #ffffff18;
        background: #ffffff08;
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
        }
    }
}
</style>
