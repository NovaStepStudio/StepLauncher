// Tarjeta de seguridad del panel: cambio de email (se confirma en el
// correo nuevo), cambio de contraseña (exige la actual) y opciones de
// privacidad (quién puede encontrarte y qué puede llegarte).
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { IconAlertCircle, IconCircleCheck } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { pedirCambioEmail, cambiarContrasena, miPrivacidad, actualizarPrivacidad, type Privacy } from '@/Auth/Api';
import { validarEmail, validarPassword } from '@/Auth/validation';
import { ApiError } from '@/Auth/Api';

const { conAuth } = useAuth();

const nuevoEmail = ref('');
const errorEmail = ref('');
const avisoEmail = ref('');
const errorEmailApi = ref('');
const enviandoEmail = ref(false);

const actual = ref('');
const nueva = ref('');
const repite = ref('');
const errorClave = ref('');
const avisoClave = ref('');
const errorClaveApi = ref('');
const cambiando = ref(false);

async function guardarEmail() {
    errorEmail.value = validarEmail(nuevoEmail.value);
    errorEmailApi.value = '';
    avisoEmail.value = '';
    if (errorEmail.value) return;
    enviandoEmail.value = true;
    try {
        await conAuth((token) => pedirCambioEmail(token, nuevoEmail.value.trim().toLowerCase()));
        avisoEmail.value = 'Listo: confirmalo desde tu correo nuevo. Tu correo viejo recibe un aviso del cambio.';
        nuevoEmail.value = '';
    } catch (err) {
        errorEmailApi.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        enviandoEmail.value = false;
    }
}

async function guardarClave() {
    const sinActual = actual.value ? '' : 'Ingresá tu contraseña actual.';
    const malaNueva = validarPassword(nueva.value);
    const distinta = nueva.value && nueva.value === actual.value ? 'La nueva tiene que ser distinta a la actual.' : '';
    const coincide = repite.value === nueva.value ? '' : 'Las contraseñas nuevas no coinciden.';
    errorClave.value = sinActual || malaNueva || distinta || coincide;
    errorClaveApi.value = '';
    avisoClave.value = '';
    if (errorClave.value) return;
    cambiando.value = true;
    try {
        await conAuth((token) => cambiarContrasena(token, { currentPassword: actual.value, newPassword: nueva.value }));
        avisoClave.value = 'Contraseña actualizada. Te avisamos por correo.';
        actual.value = '';
        nueva.value = '';
        repite.value = '';
    } catch (err) {
        errorClaveApi.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        cambiando.value = false;
    }
}

const OPCIONES = [
    { key: 'searchable', titulo: 'Que puedan buscarte', detalle: 'Si lo apagás, no aparecés en búsquedas aunque escriban tu nombre exacto.' },
    { key: 'allowEmailSearch', titulo: 'Buscarte por email', detalle: 'Permite que te encuentren y te añadan con tu correo.' },
    { key: 'receiveFriendRequests', titulo: 'Recibir solicitudes', detalle: 'Si lo apagás, nadie puede enviarte solicitudes de amistad.' },
    { key: 'receiveNotifications', titulo: 'Recibir notificaciones', detalle: 'El sistema deja de crearte notificaciones (solicitudes, aceptaciones).' },
] as const;

type ClavePrivacidad = (typeof OPCIONES)[number]['key'];

const privacidad = ref<Privacy>({ searchable: true, allowEmailSearch: true, receiveFriendRequests: true, receiveNotifications: true });
const cargandoPrivacidad = ref(true);
const guardandoClave = ref<ClavePrivacidad | null>(null);
const errorPrivacidad = ref('');

async function cargarPrivacidad(): Promise<void> {
    cargandoPrivacidad.value = true;
    errorPrivacidad.value = '';
    try {
        privacidad.value = await conAuth((token) => miPrivacidad(token));
    } catch (err) {
        errorPrivacidad.value = err instanceof ApiError ? err.message : 'No se pudo cargar tu privacidad.';
    } finally {
        cargandoPrivacidad.value = false;
    }
}

async function alternar(clave: ClavePrivacidad) {
    errorPrivacidad.value = '';
    guardandoClave.value = clave;
    const previo = privacidad.value[clave];
    privacidad.value = { ...privacidad.value, [clave]: !previo };
    try {
        privacidad.value = await conAuth((token) => actualizarPrivacidad(token, { [clave]: !previo }));
    } catch (err) {
        privacidad.value = { ...privacidad.value, [clave]: previo };
        errorPrivacidad.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        guardandoClave.value = null;
    }
}

onMounted(() => cargarPrivacidad());
</script>

