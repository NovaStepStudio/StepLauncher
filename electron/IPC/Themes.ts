import { ipcMain, dialog } from "electron"
import * as ThemeStore from "../Stores/Themes.js"

export function RegisterThemesIpc() {
  ipcMain.handle("theme:Export", (_event, metadata: { name: string; author: string; homePage: string; version: string }) => {
    try {
      const info = ThemeStore.ExportTheme(metadata)
      return { success: true, theme: info }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error exporting theme" }
    }
  })

  ipcMain.handle("theme:Import", async () => {
    const result = await dialog.showOpenDialog({
      properties: ["openDirectory"],
      title: "Selecciona la carpeta del tema",
    })
    if (result.canceled || !result.filePaths.length) {
      return { success: false, canceled: true }
    }
    try {
      const info = ThemeStore.InstallTheme(result.filePaths[0])
      if (!info) return { success: false, error: "No se encontró theme.json en la carpeta" }
      return { success: true, theme: info }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error importing theme" }
    }
  })

  ipcMain.handle("theme:List", () => {
    try {
      const themes = ThemeStore.GetInstalledThemes()
      return { success: true, themes }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error listing themes" }
    }
  })

  ipcMain.handle("theme:Delete", (_event, name: string) => {
    try {
      const ok = ThemeStore.DeleteTheme(name)
      return { success: ok }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error deleting theme" }
    }
  })

  ipcMain.handle("theme:GetConfig", (_event, name: string) => {
    try {
      const config = ThemeStore.GetThemeConfig(name)
      if (!config) return { success: false, error: "Theme not found" }
      return { success: true, config }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error loading theme" }
    }
  })

  ipcMain.handle("theme:GetGallery", (_event, name: string) => {
    try {
      const gallery = ThemeStore.GetThemeGallery(name)
      return { success: true, gallery }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Error loading gallery", gallery: [] }
    }
  })
}
