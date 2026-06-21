export type SessionStatus =
    | "pending" | "running" | "paused"
    | "cancelled" | "completed" | "failed"

export type ModuleStatus =
    | "pending" | "downloading" | "completed"
    | "verifying" | "verified" | "failed" | "retrying"

export type GcPreset =
    | "auto" | "none" | "disabled" | "off"
    | "g1gc_basic" | "g1gc_optimized"
    | "zgc" | "shenandoah"

export type GpuPreference = "none" | "auto" | "dgpu" | "igpu"

export interface ModLoaderRequest {
    loader: string
    loaderType?: string
    loaderVersion: string
    minecraftVersion: string
    instancePath: string
    sharedPath?: string
    maxThreads?: number
    debug?: boolean
}

export interface DownloadOptions {
    client?: boolean
    libraries?: boolean
    assets?: boolean
    natives?: boolean
    jvm?: boolean
}

export interface InstallRequest {
    version: string
    instancePath: string
    sharedPath?: string
    isInstance?: boolean
    download?: DownloadOptions
    verifySHA1?: boolean
    maxThreads?: number
    debug?: boolean
    modloader?: "none" | "fabric" | "forge" | "neoforge" | "quilt" | "legacyfabric" | "optifine" | string
    modloaderVersion?: string
    launcher?: LauncherBranding
}

export interface InstallResponse {
    sessionId: string
    version: string
    instancePath: string
    status: "started"
    progress: string
}

export interface AuthConfig {
    username: string
    uuid: string
    accessToken: string
    userType?: string
    clientId?: string
    xuid?: string
}

export interface AuthlibInjector {
    enabled: boolean
    jarPath: string
    serverUrl: string
}

export interface JvmConfig {
    minMemoryMb?: number
    maxMemoryMb?: number
    extraArgs?: string[]
    prependArgs?: string[]
}

export interface WindowConfig {
    width?: number
    height?: number
    fullscreen?: boolean
}

export interface LauncherBranding {
    name: string
    version?: string
}

export interface GameCustomization {
    gameDir?: string
    extraGameArgs?: string[]
    extraJvmProperties?: Record<string, string>
    serverHost?: string
    serverPort?: number
}

export interface QuickPlayConfig {
    mode: "singleplayer" | "multiplayer" | "realms"
    value: string
}

export interface LaunchFeatures {
    demo?: boolean
    quickPlay?: QuickPlayConfig
}

export interface LaunchRequest {
    version: string
    instancePath: string
    sharedPath?: string
    javaPath?: string
    hardwareAcceleration?: boolean
    gcPreset?: GcPreset
    gpuPreference?: GpuPreference
    auth?: AuthConfig
    authlibInjector?: AuthlibInjector
    jvm?: JvmConfig
    window?: WindowConfig
    launcher?: LauncherBranding
    game?: GameCustomization
    features?: LaunchFeatures
}

export interface LaunchResponse {
    launchId: string
    version: string
    status: "started"
}

export interface SessionSnapshot {
    sessionId: string
    status: SessionStatus
    createdAt: number
    totalFiles: number
    completedFiles: number
    skippedFiles: number
    failedFiles: number
    pendingFiles: number
    totalBytes: number
    downloadedBytes: number
    overallPercent: number
    error?: string
}

export interface InstanceInfo {
    launchId: string
    version: string
    username: string
    instancePath: string
    startedAt: number
    pid: number
    status: "starting" | "running" | "stopping" | "stopped"
    exitCode: number
    logFile: string | null
}

export interface EngineInfo {
    version: string
    cpu: { cores: number; optimalDlThreads: number }
    ram: { totalMb: number; estimatedFreeMb: number; reservedForOsMb: number }
    recommended: {
        downloadThreads: number
        mcMinRamMb: number
        mcMaxRamMb: number
        gcPreset: string
    }
}

export interface CrashContext {
    instanceId: string
    exitCode: number
    source: string
    reason: string
    context: string[]
    timestamp: number
}

export interface SessionRecord {
    launchId: string
    instancePath: string
    version: string
    durationMs: number
    startedAt: number
    exitCode: number
}

export interface WorldMetadata {
    folderName: string
    levelName: string
    lastPlayed: number
    gameType: number
    versionName: string
    path: string
    instanceId: string
    iconBase64: string | null
}

export interface WorldListResponse {
    worlds: WorldMetadata[]
}

export interface WsBaseEvent {
    event: string
    data: unknown
    ts: number
}

export interface NovaCoreEvents {
    connected: { message: string; version: string }

    install_step: {
        sessionId: string
        step: string
        [key: string]: unknown
    }
    module_status: {
        sessionId: string
        module: "client" | "libraries" | "assets" | "natives"
        status: ModuleStatus
    }
    manifest_resolved: { sessionId: string; versionId: string }
    offline_mode: { sessionId: string; version: string; reason: string }
    tasks_ready: {
        sessionId: string
        totalTasks: number
        totalBytes: number
        offline: boolean
        breakdown: { client: number; libraries: number; assets: number; natives: number; asset_index: number }
    }
    install_completed: { sessionId: string; version: string; modloader: string }
    install_failed: { sessionId: string; reason: string; modules: Record<string, ModuleStatus> }

    session_started: { session: string; totalFiles: number; totalBytes: number }
    session_progress: {
        sessionId: string
        completedFiles: number
        skippedFiles: number
        totalFiles: number
        overallPercent: number
        downloadedBytes: number
        totalBytes: number
    }
    session_completed: { sessionId: string; totalFiles: number; downloadedBytes: number }
    session_failed: { sessionId: string; reason: string }
    session_paused: { session: string }
    session_resumed: { session: string }
    session_cancelled: { session: string }

    download_start: { sessionId: string; category: string; file: string; size: number }
    download_progress: { sessionId: string; category: string; file: string; downloaded: number; total: number }
    download_complete: { sessionId: string; category: string; file: string; bytes: number; skipped: boolean }
    download_error: { sessionId: string; category: string; file: string; error: string }
    sha1_check: { sessionId: string; file: string; ok: boolean; expected: string; computed: string }

    launch_preparing: { launchId: string; version: string }
    launch_starting: { launchId: string; mainClass: string; version: string }
    launch_started: { launchId: string; pid: number; logFile: string }
    launch_failed: { launchId: string; error: string }
    launch_verification_failed: { launchId: string; missing: string[]; hint: string }
    launch_exited: {
        launchId: string
        exitCode: number
        normal: boolean
        durationMs: number
    }
    game_crash: { launchId: string; exitCode: number; reason: string; context?: string[]; source?: string; timestamp?: number }

    game_log: {
        launchId: string; line: string; stream: "stdout" | "stderr"
        level: string; logger: string; message: string
    }

    modloader_processor_log: { sessionId: string; line: string }

    debug: { sessionId?: string; message: string }
    recovery_state: { sessions: SessionSnapshot[] }
    engine_state: Record<string, unknown>

    engine_unreachable: {
        reason: string
        timestamp: number
        fails?: number
        code?: number
    }
}

export type NovaCoreEventName = keyof NovaCoreEvents
