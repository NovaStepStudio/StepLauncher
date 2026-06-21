import { ref, readonly } from 'vue'
import { useUserCache } from './useUserCache'

const { prefetchImages } = useUserCache()

function prefetchUserImages(u: AuthUser | null) {
  if (!u) return
  const urls: string[] = []
  if (u.avatar) urls.push(u.avatar)
  if (u.skin_url) urls.push(u.skin_url)
  if (u.cape_url) urls.push(u.cape_url)
  if (u.banner_url) urls.push(u.banner_url)
  if (urls.length) prefetchImages(urls)
}

export interface AuthUser {
  uuid: string
  username: string
  premium: string
  avatar: string | null
  email: string | null
  bio: string | null
  is_admin: boolean
  is_online: boolean
  last_seen: string | null
  last_login: string | null
  created_at: string
  skin_url: string | null
  cape_url: string | null
  banner_url?: string | null
  settings: Record<string, unknown> | null
  playing_time: number
  last_version_playing: string | null
  last_version_connected: string | null
}

const initializing = ref(true)
const isAuthenticated = ref(false)
const isOffline = ref(false)
const user = ref<AuthUser | null>(null)
const token = ref<string | null>(null)

export function useAuth() {
  async function initialize() {
    try {
      const result = await window.AuthManager.Initialize()
      isAuthenticated.value = result.authenticated
      user.value = result.user as AuthUser | null
      token.value = result.token
      prefetchUserImages(result.user as AuthUser | null)
    } catch {
      isAuthenticated.value = false
      user.value = null
      token.value = null
    } finally {
      initializing.value = false
    }
  }

  async function login(username: string, password: string): Promise<string | null> {
    try {
      const res = await window.AuthManager.Login(username, password)
      if (res.success) {
        isAuthenticated.value = true
        isOffline.value = false
        user.value = res.user as AuthUser
        token.value = res.token
        prefetchUserImages(res.user as AuthUser)
        return null
      }
      return 'Credenciales inválidas'
    } catch (e: any) {
      return e?.message ?? 'Error al conectar'
    }
  }

  function loginOffline(username: string) {
    if (!username.trim()) return
    isAuthenticated.value = true
    isOffline.value = true
    user.value = {
      uuid: '',
      username: username.trim(),
      premium: 'free',
      avatar: null,
      email: null,
      bio: null,
      is_admin: false,
      is_online: false,
      last_seen: null,
      last_login: null,
      created_at: '',
      skin_url: null,
      cape_url: null,
      banner_url: null,
      settings: null,
      playing_time: 0,
      last_version_playing: null,
      last_version_connected: null,
    }
    token.value = null
  }

  async function logout() {
    try {
      await window.AuthManager.Logout()
    } catch {}
    isAuthenticated.value = false
    isOffline.value = false
    user.value = null
    token.value = null
  }

  return {
    initializing: readonly(initializing),
    isAuthenticated: readonly(isAuthenticated),
    isOffline: readonly(isOffline),
    user: readonly(user),
    token: readonly(token),
    initialize,
    login,
    loginOffline,
    logout,
  }
}
