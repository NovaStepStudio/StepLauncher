import { spawn, ChildProcess, execSync } from "node:child_process"
import { existsSync } from "node:fs"
import { resolve } from "node:path"

export interface EngineProcessOptions {
    jar: string
    java?: string
    httpPort?: number
    wsPort?: number
    instancesDir?: string
    logDir?: string
    logLevel?: "DEBUG" | "INFO" | "WARN" | "ERROR"
    launcherName?: string
    threads?: number
    startupTimeoutMs?: number
    verbose?: boolean
    jvmArgs?: string[]
    autoKillOnExit?: boolean
}

export interface EngineProcessInfo {
    pid: number
    token: string
    httpUrl: string
    wsUrl: string
}

export class EngineProcess {
    private opts: Required<EngineProcessOptions>
    private child: ChildProcess | null = null
    private _info: EngineProcessInfo | null = null
    private exitHandler: (() => void) | null = null

    constructor(opts: EngineProcessOptions) {
        this.opts = {
            java: "java",
            httpPort: 7878,
            wsPort: 7879,
            instancesDir: resolve("instances"),
            logDir: "",
            logLevel: "INFO",
            launcherName: "Third_Party",
            threads: 2,
            startupTimeoutMs: 45_000,
            verbose: false,
            jvmArgs: [],
            autoKillOnExit: true,
            ...opts,
        }

        if (!this.opts.logDir) {
            this.opts.logDir = resolve(this.opts.instancesDir, "..", "logs")
        }
    }

    private get httpUrl(): string {
        return `http://localhost:${this.opts.httpPort}`
    }

    async healthCheck(timeoutMs = 3000): Promise<boolean> {
        if (!this._info) return false
        try {
            const ctrl = new AbortController()
            const t = setTimeout(() => ctrl.abort(), timeoutMs)
            const res = await fetch(`${this.httpUrl}/health`, {
                headers: { "X-Access-Token": this._info.token },
                signal: ctrl.signal,
            })
            clearTimeout(t)
            return res.ok
        } catch {
            return false
        }
    }

    async start(): Promise<EngineProcessInfo> {
        if (this.child && !this.child.killed) {
            throw new Error("Engine is already running.")
        }

        const { jar, java } = this.opts

        if (!existsSync(jar)) {
            throw new Error(`JAR not found: ${jar}`)
        }

        const args = [
            ...this.opts.jvmArgs,
            "-jar", jar,
            "--port", String(this.opts.httpPort),
            "--ws-port", String(this.opts.wsPort),
            "--threads", String(this.opts.threads),
            "--instances-dir", this.opts.instancesDir,
            "--log-dir", this.opts.logDir,
            "--log-level", this.opts.logLevel,
            "--launcher-name", this.opts.launcherName,
        ]

        this.child = spawn(java, args, {
            stdio: ["ignore", "pipe", "pipe"],
            detached: process.platform !== "win32",
        })

        let info: EngineProcessInfo
        try {
            info = await this.waitForReady()
        } catch (err) {
            this.kill()
            throw err
        }

        this._info = info

        if (this.opts.verbose) {
            this.child.stdout?.on("data", (d: Buffer) => process.stdout.write(d))
            this.child.stderr?.on("data", (d: Buffer) => process.stderr.write(d))
        }

        if (this.opts.autoKillOnExit) {
            this.setupExitHandlers()
        }

        return info
    }

    async stop(): Promise<void> {
        this.removeExitHandlers()

        if (this._info) {
            try {
                const ctrl = new AbortController()
                const t = setTimeout(() => ctrl.abort(), 2000)
                await fetch(`${this._info.httpUrl}/close`, {
                    method: "POST",
                    headers: { "X-Access-Token": this._info.token },
                    signal: ctrl.signal,
                })
                clearTimeout(t)
            } catch {}
        }

        if (this.child && !this.child.killed) {
            return new Promise<void>((res) => {
                const timeout = setTimeout(() => { this.kill(); res() }, 3000)
                this.child!.once("exit", () => { clearTimeout(timeout); res() })
                this.child!.kill("SIGTERM")
            })
        }
    }

    static attach(info: EngineProcessInfo): EngineProcess {
        const proc = new EngineProcess({ jar: "" } as EngineProcessOptions)
        proc._info = info
        proc.child = null
        return proc
    }

    kill(): void {
        this.removeExitHandlers()

        if (!this.child || this.child.killed) {
            this._info = null
            return
        }

        const pid = this.child.pid
        if (pid) {
            try {
                if (process.platform === "win32") {
                    execSync(`taskkill /F /T /PID ${pid}`, { stdio: "ignore" })
                } else {
                    try {
                        process.kill(-pid, "SIGKILL")
                    } catch {
                        this.child.kill("SIGKILL")
                    }
                }
            } catch {
                this.child.kill("SIGKILL")
            }
        }

        this.child = null
        this._info = null
    }

    get running(): boolean { return !!this.child && !this.child.killed }
    get info(): EngineProcessInfo | null { return this._info }

    private waitForReady(): Promise<EngineProcessInfo> {
        return new Promise<EngineProcessInfo>((resolve, reject) => {
            const { httpPort, wsPort, startupTimeoutMs } = this.opts
            const child = this.child!

            let token = ""
            let ready = false
            let buf = ""

            const timeout = setTimeout(() => {
                cleanup()
                reject(new Error(`Engine did not start within ${startupTimeoutMs}ms. Check Java and JAR.`))
            }, startupTimeoutMs)

            const cleanup = () => {
                clearTimeout(timeout)
                child.stdout?.removeListener("data", onData)
                child.stderr?.removeListener("data", onData)
                child.removeListener("error", onError)
                child.removeListener("exit", onExit)
            }

            const onData = (chunk: Buffer) => {
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
                        resolve({
                            pid: child.pid!,
                            token,
                            httpUrl: `http://localhost:${httpPort}`,
                            wsUrl: `ws://localhost:${wsPort}`,
                        })
                    }
                }
            }

            const onError = (err: Error) => {
                cleanup()
                if ((err as NodeJS.ErrnoException).code === "ENOENT") {
                    reject(new Error(`Java not found at "${this.opts.java}".`))
                } else {
                    reject(err)
                }
            }

            const onExit = (code: number | null) => {
                cleanup()
                reject(new Error(`Process exited unexpectedly (code ${code}) before ready.`))
            }

            child.stdout?.on("data", onData)
            child.stderr?.on("data", onData)
            child.once("error", onError)
            child.once("exit", onExit)
        })
    }

    private setupExitHandlers() {
        this.exitHandler = () => this.kill()
        process.once("exit", this.exitHandler)
        process.once("SIGINT", this.exitHandler)
        process.once("SIGTERM", this.exitHandler)
        process.once("SIGHUP", this.exitHandler)
    }

    private removeExitHandlers() {
        if (this.exitHandler) {
            process.off("exit", this.exitHandler)
            process.off("SIGINT", this.exitHandler)
            process.off("SIGTERM", this.exitHandler)
            process.off("SIGHUP", this.exitHandler)
            this.exitHandler = null
        }
    }
}
