import { contextBridge, ipcRenderer } from "electron"

contextBridge.exposeInMainWorld("AuthManager", {
  Initialize: () => ipcRenderer.invoke("authManager:Initialize"),
  Login: (username: string, password: string) => ipcRenderer.invoke("authManager:Login", username, password),
  Logout: () => ipcRenderer.invoke("authManager:Logout"),
  RestoreSession: () => ipcRenderer.invoke("authManager:RestoreSession"),
  StartOAuth: () => ipcRenderer.invoke("authManager:StartOAuth"),
  CompleteOAuth: (code: string, state: string) => ipcRenderer.invoke("authManager:CompleteOAuth", code, state),
  GetToken: () => ipcRenderer.invoke("authManager:GetToken"),
  GetMe: () => ipcRenderer.invoke("authManager:GetMe"),
  UpdateStats: (stats: Record<string, unknown>) => ipcRenderer.invoke("authManager:UpdateStats", stats),
  UpdateHeartbeat: (version: string, playingTimeIncrement?: number) => ipcRenderer.invoke("authManager:UpdateHeartbeat", version, playingTimeIncrement),
  GetPublicProfile: (uuid: string) => ipcRenderer.invoke("authManager:GetPublicProfile", uuid),
  GetFullPublicProfile: (uuid: string) => ipcRenderer.invoke("authManager:GetFullPublicProfile", uuid),
  GetSkins: () => ipcRenderer.invoke("authManager:GetSkins"),
  GetCapes: () => ipcRenderer.invoke("authManager:GetCapes"),
  GetKits: () => ipcRenderer.invoke("authManager:GetKits"),
  GetFriends: () => ipcRenderer.invoke("authManager:GetFriends"),
  GetFriendRequests: () => ipcRenderer.invoke("authManager:GetFriendRequests"),
  SearchUsers: (q: string) => ipcRenderer.invoke("authManager:SearchUsers", q),
  SendFriendRequest: (toUserId: number) => ipcRenderer.invoke("authManager:SendFriendRequest", toUserId),
  AcceptFriendRequest: (requestId: number) => ipcRenderer.invoke("authManager:AcceptFriendRequest", requestId),
  RejectFriendRequest: (requestId: number) => ipcRenderer.invoke("authManager:RejectFriendRequest", requestId),
  FollowUser: (uuid: string) => ipcRenderer.invoke("authManager:FollowUser", uuid),
  UnfollowUser: (uuid: string) => ipcRenderer.invoke("authManager:UnfollowUser", uuid),
  LikeProfile: (uuid: string) => ipcRenderer.invoke("authManager:LikeProfile", uuid),
  UnlikeProfile: (uuid: string) => ipcRenderer.invoke("authManager:UnlikeProfile", uuid),
  GetPresence: (uuid: string) => ipcRenderer.invoke("authManager:GetPresence", uuid),
  SetPresence: (uuid: string, presence: string, message?: string) => ipcRenderer.invoke("authManager:SetPresence", uuid, presence, message),
  GetOAuthConnections: () => ipcRenderer.invoke("authManager:GetOAuthConnections"),
  GetHealth: () => ipcRenderer.invoke("authManager:GetHealth"),
  GetLauncherToken: (launcherId: string) => ipcRenderer.invoke("authManager:GetLauncherToken", launcherId),
  GetSocialLinks: () => ipcRenderer.invoke("authManager:GetSocialLinks"),
  AddSocialLink: (platform: string, url: string) => ipcRenderer.invoke("authManager:AddSocialLink", platform, url),
  UpdateSocialLink: (id: number, url: string) => ipcRenderer.invoke("authManager:UpdateSocialLink", id, url),
  DeleteSocialLink: (id: number) => ipcRenderer.invoke("authManager:DeleteSocialLink", id),
  GetLibraries: () => ipcRenderer.invoke("authManager:GetLibraries"),
  CreateLibrary: (title: string, description: string | null, isPublic: boolean, config?: Record<string, unknown>) => ipcRenderer.invoke("authManager:CreateLibrary", title, description, isPublic, config),
  GetLibrary: (id: number) => ipcRenderer.invoke("authManager:GetLibrary", id),
  UpdateLibrary: (id: number, updates: Record<string, unknown>) => ipcRenderer.invoke("authManager:UpdateLibrary", id, updates),
  DeleteLibrary: (id: number) => ipcRenderer.invoke("authManager:DeleteLibrary", id),
  GetEntries: (libraryId: number) => ipcRenderer.invoke("authManager:GetEntries", libraryId),
  AddEntry: (libraryId: number, entry: Record<string, unknown>) => ipcRenderer.invoke("authManager:AddEntry", libraryId, entry),
  UpdateEntry: (libraryId: number, entryId: number, updates: Record<string, unknown>) => ipcRenderer.invoke("authManager:UpdateEntry", libraryId, entryId, updates),
  RemoveEntry: (libraryId: number, entryId: number) => ipcRenderer.invoke("authManager:RemoveEntry", libraryId, entryId),
  ToggleLibraryLike: (libraryId: number) => ipcRenderer.invoke("authManager:ToggleLibraryLike", libraryId),
  GetPublicLibraryEntries: (libraryId: number) => ipcRenderer.invoke("authManager:GetPublicLibraryEntries", libraryId),
  CloneLibrary: (sourceLibraryId: number, customTitle?: string) => ipcRenderer.invoke("authManager:CloneLibrary", sourceLibraryId, customTitle),
})

