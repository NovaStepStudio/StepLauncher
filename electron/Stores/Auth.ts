import { app } from "electron"
import { readFileSync, writeFileSync, mkdirSync, existsSync } from "fs"
import { join } from "path"
import { api } from "../Services/Api.js"
import {
  GenerateCodeVerifier,
  GenerateCodeChallenge,
  GenerateState,
} from "../Utils/Pkce.js"
import type {
  LoginResponse,
  MeResponse,
  OAuthTokenResponse,
  PublicProfileResponse,
  Presence,
  PublicProfile,
  OAuthConnection,
  SocialLink,
  Library,
  LibraryEntry,
  LibraryLikeResponse,
  LauncherTokenResponse,
} from "../Types/Auth.js"

let currentUser: MeResponse | null = null
let oauthState: string | null = null
let oauthVerifier: string | null = null

function GetDataPath(): string {
  try {
    const dir = join(app.getPath("userData"), "data")
    if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
    return dir
  } catch {
    return ""
  }
}

function PersistToken(token: string) {
  api.SetToken(token)
  try {
    const dir = GetDataPath()
    if (dir) writeFileSync(join(dir, "token"), token, "utf-8")
  } catch {}
}

function PersistUser(user: MeResponse) {
  currentUser = user
  try {
    const dir = GetDataPath()
    if (dir) writeFileSync(join(dir, "user.json"), JSON.stringify(user), "utf-8")
  } catch {}
}

function LoadToken(): string | null {
  try {
    const dir = GetDataPath()
    if (!dir) return null
    const path = join(dir, "token")
    if (!existsSync(path)) return null
    return readFileSync(path, "utf-8")
  } catch {
    return null
  }
}

function LoadUser(): MeResponse | null {
  try {
    const dir = GetDataPath()
    if (!dir) return null
    const path = join(dir, "user.json")
    if (!existsSync(path)) return null
    return JSON.parse(readFileSync(path, "utf-8")) as MeResponse
  } catch {
    return null
  }
}

function ClearStorage() {
  try {
    const dir = GetDataPath()
    if (!dir) return
    const tokenPath = join(dir, "token")
    const userPath = join(dir, "user.json")
    if (existsSync(tokenPath)) writeFileSync(tokenPath, "", "utf-8")
    if (existsSync(userPath)) writeFileSync(userPath, "", "utf-8")
  } catch {}
}

export function GetStoredToken(): string | null {
  return LoadToken()
}

export function GetCurrentUser(): MeResponse | null {
  return currentUser
}

export async function Initialize(): Promise<{
  authenticated: boolean
  user: MeResponse | null
  token: string | null
}> {
  const savedToken = LoadToken()
  if (!savedToken) return { authenticated: false, user: null, token: null }

  api.SetToken(savedToken)
  const savedUser = LoadUser()

  try {
    const me = await api.Get<MeResponse>("/me")
    PersistUser(me)
    return { authenticated: true, user: me, token: savedToken }
  } catch {
    if (savedUser) {
      currentUser = savedUser
      return { authenticated: true, user: savedUser, token: savedToken }
    }
    api.SetToken(null)
    return { authenticated: false, user: null, token: null }
  }
}

export async function Login(username: string, password: string, source = "launcher"): Promise<LoginResponse> {
  const res = await api.Post<LoginResponse>("/login", { username, password, source })
  PersistToken(res.token)
  return res
}

export async function FetchMe(): Promise<MeResponse> {
  const me = await api.Get<MeResponse>("/me")
  PersistUser(me)
  return me
}

export async function StartOAuthFlow(): Promise<string> {
  const verifier = GenerateCodeVerifier()
  const challenge = await GenerateCodeChallenge(verifier)
  const state = GenerateState()

  oauthState = state
  oauthVerifier = verifier

  const params = new URLSearchParams({
    client_id: "steplauncher",
    redirect_uri: "steplauncher://callback",
    response_type: "code",
    code_challenge: challenge,
    code_challenge_method: "S256",
    state,
  })

  return `https://steplauncher.pages.dev/oauth/steplauncher/authorize?${params}`
}

export async function CompleteOAuth(code: string, expectedState: string): Promise<OAuthTokenResponse> {
  if (oauthState && expectedState && expectedState !== oauthState) {
    throw new Error("State mismatch. Possible CSRF attack.")
  }

  if (!oauthVerifier) throw new Error("No PKCE verifier found. Start OAuth flow again.")

  const verifier = oauthVerifier
  oauthState = null
  oauthVerifier = null

  const res = await api.Post<OAuthTokenResponse>("/oauth/steplauncher/token", {
    code,
    client_id: "steplauncher",
    grant_type: "authorization_code",
    code_verifier: verifier,
  })

  PersistToken(res.access_token)
  return res
}

export function Logout() {
  currentUser = null
  api.SetToken(null)
  oauthState = null
  oauthVerifier = null
  ClearStorage()
}

export async function RestoreSession(): Promise<boolean> {
  const savedToken = LoadToken()
  if (savedToken) {
    api.SetToken(savedToken)
    const savedUser = LoadUser()
    if (savedUser) currentUser = savedUser
    return true
  }
  return false
}

export async function UpdateStats(stats: Record<string, unknown>): Promise<void> {
  try { await api.Patch("/me/stats", stats) } catch {}
}

export async function UpdateHeartbeat(version: string, playingTimeIncrement?: number): Promise<void> {
  const body: Record<string, unknown> = { last_version_connected: version }
  if (playingTimeIncrement !== undefined) body.playing_time_increment = playingTimeIncrement
  await api.Patch("/me/stats", body)
}

