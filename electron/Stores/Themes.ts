import { readFileSync, writeFileSync, existsSync, mkdirSync, copyFileSync, readdirSync, statSync, renameSync, unlinkSync } from "fs"
import { join, extname } from "path"
import { homedir } from "os"
import { app } from "electron"
import type { ThemeMetadata, ThemeInfo } from "../Types/Theme.js"
import type { PersonalizationConfig } from "../Types/Config.js"
import { GetConfig } from "./Config.js"

function getBaseDir(): string {
  if (process.platform === "win32") {
    return join(process.env.APPDATA || homedir(), ".StepLauncher")
  }
  return join(homedir(), ".StepLauncher")
}

function getThemesDir(): string {
  const dir = join(getBaseDir(), "themes")
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  return dir
}

function getThemeDir(name: string): string {
  return join(getThemesDir(), name)
}

function resolveAssetPath(assetUrl: string): string | null {
  let path = assetUrl.replace(/^url\(['"]?/, "").replace(/['"]?\)$/, "").trim()

  if (path.startsWith("file:///")) {
    const filePath = process.platform === "win32" ? path.replace(/^file:\/\/\//, "") : path.replace(/^file:\/\//, "")
    if (existsSync(filePath)) return filePath
    return path
  }

  const appRoot = join(app.getAppPath(), "..")
  const bases = [
    join(appRoot, "renderer"),
    join(appRoot, "dist", "Renderer"),
  ]
  for (const base of bases) {
    const fullPath = join(base, path)
    if (existsSync(fullPath)) return fullPath
  }
  return null
}

function copyToTheme(srcEntry: string, themeDir: string): string | null {
  const srcPath = resolveAssetPath(srcEntry)
  if (!srcPath || !existsSync(srcPath)) return null
  const ext = extname(srcPath) || ".png"
  const name = "asset_" + Date.now() + ext
  copyFileSync(srcPath, join(themeDir, name))
  return name
}

export function ExportTheme(metadata: { name: string; author: string; homePage: string; version: string }): ThemeInfo {
  const config = GetConfig()
  const themeDir = getThemeDir(metadata.name)
  if (existsSync(themeDir)) {
    const old = join(themeDir, "theme.json.bak")
    if (existsSync(themeDir)) renameSync(join(themeDir, "theme.json"), old)
  } else {
    mkdirSync(themeDir, { recursive: true })
  }

  const bgUrl = config.personalization.appBackground
  const bgName = bgUrl ? copyToTheme(bgUrl, themeDir) : null

  const savedConfig = { ...config.personalization } as unknown as Record<string, unknown>
  if (bgName) {
    savedConfig.appBackground = "url('./" + bgName + "')"
  }

  const themeConfig: ThemeMetadata = {
    name: metadata.name,
    author: metadata.author,
    homePage: metadata.homePage,
    version: metadata.version,
    thumbnail: bgName || "",
    gallery: [],
    config: savedConfig,
  }

  writeFileSync(join(themeDir, "theme.json"), JSON.stringify(themeConfig, null, 2), "utf-8")

  return {
    name: metadata.name,
    author: metadata.author,
    homePage: metadata.homePage,
    version: metadata.version,
    thumbnail: bgName ? join(themeDir, bgName) : "",
    path: themeDir,
  }
}

export function InstallTheme(sourcePath: string): ThemeInfo | null {
  const themeJsonPath = join(sourcePath, "theme.json")
  if (!existsSync(themeJsonPath)) return null

  let raw: string
  try {
    raw = readFileSync(themeJsonPath, "utf-8")
  } catch {
    return null
  }

  let theme: ThemeMetadata
  try {
    theme = JSON.parse(raw) as ThemeMetadata
  } catch {
    return null
  }

  if (!theme.name) return null

  const themeDir = getThemeDir(theme.name)
  if (!existsSync(themeDir)) mkdirSync(themeDir, { recursive: true })

  const files = readdirSync(sourcePath)
  for (const f of files) {
    const src = join(sourcePath, f)
    const dst = join(themeDir, f)
    if (statSync(src).isFile()) {
      copyFileSync(src, dst)
    }
  }

  const thumbPath = theme.thumbnail ? join(themeDir, theme.thumbnail) : ""
  return {
    name: theme.name,
    author: theme.author || "",
    homePage: theme.homePage || "",
    version: theme.version || "",
    thumbnail: existsSync(thumbPath) ? thumbPath : "",
    path: themeDir,
  }
}

export function GetInstalledThemes(): ThemeInfo[] {
  const dir = getThemesDir()
  if (!existsSync(dir)) return []

  return readdirSync(dir)
    .map((name) => {
      const themeDir = join(dir, name)
      if (!statSync(themeDir).isDirectory()) return null
      const jsonPath = join(themeDir, "theme.json")
      if (!existsSync(jsonPath)) return null
      try {
        const raw = readFileSync(jsonPath, "utf-8")
        const meta = JSON.parse(raw) as ThemeMetadata
        const thumbPath = meta.thumbnail ? join(themeDir, meta.thumbnail) : ""
        return {
          name: meta.name || name,
          author: meta.author || "",
          homePage: meta.homePage || "",
          version: meta.version || "",
          thumbnail: existsSync(thumbPath) ? thumbPath : "",
          path: themeDir,
        }
      } catch {
        return null
      }
    })
    .filter((t): t is ThemeInfo => t !== null)
}

export function DeleteTheme(name: string): boolean {
  const themeDir = getThemeDir(name)
  if (!existsSync(themeDir)) return false
  try {
    const files = readdirSync(themeDir)
    for (const f of files) unlinkSync(join(themeDir, f))
    try { unlinkSync(join(themeDir, ".gitkeep")) } catch {}
    try { rmdirSync(themeDir) } catch { renameSync(themeDir, themeDir + ".deleted") }
    return true
  } catch {
    return false
  }
}

function rmdirSync(path: string) {
  if (process.platform === "win32") {
    try { execSync(`rmdir /s /q "${path}"`) } catch {}
  } else {
    try { execSync(`rm -rf "${path}"`) } catch {}
  }
}

import { execSync } from "child_process"

export function GetThemeConfig(name: string): PersonalizationConfig | null {
  const themeDir = getThemeDir(name)
  const jsonPath = join(themeDir, "theme.json")
  if (!existsSync(jsonPath)) return null
  try {
    const raw = readFileSync(jsonPath, "utf-8")
    const meta = JSON.parse(raw) as ThemeMetadata
    const config = meta.config as unknown as PersonalizationConfig
    if (config.appBackground && config.appBackground.includes("./")) {
      const fileName = config.appBackground.replace(/^url\(['"]?\.\//, "").replace(/['"]?\)$/, "").trim()
      const fullPath = join(themeDir, fileName)
      if (existsSync(fullPath)) {
        config.appBackground = "url('file:///" + fullPath.replace(/\\/g, "/") + "')"
      }
    }
    return config
  } catch {
    return null
  }
}

export function GetThemeGallery(name: string): string[] {
  const themeDir = getThemeDir(name)
  const jsonPath = join(themeDir, "theme.json")
  if (!existsSync(jsonPath)) return []
  try {
    const raw = readFileSync(jsonPath, "utf-8")
    const meta = JSON.parse(raw) as ThemeMetadata
    if (!meta.gallery || !Array.isArray(meta.gallery)) return []
    return meta.gallery
      .map((g: string) => join(themeDir, g))
      .filter((p: string) => existsSync(p))
  } catch {
    return []
  }
}
