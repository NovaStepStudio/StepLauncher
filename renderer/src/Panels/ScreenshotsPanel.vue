<template>
  <section class="ss-panel">
    <div class="ss-sidebar">
      <button class="btn-icon" @click="$emit('close')">
        <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
      </button>
      <h2 class="ss-title">{{ $t('screenshots.title') }}</h2>

      <div v-if="screenshots.length" class="ss-grid">
        <div
          v-for="(s, i) in screenshots" :key="s.name"
          class="ss-thumb"
          :class="{ active: selectedIdx === i }"
          @click="selectedIdx = i"
        >
          <img :src="fileUrl(s.path)" class="ss-thumb-img" />
        </div>
      </div>
      <div v-else class="ss-empty">{{ $t('screenshots.empty') }}</div>
    </div>

    <div v-if="selected" class="ss-viewer">
      <img :src="fileUrl(selected.path)" class="ss-viewer-img" />
      <span class="ss-viewer-name">{{ selected.name }}</span>
    </div>
    <div v-else class="ss-viewer ss-viewer--empty">
      <img :src="'assets/svg/image.svg'" width="48" height="48" style="opacity:0.3" />
      <span>{{ $t('screenshots.empty_viewer') }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

defineEmits<{ (e: 'close'): void }>()

const screenshots = ref<any[]>([])
const selectedIdx = ref(-1)

const selected = computed(() =>
  selectedIdx.value >= 0 ? screenshots.value[selectedIdx.value] : null
)

function fileUrl(p: string) {
  if (!p) return ''
  if (p.startsWith('http')) return p
  if (p.startsWith('file://')) return p
  p = p.replace(/\\/g, '/')
  return `file:///${encodeURI(p)}`
}

onMounted(async () => {
  try {
    const info = await window.NovaCoreManager.GetInfo()
    const ssDir = `${info.baseDir}/game/screenshots`
    const entries = await window.ElectronAPI.ReadDir(ssDir)
    if (!entries) { screenshots.value = []; return }
    screenshots.value = entries
      .filter(e => e.isFile && /\.(png|jpg|jpeg|webp)$/i.test(e.name))
      .sort((a, b) => b.mtime.localeCompare(a.mtime))
      .map(e => ({ name: e.name, path: `${ssDir}/${e.name}` }))
  } catch { screenshots.value = [] }
})
</script>

<style scoped lang="scss" src="../Styles/Panels/ScreenshotsPanel.scss"></style>
