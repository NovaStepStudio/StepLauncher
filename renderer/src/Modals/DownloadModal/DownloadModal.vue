<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <span class="modal-title">{{ props.installState.active ? $t('modals.download.title.installing') : $t('modals.download.title.download') }}</span>
        <button class="close-btn" @click="$emit('close')">
          <img :src="'assets/svg/x.svg'" width="14" height="14" />
        </button>
      </div>

      <div class="loading" v-if="loadingVersions">{{ $t('modals.download.loading_versions') }}</div>
      <div class="error-banner" v-if="loadError">{{ loadError }}</div>

      <template v-if="!loadingVersions && !loadError && !props.installState.active">
        <div class="modal-body">
          <div class="split">
            <div class="split-left">
              <label class="form-label">{{ $t('modals.download.version_label') }}</label>
              <div class="search-box">
                <img class="search-icon" :src="'assets/svg/search.svg'" width="12" height="12" />
                <input class="search-input" v-model="searchQuery" :placeholder="$t('modals.download.search_placeholder')" />
              </div>
              <div class="version-tabs">
                <button v-for="tab in versionTabs" :key="tab.key" :class="['tab', { active: activeTab === tab.key }]" @click="activeTab = tab.key">{{ $t('modals.download.tabs.' + tab.key) }}</button>
              </div>
              <div class="version-list">
                <button v-for="v in filteredVersions" :key="v.id" :class="['version-item', { selected: selectedVersion?.id === v.id }]" @click="selectVersion(v)">
                  <div class="v-left">
                    <span class="v-name">{{ v.id }}</span>
                    <span class="v-type" :class="typeClass(v.type)">{{ typeLabel(v.type) }}</span>
                  </div>
                  <img v-if="selectedVersion?.id === v.id" class="v-check" :src="'assets/svg/check.svg'" width="14" height="14" />
                </button>
                <div v-if="filteredVersions.length === 0" class="no-results">{{ $t('modals.download.no_results') }}</div>
              </div>
            </div>

            <div class="split-right">
              <label class="form-label">{{ $t('modals.download.modloader_label') }}</label>
              <div class="modloader-list">
                <button v-for="ml in modloaderOptions" :key="ml.value" :class="['ml-item', { active: selectedModloader === ml.value }]" @click="selectModloader(ml.value)">
                  <img class="ml-icon" :src="ml.icon" alt="" />
                  <span class="ml-name">{{ ml.label }}</span>
                  <span class="ml-check" v-if="selectedModloader === ml.value">
                    <img :src="'assets/svg/check.svg'" width="10" height="10" />
                  </span>
                </button>
              </div>

              <div class="summary">
                <div class="summary-line" v-if="selectedVersion">
                  <span class="sum-key">{{ $t('modals.download.summary.version') }}</span>
                  <span class="sum-val">{{ selectedVersion.id }}</span>
                </div>
                <div class="summary-line" v-if="selectedModloader !== 'none'">
                  <span class="sum-key">{{ $t('modals.download.summary.modloader') }}</span>
                  <span class="sum-val">{{ modloaderLabel }}</span>
                </div>
              </div>

              <button class="download-btn" :disabled="!canDownload" @click="startDownload">
                <img :src="'assets/svg/download.svg'" width="14" height="14" />
                {{ $t('modals.download.buttons.download') }}
              </button>

              <span v-if="installError" class="footer-error">{{ installError }}</span>
            </div>
          </div>
        </div>
      </template>

      <template v-if="props.installState.active">
        <div class="progress-body">
          <div class="progress-info">
            <span class="progress-title">{{ props.installState.instanceName }}</span>
            <span class="progress-status">{{ props.installState.statusText || $t('modals.download.progress.status') }}</span>
          </div>

          <div class="progress-bar-track">
            <div class="progress-bar-fill" :style="{ width: props.installState.percent + '%' }"></div>
          </div>
          <div class="progress-details">
            <span>{{ $t('modals.download.progress.files', { completed: props.installState.completedFiles, total: props.installState.totalFiles }) }}</span>
            <span>{{ props.installState.downloadedMb.toFixed(1) }}/{{ props.installState.totalMb.toFixed(1) }} MB</span>
            <span>{{ props.installState.percent }}%</span>
          </div>

          <div class="section-list" v-if="props.installState.modules.length > 0">
            <div v-for="m in props.installState.modules" :key="m.module" :class="['section-item', { 'active-section': m.status === 'downloading' }]">
              <div class="section-icon">
                <img v-if="m.status === 'completed'" :src="'assets/svg/check.svg'" width="14" height="14" />
                <img v-else-if="m.status === 'failed'" :src="'assets/svg/x.svg'" width="14" height="14" />
                <img v-else-if="m.status === 'downloading'" :src="'assets/svg/clock.svg'" width="14" height="14" />
                <img v-else :src="'assets/svg/stop-circle.svg'" width="14" height="14" />
              </div>
              <div class="section-info">
                <span class="section-name">{{ moduleLabel(m.module) }}</span>
              </div>
              <span class="section-status" :class="'status-' + m.status">{{ statusLabel(m.status) }}</span>
            </div>
          </div>

          <button v-if="props.installState.logs.length > 0 || showLogs" class="logs-toggle" @click="showLogs = !showLogs">
            {{ $t(showLogs ? 'modals.download.buttons.hide_logs' : 'modals.download.buttons.show_logs', { count: props.installState.logs.length }) }}
          </button>

          <div v-if="showLogs && props.installState.logs.length > 0" class="progress-logs">
            <div v-for="(l, i) in props.installState.logs" :key="i" class="log-line">{{ l }}</div>
          </div>

          <button v-if="!props.installState.error" class="cancel-install-btn" @click="emitCloseAndCancel">
            {{ $t('modals.download.buttons.cancel_install') }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { t } from '../../i18n'
import type { InstallCardState } from '../../Widgets/InstallProgressCard/InstallProgressCard.vue'

interface VersionEntry {
  id: string
  type: string
  url?: string
  releaseTime?: string
}

const props = defineProps<{
  installState: InstallCardState
}>()

const emit = defineEmits<{
  close: []
  cancel: []
  'start-install': [req: { version: string; modloader: string; modloaderVersion: string; name: string }]
}>()

const modloaderOptions = [
  { value: 'none', label: t('modals.download.options.vanilla'), icon: 'assets/loaders/Grass.webp' },
  { value: 'fabric', label: t('modals.download.options.fabric'), icon: 'assets/loaders/fabric.png' },
  { value: 'forge', label: t('modals.download.options.forge'), icon: 'assets/loaders/forge.png' },
  { value: 'neoforge', label: t('modals.download.options.neoforge'), icon: 'assets/loaders/neoforge.png' },
  { value: 'quilt', label: t('modals.download.options.quilt'), icon: 'assets/loaders/quilt.png' },
  { value: 'optifine', label: t('modals.download.options.optifine'), icon: 'assets/loaders/optifine.png' },
]

const versionTabs = [
  { key: 'all', label: t('modals.download.tabs.all') },
  { key: 'release', label: t('modals.download.tabs.release') },
  { key: 'snapshot', label: t('modals.download.tabs.snapshot') },
  { key: 'old', label: t('modals.download.tabs.old') },
] as const

const moduleLabels: Record<string, string> = {
  client: t('modals.download.modules.client'), libraries: t('modals.download.modules.libraries'), assets: t('modals.download.modules.assets'),
  natives: t('modals.download.modules.natives'), logging: t('modals.download.modules.logging'), runtime: t('modals.download.modules.runtime'),
  modloader: t('modals.download.modules.modloader'),
}

function moduleLabel(m: string): string { return moduleLabels[m] || m }

const statusLabels: Record<string, string> = {
  pending: t('modals.download.states.pending'), downloading: t('modals.download.states.downloading'),
  completed: t('modals.download.states.completed'), failed: t('modals.download.states.failed'),
}

function statusLabel(s: string): string { return statusLabels[s] || s }

const typeLabels: Record<string, string> = {
  release: t('modals.download.types.release'), snapshot: t('modals.download.types.snapshot'),
  old_beta: t('modals.download.types.beta'), old_alpha: t('modals.download.types.alpha'),
}

const typeClasses: Record<string, string> = {
  release: 'release', snapshot: 'snapshot',
  old_beta: 'old', old_alpha: 'old',
}

function typeLabel(t: string): string { return typeLabels[t] || t }
function typeClass(t: string): string { return typeClasses[t] || '' }

const loadingVersions = ref(true)
const loadError = ref('')
const allVersions = ref<VersionEntry[]>([])
const searchQuery = ref('')
const activeTab = ref<'all' | 'release' | 'snapshot' | 'old'>('release')
const selectedVersion = ref<VersionEntry | null>(null)
const selectedModloader = ref('none')
const selectedModloaderVersion = ref('auto')
const modloaderVersions = ref<string[]>([])
const loadingModloaderVersions = ref(false)
const showLogs = ref(false)
const installError = ref('')

const modloaderLabel = computed(() => {
  const ml = modloaderOptions.find(m => m.value === selectedModloader.value)
  return ml ? ml.label : selectedModloader.value
})

const filteredVersions = computed(() => {
  let list = allVersions.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) list = list.filter(v => v.id.toLowerCase().includes(q))
  if (activeTab.value === 'release') list = list.filter(v => v.type === 'release')
  else if (activeTab.value === 'snapshot') list = list.filter(v => v.type === 'snapshot')
  else if (activeTab.value === 'old') list = list.filter(v => v.type === 'old_beta' || v.type === 'old_alpha')
  return list
})