export async function GetPublicProfile(uuid: string): Promise<PublicProfileResponse> {
  return api.Get<PublicProfileResponse>(`/profile/${uuid}`)
}

export async function GetFullPublicProfile(uuid: string): Promise<PublicProfile> {
  return api.Get<PublicProfile>(`/profile/${uuid}`)
}

export async function GetSkins(): Promise<any[]> {
  return api.Get<any[]>("/me/skins")
}

export async function GetCapes(): Promise<any[]> {
  return api.Get<any[]>("/me/capes")
}

export async function GetKits(): Promise<any[]> {
  return api.Get<any[]>("/me/kits")
}

export async function GetFriends(): Promise<any[]> {
  return api.Get<any[]>("/me/friends")
}

export async function GetFriendRequests(): Promise<{ received: any[]; sent: any[] }> {
  const [received, sent] = await Promise.all([
    api.Get<any[]>("/me/friends/requests/received"),
    api.Get<any[]>("/me/friends/requests/sent"),
  ])
  return { received, sent }
}

export async function SearchUsers(q: string): Promise<any[]> {
  return api.Get<any[]>(`/users/search?q=${encodeURIComponent(q)}`)
}

export async function SendFriendRequest(toUserId: number): Promise<void> {
  await api.Post("/me/friends/request", { to_user_id: toUserId })
}

export async function AcceptFriendRequest(requestId: number): Promise<void> {
  await api.Post(`/me/friends/requests/${requestId}/accept`)
}

export async function RejectFriendRequest(requestId: number): Promise<void> {
  await api.Post(`/me/friends/requests/${requestId}/reject`)
}

export async function FollowUser(uuid: string): Promise<void> {
  await api.Post(`/me/follow/${uuid}`)
}

export async function UnfollowUser(uuid: string): Promise<void> {
  await api.Delete(`/me/follow/${uuid}`)
}

export async function LikeProfile(uuid: string): Promise<void> {
  await api.Post(`/me/like/${uuid}`)
}

export async function UnlikeProfile(uuid: string): Promise<void> {
  await api.Delete(`/me/like/${uuid}`)
}

export async function GetPresence(uuid: string): Promise<Presence> {
  return api.Get<Presence>(`/presence/${uuid}`)
}

export async function SetPresence(uuid: string, presence: string, message?: string): Promise<void> {
  const token = api.GetToken()
  await api.Post("/presence", { uuid, accessToken: token, presence, message })
}

export async function GetOAuthConnections(): Promise<OAuthConnection[]> {
  return api.Get<OAuthConnection[]>("/me/oauth/connections")
}

export async function GetHealth(): Promise<any> {
  return api.Get("/health")
}

export async function GetLauncherToken(launcherId: string): Promise<LauncherTokenResponse> {
  return api.Post<LauncherTokenResponse>("/launcher/token", { source: "launcher", launcher_id: launcherId })
}

export async function GetSocialLinks(): Promise<SocialLink[]> {
  return api.Get<SocialLink[]>("/me/social")
}

export async function AddSocialLink(platform: string, url: string): Promise<SocialLink> {
  return api.Post<SocialLink>("/me/social", { platform, url })
}

export async function UpdateSocialLink(id: number, url: string): Promise<void> {
  await api.Patch(`/me/social/${id}`, { url })
}

export async function DeleteSocialLink(id: number): Promise<void> {
  await api.Delete(`/me/social/${id}`)
}

export async function GetLibraries(): Promise<Library[]> {
  return api.Get<Library[]>("/me/library")
}

export async function CreateLibrary(title: string, description: string | null, isPublic: boolean, config?: Record<string, unknown>): Promise<Library> {
  return api.Post<Library>("/me/library", { title, description, is_public: isPublic, config })
}

export async function GetLibrary(id: number): Promise<Library> {
  return api.Get<Library>(`/me/library/${id}`)
}

export async function UpdateLibrary(id: number, updates: Partial<Pick<Library, "title" | "description" | "is_public" | "config">>): Promise<void> {
  await api.Patch(`/me/library/${id}`, updates)
}

export async function DeleteLibrary(id: number): Promise<void> {
  await api.Delete(`/me/library/${id}`)
}

export async function GetEntries(libraryId: number): Promise<LibraryEntry[]> {
  return api.Get<LibraryEntry[]>(`/me/library/${libraryId}/entries`)
}

export async function AddEntry(libraryId: number, entry: Record<string, unknown>): Promise<LibraryEntry> {
  return api.Post<LibraryEntry>(`/me/library/${libraryId}/entries`, entry)
}

export async function UpdateEntry(libraryId: number, entryId: number, updates: Record<string, unknown>): Promise<void> {
  await api.Patch(`/me/library/${libraryId}/entries/${entryId}`, updates)
}

export async function RemoveEntry(libraryId: number, entryId: number): Promise<void> {
  await api.Delete(`/me/library/${libraryId}/entries/${entryId}`)
}

export async function ToggleLibraryLike(libraryId: number): Promise<LibraryLikeResponse> {
  return api.Post<LibraryLikeResponse>(`/me/library/${libraryId}/like`)
}

export async function GetPublicLibraryEntries(libraryId: number): Promise<LibraryEntry[]> {
  return api.Get<LibraryEntry[]>(`/library/${libraryId}/public-entries`)
}

export async function CloneLibrary(sourceLibraryId: number, customTitle?: string): Promise<Library> {
  return api.Post<Library>("/me/library/clone", { source_library_id: sourceLibraryId, title: customTitle ?? null })
}
