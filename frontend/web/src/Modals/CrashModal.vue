<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { IconAlertOctagon, IconX, IconInfoCircle, IconCode, IconFileText, IconCopy, IconCheck } from '@tabler/icons-vue';
import { crashInfo, clearCrash } from '../Stores/Launcher';
import { CLOSE_OVERLAYS_EVENT } from '../Stores/Idle';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

type TabKey = 'data' | 'codes';
const tab = ref<TabKey>('data');
const copied = ref(false);
let copyTimer: ReturnType<typeof setTimeout> | null = null;

async function copyLog() {
    const text = crashInfo.value?.crashLogText ?? '';
    if (!text) return;
    try {
        await navigator.clipboard.writeText(text);
        copied.value = true;
        if (copyTimer !== null) clearTimeout(copyTimer);
        copyTimer = setTimeout(() => {
            copied.value = false;
            copyTimer = null;
        }, 1600);
    } catch { }
}

function categoryLabel(cat: string | undefined): string {
    switch (cat) {
        case 'oom_or_killed':
        case 'oom':
            return 'Minecraft se quedó sin memoria o fue terminado por el sistema. Sube la RAM máxima o cierra otras apps.';
        case 'game_crash':
            return 'Minecraft crasheó y generó un reporte de crash. Revisa los detalles y el log para saber la causa (mods, librerías o versión incompatible).';
        case 'early_crash':
            return 'El juego falló durante el arranque, antes de abrir la ventana (módulos, argumentos o librería de arranque). Suele indicar una instalación incompleta o corrupta.';
        case 'shutdown_crash':
            return 'El juego se crasheó mientras se cerraba. No afecta al guardado: revisa el log para confirmar.';
        case 'version_error':
            return 'El juego no pudo interpretar la versión instalada (cliente corrupto o incompleto). Reinstala o actualiza la versión.';
        case 'watchdog':
            return 'El watchdog del juego detectó que el cliente quedó colgado (hilo principal bloqueado) y lo terminó.';
        case 'java_vm_crash':
            return 'La máquina virtual de Java falló (segfault/abort). Suele deberse al driver de gráficos o al Java usado.';
        case 'game_error':
            return 'Minecraft terminó con un error de juego (mods, librerías o versión incompatible).';
        case 'jvm_launch':
            return 'Java no pudo iniciarse (máquina virtual, clase principal o argumentos inválidos).';
        case 'killed':
            return 'El proceso fue terminado externamente.';
        case 'interrupted':
            return 'El juego recibió una señal de interrupción.';
        case 'signal':
            return 'El proceso terminó por una señal del sistema.';
        default:
            return 'El juego salió con un código inesperado. Revisa los detalles y el log para más información.';
    }
}

function fmtUptime(ms: number | undefined): string {
    if (typeof ms !== 'number' || ms < 0) return '—';
    const s = Math.floor(ms / 1000);
    const m = Math.floor(s / 60);
    return `${m} min ${s % 60} s`;
}

function fmtExit(code: number | undefined): string {
    return typeof code === 'number' ? String(code) : '—';
}

function close() {
    clearCrash();
    emit('update:visible', false);
}

function onCloseOverlays() {
    close();
}

onMounted(() => {
    window.addEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});

onUnmounted(() => {
    window.removeEventListener(CLOSE_OVERLAYS_EVENT, onCloseOverlays);
});
</script>

