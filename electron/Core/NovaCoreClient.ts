import { WsClient } from "./Routes/WsClient.js"
import { HttpClient } from "./Routes/HttpClient.js"
import { InstallFlow } from "./InstallFlow.js"
import { LaunchFlow } from "./LaunchFlow.js"
import type {
    InstallRequest, LaunchRequest,
    NovaCoreEventName, NovaCoreEvents,
    SessionSnapshot, InstanceInfo, EngineInfo,
    WorldListResponse,
} from "./Types/index.js"
import type { InstallCallbacks } from "./InstallFlow.js"
import type { LaunchCallbacks, LaunchHandle } from "./LaunchFlow.js"

export interface NovaCoreClientOptions {
    httpUrl?: string
    wsUrl?: string
    token: string
    timeoutMs?: number
    autoReconnect?: boolean
    healthIntervalMs?: number
}

export class NovaCoreClient {
    private readonly ws: WsClient
    private readonly http: HttpClient
    private readonly installFlow: InstallFlow
    private readonly launchFlow: LaunchFlow
    private healthCheckInterval: ReturnType<typeof setInterval> | null = null
    private healthFails = 0
    private readonly healthIntervalMs: number

    constructor(opts: NovaCoreClientOptions) {
        this.healthIntervalMs = opts.healthIntervalMs ?? 10_000
        this.http = new HttpClient(
            opts.httpUrl ?? "http://localhost:7878",
            opts.token,
            opts.timeoutMs ?? 30_000,
        )
        this.ws = new WsClient({
            url: opts.wsUrl ?? "ws://localhost:7879",
            token: opts.token,
            autoReconnect: opts.autoReconnect ?? true,
        })
        this.installFlow = new InstallFlow(this.ws, this.http)
        this.launchFlow = new LaunchFlow(this.ws, this.http)
    }

    connect(): Promise<void> {
        this.startHealthCheck()
        return this.ws.connect()
    }

    disconnect(): void {
        this.stopHealthCheck()
        this.ws.close()
    }

    async shutdown(force = false): Promise<void> {
        this.stopHealthCheck()
        this.ws.close()
        try {
            if (force) {
                const ctrl = new AbortController()
                const t = setTimeout(() => ctrl.abort(), 3000)
                await this.http.closeWithSignal(ctrl.signal)
                clearTimeout(t)
            } else {
                await this.http.close()
            }
        } catch {}
    }

    get isConnected(): boolean { return this.ws.connected }

    on<K extends NovaCoreEventName>(event: K, handler: (data: NovaCoreEvents[K]) => void): this {
        this.ws.on(event, handler); return this
    }
    off<K extends NovaCoreEventName>(event: K, handler: (data: NovaCoreEvents[K]) => void): this {
        this.ws.off(event, handler); return this
    }
    once<K extends NovaCoreEventName>(event: K, handler: (data: NovaCoreEvents[K]) => void): this {
        this.ws.once(event, handler); return this
    }
    onAny(handler: (event: NovaCoreEventName, data: unknown) => void): this {
        this.ws.onAny(handler); return this
    }
    waitFor<K extends NovaCoreEventName>(event: K, timeoutMs?: number): Promise<NovaCoreEvents[K]> {
        return this.ws.waitFor(event, timeoutMs)
    }

    install(req: InstallRequest, callbacks?: InstallCallbacks, timeoutMs?: number): Promise<void> {
        return this.installFlow.run(req, callbacks, timeoutMs)
    }

    pauseInstall(sessionId: string): Promise<void> { return this.http.pauseInstall(sessionId) }
    resumeInstall(sessionId: string): Promise<void> { return this.http.resumeInstall(sessionId) }
    cancelInstall(sessionId: string): Promise<void> { return this.http.cancelInstall(sessionId) }

    launch(req: LaunchRequest, callbacks?: LaunchCallbacks): Promise<LaunchHandle> {
        return this.launchFlow.run(req, callbacks)
    }

    killInstance(launchId: string): Promise<void> { return this.http.killInstance(launchId) }
    getRunningInstances(): Promise<InstanceInfo[]> { return this.http.getRunningInstances() }
    getRunningInstance(launchId: string) { return this.http.getRunningInstance(launchId) }

    getLatestCrash() { return this.http.getLatestCrash() }
    getSessions() { return this.http.getSessions() }
    getWorlds(): Promise<WorldListResponse> { return this.http.getWorlds() }

    getSession(sessionId: string): Promise<SessionSnapshot | null> {
        return this.http.getSession(sessionId)
    }
    async getRecoverySessions(): Promise<SessionSnapshot[]> {
        const r = await this.http.getRecoverySessions()
        return r.snapshots
    }

    getGlobalProgress() { return this.http.getGlobalProgress() }
    getEngineInfo(): Promise<EngineInfo> { return this.http.getEngineInfo() }

    getModLoaders(): Promise<{ loaders: string[] }> { return this.http.getModLoaders() }
    getModLoaderVersions(loader: string, mcVersion: string) { return this.http.getModLoaderVersions(loader, mcVersion) }
    installModLoader(req: import("./Types/index.js").ModLoaderRequest) { return this.http.installModLoader(req) }
    getModLoaderState(instancePath: string) { return this.http.getModLoaderState(instancePath) }
    deleteModLoaderState(instancePath: string) { return this.http.deleteModLoaderState(instancePath) }

    downloadRuntime(version: string, instancePath: string, sharedPath?: string) { return this.http.downloadRuntime(version, instancePath, sharedPath) }

    closeEngine(): Promise<void> { return this.http.close() }

    private startHealthCheck() {
        this.stopHealthCheck()
        this.healthCheckInterval = setInterval(async () => {
            try {
                await this.http.health()
                this.healthFails = 0
            } catch {
                this.healthFails++
                if (this.healthFails >= 2) {
                    this.ws.dispatchLocal("engine_unreachable", {
                        reason: "health_check_failed",
                        fails: this.healthFails,
                        timestamp: Date.now(),
                    })
                }
            }
        }, this.healthIntervalMs)
    }

    private stopHealthCheck() {
        if (this.healthCheckInterval) {
            clearInterval(this.healthCheckInterval)
            this.healthCheckInterval = null
        }
        this.healthFails = 0
    }
}
