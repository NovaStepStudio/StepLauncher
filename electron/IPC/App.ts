import { ipcMain, shell, BrowserWindow, dialog, Notification, nativeImage } from "electron"
import fs from "fs"
import path from "path"
import { fileURLToPath } from "url"
import { dirname, join } from "path"

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

function getAppIconPath(): string {
  if (process.env.NODE_ENV === "development" || !process.resourcesPath) {
    return join(__dirname, "../../renderer/assets/icon.png")
  }
  return join(process.resourcesPath, "assets", "icon.png")
}

export function RegisterAppIpc() {
  ipcMain.handle("app:OpenExternal", async (_event, url: string) => {
    await shell.openExternal(url)
  })

  ipcMain.handle("app:OpenFileDialog", async (event, options: { filters?: { name: string; extensions: string[] }[] }) => {
    const win = BrowserWindow.fromWebContents(event.sender)
    if (!win) return { canceled: true, filePath: null }
    const result = await dialog.showOpenDialog(win, {
      properties: ["openFile"],
      filters: options.filters,
    })
    return { canceled: result.canceled, filePath: result.filePaths[0] ?? null }
  })

  ipcMain.handle("app:ReadDir", async (_event, dirPath: string) => {
    try {
      const entries = fs.readdirSync(dirPath, { withFileTypes: true })
      return entries.map(e => ({
        name: e.name,
        isDirectory: e.isDirectory(),
        isFile: e.isFile(),
        size: e.isFile() ? fs.statSync(path.join(dirPath, e.name)).size : 0,
        mtime: fs.statSync(path.join(dirPath, e.name)).mtime.toISOString(),
      }))
    } catch {
      return null
    }
  })

  ipcMain.handle("app:Minimize", (event) => {
    BrowserWindow.fromWebContents(event.sender)?.minimize()
  })

  ipcMain.handle("app:ShowNotification", (_event, options: { title: string; body: string }) => {
    new Notification({
      title: options.title,
      body: options.body,
      icon: nativeImage.createFromPath(getAppIconPath()).resize({ width: 64, height: 64 }),
    }).show()
  })

  ipcMain.handle("app:FetchJson", async (_event, url: string) => {
    const res = await fetch(url, { signal: AbortSignal.timeout(30000) })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    return res.json()
  })

  ipcMain.handle("app:Maximize", (event) => {
    const win = BrowserWindow.fromWebContents(event.sender)
    if (!win) return
    if (win.isMaximized()) win.unmaximize()
    else win.maximize()
  })

  ipcMain.handle("app:Close", (event) => {
    BrowserWindow.fromWebContents(event.sender)?.close()
  })

  ipcMain.handle("app:OpenPath", async (_event, dirPath: string) => {
    return shell.openPath(dirPath)
  })
}
