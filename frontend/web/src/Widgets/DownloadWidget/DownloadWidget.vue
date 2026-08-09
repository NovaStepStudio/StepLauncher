<script setup lang="ts">
import { computed } from 'vue';
import { IconDownload } from '@tabler/icons-vue';
import { download } from '@/Stores/Downloads';
import { hasVersions } from '@/Stores/Launcher';

const emit = defineEmits<{ (e: 'open'): void }>();

const percent = computed(() =>
    Math.round(Math.min(100, Math.max(0, download.value?.percent ?? 0)))
);
</script>

<template>
    <button class="DownloadWidget" :class="{ gap: hasVersions }" @click="emit('open')" title="Ver descarga">
        <span class="DownloadWidget_Head">
            <span class="DownloadWidget_Icon"><IconDownload stroke="2" /></span>
            <span class="DownloadWidget_Txt">
                <span class="DownloadWidget_Title">Descargando {{ download?.version }}</span>
                <span class="DownloadWidget_Sub">Progreso {{ percent }}%</span>
            </span>
        </span>
        <span class="DownloadWidget_Bar">
            <span class="DownloadWidget_BarFill" :style="{ width: percent + '%' }"></span>
        </span>
    </button>
</template>

<style scoped lang="scss">
.DownloadWidget {
    position: fixed;
    right: 0.75rem;
    bottom: 0.75rem;
    z-index: 5;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    width: 16rem;
    max-width: calc(100vw - 1.5rem);
    padding: 0.65rem 0.75rem;
    border-radius: 0.6rem;
    background: var(--background-modal-primary);
    border: var(--border-style);
    box-shadow: var(--shadow-settings-normal) #000a;
    backdrop-filter: var(--filter-blur);
    color: var(--text-primary);
    text-shadow: var(--text-shadow-primary, none);
    font-family: var(--font-secundary), Arial, sans-serif;
    text-align: left;
    cursor: pointer;
    transition: border-color 150ms, transform 150ms;

    &.gap {
        bottom: 5rem;
    }

    &:hover {
        border-color: color-mix(in srgb, var(--background-button-primary) 75%, white 15%);
        transform: translateY(-2px);
    }
}

.DownloadWidget_Head {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    min-width: 0;
}

.DownloadWidget_Icon {
    width: 2.1rem;
    height: 2.1rem;
    flex-shrink: 0;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: color-mix(in srgb, var(--background-button-primary) 60%, transparent);
    border: var(--border-style);
    color: var(--text-primary);

    svg {
        width: 1.05rem;
        height: 1.05rem;
    }
}

.DownloadWidget_Txt {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.06rem;
}

.DownloadWidget_Title {
    font-size: 0.74rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.DownloadWidget_Sub {
    font-size: 0.64rem;
    opacity: 0.6;
}

.DownloadWidget_Bar {
    height: 0.42rem;
    border-radius: 0.3rem;
    background: rgba(0, 0, 0, 0.3);
    border: var(--border-style);
    overflow: hidden;
}

.DownloadWidget_BarFill {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--progress-color, var(--background-button-primary));
    transition: width 200ms ease;
}
</style>