<script setup lang="ts">
import { ref, reactive, computed, defineAsyncComponent, onMounted, onUnmounted, watch } from 'vue'
import { t } from './i18n'
import FloatingSidebar from './Components/FloatingSidebar/FloatingSidebar.vue'
import UserCard from './Widgets/UserCard/UserCard.vue'
import InstallProgressCard from './Widgets/InstallProgressCard/InstallProgressCard.vue'
import StatusBar from './Widgets/StatusBar/StatusBar.vue'
import FriendsWidget from './Widgets/FriendsWidget/FriendsWidget.vue'
import NotificationContainer from './Components/Notifications/NotificationContainer.vue'
import type { InstallCardState } from './Widgets/InstallProgressCard/InstallProgressCard.vue'
import { useAuth } from './Composables/useAuth'
import { useNotifications } from './Composables/useNotifications'
import { useCrashHistory } from './Composables/useCrashHistory'

const LoginPanel = defineAsyncComponent(() => import('./Panels/LoginPanel.vue'))
const ConfigurationPanel = defineAsyncComponent(() => import('./Panels/ConfigurationPanel.vue'))
const NewsSection = defineAsyncComponent(() => import('./Panels/NewsSection.vue'))
const FriendsPanel = defineAsyncComponent(() => import('./Panels/FriendsPanel.vue'))
const InstancesPanel = defineAsyncComponent(() => import('./Panels/InstancesPanel.vue'))
const ScreenshotsPanel = defineAsyncComponent(() => import('./Panels/ScreenshotsPanel.vue'))
const ModsPanel = defineAsyncComponent(() => import('./Panels/ModsPanel.vue'))
const CrashPanel = defineAsyncComponent(() => import('./Panels/CrashPanel/CrashPanel.vue'))
const DownloadModal = defineAsyncComponent(() => import('./Modals/DownloadModal/DownloadModal.vue'))
const LaunchModal = defineAsyncComponent(() => import('./Modals/LaunchModal/LaunchModal.vue'))

const { initializing, isAuthenticated, user, initialize } = useAuth()
const { success: notifySuccess, info: notifyInfo, error: notifyError, warning: notifyWarning } = useNotifications()
const { crashHistory, addCrash, removeCrash } = useCrashHistory()

function notifyOnBg(opts: { title: string; body: string }) {
  if (document.hidden) {
    window.ElectronAPI.ShowNotification(opts)
  }
}

initialize()

let welcomeShown = false

watch(isAuthenticated, (val) => {
  if (val && user.value && !welcomeShown) {
    welcomeShown = true
    setTimeout(() => notifySuccess(t('notifications.welcome', { username: user.value!.username })), 500)
  }
})

function applyPersonalization(p: PersonalizationConfig) {
  const root = document.documentElement.style
  root.setProperty('--background-title-bar', p.titleBarColor)
  document.body.style.backgroundImage = p.appBackground
  root.setProperty('--font-family-primary', `'${p.fontPrimary}'`)
  root.setProperty('--font-family-secundary', `'${p.fontSecondary}'`)
  root.setProperty('--accent-color', p.accentColor)
  root.setProperty('--modal-accent', p.modalAccent || p.accentColor)
  root.setProperty('--modal-bg', p.modalBackground || 'rgba(18, 18, 30, 0.97)')
  root.setProperty('--sidebar-bg', p.sidebarBackground)
  root.setProperty('--panel-bg', p.panelBackground)
  root.setProperty('--notif-error-text', p.notificationErrorColor)
  root.setProperty('--notif-warn-text', p.notificationWarnColor)
  root.setProperty('--notif-success-text', p.notificationSuccessColor)
  root.setProperty('--notif-error-bg', p.notificationErrorColor + '33')
  root.setProperty('--notif-warn-bg', p.notificationWarnColor + '33')
  root.setProperty('--notif-success-bg', p.notificationSuccessColor + '33')
  root.setProperty('--notif-error-border', p.notificationErrorColor + '4d')
  root.setProperty('--notif-warn-border', p.notificationWarnColor + '4d')
  root.setProperty('--notif-success-border', p.notificationSuccessColor + '4d')
  root.setProperty('--sidebar-btn-color', p.sidebarButtonColor)
  root.setProperty('--sidebar-position', p.sidebarPosition || 'left')
  document.documentElement.classList.toggle('macos-titlebar', p.macOSTitlebar)
  document.documentElement.classList.toggle('hide-icon', !p.showIcon)
  document.documentElement.classList.toggle('invert-position', p.invertPosition)
  sidebarPos.value = p.sidebarPosition || 'left'
}

