import { existsSync, mkdirSync, createWriteStream, createReadStream } from "fs"
import { readdir, rename, rm, chmod } from "fs/promises"
import { join } from "path"
import { homedir, platform } from "os"
import { get as httpsGet } from "https"
import { get as httpGet } from "http"
import { URL } from "url"
import * as tar from "tar"
import * as unzipper from "unzipper"

export type DownloadItem = "java" | "novacore" | "authlib"

export interface DownloadProgress {
  item: DownloadItem
  percent: number
  downloadedBytes: number
  totalBytes: number
  phase: "downloading" | "extracting" | "done" | "error"
  error?: string
}

export type ProgressCallback = (progress: DownloadProgress) => void

function getBaseDir(): string {
  const base = platform() === "win32"
    ? join(process.env.APPDATA || homedir(), ".StepLauncher")
    : join(homedir(), ".StepLauncher")
  return base
}

export function getJavaDir(): string {
  const dir = join(getBaseDir(), "bin", "java")
  mkdirSync(dir, { recursive: true })
  return dir
}

export function getLibsDir(): string {
  const dir = join(getBaseDir(), "bin", "libs")
  mkdirSync(dir, { recursive: true })
  return dir
}

export function getNovaCoreJarPath(): string {
  return join(getLibsDir(), "novacore-engine.jar")
}

export function getAuthlibInjectorPath(): string {
  return join(getLibsDir(), "authlib-injector.jar")
}

export function getJavaBinPath(): string {
  return join(getJavaDir(), "bin", platform() === "win32" ? "java.exe" : "java")
}

function httpGetJson<T>(urlStr: string): Promise<T> {
  return new Promise((resolve, reject) => {
    const u = new URL(urlStr)
    const mod = u.protocol === "https:" ? httpsGet : httpGet

    const req = mod(urlStr, {
      headers: { "User-Agent": "StepLauncher/1.0" },
      timeout: 15000,
    }, (res) => {
      const chunks: Buffer[] = []
      res.on("data", (c: Buffer) => chunks.push(c))
      res.on("end", () => {
        if (!res.statusCode || res.statusCode >= 400) {
          reject(new Error(`HTTP ${res.statusCode} from ${urlStr}`))
          return
        }
        try {
          resolve(JSON.parse(Buffer.concat(chunks).toString()))
        } catch (e) {
          reject(e)
        }
      })
    })
    req.on("error", reject)
    req.on("timeout", () => { req.destroy(); reject(new Error("Timeout")) })
    req.end()
  })
}

async function getLatestJavaRelease(): Promise<{
  version: string
  assets: { os: string; arch: string; archiveType: string; url: string }[]
}> {
  const releases: any[] = await httpGetJson(
    "https://api.adoptium.net/v3/assets/feature_releases/25/ga?image_type=jdk&sort_method=DEFAULT&sort_order=DESC&page_size=1"
  )

  const release = releases[0]
  if (!release) throw new Error("No Java 25 GA release found")

  const assets = release.binaries
    .filter((b: any) => b.image_type === "jdk")
    .map((b: any) => ({
      os: b.os,
      arch: b.architecture,
      archiveType: b.package.name.endsWith(".zip") ? "zip" : "tar.gz",
      url: b.package.link,
    }))

  if (assets.length === 0) throw new Error("No JDK binaries in Java 25 release")
  return { version: release.version_data.semver, assets }
}

async function getLatestGitHubRelease(owner: string, repo: string): Promise<{
  tag: string
  assets: { name: string; url: string }[]
}> {
  const data: any = await httpGetJson(
    `https://api.github.com/repos/${owner}/${repo}/releases/latest`
  )
  return {
    tag: data.tag_name,
    assets: data.assets.map((a: any) => ({
      name: a.name,
      url: a.browser_download_url,
    })),
  }
}

function matchJavaAsset(
  assets: { os: string; arch: string; archiveType: string; url: string }[]
): string {
  const sys = platform()
  const arch = process.arch

  const osKey = sys === "win32" ? "windows" : sys === "darwin" ? "mac" : "linux"
  const archKey = arch === "arm64" ? "aarch64" : "x64"

  const match = assets.find(a => a.os === osKey && a.arch === archKey)
  if (match) return match.url

  const fallback = assets.find(a => a.os === osKey)
  if (fallback) return fallback.url

  throw new Error(`No Java 25 JDK asset for ${sys}-${arch}`)
}

function downloadFile(
  urlStr: string, dest: string,
  onProgress: (downloaded: number, total: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const file = createWriteStream(dest)
    let downloaded = 0
    let total = 0

    function doRequest(url: string): void {
      const mod = url.startsWith("https") ? httpsGet : httpGet

      const req = mod(url, {
        headers: { "User-Agent": "StepLauncher/1.0" },
      }, (res) => {
        if (res.statusCode && res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          req.destroy()
          const redirectUrl = new URL(res.headers.location, url).href
          doRequest(redirectUrl)
          return
        }

        if (!res.statusCode || res.statusCode >= 400) {
          reject(new Error(`HTTP ${res.statusCode} downloading ${url}`))
          return
        }

        total = parseInt(res.headers["content-length"] || "0", 10)

        res.on("data", (chunk: Buffer) => {
          downloaded += chunk.length
          file.write(chunk)
          onProgress(downloaded, total)
        })

        res.on("end", () => {
          file.end()
        })
      })

      req.on("error", reject)
      req.end()
    }

    file.on("finish", () => resolve())
    file.on("error", reject)

    doRequest(urlStr)
  })
}

