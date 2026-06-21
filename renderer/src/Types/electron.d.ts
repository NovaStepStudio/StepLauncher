interface AuthManagerAPI {
  Initialize(): Promise<{
    authenticated: boolean
    user: {
      uuid: string
      username: string
      email: string | null
      bio: string | null
      premium: string
      is_admin: boolean
      is_online: boolean
      last_seen: string | null
      last_login: string | null
      created_at: string
      skin_url: string | null
      cape_url: string | null
      avatar: string | null
      settings: Record<string, unknown> | null
      playing_time: number
      last_version_playing: string | null
      last_version_connected: string | null
      email_verified: boolean
    } | null
    token: string | null
  }>
  Login(username: string, password: string): Promise<{
    token: string
    expires_at: number
    expires_in: number
    success: boolean
    user: {
      uuid: string
      username: string
      premium: string
      avatar: string | null
      email_verified: boolean
      email: string | null
      bio: string | null
      is_admin: boolean
      is_online: boolean
      last_seen: string | null
      last_login: string | null
      created_at: string
      skin_url: string | null
      cape_url: string | null
      settings: Record<string, unknown> | null
      playing_time: number
      last_version_playing: string | null
      last_version_connected: string | null
    }
  }>
  Logout(): Promise<void>
  RestoreSession(): Promise<boolean>
  StartOAuth(): Promise<{ code: string;
 state: string }>
  CompleteOAuth(code: string, state: string): Promise<{
    access_token: string
    token_type: string
    expires_in: number
    user: {
      uuid: string
      username: string
      premium: string
      avatar: string | null
      email_verified: boolean
      email: string | null
      bio: string | null
      is_admin: boolean
      is_online: boolean
      last_seen: string | null
      last_login: string | null
      created_at: string
      skin_url: string | null
      cape_url: string | null
      settings: Record<string, unknown> | null
      playing_time: number
      last_version_playing: string | null
      last_version_connected: string | null
    }
  }>
  GetToken(): Promise<string | null>
  GetMe(): Promise<{
    uuid: string
    username: string
    email: string | null
    bio: string | null
    premium: string
    is_admin: boolean
    is_online: boolean
    last_seen: string | null
    last_login: string | null
    created_at: string
    skin_url: string | null
    cape_url: string | null
    avatar: string | null
    settings: Record<string, unknown> | null
    playing_time: number
    last_version_playing: string | null
    last_version_connected: string | null
    email_verified: boolean
  }>
  UpdateStats(stats: { playing_time_increment?: number;
 last_version_playing?: string;
 last_version_connected?: string }): Promise<void>
  UpdateHeartbeat(version: string, playingTimeIncrement?: number): Promise<void>
  GetPublicProfile(uuid: string): Promise<{ uuid: string;
 username: string;
 bio: string | null;
 premium: string;
 is_online: boolean;
 last_seen: string | null;
 avatar: string | null;
 skin_url: string | null;
 cape_url: string | null }>
  GetFullPublicProfile(uuid: string): Promise<any>
  GetSkins(): Promise<any[]>
  GetCapes(): Promise<any[]>
  GetKits(): Promise<any[]>
  GetFriends(): Promise<any[]>
  GetFriendRequests(): Promise<{ received: any[];
 sent: any[] }>
  SearchUsers(q: string): Promise<any[]>
  SendFriendRequest(toUserId: number): Promise<void>
  AcceptFriendRequest(requestId: number): Promise<void>
  RejectFriendRequest(requestId: number): Promise<void>
  FollowUser(uuid: string): Promise<void>
  UnfollowUser(uuid: string): Promise<void>
  LikeProfile(uuid: string): Promise<void>
  UnlikeProfile(uuid: string): Promise<void>
  GetPresence(uuid: string): Promise<{ uuid: string;
 presence: string;
 message: string | null;
 image_b64: string | null;
 updated_at: string | null }>
  SetPresence(uuid: string, presence: string, message?: string): Promise<void>
  GetOAuthConnections(): Promise<Array<{ provider: string;
 provider_id: string;
 email: string | null;
 created_at: string }>>
  GetHealth(): Promise<any>
  GetLauncherToken(launcherId: string): Promise<{ launcher_token: string;
 launcher_id: string;
 expires_at: number;
 expires_in: number }>
  GetSocialLinks(): Promise<Array<{ id: number;
 platform: string;
 url: string;
 display_order: number }>>
  AddSocialLink(platform: string, url: string): Promise<{ id: number;
 platform: string;
 url: string;
 display_order: number }>
  UpdateSocialLink(id: number, url: string): Promise<void>
  DeleteSocialLink(id: number): Promise<void>
  GetLibraries(): Promise<any[]>
  CreateLibrary(title: string, description: string | null, isPublic: boolean, config?: Record<string, unknown>): Promise<any>
  GetLibrary(id: number): Promise<any>
  UpdateLibrary(id: number, updates: Record<string, unknown>): Promise<void>
  DeleteLibrary(id: number): Promise<void>
  GetEntries(libraryId: number): Promise<any[]>
  AddEntry(libraryId: number, entry: Record<string, unknown>): Promise<any>
  UpdateEntry(libraryId: number, entryId: number, updates: Record<string, unknown>): Promise<void>
  RemoveEntry(libraryId: number, entryId: number): Promise<void>
  ToggleLibraryLike(libraryId: number): Promise<{ success: boolean;
 liked: boolean;
 likes: number;
 error?: string }>
  GetPublicLibraryEntries(libraryId: number): Promise<any[]>
  CloneLibrary(sourceLibraryId: number, customTitle?: string): Promise<any>
}

