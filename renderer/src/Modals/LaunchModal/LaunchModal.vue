<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <div class="header-left">
          <img :src="'assets/svg/clock.svg'" width="16" height="16" />
          <span class="modal-title">{{ $t('modals.launch.title') }}</span>
        </div>
        <button class="close-btn" @click="$emit('close')">
          <img :src="'assets/svg/x.svg'" width="14" height="14" />
        </button>
      </div>

      <div class="search-box">
        <img class="search-icon" :src="'assets/svg/search.svg'" width="14" height="14" />
        <input v-model="search" class="search-input" :placeholder="$t('modals.launch.search_placeholder')" />
      </div>

      <div class="filter-tabs">
        <button v-for="tab in filterTabs" :key="tab.id" :class="['filter-tab', { active: activeTab === tab.id }]" @click="activeTab = tab.id">
          {{ $t('modals.launch.tabs.' + tab.id) }}
        </button>
      </div>

      <div class="loading" v-if="loading">
        <div class="spinner" />
        <span>{{ $t('modals.launch.loading_versions') }}</span>
      </div>

      <div class="modal-body" v-if="!loading">
        <div v-if="groupedFiltered.length === 0" class="no-results">{{ $t('modals.launch.no_results') }}</div>
        <div v-for="group in groupedFiltered" :key="group.major" class="version-group">
          <div class="group-header" @click="toggleGroup(group.major)">
            <img :class="['group-arrow', { open: expandedGroups.has(group.major) }]" :src="'assets/svg/chevron-right.svg'" width="10" height="10" />
            <span class="group-title">{{ group.major }}</span>
            <span class="group-count">{{ group.items.length === 1 ? $t('modals.launch.group_count', { count: group.items.length }) : $t('modals.launch.group_count_plural', { count: group.items.length }) }}</span>
          </div>
          <div v-if="expandedGroups.has(group.major)" class="group-items">
            <button v-for="v in group.items" :key="v.full" :class="['version-item', { selected: selectedVersion === v.full }]" @click="select(v.full)">
              <div class="v-info">
                <span class="v-name">{{ v.full }}</span>
                <span :class="['v-badge', v.type]">{{ typeLabel(v.type) }}</span>
              </div>
              <img v-if="selectedVersion === v.full" class="v-check" :src="'assets/svg/check.svg'" width="14" height="14" />
            </button>
          </div>
        </div>
      </div>

      <div class="footer" v-if="selectedVersion">
        <button class="play-btn" @click="play">
          <img :src="'assets/svg/play.svg'" width="16" height="16" />
          {{ $t('modals.launch.play_button', { version: selectedVersion }) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { t } from '../../i18n'

const emit = defineEmits<{
  close: []
  play: [payload: { version: string }]
}>()

const loading = ref(true)
const versions = ref<string[]>([])
const selectedVersion = ref<string | null>(null)
const search = ref('')
const activeTab = ref('all')
const expandedGroups = ref(new Set<string>())

const filterTabs = [
  { id: 'all', label: t('modals.launch.tabs.all') },
  { id: 'vanilla', label: t('modals.launch.tabs.vanilla') },
  { id: 'release', label: t('modals.launch.tabs.release') },
  { id: 'snapshot', label: t('modals.launch.tabs.snapshot') },
  { id: 'beta', label: t('modals.launch.tabs.beta') },
  { id: 'alpha', label: t('modals.launch.tabs.alpha') },
]

type VersionType = 'vanilla' | 'release' | 'snapshot' | 'beta' | 'alpha' | 'forge' | 'fabric' | 'quilt' | 'neoforge' | 'optifine' | 'legacyfabric'

interface VersionInfo {
  full: string
  major: string
  type: VersionType
}

function getVersionInfo(v: string): VersionInfo {
  let type: VersionType = 'release'
  const lower = v.toLowerCase()
  if (lower.includes('snapshot') || lower.includes('pre') || lower.includes('rc') || lower.includes('w'))
    type = 'snapshot'
  else if (lower.startsWith('b') || lower.startsWith('beta') || lower.includes('beta'))
    type = 'beta'
  else if (lower.startsWith('a') || lower.startsWith('alpha') || lower.includes('alpha'))
    type = 'alpha'
  if (lower.includes('forge')) type = 'forge'
  else if (lower.includes('fabric')) type = 'fabric'
  else if (lower.includes('quilt')) type = 'quilt'
  else if (lower.includes('neoforge')) type = 'neoforge'
  else if (lower.includes('optifine')) type = 'optifine'
  else if (lower.includes('legacyfabric')) type = 'legacyfabric'

  let major = v
  const parts = v.split('.')
  if (parts.length >= 2) {
    if (parts[0] === '1' && parts[1]) major = `1.${parts[1]}`
    else major = `${parts[0]}.${parts[1] || '0'}`
  }
  return { full: v, major, type }
}

function typeLabel(type: VersionType): string {
  return t('modals.launch.badges.' + type) || type
}

const groupedFiltered = computed(() => {
  let list = versions.value.map(getVersionInfo)

  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(v => v.full.toLowerCase().includes(q) || v.major.includes(q))
  }

  if (activeTab.value !== 'all') {
    list = list.filter(v => v.type === activeTab.value)
  }

  const groups = new Map<string, VersionInfo[]>()
  for (const v of list) {
    if (!groups.has(v.major)) groups.set(v.major, [])
    groups.get(v.major)!.push(v)
  }

  const sorted = [...groups.entries()].sort((a, b) => {
    const ap = a[0].split('.').map(Number)
    const bp = b[0].split('.').map(Number)
    for (let i = 0; i < Math.max(ap.length, bp.length); i++) {
      const diff = (bp[i] ?? 0) - (ap[i] ?? 0)
      if (diff !== 0) return diff
    }
    return 0
  })

  return sorted.map(([major, items]) => ({
    major,
    items: items.sort((a, b) => {
      const ap = a.full.split(/[.\-]/).map(Number)
      const bp = b.full.split(/[.\-]/).map(Number)
      for (let i = 0; i < Math.max(ap.length, bp.length); i++) {
        const diff = (bp[i] ?? 0) - (ap[i] ?? 0)
        if (diff !== 0) return diff
      }
      return 0
    }),
  }))
})

function toggleGroup(major: string) {
  const s = new Set(expandedGroups.value)
  if (s.has(major)) s.delete(major)
  else s.add(major)
  expandedGroups.value = s
}

function select(version: string) {
  selectedVersion.value = version
}

function play() {
  if (!selectedVersion.value) return
  emit('play', { version: selectedVersion.value })
}

onMounted(async () => {
  try {
    const result = await window.NovaCoreManager.GetDownloadedVersions()
    versions.value = Array.isArray(result) ? result : []
    const s = new Set<string>()
    for (const v of versions.value) {
      const info = getVersionInfo(v)
      s.add(info.major)
    }
    expandedGroups.value = s
  } catch {
    versions.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped lang="scss" src="../../Styles/Modals/LaunchModal.scss"></style>
