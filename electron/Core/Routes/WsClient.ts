import type { NovaCoreEventName, NovaCoreEvents, WsBaseEvent } from "../Types/index.js"

type Callback<K extends NovaCoreEventName> = (data: NovaCoreEvents[K], raw: WsBaseEvent) => void
type AnyCallback = (data: unknown, raw: WsBaseEvent) => void

export interface WsClientOptions {
    url: string
    token: string
    autoReconnect?: boolean
    reconnectDelay?: number
    maxReconnectAttempts?: number
}

export class WsClient {
    private readonly opts: Required<WsClientOptions>
    private ws: WebSocket | null = null
    private pendingConnect: Promise<void> | null = null
    private attempts = 0
    private timer: ReturnType<typeof setTimeout> | null = null
    private dead = false
    private readonly map = new Map<NovaCoreEventName | "*", Set<AnyCallback>>()
    private heartbeatTimer: ReturnType<typeof setInterval> | null = null

    constructor(opts: WsClientOptions) {
        this.opts = {
            autoReconnect: true,
            reconnectDelay: 1500,
            maxReconnectAttempts: 0,
            ...opts,
        }
    }

    connect(): Promise<void> {
        if (this.connected) return Promise.resolve()
        if (this.pendingConnect) return this.pendingConnect

        this.dead = false
        this.pendingConnect = new Promise((resolve, reject) => {
            const url = `${this.opts.url}?token=${encodeURIComponent(this.opts.token)}`
            try {
                this.ws = new WebSocket(url)
            } catch (e) {
                this.pendingConnect = null
                reject(e)
                return
            }

            let settled = false
            const timeout = setTimeout(() => {
                complete(() => reject(new Error(`WebSocket connect timeout after 15000ms`)))
            }, 15_000)

            const complete = (fn: () => void) => {
                if (settled) return
                settled = true
                this.pendingConnect = null
                clearTimeout(timeout)
                fn()
            }

            const onConnected = () => {
                complete(() => resolve())
            }

            this.once("connected", onConnected)

            const cleanupConnect = () => {
                this.off("connected", onConnected)
            }

            this.ws.onopen = () => {
                this.attempts = 0
                this.startHeartbeat()
            }
            this.ws.onmessage = (e) => this.dispatch(e.data as string)
            this.ws.onerror = () => {
                cleanupConnect()
                complete(() => reject(new Error("WebSocket connection failed")))
            }
            this.ws.onclose = (e) => {
                cleanupConnect()
                if (!settled) {
                    complete(() => reject(new Error(`WebSocket closed before connected (code=${e.code})`)))
                }
                if (e.code === 1008) {
                    this.dead = true
                    return
                }
                this.stopHeartbeat()
                if (!this.dead) {
                    this.dispatchLocal("engine_unreachable", {
                        reason: "ws_disconnected",
                        code: e.code,
                        timestamp: Date.now(),
                    })
                }
                this.scheduleReconnect()
            }
        })

        return this.pendingConnect
    }

    close(): void {
        this.dead = true
        this.stopHeartbeat()
        if (this.timer) { clearTimeout(this.timer); this.timer = null }
        this.ws?.close(1000, "bye")
        this.ws = null
    }

    get connected(): boolean { return this.ws?.readyState === WebSocket.OPEN }

    on<K extends NovaCoreEventName>(event: K, cb: Callback<K>): this {
        this.add(event, cb as AnyCallback); return this
    }
    off<K extends NovaCoreEventName>(event: K, cb: Callback<K>): this {
        this.map.get(event)?.delete(cb as AnyCallback); return this
    }
    once<K extends NovaCoreEventName>(event: K, cb: Callback<K>): this {
        const w: AnyCallback = (d, r) => { this.off(event, w as unknown as Callback<K>); (cb as AnyCallback)(d, r) }
        this.add(event, w); return this
    }
    onAny(cb: (event: NovaCoreEventName, data: unknown) => void): this {
        this.add("*", (d, r) => cb(r.event as NovaCoreEventName, d)); return this
    }
    waitFor<K extends NovaCoreEventName>(event: K, ms = 30_000): Promise<NovaCoreEvents[K]> {
        return new Promise((res, rej) => {
            const t = setTimeout(() => { this.off(event, cb); rej(new Error(`Timeout: "${event}"`)) }, ms)
            const cb: Callback<K> = (d) => { clearTimeout(t); res(d) }
            this.once(event, cb)
        })
    }

    dispatchLocal(event: NovaCoreEventName, data: unknown) {
        const ts = Date.now()
        const raw = { event, data, ts } as WsBaseEvent
        this.map.get(event)?.forEach(cb => { try { cb(data, raw) } catch {} })
        this.map.get("*")?.forEach(cb => { try { cb(data, raw) } catch {} })
    }

    private add(k: NovaCoreEventName | "*", cb: AnyCallback) {
        if (!this.map.has(k)) this.map.set(k, new Set())
        this.map.get(k)!.add(cb)
    }
    private dispatch(raw: string) {
        let p: WsBaseEvent
        try { p = JSON.parse(raw) as WsBaseEvent } catch { return }
        const key = p.event as NovaCoreEventName
        this.map.get(key)?.forEach(cb => { try { cb(p.data, p) } catch {} })
        this.map.get("*")?.forEach(cb => { try { cb(p.data, p) } catch {} })
    }
    private scheduleReconnect() {
        if (this.dead || !this.opts.autoReconnect) return
        if (this.timer) return

        const { maxReconnectAttempts: max, reconnectDelay: base } = this.opts
        if (max > 0 && this.attempts >= max) return

        this.attempts++
        const delay = base * Math.min(this.attempts, 6)

        this.timer = setTimeout(() => {
            this.timer = null
            if (!this.dead) {
                this.connect().catch(() => {})
            }
        }, delay)
    }

    private startHeartbeat() {
        this.stopHeartbeat()
        this.heartbeatTimer = setInterval(() => {
            if (this.connected) {
                try { this.ws?.send(JSON.stringify({ event: "heartbeat", data: { timestamp: Date.now() } })) } catch {}
            }
        }, 30_000)
    }

    private stopHeartbeat() {
        if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer)
            this.heartbeatTimer = null
        }
    }
}