function applyLauncherConfig(l: LauncherConfig) {
  const root = document.documentElement
  root.classList.toggle('blur-disabled', !l.blur)
  root.classList.toggle('filters-disabled', !l.filters)
  root.classList.toggle('shadows-disabled', !l.shadows)
  root.classList.toggle('hide-console', !l.showConsole)
  root.classList.toggle('hide-news', !l.showNews)
}

onMounted(async () => {
  try {
    const c = await window.ConfigManager.Get()
    applyPersonalization(c.personalization)
    applyLauncherConfig(c.launcher)
  } catch {}

  window.addEventListener('sidebar-position-changed', ((e: CustomEvent) => {
    sidebarPos.value = e.detail
  }) as EventListener)

  setupInstallEvents()

  window.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      if (currentView.value) {
        currentView.value = null
      }
      consoleVisible.value = false
    }
  })

})

const currentView = ref<string | null>(null)
const selectedProfileUuid = ref<string | null>(null)
const sidebarPos = ref<'left' | 'right'>('left')

function openProfile(uuid: string) {
  selectedProfileUuid.value = uuid
  currentView.value = 'friends'
}

const consoleVisible = ref(false)
const consoleLogs = ref<string[]>([])
const consoleRef = ref<HTMLElement | null>(null)

const crashInfo = ref<{ reason: string; context: string[]; exitCode: number; timestamp?: number } | null>(null)
const crashPanelVisible = ref(false)
const consoleTab = ref<'logs' | 'crashes'>('logs')