<template>
    <Teleport to="body">
        <Transition name="CrashModal">
            <div v-if="visible" class="CrashModal_Overlay" @click.self="close">
                <div class="CrashModal_Dialog">
                    <div class="CrashModal_Head">
                        <span class="CrashModal_Icon"><IconAlertOctagon stroke="2" /></span>
                        <div class="CrashModal_Titles">
                            <h3>Minecraft se ha crasheado</h3>
                            <p>El juego terminó de forma inesperada</p>
                        </div>
                        <button class="CrashModal_Close" title="Cerrar" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>

                    <div class="CrashModal_Tabs">
                        <button class="CrashModal_Tab" :class="{ active: tab === 'data' }" @click="tab = 'data'">
                            <IconInfoCircle stroke="2" />
                            Datos del error
                        </button>
                        <button class="CrashModal_Tab" :class="{ active: tab === 'codes' }" @click="tab = 'codes'">
                            <IconCode stroke="2" />
                            Códigos de errores
                        </button>
                    </div>

                    <div class="CrashModal_Body">
                        <template v-if="tab === 'data'">
                            <div class="CrashModal_Grid">
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Versión</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.version ?? '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Código de salida</span>
                                    <span class="CrashModal_Val">{{ fmtExit(crashInfo?.exitCode) }}</span>
                                </div>
                                <div class="CrashModal_Cell CrashModal_CellFull">
                                    <span class="CrashModal_Key">Categoría</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.crashCategory ?? '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell CrashModal_CellFull">
                                    <span class="CrashModal_Key">Motivo</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.crashReason ?? '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Tiempo en juego</span>
                                    <span class="CrashModal_Val">{{ fmtUptime(crashInfo?.uptimeMs) }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">PID</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.pid ?? '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Jugador</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.playerName ?? '—' }}</span>
                                </div>
                            </div>

                            <p class="CrashModal_Explain">{{ categoryLabel(crashInfo?.crashCategory) }}</p>
                        </template>

                        <template v-else>
                            <div class="CrashModal_CodesRow">
                                <div class="CrashModal_Cell CrashModal_CellCode">
                                    <span class="CrashModal_Key">Código de error</span>
                                    <span class="CrashModal_Val">{{ fmtExit(crashInfo?.exitCode) }}</span>
                                </div>
                                <div class="CrashModal_Cell CrashModal_CellCode">
                                    <span class="CrashModal_Key">Categoría</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.crashCategory ?? '—' }}</span>
                                </div>
                            </div>

                            <div v-if="crashInfo?.crashLogText" class="CrashModal_LogBox">
                                <div class="CrashModal_LogHead">
                                    <span class="CrashModal_LogTitle">
                                        <IconFileText stroke="2" />
                                        Log de error
                                    </span>
                                    <button
                                        class="CrashModal_Copy"
                                        :class="{ copied }"
                                        title="Copiar log"
                                        @click="copyLog"
                                    >
                                        <IconCheck v-if="copied" stroke="2" />
                                        <IconCopy v-else stroke="2" />
                                        <span>{{ copied ? 'Copiado' : 'Copiar' }}</span>
                                    </button>
                                </div>
                                <pre class="CrashModal_Log">{{ crashInfo.crashLogText }}</pre>
                            </div>

                            <p v-else class="CrashModal_Explain">
                                No se pudo recuperar el texto del log para este crash. Revisa la pestaña
                                «Datos del error» para más información.
                            </p>
                        </template>
                    </div>

                    <div class="CrashModal_Footer">
                        <button class="SsBtn" @click="close">Cerrar</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Settings.scss' as *;

.CrashModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 130;
}

.CrashModal_Dialog {
    width: min(80vw, 62rem);
    min-width: 65vw;
    max-width: 92vw;
    max-height: 88vh;
    background: var(--background-modal-primary);
    border: var(--border-modal-style);
    border-radius: 0.85rem;
    box-shadow: var(--shadow-settings-normal) #000a;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.CrashModal_Head {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.1rem 1.35rem 0.85rem;
    border-bottom: var(--border-modal-style);
    background: #0005;
}

