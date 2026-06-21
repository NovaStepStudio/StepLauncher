<template>
  <div v-if="profile">
    <div v-if="profile.user.banner_url" ref="bannerRef" class="profile-banner-container">
      <StaticLight class="profile-banner-light" :container-ref="bannerRef" :source="cache.getCachedSrc(profile.user.banner_url)" :options="{ blur: 180, canvasBlur: 25, brightness: 2.5, saturation: 2.5, opacity: 1, scale: 1.5, zIndex: -1, fadeInDuration: 400 }" />
      <div class="profile-banner" :style="{ backgroundImage: `url(${cache.getCachedSrc(profile.user.banner_url)})` }"></div>
    </div>
    <div class="profile-view">
      <div class="profile-left">
        <div class="avatar-large-container" ref="avatarContainerRef">
          <div class="avatar-large">
            <img v-if="profile.user.avatar" :src="cache.getCachedSrc(profile.user.avatar)" />
            <div v-else class="avatar-placeholder">{{ profile.user.username?.[0]?.toUpperCase() }}</div>
          </div>
          <StaticLight class="avatar-glow" :container-ref="avatarContainerRef" :source="cache.getCachedSrc(profile.user.avatar)" :options="{ blur: 180, canvasBlur: 25, brightness: 2.5, saturation: 2.5, opacity: 1, scale: 1.5, zIndex: -1, fadeInDuration: 400 }" />
          <div v-if="profile.user.is_verified" class="verified-badge-avatar" title="Cuenta verificada">
            <img :src="'assets/svg/verified.svg'" width="14" height="14" />
          </div>
        </div>

        <div class="name-row">
          <h1>{{ profile.user.username }}</h1>
          <span :class="['tier', profile.user.premium]">{{ tierLabel(profile.user.premium) }}</span>
          <span v-if="profile.user.is_admin" class="admin-badge">Admin</span>
        </div>

        <div class="stats-row">
          <div class="stat">
            <strong>{{ profile.user.followers_count ?? 0 }}</strong>
            <span>Seguidores</span>
          </div>
          <div class="stat">
            <strong>{{ profile.user.following_count ?? 0 }}</strong>
            <span>Siguiendo</span>
          </div>
          <div class="stat">
            <strong>{{ profile.user.likes_count ?? 0 }}</strong>
            <span>Me gusta</span>
          </div>
        </div>

        <div v-if="profile.user.bio" class="bio">{{ profile.user.bio }}</div>

        <div class="details-grid">
          <div class="detail-card">
            <span class="detail-label">Miembro desde</span>
            <span class="detail-value">{{ formatDate(profile.user.created_at) }}</span>
          </div>
          <div v-if="profile.user.last_version_playing" class="detail-card">
            <span class="detail-label">Última versión jugada</span>
            <span class="detail-value">{{ profile.user.last_version_playing }}</span>
          </div>
          <div v-if="profile.user.last_launcher_version" class="detail-card">
            <span class="detail-label">Última versión del launcher</span>
            <span class="detail-value">{{ profile.user.last_launcher_version }}</span>
          </div>
          <div v-if="profile.user.playing_time !== undefined" class="detail-card">
            <span class="detail-label">Tiempo jugado</span>
            <span class="detail-value">{{ formatPlayTime(profile.user.playing_time) }}</span>
          </div>
          <div class="detail-card">
            <span class="detail-label">Estado</span>
            <span :class="['detail-value', profile.user.is_online ? 'online' : 'offline']">{{ profile.user.is_online ? 'En línea' : 'Desconectado' }}</span>
          </div>
          <div v-if="profile.user.last_seen" class="detail-card">
            <span class="detail-label">Última vez</span>
            <span class="detail-value">{{ formatDate(profile.user.last_seen) }}</span>
          </div>
        </div>

        <div v-if="profile.presence" class="card-section">
          <div class="section-header">Presencia</div>
          <div class="presence-status">
            <span :class="['presence-dot', presenceClass]" />
            <strong>{{ profile.presence.presence || 'Desconectado' }}</strong>
            <span v-if="profile.presence.message" class="presence-msg"> — {{ profile.presence.message }}</span>
          </div>
          <small v-if="profile.presence.updated_at" class="presence-date">Actualizado: {{ formatDate(profile.presence.updated_at) }}</small>
        </div>

        <div v-if="profile.common_friends && profile.common_friends.length" class="card-section">
          <div class="section-header">Amigos en común ({{ profile.common_friends.length }})</div>
          <div class="common-friends">
            <button v-for="cf in profile.common_friends.slice(0, 12)" :key="cf.uuid" class="cf-link" @click="$emit('select-user', cf.uuid)">
              <div class="cf-avatar">
                <img v-if="cf.avatar" :src="cache.getCachedSrc(cf.avatar)" />
                <div v-else class="cf-placeholder">{{ cf.username?.[0]?.toUpperCase() }}</div>
              </div>
              <span class="cf-name">{{ cf.username }}</span>
            </button>
          </div>
        </div>

        <div v-if="profile.libraries && profile.libraries.length" class="card-section">
          <div class="section-header">
            <img :src="'assets/svg/book.svg'" width="13" height="13" />
            Librerías
          </div>
          <div class="libs-grid">
            <div v-for="lib in profile.libraries" :key="lib.id" class="lib-card" @click="$emit('open-library', lib)">
              <div class="lib-card-preview">
                <div v-if="lib.entries && lib.entries.length" class="lib-card-grid">
                  <img v-for="(e, i) in lib.entries.slice(0, 4)" :key="i" :src="cache.getCachedSrc(e.icon || '')" class="lib-card-thumb" @error="hideImg" />
                </div>
                <div v-else class="lib-card-grid lib-card-grid--empty">
                  <img :src="'assets/svg/book.svg'" width="18" height="18" />
                </div>
              </div>
              <div class="lib-card-body">
                <div class="lib-card-top">
                  <h4>{{ lib.title }}</h4>
                  <span class="lib-badge" v-if="lib.is_public">Pública</span>
                </div>
                <p v-if="lib.description" class="lib-card-desc">{{ lib.description }}</p>
                <div class="lib-card-footer">
                  <span>{{ lib.entries_count }} elementos</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="hasCosmetics" class="card-section">
          <div class="section-header">Cosméticos</div>

          <div v-if="profile.cosmetics!.skins && profile.cosmetics!.skins.length" class="cos-group">
            <div class="cos-subtitle">Skins ({{ profile.cosmetics!.skins.length }})</div>
            <div class="cos-grid">
              <div v-for="s in profile.cosmetics!.skins" :key="s.id" @click="$emit('preview-cosmetic', s.skin_url)">
                <SkinCard :name="s.name" :skin-url="s.skin_url" :active="s.is_active" />
              </div>
            </div>
          </div>

          <div v-if="profile.cosmetics!.capes && profile.cosmetics!.capes.length" class="cos-group">
            <div class="cos-subtitle">Capas ({{ profile.cosmetics!.capes.length }})</div>
            <div class="cos-grid">
              <div v-for="c in profile.cosmetics!.capes" :key="c.id" @click="$emit('preview-cosmetic', profile.user.skin_url || defaultSkin, c.cape_url)">
                <SkinCard :name="c.name" :skin-url="profile.user.skin_url || defaultSkin" :cape-url="c.cape_url" :active="c.is_active" />
              </div>
            </div>
          </div>

          <div v-if="profile.cosmetics!.kits && profile.cosmetics!.kits.length" class="cos-group">
            <div class="cos-subtitle">Kits ({{ profile.cosmetics!.kits.length }})</div>
            <div class="cos-grid-kits">
              <div v-for="k in profile.cosmetics!.kits" :key="k.id" class="kit-card-wrapper" @click="$emit('preview-cosmetic', getKitSkinUrl(k.skin_id), getKitCapeUrl(k.cape_id))">
                <SkinCard :name="k.title" :skin-url="getKitSkinUrl(k.skin_id)" :cape-url="getKitCapeUrl(k.cape_id)" :active="k.is_active" />
                <div class="kit-info-overlay">
                  <strong>
                    <img :src="'assets/svg/backpack.svg'" width="11" height="11" />
                    {{ k.title }}
                  </strong>
                  <span v-if="k.description">{{ k.description }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="profile-right">
        <div class="card-section character-card">
          <div class="section-header">Personaje</div>
          <div class="viewer-wrapper">
            <canvas ref="characterCanvas" class="character-canvas"></canvas>
          </div>
        </div>

        <div v-if="profile.social_links && profile.social_links.length" class="card-section">
          <div class="section-header">Redes sociales</div>
          <div class="social-links">
            <a v-for="link in profile.social_links" :key="link.id" :href="link.url" target="_blank" class="social-link" :title="link.platform" @click.prevent="$emit('open-external', link.url)">
              <img class="social-icon" :src="getSocialIconSrc(link.platform)" width="16" height="16" />
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import type { SkinViewer } from 'skinview3d'
import StaticLight from '../../Components/StaticLight/StaticLight.vue'
import SkinCard from '../../Components/SkinCard/SkinCard.vue'
import { useUserCache } from '../../Composables/useUserCache'

const props = defineProps<{
  profile: any
  defaultSkin: string
}>()

defineEmits<{
  'preview-cosmetic': [skinUrl: string, capeUrl?: string]
  'open-library': [lib: any]
  'select-user': [uuid: string]
  'open-external': [url: string]
}>()

const cache = useUserCache()

const bannerRef = ref<HTMLElement | null>(null)
const avatarContainerRef = ref<HTMLElement | null>(null)
const characterCanvas = ref<HTMLCanvasElement | null>(null)
let characterViewer: SkinViewer | null = null

const presenceClass = computed(() => {
  const p = props.profile?.presence?.presence
  if (!p) return 'presence-desconectado'
  const lower = p.toLowerCase()
  if (lower.includes('conectado')) return 'presence-conectado'
  if (lower.includes('molestar')) return 'presence-molestar'
  return 'presence-desconectado'
})

const hasCosmetics = computed(() => {
  if (!props.profile?.cosmetics) return false
  const c = props.profile.cosmetics
  return (c.skins && c.skins.length > 0) || (c.capes && c.capes.length > 0) || (c.kits && c.kits.length > 0)
})

function hideImg(e: Event) {
  (e.target as HTMLElement)?.style?.setProperty('display', 'none')
}

function getSocialIconSrc(platform: string): string {
  const icons: Record<string, string> = {
    twitter: 'assets/svg/logo-x.svg',
    github: 'assets/svg/logo-github.svg',
    discord: 'assets/svg/logo-discord.svg',
    youtube: 'assets/svg/logo-youtube.svg',
    twitch: 'assets/svg/logo-twitch.svg',
    instagram: 'assets/svg/logo-instagram.svg',
    linkedin: 'assets/svg/logo-linkedin.svg',
    tiktok: 'assets/svg/logo-tiktok.svg',
  }
  return icons[platform.toLowerCase()] || 'assets/svg/user-circle.svg'
}

function formatPlayTime(seconds: number): string {
  if (!seconds || seconds <= 0) return '0h'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h === 0) return `${m}m`
  if (m === 0) return `${h}h`
  return `${h}h ${m}m`
}

