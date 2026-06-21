import type {
    InstallRequest, InstallResponse,
    LaunchRequest, LaunchResponse,
    SessionSnapshot, InstanceInfo, EngineInfo,
} from "../Types/index.js"

export class HttpError extends Error {
    constructor(public readonly status: number, message: string) {
        super(message)
        this.name = "NovaCoreHttpError"
    }
}

export class HttpClient {
    constructor(
        private readonly base: string,
        private readonly token: string,
        private readonly timeoutMs: number,
    ) {}

    health() { return this.get<{ status: string; version: string }>("/health") }

    install(req: InstallRequest) { return this.post<InstallResponse>("/install", req) }
    pauseInstall(id: string) { return this.post<void>(`/install/pause/${id}`, null) }
    resumeInstall(id: string) { return this.post<void>(`/install/resume/${id}`, null) }
    cancelInstall(id: string) { return this.post<void>(`/install/cancel/${id}`, null) }
    getRecoverySessions() { return this.get<{ count: number; snapshots: SessionSnapshot[] }>("/install/recovery") }

    launch(req: LaunchRequest) { return this.post<LaunchResponse>("/launch", req) }
    killInstance(id: string) { return this.post<void>(`/launch/kill/${id}`, null) }
    getRunningInstances(): Promise<InstanceInfo[]> { return this.get<{ instances: InstanceInfo[] }>("/launch/instances").then(r => r.instances ?? []) }
    getRunningInstance(id: string) { return this.get<InstanceInfo>(`/launch/instances/${id}`).catch(e => { if (e instanceof HttpError && e.status === 404) return null; throw e }) }

    getLatestCrash() { return this.get<Record<string, unknown> | null>("/crashes/latest").catch(e => { if (e instanceof HttpError && e.status === 404) return null; throw e }) }
    getSessions() { return this.get<Record<string, unknown>[]>("/sessions") }
    getWorlds() { return this.get<import("../Types/index.js").WorldListResponse>("/novacore/worlds") }

    getSession(id: string) { return this.get<SessionSnapshot>(`/progress?sessionId=${id}`).catch(e => { if (e instanceof HttpError && e.status === 404) return null; throw e }) }
    getGlobalProgress() { return this.get<Record<string, unknown>>("/progress/global") }

    getEngineInfo() { return this.get<EngineInfo>("/system/resources") }

    getModLoaders() { return this.get<{ loaders: string[] }>("/modloaders") }
    getModLoaderVersions(loader: string, mcVer: string) {
        return this.get<{ versions: Record<string, unknown>[] }>(`/modloaders/versions/${loader}/${mcVer}`)
    }
    installModLoader(req: import("../Types/index.js").ModLoaderRequest) {
        return this.post<Record<string, unknown>>("/modloaders/install", req)
    }
    getModLoaderState(instancePath: string) {
        return this.get<Record<string, unknown>>(`/modloaders/state/${encodeURIComponent(instancePath)}`)
    }
    deleteModLoaderState(instancePath: string) {
        return this.req<Record<string, unknown>>("DELETE", `/modloaders/state/${encodeURIComponent(instancePath)}`, null)
    }

    downloadRuntime(version: string, instancePath: string, sharedPath?: string) {
        return this.post<Record<string, unknown>>("/runtime", { version, instancePath, sharedPath })
    }

    close() { return this.post<void>("/close", null) }
    closeWithSignal(signal: AbortSignal) {
        return this.reqRaw("POST", "/close", null, signal)
    }

    private async get<T>(path: string): Promise<T> { return this.req<T>("GET", path, null) }
    private async post<T>(path: string, body: unknown): Promise<T> { return this.req<T>("POST", path, body) }

    private async reqRaw(method: string, path: string, body: unknown, signal: AbortSignal): Promise<void> {
        const init: RequestInit = {
            method,
            signal,
            headers: {
                "Content-Type": "application/json",
                "X-Access-Token": this.token,
            },
        }
        if (body !== null) init.body = JSON.stringify(body)
        await fetch(`${this.base}${path}`, init)
    }

    private async req<T>(method: string, path: string, body: unknown, attempt = 1): Promise<T> {
        const ctrl = new AbortController()
        const timer = setTimeout(() => ctrl.abort(), this.timeoutMs)
        let res: Response
        try {
            const init: RequestInit = {
                method,
                signal: ctrl.signal,
                headers: {
                    "Content-Type": "application/json",
                    "X-Access-Token": this.token,
                },
            }
            if (body !== null) {
                init.body = JSON.stringify(body)
            }
            res = await fetch(`${this.base}${path}`, init)
        } catch (e) {
            if (attempt < 2) {
                clearTimeout(timer)
                return this.req(method, path, body, attempt + 1)
            }
            throw new HttpError(0, `Network: ${e instanceof Error ? e.message : String(e)}`)
        } finally {
            clearTimeout(timer)
        }
        let json: unknown
        try { json = await res.json() } catch { json = {} }
        if (!res.ok) throw new HttpError(res.status, (json as Record<string, string>)["error"] ?? `HTTP ${res.status}`)
        return json as T
    }
}
