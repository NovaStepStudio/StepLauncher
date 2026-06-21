<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { CSSProperties, PropType } from 'vue'

type StaticLightOptions = {
  blur?: number
  canvasBlur?: number
  brightness?: number
  saturation?: number
  opacity?: number
  scale?: number
  zIndex?: number
  fadeInDuration?: number
}

const props = defineProps({
  containerRef: {
    type: Object as PropType<HTMLElement | null>,
    default: null,
  },
  source: {
    type: [String, Object] as PropType<string | HTMLImageElement | null | undefined>,
    default: null,
  },
  options: {
    type: Object as PropType<StaticLightOptions>,
    default: () => ({}),
  },
  className: {
    type: String,
    default: '',
  },
  style: {
    type: Object as PropType<CSSProperties>,
    default: () => ({}),
  },
})

const defaultOptions: Required<StaticLightOptions> = {
  blur: 80,
  canvasBlur: 15,
  brightness: 1.2,
  saturation: 1.3,
  opacity: 0.6,
  scale: 0.15,
  zIndex: -1,
  fadeInDuration: 800,
}

const config = computed(() => ({
  ...defaultOptions,
  ...props.options,
}))

const rootRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const ctxRef = ref<CanvasRenderingContext2D | null>(null)
const imageRef = ref<HTMLImageElement | null>(null)
const resizeObserverRef = ref<ResizeObserver | null>(null)
const rafRef = ref<number | null>(null)
const fadeTimerRef = ref<number | null>(null)
const classWatcherRef = ref<MutationObserver | null>(null)

let token = 0

function get2DContext(canvas: HTMLCanvasElement) {
  return canvas.getContext('2d', {
    alpha: false,
    desynchronized: true,
  })
}

function loadImage(source: string | HTMLImageElement): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    if (typeof source !== 'string') {
      if (source.complete && source.naturalWidth > 0) {
        resolve(source)
        return
      }
      const onLoad = () => { cleanup(); resolve(source) }
      const onError = () => { cleanup(); reject(new Error('[StaticLight] Image element failed to load')) }
      const cleanup = () => {
        source.removeEventListener('load', onLoad)
        source.removeEventListener('error', onError)
      }
      source.addEventListener('load', onLoad, { once: true })
      source.addEventListener('error', onError, { once: true })
      return
    }

    const img = new Image()
    img.crossOrigin = 'anonymous'
    const onLoad = () => {
      img.removeEventListener('load', onLoad)
      img.removeEventListener('error', onError)
      resolve(img)
    }
    const onError = () => {
      img.removeEventListener('load', onLoad)
      img.removeEventListener('error', onError)
      reject(new Error(`[StaticLight] Failed to load image: ${source}`))
    }
    img.addEventListener('load', onLoad, { once: true })
    img.addEventListener('error', onError, { once: true })
    img.src = source
  })
}

function lightEnabled(): boolean {
  return !document.documentElement.classList.contains('filters-disabled') && !document.documentElement.classList.contains('shadows-disabled')
}

function render() {
  const currentCanvas = canvasRef.value
  const currentCtx = ctxRef.value
  const currentImage = imageRef.value

  if (!currentCanvas || !currentCtx || !currentImage || !currentImage.naturalWidth || !currentImage.naturalHeight || !lightEnabled()) {
    return
  }

  const sw = currentImage.naturalWidth
  const sh = currentImage.naturalHeight
  const w = Math.max(1, Math.floor(sw * config.value.scale))
  const h = Math.max(1, Math.floor(sh * config.value.scale))

  if (currentCanvas.width !== w) currentCanvas.width = w
  if (currentCanvas.height !== h) currentCanvas.height = h

  currentCtx.clearRect(0, 0, w, h)
  currentCtx.filter = `blur(${config.value.canvasBlur}px) brightness(${config.value.brightness}) saturate(${config.value.saturation})`

  try {
    currentCtx.drawImage(currentImage, 0, 0, w, h)
  } catch {}
}

async function applyImage() {
  if (!props.source || !lightEnabled()) return
  const currentToken = ++token
  try {
    const img = await loadImage(props.source)
    if (currentToken !== token) return
    imageRef.value = img

    const canvas = canvasRef.value
    if (!canvas) return

    if (fadeTimerRef.value !== null) {
      window.clearTimeout(fadeTimerRef.value)
      fadeTimerRef.value = null
    }

    if (canvas.style.opacity !== '0') {
      canvas.style.opacity = '0'
      fadeTimerRef.value = window.setTimeout(() => {
        if (currentToken !== token) return
        render()
        requestAnimationFrame(() => {
          const latestCanvas = canvasRef.value
          if (latestCanvas) latestCanvas.style.opacity = String(config.value.opacity)
        })
      }, config.value.fadeInDuration)
      return
    }

    render()
    requestAnimationFrame(() => {
      const latestCanvas = canvasRef.value
      if (latestCanvas) latestCanvas.style.opacity = String(config.value.opacity)
    })
  } catch {}
}

function cleanup() {
  resizeObserverRef.value?.disconnect()
  resizeObserverRef.value = null

  if (rafRef.value !== null) {
    cancelAnimationFrame(rafRef.value)
    rafRef.value = null
  }

  if (fadeTimerRef.value !== null) {
    window.clearTimeout(fadeTimerRef.value)
    fadeTimerRef.value = null
  }

  token += 1
  ctxRef.value = null
  imageRef.value = null
}

function setup() {
  if (!lightEnabled()) return

  const container = props.containerRef
  const root = rootRef.value
  const canvas = canvasRef.value

  if (!container || !root || !canvas) return

  if (getComputedStyle(container).position === 'static') {
    container.style.position = 'relative'
  }
  container.style.isolation = 'isolate'

  const ctx = get2DContext(canvas)
  if (!ctx) return
  ctxRef.value = ctx

  root.style.cssText = `
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: ${config.value.zIndex};
  `

  canvas.style.cssText = `
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    filter: blur(${config.value.blur}px);
    opacity: 0;
    transition: opacity ${config.value.fadeInDuration}ms ease;
    will-change: filter, opacity;
    display: block;
  `

  resizeObserverRef.value?.disconnect()
  resizeObserverRef.value = new ResizeObserver(() => { render() })
  resizeObserverRef.value.observe(container)
  window.addEventListener('resize', render)

  render()
  applyImage()
}

onMounted(async () => {
  await nextTick()
  setup()

  classWatcherRef.value = new MutationObserver(() => {
    if (lightEnabled()) {
      if (!ctxRef.value) setup()
    } else {
      const cvs = canvasRef.value
      if (cvs) cvs.style.opacity = '0'
      cleanup()
    }
  })
  classWatcherRef.value.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})

onBeforeUnmount(() => {
  classWatcherRef.value?.disconnect()
  cleanup()
})

watch(
  () => [
    props.source,
    config.value.blur,
    config.value.canvasBlur,
    config.value.brightness,
    config.value.saturation,
    config.value.opacity,
    config.value.scale,
    config.value.zIndex,
    config.value.fadeInDuration,
  ],
  () => {
    applyImage()
    render()
  },
)

watch(() => props.containerRef, (newVal) => {
  if (newVal) {
    resizeObserverRef.value?.disconnect()
    setup()
  }
})
</script>

<template>
  <div
    ref="rootRef"
    :class="className"
    :style="style"
  >
    <canvas ref="canvasRef" />
    <slot />
  </div>
</template>