function formatDate(d?: string | null): string {
  if (!d) return '—'
  const date = new Date(d)
  if (isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('es-ES', { day: 'numeric', month: 'short', year: 'numeric' })
}

function tierLabel(tier: string): string {
  const labels: Record<string, string> = {
    free: 'Free',
    premium: 'Premium',
    team: 'Team',
    ultimate: 'Ultimate',
  }
  return labels[tier] || tier
}

function getKitSkinUrl(skinId?: number): string {
  if (!skinId || !props.profile?.cosmetics?.skins) return props.defaultSkin
  const skin = props.profile.cosmetics.skins.find((s: { id: number; skin_url: string }) => s.id === skinId)
  return skin ? skin.skin_url : props.defaultSkin
}

function getKitCapeUrl(capeId?: number): string | undefined {
  if (!capeId || !props.profile?.cosmetics?.capes) return undefined
  const cape = props.profile.cosmetics.capes.find((c: { id: number; cape_url: string }) => c.id === capeId)
  return cape ? cape.cape_url : undefined
}

async function initCharacterViewer(skinUrl: string, capeUrl?: string) {
  if (characterViewer) {
    characterViewer.dispose()
    characterViewer = null
  }
  if (!characterCanvas.value) return
  try {
    let skinImg = await cache.preloadImage(skinUrl)
    if (!skinImg) {
      skinImg = await cache.preloadImage(props.defaultSkin)
    }
    if (!characterCanvas.value || !skinImg) return
    let capeImg: HTMLImageElement | null | undefined
    if (capeUrl) {
      capeImg = await cache.preloadImage(capeUrl)
    }
    if (!characterCanvas.value) return
    const { SkinViewer } = await import('skinview3d')
    characterViewer = new SkinViewer({
      canvas: characterCanvas.value,
      width: 180,
      height: 250,
      skin: skinImg,
      ...(capeImg ? { cape: capeImg } : {}),
    })
    characterViewer.autoRotate = true
    characterViewer.autoRotateSpeed = 1.0
    characterViewer.zoom = 0.85
  } catch (e) {
    console.warn('SkinViewer init failed:', e)
  }
}

function prefetchProfileImages() {
  const p = props.profile
  if (!p) return
  const urls: string[] = []
  if (p.user?.avatar) urls.push(p.user.avatar)
  if (p.user?.banner_url) urls.push(p.user.banner_url)
  if (p.user?.skin_url) urls.push(p.user.skin_url)
  if (p.user?.cape_url) urls.push(p.user.cape_url)
  if (p.cosmetics?.skins) for (const s of p.cosmetics.skins) if (s.skin_url) urls.push(s.skin_url)
  if (p.cosmetics?.capes) for (const c of p.cosmetics.capes) if (c.cape_url) urls.push(c.cape_url)
  if (p.cosmetics?.kits) for (const k of p.cosmetics.kits) {
    const sk = getKitSkinUrl(k.skin_id)
    if (sk && !sk.startsWith('assets/')) urls.push(sk)
    const ck = getKitCapeUrl(k.cape_id)
    if (ck && !ck.startsWith('assets/')) urls.push(ck)
  }
  if (p.common_friends) for (const cf of p.common_friends) if (cf.avatar) urls.push(cf.avatar)
  if (p.libraries) for (const lib of p.libraries) if (lib.entries) for (const e of lib.entries) if (e.icon) urls.push(e.icon)
  cache.prefetchImages(urls)
}

watch(() => props.profile, () => {
  prefetchProfileImages()
}, { immediate: true })

onMounted(async () => {
  await nextTick()
  const p = props.profile
  if (p?.user?.skin_url) {
    initCharacterViewer(p.user.skin_url, p.user.cape_url || undefined)
  }
})

onBeforeUnmount(() => {
  if (characterViewer) {
    characterViewer.dispose()
    characterViewer = null
  }
})
</script>

<style scoped lang="scss" src="../../Styles/Layouts/ProfileSection.scss"></style>
