import { ipcMain } from "electron"
import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, unlinkSync } from "fs"
import { createHash } from "crypto"
import { join } from "path"
import { homedir } from "os"

const CACHE_DIR = join(homedir(), ".StepLauncher", "bin", "cache")
const IMAGES_DIR = join(CACHE_DIR, "images")
const DEFAULT_TTL = 5 * 60 * 1000

function ensureDir(dir: string): void {
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true })
  }
}

function sanitizeKey(key: string): string {
  return key.replace(/[^a-zA-Z0-9_-]/g, "_")
}

function hashUrl(url: string): string {
  return createHash("md5").update(url).digest("hex")
}

function dataPath(key: string): string {
  return join(CACHE_DIR, `${sanitizeKey(key)}.json`)
}

interface CacheEntry {
  data: unknown
  expiry: number
}

function isExpired(entry: CacheEntry): boolean {
  return Date.now() > entry.expiry
}

function get(key: string): unknown | null {
  try {
    ensureDir(CACHE_DIR)
    const path = dataPath(key)
    if (!existsSync(path)) return null
    const raw = readFileSync(path, "utf-8")
    const entry: CacheEntry = JSON.parse(raw)
    if (isExpired(entry)) {
      try { unlinkSync(path) } catch {}
      return null
    }
    return entry.data
  } catch {
    return null
  }
}

function set(key: string, data: unknown, ttl: number = DEFAULT_TTL): void {
  try {
    ensureDir(CACHE_DIR)
    const entry: CacheEntry = { data, expiry: Date.now() + ttl }
    writeFileSync(dataPath(key), JSON.stringify(entry), "utf-8")
  } catch {}
}

function remove(key: string): void {
  try {
    const path = dataPath(key)
    if (existsSync(path)) unlinkSync(path)
  } catch {}
}

function clearAll(): void {
  try {
    for (const dir of [CACHE_DIR, IMAGES_DIR]) {
      if (!existsSync(dir)) continue
      const files = readdirSync(dir)
      for (const file of files) {
        try { unlinkSync(join(dir, file)) } catch {}
      }
    }
  } catch {}
}

async function downloadImage(url: string, destPath: string): Promise<void> {
  const res = await fetch(url, { signal: AbortSignal.timeout(15000) })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const buf = await res.arrayBuffer()
  writeFileSync(destPath, Buffer.from(buf))
}

function getMime(ext: string): string {
  const map: Record<string, string> = {
    ".png": "image/png",
    ".jpg": "image/jpeg",
    ".jpeg": "image/jpeg",
    ".webp": "image/webp",
    ".gif": "image/gif",
    ".svg": "image/svg+xml",
  }
  return map[ext.toLowerCase()] || "application/octet-stream"
}

async function getImage(url: string): Promise<string> {
  ensureDir(IMAGES_DIR)
  const hash = hashUrl(url)
  const ext = ".png"
  const filename = `${hash}${ext}`
  const filePath = join(IMAGES_DIR, filename)

  if (existsSync(filePath)) {
    const data = readFileSync(filePath)
    const b64 = data.toString("base64")
    return `data:${getMime(ext)};base64,${b64}`
  }

  await downloadImage(url, filePath)
  const data = readFileSync(filePath)
  const b64 = data.toString("base64")
  return `data:${getMime(ext)};base64,${b64}`
}

async function prefetchImage(url: string): Promise<void> {
  ensureDir(IMAGES_DIR)
  const filename = `${hashUrl(url)}.png`
  const filePath = join(IMAGES_DIR, filename)
  if (!existsSync(filePath)) {
    await downloadImage(url, filePath)
  }
}

export function RegisterCacheIpc() {
  ensureDir(CACHE_DIR)
  ensureDir(IMAGES_DIR)

  ipcMain.handle("cache:Get", async (_event, key: string) => {
    return get(key)
  })

  ipcMain.handle("cache:Set", async (_event, key: string, data: unknown, ttl?: number) => {
    set(key, data, ttl)
  })

  ipcMain.handle("cache:Remove", async (_event, key: string) => {
    remove(key)
  })

  ipcMain.handle("cache:Clear", async () => {
    clearAll()
  })

  ipcMain.handle("cache:Size", async () => {
    try {
      if (!existsSync(CACHE_DIR)) return 0
      const files = readdirSync(CACHE_DIR)
      let count = 0
      for (const file of files) {
        if (!file.endsWith(".json")) continue
        try {
          const raw = readFileSync(join(CACHE_DIR, file), "utf-8")
          const entry: CacheEntry = JSON.parse(raw)
          if (!isExpired(entry)) count++
        } catch { try { unlinkSync(join(CACHE_DIR, file)) } catch {} }
      }
      return count
    } catch { return 0 }
  })

  ipcMain.handle("cache:GetImage", async (_event, url: string) => {
    try {
      return await getImage(url)
    } catch {
      return ''
    }
  })

  ipcMain.handle("cache:PrefetchImage", async (_event, url: string) => {
    try {
      await prefetchImage(url)
    } catch {}
  })
}
