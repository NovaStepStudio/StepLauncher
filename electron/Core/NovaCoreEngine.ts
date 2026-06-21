import { EngineProcess } from "./EngineProcess.js"
import { NovaCoreClient } from "./NovaCoreClient.js"
import type { EngineProcessOptions } from "./EngineProcess.js"
import type { NovaCoreClientOptions } from "./NovaCoreClient.js"

export type NovaCoreEngineOptions = EngineProcessOptions & {
    client?: Pick<NovaCoreClientOptions, "timeoutMs" | "autoReconnect" | "healthIntervalMs">
}

export interface AttachOptions {
    httpUrl: string
    wsUrl: string
    token: string
    timeoutMs?: number
    autoReconnect?: boolean
    healthIntervalMs?: number
}

export class NovaCoreEngine {
    private static instances = new Set<{ client: NovaCoreClient; process?: EngineProcess }>()
    private static cleanupInited = false

    private static initCleanup() {
        if (NovaCoreEngine.cleanupInited) return
        NovaCoreEngine.cleanupInited = true
        const run = () => {
            for (const inst of NovaCoreEngine.instances) {
                try { inst.client.disconnect() } catch {}
                try { inst.process?.kill() } catch {}
            }
        }
        process.on("exit", run)
        process.on("SIGINT", () => { run(); process.exit(0) })
        process.on("SIGTERM", () => { run(); process.exit(0) })
    }

    static async start(opts: NovaCoreEngineOptions): Promise<NovaCoreClient> {
        NovaCoreEngine.initCleanup()
        const proc = new EngineProcess(opts)
        const info = await proc.start()

        const client = new NovaCoreClient({
            httpUrl: info.httpUrl,
            wsUrl: info.wsUrl,
            token: info.token,
            timeoutMs: opts.client?.timeoutMs,
            autoReconnect: opts.client?.autoReconnect,
            healthIntervalMs: opts.client?.healthIntervalMs,
        })

        await client.connect()
        NovaCoreEngine.instances.add({ client, process: proc })
        return client
    }

    static async startWithHandle(opts: NovaCoreEngineOptions): Promise<{
        client: NovaCoreClient
        process: EngineProcess
    }> {
        NovaCoreEngine.initCleanup()
        const proc = new EngineProcess(opts)
        const info = await proc.start()

        const client = new NovaCoreClient({
            httpUrl: info.httpUrl,
            wsUrl: info.wsUrl,
            token: info.token,
            timeoutMs: opts.client?.timeoutMs,
            autoReconnect: opts.client?.autoReconnect,
            healthIntervalMs: opts.client?.healthIntervalMs,
        })

        await client.connect()
        const pair = { client, process: proc }
        NovaCoreEngine.instances.add(pair)
        return pair
    }

    static async attach(opts: AttachOptions): Promise<NovaCoreClient> {
        const client = new NovaCoreClient({
            httpUrl: opts.httpUrl,
            wsUrl: opts.wsUrl,
            token: opts.token,
            timeoutMs: opts.timeoutMs,
            autoReconnect: opts.autoReconnect ?? true,
            healthIntervalMs: opts.healthIntervalMs,
        })
        await client.connect()
        NovaCoreEngine.instances.add({ client })
        return client
    }

    static async attachWithHandle(opts: AttachOptions): Promise<{
        client: NovaCoreClient
        process: EngineProcess
    }> {
        const client = new NovaCoreClient({
            httpUrl: opts.httpUrl,
            wsUrl: opts.wsUrl,
            token: opts.token,
            timeoutMs: opts.timeoutMs,
            autoReconnect: opts.autoReconnect ?? true,
            healthIntervalMs: opts.healthIntervalMs,
        })
        await client.connect()
        const process = EngineProcess.attach({
            httpUrl: opts.httpUrl,
            wsUrl: opts.wsUrl,
            token: opts.token,
            pid: 0,
        })
        NovaCoreEngine.instances.add({ client, process })
        return { client, process }
    }
}