contextBridge.exposeInMainWorld("ConfigManager", {
  Get: () => ipcRenderer.invoke("config:Get"),
  UpdateLauncher: (updates: Record<string, unknown>) => ipcRenderer.invoke("config:UpdateLauncher", updates),
  UpdateMinecraft: (updates: Record<string, unknown>) => ipcRenderer.invoke("config:UpdateMinecraft", updates),
  UpdatePersonalization: (updates: Record<string, unknown>) => ipcRenderer.invoke("config:UpdatePersonalization", updates),
  GetSystemInfo: () => ipcRenderer.invoke("config:GetSystemInfo"),
})

contextBridge.exposeInMainWorld("ElectronAPI", {
  OpenExternal: (url: string) => ipcRenderer.invoke("app:OpenExternal", url),
  OpenFileDialog: (options?: { filters?: { name: string; extensions: string[] }[] }) => ipcRenderer.invoke("app:OpenFileDialog", options),
  OpenPath: (dirPath: string) => ipcRenderer.invoke("app:OpenPath", dirPath),
  ReadDir: (dirPath: string) => ipcRenderer.invoke("app:ReadDir", dirPath),
  Hide: () => ipcRenderer.invoke("app:Hide"),
  Show: () => ipcRenderer.invoke("app:Show"),
  Minimize: () => ipcRenderer.invoke("app:Minimize"),
  Maximize: () => ipcRenderer.invoke("app:Maximize"),
  Close: () => ipcRenderer.invoke("app:Close"),
  ShowNotification: (options: { title: string; body: string }) => ipcRenderer.invoke("app:ShowNotification", options),
  FetchJson: (url: string) => ipcRenderer.invoke("app:FetchJson", url),
  ReadLocaleFile: (locale: string) => ipcRenderer.invoke("app:ReadLocaleFile", locale),
})

contextBridge.exposeInMainWorld("DownloadManager", {
  CheckAll: () => ipcRenderer.invoke("downloads:CheckAll"),
  StartAll: () => ipcRenderer.invoke("downloads:StartAll"),
  OnProgress: (callback: (progress: unknown) => void) => {
    const handler = (_event: Electron.IpcRendererEvent, progress: unknown) => callback(progress)
    ipcRenderer.on("downloads:Progress", handler)
    return () => ipcRenderer.removeListener("downloads:Progress", handler)
  },
})

