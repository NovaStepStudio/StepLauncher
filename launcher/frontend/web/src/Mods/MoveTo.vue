<script setup lang="ts">
import { ref, computed } from 'vue';
import { IconCheck, IconX } from '@tabler/icons-vue';

const props = defineProps<{
    fileName: string;
    fromLabel: string;
    fromDestination: string;
    fromInstance: string;
    instances: Array<{ name: string; title: string }>;
}>();

const emit = defineEmits<{
    (e: 'confirm', dest: string, inst: string): void;
    (e: 'cancel'): void;
}>();

const dest = ref<'global' | 'instance'>('global');
const inst = ref('');

const options = computed(() =>
    props.instances.filter(
        (i) => !(props.fromDestination === 'instance' && i.name === props.fromInstance)
    )
);

const canGlobal = computed(() => props.fromDestination !== 'global');

const valid = computed(() => {
    if (dest.value === 'global') return canGlobal.value;
    return !!inst.value;
});

function targetLabel(): string {
    if (dest.value === 'global') return 'Juego global';
    const found = props.instances.find((i) => i.name === inst.value);
    return found?.title || inst.value || 'Instancia';
}
</script>

<template>
    <div class="ModsMove">
        <p class="ModsMove_Title">Mover <b :title="fileName">{{ fileName }}</b></p>
        <p class="ModsMove_Sub">Desde {{ fromLabel }}</p>
        <div class="ModsMove_Radios">
            <label class="ModsMove_Radio" :class="{ on: dest === 'global', off: !canGlobal }">
                <input v-model="dest" type="radio" value="global" :disabled="!canGlobal" />
                <span class="ModsMove_RadioDot"></span>
                <span>Juego global</span>
            </label>
            <label class="ModsMove_Radio" :class="{ on: dest === 'instance', off: !options.length }">
                <input v-model="dest" type="radio" value="instance" :disabled="!options.length" />
                <span class="ModsMove_RadioDot"></span>
                <span>Instancia</span>
            </label>
        </div>
        <label v-if="dest === 'instance'" class="ModsMove_Field">
            <select v-model="inst" class="SsSel">
                <option value="" disabled>Elige la instancia…</option>
                <option v-for="i in options" :key="i.name" :value="i.name">
                    {{ i.title || i.name }}
                </option>
            </select>
        </label>
        <p v-else class="ModsMove_Target">Destino: Juego global</p>
        <p v-if="dest === 'instance' && inst" class="ModsMove_Target">Destino: {{ targetLabel() }}</p>
        <div class="ModsMove_Actions">
            <button class="SsBtn" @click="emit('cancel')">
                <IconX stroke="2" /> Cancelar
            </button>
            <button class="SsBtn SsBtnPrimary" :disabled="!valid" @click="emit('confirm', dest, inst)">
                <IconCheck stroke="2" /> Mover
            </button>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use './Styles/MoveTo.scss';
</style>
