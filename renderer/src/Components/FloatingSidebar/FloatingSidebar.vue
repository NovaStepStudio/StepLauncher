<template>
  <div
    class="fixed-sidebar"
    :class="[position, { expanded: isExpanded, dragging: isDragging }]"
    :style="sidebarStyle"
    @mouseenter="isExpanded = true"
    @mouseleave="isExpanded = false"
  >
    <div
      v-if="isExpanded && (position === 'left' || position === 'right')"
      class="drag-handle"
      :class="position"
      @mousedown.prevent="startDrag"
    />
    <nav class="menu">
      <div
        v-for="(item, idx) in items"
        :key="idx"
        class="item"
        @click="item.action && item.action()"
      >
        <span class="icon-wrap"><img class="icon" v-bind:src="item.icon"></span>
        <span class="label">{{ item.text }}</span>
        <span v-if="item.badge && item.badge > 0" class="badge">{{ item.badge }}</span>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface SidebarItem {
  text: string
  icon: string
  action?: () => void
  badge?: number
}

interface Props {
  items: SidebarItem[]
  position?: 'left' | 'right' | 'top' | 'bottom'
}

interface Emits {
  (e: 'update:position', pos: 'left' | 'right'): void
}

const props = withDefaults(defineProps<Props>(), { 
  position: 'left' 
})
const emit = defineEmits<Emits>()

const isExpanded = ref(false)
const isDragging = ref(false)
const dragX = ref(0)

const sidebarStyle = computed(() => {
  if (isDragging.value) {
    return {
      left: dragX.value + 'px',
      right: 'auto',
      transform: 'translateY(-50%)',
    }
  }
  if (props.position === 'right') return { right: 0, left: 'auto' }
  return { left: 0, right: 'auto' }
})

function startDrag(e: MouseEvent) {
  const el = (e.currentTarget as HTMLElement).parentElement!
  const rect = el.getBoundingClientRect()
  const offsetX = e.clientX - rect.left

  isDragging.value = true
  dragX.value = e.clientX - offsetX

  function onMove(ev: MouseEvent) {
    dragX.value = ev.clientX - offsetX
  }

  function onUp(ev: MouseEvent) {
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
    isDragging.value = false

    const center = window.innerWidth / 2
    const newPos = ev.clientX < center ? 'left' : 'right'
    if (newPos !== props.position) {
      emit('update:position', newPos)
    }
  }

  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}
</script>

<style scoped lang="scss" src="../../Styles/Components/FloatingSidebar.scss"></style>
