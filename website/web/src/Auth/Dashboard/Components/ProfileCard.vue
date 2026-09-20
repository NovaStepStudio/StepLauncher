// Tarjeta de perfil del panel: muestra email, usuario, nombre, bio,
// última versión de MC y fechas de cuenta; permite editar usuario, nombre
// y bio (PATCH /v1/accounts/me; bio vacía la borra).
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { IconUser, IconPencil, IconCheck, IconAlertCircle, IconCircleCheck, IconMail, IconId, IconBrandMinecraft, IconCalendar, IconClock, IconCopy, IconHash, IconEye, IconEyeOff } from '@tabler/icons-vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { actualizarPerfil } from '@/Auth/Api';
import { validarUsername, validarDisplayName, validarBio } from '@/Auth/validation';
import { ApiError } from '@/Auth/Api';

const { profile, conAuth } = useAuth();

const editando = ref(false);
const username = ref('');
const displayName = ref('');
const bio = ref('');
const errorUsername = ref('');
const errorDisplay = ref('');
const errorBio = ref('');
const errorGeneral = ref('');
const aviso = ref('');
const guardando = ref(false);
const copiado = ref<string | null>(null);

// El correo se muestra enmascarado por defecto (seguro para compartir
// pantalla): se revela 15 segundos y se vuelve a ocultar solo.
const emailVisible = ref(false);
let temporizadorEmail: ReturnType<typeof setTimeout> | null = null;

const emailMascara = computed(() => {
    const email = profile.value?.email || '';
    const partes = email.split('@');
    const local = partes[0] || '';
    const dominio = partes[1] || '';
    if (!local || !dominio) return '—';
    return `${local.charAt(0)}••••••@${dominio}`;
});

function revelarEmail() {
    if (temporizadorEmail) clearTimeout(temporizadorEmail);
    emailVisible.value = true;
    temporizadorEmail = setTimeout(() => {
        emailVisible.value = false;
        temporizadorEmail = null;
    }, 15000);
}

onBeforeUnmount(() => {
    if (temporizadorEmail) clearTimeout(temporizadorEmail);
});

async function copiarTexto(valor: string, clave: string) {
    if (!valor) return;
    try {
        await navigator.clipboard.writeText(valor);
    } catch {
        const area = document.createElement('textarea');
        area.value = valor;
        document.body.appendChild(area);
        area.select();
        document.execCommand('copy');
        area.remove();
    }
    copiado.value = clave;
    setTimeout(() => {
        if (copiado.value === clave) copiado.value = null;
    }, 2000);
}

const bioLongitud = computed(() => bio.value.length);

function fechaCorta(iso: string | null | undefined): string {
    if (!iso) return '—';
    try {
        return new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
        return '—';
    }
}