<template>
    <section class="Card" aria-label="Seguridad">
        <form class="Block" @submit.prevent="guardarEmail" novalidate>
            <b>Cambiar correo</b>
            <div class="Field">
                <label for="dash-email">Correo nuevo</label>
                <input id="dash-email" v-model="nuevoEmail" type="email" autocomplete="email" placeholder="nuevo@ejemplo.com">
                <small v-if="errorEmail" class="FieldError">{{ errorEmail }}</small>
            </div>
            <p v-if="errorEmailApi" class="FormError" role="alert">
                <IconAlertCircle stroke="2" />
                {{ errorEmailApi }}
            </p>
            <p v-if="avisoEmail" class="Ok" role="status">
                <IconCircleCheck stroke="2" />
                {{ avisoEmail }}
            </p>
            <button type="submit" class="GhostBtn" :disabled="enviandoEmail">
                {{ enviandoEmail ? 'Enviando...' : 'Pedir cambio de correo' }}
            </button>
        </form>
        <form class="Block" @submit.prevent="guardarClave" novalidate>
            <b>Cambiar contraseña</b>
            <div class="Field">
                <label for="dash-cur">Contraseña actual</label>
                <input id="dash-cur" v-model="actual" type="password" autocomplete="current-password">
            </div>
            <div class="Field">
                <label for="dash-new">Contraseña nueva</label>
                <input id="dash-new" v-model="nueva" type="password" autocomplete="new-password" placeholder="Mínimo 8 caracteres">
            </div>
            <div class="Field">
                <label for="dash-rep">Repetí la nueva</label>
                <input id="dash-rep" v-model="repite" type="password" autocomplete="new-password">
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
            <button type="submit" class="GhostBtn" :disabled="cambiando">
                {{ cambiando ? 'Guardando...' : 'Cambiar contraseña' }}
            </button>
        </form>
        <div class="Block">
            <b>Privacidad</b>
            <p v-if="cargandoPrivacidad" class="State">Cargando tu privacidad...</p>
            <ul v-else class="Toggles">
                <li v-for="op in OPCIONES" :key="op.key">
                    <div class="Txt">
                        <b>{{ op.titulo }}</b>
                        <small>{{ op.detalle }}</small>
                    </div>
                    <button
                        type="button"
                        class="Toggle"
                        :class="{ on: privacidad[op.key] }"
                        role="switch"
                        :aria-checked="privacidad[op.key]"
                        :aria-label="op.titulo"
                        :disabled="guardandoClave !== null"
                        @click="alternar(op.key)"
                    >
                        <i></i>
                    </button>
                </li>
            </ul>
            <p v-if="errorPrivacidad" class="FormError" role="alert">
                <IconAlertCircle stroke="2" />
                {{ errorPrivacidad }}
            </p>
        </div>
    </section>
</template>

<style scoped lang="scss">
.Card{
    display: flex;
    flex-direction: column;
    gap: 1.4rem;
    padding: 1.5rem;
    border-radius: .9rem;
    border: 1px solid #ffffff18;
    background: #ffffff08;
    .Block{
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: .8rem;
        b{
            font-size: .85rem;
            font-family: 'Lexend';
        }
        .Field{
            display: flex;
            flex-direction: column;
            gap: .45rem;
            width: 100%;
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
    }
    .GhostBtn{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .5rem 1rem;
        border-radius: .55rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .78rem;
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
    .State{
        margin: 0;
        font-size: .8rem;
        opacity: .55;
    }
    .Toggles{
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;
        flex-direction: column;
        width: 100%;
        li{
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 1rem;
            padding: .7rem 0;
            border-bottom: 1px solid #ffffff10;
            &:last-child{
                border-bottom: 0;
            }
            .Txt{
                display: flex;
                flex-direction: column;
                gap: .2rem;
                b{
                    font-size: .8rem;
                }
                small{
                    font-size: .7rem;
                    opacity: .5;
                    line-height: 1.5;
                }
            }
            .Toggle{
                position: relative;
                flex-shrink: 0;
                width: 2.8rem;
                height: 1.6rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff10;
                cursor: pointer;
                transition: background 150ms, border-color 150ms, opacity 150ms;
                i{
                    position: absolute;
                    top: 50%;
                    left: .2rem;
                    width: 1.1rem;
                    height: 1.1rem;
                    border-radius: 99rem;
                    background: #ffffffa6;
                    transform: translate(0, -50%);
                    transition: transform 150ms, background 150ms;
                }
                &.on{
                    background: #fff;
                    border-color: #fff;
                    i{
                        background: #000;
                        transform: translate(1.2rem, -50%);
                    }
                }
                &:disabled{
                    opacity: .5;
                    cursor: wait;
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
        font-size: .8rem;
        color: #9dffb0;
        svg{
            width: 1.1rem;
            height: 1.1rem;
        }
    }
}
@media (max-width: 600px){
    .Card{
        min-width: 0;
        padding: 1rem;
        .Block{
            min-width: 0;
            width: 100%;
        }
        .GhostBtn{
            width: 100%;
            min-height: 2.75rem;
        }
        .Toggles{
            li{
                align-items: flex-start;
                .Txt{
                    min-width: 0;
                    small{
                        overflow-wrap: anywhere;
                    }
                }
            }
        }
    }
}
</style>
