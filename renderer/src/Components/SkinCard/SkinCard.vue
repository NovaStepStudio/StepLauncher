<template>
  <div class="skin-card" :class="{ 'skin-card--active': active }">
    <div class="skin-card__viewer">
      <canvas ref="canvas" class="skin-card__canvas"></canvas>
      <div v-if="active" class="skin-card__badge">{{ $t('skin_card.badge_active') }}</div>
    </div>
    <span class="skin-card__name">{{ name }}</span>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import type { SkinViewer } from 'skinview3d'
import { useUserCache } from '../../Composables/useUserCache'

const props = withDefaults(defineProps<{
  name: string
  skinUrl: string
  capeUrl?: string
  active?: boolean
  width?: number
  height?: number
  fallbackSkin?: string
}>(), {
  fallbackSkin: 'assets/defaults/steve_skin.png',
})

const cache = useUserCache()
const canvas = ref<HTMLCanvasElement | null>(null)
let viewer: SkinViewer | null = null

async function init() {
  if (!canvas.value) return
  viewer?.dispose()
  viewer = null
  try {
    let skinImg = await cache.preloadImage(props.skinUrl)
    if (!skinImg) {
      skinImg = await cache.preloadImage(props.fallbackSkin)
    }
    if (!canvas.value || !skinImg) return
    let capeImg: HTMLImageElement | null | undefined
    if (props.capeUrl) {
      capeImg = await cache.preloadImage(props.capeUrl)
    }
    if (!canvas.value) return
    const { SkinViewer } = await import('skinview3d')
    viewer = new SkinViewer({
      canvas: canvas.value,
      width: props.width ?? 110,
      height: props.height ?? 160,
      skin: skinImg,
      ...(capeImg ? { cape: capeImg } : {}),
    })
    viewer.autoRotate = true
    viewer.autoRotateSpeed = 1.2
    viewer.zoom = 0.9
  } catch {}
}

onMounted(init)
onUnmounted(() => viewer?.dispose())
watch(() => [props.skinUrl, props.capeUrl], init)
</script>

<style scoped lang="scss" src="../../Styles/Components/SkinCard.scss"></style>
