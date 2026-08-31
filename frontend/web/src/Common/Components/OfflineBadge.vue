<script setup lang="ts">
import { ref, computed } from 'vue';
import { IconAlertTriangle } from '@tabler/icons-vue';

const props = withDefaults(defineProps<{
    message?: string;
    placement?: 'outside' | 'inside';
    tooltip?: 'top' | 'left' | 'right' | 'bottom';
}>(), {
    placement: 'outside',
    tooltip: 'top',
});

const tip = computed(() => props.message ?? 'Sin conexión — Esta función requiere internet y no está disponible sin conexión.');

const show = ref(false);
const badgeRef = ref<HTMLElement | null>(null);
const tipStyle = ref<Record<string, string>>({});

function updatePos() {
    const el = badgeRef.value;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const gap = 8;
    const style: Record<string, string> = {
        position: 'fixed',
        zIndex: '9999',
        maxWidth: '240px',
    };
    // Medidas aproximadas del tip para centrar
    const tipW = 220;
    const tipH = 36;
    switch (props.tooltip) {
        case 'top':
            style.left = `${r.left + r.width / 2}px`;
            style.top = `${r.top - gap}px`;
            style.transform = 'translate(-50%, -100%)';
            break;
        case 'bottom':
            style.left = `${r.left + r.width / 2}px`;
            style.top = `${r.bottom + gap}px`;
            style.transform = 'translate(-50%, 0)';
            break;
        case 'left':
            style.left = `${r.left - gap}px`;
            style.top = `${r.top + r.height / 2}px`;
            style.transform = 'translate(-100%, -50%)';
            break;
        case 'right':
            style.left = `${r.right + gap}px`;
            style.top = `${r.top + r.height / 2}px`;
            style.transform = 'translate(0, -50%)';
            break;
    }
    // Evitar salir de viewport
    tipStyle.value = style;
}

function onEnter() {
    updatePos();
    show.value = true;
}
function onLeave() {
    show.value = false;
}
</script>

<template>
    <span
        ref="badgeRef"
        class="OfflineBadge"
        :class="[`placement-${placement}`]"
        :title="tip"
        aria-label="Sin conexión"
        role="img"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <IconAlertTriangle class="OfflineBadge_Icon" size="12" stroke="2.6" />
        <Teleport to="body">
            <div v-if="show" class="OfflineTip" :style="tipStyle">
                {{ tip }}
            </div>
        </Teleport>
    </span>
</template>

<style scoped>
.OfflineBadge {
    position: absolute;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #ff3b30;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid rgb(255,0,0,0.5);
    z-index: 3;
    user-select: none;
    pointer-events: auto;
}

.OfflineBadge.placement-outside {
    top: -6px;
    right: -6px;
}

.OfflineBadge.placement-inside {
    top: 4px;
    right: 6px;
}

:global(.offline-wrap) {
    overflow: visible !important;
}

.OfflineTip {
    background: #1e1e1e;
    color: #fff;
    font-size: 11px;
    font-weight: 500;
    line-height: 1.35;
    padding: 6px 8px;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
    white-space: normal;
    width: max-content;
    max-width: min(240px, 70vw);
    text-align: center;
    pointer-events: none;
    font-family: Inter, sans-serif;
}
</style>