function formatCrashTime(ts: number) {
  const d = new Date(ts)
  return d.toLocaleString('es-ES', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function viewCrash(record: { reason: string; context: string[]; exitCode: number; timestamp?: number }) {
  crashInfo.value = { reason: record.reason, context: record.context, exitCode: record.exitCode, timestamp: record.timestamp }
  crashPanelVisible.value = true
  consoleVisible.value = false
}

function fmtLog(event: string, data: unknown): string {
  if (typeof data === 'string') return `[${event}] ${data}`
  if (!data || typeof data !== 'object') return `[${event}] ${String(data)}`
  const d = data as Record<string, unknown>
  if (event === 'game_log' || event === 'game_stdout' || event === 'game_stderr' || event === 'game_log_error') {
    const msg = (d.message as string) || (d.line as string) || JSON.stringify(d)
    return msg
  }
  if (event === 'engine_log') {
    return (d.line as string) || JSON.stringify(d)
  }
  if (event === 'game_crash' || event === 'launch_failed') {
    const reason = (d.reason as string) || (d.error as string) || 'Unknown'
    const context = (d.context as string[]) || []
    const exitCode = (d.exitCode as number) ?? -1
    const timestamp = Date.now()
    crashInfo.value = { reason, context, exitCode, timestamp }
    addCrash(reason, context, exitCode)
    crashPanelVisible.value = true
    return t('console.crash.prefix', { exitCode, reason })
  }
  if (event === 'launch_exited') {
    const normal = d.normal as boolean
    const exitCode = (d.exitCode as number) ?? 0
    if (!normal && exitCode !== 0) {
      const reason = (d.reason as string) || `Game exited with code ${exitCode}`
      const timestamp = Date.now()
      crashInfo.value = { reason, context: [], exitCode, timestamp }
      addCrash(reason, [], exitCode)
      crashPanelVisible.value = true
      return t('console.crash.prefix', { exitCode, reason })
    }
    return t('console.log_prefix.launch_exit', { exitCode, normal })
  }
  return `[${event}] ${JSON.stringify(d)}`
}

let consoleUnsub: (() => void) | null = null

watch(consoleVisible, async (val) => {
  if (val) {
    crashPanelVisible.value = false
    if (consoleLogs.value.length === 0) {
      try {
        const [info, sys] = await Promise.all([
          window.NovaCoreManager.GetInfo().catch(() => null),
          window.ConfigManager.GetSystemInfo().catch(() => null),
        ])
        if (info) consoleLogs.value.push(t('console.log_prefix.engine', { version: info.version || '?', baseDir: info.baseDir || '?' }))
        if (sys) {
          consoleLogs.value.push(`[system] ${sys.os || '?'} | ${sys.arch || '?'} | ${sys.cpu || '?'}`)
          consoleLogs.value.push(t('console.log_prefix.system_ram', { totalRam: sys.totalRam || '?', ramFree: sys.ramFree || '?', java: sys.java || '?' }))
          if (sys.gpu && sys.gpu !== 'No detectado') consoleLogs.value.push(t('console.log_prefix.system_gpu', { gpu: sys.gpu }))
        }
      } catch {}
    }
    if (consoleUnsub) consoleUnsub()
    consoleUnsub = window.NovaCoreManager.OnEvent((data) => {
      const text = fmtLog(data.event, data.data)
      consoleLogs.value.push(text)
      if (consoleLogs.value.length > 1000) consoleLogs.value.splice(0, 500)
      if (consoleRef.value) setTimeout(() => { consoleRef.value!.scrollTop = consoleRef.value!.scrollHeight }, 0)
    })
  } else {
    if (consoleUnsub) {
      consoleUnsub()
      consoleUnsub = null
    }
  }
})

const installState = reactive<InstallCardState>({
  active: false,
  instanceName: '',
  percent: 0,
  statusText: '',
  completedFiles: 0,
  skippedFiles: 0,
  totalFiles: 0,
  downloadedMb: 0,
  totalMb: 0,
  modules: [],
  logs: [],
  error: '',
})

let currentSessionId = ''
let installUnsub: (() => void) | null = null
let dismissedSession = ''

const stepLabels: Record<string, string> = {
  resolving_version: t('install.steps.resolving_version'),
  fetching_asset_index: t('install.steps.fetching_asset_index'),
  downloading_jvm: t('install.steps.downloading_jvm'),
  building_task_list: t('install.steps.building_task_list'),
  downloading: t('install.steps.downloading'),
  verifying: t('install.steps.verifying'),
  retrying: t('install.steps.retrying'),
  extracting_natives: t('install.steps.extracting_natives'),
  modloader: t('install.steps.modloader'),
}

function stepLabel(step: string): string { return stepLabels[step] || step }

function setupInstallEvents() {
  if (installUnsub) installUnsub()
  try {
    if (window.NovaCoreManager?.OnEvent) {
      installUnsub = window.NovaCoreManager.OnEvent((ev: any) => {
        const { event, data } = ev
        if (!data) return

        switch (event) {
          case 'install_step':
            installState.statusText = stepLabel(data.step)
            break
          case 'tasks_ready':
            installState.totalFiles = data.totalTasks ?? 0
            installState.totalMb = (data.totalBytes ?? 0) / 1048576
            installState.statusText = t('install.status.preparing_download')
            break
          case 'session_started':
            installState.active = true
            installState.totalFiles = data.totalFiles ?? 0
            installState.totalMb = (data.totalBytes ?? 0) / 1048576
            if (!installState.instanceName) installState.instanceName = t('install.default_instance_name')
            break
          case 'session_progress':
            installState.percent = data.overallPercent ?? data.percent ?? 0
            installState.completedFiles = data.completedFiles ?? 0
            installState.skippedFiles = data.skippedFiles ?? 0
            installState.downloadedMb = (data.downloadedBytes ?? data.totalBytes ?? 0) / 1048576
            break
          case 'module_status': {
            const idx = installState.modules.findIndex((m: any) => m.module === data.module)
            if (idx >= 0) installState.modules[idx]!.status = data.status
            else installState.modules.push({ module: data.module, status: data.status })
            break
          }
          case 'install_failed':
            installState.active = true
            installState.error = data.reason ?? t('install.errors.installation')
            installState.statusText = t('install.status.error')
            notifyError(t('notifications.install.error', { error: installState.error }), 8000)
            notifyOnBg({ title: t('notifications.install.error_title'), body: t('notifications.install.error_body', { instance: installState.instanceName, error: installState.error }) })
            break
          case 'session_completed':
            installState.active = true
            installState.statusText = t('install.status.completed')
            installState.percent = 100
            if (installState.instanceName && installState.instanceName !== t('install.default_instance_name')) {
              notifySuccess(t('notifications.install.complete', { instance: installState.instanceName }))
              notifyOnBg({ title: t('notifications.install.complete_title'), body: t('notifications.install.complete_body', { instance: installState.instanceName }) })
            }
            break
          case 'session_failed':
            if (!installState.error) {
              installState.active = true
              installState.error = data.reason ?? t('install.errors.session')
              installState.statusText = t('install.status.error')
              if (installState.instanceName && installState.instanceName !== t('install.default_instance_name')) {
                notifyError(t('notifications.install.session_error', { error: installState.error }), 8000)
                notifyOnBg({ title: t('notifications.install.session_error_title'), body: t('notifications.install.session_error_body', { error: installState.error }) })
              }
            }
            break
          case 'download_start':
            installState.logs.push(t('install.log.downloading_file', { file: data.file }))
            if (installState.logs.length > 100) installState.logs.splice(0, 50)
            break
          case 'modloader_processor_log':
            if (data.message) {
              installState.logs.push(data.message)
              installState.statusText = t('install.status.processing_modloader')
            }
            break
        }
      })
    }
  } catch {}
}

async function handleInstall(req: { version: string; modloader: string; modloaderVersion: string; name: string }) {
  const info = await window.NovaCoreManager.GetInfo()

  installState.active = true
  installState.instanceName = req.name
  installState.percent = 0
  installState.statusText = t('install.status.starting')
  installState.completedFiles = 0
  installState.skippedFiles = 0
  installState.totalFiles = 0
  installState.downloadedMb = 0
  installState.totalMb = 0
  installState.modules = []
  installState.logs = []
  installState.error = ''

  currentSessionId = ''

  try {
    const result = await window.NovaCoreManager.Install({
      version: req.version,
      instancePath: info.baseDir,
      sharedPath: info.baseDir,
      launcher: { name: 'StepLauncher' },
      modloader: req.modloader !== 'none' ? req.modloader : undefined,
      modloaderVersion: req.modloaderVersion || undefined,
      download: { jvm: true },
    })

    if (!result.success) {
      installState.error = result.error ?? t('install.errors.unknown')
      installState.statusText = t('install.status.error')
      return
    }

    currentSessionId = result.sessionId ?? ''
  } catch (err: any) {
    installState.error = err?.message ?? t('install.errors.connecting_engine')
    installState.statusText = t('install.status.error')
  }
}

function cancelInstall() {
  if (currentSessionId) {
    window.NovaCoreManager.CancelInstall(currentSessionId).catch(() => {})
  }
  currentSessionId = ''
  installState.error = t('install.status.cancelled')
  installState.statusText = t('install.status.cancelled')
  notifyWarning(t('notifications.install.cancelled'))
}

function closeInstallCard() {
  installState.active = false
  currentSessionId = ''
}

function onSidebarPosChange(pos: 'left' | 'right') {
  sidebarPos.value = pos
  window.dispatchEvent(new CustomEvent('sidebar-position-changed', { detail: pos }))
  window.ConfigManager.UpdatePersonalization({ sidebarPosition: pos }).catch(() => {})
}

function onInstallStarted(payload: { name: string }) {
  installState.active = true
  installState.instanceName = payload.name
  installState.percent = 0
  installState.statusText = t('install.status.starting')
  installState.completedFiles = 0
  installState.skippedFiles = 0
  installState.totalFiles = 0
  installState.downloadedMb = 0
  installState.totalMb = 0
  installState.modules = []
  installState.logs = []
  installState.error = ''
}

async function handleLaunch(payload: { version: string }) {
  currentView.value = null
  try {
    const [info, token, freshConfig] = await Promise.all([
      window.NovaCoreManager.GetInfo(),
      window.AuthManager.GetToken(),
      window.ConfigManager.Get(),
    ])
    const mc = freshConfig.minecraft
    infoVersion = info.version

    const jvmArgs = (mc?.jvmArgs ?? '').split('\n').filter(Boolean)

    const result = await window.NovaCoreManager.Launch({
      version: payload.version,
      instancePath: info.baseDir,
      sharedPath: info.baseDir,
      javaPath: mc?.javaPath || undefined,
      gcPreset: (mc?.gcPreset && mc.gcPreset !== 'none') ? mc.gcPreset : undefined,
      gpuPreference: mc?.gpuPreference || undefined,
      auth: {
        username: user.value?.username ?? t('electron.default_username'),
        uuid: user.value?.uuid ?? '',
        accessToken: token ?? '',
        userType: 'mojang',
      },
      authlibInjector: {
        enabled: true,
        jarPath: info.authlibPath,
        serverUrl: 'https://api.stepnicka012.workers.dev',
      },
      jvm: {
        maxMemoryMb: mc?.maxRam ?? 4096,
        minMemoryMb: mc?.minRam ?? 2048,
        extraArgs: jvmArgs,
      },
      window: {
        width: mc?.windowWidth ?? 854,
        height: mc?.windowHeight ?? 480,
        fullscreen: mc?.fullscreen ?? false,
      },
      game: {
        extraGameArgs: (mc?.gameArgs ?? '').split('\n').filter(Boolean),
      },
      launcher: { name: 'StepLauncher' },
    })
    if (!result.success) {
      notifyError(t('notifications.launch.error', { error: result.error ?? t('general.error') }))
    } else {
      notifySuccess(t('notifications.launch.success', { version: payload.version }))
      notifyOnBg({ title: t('notifications.launch.success_title'), body: t('notifications.launch.success_body', { version: payload.version }) })

      window.AuthManager.UpdateStats({ last_version_playing: payload.version, last_version_connected: info.version }).catch(() => {})
      stopHeartbeat()
      heartbeatInterval = setInterval(() => {
        window.AuthManager.UpdateHeartbeat(info.version, 30).catch(() => {})
      }, 30000)

      if (freshConfig.launcher.hideOnLaunch) {
        window.ElectronAPI.Hide()
        const launchId = result.launchId
        if (launchId) {
          let gameStarted = false
          const checkInterval = setInterval(async () => {
            try {
              const instances = await window.NovaCoreManager.RunningInstances()
              const isRunning = instances.some((i: any) => {
                const id = i?.launchId ?? i?.id ?? i?.instanceId
                return id === launchId
              })
              if (isRunning) {
                gameStarted = true
                return
              }
              if (gameStarted && !isRunning) {
                clearInterval(checkInterval)
                stopHeartbeat()
                window.ElectronAPI.Show()
                notifyOnBg({ title: t('notifications.game.closed_title'), body: t('notifications.game.closed_body', { version: payload.version }) })
              }
            } catch {
              clearInterval(checkInterval)
              stopHeartbeat()
              window.ElectronAPI.Show()
            }
          }, 5000)
        }
      }
    }
  } catch (err: any) {
    notifyError(t('notifications.launch.generic_error', { error: err?.message ?? t('general.error') }))
  }
}

const pendingFriendRequests = ref(0)

const Items = computed(() => [
  { text: t('navigation.sidebar.config'), icon: 'assets/svg/settings.svg', action: () => { currentView.value = 'config' } },
  { text: t('navigation.sidebar.friends'), icon: 'assets/svg/friends.svg', action: () => { currentView.value = 'friends' }, badge: pendingFriendRequests.value },
  { text: t('navigation.sidebar.screenshots'), icon: 'assets/svg/photo.svg', action: () => { currentView.value = 'screenshots' } },
  { text: t('navigation.sidebar.mods'), icon: 'assets/svg/mod.svg', action: () => { currentView.value = 'mods' } },
  { text: t('navigation.sidebar.instances'), icon: 'assets/svg/instance.svg', action: () => { currentView.value = 'instances' } },
  { text: t('navigation.sidebar.downloads'), icon: 'assets/svg/download.svg', action: () => { currentView.value = 'download' } },
  { text: t('navigation.sidebar.play'), icon: 'assets/svg/play.svg', action: () => { currentView.value = 'launch' } },
])

let heartbeatInterval: ReturnType<typeof setInterval> | null = null
let friendPollTimer: ReturnType<typeof setInterval> | null = null
let infoVersion = ''


function stopHeartbeat() {
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval)
    heartbeatInterval = null
  }
  if (infoVersion) window.AuthManager.UpdateHeartbeat(infoVersion, 0).catch(() => {})
}

