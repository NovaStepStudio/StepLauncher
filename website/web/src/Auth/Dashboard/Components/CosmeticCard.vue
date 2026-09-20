// Card de un cosmético (skin o capa): render 3D del personaje con SkinStage,
// botones para ampliar (modal grande), descargar y borrar.
<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import { IconMaximize, IconDownload, IconTrash, IconX, IconAlertCircle } from '@tabler/icons-vue';
import SkinStage from './SkinStage.vue';
import { useAuth } from '@/Auth/Composables/useAuth';
import { borrarArchivo, type FileUpload } from '@/Auth/Api';
import { ApiError } from '@/Auth/Api';
import { urlFirmada, olvidarUrl } from '@/Auth/firmadas';

const props = defineProps<{ item: FileUpload }>();
const emit = defineEmits<{ (e: 'borrado', id: string): void }>();

const { conAuth } = useAuth();

const vista = ref<string | null>(null);
const errorVista = ref('');
const errorAccion = ref('');
const confirmar = ref(false);
const borrando = ref(false);
const ampliado = ref(false);

function tamano(bytes: number): string {
    if (!bytes || bytes <= 0) return '—';
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function fecha(iso: string): string {
    if (!iso) return '—';
    try {
        return new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
        return '—';
    }
}

async function cargarVista(): Promise<void> {
    try {
        vista.value = await urlFirmada(conAuth, props.item.id);
    } catch (err) {
        errorVista.value = err instanceof ApiError ? err.message : 'No se pudo cargar la vista 3D.';
    }
}

cargarVista();

function descargar() {
    if (!vista.value) return;
    errorAccion.value = '';
    const enlace = document.createElement('a');
    enlace.href = vista.value;
    enlace.download = `${props.item.kind}-${props.item.id.slice(0, 8)}.png`;
    enlace.target = '_blank';
    enlace.rel = 'noopener';
    document.body.appendChild(enlace);
    enlace.click();
    enlace.remove();
}

async function borrar() {
    if (!confirmar.value) {
        confirmar.value = true;
        return;
    }
    confirmar.value = false;
    errorAccion.value = '';
    borrando.value = true;
    try {
        await conAuth((token) => borrarArchivo(token, props.item.id));
        olvidarUrl(props.item.id);
        ampliado.value = false;
        emit('borrado', props.item.id);
    } catch (err) {
        errorAccion.value = err instanceof ApiError ? err.message : 'Servicio no disponible. Probá más tarde.';
    } finally {
        borrando.value = false;
    }
}

function tecla(evento: KeyboardEvent) {
    if (evento.key === 'Escape') ampliado.value = false;
}

function abrir() {
    errorAccion.value = '';
    ampliado.value = true;
    window.addEventListener('keydown', tecla);
}

function cerrar() {
    ampliado.value = false;
    window.removeEventListener('keydown', tecla);
}

onBeforeUnmount(() => window.removeEventListener('keydown', tecla));
</script>

<template>
    <article class="Cosmetic">
        <div class="Preview">
            <SkinStage
                v-if="vista"
                :skin="item.kind === 'skin' ? vista : null"
                :cape="item.kind === 'cape' ? vista : null"
                :ancho="220"
                :alto="250"
            />
            <span v-else class="State">{{ errorVista || 'Buscando imagen...' }}</span>
            <span class="Kind">{{ item.kind === 'skin' ? 'Skin' : 'Capa' }}</span>
        </div>
        <div class="Meta">
            <span>{{ tamano(item.sizeBytes) }} · {{ fecha(item.createdAt) }}</span>
        </div>
        <div class="RowBtns">
            <button type="button" class="Btn" :disabled="!vista" @click="abrir">
                <IconMaximize stroke="2" />
                Ampliar
            </button>
            <button type="button" class="Btn" :disabled="!vista" @click="descargar" aria-label="Descargar archivo">
                <IconDownload stroke="2" />
            </button>
            <button
                type="button"
                class="Btn danger"
                :class="{ confirm: confirmar }"
                :disabled="borrando"
                @click="borrar"
                :aria-label="confirmar ? 'Confirmar borrado' : 'Borrar archivo'"
            >
                <IconTrash stroke="2" />
                <i v-if="confirmar">¿Seguro?</i>
            </button>
        </div>
        <p v-if="errorAccion" class="FormError" role="alert">
            <IconAlertCircle stroke="2" />
            {{ errorAccion }}
        </p>
        <Teleport to="body">
            <div v-if="ampliado" class="Overlay" @click.self="cerrar" role="dialog" aria-modal="true" aria-label="Vista ampliada">
                <div class="Modal sl-fade">
                    <button type="button" class="Close" aria-label="Cerrar vista" @click="cerrar">
                        <IconX stroke="2" />
                    </button>
                    <SkinStage
                        v-if="vista"
                        :skin="item.kind === 'skin' ? vista : null"
                        :cape="item.kind === 'cape' ? vista : null"
                        :ancho="340"
                        :alto="400"
                        :zoom="true"
                    />
                    <div class="RowBtns center">
                        <button type="button" class="Btn" @click="descargar">
                            <IconDownload stroke="2" />
                            Descargar
                        </button>
                        <button
                            type="button"
                            class="Btn danger"
                            :class="{ confirm: confirmar }"
                            :disabled="borrando"
                            @click="borrar"
                        >
                            <IconTrash stroke="2" />
                            {{ confirmar ? '¿Seguro?' : 'Borrar' }}
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>
    </article>
</template>

<style scoped lang="scss">
.Cosmetic{
    display: flex;
    flex-direction: column;
    gap: .7rem;
    padding: 1rem;
    border-radius: .8rem;
    border: 1px solid #ffffff14;
    background: #00000060;
    .Preview{
        position: relative;
        display: flex;
        justify-content: center;
        .State{
            display: flex;
            justify-content: center;
            align-items: center;
            width: 220px;
            height: 250px;
            font-size: .75rem;
            opacity: .5;
            text-align: center;
            padding: 0 1rem;
        }
        .Kind{
            position: absolute;
            top: .4rem;
            left: .4rem;
            font-size: .62rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: .08em;
            padding: .2rem .55rem;
            border-radius: 99rem;
            background: #fff;
            color: #000;
        }
    }
    .Meta{
        text-align: center;
        font-size: .72rem;
        opacity: .55;
    }
    .RowBtns{
        display: flex;
        gap: .4rem;
        &.center{
            justify-content: center;
        }
        .Btn{
            display: flex;
            justify-content: center;
            align-items: center;
            gap: .4rem;
            flex: 1;
            padding: .5rem .6rem;
            border-radius: .55rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            color: #fff;
            font-size: .75rem;
            font-weight: 600;
            font-family: inherit;
            cursor: pointer;
            white-space: nowrap;
            transition: background 150ms, opacity 150ms;
            svg{
                width: 1rem;
                height: 1rem;
                flex-shrink: 0;
            }
            i{
                font-style: normal;
                font-size: .72rem;
                font-weight: 700;
            }
            &:hover:not(:disabled){
                background: #ffffff14;
            }
            &:disabled{
                opacity: .4;
                cursor: default;
            }
            &.danger{
                flex: 0 1 auto;
                &:hover:not(:disabled){
                    background: #ff6b6b22;
                }
                &.confirm{
                    border-color: #ff9d9d60;
                    background: #ff6b6b22;
                }
            }
        }
    }
    .FormError{
        display: flex;
        align-items: center;
        gap: .5rem;
        margin: 0;
        padding: .55rem .75rem;
        border-radius: .55rem;
        border: 1px solid #ff9d9d45;
        background: #ff6b6b14;
        font-size: .75rem;
        line-height: 1.5;
        svg{
            flex-shrink: 0;
            width: 1rem;
            height: 1rem;
            color: #ff9d9d;
        }
    }
}
.Overlay{
    position: fixed;
    inset: 0;
    z-index: 200;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 1.5rem;
    background: #000000d9;
    backdrop-filter: blur(8px);
    .Modal{
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        max-width: min(26rem, 100%);
        max-height: 100%;
        overflow-y: auto;
        padding: 2.2rem 1.2rem 1.2rem 1.2rem;
        border-radius: .9rem;
        border: 1px solid #ffffff20;
        background: #0b0b0f;
        .Close{
            position: absolute;
            top: .6rem;
            right: .6rem;
            display: flex;
            justify-content: center;
            align-items: center;
            width: 2.2rem;
            height: 2.2rem;
            border-radius: .55rem;
            border: 1px solid #ffffff25;
            background: #ffffff08;
            color: #fff;
            cursor: pointer;
            transition: background 150ms;
            svg{
                width: 1.1rem;
                height: 1.1rem;
            }
            &:hover{
                background: #ffffff14;
            }
        }
        .RowBtns{
            display: flex;
            gap: .4rem;
            width: 100%;
            justify-content: center;
            .Btn{
                display: flex;
                justify-content: center;
                align-items: center;
                gap: .4rem;
                padding: .55rem 1rem;
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
                    opacity: .4;
                    cursor: default;
                }
                &.danger{
                    &:hover:not(:disabled){
                        background: #ff6b6b22;
                    }
                    &.confirm{
                        border-color: #ff9d9d60;
                        background: #ff6b6b22;
                    }
                }
            }
        }
    }
}
</style>
