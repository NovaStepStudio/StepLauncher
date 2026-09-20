<script setup lang="ts">
import { computed } from 'vue';
import { IconDownload } from '@tabler/icons-vue';
import { allActiveDownloads } from '@/Instances/Store';
import { hasVersions } from '@/Launcher/Store';

const emit = defineEmits<{ (e: 'open'): void }>();

const count = computed(() => allActiveDownloads.value.length);

const percent = computed(() => {
    const d = allActiveDownloads.value[0];
    const p = d?.percent ?? 0;
    return Math.round(Math.min(100, Math.max(0, p)));
});

const title = computed(() => {
    if (count.value > 1) return `${count.value} descargas activas`;
    const d = allActiveDownloads.value[0];
    if (!d) return 'Descargando…';
    if (d.kind === 'loader') return `Instalando ${d.loader ?? 'modloader'} en ${d.label}`;
    return `Descargando ${d.label}`;
});

const sub = computed(() => {
    if (count.value > 1) return 'Pulsa para ver el detalle';
    const d = allActiveDownloads.value[0];
    if (!d) return 'Progreso 0%';
    if (d.kind === 'loader') {
        switch (d.phase) {
            case 'resolving':
                return 'Resolviendo modloader…';
            case 'downloading':
                return d.message || `Descargando ${d.loader ?? 'modloader'}…`;
            case 'installing':
                return d.message || `Instalando ${d.loader ?? 'modloader'}…`;
            default:
                return `Progreso ${percent.value}%`;
        }
    }
    if (d.state === 'verifying' || d.state === 'redownloading') {
        return `Verificando · ${d.filesDownloaded}/${d.filesTotal}`;
    }
    return `Progreso ${percent.value}%`;
});
</script>

<template>
    <button class="DownloadWidget" :class="{ gap: hasVersions }" @click="emit('open')" title="Ver descargas">
        <span class="DownloadWidget_Head">
            <span class="DownloadWidget_Icon"><IconDownload stroke="2" /></span>
            <span class="DownloadWidget_Txt">
                <span class="DownloadWidget_Title">{{ title }}</span>
                <span class="DownloadWidget_Sub">{{ sub }}</span>
            </span>
        </span>
        <span class="DownloadWidget_Bar">
            <span class="DownloadWidget_BarFill" :style="{ width: percent + '%' }"></span>
        </span>
    </button>
</template>

<style scoped lang="scss">
@use './Styles/Widget.scss';
</style>
