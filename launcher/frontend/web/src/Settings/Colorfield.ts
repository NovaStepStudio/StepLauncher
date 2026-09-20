import { ref } from 'vue';

let nextId = 0;

export function nextColorFieldId(): number {
    return ++nextId;
}

export const openColorFieldId = ref<number | null>(null);

export const previewColorFieldId = ref<number | null>(null);
