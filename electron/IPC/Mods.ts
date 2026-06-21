import { ipcMain, dialog } from "electron"
import { writeFileSync, mkdirSync, readdirSync, readFileSync, existsSync, cpSync, rmSync, createReadStream, createWriteStream } from "fs"
import { extname, dirname, join } from "path"
import { homedir, platform } from "os"
import * as unzipper from "unzipper"

function getBaseDir(): string {
  return platform() === "win32"
    ? join(process.env.APPDATA || homedir(), ".StepLauncher")
    : join(homedir(), ".StepLauncher")
}

function getGameDir(): string {
  const dir = join(getBaseDir(), "game")
  mkdirSync(dir, { recursive: true })
  return dir
}

export function RegisterModsIpc() {
  ipcMain.handle("mods:GetGameDir", async () => {
    return getGameDir()
  })

  ipcMain.handle("mods:GetBaseDir", async () => {
    return getBaseDir()
  })

  ipcMain.handle("mods:DownloadFile", async (_event, url: string, suggestedName: string) => {
    try {
      const ext = extname(suggestedName) || ".jar"
      const result = await dialog.showSaveDialog({
        defaultPath: suggestedName,
        filters: [
          { name: "Archivo", extensions: [ext.replace(".", "")] },
          { name: "Todos los archivos", extensions: ["*"] },
        ],
      })
      if (result.canceled || !result.filePath) {
        return { success: false, canceled: true }
      }

      const response = await fetch(url)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const buffer = Buffer.from(await response.arrayBuffer())
      writeFileSync(result.filePath, buffer)

      return { success: true, filePath: result.filePath }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Download failed" }
    }
  })

  ipcMain.handle("mods:DownloadToPath", async (_event, url: string, destPath: string) => {
    try {
      mkdirSync(dirname(destPath), { recursive: true })
      const response = await fetch(url)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const buffer = Buffer.from(await response.arrayBuffer())
      writeFileSync(destPath, buffer)
      return { success: true, filePath: destPath }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Download failed" }
    }
  })

  ipcMain.handle("mods:DownloadToGameDir", async (_event, contentType: string, url: string, filename: string, worldName?: string) => {
    try {
      const gameDir = getGameDir()
      let destDir: string

      switch (contentType) {
        case "shader":
          destDir = join(gameDir, "shaderpacks")
          break
        case "resourcepack":
          destDir = join(gameDir, "resourcepacks")
          break
        case "datapack":
          if (worldName) {
            destDir = join(gameDir, "saves", worldName, "datapacks")
          } else {
            destDir = join(gameDir, "datapacks")
          }
          break
        case "mod":
        default:
          destDir = join(gameDir, "mods")
          break
      }

      const destPath = join(destDir, filename)
      mkdirSync(destDir, { recursive: true })

      const response = await fetch(url)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const buffer = Buffer.from(await response.arrayBuffer())
      writeFileSync(destPath, buffer)

      return { success: true, filePath: destPath }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Download failed" }
    }
  })

  ipcMain.handle("mods:GetWorlds", async () => {
    try {
      const savesDir = join(getGameDir(), "saves")
      if (!existsSync(savesDir)) {
        mkdirSync(savesDir, { recursive: true })
        return { success: true, worlds: [] }
      }
      const worlds = readdirSync(savesDir, { withFileTypes: true })
        .filter(d => d.isDirectory())
        .map(d => d.name)
      return { success: true, worlds }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Failed to list worlds" }
    }
  })

  ipcMain.handle("mods:InstallModpack", async (_event, url: string, _filename: string) => {
    try {
      const gameDir = getGameDir()
      const tmpDir = join(getBaseDir(), "tmp", "modpack_" + Date.now())
      const mrpackPath = join(tmpDir, "pack.mrpack")

      mkdirSync(tmpDir, { recursive: true })

      const response = await fetch(url)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const buffer = Buffer.from(await response.arrayBuffer())
      writeFileSync(mrpackPath, buffer)

      await new Promise<void>((resolve, reject) => {
        const readStream = createReadStream(mrpackPath)
        const parseStream = unzipper.Parse()
        readStream.pipe(parseStream)

        parseStream.on("entry", (entry: any) => {
          const entryPath = join(tmpDir, entry.path)
          if (entry.type === "Directory") {
            mkdirSync(entryPath, { recursive: true })
            entry.autodrain()
          } else {
            mkdirSync(dirname(entryPath), { recursive: true })
            entry.pipe(createWriteStream(entryPath))
          }
        })
        parseStream.on("close", resolve)
        parseStream.on("error", reject)
      })

      const indexPath = join(tmpDir, "modrinth.index.json")
      if (!existsSync(indexPath)) {
        rmSync(tmpDir, { recursive: true, force: true })
        return { success: false, error: "Invalid modpack: missing modrinth.index.json" }
      }

      const index = JSON.parse(readFileSync(indexPath, "utf-8"))

      for (const file of index.files || []) {
        const downloadUrl = file.downloads?.[0]
        if (!downloadUrl) continue

        const filePath = join(gameDir, file.path)
        mkdirSync(dirname(filePath), { recursive: true })

        try {
          const dlRes = await fetch(downloadUrl)
          if (dlRes.ok) {
            const dlBuf = Buffer.from(await dlRes.arrayBuffer())
            writeFileSync(filePath, dlBuf)
          }
        } catch {}
      }

      const overrideDir = join(tmpDir, "override")
      if (existsSync(overrideDir)) {
        const copyRecursive = (src: string, dest: string) => {
          if (!existsSync(src)) return
          const entries = readdirSync(src, { withFileTypes: true })
          for (const entry of entries) {
            const srcPath = join(src, entry.name)
            const destPath = join(dest, entry.name)
            if (entry.isDirectory()) {
              mkdirSync(destPath, { recursive: true })
              copyRecursive(srcPath, destPath)
            } else {
              mkdirSync(dirname(destPath), { recursive: true })
              cpSync(srcPath, destPath)
            }
          }
        }
        copyRecursive(overrideDir, gameDir)
      }

      rmSync(tmpDir, { recursive: true, force: true })

      return { success: true, gameDir }
    } catch (e: any) {
      return { success: false, error: e?.message ?? "Modpack installation failed" }
    }
  })
}
