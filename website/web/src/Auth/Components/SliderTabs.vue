// Pestañas con pastilla blanca deslizante, mismo patrón que los tabs de
// plataforma de Descargas (ReleaseCard): un span absoluto se mueve con
// translateX(índice * 100%) porque todas las columnas miden lo mismo.
// En móvil la pastilla se oculta y la tab activa lleva el fondo blanco.
<script setup lang="ts" generic="T extends string">
import { computed, type Component } from 'vue';

const props = defineProps<{
    tabs: ReadonlyArray<{ key: T; label: string; icon?: Component }>;
    modelValue: T;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: T): void;
}>();

// Índice de la tab activa: mueve la pastilla deslizante.
const indice = computed(() => Math.max(0, props.tabs.findIndex((t) => t.key === props.modelValue)));
</script>

<template>
    <div
        class="Tabs"
        role="tablist"
        :style="{ gridTemplateColumns: `repeat(${tabs.length}, minmax(0, 1fr))`, '--tab-count': tabs.length, '--tab-index': indice }"
    >
        <span class="TabSlider" aria-hidden="true"></span>
        <button
            v-for="t in tabs"
            :key="t.key"
            type="button"
            role="tab"
            class="Tab"
            :class="{ active: modelValue === t.key }"
            :aria-selected="modelValue === t.key"
            @click="emit('update:modelValue', t.key)"
        >
            <component v-if="t.icon" :is="t.icon" stroke="2" />
            <span>{{ t.label }}</span>
        </button>
    </div>
</template>

<style scoped lang="scss">
.Tabs{
    position: relative;
    display: grid;
    padding: .35rem;
    border-radius: .7rem;
    border: 1px solid #ffffff14;
    background: #ffffff05;
    overflow-x: auto;
    .TabSlider{
        position: absolute;
        top: .35rem;
        bottom: .35rem;
        left: .35rem;
        width: calc((100% - .7rem) / var(--tab-count, 1));
        border-radius: .5rem;
        background: #fff;
        box-shadow: 0 2px 12px #00000080;
        transform: translateX(calc(var(--tab-index, 0) * 100%));
        transition: transform .38s cubic-bezier(.22, .61, .36, 1);
        pointer-events: none;
    }
    .Tab{
        position: relative;
        z-index: 1;
        display: flex;
        justify-content: center;
        align-items: center;
        gap: .45rem;
        padding: .55rem .8rem;
        border-radius: .5rem;
        border: 1px solid transparent;
        background: transparent;
        color: #ffffffa6;
        font-size: .8rem;
        font-weight: 600;
        font-family: 'Lexend';
        cursor: pointer;
        white-space: nowrap;
        transition: color 150ms;
        svg{
            width: 1.05rem;
            height: 1.05rem;
            flex-shrink: 0;
        }
        &:hover:not(.active){
            color: #fff;
        }
        &.active{
            background: transparent;
            color: #000;
            border-color: transparent;
        }
    }
}
@media (max-width: 600px){
    .Tabs{
        display: flex;
        .TabSlider{
            display: none;
        }
        .Tab{
            flex: 1 0 auto;
            &.active{
                background: #fff;
            }
        }
    }
}
@media (prefers-reduced-motion: reduce){
    .Tabs{
        .TabSlider{
            transition: none;
        }
    }
}
</style>
