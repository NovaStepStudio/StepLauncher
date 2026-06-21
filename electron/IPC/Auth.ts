import { ipcMain, shell } from "electron"
import * as AuthStore from "../Stores/Auth.js"
import { RegisterAppIpc } from "./App.js"

let oauthResolve: ((data: { code: string; state: string }) => void) | null = null
let pendingOAuthUrl: string | null = null
let oauthTimeout: ReturnType<typeof setTimeout> | null = null

export function HandleOAuthCallback(url: string) {
  if (oauthResolve) {
    try {
      const parsed = new URL(url)
      const code = parsed.searchParams.get("code")
      const state = parsed.searchParams.get("state") ?? ""
      if (code) {
        if (oauthTimeout) clearTimeout(oauthTimeout)
        oauthResolve({ code, state })
        oauthResolve = null
      }
    } catch {}
  } else {
    pendingOAuthUrl = url
  }
}

export function RegisterAuthIpc() {
  RegisterAppIpc()

  ipcMain.handle("authManager:Initialize", async () => AuthStore.Initialize())

  ipcMain.handle("authManager:Login", async (_event, username: string, password: string) => {
    const res = await AuthStore.Login(username, password)
    const me = await AuthStore.FetchMe()
    return { token: res.token, expires_at: res.expires_at, expires_in: res.expires_in, success: true, user: me }
  })

  ipcMain.handle("authManager:Logout", () => AuthStore.Logout())
  ipcMain.handle("authManager:RestoreSession", () => AuthStore.RestoreSession())
  ipcMain.handle("authManager:GetToken", () => AuthStore.GetStoredToken())
  ipcMain.handle("authManager:GetMe", async () => AuthStore.FetchMe())

  ipcMain.handle("authManager:StartOAuth", async () => {
    const authUrl = await AuthStore.StartOAuthFlow()
    await shell.openExternal(authUrl)
    return new Promise<{ code: string; state: string }>((resolve) => {
      oauthResolve = resolve
      if (pendingOAuthUrl) {
        HandleOAuthCallback(pendingOAuthUrl)
        pendingOAuthUrl = null
      }
      oauthTimeout = setTimeout(() => {
        if (oauthResolve === resolve) {
          oauthResolve = null
          resolve({ code: "", state: "" })
        }
      }, 300_000)
    })
  })

  ipcMain.handle("authManager:CompleteOAuth", async (_event, code: string, state: string) => {
    const res = await AuthStore.CompleteOAuth(code, state)
    const me = await AuthStore.FetchMe()
    return { access_token: res.access_token, token_type: res.token_type, expires_in: res.expires_in, user: me }
  })

  ipcMain.handle("authManager:UpdateStats", async (_event, stats) => { await AuthStore.UpdateStats(stats) })
  ipcMain.handle("authManager:UpdateHeartbeat", async (_event, version: string, playingTimeIncrement?: number) => { await AuthStore.UpdateHeartbeat(version, playingTimeIncrement) })

  ipcMain.handle("authManager:GetPublicProfile", async (_event, uuid: string) => {
    if (!uuid) throw new Error("uuid is required")
    return AuthStore.GetPublicProfile(uuid)
  })

  ipcMain.handle("authManager:GetFullPublicProfile", async (_event, uuid: string) => {
    if (!uuid) throw new Error("uuid is required")
    return AuthStore.GetFullPublicProfile(uuid)
  })

  ipcMain.handle("authManager:GetSkins", async () => AuthStore.GetSkins())
  ipcMain.handle("authManager:GetCapes", async () => AuthStore.GetCapes())
  ipcMain.handle("authManager:GetKits", async () => AuthStore.GetKits())

  ipcMain.handle("authManager:GetFriends", async () => AuthStore.GetFriends())
  ipcMain.handle("authManager:GetFriendRequests", async () => AuthStore.GetFriendRequests())
  ipcMain.handle("authManager:SearchUsers", async (_event, q: string) => AuthStore.SearchUsers(q))
  ipcMain.handle("authManager:SendFriendRequest", async (_event, toUserId: number) => { await AuthStore.SendFriendRequest(toUserId) })
  ipcMain.handle("authManager:AcceptFriendRequest", async (_event, requestId: number) => { await AuthStore.AcceptFriendRequest(requestId) })
  ipcMain.handle("authManager:RejectFriendRequest", async (_event, requestId: number) => { await AuthStore.RejectFriendRequest(requestId) })
  ipcMain.handle("authManager:FollowUser", async (_event, uuid: string) => { await AuthStore.FollowUser(uuid) })
  ipcMain.handle("authManager:UnfollowUser", async (_event, uuid: string) => { await AuthStore.UnfollowUser(uuid) })
  ipcMain.handle("authManager:LikeProfile", async (_event, uuid: string) => { await AuthStore.LikeProfile(uuid) })
  ipcMain.handle("authManager:UnlikeProfile", async (_event, uuid: string) => { await AuthStore.UnlikeProfile(uuid) })

  ipcMain.handle("authManager:GetPresence", async (_event, uuid: string) => {
    if (!uuid) throw new Error("uuid is required")
    return AuthStore.GetPresence(uuid)
  })

  ipcMain.handle("authManager:SetPresence", async (_event, uuid: string, presence: string, message?: string) => { await AuthStore.SetPresence(uuid, presence, message) })

  ipcMain.handle("authManager:GetOAuthConnections", async () => AuthStore.GetOAuthConnections())
  ipcMain.handle("authManager:GetHealth", async () => {
    try { return await AuthStore.GetHealth() }
    catch { return null }
  })
  ipcMain.handle("authManager:GetLauncherToken", async (_event, launcherId: string) => AuthStore.GetLauncherToken(launcherId))

  ipcMain.handle("authManager:GetSocialLinks", async () => AuthStore.GetSocialLinks())
  ipcMain.handle("authManager:AddSocialLink", async (_event, platform: string, url: string) => AuthStore.AddSocialLink(platform, url))
  ipcMain.handle("authManager:UpdateSocialLink", async (_event, id: number, url: string) => { await AuthStore.UpdateSocialLink(id, url) })
  ipcMain.handle("authManager:DeleteSocialLink", async (_event, id: number) => { await AuthStore.DeleteSocialLink(id) })

  ipcMain.handle("authManager:GetLibraries", async () => AuthStore.GetLibraries())
  ipcMain.handle("authManager:CreateLibrary", async (_event, title: string, description: string | null, isPublic: boolean, config?: Record<string, unknown>) => AuthStore.CreateLibrary(title, description, isPublic, config))
  ipcMain.handle("authManager:GetLibrary", async (_event, id: number) => AuthStore.GetLibrary(id))
  ipcMain.handle("authManager:UpdateLibrary", async (_event, id: number, updates) => { await AuthStore.UpdateLibrary(id, updates) })
  ipcMain.handle("authManager:DeleteLibrary", async (_event, id: number) => { await AuthStore.DeleteLibrary(id) })
  ipcMain.handle("authManager:GetEntries", async (_event, libraryId: number) => AuthStore.GetEntries(libraryId))
  ipcMain.handle("authManager:AddEntry", async (_event, libraryId: number, entry: Record<string, unknown>) => AuthStore.AddEntry(libraryId, entry))
  ipcMain.handle("authManager:UpdateEntry", async (_event, libraryId: number, entryId: number, updates: Record<string, unknown>) => { await AuthStore.UpdateEntry(libraryId, entryId, updates) })
  ipcMain.handle("authManager:RemoveEntry", async (_event, libraryId: number, entryId: number) => { await AuthStore.RemoveEntry(libraryId, entryId) })
  ipcMain.handle("authManager:ToggleLibraryLike", async (_event, libraryId: number) => AuthStore.ToggleLibraryLike(libraryId))
  ipcMain.handle("authManager:GetPublicLibraryEntries", async (_event, libraryId: number) => AuthStore.GetPublicLibraryEntries(libraryId))
  ipcMain.handle("authManager:CloneLibrary", async (_event, sourceLibraryId: number, customTitle?: string) => AuthStore.CloneLibrary(sourceLibraryId, customTitle))
}
