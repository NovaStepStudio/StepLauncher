<script lang="ts">
export interface SectionConfig {
    name: string;
    icon?: any;
    component: any;
}
</script>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { IconX, IconSettings } from '@tabler/icons-vue';
import appInfo from '../../../../wails.json';

const appName = appInfo.name ?? 'StepLauncher';
const appVersion = appInfo.version ?? '0.0.0';

const props = defineProps<{
    visible: boolean;
    sections?: SectionConfig[];
}>();

const emit = defineEmits<{
    (e: 'update:visible', val: boolean): void;
}>();

const defaultSections: SectionConfig[] = [];
const mergedSections = computed(() => props.sections ?? defaultSections);
const activeIndex = ref(0);
const activeSection = computed(() => mergedSections.value[activeIndex.value] ?? null);

function close() {
    emit('update:visible', false);
}

function onOverlayClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('SettingsModal_Overlay')) {
        close();
    }
}
</script>

<template>
    <Teleport to="body">
        <Transition name="SettingsModal">
            <div v-if="visible" class="SettingsModal_Overlay" @click="onOverlayClick">
                <div class="SettingsModal">
                    <div class="SettingsModal_Header">
                        <div class="SettingsModal_Title">
                            <IconSettings stroke="2" />
                            <h2>Configuracion</h2>
                        </div>
                        <button class="SettingsModal_Close" @click="close">
                            <IconX stroke="2" />
                        </button>
                    </div>
                    <div class="SettingsModal_Body">
                        <aside class="SettingsModal_Sidebar">
                            <button
                                v-for="(section, i) in mergedSections"
                                :key="i"
                                :class="['SettingsModal_SidebarItem', { active: activeIndex === i }]"
                                @click="activeIndex = i"
                            >
                                <component v-if="section.icon" :is="section.icon" stroke="2" class="SettingsModal_SidebarIcon" />
                                <span>{{ section.name }}</span>
                            </button>
                            <div v-if="mergedSections.length === 0" class="SettingsModal_Empty">
                                <p>No hay secciones registradas.</p>
                            </div>
                            <span class="SettingsModal_Version">{{ appName }} v{{ appVersion }}</span>
                        </aside>
                        <main class="SettingsModal_Content">
                            <Transition name="SettingsModal_Section" mode="out-in">
                                <component
                                    v-if="activeSection"
                                    :key="activeIndex"
                                    :is="activeSection.component"
                                />
                            </Transition>
                        </main>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
.SettingsModal_Overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: var(--filter-blur);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 10;
}

.SettingsModal {
    width: 90%;
    height: 90%;
    background: var(--background-modal-primray);
    border: 1px solid var(--border-modal-style);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: var(--shadow-settings-normal) #0005;
}

.SettingsModal_Header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 1rem;
    padding: 0.5rem 1rem;
    flex-shrink: 0;
    background: #0005;
    border-bottom: var(--border-modal-style);

    .SettingsModal_Title {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--text-primary);

        h2 {
            margin: 0;
            font-size: .75rem;
            font-family: var(--font-primary), Arial, sans-serif;
            font-weight: 600;
        }

        svg {
            width: 1rem;
        }
    }

    .SettingsModal_Close {
        background: none;
        border: none;
        color: var(--text-secondary);
        cursor: pointer;
        border-radius: 0.35rem;
        display: flex;
        justify-content: center;
        align-items: center;
        height: 1.85rem;
        width: 1.85rem;
        transition: background 150ms, color 150ms;

        &:hover {
            background: #1111;
            color: var(--text-primary);
        }
    }
}

.SettingsModal_Body {
    display: flex;
    flex: 1;
    overflow: hidden;
}

.SettingsModal_Sidebar {
    position: relative;
    width: 12.5rem;
    flex-shrink: 0;
    background: #0005;
    border-right: var(--border-modal-style);
    padding: 0.75rem;
    padding-bottom: 2.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    overflow-y: auto;
}

.SettingsModal_Version {
    position: absolute;
    bottom: .5rem;
    left: .5rem;
    font-family: var(--font-secundary), Arial, sans-serif;
    font-size: .65rem;
    opacity: .4;
    pointer-events: none;
    white-space: nowrap;
}

.SettingsModal_SidebarItem {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: .45rem .6rem;
    border-radius: 0.4rem;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.85rem;
    font-family: var(--font-secundary), Arial, sans-serif;
    cursor: pointer;
    transition: background 150ms, color 150ms;
    &:hover {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 25%);
        color: var(--text-primary);
    }
    &:active{
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 30%);
        color: var(--text-primary);        
    }
    &.active {
        background: color-mix(in srgb, var(--background-button-primary) 50%, gray 20%);
        color: var(--text-primary);
    }

    .SettingsModal_SidebarIcon {
        width: 1.1rem;
        flex-shrink: 0;
    }
}

.SettingsModal_Empty {
    padding: 1rem 0.5rem;
    text-align: center;
    color: #1115;
    font-size: 0.8rem;
}

.SettingsModal_Content {
    flex: 1;
    overflow-y: auto;
    color: var(--text-primary);
}

.SettingsModal-enter-active,
.SettingsModal-leave-active {
    transition: opacity 200ms ease;

    .SettingsModal {
        transition: transform 200ms ease, opacity 200ms ease;
    }
}

.SettingsModal-enter-from,
.SettingsModal-leave-to {
    opacity: 0;

    .SettingsModal {
        transform: scale(0.95);
        opacity: 0;
    }
}

.SettingsModal_Section-enter-active,
.SettingsModal_Section-leave-active {
    transition: opacity 80ms ease;
}

.SettingsModal_Section-enter-from,
.SettingsModal_Section-leave-to {
    opacity: 0;
}
</style>
