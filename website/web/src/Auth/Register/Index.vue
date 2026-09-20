// Página de registro: email + usuario + contraseña (lo mismo que pide la API).
// Al crear la cuenta se inicia sesión automáticamente y se va al panel.
<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { IconUserPlus, IconEye, IconEyeOff, IconAlertCircle } from '@tabler/icons-vue';
import AccountHero from '@/Auth/Components/AccountHero.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { validarEmail, validarUsername, validarPassword } from '@/Auth/validation';
import { ApiError } from '@/Auth/Api';

const router = useRouter();
const { register, ocupado } = useAuth();

const email = ref('');
const username = ref('');
const password = ref('');
const confirm = ref('');
const verClave = ref(false);
const errorEmail = ref('');
const errorUsername = ref('');
const errorPassword = ref('');
const errorConfirm = ref('');
const errorGeneral = ref('');

async function enviar() {
    errorEmail.value = validarEmail(email.value);
    errorUsername.value = validarUsername(username.value);
    errorPassword.value = validarPassword(password.value);
    errorConfirm.value = confirm.value === password.value ? '' : 'Las contraseñas no coinciden.';
    errorGeneral.value = '';
    if (errorEmail.value || errorUsername.value || errorPassword.value || errorConfirm.value) return;
    try {
        await register(email.value, username.value, password.value);
        await router.push('/dashboard');
    } catch (err) {
        errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    }
}
</script>

<template>
    <div class="Register">
        <AccountHero
            titulo="Crear cuenta"
            subtitulo="Tu cuenta, tu Minecraft."
            descripcion="Un email, un usuario y una contraseña: eso es todo lo que necesitás para empezar."
            insignia="Gratis para siempre"
        />
        <div class="FormWrap" v-reveal>
            <form class="Card sl-stagger" @submit.prevent="enviar" novalidate>
                <div class="Field">
                    <label for="reg-email">Correo</label>
                    <input
                        id="reg-email"
                        v-model="email"
                        type="email"
                        autocomplete="email"
                        placeholder="correo@ejemplo.com"
                        :aria-invalid="!!errorEmail"
                    >
                    <small v-if="errorEmail" class="FieldError">{{ errorEmail }}</small>
                </div>
                <div class="Field">
                    <label for="reg-user">Nombre de usuario</label>
                    <input
                        id="reg-user"
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
                    <label for="reg-pass">Contraseña</label>
                    <div class="PassRow">
                        <input
                            id="reg-pass"
                            v-model="password"
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
                    <label for="reg-confirm">Repetí la contraseña</label>
                    <input
                        id="reg-confirm"
                        v-model="confirm"
                        :type="verClave ? 'text' : 'password'"
                        autocomplete="new-password"
                        placeholder="Una vez más, por las dudas"
                        :aria-invalid="!!errorConfirm"
                    >
                    <small v-if="errorConfirm" class="FieldError">{{ errorConfirm }}</small>
                </div>
                <p v-if="errorGeneral" class="FormError" role="alert">
                    <IconAlertCircle stroke="2" />
                    {{ errorGeneral }}
                </p>
                <button class="BtnPrimary" type="submit" :disabled="ocupado">
                    <IconUserPlus stroke="2" />
                    {{ ocupado ? 'Creando...' : 'Crear mi cuenta' }}
                </button>
                <p class="Switch">
                    ¿Ya tenés cuenta?
                    <RouterLink to="/login">Entrá acá</RouterLink>
                </p>
            </form>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Register{
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
