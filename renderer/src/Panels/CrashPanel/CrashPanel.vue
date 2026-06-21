<template>
  <div v-if="visible" class="crash-overlay" @click.self="$emit('close')">
    <div class="crash-modal">
      <div class="crash-topbar">
        <div class="topbar-left">
          <div class="crash-badge">
            <img :src="'assets/svg/alert-triangle.svg'" width="20" height="20" />
          </div>
          <div class="topbar-info">
            <span class="topbar-title">{{ $t('modals.crash.title') }}</span>
            <span class="topbar-time">{{ formatTime(info?.timestamp) }}</span>
          </div>
        </div>
        <button class="topbar-close" @click="$emit('close')">
          <img :src="'assets/svg/x.svg'" width="16" height="16" />
        </button>
      </div>

      <div class="crash-content">
        <div class="crash-hero">
          <div class="hero-icon">
            <img :src="'assets/svg/alert-circle.svg'" width="48" height="48" />
          </div>
          <div class="hero-text">
            <span class="hero-reason">{{ info?.reason }}</span>
            <span class="hero-code">{{ $t('modals.crash.exit_code', { exitCode: info?.exitCode || '?' }) }}</span>
          </div>
        </div>

        <div class="crash-actions">
          <button class="act-btn act-btn--danger" @click="$emit('close')">
            <img :src="'assets/svg/x.svg'" width="14" height="14" />
            {{ $t('modals.crash.buttons.close') }}
          </button>
          <button class="act-btn" @click="copyCrash">
            <img :src="'assets/svg/copy.svg'" width="14" height="14" />
            {{ $t('modals.crash.buttons.copy') }}
          </button>
          <button class="act-btn" @click="$emit('open-console')">
            <img :src="'assets/svg/log-out.svg'" width="14" height="14" />
            {{ $t('modals.crash.buttons.console') }}
          </button>
          <button class="act-btn" @click="showLogs = !showLogs">
            <img :src="'assets/svg/file-text.svg'" width="14" height="14" />
            {{ $t(showLogs ? 'modals.crash.buttons.hide_logs' : 'modals.crash.buttons.show_logs') }}
          </button>

        </div>

        <div v-if="showLogs && info?.context?.length" class="crash-logs">
          <div class="logs-bar">
            <span class="logs-bar-title">{{ $t('modals.crash.logs.bar_title') }}</span>
            <span class="logs-bar-count">{{ $t('modals.crash.logs.bar_count', { count: info!.context.length }) }}</span>
          </div>
          <div class="logs-body">
            <div v-for="(line, i) in info!.context" :key="i" class="log-line">
              <span class="log-num">{{ i + 1 }}</span>
              <span class="log-text">{{ line }}</span>
            </div>
          </div>
        </div>

        <div v-if="!showLogs && info?.context?.length" class="logs-preview" @click="showLogs = true">
          <img :src="'assets/svg/file.svg'" width="12" height="12" />
          {{ $t('modals.crash.logs.preview', { count: info!.context.length }) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { t } from '../../i18n'

export interface CrashInfo {
  reason: string
  context: string[]
  exitCode: number
  timestamp?: number
}

const props = defineProps<{
  info: CrashInfo | null
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
  'open-console': []
}>()

const showLogs = ref(false)

function formatTime(ts?: number) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function copyCrash() {
  if (!props.info) return
  const text = t('modals.crash.copy_format', { reason: props.info.reason, exitCode: props.info.exitCode, context: (props.info.context || []).join('\n') })
  navigator.clipboard.writeText(text)
}
</script>

<style scoped>
.crash-overlay {
  position: fixed;
  inset: 1.75rem 0 0 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: cfade 0.2s ease;
}

@keyframes cfade { from { opacity: 0; } to { opacity: 1; } }

.crash-modal {
  width: 520px;
  max-height: 80vh;
  background: var(--modal-bg, rgba(17, 17, 24, 0.97));
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--modal-border, rgba(255,255,255,0.06));
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 48px rgba(0, 0, 0, 0.6);
}

.crash-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--modal-border, rgba(255,255,255,0.06));
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.crash-badge {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 85%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-color, #5cd0e7);
  flex-shrink: 0;
}

