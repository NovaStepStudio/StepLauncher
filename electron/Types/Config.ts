export interface LauncherConfig {
  autoStartMinecraft: boolean
  hideOnLaunch: boolean
  showConsole: boolean
  showNews: boolean
  filters: boolean
  blur: boolean
  hardwareAcceleration: boolean
  shadows: boolean
  discordRpc: boolean
  locale: string
}

export interface MinecraftConfig {
  useRecommendedJava: boolean
  maxConsoleEvents: number
  showNotificationOnLaunch: boolean
  cleanBeforeLaunch: boolean
  javaPath: string
  fullscreen: boolean
  windowWidth: number
  windowHeight: number
  maxRam: number
  minRam: number
  gcPreset: string
  gpuPreference: string
  jvmArgs: string
  gameArgs: string
}

export interface PersonalizationConfig {
  titleBarColor: string
  appBackground: string
  fontPrimary: string
  fontSecondary: string
  accentColor: string
  modalAccent?: string
  modalBackground?: string
  sidebarBackground: string
  panelBackground: string
  notificationErrorColor: string
  notificationWarnColor: string
  notificationSuccessColor: string
  sidebarButtonColor: string
  macOSTitlebar: boolean
  showIcon: boolean
  invertPosition: boolean
  sidebarPosition: 'left' | 'right'
}

export interface AppConfig {
  launcher: LauncherConfig
  minecraft: MinecraftConfig
  personalization: PersonalizationConfig
}
