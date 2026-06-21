<template>
  <div class="friends-widget" v-if="show">
    <div class="header" @click="toggleExpanded">
      <img :src="'assets/svg/friends.svg'" width="14" height="14" class="icon" />
      <span class="title">{{ $t('friends.widget.title') }}</span>
      <div class="header-right">
        <span class="count">{{ onlineCount }}/{{ friends.length }}</span>
        <img :class="['chevron', { open: expanded }]" :src="'assets/svg/chevron-down.svg'" width="10" height="10" />
      </div>
    </div>
    <transition name="collapse">
      <div class="list" v-if="expanded">
        <div v-if="loading" class="status-row">{{ $t('friends.widget.loading') }}</div>
        <template v-else-if="friends.length > 0">
          <div v-for="group in grouped" :key="group.key" class="group">
            <div class="group-header">{{ group.label }} — {{ group.list.length }}</div>
            <div v-for="f in group.list" :key="f.uuid" class="friend" @click="emit('open-profile', f.uuid)">
              <div class="avatar-wrap">
                <img :src="cache.getCachedSrc(avatarSrc(f.avatar))" alt="" class="avatar" />
                <span class="dot" :class="f.statusClass" />
              </div>
              <div class="info">
                <span class="name">{{ f.username }}</span>
                <span class="presence" :class="f.statusClass">{{ f.presenceLabel }}</span>
              </div>
            </div>
          </div>
        </template>
        <div v-else class="status-row">{{ $t('friends.widget.empty') }}</div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { t } from '../../i18n'
import { useUserCache } from '../../Composables/useUserCache'

interface Friend {
  uuid: string
  username: string
  avatar: string
  statusClass: 'online' | 'in_game' | 'idle' | 'busy' | 'offline'
  presenceLabel: string
}

interface StatusGroup {
  key: string
  label: string
  list: Friend[]
}

const emit = defineEmits<{ 'open-profile': [uuid: string] }>()
const cache = useUserCache()

const show = ref(false)
const expanded = ref(localStorage.getItem('fw_expanded') !== 'false')
const friends = ref<Friend[]>([])
const loading = ref(true)

let splashTimer: ReturnType<typeof setTimeout> | null = null
let splashInterval: ReturnType<typeof setInterval> | null = null

const onlineCount = computed(() => {
  return friends.value.filter(f => f.statusClass === 'online' || f.statusClass === 'in_game').length
})

const grouped = computed(() => {
  const groups: StatusGroup[] = [
    { key: 'in_game', label: t('friends.widget.groups.in_game'), list: [] },
    { key: 'online', label: t('friends.widget.groups.online'), list: [] },
    { key: 'idle', label: t('friends.widget.groups.idle'), list: [] },
    { key: 'busy', label: t('friends.widget.groups.busy'), list: [] },
    { key: 'offline', label: t('friends.widget.groups.offline'), list: [] },
  ]
  for (const f of friends.value) {
    const g = groups.find(g => g.key === f.statusClass)
    if (g) g.list.push(f)
  }
  return groups.filter(g => g.list.length > 0)
})

function avatarSrc(url: string): string {
  if (url) return url
  return 'assets/defaults/steve_avatar.png'
}

function toggleExpanded() {
  expanded.value = !expanded.value
  localStorage.setItem('fw_expanded', String(expanded.value))
}

function relativeTime(iso: string | null | undefined): string | null {
  if (!iso) return null
  try {
    const diff = Date.now() - new Date(iso).getTime()
    if (diff < 0) return null
    const mins = Math.floor(diff / 60000)
    if (mins < 1) return t('friends.widget.presence.seen_just_now')
    if (mins < 60) return t('friends.widget.presence.seen_minutes', { m: mins })
    const hrs = Math.floor(mins / 60)
    if (hrs < 24) return t('friends.widget.presence.seen_hours', { h: hrs })
    const days = Math.floor(hrs / 24)
    if (days < 7) return t('friends.widget.presence.seen_days', { d: days })
    return new Date(iso).toLocaleDateString('es')
  } catch {
    return null
  }
}

function parseFriend(raw: any): Friend | null {
  try {
    if (!raw || typeof raw !== 'object') return null

    const uuid = raw.friend_uuid ?? raw.uuid ?? raw.id ?? ''
    const username = raw.friend_username ?? raw.username ?? raw.name ?? ''
    const avatar = raw.friend_avatar_url ?? raw.avatar ?? ''
    const isOnline = raw.is_online === true || raw.online === true
    const lastSeen = raw.last_seen ?? raw.lastSeen ?? null

    if (!uuid || !username) return null

    if (isOnline) {
      return { uuid, username, avatar, statusClass: 'online', presenceLabel: t('friends.widget.presence.online') }
    }

    const seen = relativeTime(lastSeen)
    return {
      uuid,
      username,
      avatar,
      statusClass: 'offline',
      presenceLabel: seen ? t('friends.widget.presence.seen_format', { time: seen }) : t('friends.widget.presence.offline'),
    }
  } catch {
    return null
  }
}

onMounted(() => {
  try {
    if (window.__splashDone) {
      show.value = true
    } else {
      splashInterval = setInterval(() => {
        try {
          if (window.__splashDone) {
            show.value = true
            if (splashInterval) clearInterval(splashInterval)
            splashInterval = null
          }
        } catch {}
      }, 100)
      splashTimer = setTimeout(() => {
        if (splashInterval) {
          clearInterval(splashInterval)
          splashInterval = null
        }
        show.value = true
      }, 15000)
    }

    ;(async () => {
      try {
        const cached = await cache.getFriendList()
        if (cached) {
          friends.value = cached
          loading.value = false
          return
        }

        if (typeof window.AuthManager?.GetFriends !== 'function') {
          console.warn('[FriendsWidget] AuthManager.GetFriends not available')
          return
        }

        const data = await window.AuthManager.GetFriends()

        let rawList: any[] = []
        if (Array.isArray(data)) {
          rawList = data
        } else if (data && typeof data === 'object') {
          const d = data as Record<string, unknown>
          rawList = (d.data ?? d.friends ?? d.users ?? d.results ?? d.list ?? []) as any[]
        }

        const parsed: Friend[] = []
        for (const item of rawList) {
          const f = parseFriend(item)
          if (f) parsed.push(f)
        }

        friends.value = parsed
        cache.setFriendList(parsed)
        cache.prefetchImages(parsed.map(f => f.avatar).filter(Boolean))
      } catch (err) {
        console.error('[FriendsWidget] fetch error:', err)
      } finally {
        loading.value = false
      }
    })()
  } catch (err) {
    console.error('[FriendsWidget] mount error:', err)
    show.value = true
    loading.value = false
    expanded.value = true
  }
})

onUnmounted(() => {
  if (splashInterval) {
    clearInterval(splashInterval)
    splashInterval = null
  }
  if (splashTimer) {
    clearTimeout(splashTimer)
    splashTimer = null
  }
})
</script>

<style scoped lang="scss" src="../../Styles/Widgets/FriendsWidget.scss"></style>
