import { spawn, execSync, ChildProcess } from "child_process"
import { existsSync, mkdirSync, readFileSync } from "fs"
import { request as httpRequest } from "http"
import { URL } from "url"
import { join, dirname } from "path"
import { homedir, platform } from "os"
import { fileURLToPath } from "url"
import { getJavaBinPath, getNovaCoreJarPath, getAuthlibInjectorPath } from "./DownloadManager.js"

const MANAGER_DIR = dirname(fileURLToPath(import.meta.url))

type EngineStatus = "stopped" | "starting" | "running" | "error"

export type EventCallback = (event: string, data: unknown) => void

let instance: NovaCoreManagerService | null = null

export function getNovaCoreManager(): NovaCoreManagerService {
    if (!instance) instance = new NovaCoreManagerService()
    return instance
}

function getBaseDir(): string {
    return platform() === "win32"
        ? join(process.env.APPDATA || homedir(), ".StepLauncher")
        : join(homedir(), ".StepLauncher")
}

function getInstancesDir(): string {
    const dir = join(getBaseDir(), "instances")
    mkdirSync(dir, { recursive: true })
    mkdirSync(join(dir, "shared"), { recursive: true })
    return dir
}

function getVersionsDir(): string {
    const dir = join(getBaseDir(), "versions")
    mkdirSync(dir, { recursive: true })
    return dir
}

function getLogDir(): string {
    const dir = join(getBaseDir(), "logs")
    mkdirSync(dir, { recursive: true })
    return dir
}

export class NovaCoreManagerService {
    private proc: ChildProcess | null = null
    private token = ""
    private httpPort = 17878
    private wsPort = 17879
    private ws: WebSocket | null = null
    private status: EngineStatus = "stopped"
    private errorMessage = ""
    private listeners = new Set<EventCallback>()
    private healthTimer: ReturnType<typeof setInterval> | null = null
    private wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
    private wsDead = false
    private restartCount = 0
    private maxRestarts = 5
    private restartTimer: ReturnType<typeof setTimeout> | null = null
    private intentionalStop = false
    private reportedCrashes = new Set<string>()
    private forwardingRegistered = false
    private unreachableEmitted = false
    private heartbeatTimer: ReturnType<typeof setInterval> | null = null
    private lastPong = 0

    private httpReq<T>(method: string, path: string, body?: unknown): Promise<T> {
        return new Promise((resolve, reject) => {
            const url = new URL(`http://localhost:${this.httpPort}${path}`)
            const bodyStr = body !== undefined ? JSON.stringify(body) : null

            const req = httpRequest(url, {
                method,
                headers: {
                    "Content-Type": "application/json",
                    "X-Access-Token": this.token,
                    ...(bodyStr ? { "Content-Length": Buffer.byteLength(bodyStr).toString() } : {}),
                },
                timeout: 30000,
            }, (res) => {
                const chunks: Buffer[] = []
                res.on("data", (c: Buffer) => chunks.push(c))
                res.on("end", () => {
                    let data: any = null
                    const raw = Buffer.concat(chunks).toString()
                    if (raw) { try { data = JSON.parse(raw) } catch {} }

                    if (!res.statusCode || res.statusCode >= 400) {
                        const errMsg = data?.error ?? `HTTP ${res.statusCode}`
                        reject(new Error(errMsg))
                        return
                    }
                    resolve(data as T)
                })
            })

            req.on("error", reject)
            req.on("timeout", () => { req.destroy(); reject(new Error("Request timeout")) })
            if (bodyStr) req.write(bodyStr)
            req.end()
        })
    }

    addListener(cb: EventCallback): () => void {
        this.listeners.add(cb)
        return () => { this.listeners.delete(cb) }
    }

    private emit(event: string, data: unknown) {
        const cbs = [...this.listeners]
        for (const cb of cbs) {
            try { cb(event, data) } catch {}
        }
    }

    private freePort(port: number) {
        try {
            execSync(`fuser -k ${port}/tcp 2>/dev/null`, { stdio: "ignore", timeout: 3000 })
        } catch {}
    }