contextBridge.exposeInMainWorld("NovaCoreManager", {
  Start: () => ipcRenderer.invoke("novacore:Start"),
  Stop: () => ipcRenderer.invoke("novacore:Stop"),
  Status: () => ipcRenderer.invoke("novacore:Status"),
  Health: () => ipcRenderer.invoke("novacore:Health"),
  Versions: (limit?: number) => ipcRenderer.invoke("novacore:Versions", limit),
  Install: (req: unknown) => ipcRenderer.invoke("novacore:Install", req),
  Launch: (req: unknown) => ipcRenderer.invoke("novacore:Launch", req),
  KillInstance: (id: string) => ipcRenderer.invoke("novacore:KillInstance", id),
  EngineInfo: () => ipcRenderer.invoke("novacore:EngineInfo"),
  ModLoaders: () => ipcRenderer.invoke("novacore:ModLoaders"),
  ModLoaderVersions: (loader: string, mcVersion: string) => ipcRenderer.invoke("novacore:ModLoaderVersions", loader, mcVersion),
  Progress: (sessionId: string) => ipcRenderer.invoke("novacore:Progress", sessionId),
  GlobalProgress: () => ipcRenderer.invoke("novacore:GlobalProgress"),
  PauseInstall: (sessionId: string) => ipcRenderer.invoke("novacore:PauseInstall", sessionId),
  ResumeInstall: (sessionId: string) => ipcRenderer.invoke("novacore:ResumeInstall", sessionId),
  CancelInstall: (sessionId: string) => ipcRenderer.invoke("novacore:CancelInstall", sessionId),
  RunningInstances: () => ipcRenderer.invoke("novacore:RunningInstances"),
  LatestCrash: () => ipcRenderer.invoke("novacore:LatestCrash"),
  Sessions: () => ipcRenderer.invoke("novacore:Sessions"),
  RecoverySessions: () => ipcRenderer.invoke("novacore:RecoverySessions"),
  InstallModLoader: (req: unknown) => ipcRenderer.invoke("novacore:InstallModLoader", req),
  DownloadRuntime: (version: string, instancePath: string, sharedPath?: string) => ipcRenderer.invoke("novacore:DownloadRuntime", version, instancePath, sharedPath),
  Worlds: () => ipcRenderer.invoke("novacore:Worlds"),
  GetInfo: () => ipcRenderer.invoke("novacore:GetInfo"),
  OnEvent: (callback: (data: { event: string; data: unknown }) => void) => {
    const handler = (_event: Electron.IpcRendererEvent, data: { event: string; data: unknown }) => callback(data)
    ipcRenderer.on("novacore:Event", handler)
    return () => ipcRenderer.removeListener("novacore:Event", handler)
  },
  GetDownloadedVersions: () => ipcRenderer.invoke("novacore:GetDownloadedVersions"),
  GetLogFiles: () => ipcRenderer.invoke("novacore:GetLogFiles"),
  ReadLogFile: (fileName: string, maxLines?: number) => ipcRenderer.invoke("novacore:ReadLogFile", fileName, maxLines),
  OnEngineLog: (callback: (line: string) => void) => {
    const handler = (_event: Electron.IpcRendererEvent, data: { event: string; data: unknown }) => {
      if (data.event === "engine_log") callback((data.data as any).line ?? "")
    }
    ipcRenderer.on("novacore:Event", handler)
    return () => ipcRenderer.removeListener("novacore:Event", handler)
  },
})

contextBridge.exposeInMainWorld("ThemeManager", {
  Export: (metadata: { name: string; author: string; homePage: string; version: string }) => ipcRenderer.invoke("theme:Export", metadata),
  Import: () => ipcRenderer.invoke("theme:Import"),
  List: () => ipcRenderer.invoke("theme:List"),
  Delete: (name: string) => ipcRenderer.invoke("theme:Delete", name),
  GetConfig: (name: string) => ipcRenderer.invoke("theme:GetConfig", name),
  GetGallery: (name: string) => ipcRenderer.invoke("theme:GetGallery", name),
})

contextBridge.exposeInMainWorld("ModsManager", {
  DownloadFile: (url: string, suggestedName: string) => ipcRenderer.invoke("mods:DownloadFile", url, suggestedName),
  DownloadToPath: (url: string, destPath: string) => ipcRenderer.invoke("mods:DownloadToPath", url, destPath),
  GetGameDir: () => ipcRenderer.invoke("mods:GetGameDir"),
  GetBaseDir: () => ipcRenderer.invoke("mods:GetBaseDir"),
  DownloadToGameDir: (contentType: string, url: string, filename: string, worldName?: string) => ipcRenderer.invoke("mods:DownloadToGameDir", contentType, url, filename, worldName),
  GetWorlds: () => ipcRenderer.invoke("mods:GetWorlds"),
  InstallModpack: (url: string, filename: string) => ipcRenderer.invoke("mods:InstallModpack", url, filename),
})

contextBridge.exposeInMainWorld("InstancesManager", {
  GetDir: () => ipcRenderer.invoke("instances:GetDir"),
  GetSharedDir: () => ipcRenderer.invoke("instances:GetSharedDir"),
  List: () => ipcRenderer.invoke("instances:List"),
  Get: (instanceId: string) => ipcRenderer.invoke("instances:Get", instanceId),
  UpdateConfig: (instanceId: string, updates: Record<string, any>) => ipcRenderer.invoke("instances:UpdateConfig", instanceId, updates),
  Delete: (instanceId: string) => ipcRenderer.invoke("instances:Delete", instanceId),
})

contextBridge.exposeInMainWorld("CacheManager", {
  Get: (key: string) => ipcRenderer.invoke("cache:Get", key),
  Set: (key: string, data: unknown, ttl?: number) => ipcRenderer.invoke("cache:Set", key, data, ttl),
  Remove: (key: string) => ipcRenderer.invoke("cache:Remove", key),
  Clear: () => ipcRenderer.invoke("cache:Clear"),
  Size: () => ipcRenderer.invoke("cache:Size"),
  GetImage: (url: string) => ipcRenderer.invoke("cache:GetImage", url),
  PrefetchImage: (url: string) => ipcRenderer.invoke("cache:PrefetchImage", url),
})
