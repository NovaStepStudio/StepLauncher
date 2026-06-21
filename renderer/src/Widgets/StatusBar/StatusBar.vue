<template>
  <div class="status-bar" v-if="show" @mouseenter="showPanel = true" @mouseleave="showPanel = false">
    <div class="status-item" :class="apiStatus">
      <span class="dot" />
      <span class="label">{{ $t('widgets.status_bar.labels.api') }}</span>
    </div>
    <div class="status-item" :class="novaStatus">
      <span class="dot" />
      <span class="label">{{ $t('widgets.status_bar.labels.novacore') }}</span>
    </div>

    <Transition name="fade">
      <div v-if="showPanel" class="status-popup">
        <div class="popup-header">{{ $t('widgets.status_bar.popup_title') }}</div>

        <div class="popup-graph">
          <div class="graph-labels">
            <div class="graph-label">
              <span class="gl-dot" :class="apiStatus" />
              <span class="gl-text">{{ $t('widgets.status_bar.labels.api') }}</span>
              <span class="gl-badge" :class="apiStatus">{{ statusLabel(apiStatus) }}</span>
            </div>
            <div class="graph-label">
              <span class="gl-dot" :class="novaStatus" />
              <span class="gl-text">{{ $t('widgets.status_bar.labels.novacore') }}</span>
              <span class="gl-badge" :class="novaStatus">{{ statusLabel(novaStatus) }}</span>
            </div>
          </div>
          <div class="graph-chart">
            <div class="chart-legend">
              <span class="legend-item"><span class="legend-dot online" /> {{ $t('widgets.status_bar.legend.online') }}</span>
              <span class="legend-item"><span class="legend-dot checking" /> {{ $t('widgets.status_bar.legend.checking') }}</span>
              <span class="legend-item"><span class="legend-dot offline" /> {{ $t('widgets.status_bar.legend.offline') }}</span>
            </div>
          <div class="chart-bars">
            <div
              v-for="(h, i) in apiHistory"
              :key="i"
              :class="['chart-bar', h.status]"
              :style="{ height: barHeight(h.status, h.latency) }"
              :title="barTooltip(i, h)"
            />
          </div>
          <div class="chart-time">
            <span class="ct-now">{{ $t('widgets.status_bar.chart.now') }}</span>
            <span class="ct-ago">{{ $t('widgets.status_bar.chart.ago', { seconds: apiHistory.length * 3 }) }}</span>
          </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { t } from '../../i18n'

const show = ref(false)
const apiStatus = ref<'online' | 'offline' | 'checking'>('checking')
const novaStatus = ref<'online' | 'offline' | 'starting' | 'checking'>('checking')
const showPanel = ref(false)

const apiHistory = ref<{ status: string; latency: number; ts: number }[]>([])

let pollTimer: ReturnType<typeof setInterval> | null = null
let unsubNova: (() => void) | null = null

function statusLabel(s: string): string {
  return t('widgets.status_bar.status.' + s)
}

function barHeight(s: string, latency: number): string {
  if (s === 'offline') return '6px'
  if (s === 'checking') return '12px'
  const h = Math.max(8, Math.min(24, 24 - latency / 50))
  return h + 'px'
}

function barTooltip(i: number, h: { status: string; latency: number; ts: number }): string {
  const secs = Math.round((Date.now() - h.ts) / 1000)
  const when = secs === 0 ? t('widgets.status_bar.tooltips.now') : t('widgets.status_bar.tooltips.ago', { seconds: secs })
  if (h.status === 'offline') return t('widgets.status_bar.tooltips.offline', { when })
  if (h.status === 'checking') return t('widgets.status_bar.tooltips.checking', { when })
  return t('widgets.status_bar.tooltips.latency', { when, latency: h.latency })
}

async function checkAll() {
  apiStatus.value = 'checking'
  if (novaStatus.value !== 'starting') novaStatus.value = 'checking'

  let apiLatency = 0
  try {
    const t0 = performance.now()
    await window.AuthManager.GetHealth()
    apiLatency = Math.round(performance.now() - t0)
    apiStatus.value = 'online'
  } catch {
    apiStatus.value = 'offline'
    apiLatency = 0
  }

  try {
    const s = await window.NovaCoreManager.Status()
    if (s.status === 'running') novaStatus.value = 'online'
    else if (s.status === 'starting') novaStatus.value = 'starting'
    else novaStatus.value = 'offline'
  } catch {
    if (novaStatus.value !== 'starting') novaStatus.value = 'offline'
  }

  apiHistory.value.push({ status: apiStatus.value, latency: apiLatency, ts: Date.now() })
  if (apiHistory.value.length > 40) apiHistory.value.shift()
}

onMounted(() => {
  if (window.__splashDone) {
    show.value = true
  } else {
    const checkSplash = setInterval(() => {
      if (window.__splashDone) {
        show.value = true
        clearInterval(checkSplash)
      }
    }, 100)
    setTimeout(() => { clearInterval(checkSplash); show.value = true }, 15000)
  }

  checkAll()
  pollTimer = setInterval(checkAll, 3000)

  try {
    if (window.NovaCoreManager?.OnEvent) {
      unsubNova = window.NovaCoreManager.OnEvent((ev) => {
        if (ev.event === 'engine_unreachable') {
          novaStatus.value = 'offline'
        }
        checkAll()
      })
    }
  } catch {}
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (unsubNova) unsubNova()
})
</script>

<style scoped lang="scss" src="../../Styles/Widgets/StatusBar.scss"></style>