async function extractTarGz(src: string, dest: string): Promise<void> {
  await tar.extract({ file: src, cwd: dest })
}

async function extractZip(src: string, dest: string): Promise<void> {
  await new Promise<void>((resolve, reject) => {
    createReadStream(src)
      .pipe(unzipper.Extract({ path: dest }))
      .on("close", () => resolve())
      .on("error", reject)
  })
}

async function findJavaHome(extractDir: string): Promise<string> {
  const entries = await readdir(extractDir)
  const jdkDir = entries.find(e => e.startsWith("jdk-"))
  if (!jdkDir) throw new Error("Could not find JDK directory after extraction")
  return join(extractDir, jdkDir)
}

async function moveContents(src: string, dest: string): Promise<void> {
  const entries = await readdir(src)
  for (const entry of entries) {
    const s = join(src, entry)
    const d = join(dest, entry)
    if (existsSync(d)) await rm(d, { recursive: true, force: true })
    await rename(s, d)
  }
}

export async function checkJavaExists(): Promise<boolean> {
  return existsSync(getJavaBinPath())
}

export async function checkNovaCoreExists(): Promise<boolean> {
  return existsSync(getNovaCoreJarPath())
}

export async function checkAuthlibExists(): Promise<boolean> {
  return existsSync(getAuthlibInjectorPath())
}

export async function checkAllExist(): Promise<Record<DownloadItem, boolean>> {
  return {
    java: await checkJavaExists(),
    novacore: await checkNovaCoreExists(),
    authlib: await checkAuthlibExists(),
  }
}

export async function downloadJava(onProgress: ProgressCallback): Promise<void> {
  const baseDir = getBaseDir()
  const tmpDir = join(baseDir, ".tmp")
  mkdirSync(tmpDir, { recursive: true })

  try {
    onProgress({ item: "java", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "downloading" })

    const { assets } = await getLatestJavaRelease()
    const url = matchJavaAsset(assets)
    const archiveName = url.split("/").pop()!
    const archivePath = join(tmpDir, archiveName)

    await downloadFile(url, archivePath, (downloaded, total) => {
      const percent = total > 0 ? Math.round((downloaded / total) * 100) : 0
      onProgress({ item: "java", percent, downloadedBytes: downloaded, totalBytes: total, phase: "downloading" })
    })

    onProgress({ item: "java", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "extracting" })

    const extractDir = join(tmpDir, "jdk_extract")
    mkdirSync(extractDir, { recursive: true })

    if (archiveName.endsWith(".zip")) {
      await extractZip(archivePath, extractDir)
    } else {
      await extractTarGz(archivePath, extractDir)
    }

    const jdkHome = await findJavaHome(extractDir)
    await moveContents(jdkHome, getJavaDir())
    await rm(tmpDir, { recursive: true, force: true })

    const javaBin = getJavaBinPath()
    if (platform() !== "win32" && existsSync(javaBin)) {
      await chmod(javaBin, 0o755)
    }

    onProgress({ item: "java", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
  } catch (err) {
    await rm(tmpDir, { recursive: true, force: true }).catch(() => {})
    onProgress({ item: "java", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "error", error: String(err) })
    throw err
  }
}

export async function downloadNovaCore(onProgress: ProgressCallback): Promise<void> {
  try {
    onProgress({ item: "novacore", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "downloading" })

    const { assets } = await getLatestGitHubRelease("NovaStepStudio", "NovaCore-Engine")

    const jar = assets.find(a => a.name.endsWith(".jar") && !a.name.includes("javadoc") && !a.name.includes("sources"))
    if (!jar) throw new Error("No novacore-engine JAR in latest release")

    mkdirSync(getLibsDir(), { recursive: true })
    await downloadFile(jar.url, getNovaCoreJarPath(), (downloaded, total) => {
      const percent = total > 0 ? Math.round((downloaded / total) * 100) : 0
      onProgress({ item: "novacore", percent, downloadedBytes: downloaded, totalBytes: total, phase: "downloading" })
    })

    onProgress({ item: "novacore", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
  } catch (err) {
    onProgress({ item: "novacore", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "error", error: String(err) })
    throw err
  }
}

export async function downloadAuthlibInjector(onProgress: ProgressCallback): Promise<void> {
  try {
    onProgress({ item: "authlib", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "downloading" })

    const { assets } = await getLatestGitHubRelease("yushijinhun", "authlib-injector")

    const jar = assets.find(a => a.name.endsWith(".jar"))
    if (!jar) throw new Error("No authlib-injector JAR in latest release")

    mkdirSync(getLibsDir(), { recursive: true })
    await downloadFile(jar.url, getAuthlibInjectorPath(), (downloaded, total) => {
      const percent = total > 0 ? Math.round((downloaded / total) * 100) : 0
      onProgress({ item: "authlib", percent, downloadedBytes: downloaded, totalBytes: total, phase: "downloading" })
    })

    onProgress({ item: "authlib", percent: 100, downloadedBytes: 0, totalBytes: 0, phase: "done" })
  } catch (err) {
    onProgress({ item: "authlib", percent: 0, downloadedBytes: 0, totalBytes: 0, phase: "error", error: String(err) })
    throw err
  }
}
