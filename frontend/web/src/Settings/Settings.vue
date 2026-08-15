<script lang="ts">
export interface SectionConfig {
    name: string;
    icon?: any;
    component: any;
}
</script>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { IconX, IconSettings } from '@tabler/icons-vue';
import appInfo from '../../../../wails.json';
import { previewColorFieldId } from './Colorfield';
import { useOverlayEscape } from '@/Common/Composables/useOverlayEscape';

const appName = appInfo.name ?? 'StepLauncher';
const appVersion = appInfo.version ?? '0.0.0';

const props = defineProps<{
    visible: boolean;
    sections?: SectionConfig[];
    initialSection?: string;
}>();

const emit = defineEmits<{
    (e: 'update:visible', val: boolean): void;
}>();

const previewMode = computed(() => previewColorFieldId.value !== null);

const defaultSections: SectionConfig[] = [];
const mergedSections = computed(() => props.sections ?? defaultSections);
const activeIndex = ref(0);
const activeSection = computed(() => mergedSections.value[activeIndex.value] ?? null);

watch(
    () => props.visible,
    (v) => {
        if (!v) {
            activeIndex.value = 0;
            return;
        }
        if (props.initialSection) {
            const idx = mergedSections.value.findIndex((s) => s.name === props.initialSection);
            if (idx >= 0) activeIndex.value = idx;
        }
    }
);

function close() {
    emit('update:visible', false);
}

useOverlayEscape(close, { isActive: () => props.visible });

function onOverlayClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains('SettingsModal_Overlay')) {
        if (previewMode.value) return;
        close();
    }
}
</script>

<template>
    <Teleport to="body">
        <Transition name="SettingsModal">
            <div v-if="visible" class="SettingsModal_Overlay" :class="{ preview: previewMode }" @click="onOverlayClick">
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
                                <KeepAlive>
                                    <component
                                        v-if="activeSection"
                                        :key="activeSection.name"
                                        :is="activeSection.component"
                                    />
                                </KeepAlive>
                            </Transition>
                        </main>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped lang="scss">
@use './Styles/Settings.scss';
</style>
