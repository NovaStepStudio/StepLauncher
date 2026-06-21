import { ipcMain, BrowserWindow } from "electron"
import * as DL from "../Services/DownloadManager.js"

export function RegisterDownloadsIpc() {
  ipcMain.handle("downloads:CheckAll", async () => {
    return DL.checkAllExist()
  })

  ipcMain.handle("downloads:StartAll", async (event) => {
    const win = BrowserWindow.fromWebContents(event.sender)
    if (!win) return { success: false, item: "unknown", error: "No window" }

    const sendProgress = (progress: DL.DownloadProgress) => {
      if (win.isDestroyed() || win.webContents.isDestroyed()) return
      try {
        win.webContents.send("downloads:Progress", progress)
      } catch {
        // Renderer may die between isDestroyed check and send
      }
    }

    const status = await DL.checkAllExist()

    if (status.java) {
      sendProgress({ item: "java", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
    } else {
      try {
        await DL.downloadJava(sendProgress)
      } catch (e: any) {
        return { success: false, item: "java" as const, error: e?.message ?? String(e) }
      }
    }

    if (status.novacore) {
      sendProgress({ item: "novacore", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
    } else {
      try {
        await DL.downloadNovaCore(sendProgress)
      } catch (e: any) {
        return { success: false, item: "novacore" as const, error: e?.message ?? String(e) }
      }
    }

    if (status.authlib) {
      sendProgress({ item: "authlib", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
    } else {
      try {
        await DL.downloadAuthlibInjector(sendProgress)
      } catch (e: any) {
        return { success: false, item: "authlib" as const, error: e?.message ?? String(e) }
      }
    }

    return { success: true }
  })
}
