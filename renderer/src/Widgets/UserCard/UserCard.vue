<template>
  <div class="user-card" @click="emit('open-profile', user.uuid)" v-if="show">
    <div v-if="user.banner_url" ref="bannerRef" class="user-card-banner" :style="{ backgroundImage: `url(${bannerSrc})` }">
      <StaticLight
        :container-ref="bannerRef"
        :source="bannerSrc"
        :options="{ blur: 180, canvasBlur: 25, brightness: 2.5, saturation: 2.5, opacity: 0.6, scale: 1.5, zIndex: -1, fadeInDuration: 400 }"
      />
    </div>
    <div ref="avatarWrapRef" class="avatar-wrap">
      <img :src="avatarSrc" class="avatar" alt="" />
      <StaticLight
        :container-ref="avatarWrapRef"
        :source="avatarSrc"
        :options="{
          scale: 1.5,
          brightness: 1.5,
          blur: 120,
          canvasBlur: 150,
          opacity: 1,
          zIndex: 1,
        }"
      />
      <span class="status-dot" :class="user.is_online ? 'online' : 'offline'" />
    </div>
    <div class="info">
      <span class="username">{{ user.username }}</span>
      <span class="presence">{{
        $t(user.is_online ? 'widgets.user_card.online' : 'widgets.user_card.offline')
      }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { AuthUser } from '../../Composables/useAuth'
import { useUserCache } from '../../Composables/useUserCache'
import StaticLight from '../../Components/StaticLight/StaticLight.vue'

interface Props {
  user: AuthUser
}

const props = defineProps<Props>()
const emit = defineEmits<{ 'open-profile': [uuid: string] }>()

const cache = useUserCache()
const show = ref(false)
const defaultAvatar = 'assets/defaults/steve_avatar.png'
const avatarWrapRef = ref<HTMLElement | null>(null)
const bannerRef = ref<HTMLElement | null>(null)

const avatarSrc = computed(() => cache.getCachedSrc(props.user.avatar || defaultAvatar))
const bannerSrc = computed(() => props.user.banner_url ? cache.getCachedSrc(props.user.banner_url) : '')

onMounted(() => {
  if (window.__splashDone) {
    show.value = true
  } else {
    const check = setInterval(() => {
      if (window.__splashDone) {
        show.value = true
        clearInterval(check)
      }
    }, 100)
    setTimeout(() => {
      clearInterval(check)
      show.value = true
    }, 15000)
  }
})
</script>

<style scoped lang="scss" src="../../Styles/Widgets/UserCard.scss"></style>
