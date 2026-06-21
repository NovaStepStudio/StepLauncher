export interface AuthUser {
  uuid: string
  username: string
  tier: string
  avatar: string | null
  email_verified?: boolean
}

export interface LoginResponse {
  token: string
  expires_at: number
  expires_in: number
  success: boolean
  user: AuthUser
}

export interface RegisterResponse {
  success: boolean
  user: { uuid: string; username: string }
}

export interface OAuthTokenResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

export interface MeResponse {
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
}

export interface PublicProfileResponse {
  uuid: string
  username: string
  bio: string | null
  premium: string
  is_online: boolean
  last_seen: string | null
  avatar: string | null
  skin_url: string | null
  cape_url: string | null
}

export interface ApiError {
  error: string
  detail?: string
}

export interface UserStatsBody {
  playing_time_increment?: number
  last_version_playing?: string | null
  last_version_connected?: string | null
}

export interface OAuthCallbackData {
  code: string
  state: string
}

export interface OAuthConnection {
  provider: string
  provider_id: string
  email: string | null
  created_at: string
}

export interface UserFull {
  uuid: string
  username: string
  password_hash: string
  salt: string
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
  playing_time: number
  last_version_playing: string | null
  last_version_connected: string | null
  settings: Record<string, unknown> | null
  email_verified?: boolean
}

export interface Presence {
  uuid: string
  presence: string
  message: string | null
  image_b64: string | null
  updated_at: string | null
}

export interface PublicProfile {
  user: {
    id?: number
    uuid: string
    username: string
    bio: string | null
    avatar: string | null
    skin_url: string | null
    cape_url: string | null
    premium: string
    created_at: string
    last_version_playing: string | null
    followers_count: number
    following_count: number
    likes_count: number
    is_online?: boolean
    last_seen?: string | null
    playing_time?: number
    is_admin?: boolean
    is_verified?: boolean
    last_launcher_version?: string | null
    private?: boolean
    message?: string
    banned?: boolean
  }
  is_following: boolean
  is_liked: boolean
  presence?: {
    presence: string
    message: string | null
    updated_at: string | null
  }
  social_links?: Array<{
    id: number
    platform: string
    url: string
    display_order: number
  }>
  common_friends?: Array<{ uuid: string; username: string; avatar?: string }>
  cosmetics?: {
    skins: any[]
    capes: any[]
    kits: any[]
  }
  libraries?: Library[]
}

export interface SocialLink {
  id: number
  platform: string
  url: string
  display_order: number
}

export interface Library {
  id: number
  user_id: number
  title: string
  description: string | null
  is_public: boolean
  config: Record<string, unknown>
  entries_count: number
  follows: number
  likes: number
  liked?: boolean
  created_at: string
  updated_at: string
  entries?: LibraryEntry[]
}

export interface LibraryEntry {
  id: number
  library_id: number
  project_id: string
  provider: string
  entry_type: string
  slug: string | null
  title: string
  description: string | null
  body: string | null
  icon: string | null
  color: string | null
  author_name: string | null
  author_team_id: string | null
  project_url: string | null
  issues_url: string | null
  source_url: string | null
  wiki_url: string | null
  discord_url: string | null
  donate_url: string | null
  favorite: boolean
  enabled: boolean
  hidden: boolean
  pinned: boolean
  archived: boolean
  notes: string | null
  compatibility: Record<string, unknown> | null
  categories: string[] | null
  tags: string[] | null
  added_at: string
  last_used_at: string | null
  last_sync_at: string | null
  published_at: string | null
  project_updated_at: string | null
}

export interface ApiResult<T = void> {
  success: boolean
  data?: T
  error?: string
}

export interface LauncherTokenResponse {
  launcher_token: string
  launcher_id: string
  expires_at: number
  expires_in: number
}

export interface LibraryLikeResponse {
  success: boolean
  liked: boolean
  likes: number
  error?: string
}
