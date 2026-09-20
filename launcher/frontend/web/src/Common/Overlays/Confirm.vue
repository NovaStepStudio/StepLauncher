<script setup lang="ts">
import { ref } from 'vue';
import { resolveConfirm, type AskResult } from './Store';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const props = withDefaults(
    defineProps<{
        title: string;
        message?: string;
        confirmLabel?: string;
        cancelLabel?: string;
        danger?: boolean;
        inputLabel?: string;
        inputPlaceholder?: string;
        inputValue?: string;
    }>(),
    { confirmLabel: 'Confirmar', cancelLabel: 'Cancelar', danger: false, inputValue: '' }
);

const input = ref(props.inputValue);

function finish(confirmed: boolean): void {
    const result: AskResult = { confirmed, value: confirmed ? input.value : '' };
    resolveConfirm(result);
}

useOverlayEscape(() => finish(false), { priority: 3 });
</script>

<template>
    <Teleport to="body">
        <Transition name="ConfirmDialog">
            <div class="ConfirmDialog_Backdrop" @click.self="finish(false)">
                <div class="ConfirmDialog_Dialog">
                    <div class="ConfirmDialog_Icon" :class="{ danger }">
                        <svg
                            v-if="danger"
                            width="24"
                            height="24"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        ><polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /></svg>
                        <svg
                            v-else
                            width="24"
                            height="24"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        ><path d="M15 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0z" /><path d="M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2" /></svg>
                    </div>
                    <h4 class="ConfirmDialog_Title">{{ title }}</h4>
                    <p v-if="message" class="ConfirmDialog_Message">{{ message }}</p>
                    <label v-if="inputLabel" class="ConfirmDialog_Field">
                        <span>{{ inputLabel }}</span>
                        <input
                            class="SsIn"
                            v-model="input"
                            :placeholder="inputPlaceholder"
                            autocomplete="off"
                            spellcheck="false"
                            @keydown.enter="finish(true)"
                        />
                    </label>
                    <div class="ConfirmDialog_Actions">
                        <button class="SsBtn" @click="finish(false)">{{ cancelLabel }}</button>
                        <button class="SsBtn" :class="danger ? 'SsBtnDanger' : 'SsBtnPrimary'" @click="finish(true)">
                            {{ confirmLabel }}
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use '../Styles/Components/Confirm.scss';
</style>