async function pollFriendRequests() {
  if (!isAuthenticated.value) return
  try {
    const requests = await window.AuthManager.GetFriendRequests()
    const received = Array.isArray((requests as any)?.received) ? (requests as any).received : []
    pendingFriendRequests.value = received.length
  } catch {}
}

onMounted(() => {
  if (friendPollTimer) clearInterval(friendPollTimer)
  friendPollTimer = setInterval(pollFriendRequests, 30000)
  setTimeout(pollFriendRequests, 1000)
})

onUnmounted(() => {
  if (installUnsub) { installUnsub(); installUnsub = null }
  stopHeartbeat()
  if (friendPollTimer) {
    clearInterval(friendPollTimer)
    friendPollTimer = null
  }
})
</script>

<template>
  <div class="app-container">
    <LoginPanel v-if="!initializing && !isAuthenticated" />
    <template v-else-if="!initializing && isAuthenticated">
      <UserCard v-if="user && !currentView && !consoleVisible" :user="user" @open-profile="openProfile" />

      <template v-if="currentView === 'config'">
        <div class="view-overlay" />
        <ConfigurationPanel @back="currentView = null" />
      </template>

      <DownloadModal
        v-else-if="currentView === 'download'"
        :install-state="installState"
        @close="currentView = null"
        @start-install="handleInstall"
        @cancel="cancelInstall"
      />

      <LaunchModal
        v-else-if="currentView === 'launch'"
        @close="currentView = null"
        @play="handleLaunch"
      />

      <NewsSection
        v-else-if="currentView === 'news'"
        @close="currentView = null"
      />

      <FriendsPanel
        v-else-if="currentView === 'friends'"
        :initial-uuid="selectedProfileUuid"
        @close="currentView = null; selectedProfileUuid = null"
      />

      <InstancesPanel
        v-else-if="currentView === 'instances'"
        :app-install-state="installState"
        @close="currentView = null"
        @install-started="onInstallStarted"
      />

      <ScreenshotsPanel
        v-else-if="currentView === 'screenshots'"
        @close="currentView = null"
      />

      <ModsPanel
        v-else-if="currentView === 'mods'"
        @close="currentView = null"
      />

      <FriendsWidget v-if="!currentView && !consoleVisible" @open-profile="openProfile" />
      <StatusBar v-if="!currentView && !consoleVisible" />
      <FloatingSidebar v-if="!currentView && !consoleVisible" :items="Items" :position="sidebarPos" @update:position="onSidebarPosChange" />

      <div v-if="!currentView && !consoleVisible" class="bottom-btns">
        <button class="bottom-btn bottom-btn-news" @click="currentView = 'news'">
          <img :src="'assets/svg/news.svg'" width="12" height="12" />
          {{ $t('navigation.bottom.news') }}
        </button>
        <button class="bottom-btn bottom-btn-console" @click="consoleVisible = !consoleVisible">
          <img :src="'assets/svg/terminal.svg'" width="12" height="12" />
          {{ $t('navigation.bottom.console') }}
        </button>
      </div>

      <div v-if="consoleVisible" class="console-panel">
        <div class="panel-header">
          <div class="console-tabs">
            <button :class="['console-tab', { active: consoleTab === 'logs' }]" @click="consoleTab = 'logs'">
              <img :src="'assets/svg/terminal.svg'" width="12" height="12" />
              {{ $t('console.tabs.events') }}
            </button>
            <button :class="['console-tab', { active: consoleTab === 'crashes' }]" @click="consoleTab = 'crashes'">
              <img :src="'assets/svg/alert-circle.svg'" width="12" height="12" />
              {{ $t('console.tabs.crash_history') }}
              <span v-if="crashHistory.length" class="crash-badge-count">{{ crashHistory.length }}</span>
            </button>
          </div>
          <button class="panel-close" @click="consoleVisible = false">&times;</button>
        </div>

        <div v-if="consoleTab === 'logs'" class="console-body" ref="consoleRef">
          <div v-if="consoleLogs.length === 0" class="panel-empty">{{ $t('console.empty.no_events') }}</div>
          <div v-for="(line, i) in consoleLogs.slice(-500)" :key="i" class="console-line">{{ line }}</div>
        </div>

        <div v-if="consoleTab === 'crashes'" class="console-body">
          <div v-if="crashHistory.length === 0" class="panel-empty">{{ $t('console.empty.no_crashes') }}</div>
          <div v-for="crash in crashHistory" :key="crash.id" class="crash-history-item">
            <div class="chi-header" @click="crash._expanded = !crash._expanded">
              <div class="chi-left">
                <div class="chi-dot" />
                <div class="chi-info">
                  <span class="chi-reason">{{ crash.reason }}</span>
                  <span class="chi-meta">{{ $t('console.crash.code', { exitCode: crash.exitCode }) }} · {{ formatCrashTime(crash.timestamp) }}</span>
                </div>
              </div>
              <div class="chi-actions">
                <button class="chi-btn" @click.stop="viewCrash(crash)">{{ $t('console.crash.view') }}</button>
                <button class="chi-btn chi-btn--danger" @click.stop="removeCrash(crash.id)">×</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <CrashPanel :visible="crashPanelVisible" :info="crashInfo" @close="crashPanelVisible = false" @open-console="crashPanelVisible = false; consoleVisible = true" />

      <InstallProgressCard v-if="installState.active && currentView !== 'download' && currentView !== 'instances'" :state="installState" @cancel="cancelInstall" @close="closeInstallCard" />
    </template>
    <NotificationContainer />
  </div>
</template>

<style lang="scss" src="./Styles/Global.scss"></style>
<style scoped lang="scss" src="./Styles/App.scss"></style>