.CrashModal_Icon {
    width: 2.4rem;
    height: 2.4rem;
    flex-shrink: 0;
    border-radius: 0.6rem;
    background: linear-gradient(135deg, color-mix(in srgb, var(--color-error) 35%, transparent), color-mix(in srgb, var(--color-error) 10%, transparent));
    border: 1px solid color-mix(in srgb, var(--color-error) 35%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;
    color: var(--color-error);
}

.CrashModal_Titles {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;

    h3 {
        margin: 0;
        font-family: var(--font-primary), Arial, sans-serif;
        font-size: calc(0.95rem * var(--font-size-primary, 1));
        font-weight: 600;
    }

    p {
        margin: 0;
        font-size: 0.68rem;
        opacity: 0.5;
    }
}

.CrashModal_Close {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 0.4rem;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 2rem;
    width: 2rem;
    flex-shrink: 0;
    transition: background 150ms, color 150ms;

    svg {
        width: 1.1rem;
        height: 1.1rem;
    }

    &:hover {
        background: #1111;
        color: var(--text-primary);
    }
}

.CrashModal_Icon svg {
    width: 1.3rem;
    height: 1.3rem;
}

.CrashModal_Tabs {
    display: flex;
    gap: 0.5rem;
    padding: 0.9rem 1.35rem 0;
    flex-shrink: 0;
}

.CrashModal_Tab {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    border-radius: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.72rem;
    padding: 0.45rem 0.9rem;
    cursor: pointer;
    transition: background 150ms, border-color 150ms, color 150ms;

    svg {
        width: 0.95rem;
        height: 0.95rem;
    }

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 10%);
        color: var(--text-primary);
    }

    &.active {
        border-color: color-mix(in srgb, var(--background-button-primary) 55%, gray 50%);
        background-color: color-mix(in srgb, var(--background-button-primary) 80%, gray 15%);
        color: var(--text-primary);
    }
}

.CrashModal_Body {
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    padding: 1.1rem 1.35rem;
    overflow-y: auto;
}

.CrashModal_Grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.6rem;
}

.CrashModal_Cell {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.55rem 0.7rem;
    border-radius: 0.5rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
}

.CrashModal_CellFull {
    grid-column: 1 / -1;
}

.CrashModal_CodesRow {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.6rem;
}

.CrashModal_Key {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    opacity: 0.45;
}

.CrashModal_Val {
    font-size: 0.78rem;
    font-family: var(--font-secundary), Arial, sans-serif;
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    word-break: break-word;
}

.CrashModal_Explain {
    margin: 0;
    font-size: 0.72rem;
    line-height: 1.55;
    opacity: 0.6;
}

.CrashModal_LogBox {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-radius: 0.5rem;
    background: rgba(0, 0, 0, 0.35);
    border: var(--border-style);
    overflow: hidden;
}

.CrashModal_LogHead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.6rem;
    padding: 0.55rem 0.8rem;
    border-bottom: var(--border-style);
    flex-shrink: 0;
}

.CrashModal_LogTitle {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.68rem;
    opacity: 0.75;

    svg {
        width: 0.95rem;
        height: 0.95rem;
    }
}

.CrashModal_Copy {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    background: var(--control-bg);
    border: 1px solid var(--control-border);
    color: var(--text-secondary);
    border-radius: 0.4rem;
    font-size: 0.65rem;
    padding: 0.3rem 0.6rem;
    cursor: pointer;
    transition: background 150ms, border-color 150ms, color 150ms;

    svg {
        width: 0.85rem;
        height: 0.85rem;
    }

    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 10%);
        color: var(--text-primary);
    }

    &.copied {
        border-color: color-mix(in srgb, var(--color-success, #3fb950) 60%, transparent);
        color: var(--color-success, #3fb950);
    }
}

.CrashModal_Log {
    margin: 0;
    padding: 0.8rem;
    font-family: Consolas, 'Courier New', monospace;
    font-size: 0.68rem;
    line-height: 1.5;
    color: var(--text-primary);
    opacity: 0.85;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-y: auto;
    max-height: 46vh;
}

.CrashModal_Footer {
    display: flex;
    justify-content: flex-end;
    padding: 0.85rem 1.35rem 1.1rem;
    border-top: var(--border-modal-style);
}

.CrashModal-enter-active,
.CrashModal-leave-active {
    transition: opacity 180ms ease;

    .CrashModal_Dialog {
        transition: transform 200ms ease, opacity 180ms ease;
    }
}

.CrashModal-enter-from,
.CrashModal-leave-to {
    opacity: 0;

    .CrashModal_Dialog {
        transform: translateY(8px) scale(0.97);
        opacity: 0;
    }
}
</style>
