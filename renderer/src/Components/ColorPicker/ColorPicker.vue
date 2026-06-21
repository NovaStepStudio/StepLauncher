<template>
  <div class="color-picker" ref="root" @mouseenter="expanded = true" @mouseleave="expanded = false">
    <div class="color-trigger" :style="{ background: hex8 }">
      <img class="chevron" :class="{ open: expanded }" :src="'assets/svg/chevron-down.svg'" width="9" height="9" />
    </div>
    <div class="color-body" :class="{ expanded }">
      <div class="color-body-inner">
        <div class="sv-picker" ref="svArea" @mousedown="onSvDown" :style="{ background: `hsl(${hue}, 100%, 50%)` }">
          <div class="sv-white" />
          <div class="sv-black" />
          <div class="sv-cursor" :style="{ top: svY + '%', left: svX + '%' }" />
        </div>
        <div class="sliders">
          <div class="hue-track" ref="hueTrack" @mousedown="onHueDown">
            <div class="hue-cursor" :style="{ left: huePct + '%' }" />
          </div>
          <div class="alpha-track" ref="alphaTrack" @mousedown="onAlphaDown">
            <div class="alpha-fill" :style="{ background: `linear-gradient(90deg, transparent, hsl(${hue}, 100%, 50%))` }" />
            <div class="alpha-cursor" :style="{ left: alphaPct + '%' }" />
          </div>
        </div>
        <div class="controls">
          <div class="preview" :style="{ background: hex8 }" />
          <div class="hex-wrap">
            <span class="hash">#</span>
            <input class="hex-input" :value="cleanHex" @input="onHexInput" @blur="onHexBlur" maxlength="8" spellcheck="false" placeholder="000000" />
          </div>
        </div>
        <div class="swatches">
          <button v-for="c in swatches" :key="c" :class="['swatch', { active: hex8.toLowerCase() === c.toLowerCase() }]" :style="{ background: c }" @click="selectSwatch(c)" :title="c" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const expanded = ref(false)
const root = ref<HTMLElement>()
const svArea = ref<HTMLElement>()
const hueTrack = ref<HTMLElement>()
const alphaTrack = ref<HTMLElement>()

const swatches = [
  '#111111', '#1a1a2e', '#16213e', '#0f3460',
  '#e94560', '#533483', '#0d7377', '#14a76c',
  '#ff652f', '#ffe400', '#00c9b7', '#5cd0e7',
  '#ffffff', '#888888', '#555555', '#222222',
]

function parseHex(hex: string) {
  const h = hex.replace('#', '')
  let r = 0, g = 0, b = 0, a = 1
  if (h.length === 3) {
    r = parseInt(h[0]! + h[0]!, 16)
    g = parseInt(h[1]! + h[1]!, 16)
    b = parseInt(h[2]! + h[2]!, 16)
  } else if (h.length >= 6) {
    r = parseInt(h[0]! + h[1]!, 16)
    g = parseInt(h[2]! + h[3]!, 16)
    b = parseInt(h[4]! + h[5]!, 16)
    if (h.length === 8) a = parseInt(h[6]! + h[7]!, 16) / 255
  }
  return { r, g, b, a }
}

function toHex(r: number, g: number, b: number, a: number) {
  const rh = Math.round(r).toString(16).padStart(2, '0')
  const gh = Math.round(g).toString(16).padStart(2, '0')
  const bh = Math.round(b).toString(16).padStart(2, '0')
  const ah = Math.round(a * 255).toString(16).padStart(2, '0')
  return `#${rh}${gh}${bh}${ah}`
}

function rgbToHsv(r: number, g: number, b: number) {
  const rr = r / 255, gg = g / 255, bb = b / 255
  const mx = Math.max(rr, gg, bb), mn = Math.min(rr, gg, bb)
  const d = mx - mn
  let h = 0
  if (d !== 0) {
    if (mx === rr) h = ((gg - bb) / d + (gg < bb ? 6 : 0)) * 60
    else if (mx === gg) h = ((bb - rr) / d + 2) * 60
    else h = ((rr - gg) / d + 4) * 60
  }
  const s = mx === 0 ? 0 : d / mx
  return { h, s, v: mx }
}

