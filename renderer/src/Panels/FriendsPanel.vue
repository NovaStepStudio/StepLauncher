<template>
  <section class="friends-panel">
    <div class="friends-sidebar">
      <button class="btn-icon" @click="$emit('close')">
        <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
      </button>
      <h2 class="friends-title">{{ $t('friends.panel.title') }}</h2>

      <div class="search-wrap">
        <input v-model="searchQ" @keyup.enter="doSearch" class="search-input" :placeholder="$t('friends.panel.search_placeholder')" />
        <button class="btn-icon btn-search" @click="doSearch" :disabled="searching" :title="$t('general.search')">
          <img :src="'assets/svg/search.svg'" width="14" height="14" />
        </button>
      </div>

      <div v-if="searchResults.length" class="results-section">
        <h3>{{ $t('friends.panel.search_results') }}</h3>
        <div v-for="u in searchResults" :key="u.id" class="row-item" @click="selectUserByUuid(u.uuid)">
          <div class="av-wrap">
            <img v-if="u.avatar" :src="cache.getCachedSrc(u.avatar)" />
            <div v-else class="av-init">{{ u.username?.[0]?.toUpperCase() }}</div>
          </div>
          <div class="info-col">
            <strong>{{ u.username }}</strong>
            <span>{{ u.bio ?? $t('friends.panel.no_bio') }}</span>
          </div>
          <button v-if="authUser && u.id !== authUser.uuid" class="btn btn--sm btn--add" @click.stop="doSendReq(u.id)" :title="$t('friends.panel.buttons.add')"><img :src="'assets/svg/plus.svg'" width="12" height="12" /></button>
        </div>
      </div>

      <div v-if="onlineFriends.length" class="section">
        <h3><span class="dot dot--online"></span> {{ $t('friends.panel.sections.online') }} <span class="badge">{{ onlineFriends.length }}</span></h3>
        <div v-for="f in onlineFriends" :key="f.friend_id" class="row-item" @click="selectUser(f)">
          <div class="av-wrap">
            <img v-if="f.friend_avatar_url" :src="cache.getCachedSrc(f.friend_avatar_url)" />
            <div v-else class="av-init">{{ f.friend_username?.[0]?.toUpperCase() }}</div>
          </div>
          <div class="info-col">
            <strong>{{ f.friend_username }}</strong>
            <span>{{ f.friend_bio ?? '' }}</span>
          </div>
          <span class="dot dot--online"></span>
        </div>
      </div>

      <div class="section">
        <h3>{{ $t('friends.panel.sections.all') }} <span class="badge">{{ friends.length }}</span></h3>
        <div v-if="!friends.length" class="empty-msg">{{ $t('friends.panel.empty_friends') }}</div>
        <div v-for="f in offlineFriends" :key="f.friend_id" class="row-item" @click="selectUser(f)">
          <div class="av-wrap">
            <img v-if="f.friend_avatar_url" :src="cache.getCachedSrc(f.friend_avatar_url)" />
            <div v-else class="av-init">{{ f.friend_username?.[0]?.toUpperCase() }}</div>
          </div>
          <div class="info-col">
            <strong>{{ f.friend_username }}</strong>
            <span>{{ f.friend_bio ?? '' }}</span>
          </div>
          <span class="dot dot--offline"></span>
        </div>
      </div>

      <div v-if="reqReceived.length" class="section">
        <h3>
          <img :src="'assets/svg/plus.svg'" width="13" height="13" />
          {{ $t('friends.panel.sections.received') }} <span class="badge">{{ reqReceived.length }}</span>
        </h3>
        <div v-for="r in reqReceived" :key="r.request_id" class="row-item">
          <div class="av-wrap">
            <img v-if="r.from_avatar" :src="cache.getCachedSrc(r.from_avatar)" />
            <div v-else class="av-init">{{ r.from_username?.[0]?.toUpperCase() }}</div>
          </div>
          <div class="info-col"><strong>{{ r.from_username ?? '?' }}</strong></div>
          <div class="acts">
            <button class="btn btn--sm btn--accept" @click="doAccept(r.request_id)" :title="$t('friends.panel.buttons.accept')"><img :src="'assets/svg/check.svg'" width="12" height="12" /></button>
            <button class="btn btn--sm btn--reject" @click="doReject(r.request_id)" :title="$t('friends.panel.buttons.reject')"><img :src="'assets/svg/x.svg'" width="12" height="12" /></button>
          </div>
        </div>
      </div>

      <div v-if="reqSent.length" class="section">
        <h3>
          <img :src="'assets/svg/upload.svg'" width="13" height="13" />
          {{ $t('friends.panel.sections.sent') }} <span class="badge">{{ reqSent.length }}</span>
        </h3>
        <div v-for="r in reqSent" :key="r.request_id" class="row-item">
          <div class="info-col"><strong>{{ r.to_username ?? '—' }}</strong></div>
          <span class="pending-badge">{{ $t('friends.panel.badge_pending') }}</span>
        </div>
      </div>
    </div>

    <div class="friends-main">
      <div class="friends-main-scroll">
        <div class="empty-state" v-if="!selectedUuid">
          <img :src="'assets/svg/friends.svg'" width="34" height="34" />
          <span>{{ $t('friends.panel.empty_state') }}</span>
        </div>

        <div v-if="selectedUuid && loadingProfile" class="profile-loading">
          <div class="spinner"></div>
          <p>{{ $t('friends.panel.profile_loading') }}</p>
        </div>

        <ProfileSection
          v-if="selectedUuid && profile && !loadingProfile"
          :profile="profile"
          :default-skin="defaultSkin"
          @preview-cosmetic="openCosmeticPreview"
          @open-library="openProfileLib"
          @select-user="selectUserByUuid"
          @open-external="openExternal"
        />
      </div>
    </div>

    <div v-if="previewCosmetic" class="modal-overlay" @click.self="closeCosmeticPreview">
      <button class="modal-close" @click="closeCosmeticPreview">
        <img :src="'assets/svg/x.svg'" width="20" height="20" />
      </button>
      <canvas ref="previewCanvas" class="preview-canvas-large"></canvas>
    </div>

    <div v-if="profileLibDetail" class="modal-overlay" @click.self="profileLibEntry ? (profileLibEntry = null) : (profileLibDetail = null)">
      <div class="lib-modal">
        <div class="lib-modal-header">
          <template v-if="profileLibEntry">
            <button @click="profileLibEntry = null" class="btn-icon"><img :src="'assets/svg/arrow-left.svg'" width="16" height="16" /></button>
            <span class="lib-modal-title">{{ $t('mods.library.back') }}</span>
          </template>
          <template v-else>
            <span class="lib-modal-title">{{ profileLibDetail.title }}</span>
          </template>
          <button class="btn-icon" @click="profileLibDetail = null"><img :src="'assets/svg/x.svg'" width="16" height="16" /></button>
        </div>
        <p v-if="!profileLibEntry && profileLibDetail.description" class="lib-modal-desc">{{ profileLibDetail.description }}</p>

        <template v-if="profileLibEntry">
          <div class="lib-entry-detail">
            <div class="lib-entry-info">
              <div class="lib-entry-head">
                <img v-if="profileLibEntry.icon" :src="cache.getCachedSrc(profileLibEntry.icon)" class="lib-entry-icon-lg" />
                <div>
                  <h3>{{ profileLibEntry.title }}</h3>
                  <div class="lib-entry-meta">
                    <span class="lib-entry-tag">{{ profileLibEntry.entry_type }}</span>
                    <span v-if="profileLibEntry.color" class="lib-color-dot" :style="{ background: '#' + profileLibEntry.color }"></span>
                  </div>
                </div>
              </div>

              <div class="lib-entry-grid">
                <div v-if="profileLibEntry.author_name || entryDetailData?.author"><strong>{{ $t('mods.entry.author') }}</strong><span>{{ profileLibEntry.author_name || entryDetailData?.author }}</span></div>
                <div v-if="entryDetailData?.downloads !== undefined"><strong>{{ $t('mods.entry.downloads') }}</strong><span>{{ fmtNum(entryDetailData.downloads) }}</span></div>
                <div v-if="entryDetailData?.followers !== undefined || entryDetailData?.follows !== undefined"><strong>{{ $t('mods.entry.followers') }}</strong><span>{{ fmtNum(entryDetailData.followers || entryDetailData.follows) }}</span></div>
                <div v-if="entryDetailData?.client_side"><strong>{{ $t('mods.entry.client') }}</strong><span>{{ entryDetailData.client_side }}</span></div>
                <div v-if="entryDetailData?.server_side"><strong>{{ $t('mods.entry.server') }}</strong><span>{{ entryDetailData.server_side }}</span></div>
                <div v-if="entryDetailData?.license?.id"><strong>{{ $t('mods.entry.license') }}</strong><span>{{ entryDetailData.license.id }}</span></div>
                <div v-if="profileLibEntry.slug"><strong>{{ $t('mods.entry.slug') }}</strong><span>{{ profileLibEntry.slug }}</span></div>
                <div v-if="profileLibEntry.provider"><strong>{{ $t('mods.entry.provider') }}</strong><span>{{ profileLibEntry.provider }}</span></div>
              </div>

              <div v-if="profileLibEntry.categories?.length" class="lib-entry-sec">
                <strong>{{ $t('mods.entry.categories') }}</strong>
                <div class="lib-entry-tags"><span v-for="c in profileLibEntry.categories" :key="c" class="lib-entry-tag">{{ c }}</span></div>
              </div>
              <div v-if="profileLibEntry.tags?.length" class="lib-entry-sec">
                <strong>{{ $t('mods.entry.tags') }}</strong>
                <div class="lib-entry-tags"><span v-for="t in profileLibEntry.tags" :key="t" class="lib-entry-tag">{{ t }}</span></div>
              </div>
              <div v-if="entryDetailData?.loaders?.length" class="lib-entry-sec">
                <strong>{{ $t('mods.entry.loaders') }}</strong>
                <div class="lib-entry-tags"><span v-for="l in entryDetailData.loaders" :key="l" class="lib-entry-tag">{{ l }}</span></div>
              </div>
              <div v-if="entryDetailData?.game_versions?.length" class="lib-entry-sec">
                <strong>{{ $t('mods.entry.versions') }}</strong>
                <div class="lib-entry-tags">
                  <span v-for="v in entryDetailData.game_versions.slice(0, 10)" :key="v" class="lib-entry-tag">{{ v }}</span>
                  <span v-if="entryDetailData.game_versions.length > 10" class="lib-entry-tag">+{{ entryDetailData.game_versions.length - 10 }}</span>
                </div>
              </div>

              <div class="lib-entry-sec">
                <strong>{{ $t('mods.entry.links') }}</strong>
                <div class="lib-entry-links">
                  <a v-if="profileLibEntry.project_url" :href="profileLibEntry.project_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_modrinth')"><img :src="modrinthSvg" class="modrinth-icon" /> {{ $t('mods.entry.link_modrinth') }}</a>
                  <a v-if="profileLibEntry.issues_url" :href="profileLibEntry.issues_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_issues')">{{ $t('mods.entry.link_issues') }}</a>
                  <a v-if="profileLibEntry.source_url" :href="profileLibEntry.source_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_source')">{{ $t('mods.entry.link_source') }}</a>
                  <a v-if="profileLibEntry.wiki_url" :href="profileLibEntry.wiki_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_wiki')">{{ $t('mods.entry.link_wiki') }}</a>
                  <a v-if="profileLibEntry.discord_url" :href="profileLibEntry.discord_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_discord')">{{ $t('mods.entry.link_discord') }}</a>
                  <a v-if="profileLibEntry.donate_url" :href="profileLibEntry.donate_url" target="_blank" class="lib-entry-link" :title="$t('mods.entry.link_donate')">{{ $t('mods.entry.link_donate') }}</a>
                </div>
              </div>
            </div>
            <div class="lib-entry-body">
              <div v-if="loadingEntryBody" class="lib-entry-state">{{ $t('general.loading') }}</div>
              <div v-else-if="entryBodyHtml" class="markdown-body" v-html="entryBodyHtml"></div>
              <div v-else-if="profileLibEntry.description" class="lib-entry-body-desc">{{ profileLibEntry.description }}</div>
              <div v-else class="lib-entry-state">{{ $t('mods.library.no_desc') }}</div>
            </div>
          </div>
        </template>

        <template v-else>
          <div v-if="loadingLibEntries" class="lib-entry-state"><div class="spinner"></div></div>
          <div v-else-if="profileLibEntries.length" class="lib-entry-list">
            <div v-for="e in profileLibEntries" :key="e.id" class="lib-entry-row" @click="showProfileLibEntry(e)">
              <img v-if="e.icon" :src="cache.getCachedSrc(e.icon)" class="lib-entry-icon" />
              <div class="lib-entry-info">
                <strong>{{ e.title }}</strong>
                <span>{{ e.entry_type }}{{ e.author_name ? ' · ' + e.author_name : '' }}</span>
              </div>
              <a v-if="e.project_url" :href="e.project_url" target="_blank" class="lib-entry-ext-link" @click.stop><img :src="modrinthSvg" class="modrinth-icon-sm" /></a>
            </div>
          </div>
          <div v-else class="lib-entry-state">{{ $t('mods.library.empty_entries') }}</div>
          <div v-if="currentUserId && currentUserId !== profile?.user?.uuid" class="lib-modal-acts">
            <button @click="toggleProfileLibLike(profileLibDetail)" class="btn-like-lib" :class="{ liked: profileLibDetail.liked }">
              <img :src="'assets/svg/heart.svg'" width="14" height="14" />
              {{ profileLibDetail.likes ?? 0 }}
            </button>
            <button @click="cloneProfileLib(profileLibDetail)" class="btn-save-lib">
              <img :src="'assets/svg/save.svg'" width="14" height="14" />
              {{ $t('mods.library.save') }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { t } from '../i18n'
import type { SkinViewer } from 'skinview3d'
import ProfileSection from '../Layouts/ProfileSection/ProfileSection.vue'
import { useAuth } from '../Composables/useAuth'
import { useUserCache } from '../Composables/useUserCache'
const modrinthSvg = 'assets/svg/modrinth.svg'
const { user: authUser } = useAuth()
const cache = useUserCache()

const props = defineProps<{ initialUuid?: string | null }>()
const emit = defineEmits<{ close: [] }>()

const defaultSkin = 'assets/defaults/steve_skin.png'

const friends = ref<any[]>([])
const reqReceived = ref<any[]>([])
const reqSent = ref<any[]>([])
const searchQ = ref('')
const searching = ref(false)
const searchResults = ref<any[]>([])
const selectedUuid = ref<string | null>(null)
const loadingProfile = ref(false)
const profile = ref<any>(null)
const previewCanvas = ref<HTMLCanvasElement | null>(null)

const previewCosmetic = ref<{ skinUrl: string; capeUrl?: string } | null>(null)

const profileLibDetail = ref<any>(null)
const profileLibEntries = ref<any[]>([])
const loadingLibEntries = ref(false)
const profileLibEntry = ref<any>(null)
const entryBodyHtml = ref('')
const loadingEntryBody = ref(false)
const entryDetailData = ref<any>(null)

let previewViewer: SkinViewer | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null

const onlineFriends = computed(() => friends.value.filter((f: any) => f.is_online))
const offlineFriends = computed(() => friends.value.filter((f: any) => !f.is_online))
const currentUserId = computed(() => profile.value?.current_user_id || null)

watch(() => props.initialUuid, (uuid) => {
  if (uuid) loadProfile(uuid)
}, { immediate: true })

let prevReceivedCount = 0

function fmtNum(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toLocaleString()
}

function openExternal(url: string) {
  window.ElectronAPI.OpenExternal(url)
}

async function loadAll() {
  try {
    const [list, requests] = await Promise.all([
      window.AuthManager.GetFriends(),
      window.AuthManager.GetFriendRequests(),
    ])
    friends.value = Array.isArray(list) ? list : []
    cache.setFriendList(friends.value)
    if (requests) {
      const newReceived = Array.isArray(requests.received) ? requests.received : []
      reqSent.value = Array.isArray(requests.sent) ? requests.sent : []
      if (newReceived.length > prevReceivedCount && prevReceivedCount > 0) {
        const diff = newReceived.length - prevReceivedCount
        if (document.hidden) window.ElectronAPI.ShowNotification({
          title: 'StepLauncher',
          body: diff > 1 ? t('install.friend_request_plural', { count: diff }) : t('install.friend_request', { count: diff }),
        })
      }
      prevReceivedCount = newReceived.length
      reqReceived.value = newReceived
    }
  } catch {
    const cached = await cache.getFriendList()
    if (cached) friends.value = cached
  }
}

onMounted(() => {
  loadAll()
  pollTimer = setInterval(loadAll, 30000)
})

async function selectUser(f: any) {
  const uuid = f.friend_uuid || f.uuid
  if (!uuid) return
  await loadProfile(uuid)
}

async function selectUserByUuid(uuid: string) {
  if (!uuid) return
  await loadProfile(uuid)
}

async function loadProfile(uuid: string) {
  selectedUuid.value = uuid
  loadingProfile.value = true
  profile.value = null
  try {
    let p: any
    try {
      p = await window.AuthManager.GetFullPublicProfile(uuid)
    } catch {}
    if (p?.user) {
      profile.value = p
      await cache.setUser(uuid, p.user, 5 * 60 * 1000)
      const imgUrls: string[] = []
      if (p.user.avatar) imgUrls.push(p.user.avatar)
      if (p.user.banner_url) imgUrls.push(p.user.banner_url)
      if (p.user.skin_url) imgUrls.push(p.user.skin_url)
      if (p.user.cape_url) imgUrls.push(p.user.cape_url)
      if (p.cosmetics?.skins) for (const s of p.cosmetics.skins) if (s.skin_url) imgUrls.push(s.skin_url)
      if (p.cosmetics?.capes) for (const c of p.cosmetics.capes) if (c.cape_url) imgUrls.push(c.cape_url)
      cache.prefetchImages(imgUrls)
    } else {
      const cached = await cache.getUser(uuid)
      if (cached) profile.value = { user: cached }
    }
  } catch {}
  loadingProfile.value = false
}

async function openCosmeticPreview(skinUrl: string, capeUrl?: string) {
  const { SkinViewer } = await import('skinview3d')
  let skinImg = await cache.preloadImage(skinUrl)
  if (!skinImg) {
    skinImg = await cache.preloadImage(defaultSkin)
  }
  if (!skinImg) return
  let capeImg: HTMLImageElement | null | undefined
  if (capeUrl) {
    capeImg = await cache.preloadImage(capeUrl)
  }
  previewCosmetic.value = { skinUrl: skinUrl, capeUrl: capeUrl }
  nextTick(() => {
    if (previewViewer) {
      previewViewer.dispose()
      previewViewer = null
    }
    if (!previewCanvas.value) return
    try {
      previewViewer = new SkinViewer({
        canvas: previewCanvas.value,
        width: 280,
        height: 400,
        skin: skinImg,
        ...(capeImg ? { cape: capeImg } : {}),
      })
      previewViewer.autoRotate = true
      previewViewer.autoRotateSpeed = 1.0
      previewViewer.zoom = 0.85
    } catch {}
  })
}

function closeCosmeticPreview() {
  previewCosmetic.value = null
  previewViewer?.dispose()
  previewViewer = null
}

async function doSearch() {
  if (searchQ.value.length < 2) return
  searching.value = true
  try {
    const cached = await cache.getSearchResults(searchQ.value)
    if (cached) {
      searchResults.value = cached
      searching.value = false
      return
    }
    const results = await window.AuthManager.SearchUsers(searchQ.value)
    const arr = Array.isArray(results) ? results : []
    searchResults.value = arr
    await cache.setSearchResults(searchQ.value, arr)
    for (const u of arr) {
      if (u?.uuid) {
        await cache.setUser(u.uuid, u, 5 * 60 * 1000)
        if (u.avatar) await cache.prefetchImage(u.avatar)
      }
    }
  } catch {
    searchResults.value = []
  }
  searching.value = false
}

async function doSendReq(id: number) {
  try {
    await window.AuthManager.SendFriendRequest(id)
    searchResults.value = searchResults.value.filter((u: any) => u.id !== id)
    await loadAll()
  } catch {}
}

async function doAccept(requestId: number) {
  try {
    await window.AuthManager.AcceptFriendRequest(requestId)
    await loadAll()
  } catch {}
}

async function doReject(requestId: number) {
  try {
    await window.AuthManager.RejectFriendRequest(requestId)
    await loadAll()
  } catch {}
}

async function openProfileLib(lib: any) {
  profileLibDetail.value = lib
  await viewProfileLibEntries(lib.id)
}

async function viewProfileLibEntries(libId: number) {
  loadingLibEntries.value = true
  profileLibEntries.value = []
  try {
    const r = await window.AuthManager.GetPublicLibraryEntries(libId)
    profileLibEntries.value = Array.isArray(r) ? r : []
  } catch {
  } finally {
    loadingLibEntries.value = false
  }
}

async function showProfileLibEntry(e: any) {
  entryBodyHtml.value = ''
  entryDetailData.value = null

  if (e.provider === 'modrinth' && e.project_id) {
    loadingEntryBody.value = true
    try {
      const res = await fetch(`https://api.modrinth.com/v2/project/${e.project_id}`)
      if (res.ok) {
        const data = await res.json()
        entryDetailData.value = data
        if (data.body) {
          const { marked } = await import('marked')
          entryBodyHtml.value = await marked.parse(data.body)
        }
      }
    } catch {
    } finally {
      loadingEntryBody.value = false
    }
  } else {
    loadingEntryBody.value = false
  }

  if (!entryBodyHtml.value && e.body) {
    try {
      const { marked } = await import('marked')
      entryBodyHtml.value = await marked.parse(e.body)
    } catch {}
  }
}

async function toggleProfileLibLike(lib: any) {
  try {
    const r = await window.AuthManager.ToggleLibraryLike(lib.id)
    if (r?.success) {
      lib.liked = r.liked
      lib.likes = r.likes
    } else if (r?.error) {
      console.warn(r.error)
    }
  } catch {}
}

async function cloneProfileLib(lib: any) {
  try {
    const r = await window.AuthManager.CloneLibrary(lib.id)
      if (r) {
        if (document.hidden) window.ElectronAPI.ShowNotification({ title: 'StepLauncher', body: t('install.library_cloned', { title: r.title || 'guardada' }) })
    }
  } catch {}
}

onUnmounted(() => {
  if (previewViewer) {
    previewViewer.dispose()
    previewViewer = null
  }
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped lang="scss" src="../Styles/Panels/FriendsPanel.scss"></style>