interface LauncherConfig {
  autoStartMinecraft: boolean
  hideOnLaunch: boolean
  showConsole: boolean
  showNews: boolean
  filters: boolean
  blur: boolean
  hardwareAcceleration: boolean
  shadows: boolean
  discordRpc: boolean
  locale: string
}

interface MinecraftConfig {
  useRecommendedJava: boolean
  maxConsoleEvents: number
  showNotificationOnLaunch: boolean
  cleanBeforeLaunch: boolean
  javaPath: string
  fullscreen: boolean
  windowWidth: number
  windowHeight: number
  maxRam: number
  minRam: number
  gcPreset: string
  gpuPreference: string
  jvmArgs: string
  gameArgs: string
}

interface PersonalizationConfig {
  titleBarColor: string
  appBackground: string
  fontPrimary: string
  fontSecondary: string
  accentColor: string
  modalAccent?: string
  modalBackground?: string
  sidebarBackground: string
  panelBackground: string
  notificationErrorColor: string
  notificationWarnColor: string
  notificationSuccessColor: string
  sidebarButtonColor: string
  macOSTitlebar: boolean
  showIcon: boolean
  invertPosition: boolean
  sidebarPosition: 'left' | 'right'
}

interface AppConfig {
  launcher: LauncherConfig
  minecraft: MinecraftConfig
  personalization: PersonalizationConfig
}

interface ConfigManagerAPI {
  Get(): Promise<AppConfig>
  UpdateLauncher(updates: Partial<LauncherConfig>): Promise<AppConfig>
  UpdateMinecraft(updates: Partial<MinecraftConfig>): Promise<AppConfig>
  UpdatePersonalization(updates: Partial<PersonalizationConfig>): Promise<AppConfig>
  GetSystemInfo(): Promise<{
    os: string
    arch: string
    hostname: string
    user: string
    cpu: string
    totalRam: number
    ramFree: string
    java: string
    gpu: string
  }>
}

interface ElectronAPI {
  OpenExternal(url: string): Promise<void>
  OpenFileDialog(options?: { filters?: { name: string;
 extensions: string[] }[] }): Promise<{ canceled: boolean;
 filePath: string | null }>
  OpenPath(dirPath: string): Promise<string>
  ReadDir(dirPath: string): Promise<{ name: string;
 isDirectory: boolean;
 isFile: boolean;
 size: number;
 mtime: string }[] | null>
  Hide(): Promise<void>
  Show(): Promise<void>
  Minimize(): Promise<void>
  Maximize(): Promise<void>
  Close(): Promise<void>
  ShowNotification(options: { title: string;
 body: string }): Promise<void>
  FetchJson(url: string): Promise<any>
  ReadLocaleFile(locale: string): Promise<Record<string, unknown> | null>
}

interface DownloadProgress {
  item: "java" | "novacore" | "authlib"
  percent: number
  downloadedBytes: number
  totalBytes: number
  phase: "downloading" | "extracting" | "done" | "error"
  error?: string
}

interface DownloadManagerAPI {
  CheckAll(): Promise<Record<string, boolean>>
  StartAll(): Promise<{ success: boolean;
 item?: string;
 error?: string }>
  OnProgress(callback: (progress: DownloadProgress) => void): () => void
}