function hsvToRgb(h: number, s: number, v: number) {
  const c = v * s
  const hp = h / 60
  const x = c * (1 - Math.abs(hp % 2 - 1))
  let r = 0, g = 0, b = 0
  if (hp < 1) { r = c; g = x }
  else if (hp < 2) { r = x; g = c }
  else if (hp < 3) { g = c; b = x }
  else if (hp < 4) { g = x; b = c }
  else if (hp < 5) { r = x; b = c }
  else { r = c; b = x }
  const m = v - c
  return { r: (r + m) * 255, g: (g + m) * 255, b: (b + m) * 255 }
}

const hue = ref(180)
const sat = ref(0.5)
const val = ref(0.8)
const alpha = ref(1)

function fromHex(hex: string) {
  const { r, g, b, a } = parseHex(hex)
  alpha.value = a
  const hsv = rgbToHsv(r, g, b)
  hue.value = hsv.h
  sat.value = hsv.s
  val.value = hsv.v
}

function emitColor() {
  const { r, g, b } = hsvToRgb(hue.value, sat.value, val.value)
  emit('update:modelValue', toHex(r, g, b, alpha.value))
}

fromHex(props.modelValue)

const hex8 = computed(() => {
  const { r, g, b } = hsvToRgb(hue.value, sat.value, val.value)
  return toHex(r, g, b, alpha.value)
})

const cleanHex = computed(() => {
  const { r, g, b } = hsvToRgb(hue.value, sat.value, val.value)
  const rh = Math.round(r).toString(16).padStart(2, '0')
  const gh = Math.round(g).toString(16).padStart(2, '0')
  const bh = Math.round(b).toString(16).padStart(2, '0')
  return `${rh}${gh}${bh}`
})

const huePct = computed(() => (hue.value / 360) * 100)
const svX = computed(() => sat.value * 100)
const svY = computed(() => (1 - val.value) * 100)
const alphaPct = computed(() => alpha.value * 100)

function selectSwatch(c: string) {
  fromHex(c)
  emitColor()
}

function onSvDown(e: MouseEvent) {
  e.preventDefault()
  updateSv(e)
  const handler = (ev: MouseEvent) => { ev.preventDefault(); updateSv(ev) }
  const up = () => { window.removeEventListener('mousemove', handler); window.removeEventListener('mouseup', up) }
  window.addEventListener('mousemove', handler)
  window.addEventListener('mouseup', up)
}

function updateSv(e: MouseEvent) {
  const el = svArea.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  sat.value = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
  val.value = Math.max(0, Math.min(1, 1 - (e.clientY - rect.top) / rect.height))
  emitColor()
}

function onHueDown(e: MouseEvent) {
  e.preventDefault()
  updateHue(e)
  const handler = (ev: MouseEvent) => { ev.preventDefault(); updateHue(ev) }
  const up = () => { window.removeEventListener('mousemove', handler); window.removeEventListener('mouseup', up) }
  window.addEventListener('mousemove', handler)
  window.addEventListener('mouseup', up)
}

function updateHue(e: MouseEvent) {
  const el = hueTrack.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  hue.value = Math.max(0, Math.min(360, ((e.clientX - rect.left) / rect.width) * 360))
  emitColor()
}

function onAlphaDown(e: MouseEvent) {
  e.preventDefault()
  updateAlpha(e)
  const handler = (ev: MouseEvent) => { ev.preventDefault(); updateAlpha(ev) }
  const up = () => { window.removeEventListener('mousemove', handler); window.removeEventListener('mouseup', up) }
  window.addEventListener('mousemove', handler)
  window.addEventListener('mouseup', up)
}

function updateAlpha(e: MouseEvent) {
  const el = alphaTrack.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  alpha.value = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
  emitColor()
}

function onHexInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value.replace(/[^0-9a-fA-F]/g, '').slice(0, 8)
  if (raw.length === 6) {
    const r = parseInt(raw[0]! + raw[1]!, 16)
    const g = parseInt(raw[2]! + raw[3]!, 16)
    const b = parseInt(raw[4]! + raw[5]!, 16)
    fromHex(toHex(r, g, b, alpha.value))
    emitColor()
  } else if (raw.length === 8) {
    fromHex('#' + raw)
    emitColor()
  }
}

function onHexBlur(e: Event) {
  const el = e.target as HTMLInputElement
  if (el.value.replace(/[^0-9a-fA-F]/g, '').length < 6) {
    el.value = cleanHex.value
  }
}
</script>

<style scoped lang="scss" src="../../Styles/Components/ColorPicker.scss"></style>
