import { ref, reactive } from 'vue'

const DEFAULT_TTL = 5 * 60 * 1000

const cm = (): CacheManagerAPI | null => {
  return window.CacheManager ?? null
}

const pendingRequests = new Map<string, Promise<unknown>>()
const resolvedImages = reactive(new Map<string, string>())

function dedupedFetch<T>(key: string, fetcher: () => Promise<T>): Promise<T> {
  const existing = pendingRequests.get(key)
  if (existing) return existing as Promise<T>
  const promise = fetcher().finally(() => pendingRequests.delete(key))
  pendingRequests.set(key, promise)
  return promise
}

interface CachedUser {
  uuid: string
  username: string
  avatar: string | null
  premium: string
  bio: string | null
  is_admin: boolean
  is_online: boolean
  last_seen: string | null
  skin_url: string | null
  cape_url: string | null
  banner_url: string | null
}

interface CachedFriend {
  uuid: string
  username: string
  avatar: string
  statusClass: 'online' | 'in_game' | 'idle' | 'busy' | 'offline'
  presenceLabel: string
}

export function useUserCache() {
  const cacheSize = ref(0)

  async function updateSize() {
    try {
      const mgr = cm()
      if (mgr) cacheSize.value = await mgr.Size()
      else cacheSize.value = 0
    } catch {}
  }

  async function getUser(uuid: string): Promise<CachedUser | null> {
    try {
      const mgr = cm()
      if (!mgr) return null
      return await mgr.Get(`user_${uuid}`) as CachedUser | null
    } catch {
      return null
    }
  }

  const IMAGE_FIELDS = ['avatar', 'skin_url', 'cape_url', 'banner_url'] as const

  async function setUser(uuid: string, user: Partial<CachedUser>, ttl?: number): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      const existing = await getUser(uuid) || {} as CachedUser
      const changedUrls: string[] = []
      for (const field of IMAGE_FIELDS) {
        const oldVal = existing[field]
        const newVal = (user as any)[field]
        if (newVal && newVal !== oldVal) {
          changedUrls.push(newVal)
        }
      }
      if (changedUrls.length) {
        prefetchImages(changedUrls)
      }
      await mgr.Set(`user_${uuid}`, { ...existing, ...user }, ttl ?? DEFAULT_TTL)
    } catch {}
  }

  async function getFriendList(): Promise<CachedFriend[] | null> {
    try {
      const mgr = cm()
      if (!mgr) return null
      return await mgr.Get('friends_list') as CachedFriend[] | null
    } catch {
      return null
    }
  }

  async function setFriendList(friends: CachedFriend[]): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      prefetchImages(friends.map(f => f.avatar).filter(Boolean))
      await mgr.Set('friends_list', friends, 60 * 1000)
    } catch {}
  }

  async function getSearchResults(query: string): Promise<CachedUser[] | null> {
    try {
      const mgr = cm()
      if (!mgr) return null
      return await mgr.Get(`search_${query.toLowerCase().trim()}`) as CachedUser[] | null
    } catch {
      return null
    }
  }

  async function setSearchResults(query: string, results: CachedUser[]): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      await mgr.Set(`search_${query.toLowerCase().trim()}`, results, 2 * 60 * 1000)
    } catch {}
  }

  function isLocalSrc(src: string): boolean {
    return src.startsWith('data:') || src.startsWith('assets/') || src.startsWith('blob:')
  }

  async function getImageUrl(url: string): Promise<string> {
    if (!url || isLocalSrc(url)) return url

    const cached = resolvedImages.get(url)
    if (cached) return cached

    try {
      const mgr = cm()
      if (!mgr) return url
      const dataUrl = await dedupedFetch(url, async () => {
        return await mgr!.GetImage(url) as string
      })
      if (!dataUrl) return url
      resolvedImages.set(url, dataUrl)
      return dataUrl
    } catch {
      return url
    }
  }

  async function preloadImage(url: string): Promise<HTMLImageElement | null> {
    try {
      const src = await getImageUrl(url)
      if (!src) return null
      const img = new Image()
      await new Promise<void>((resolve, reject) => {
        img.onload = () => resolve()
        img.onerror = () => reject()
        img.src = src
      })
      return img
    } catch {
      return null
    }
  }

  function getCachedSrc(url: string): string {
    if (!url || isLocalSrc(url)) return url
    return resolvedImages.get(url) || url
  }

  async function prefetchImage(url: string): Promise<void> {
    if (!url || isLocalSrc(url)) return
    if (resolvedImages.has(url)) return
    try {
      const mgr = cm()
      if (!mgr) return
      const dataUrl = await dedupedFetch(url, async () => {
        return await mgr!.GetImage(url) as string
      })
      if (dataUrl) resolvedImages.set(url, dataUrl)
    } catch {}
  }

  async function prefetchImages(urls: string[]): Promise<void> {
    await Promise.allSettled(urls.map(url => prefetchImage(url)))
  }

  async function invalidateUser(uuid: string): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      await mgr.Remove(`user_${uuid}`)
    } catch {}
  }

  async function invalidateFriends(): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      await mgr.Remove('friends_list')
    } catch {}
  }

  async function clearAll(): Promise<void> {
    try {
      const mgr = cm()
      if (!mgr) return
      await mgr.Clear()
      cacheSize.value = 0
      resolvedImages.clear()
    } catch {}
  }

  async function clearExpired(): Promise<void> {
    await updateSize()
  }

  async function fetchWithCache<T>(
    key: string,
    fetcher: () => Promise<T>,
    ttl: number = DEFAULT_TTL,
  ): Promise<T> {
    try {
      const mgr = cm()
      if (mgr) {
        const cached = await mgr.Get(key) as T | null
        if (cached !== null) return cached
      }
    } catch {}
    return dedupedFetch(key, async () => {
      const data = await fetcher()
      if (data !== null && data !== undefined) {
        try {
          const mgr = cm()
          if (mgr) await mgr.Set(key, data, ttl)
        } catch {}
      }
      return data
    })
  }

  return {
    cacheSize,
    getUser,
    setUser,
    getFriendList,
    setFriendList,
    getSearchResults,
    setSearchResults,
    getImageUrl,
    getCachedSrc,
    preloadImage,
    prefetchImage,
    prefetchImages,
    invalidateUser,
    invalidateFriends,
    clearAll,
    clearExpired,
    fetchWithCache,
  }
}