const canDownload = computed(() => !props.installState.active && !!selectedVersion.value)

onMounted(async () => {
  try {
    const vRes = await window.NovaCoreManager.Versions(200).catch(() => null)
    if (vRes) {
      const raw = (vRes as any).versions ?? vRes
      if (Array.isArray(raw)) {
        const seen = new Set<string>()
        allVersions.value = raw.filter((v: any) => {
          if (!v.id || seen.has(v.id)) return false
          seen.add(v.id)
          return true
        })
        const first = allVersions.value.find(v => v.type === 'release')
        if (first) selectedVersion.value = first
      }
    }
  } catch (err: any) {
    loadError.value = err?.message ?? t('modals.download.errors.loading_versions')
  } finally {
    loadingVersions.value = false
  }
})

function selectVersion(v: VersionEntry) {
  selectedVersion.value = v
  if (selectedModloader.value !== 'none') fetchModloaderVersions()
}

function selectModloader(value: string) {
  selectedModloader.value = value
  selectedModloaderVersion.value = ''
  if (value !== 'none' && selectedVersion.value) fetchModloaderVersions()
}

async function fetchModloaderVersions() {
  if (!selectedVersion.value || !selectedModloader.value || selectedModloader.value === 'none') return
  loadingModloaderVersions.value = true
  try {
    const res = await window.NovaCoreManager.ModLoaderVersions(selectedModloader.value, selectedVersion.value.id)
    const raw = (res as any)?.versions ?? res
    if (Array.isArray(raw)) {
      modloaderVersions.value = raw.reduce((acc: string[], v: any) => {
        const str = typeof v === 'string' ? v : (v.version ?? v.id ?? v.name)
        if (str && typeof str === 'string') acc.push(str)
        return acc
      }, [])
    } else {
      modloaderVersions.value = []
    }
  } catch {
    modloaderVersions.value = []
  } finally {
    loadingModloaderVersions.value = false
  }
}

function emitCloseAndCancel() {
  emit('cancel')
  emit('close')
}

function startDownload() {
  if (!selectedVersion.value || props.installState.active) return

  const versionId = selectedVersion.value.id
  const ml = selectedModloader.value
  const mlVer = ''
  const name = ml !== 'none' ? `${versionId}-${ml}` : versionId

  emit('start-install', { version: versionId, modloader: ml, modloaderVersion: mlVer, name })
}
</script>

<style scoped lang="scss" src="../../Styles/Modals/DownloadModal.scss"></style>
