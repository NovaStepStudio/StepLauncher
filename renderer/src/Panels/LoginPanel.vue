<template>
  <div class="login-panel">
    <div class="panel-bg">
      <div
        v-for="(img, i) in backgrounds"
        :key="i"
        class="bg-slide"
        :class="{ active: i === activeIdx, prev: i === prevIdx }"
        :style="{ backgroundImage: `url(${img})` }"
      />
      <div class="bg-overlay" />
    </div>
    <div class="panel-inner">
      <div class="brand">
        <div class="logo-wrap">
          <img :src="iconUrl" alt="StepLauncher" class="logo" />
          <div class="logo-glow" />
        </div>
        <h1>{{ $t('auth.login.brand_title') }}</h1>
        <p>{{ $t('auth.login.brand_subtitle') }}</p>
      </div>

      <div class="tabs">
        <button
          :class="['tab', { active: mode === 'online' }]"
          @click="mode = 'online'"
        >{{ $t('auth.login.tabs.online') }}</button>
        <button
          :class="['tab', { active: mode === 'offline' }]"
          @click="mode = 'offline'"
        >{{ $t('auth.login.tabs.offline') }}</button>
      </div>

      <form v-if="mode === 'online'" @submit.prevent="handleOnlineLogin" class="form">
        <div class="field">
          <label>{{ $t('auth.login.fields.username_label') }}</label>
          <div class="input-wrap">
            <img class="input-icon" :src="'assets/svg/user-circle.svg'" width="14" height="14" />
            <input
              v-model="loginUsername"
              type="text"
              :placeholder="$t('auth.login.fields.username_placeholder')"
              :disabled="loading"
              autocomplete="username"
            />
          </div>
        </div>
        <div class="field">
          <label>{{ $t('auth.login.fields.password_label') }}</label>
          <div class="input-wrap">
            <img class="input-icon" :src="'assets/svg/lock.svg'" width="14" height="14" />
            <input
              v-model="loginPassword"
              type="password"
              :placeholder="$t('auth.login.fields.password_placeholder')"
              :disabled="loading"
              autocomplete="current-password"
            />
          </div>
        </div>
        <p v-if="onlineError" class="error">{{ onlineError }}</p>
        <button type="submit" class="btn btn-primary" :disabled="loading">
          <span v-if="loading" class="spinner" />
          <span v-else>{{ $t('auth.login.buttons.login') }}</span>
        </button>
        <button type="button" class="btn btn-ghost" :disabled="loading" @click="handleOAuth">
          <img :src="'assets/svg/external-link.svg'" width="15" height="15" />
          {{ $t('auth.login.buttons.oauth') }}
        </button>
      </form>

      <form v-else @submit.prevent="handleOfflineLogin" class="form">
        <div class="field">
          <label>{{ $t('auth.login.fields.offline_username_label') }}</label>
          <div class="input-wrap">
            <img class="input-icon" :src="'assets/svg/user-circle.svg'" width="14" height="14" />
            <input
              v-model="offlineUsername"
              type="text"
              :placeholder="$t('auth.login.fields.offline_username_placeholder')"
              maxlength="16"
            />
          </div>
        </div>
        <button type="submit" class="btn btn-secondary">{{ $t('auth.login.buttons.offline_play') }}</button>
      </form>

      <div class="divider" />

      <div class="links">
        <button class="link-btn" @click="openUrl('https://steplauncher.pages.dev/Register')">
          <img :src="'assets/svg/user-plus.svg'" width="13" height="13" />
          {{ $t('auth.login.links.create_account') }}
        </button>
        <button class="link-btn" @click="openUrl('https://steplauncher.pages.dev/forgot-password')">
          <img :src="'assets/svg/info.svg'" width="13" height="13" />
          {{ $t('auth.login.links.forgot_password') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { t } from '../i18n'
import { useAuth } from '../Composables/useAuth'
const iconUrl = 'assets/icon.png'

const emit = defineEmits<{ loggedIn: [] }>()

const { login, loginOffline } = useAuth()

const mode = ref<'online' | 'offline'>('online')
const loading = ref(false)
const loginUsername = ref('')
const loginPassword = ref('')
const onlineError = ref('')
const offlineUsername = ref('')

async function handleOnlineLogin() {
  onlineError.value = ''
  if (!loginUsername.value.trim() || !loginPassword.value.trim()) {
    onlineError.value = t('auth.login.errors.validation')
    return
  }
  loading.value = true
  const err = await login(loginUsername.value, loginPassword.value)
  loading.value = false
  if (err) {
    onlineError.value = err
    return
  }
  emit('loggedIn')
}

async function handleOAuth() {
  loading.value = true
  try {
    const { code, state } = await window.AuthManager.StartOAuth()
    if (!code) {
      loading.value = false
      return
    }
    await window.AuthManager.CompleteOAuth(code, state)
    const result = await window.AuthManager.Initialize()
    if (result.authenticated) emit('loggedIn')
  } catch {
    onlineError.value = t('auth.login.errors.auth_web')
  }
  loading.value = false
}

function handleOfflineLogin() {
  if (!offlineUsername.value.trim()) return
  loginOffline(offlineUsername.value)
  emit('loggedIn')
}

function openUrl(url: string) {
  window.ElectronAPI.OpenExternal(url)
}

const backgrounds = Array.from({ length: 61 }, (_, i) => `assets/background/RRE36/${i + 1}.webp`)

const activeIdx = ref(Math.floor(Math.random() * backgrounds.length))
const prevIdx = ref(-1)
let interval: ReturnType<typeof setInterval> | null = null

function nextBg() {
  let next = activeIdx.value
  while (next === activeIdx.value) {
    next = Math.floor(Math.random() * backgrounds.length)
  }
  prevIdx.value = activeIdx.value
  activeIdx.value = next
}

onMounted(() => {
  interval = setInterval(nextBg, 8000)
})

onUnmounted(() => {
  if (interval) clearInterval(interval)
})
</script>

<style scoped lang="scss" src="../Styles/Panels/LoginPanel.scss"></style>