interface NovaCoreManagerAPI {
  Start(): Promise<{ success: boolean;
 error?: string }>
  Stop(): Promise<void>
  Status(): Promise<{ status: string;
 error?: string }>
  Health(): Promise<{ status: string;
 version: string;
 uptime: number }>
  Versions(limit?: number): Promise<any[]>
  Install(req: any): Promise<{ success: boolean;
 sessionId?: string;
 error?: string }>
  Launch(req: any): Promise<{ success: boolean;
 launchId?: string;
 error?: string }>
  KillInstance(id: string): Promise<{ success: boolean;
 error?: string }>
  EngineInfo(): Promise<any>
  ModLoaders(): Promise<{ loaders: string[] }>
  ModLoaderVersions(loader: string, mcVersion: string): Promise<any>
  Progress(sessionId: string): Promise<any>
  GlobalProgress(): Promise<any>
  PauseInstall(sessionId: string): Promise<void>
  ResumeInstall(sessionId: string): Promise<void>
  CancelInstall(sessionId: string): Promise<void>
  RunningInstances(): Promise<any[]>
  LatestCrash(): Promise<any>
  Sessions(): Promise<any[]>
  RecoverySessions(): Promise<any>
  InstallModLoader(req: any): Promise<{ success: boolean;
 error?: string }>
  DownloadRuntime(version: string, instancePath: string, sharedPath?: string): Promise<void>
  Worlds(): Promise<any>
  GetInfo(): Promise<{
    version: string
    jarPath: string
    authlibPath: string
    instancesDir: string
    versionsDir: string
    baseDir: string
    javaPath: string
  }>
  OnEvent(callback: (data: { event: string;
 data: unknown }) => void): () => void
  GetDownloadedVersions(): Promise<string[]>
  GetLogFiles(): Promise<{ name: string;
 size: number;
 mtime: string }[]>
  ReadLogFile(fileName: string, maxLines?: number): Promise<string[]>
  OnEngineLog(callback: (line: string) => void): () => void
}

interface InstancesManagerAPI {
  GetDir(): Promise<string>
  GetSharedDir(): Promise<string>
  List(): Promise<any[]>
  Get(instanceId: string): Promise<any>
  UpdateConfig(instanceId: string, updates: Record<string, any>): Promise<{ success: boolean;
 error?: string }>
  Delete(instanceId: string): Promise<{ success: boolean;
 error?: string }>
}

interface ThemeInfo {
  name: string
  author: string
  homePage: string
  version: string
  thumbnail: string
  path: string
}

interface ThemeManagerAPI {
  Export(metadata: { name: string; author: string; homePage: string; version: string }): Promise<{ success: boolean; theme?: ThemeInfo; error?: string }>
  Import(): Promise<{ success: boolean; theme?: ThemeInfo; error?: string; canceled?: boolean }>
  List(): Promise<{ success: boolean; themes?: ThemeInfo[]; error?: string }>
  Delete(name: string): Promise<{ success: boolean; error?: string }>
  GetConfig(name: string): Promise<{ success: boolean; config?: PersonalizationConfig; error?: string }>
  GetGallery(name: string): Promise<{ success: boolean; gallery?: string[]; error?: string }>
}

interface ModsManagerAPI {
  DownloadFile(url: string, suggestedName: string): Promise<{ success: boolean; filePath?: string; error?: string; canceled?: boolean }>
  DownloadToPath(url: string, destPath: string): Promise<{ success: boolean; filePath?: string; error?: string }>
  GetGameDir(): Promise<string>
  GetBaseDir(): Promise<string>
  DownloadToGameDir(contentType: string, url: string, filename: string, worldName?: string): Promise<{ success: boolean; filePath?: string; error?: string }>
  GetWorlds(): Promise<{ success: boolean; worlds?: string[]; error?: string }>
  InstallModpack(url: string, filename: string): Promise<{ success: boolean; gameDir?: string; error?: string }>
}

interface CacheManagerAPI {
  Get(key: string): Promise<unknown>
  Set(key: string, data: unknown, ttl?: number): Promise<void>
  Remove(key: string): Promise<void>
  Clear(): Promise<void>
  Size(): Promise<number>
  GetImage(url: string): Promise<string>
  PrefetchImage(url: string): Promise<void>
}

interface Window {
  AuthManager: AuthManagerAPI
  ConfigManager: ConfigManagerAPI
  ElectronAPI: ElectronAPI
  DownloadManager: DownloadManagerAPI
  NovaCoreManager: NovaCoreManagerAPI
  InstancesManager: InstancesManagerAPI
  ThemeManager: ThemeManagerAPI
  ModsManager: ModsManagerAPI
  CacheManager: CacheManagerAPI
  __splashDone?: boolean
}
