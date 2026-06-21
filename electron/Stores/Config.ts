import { readFileSync, writeFileSync, existsSync, mkdirSync, renameSync, copyFileSync, unlinkSync } from "fs"
import { join } from "path"
import { homedir } from "os"
import type { AppConfig, LauncherConfig, MinecraftConfig, PersonalizationConfig } from "../Types/Config.js"

const DEFAULT_CONFIG: AppConfig = {
  launcher: {
    autoStartMinecraft: true,
    hideOnLaunch: true,
    showConsole: false,
    showNews: true,
    filters: true,
    blur: true,
    hardwareAcceleration: true,
    shadows: true,
    discordRpc: true,
    locale: "es-AR",
  },
  minecraft: {
    useRecommendedJava: true,
    maxConsoleEvents: 100,
    showNotificationOnLaunch: true,
    cleanBeforeLaunch: true,
    javaPath: "",
    fullscreen: false,
    windowWidth: 854,
    windowHeight: 480,
    maxRam: 4096,
    minRam: 2048,
    gcPreset: "g1gc_optimized",
    gpuPreference: "auto",
    jvmArgs: "-XX:+AlwaysPreTouch\n-XX:+DisableExplicitGC",
    gameArgs: "",
  },
  personalization: {
    titleBarColor: "#111",
    appBackground: "url('assets/background/RRE36/19.webp')",
    fontPrimary: "Lexend",
    fontSecondary: "Inter",
    accentColor: "#5cd0e7",
    modalAccent: "#5cd0e7",
    modalBackground: "rgba(18, 18, 30, 0.97)",
    sidebarBackground: "rgba(8,8,16,0.65)",
    panelBackground: "rgba(8,8,16,0.55)",
    notificationErrorColor: "#e57373",
    notificationWarnColor: "#ffd54f",
    notificationSuccessColor: "#81c784",
    sidebarButtonColor: "rgba(255,255,255,0.06)",
    macOSTitlebar: false,
    showIcon: true,
    invertPosition: false,
    sidebarPosition: "left",
  },
}

const DEFAULT_PERSONALIZATION: PersonalizationConfig = {
  ...DEFAULT_CONFIG.personalization,
}

function getBaseDir(): string {
  if (process.platform === "win32") {
    return join(process.env.APPDATA || homedir(), ".StepLauncher")
  }
  return join(homedir(), ".StepLauncher")
}

function getConfigDir(): string {
  const dir = join(getBaseDir(), "bin", "config")
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  return dir
}

function getConfigPath(): string {
  return join(getConfigDir(), "settings.json")
}

function getBackupPath(): string {
  return getConfigPath() + ".bak"
}

function getPersonalizationPath(): string {
  return join(getConfigDir(), "personalization.json")
}

let cached: AppConfig | null = null

function mergeConfig(parsed: Partial<AppConfig>): AppConfig {
  return {
    launcher: { ...DEFAULT_CONFIG.launcher, ...parsed.launcher },
    minecraft: { ...DEFAULT_CONFIG.minecraft, ...parsed.minecraft },
    personalization: { ...DEFAULT_CONFIG.personalization, ...parsed.personalization },
  }
}

function mergePersonalization(parsed: Partial<PersonalizationConfig>): PersonalizationConfig {
  return { ...DEFAULT_PERSONALIZATION, ...parsed }
}

function tryReadFile(path: string): AppConfig | null {
  try {
    const raw = readFileSync(path, "utf-8")
    return mergeConfig(JSON.parse(raw))
  } catch {
    return null
  }
}

export function LoadConfig(): AppConfig {
  if (cached) return cached

  const configPath = getConfigPath()
  const backupPath = getBackupPath()

  let loaded = tryReadFile(configPath)
  if (loaded) {
    cached = loaded
    return cached
  }

  loaded = tryReadFile(backupPath)
  if (loaded) {
    cached = loaded
    try { copyFileSync(backupPath, configPath) } catch {}
    return cached
  }

  cached = structuredClone(DEFAULT_CONFIG)
  SaveConfig(cached)
  return cached
}

export function LoadPersonalizationConfig(): PersonalizationConfig {
  const pPath = getPersonalizationPath()
  try {
    const raw = readFileSync(pPath, "utf-8")
    return mergePersonalization(JSON.parse(raw))
  } catch {
    const fromMain = cached?.personalization
    if (fromMain) {
      SavePersonalizationConfig(fromMain)
      return fromMain
    }
    return structuredClone(DEFAULT_PERSONALIZATION)
  }
}

export function SavePersonalizationConfig(personalization: PersonalizationConfig): void {
  const pPath = getPersonalizationPath()
  const tmpPath = pPath + ".tmp"
  try {
    writeFileSync(tmpPath, JSON.stringify(personalization, null, 2), "utf-8")
    renameSync(tmpPath, pPath)
  } catch (e) {
    try { if (existsSync(tmpPath)) unlinkSync(tmpPath) } catch {}
    throw e
  }
}

export function SaveConfig(config: AppConfig): void {
  cached = config
  const configPath = getConfigPath()
  const tmpPath = configPath + ".tmp"
  try {
    writeFileSync(tmpPath, JSON.stringify(config, null, 2), "utf-8")
    if (existsSync(configPath)) {
      copyFileSync(configPath, getBackupPath())
    }
    renameSync(tmpPath, configPath)
  } catch (e) {
    try { if (existsSync(tmpPath)) unlinkSync(tmpPath) } catch {}
    throw e
  }
}

export function GetConfig(): AppConfig {
  const c = LoadConfig()
  c.personalization = LoadPersonalizationConfig()
  return c
}

export function UpdateLauncherConfig(updates: Partial<LauncherConfig>): AppConfig {
  const config = LoadConfig()
  config.launcher = { ...config.launcher, ...updates }
  SaveConfig(config)
  return config
}

export function UpdateMinecraftConfig(updates: Partial<MinecraftConfig>): AppConfig {
  const config = LoadConfig()
  config.minecraft = { ...config.minecraft, ...updates }
  SaveConfig(config)
  return config
}

export function UpdatePersonalizationConfig(updates: Partial<PersonalizationConfig>): AppConfig {
  const config = LoadConfig()
  const personalization = { ...LoadPersonalizationConfig(), ...updates }
  SavePersonalizationConfig(personalization)
  config.personalization = personalization
  cached = config
  return config
}