    async start(): Promise<void> {
        if (this.status === "running") return

        this.freePort(this.httpPort)
        this.freePort(this.wsPort)

        this.status = "starting"

        const javaPath = getJavaBinPath()
        const jarPath = getNovaCoreJarPath()

        if (!existsSync(javaPath)) throw new Error(`Java no encontrado en ${javaPath}`)
        if (!existsSync(jarPath)) throw new Error(`NovaCore JAR no encontrado en ${jarPath}`)

        const instancesDir = getInstancesDir()
        const logDir = getLogDir()

        const args = [
            "-jar", jarPath,
            "--port", String(this.httpPort),
            "--ws-port", String(this.wsPort),
            "--threads", "4",
            "--instances-dir", instancesDir,
            "--log-dir", logDir,
            "--launcher-name", "StepLauncher",
            "--log-level", "INFO",
        ]

        this.proc = spawn(javaPath, args, {
            stdio: ["ignore", "pipe", "pipe"],
            detached: platform() === "linux" || platform() === "darwin",
        })

        return new Promise<void>((resolve, reject) => {
            const proc = this.proc!
            let token = ""
            let ready = false
            let buf = ""
            let errBuf = ""
            let settled = false

            const timeout = setTimeout(() => {
                cleanup()
                if (!settled) {
                    settled = true
                    this.status = "error"
                    this.errorMessage = "Engine did not start within 45s"
                    reject(new Error(this.errorMessage))
                }
            }, 45000)

            const forwardToWindows = () => {
                if (this.forwardingRegistered) return
                this.forwardingRegistered = true
                import("electron").then(({ BrowserWindow }) => {
                    this.addListener((event, data) => {
                        const wins = BrowserWindow.getAllWindows()
                        for (let i = 0; i < wins.length; i++) {
                            const w = wins[i]
                            if (w.isDestroyed() || w.webContents.isDestroyed()) continue
                            try {
                                w.webContents.send("novacore:Event", { event, data })
                            } catch {
                                // webContents.send can throw write EIO if renderer dies
                                // between isDestroyed check and send — swallow silently
                            }
                        }
                    })
                })
            }

            const cleanup = () => {
                clearTimeout(timeout)
                if (proc.stdout) proc.stdout.removeListener("data", onStdout)
                if (proc.stderr) proc.stderr.removeListener("data", onStderr)
                proc.removeListener("error", onError)
                proc.removeListener("exit", onExit)
            }

            const onStdout = (chunk: Buffer) => {
                buf += chunk.toString("utf8")
                const lines = buf.split("\n")
                buf = lines.pop() ?? ""

                for (const line of lines) {
                    const trimmed = line.trim()
                    if (!token && trimmed.startsWith("TOKEN:")) {
                        token = trimmed.slice(6).trim()
                    }
                    if (!ready && trimmed.includes("initialization completed successfully") && token) {
                        ready = true
                        cleanup()
                        this.token = token
                        this.restartCount = 0
                        this.status = "running"
                        this.connectWs()
                        this.startHealthCheck()
                        forwardToWindows()

                        const logForwarder = (line: string) => {
                            const clean = line.replace(/\x1b\[[0-9;]*m/g, '')
                            this.emit("engine_log", { line: clean })

                            let launchId = ''
                            let exitCode = 0
                            let matched = false

                            const crashMatch = clean.match(/CRASH exitCode=(\d+)/)
                            if (crashMatch) {
                                exitCode = parseInt(crashMatch[1], 10)
                                const lm = clean.match(/\[launch-([^\]]+)\]/)
                                if (lm) launchId = `launch-${lm[1]}`
                                matched = true
                            }

                            const exitMatch = clean.match(/Exited:\s*(-?\d+)/)
                            if (!matched && exitMatch) {
                                exitCode = parseInt(exitMatch[1], 10)
                                if (exitCode !== 0) {
                                    const lm = clean.match(/\[launch-([^\]]+)\]/)
                                    if (lm) launchId = `launch-${lm[1]}`
                                    matched = true
                                }
                            }

                            if (matched) {
                                if (this.reportedCrashes.has(launchId)) return
                                if (launchId) {
                                    this.reportedCrashes.add(launchId)
                                    setTimeout(() => this.reportedCrashes.delete(launchId), 10000)
                                }
                                this.emit("game_crash", {
                                    launchId,
                                    exitCode,
                                    reason: `Game crashed with exit code ${exitCode}`,
                                    source: "engine_stdout",
                                    timestamp: Date.now(),
                                })
                            }
                        }
                        if (proc.stdout) {
                            proc.stdout.on("data", (c: Buffer) => {
                                for (const l of c.toString("utf8").split("\n").filter(Boolean)) logForwarder(l)
                            })
                        }
                        if (proc.stderr) {
                            proc.stderr.on("data", (c: Buffer) => {
                                for (const l of c.toString("utf8").split("\n").filter(Boolean)) logForwarder(l)
                            })
                        }
                        if (!settled) {
                            settled = true
                            resolve()
                        }

                        proc.once("exit", (code) => this.onEngineExit(code, proc))
                    }
                }
            }

            const onStderr = (chunk: Buffer) => {
                errBuf += chunk.toString("utf8")
            }

            const onError = (err: Error) => {
                cleanup()
                this.status = "error"
                this.errorMessage = err.message
                if (!settled) {
                    settled = true
                    reject(err)
                }
            }

            const onExit = (code: number | null) => {
                cleanup()
                const stderrTail = errBuf.trim().split("\n").slice(-3).join(" | ")
                const detail = stderrTail ? ` — ${stderrTail}` : ""
                this.errorMessage = `Engine process exited with code ${code}${detail}`
                this.status = "error"
                if (!settled) {
                    settled = true
                    reject(new Error(this.errorMessage))
                }
            }

            if (proc.stdout) proc.stdout.on("data", onStdout)
            if (proc.stderr) proc.stderr.on("data", onStderr)
            proc.once("error", onError)
            proc.once("exit", onExit)
        })
    }

    async stop(): Promise<void> {
        this.intentionalStop = true
        this.cancelRestart()
        this.stopHealthCheck()
        this.stopHeartbeat()
        this.wsDead = true
        if (this.wsReconnectTimer) {
            clearTimeout(this.wsReconnectTimer)
            this.wsReconnectTimer = null
        }
        this.ws?.close(1000, "bye")
        this.ws = null
        this.unreachableEmitted = false

        try {
            await this.httpReq("POST", "/close")
        } catch {}

        if (this.proc && !this.proc.killed) {
            await new Promise<void>((resolve) => {
                const timeout = setTimeout(() => {
                    this.forceKill()
                    resolve()
                }, 3000)
                this.proc!.once("exit", () => {
                    clearTimeout(timeout)
                    resolve()
                })
                this.proc!.kill("SIGTERM")
            })
        }

        this.proc = null
        this.token = ""
        this.status = "stopped"
    }

    private forceKill() {
        if (!this.proc || this.proc.killed) return
        const pid = this.proc.pid
        if (pid) {
            try {
                if (platform() === "win32") {
                    const { execSync } = require("child_process")
                    execSync(`taskkill /F /T /PID ${pid}`, { stdio: "ignore" })
                } else {
                    try { process.kill(-pid, "SIGKILL") } catch { this.proc.kill("SIGKILL") }
                }
            } catch { this.proc.kill("SIGKILL") }
        }
        this.proc = null
    }

    getStatus(): { status: EngineStatus; error?: string } {
        return { status: this.status, error: this.errorMessage || undefined }
    }

    private cancelRestart() {
        if (this.restartTimer) {
            clearTimeout(this.restartTimer)
            this.restartTimer = null
        }
    }

    private async onEngineExit(code: number | null, proc: ChildProcess) {
        if (this.intentionalStop || this.proc !== proc) return

        this.status = "error"
        this.errorMessage = `Engine exited with code ${code}`
        this.emitUnreachable("process_exit", { code })

        this.restartCount++
        if (this.restartCount > this.maxRestarts) {
            this.emit("engine_log", { line: `[engine] Reinicio automático agotado (${this.maxRestarts} intentos)` })
            return
        }

        const delay = Math.min(this.restartCount * 2000, 10000)
        this.emit("engine_log", { line: `[engine] Reiniciando en ${delay / 1000}s (intento ${this.restartCount}/${this.maxRestarts})...` })

        this.restartTimer = setTimeout(async () => {
            this.restartTimer = null
            try {
                await this.start()
                this.emit("engine_log", { line: "[engine] Motor reiniciado correctamente" })
            } catch (e: any) {
                this.emit("engine_log", { line: `[engine] Error al reiniciar: ${e.message}` })
            }
        }, delay)
    }

    private async ensureRunning(): Promise<void> {
        if (this.status === "running") return
        if (this.status === "starting") {
            await new Promise<void>((resolve, reject) => {
                const check = setInterval(() => {
                    if (this.status === "running") { clearInterval(check); resolve() }
                    if (this.status === "error") { clearInterval(check); reject(new Error(this.errorMessage)) }
                }, 200)
                setTimeout(() => { clearInterval(check); reject(new Error("Engine start timeout")) }, 50000)
            })
            return
        }
        await this.start()
    }
    // forwarding now registered once in start() via forwardToWindows() with forwardingRegistered guard

    private emitUnreachable(reason: string, extra: Record<string, unknown> = {}) {
        if (this.unreachableEmitted) return
        this.unreachableEmitted = true
        this.emit("engine_unreachable", { reason, timestamp: Date.now(), ...extra })
        setTimeout(() => { this.unreachableEmitted = false }, 2000)
    }

    private startHeartbeat() {
        this.stopHeartbeat()
        this.lastPong = Date.now()
        this.heartbeatTimer = setInterval(() => {
            if (this.ws?.readyState !== WebSocket.OPEN) {
                this.stopHeartbeat()
                return
            }
            try {
                this.ws!.send(JSON.stringify({ event: "ping", data: { timestamp: Date.now() } }))
            } catch {}
            if (Date.now() - this.lastPong > 60000) {
                this.ws?.close()
                this.emitUnreachable("ws_timeout", {})
                this.scheduleWsReconnect()
            }
        }, 15000)
    }

    private stopHeartbeat() {
        if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer)
            this.heartbeatTimer = null
        }
    }

    private connectWs() {
        if (this.wsDead) return
        if (this.ws?.readyState === WebSocket.OPEN) return

        try {
            this.ws = new WebSocket(`ws://localhost:${this.wsPort}?token=${encodeURIComponent(this.token)}`)
        } catch {
            this.scheduleWsReconnect()
            return
        }

        const timeout = setTimeout(() => {
            if (this.ws?.readyState !== WebSocket.OPEN) {
                this.ws?.close()
                this.scheduleWsReconnect()
            }
        }, 15000)

        this.ws.onopen = () => {
            clearTimeout(timeout)
            this.startHeartbeat()
        }

        this.ws.onmessage = (e: MessageEvent) => {
            try {
                const raw = e.data as string
                const msg = JSON.parse(raw)
                if (msg.event === "pong") {
                    this.lastPong = Date.now()
                    return
                }
                this.emit(msg.event, msg.data)
            } catch {}
        }

        this.ws.onclose = (e: CloseEvent) => {
            clearTimeout(timeout)
            this.stopHeartbeat()
            if (e.code === 1008) {
                this.wsDead = true
                return
            }
            if (!this.wsDead) {
                this.emitUnreachable("ws_disconnected", { code: e.code })
                this.scheduleWsReconnect()
            }
        }

        this.ws.onerror = () => {
            clearTimeout(timeout)
        }
    }

    private scheduleWsReconnect() {
        if (this.wsDead || this.wsReconnectTimer) return
        this.wsReconnectTimer = setTimeout(() => {
            this.wsReconnectTimer = null
            if (!this.wsDead && this.status === "running") {
                this.connectWs()
            }
        }, 1500)
    }

    private startHealthCheck() {
        this.stopHealthCheck()
        let fails = 0
        this.healthTimer = setInterval(async () => {
            try {
                await this.httpReq("GET", "/health")
                fails = 0
            } catch {
                fails++
                if (fails >= 2) {
                    this.emitUnreachable("health_check_failed", { fails })
                }
            }
        }, 10000)
    }

    private stopHealthCheck() {
        if (this.healthTimer) {
            clearInterval(this.healthTimer)
            this.healthTimer = null
        }
    }

    async health(): Promise<{ status: string; version: string; uptime: number }> {
        await this.ensureRunning()
        return this.httpReq("GET", "/health")
    }

    async getVersions(limit = 50): Promise<any[]> {
        await this.ensureRunning()
        return this.httpReq("GET", `/versions?limit=${limit}`)
    }

    async install(req: any): Promise<string> {
        await this.ensureRunning()
        return this.httpReq("POST", "/install", req).then((r: any) => r.sessionId as string)
    }

    async launch(req: any): Promise<string> {
        await this.ensureRunning()
        return this.httpReq("POST", "/launch", req).then((r: any) => r.launchId as string)
    }

    async killInstance(id: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", `/launch/kill/${id}`)
    }

    async getEngineInfo(): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", "/system/resources")
    }

    async getModLoaders(): Promise<{ loaders: string[] }> {
        await this.ensureRunning()
        return this.httpReq("GET", "/modloaders")
    }

    async getModLoaderVersions(loader: string, mcVersion: string): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", `/modloaders/versions/${loader}/${mcVersion}`)
    }

    async getProgress(sessionId: string): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", `/progress?sessionId=${sessionId}`)
    }

    async getGlobalProgress(): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", "/progress/global")
    }

    async pauseInstall(sessionId: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", `/install/pause/${sessionId}`)
    }

    async resumeInstall(sessionId: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", `/install/resume/${sessionId}`)
    }

    async cancelInstall(sessionId: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", `/install/cancel/${sessionId}`)
    }

    async getRunningInstances(): Promise<any[]> {
        await this.ensureRunning()
        return this.httpReq("GET", "/launch/instances").then((r: any) => r.instances ?? [])
    }

    async getRunningInstance(id: string): Promise<any | null> {
        await this.ensureRunning()
        return this.httpReq("GET", `/launch/instances/${id}`).catch((e: any) => {
            if (e.message?.includes("404")) return null
            throw e
        })
    }

    async getLatestCrash(): Promise<any | null> {
        await this.ensureRunning()
        return this.httpReq("GET", "/crashes/latest").catch((e: any) => {
            if (e.message?.includes("404")) return null
            throw e
        })
    }

    async getSessions(): Promise<any[]> {
        await this.ensureRunning()
        return this.httpReq("GET", "/sessions")
    }

    async getRecoverySessions(): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", "/install/recovery")
    }

    async installModLoader(req: any): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", "/modloaders/install", req)
    }

    async downloadRuntime(version: string, instancePath: string, sharedPath?: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("POST", "/runtime", { version, instancePath, sharedPath })
    }

    async getWorlds(): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", "/novacore/worlds")
    }

    async getApiInfo(): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", "/api")
    }

    async getModLoaderState(instancePath: string): Promise<any> {
        await this.ensureRunning()
        return this.httpReq("GET", `/modloaders/state/${encodeURIComponent(instancePath)}`)
    }

    async deleteModLoaderState(instancePath: string): Promise<void> {
        await this.ensureRunning()
        return this.httpReq("DELETE", `/modloaders/state/${encodeURIComponent(instancePath)}`)
    }

    getJarPath(): string {
        return getNovaCoreJarPath()
    }

    getAuthlibPath(): string {
        return getAuthlibInjectorPath()
    }

    getInstancesDir(): string {
        return getInstancesDir()
    }

    getVersionsDir(): string {
        return getVersionsDir()
    }

    getBaseDir(): string {
        return getBaseDir()
    }

    getJavaPath(): string {
        return getJavaBinPath()
    }

    getVersion(): string {
        try {
            const pkgPath = join(MANAGER_DIR, "../../package.json")
            const pkg = JSON.parse(readFileSync(pkgPath, "utf-8"))
            return pkg.version ?? "1.0.0"
        } catch {
            return "1.0.0"
        }
    }

    async getDownloadedVersions(): Promise<string[]> {
        const { readdirSync, statSync, existsSync } = await import("fs")
        const { join } = await import("path")
        const dir = getVersionsDir()
        if (!existsSync(dir)) return []
        return readdirSync(dir).filter(f => statSync(join(dir, f)).isDirectory()).sort()
    }

    async getLogFiles(): Promise<{ name: string; size: number; mtime: string }[]> {
        const logDir = getLogDir()
        const { readdirSync, statSync, existsSync } = await import("fs")
        const { join } = await import("path")
        if (!existsSync(logDir)) return []
        return readdirSync(logDir)
            .filter((f: string) => f.endsWith('.log'))
            .map((f: string) => {
                const stat = statSync(join(logDir, f))
                return { name: f, size: stat.size, mtime: stat.mtime.toISOString() }
            })
            .sort((a: any, b: any) => b.mtime.localeCompare(a.mtime))
    }

    async readLogFile(fileName: string, maxLines = 2000): Promise<string[]> {
        const logDir = getLogDir()
        const { readFileSync, existsSync } = await import("fs")
        const { join } = await import("path")
        const fullPath = join(logDir, fileName)
        if (!existsSync(fullPath)) return []
        const content = readFileSync(fullPath, "utf-8")
        const lines = content.split("\n")
        return lines.slice(-maxLines)
    }
}
