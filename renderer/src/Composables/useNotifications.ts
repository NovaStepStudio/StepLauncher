import { ref } from 'vue'

export interface Notification {
  id: number
  message: string
  type: 'success' | 'info' | 'warning' | 'error'
  timeout: number
}

const notifications = ref<Notification[]>([])
let nextId = 0

export function useNotifications() {
  function notify(message: string, type: Notification['type'] = 'info', timeout = 4500): number {
    const id = nextId++
    notifications.value.push({ id, message, type, timeout })
    if (timeout > 0) {
      setTimeout(() => remove(id), timeout)
    }
    return id
  }

  function remove(id: number) {
    const idx = notifications.value.findIndex(n => n.id === id)
    if (idx >= 0) notifications.value.splice(idx, 1)
  }

  function success(msg: string, timeout?: number) { return notify(msg, 'success', timeout) }
  function info(msg: string, timeout?: number) { return notify(msg, 'info', timeout) }
  function warning(msg: string, timeout?: number) { return notify(msg, 'warning', timeout) }
  function error(msg: string, timeout?: number) { return notify(msg, 'error', timeout) }

  return { notifications, notify, remove, success, info, warning, error }
}
