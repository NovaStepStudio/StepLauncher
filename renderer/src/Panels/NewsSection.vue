<template>
  <div class="news-wrapper">
    <div class="news-sidebar">
      <button class="back-btn" @click="selectedEntry ? (selectedEntry = null) : emit('close')">
        <img :src="'assets/svg/arrow-left.svg'" width="14" height="14" />
      </button>
      <span class="news-title">{{ $t('news.title') }}</span>
      <div class="sidebar-filters">
        <button :class="['filter-btn', { active: filter === 'all' }]" @click="filter = 'all'">{{ $t('news.filters.all') }}</button>
        <button :class="['filter-btn', { active: filter === 'release' }]" @click="filter = 'release'">{{ $t('news.filters.releases') }}</button>
        <button :class="['filter-btn', { active: filter === 'snapshot' }]" @click="filter = 'snapshot'">{{ $t('news.filters.snapshots') }}</button>
      </div>
    </div>

    <div class="news-main" v-if="!selectedEntry">
      <div class="loading" v-if="loading">
        <div class="spinner" />
        <span>{{ $t('news.loading') }}</span>
      </div>
      <div class="error" v-if="error">
        <span>{{ error }}</span>
        <button class="retry-btn" @click="loadNews">{{ $t('news.buttons.retry') }}</button>
      </div>
      <div v-if="!loading && !error" class="news-list">
        <div v-for="entry in filteredEntries" :key="entry.id" class="card" @click="selectedEntry = entry">
          <div class="card-img" v-if="entry.image" :style="{ backgroundImage: `url(${getEntryImg(entry.image.url)})` }" />
          <div class="card-body">
            <div class="card-meta">
              <span :class="['card-badge', entry.type]">{{ typeLabel(entry.type) }}</span>
              <span class="card-version">{{ entry.version }}</span>
            </div>
            <h3 class="card-title">{{ entry.title }}</h3>
          </div>
        </div>
        <div v-if="filteredEntries.length === 0" class="no-matches">{{ $t('news.no_matches') }}</div>
      </div>
    </div>

    <div class="detail-view" v-if="selectedEntry">
      <div class="detail-header">
        <span :class="['detail-badge', selectedEntry.type]">{{ typeLabel(selectedEntry.type) }}</span>
        <span class="detail-version">{{ selectedEntry.version }}</span>
      </div>
      <h2 class="detail-title">{{ selectedEntry.title }}</h2>
      <div class="detail-body" v-html="`<base target='_blank'>${selectedEntry.body}`" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { t } from '../i18n'
import { useUserCache } from '../Composables/useUserCache'

const NEWS_BASE = 'https://launchercontent.mojang.com'

interface NewsEntry {
  title: string
  type: string
  version: string
  image?: { url: string; title: string }
  body: string
  id: string
}

const cache = useUserCache()
const entries = ref<NewsEntry[]>([])
const loading = ref(true)
const error = ref('')
const filter = ref<'all' | 'release' | 'snapshot'>('all')
const selectedEntry = ref<NewsEntry | null>(null)

const filteredEntries = computed(() => {
  if (filter.value === 'all') return entries.value
  return entries.value.filter(e => e.type === filter.value)
})

function getEntryImg(path: string): string {
  return cache.getCachedSrc(NEWS_BASE + path)
}

function typeLabel(type: string): string {
  if (type === 'release') return 'Release'
  if (type === 'snapshot') return 'Snapshot'
  return type
}

const emit = defineEmits<{ close: [] }>()

async function loadNews() {
  loading.value = true
  error.value = ''
  try {
    const data = await window.ElectronAPI.FetchJson(NEWS_BASE + '/javaPatchNotes.json')
    entries.value = data.entries ?? []
    const urls = (data.entries ?? [])
      .map((e: NewsEntry) => e.image ? NEWS_BASE + e.image.url : '')
      .filter(Boolean)
    cache.prefetchImages(urls)
  } catch (e: any) {
    error.value = e?.message ?? t('news.error')
  } finally {
    loading.value = false
  }
}

onMounted(loadNews)
</script>

<style scoped lang="scss" src="../Styles/Panels/NewsSection.scss"></style>
