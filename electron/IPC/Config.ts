import { ipcMain } from "electron"
import { totalmem, freemem, cpus, arch, platform, release, hostname, userInfo } from "os"
import { execSync } from "child_process"
import * as ConfigStore from "../Stores/Config.js"
import { startRPC, stopRPC } from "../Services/DiscordRPC.js"

function getJavaVersion(): string {
  try {
    const out = execSync("java -version 2>&1", { timeout: 5000, encoding: "utf-8" })
    const match = out.match(/(\d+\.\d+\.\d+[^"]*)/)
    return match?.[1] ?? out.split("\n")[0]?.trim() ?? "No detectado"
  } catch {
    return "No detectado"
  }
}

function getGpuInfo(): string {
  try {
    if (process.platform === "win32") {
      const out = execSync('wmic path win32_VideoController get name 2>&1', { timeout: 3000, encoding: "utf-8" })
      const lines = out.split("\n").map(l => l.trim()).filter(Boolean)
      return lines.slice(1).join(", ") || "No detectado"
    }
    if (process.platform === "linux") {
      const out = execSync("glxinfo -B 2>&1 | grep 'Device\\|OpenGL renderer'", { timeout: 3000, encoding: "utf-8" })
      return out.split("\n")[0]?.trim() ?? "No detectado"
    }
  } catch {}
  return "No detectado"
}

function syncRPC(config: Record<string, unknown>): void {
  if (config.discordRpc === true) {
    startRPC()
  } else if (config.discordRpc === false) {
    stopRPC()
  }
}

export function RegisterConfigIpc() {
  ipcMain.handle("config:Get", () => ConfigStore.GetConfig())
  ipcMain.handle("config:UpdateLauncher", (_event, updates) => {
    ConfigStore.UpdateLauncherConfig(updates)
    syncRPC(updates)
  })
  ipcMain.handle("config:UpdateMinecraft", (_event, updates) => ConfigStore.UpdateMinecraftConfig(updates))
  ipcMain.handle("config:UpdatePersonalization", (_event, updates) => ConfigStore.UpdatePersonalizationConfig(updates))
  ipcMain.handle("config:GetSystemInfo", () => ({
    totalRam: totalmem(),
    os: `${platform()} ${release()}`,
    arch: arch(),
    hostname: hostname(),
    user: userInfo().username,
    cpu: `${cpus()[0]?.model ?? "?"} (${cpus().length} cores)`,
    ram: `${Math.round(totalmem() / 1073741824)} GB`,
    ramFree: `${Math.round(freemem() / 1073741824)} GB`,
    java: getJavaVersion(),
    gpu: getGpuInfo(),
  }))
}