function fechaSesion(iso: string | null | undefined): string {
    if (!iso) return 'Todavía no entraste';
    try {
        return new Date(iso).toLocaleString('es-AR', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
    } catch {
        return '—';
    }
}

watch(
    () => profile.value,
    (p) => {
        if (!editando.value && p) {
            username.value = p.username ?? '';
            displayName.value = p.displayName ?? '';
            bio.value = p.bio ?? '';
        }
    },
    { immediate: true },
);

function empezarEdicion() {
    username.value = profile.value?.username ?? '';
    displayName.value = profile.value?.displayName ?? '';
    bio.value = profile.value?.bio ?? '';
    errorUsername.value = '';
    errorDisplay.value = '';
    errorBio.value = '';
    errorGeneral.value = '';
    aviso.value = '';
    editando.value = true;
}

async function guardar() {
    errorUsername.value = validarUsername(username.value);
    errorDisplay.value = validarDisplayName(displayName.value);
    errorBio.value = validarBio(bio.value);
    errorGeneral.value = '';
    aviso.value = '';
    if (errorUsername.value || errorDisplay.value || errorBio.value) return;
    guardando.value = true;
    try {
        const actualizado = await conAuth((token) =>
            actualizarPerfil(token, {
                username: username.value.trim(),
                displayName: displayName.value.trim(),
                bio: bio.value,
            }),
        );
        profile.value = actualizado;
        editando.value = false;
        aviso.value = 'Perfil actualizado.';
    } catch (err) {
        errorGeneral.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        guardando.value = false;
    }
}
</script>

<template>
    <section class="Perfil" aria-label="Perfil">
        <div class="TopRow">
            <button v-if="!editando" type="button" class="GhostBtn" @click="empezarEdicion">
                <IconPencil stroke="2" />
                Editar
            </button>
        </div>
        <div v-if="!editando" class="Vista">
            <blockquote v-if="profile?.bio" class="Bio">
                <p>{{ profile.bio }}</p>
            </blockquote>
            <div class="Tiles">
                <div class="Tile wide">
                    <span class="Ico"><IconMail stroke="2" /></span>
                    <div>
                        <span>Correo</span>
                        <b>{{ emailVisible ? (profile?.email ?? '—') : emailMascara }}</b>
                    </div>
                    <button
                        v-if="profile?.email"
                        type="button"
                        class="Eye"
                        :aria-label="emailVisible ? 'Ocultar correo' : 'Mostrar correo'"
                        :title="emailVisible ? 'Ocultar (se oculta solo)' : 'Mostrar 15 segundos'"
                        @click="emailVisible ? (emailVisible = false) : revelarEmail()"
                    >
                        <component :is="emailVisible ? IconEyeOff : IconEye" stroke="2" />
                    </button>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconUser stroke="2" /></span>
                    <div>
                        <span>Usuario</span>
                        <b>{{ profile?.username ?? '—' }}</b>
                    </div>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconId stroke="2" /></span>
                    <div>
                        <span>Nombre visible</span>
                        <b>{{ profile?.displayName ?? '—' }}</b>
                    </div>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconBrandMinecraft stroke="2" /></span>
                    <div>
                        <span>Último MC jugado</span>
                        <b>{{ profile?.lastMcVersion ?? '—' }}</b>
                    </div>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconCalendar stroke="2" /></span>
                    <div>
                        <span>Miembro desde</span>
                        <b>{{ fechaCorta(profile?.createdAt) }}</b>
                    </div>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconClock stroke="2" /></span>
                    <div>
                        <span>Última sesión</span>
                        <b>{{ fechaSesion(profile?.lastSessionAt) }}</b>
                    </div>
                </div>
                <div class="Tile">
                    <span class="Ico"><IconHash stroke="2" /></span>
                    <div>
                        <span>UUID de cuenta</span>
                        <b class="Mono">{{ profile?.id ?? '—' }}</b>
                    </div>
                    <button
                        v-if="profile?.id"
                        type="button"
                        class="MiniCopy"
                        :aria-label="copiado === 'id' ? 'Copiado' : 'Copiar UUID de cuenta'"
                        :title="copiado === 'id' ? 'Copiado' : 'Copiar'"
                        @click="copiarTexto(profile.id, 'id')"
                    >
                        <component :is="copiado === 'id' ? IconCheck : IconCopy" stroke="2" />
                    </button>
                </div>
            </div>
            <div class="Uuid">
                <div>
                    <span>UUID de Minecraft</span>
                    <b>{{ profile?.mcUuid ?? 'Sin enlazar' }}</b>
                </div>
                <button v-if="profile?.mcUuid" type="button" class="GhostBtn" @click="copiarTexto(profile.mcUuid, 'mc')">
                    <component :is="copiado === 'mc' ? IconCheck : IconCopy" stroke="2" />
                    {{ copiado === 'mc' ? 'Copiado' : 'Copiar' }}
                </button>
            </div>
        </div>
        <form v-else class="Form" @submit.prevent="guardar" novalidate>
            <div class="Field">
                <label for="dash-user">Usuario</label>
                <input id="dash-user" v-model="username" type="text" autocomplete="username" :aria-invalid="!!errorUsername">
                <small v-if="errorUsername" class="FieldError">{{ errorUsername }}</small>
            </div>
            <div class="Field">
                <label for="dash-display">Nombre visible</label>
                <input id="dash-display" v-model="displayName" type="text" autocomplete="nickname" :aria-invalid="!!errorDisplay">
                <small v-if="errorDisplay" class="FieldError">{{ errorDisplay }}</small>
            </div>
            <div class="Field">
                <label for="dash-bio">Bio</label>
                <textarea
                    id="dash-bio"
                    v-model="bio"
                    rows="4"
                    maxlength="5000"
                    placeholder="Contanos algo sobre vos..."
                    :aria-invalid="!!errorBio"
                ></textarea>
                <small v-if="errorBio" class="FieldError">{{ errorBio }}</small>
                <small v-else class="Hint">{{ bioLongitud }} / 5000 · vacía la borra</small>
            </div>
            <p v-if="errorGeneral" class="FormError" role="alert">
                <IconAlertCircle stroke="2" />
                {{ errorGeneral }}
            </p>
            <div class="Actions">
                <button type="button" class="GhostBtn" :disabled="guardando" @click="editando = false">Cancelar</button>
                <button type="submit" class="BtnPrimary" :disabled="guardando">
                    <IconCheck stroke="2" />
                    {{ guardando ? 'Guardando...' : 'Guardar' }}
                </button>
            </div>
        </form>
        <p v-if="aviso" class="Ok" role="status">
            <IconCircleCheck stroke="2" />
            {{ aviso }}
        </p>
    </section>
</template>

<style scoped lang="scss">
.Perfil{
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
    .TopRow{
        display: flex;
        justify-content: flex-end;
    }
    .Vista{
        display: flex;
        flex-direction: column;
        gap: 1rem;
        .Bio{
            margin: 0;
            padding: .2rem 0 .2rem 1rem;
            border-left: 2px solid #ffffff30;
            p{
                margin: 0;
                font-size: .86rem;
                line-height: 1.65;
                opacity: .85;
                overflow-wrap: anywhere;
                white-space: pre-line;
            }
        }
        .Tiles{
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: .6rem;
            .Tile{
                display: flex;
                align-items: center;
                gap: .7rem;
                min-width: 0;
                padding: .7rem .8rem;
                border-radius: .65rem;
                border: 1px solid #ffffff14;
                background: linear-gradient(#ffffff0a, transparent), #00000060;
                transition: border-color 150ms, transform 150ms;
                &.wide{
                    grid-column: 1 / -1;
                }
                .Ico{
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    flex-shrink: 0;
                    width: 2.1rem;
                    height: 2.1rem;
                    border-radius: .55rem;
                    border: 1px solid #ffffff14;
                    background: #ffffff0d;
                    svg{
                        width: 1.05rem;
                        height: 1.05rem;
                        opacity: .75;
                    }
                }
                div{
                    display: flex;
                    flex-direction: column;
                    gap: .15rem;
                    min-width: 0;
                    flex: 1;
                    span{
                        font-size: .65rem;
                        text-transform: uppercase;
                        letter-spacing: .08em;
                        opacity: .45;
                    }
                    b{
                        font-size: .8rem;
                        font-weight: 600;
                        overflow: hidden;
                        text-overflow: ellipsis;
                        white-space: nowrap;
                        &.Mono{
                            font-family: monospace;
                            font-size: .68rem;
                            font-weight: 400;
                        }
                    }
                }
                &:hover{
                    border-color: #ffffff25;
                    transform: translateY(-1px);
                }
            }
            .Eye,
            .MiniCopy{
                display: flex;
                justify-content: center;
                align-items: center;
                flex-shrink: 0;
                width: 1.9rem;
                height: 1.9rem;
                border-radius: .5rem;
                border: 1px solid transparent;
                background: transparent;
                color: #ffffffa6;
                cursor: pointer;
                transition: background 150ms, color 150ms;
                svg{
                    width: 1rem;
                    height: 1rem;
                }
                &:hover{
                    background: #ffffff14;
                    color: #fff;
                }
            }
        }
        .Uuid{
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: .8rem;
            padding: .7rem .8rem;
            border-radius: .65rem;
            border: 1px dashed #ffffff25;
            background: transparent;
            div{
                display: flex;
                flex-direction: column;
                gap: .15rem;
                min-width: 0;
                span{
                    font-size: .65rem;
                    text-transform: uppercase;
                    letter-spacing: .08em;
                    opacity: .45;
                }
                b{
                    font-family: monospace;
                    font-size: .72rem;
                    font-weight: 400;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    white-space: nowrap;
                }
            }
        }
    }
    .Form{
        display: flex;
        flex-direction: column;
        gap: .9rem;
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
            input,
            textarea{
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
            textarea{
                resize: vertical;
                min-height: 5.5rem;
                line-height: 1.6;
            }
            .Hint{
                font-size: .7rem;
                opacity: .45;
            }
            .FieldError{
                font-size: .72rem;
                color: #ff9d9d;
            }
        }
        .Actions{
            display: flex;
            justify-content: flex-end;
            gap: .5rem;
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
        transition: background 150ms;
        svg{
            width: 1rem;
            height: 1rem;
        }
        &:hover:not(:disabled){
            background: #ffffff14;
        }
        &:disabled{
            opacity: .5;
            cursor: default;
        }
    }
    .BtnPrimary{
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .4rem;
        padding: .5rem 1.1rem;
        border-radius: .55rem;
        border: 1px solid #fff;
        background: #fff;
        color: #000;
        font-size: .78rem;
        font-weight: 700;
        font-family: inherit;
        cursor: pointer;
        transition: filter 150ms, opacity 150ms;
        svg{
            width: 1rem;
            height: 1rem;
        }
        &:hover:not(:disabled){
            filter: brightness(.85);
        }
        &:disabled{
            opacity: .6;
            cursor: wait;
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
@media (max-width: 500px){
    .Card{
        .Vista{
            .Tiles{
                grid-template-columns: minmax(0, 1fr);
            }
            .Uuid{
                flex-direction: column;
                align-items: stretch;
                text-align: center;
            }
        }
    }
}
</style>
