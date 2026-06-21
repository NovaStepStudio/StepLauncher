import { app, BrowserWindow, Tray, Menu, nativeImage, ipcMain, shell } from "electron"
import { fileURLToPath } from "url"
import { dirname, join } from "path"
import { initLogger, restoreConsole } from "./Services/Logger.js"
import { RegisterAuthIpc, HandleOAuthCallback } from "./IPC/Auth.js"
import { RegisterConfigIpc } from "./IPC/Config.js"
import { RegisterDownloadsIpc } from "./IPC/Downloads.js"
import { RegisterNovaCoreIpc } from "./IPC/NovaCore.js"
import { RegisterInstancesIpc } from "./IPC/Instances.js"
import { RegisterThemesIpc } from "./IPC/Themes.js"
import { RegisterModsIpc } from "./IPC/Mods.js"
import { RegisterCacheIpc } from "./IPC/Cache.js"
import { RegisterLangManagerIpc } from "./IPC/LangManager.js"
import { startRPC, stopRPC } from "./Services/DiscordRPC.js"
import * as ConfigStore from "./Stores/Config.js"

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

function getIcon(): string {
  const extension: Record<string, string> = {
    win32: "ico",
    linux: "png",
  }
  const ext = extension[process.platform] ?? "png"

  if (app.isPackaged) {
    return join(process.resourcesPath, "assets", `icon.${ext}`)
  }

  return join(__dirname, "../renderer/assets", `icon.${ext}`)
}

let win: BrowserWindow | null = null
let tray: Tray | null = null

function createTray(): void {
  const iconPath = getIcon()
  const icon = nativeImage.createFromPath(iconPath).resize({ width: 16, height: 16 })
  tray = new Tray(icon)
  tray.setToolTip("StepLauncher")

  const contextMenu = Menu.buildFromTemplate([
    {
      label: "Mostrar",
      click: () => {
        if (win) { win.show(); win.focus() }
      }
    },
    { type: "separator" },
    {
      label: "Salir",
      click: () => app.quit()
    }
  ])

  tray.setContextMenu(contextMenu)
  tray.on("click", () => {
    if (win) { win.show(); win.focus() }
  })
}

const gotLock = app.requestSingleInstanceLock()

if (!gotLock) {
  app.quit()
} else {
  app.on("second-instance", (_event, argv) => {
    const url = argv.find((arg) => arg.startsWith("steplauncher://"))
    if (url) HandleOAuthCallback(url)

    if (win) {
      if (win.isMinimized()) win.restore()
      win.focus()
    }
  })
}

function CreateWindow(): void {
  win = new BrowserWindow({
    height: 600,
    width: 1000,
    minHeight: 600,
    minWidth: 900,
    frame: false,
    title: "StepLauncher",
    icon: getIcon(),
    webPreferences: {
      webSecurity: false,
      backgroundThrottling: true,
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
      preload: join(__dirname, "Preload.js"),
    },
  })

  ipcMain.handle("app:Hide", () => {
    if (win) {
      win.hide()
      if (!tray) createTray()
    }
  })

  ipcMain.handle("app:Show", () => {
    if (win) {
      win.show()
      win.focus()
    }
    if (tray) { tray.destroy(); tray = null }
  })

  if (app.isPackaged) {
    win.loadFile(join(__dirname, "Renderer", "index.html"))
  } else {
    win.loadURL("http://localhost:5173/")
  }

  win.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url)
    return { action: "deny" }
  })

  win.webContents.on("will-navigate", (event, url) => {
    if (url !== win?.webContents.getURL()) {
      event.preventDefault()
      shell.openExternal(url)
    }
  })
}

app.whenReady().then(() => {
  initLogger("DEBUG")
  console.log("[App] StepLauncher starting...")

  if (process.defaultApp) {
    if (process.argv.length >= 2) {
      app.setAsDefaultProtocolClient(
        "steplauncher",
        process.execPath,
        [process.argv[1]]
      )
    }
  } else {
    app.setAsDefaultProtocolClient("steplauncher")
  }

  RegisterAuthIpc()
  RegisterConfigIpc()
  RegisterDownloadsIpc()
  RegisterNovaCoreIpc()
  RegisterInstancesIpc()
  RegisterLangManagerIpc()
  RegisterThemesIpc()
  RegisterModsIpc()
  RegisterCacheIpc()

  try {
    const cfg = ConfigStore.GetConfig()
    if (cfg.launcher.discordRpc) startRPC()
  } catch {}

  CreateWindow()
})

app.on("open-url", (event, url) => {
  event.preventDefault()
  HandleOAuthCallback(url)
})

app.on("window-all-closed", () => {
  stopRPC()
  restoreConsole()
  if (process.platform !== "darwin") {
    app.quit()
  }
})

app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    CreateWindow()
  }
})