.topbar-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.topbar-title {
  font-family: var(--font-family-primary, 'Lexend'), sans-serif;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-primary, #fff);
}

.topbar-time {
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.62rem;
  color: var(--text-muted, rgba(255,255,255,0.35));
}

.topbar-close {
  all: unset;
  color: var(--text-muted, rgba(255,255,255,0.35));
  padding: 4px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.12s;
  display: flex;
  &:hover {
    color: var(--text-primary, #fff);
    background: var(--modal-hover, rgba(255,255,255,0.04));
  }
}

.crash-content {
  padding: 16px;
  overflow-y: auto;
  flex: 1;
}

.crash-hero {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px;
  background: color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 92%);
  border: 1px solid color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 82%);
  border-radius: 10px;
  margin-bottom: 12px;
}

.hero-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 88%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.hero-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.hero-reason {
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-primary, #fff);
  word-break: break-word;
  line-height: 1.4;
}

.hero-code {
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.65rem;
  color: var(--text-muted, rgba(255,255,255,0.35));
}

.crash-actions {
  display: flex;
  gap: 5px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.act-btn {
  all: unset;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 10px;
  border-radius: 6px;
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.67rem;
  font-weight: 500;
  color: var(--text-secondary, rgba(255,255,255,0.6));
  background: var(--modal-hover, rgba(255,255,255,0.04));
  cursor: pointer;
  transition: all 0.12s;
  white-space: nowrap;
  &:hover {
    background: color-mix(in srgb, var(--text-primary, #fff), transparent 92%);
    color: var(--text-primary, #fff);
  }
}

.act-btn--danger {
  color: var(--modal-danger, #ff7675);
  background: color-mix(in srgb, var(--modal-danger, #ff7675), transparent 90%);
  &:hover {
    background: color-mix(in srgb, var(--modal-danger, #ff7675), transparent 80%);
  }
}

.crash-logs {
  border: 1px solid var(--modal-border, rgba(255,255,255,0.06));
  border-radius: 8px;
  overflow: hidden;
}

.logs-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 7px 10px;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid var(--modal-border, rgba(255,255,255,0.06));
}

.logs-bar-title {
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.65rem;
  font-weight: 600;
  color: var(--text-muted, rgba(255,255,255,0.35));
}

.logs-bar-count {
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.6rem;
  color: var(--text-faint, rgba(255,255,255,0.15));
}

.logs-body {
  max-height: 35vh;
  overflow-y: auto;
  padding: 0;
  background: rgba(0, 0, 0, 0.3);
}

.log-line {
  display: flex;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.6rem;
  line-height: 1.7;
  border-bottom: 1px solid rgba(255, 255, 255, 0.02);
}

.log-line:last-child { border-bottom: none; }

.log-num {
  width: 32px;
  flex-shrink: 0;
  text-align: right;
  padding: 0 8px 0 10px;
  color: rgba(255, 255, 255, 0.12);
  user-select: none;
}

.log-text {
  flex: 1;
  padding: 0 10px 0 0;
  color: rgba(255, 255, 255, 0.55);
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-preview {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border: 1px dashed var(--modal-border, rgba(255,255,255,0.06));
  border-radius: 6px;
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.65rem;
  color: var(--text-muted, rgba(255,255,255,0.35));
  cursor: pointer;
  transition: all 0.12s;
  &:hover {
    border-color: var(--accent-color, #5cd0e7);
    color: var(--text-secondary, rgba(255,255,255,0.6));
  }
}

.act-btn--upload {
  color: var(--accent-color, #5cd0e7);
  background: color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 88%);
  &:hover {
    background: color-mix(in srgb, var(--accent-color, #5cd0e7), transparent 76%);
  }
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.upload-result {
  padding: 7px 10px;
  margin-bottom: 12px;
  border-radius: 6px;
  font-family: var(--font-family-secundary, 'Inter'), sans-serif;
  font-size: 0.65rem;
  font-weight: 500;
  background: rgba(255, 100, 100, 0.12);
  color: #ff7675;
  &.success {
    background: rgba(129, 199, 132, 0.12);
    color: #81c784;
  }
}
</style>
