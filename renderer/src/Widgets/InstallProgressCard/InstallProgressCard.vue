<script setup lang="ts">
import { ref } from 'vue'
import { t } from '../../i18n'

export interface ModuleInfo {
    module: string
    status: string
}

export interface InstallCardState {
    active: boolean
    instanceName: string
    percent: number
    statusText: string
    completedFiles: number
    skippedFiles: number
    totalFiles: number
    downloadedMb: number
    totalMb: number
    modules: ModuleInfo[]
    logs: string[]
    error: string
}

const props = defineProps<{
    state: InstallCardState
}>()

const emit = defineEmits<{
    cancel: []
    close: []
}>()

const minimized = ref(false)
const showLogs = ref(false)

const moduleLabels: Record<string, string> = {
    client: t('widgets.install_progress.modules.client'),
    libraries: t('widgets.install_progress.modules.libraries'),
    assets: t('widgets.install_progress.modules.assets'),
    natives: t('widgets.install_progress.modules.natives'),
}

const moduleOrder = ['client', 'libraries', 'assets', 'natives']

const statusLabels: Record<string, string> = {
    pending: t('widgets.install_progress.states.pending'),
    downloading: t('widgets.install_progress.states.downloading'),
    completed: t('widgets.install_progress.states.completed'),
    verifying: t('widgets.install_progress.states.verifying'),
    verified: t('widgets.install_progress.states.verified'),
    failed: t('widgets.install_progress.states.failed'),
    retrying: t('widgets.install_progress.states.retrying'),
}

function modStatus(mod: string): string {
    const m = props.state.modules.find(x => x.module === mod)
    return m?.status ?? 'pending'
}

function modLabel(mod: string): string {
    const m = props.state.modules.find(x => x.module === mod)
    return m ? (statusLabels[m.status] || m.status) : t('widgets.install_progress.states.pending')
}
</script>

<template>
    <div v-if="state.active" class="install-island" :class="{ minimized, done: state.percent >= 100 && !state.error }">
        <div class="island-header" @click="minimized = !minimized">
            <div class="ih-left">
                <div class="spinner" :class="{ done: state.percent >= 100 && !state.error }"></div>
                <span class="ih-name">{{ state.instanceName }}</span>
            </div>
            <div class="ih-right">
                <span class="ih-pct">{{ state.percent }}%</span>
                <button class="ih-btn" @click.stop="showLogs = !showLogs" :title="$t('widgets.install_progress.tooltips.logs')">
                    <img :src="'assets/svg/hamburger.svg'" width="12" height="12" />
                </button>
                <button class="ih-btn ih-close" @click.stop="emit('close')" :title="$t('widgets.install_progress.tooltips.close')">
                    <img :src="'assets/svg/x.svg'" width="12" height="12" />
                </button>
            </div>
        </div>
        <div v-show="!minimized" class="island-body">
            <div class="status">{{ state.statusText }}</div>
            <div class="progress-track">
                <div class="progress-fill" :style="{ width: state.percent + '%' }"></div>
            </div>
            <div class="file-info" v-if="state.totalFiles > 0">
                {{ $t('widgets.install_progress.files_info', { completed: state.completedFiles + state.skippedFiles, total: state.totalFiles }) }}
                <span v-if="state.totalMb > 0">
                    · {{ state.downloadedMb.toFixed(1) }} / {{ state.totalMb.toFixed(1) }} MB
                </span>
            </div>
            <div class="modules">
                <template v-for="mod in moduleOrder" :key="mod">
                    <div v-if="state.modules.find(m => m.module === mod)" class="module-row">
                        <span class="module-name">{{ moduleLabels[mod] || mod }}</span>
                        <span class="module-status" :class="modStatus(mod)">{{ modLabel(mod) }}</span>
                    </div>
                </template>
                <template v-for="mod in state.modules" :key="mod.module">
                    <div v-if="!moduleOrder.includes(mod.module)" class="module-row">
                        <span class="module-name">{{ moduleLabels[mod.module] || mod.module }}</span>
                        <span class="module-status" :class="mod.status">{{ statusLabels[mod.status] || mod.status }}</span>
                    </div>
                </template>
            </div>
            <div v-if="state.error" class="error-msg">{{ state.error }}</div>
            <div v-if="showLogs && state.logs.length > 0" class="logs">
                <div v-for="(line, i) in state.logs" :key="i" class="log-line">{{ line }}</div>
            </div>
            <div v-else-if="showLogs" class="logs empty">{{ $t('widgets.install_progress.empty_logs') }}</div>
            <button v-if="state.percent < 100 && !state.error" class="cancel-btn"
                @click="emit('cancel')">{{ $t('widgets.install_progress.buttons.cancel') }}</button>
        </div>
    </div>
</template>

<style scoped lang="scss" src="../../Styles/Widgets/InstallProgressCard.scss"></style>
