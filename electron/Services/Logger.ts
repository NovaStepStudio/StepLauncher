import { appendFileSync, mkdirSync, readdirSync, rmSync, writeFileSync } from "fs"
import { join } from "path"
import { homedir, platform, EOL } from "os"

type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL"

const LOG_LEVELS: Record<LogLevel, number> = {
    DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3, FATAL: 4,
}

let _minLevel: LogLevel = "DEBUG"
let _logDir = ""
let _logPath = ""
let _originalConsole: Record<string, (...args: unknown[]) => void> = {}

function getBaseDir(): string {
    return platform() === "win32"
        ? join(process.env.APPDATA || homedir(), ".StepLauncher")
        : join(homedir(), ".StepLauncher")
}

function ensureLogDir(): string {
    const dir = join(getBaseDir(), "logs")
    mkdirSync(dir, { recursive: true })
    return dir
}

function rotateLogs(maxFiles = 10) {
    try {
        const dir = ensureLogDir()
        const files = readdirSync(dir)
            .filter(f => f.startsWith("launcher-") && f.endsWith(".log"))
            .sort()
            .reverse()
        while (files.length >= maxFiles) {
            const old = files.pop()!
            rmSync(join(dir, old), { force: true })
        }
    } catch {}
}

function formatLine(level: LogLevel, msg: string): string {
    const ts = new Date().toISOString()
    return `[${ts}] [${level.padEnd(5)}] ${msg}${EOL}`
}

function write(level: LogLevel, msg: string) {
    if (LOG_LEVELS[level] < LOG_LEVELS[_minLevel]) return
    const line = formatLine(level, msg)
    try {
        appendFileSync(_logPath, line)
    } catch {}
    try {
        if (level === "ERROR" || level === "FATAL") {
            _originalConsole.error?.(msg)
        } else if (level === "WARN") {
            _originalConsole.warn?.(msg)
        } else {
            _originalConsole.log?.(msg)
        }
    } catch {}
}

function formatArgs(args: unknown[]): string {
    return args.map(a => {
        if (a instanceof Error) {
            return `${a.message}${a.stack ? EOL + a.stack : ""}`
        }
        if (typeof a === "object" && a !== null) {
            try { return JSON.stringify(a, null, 2) } catch { return String(a) }
        }
        return String(a)
    }).join(" ")
}

export function initLogger(minLevel: LogLevel = "DEBUG") {
    _minLevel = minLevel
    _logDir = ensureLogDir()
    _logPath = join(_logDir, `launcher-${Date.now()}.log`)

    _originalConsole = {
        log: console.log.bind(console),
        warn: console.warn.bind(console),
        error: console.error.bind(console),
        debug: console.debug?.bind(console),
    }

    rotateLogs(10)

    writeFileSync(_logPath, formatLine("INFO", `Logger initialized — StepLauncher v${getVersion()}`))
    writeFileSync(_logPath, formatLine("INFO", `Platform: ${platform()}, Node: ${process.version}, Electron: ${process.versions?.electron ?? "?"}`), { flag: "a" })

    console.log = (...args: unknown[]) => write("INFO", formatArgs(args))
    console.warn = (...args: unknown[]) => write("WARN", formatArgs(args))
    console.error = (...args: unknown[]) => write("ERROR", formatArgs(args))
    console.debug = (...args: unknown[]) => write("DEBUG", formatArgs(args))

    process.on("uncaughtException", (err: Error) => {
        write("FATAL", `UNCAUGHT EXCEPTION: ${err.message}${EOL}${err.stack ?? ""}`)
        try {
            _originalConsole.error?.("[Logger] Uncaught exception captured:", err)
        } catch {}
    })

    process.on("unhandledRejection", (reason: unknown) => {
        const msg = reason instanceof Error
            ? `${reason.message}${EOL}${reason.stack ?? ""}`
            : String(reason)
        write("FATAL", `UNHANDLED REJECTION: ${msg}`)
        try {
            _originalConsole.error?.("[Logger] Unhandled rejection captured:", reason)
        } catch {}
    })

    write("INFO", "Logger system ready — console.error/warn/log safely captured")
}

export function getLogPath(): string {
    return _logPath
}

export function getLogDir(): string {
    return _logDir
}

export function flushLogs(): void {}

export function restoreConsole() {
    console.log = _originalConsole.log
    console.warn = _originalConsole.warn
    console.error = _originalConsole.error
    console.debug = _originalConsole.debug ?? console.debug
}

function getVersion(): string {
    try {
        const { readFileSync } = require("fs")
        const { join } = require("path")
        const pkg = JSON.parse(readFileSync(join(__dirname, "../../package.json"), "utf-8"))
        return pkg.version ?? "1.0.0"
    } catch {
        return "1.0.0"
    }
}
