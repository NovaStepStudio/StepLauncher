<template>
  <div class="notif-container">
    <transition-group name="notif-slide" tag="div" class="notif-list">
      <div v-for="n in notifications" :key="n.id" :class="['notif', n.type]" @click="remove(n.id)">
        <div class="notif-icon">
          <img v-if="n.type === 'success'" :src="'assets/svg/check.svg'" width="14" height="14" />
          <img v-else-if="n.type === 'error'" :src="'assets/svg/x-circle.svg'" width="14" height="14" />
          <img v-else-if="n.type === 'warning'" :src="'assets/svg/alert-triangle.svg'" width="14" height="14" />
          <img v-else :src="'assets/svg/alert-circle.svg'" width="14" height="14" />
        </div>
        <span class="notif-text">{{ n.message }}</span>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { useNotifications } from '../../Composables/useNotifications'

const { notifications, remove } = useNotifications()
</script>

<style scoped>
.notif-container {
  position: fixed;
  top: 2.5rem;
  right: 1rem;
  z-index: 99999;
  display: flex;
  flex-direction: column;
  gap: 6px;
  pointer-events: none;
}
.notif-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.notif {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 8px;
  backdrop-filter: blur(12px);
  cursor: pointer;
  pointer-events: auto;
  max-width: 340px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  transition: opacity 0.15s;
  &:hover { opacity: 0.85; }
  .notif-text {
    font-family: var(--font-family-secundary, 'Inter'), sans-serif;
    font-size: 0.7rem;
    font-weight: 500;
    line-height: 1.3;
  }
  &.success {
    background: var(--notif-success-bg, rgba(76, 175, 80, 0.2));
    border: 1px solid var(--notif-success-border, rgba(76, 175, 80, 0.3));
    color: var(--notif-success-text, #81c784);
  }
  &.info {
    background: rgba(33, 150, 243, 0.2);
    border: 1px solid rgba(33, 150, 243, 0.3);
    color: #64b5f6;
  }
  &.warning {
    background: var(--notif-warn-bg, rgba(255, 193, 7, 0.2));
    border: 1px solid var(--notif-warn-border, rgba(255, 193, 7, 0.3));
    color: var(--notif-warn-text, #ffd54f);
  }
  &.error {
    background: var(--notif-error-bg, rgba(244, 67, 54, 0.2));
    border: 1px solid var(--notif-error-border, rgba(244, 67, 54, 0.3));
    color: var(--notif-error-text, #e57373);
  }
}
.notif-slide-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.notif-slide-leave-active { transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1); }
.notif-slide-enter-from { opacity: 0; transform: translateX(40px); }
.notif-slide-leave-to { opacity: 0; transform: translateX(40px); }
</style>
