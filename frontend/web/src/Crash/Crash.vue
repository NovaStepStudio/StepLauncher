<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { IconAlertOctagon, IconX, IconInfoCircle, IconFolderOpen, IconFileText, IconCopy, IconCheck } from '@tabler/icons-vue';
import { crashInfo, clearCrash } from '@/Launcher/Store';
import { CLOSE_OVERLAYS_EVENT } from '@/Common/Stores/Idle';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const props = defineProps<{
    visible: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void;
}>();

type TabKey = 'data' | 'logs';
const tab = ref<TabKey>('data');
const copied = ref(false);
let copyTimer: ReturnType<typeof setTimeout> | null = null;

async function copyLog() {
    const text = crashInfo.value?.gameOutputText ?? crashInfo.value?.crashLogText ?? '';
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

async function openLog(path: string | undefined) {
    if (!path) return;
    try {
        await (window as any).go?.main?.App?.OpenPath?.(path);
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

function categoryName(cat: string | undefined): string {
    switch (cat) {
        case 'oom_or_killed':
        case 'oom':
            return 'Memoria agotada';
        case 'game_crash':
            return 'Crash del juego';
        case 'early_crash':
            return 'Fallo en el arranque';
        case 'shutdown_crash':
            return 'Crash al cerrar';
        case 'version_error':
            return 'Versión inválida';
        case 'watchdog':
            return 'Cliente colgado';
        case 'java_vm_crash':
            return 'Fallo de la JVM';
        case 'game_error':
            return 'Error del juego';
        case 'jvm_launch':
            return 'Fallo al iniciar Java';
        case 'killed':
            return 'Proceso terminado';
        case 'interrupted':
            return 'Interrumpido';
        case 'signal':
            return 'Señal del sistema';
        default:
            return cat && cat !== 'unknown' ? cat : 'Crash inesperado';
    }
}

function fmtTimestamp(ts: string | undefined): string {
    if (!ts) return '—';
    const d = new Date(ts);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString(undefined, {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

function javaName(javaExec: string | undefined): string {
    if (!javaExec) return '—';
    const clean = javaExec.replace(/\\/g, '/');
    const parts = clean.split('/');
    const last = parts[parts.length - 1] || clean;
    return last.includes('java') ? last : last || clean;
}

function fmtUptime(ms: number | undefined): string {
    if (typeof ms !== 'number' || ms < 0) return '—';
    const s = Math.floor(ms / 1000);
    const m = Math.floor(s / 60);
    return `${m} min ${s % 60} s`;
}

function close() {
    clearCrash();
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible });

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
                            <span class="CrashModal_CatBadge">{{ categoryName(crashInfo?.crashCategory) }}</span>
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
                        <button class="CrashModal_Tab" :class="{ active: tab === 'logs' }" @click="tab = 'logs'">
                            <IconFolderOpen stroke="2" />
                            Logs y lanzamiento
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
                                    <span class="CrashModal_Key">Fecha y hora</span>
                                    <span class="CrashModal_Val">{{ fmtTimestamp(crashInfo?.timestamp) }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Instancia</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.instanceId || '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Jugador</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.playerName || '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">PID</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.pid ?? '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Tiempo en juego</span>
                                    <span class="CrashModal_Val">{{ fmtUptime(crashInfo?.uptimeMs) }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Estado</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.status || '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">RAM máxima</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.maxRam ? `${crashInfo.maxRam} MB` : '—' }}</span>
                                </div>
                                <div class="CrashModal_Cell">
                                    <span class="CrashModal_Key">Java</span>
                                    <span class="CrashModal_Val">{{ javaName(crashInfo?.javaExec) }}</span>
                                </div>
                                <div class="CrashModal_Cell CrashModal_CellFull">
                                    <span class="CrashModal_Key">Motivo</span>
                                    <span class="CrashModal_Val">{{ crashInfo?.crashReason || '—' }}</span>
                                </div>
                            </div>

                            <div class="CrashModal_Explain">
                                <IconAlertOctagon stroke="2" />
                                <span>{{ categoryLabel(crashInfo?.crashCategory) }}</span>
                            </div>
                        </template>

                        <template v-else>
                            <div class="CrashModal_LogsList">
                                <div v-for="(item, i) in [
                                    { label: 'Log del launcher', path: crashInfo?.launcherLogPath },
                                    { label: 'Log de Minecraft', path: crashInfo?.minecraftLogPath },
                                    { label: 'Reporte de la JVM', path: crashInfo?.jvmLogPath },
                                ].filter((l) => !!l.path)" :key="i" class="CrashModal_LogRow">
                                    <div class="CrashModal_LogRowInfo">
                                        <span class="CrashModal_Key">{{ item.label }}</span>
                                        <span class="CrashModal_LogPath">{{ item.path }}</span>
                                    </div>
                                    <button class="CrashModal_OpenBtn" title="Abrir en el explorador" @click="openLog(item.path)">
                                        <IconFolderOpen stroke="2" />
                                        Abrir
                                    </button>
                                </div>
                                <p v-if="!crashInfo?.launcherLogPath && !crashInfo?.minecraftLogPath && !crashInfo?.jvmLogPath" class="CrashModal_Explain">
                                    No hay rutas de logs disponibles para este crash.
                                </p>
                            </div>
                            <div class="Logs_Code">
                                <div v-if="crashInfo?.gameOutputText" class="CrashModal_LogBox CrashModal_LogBoxMain">
                                    <div class="CrashModal_LogHead">
                                        <span class="CrashModal_LogTitle">
                                            <IconFileText stroke="2" />
                                            Output del juego
                                        </span>
                                        <button
                                            class="CrashModal_Copy"
                                            :class="{ copied }"
                                            title="Copiar output"
                                            @click="copyLog"
                                        >
                                            <IconCheck v-if="copied" stroke="2" />
                                            <IconCopy v-else stroke="2" />
                                            <span>{{ copied ? 'Copiado' : 'Copiar' }}</span>
                                        </button>
                                    </div>
                                    <pre class="CrashModal_Log">{{ crashInfo.gameOutputText }}</pre>
                                </div>
                                <p v-else class="CrashModal_Explain">
                                    No se pudo recuperar el output del juego para este crash.
                                </p>

                                <div class="CrashModal_LaunchBox">
                                    <div class="CrashModal_LogHead">
                                        <span class="CrashModal_LogTitle">
                                            <IconFileText stroke="2" />
                                            Resumen de lanzamiento
                                        </span>
                                    </div>
                                    <pre v-if="crashInfo?.launchInfo" class="CrashModal_Log CrashModal_LaunchPre">{{ crashInfo.launchInfo }}</pre>
                                    <p v-else class="CrashModal_Explain">
                                        No se pudo recuperar el resumen de lanzamiento para este crash.
                                    </p>
                                </div>
                            </div>
                        </template>
                    </div>

                    <div class="CrashModal_Footer">
                        <button
                            v-if="crashInfo?.gameOutputText || crashInfo?.crashLogText"
                            class="SsBtn CrashModal_CopyBtn"
                            :class="{ copied }"
                            @click="copyLog"
                        >
                            <IconCheck v-if="copied" stroke="2" />
                            <IconCopy v-else stroke="2" />
                            {{ copied ? 'Copiado' : 'Copiar output' }}
                        </button>
                        <button class="SsBtn SsBtnPrimary" @click="close">Cerrar</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Crash.scss';
</style>
