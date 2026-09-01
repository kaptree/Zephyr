import { ref } from 'vue'

export interface ToastItem {
  id: number
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  /** 展示时长（ms），驱动底部进度条倒计时 */
  duration: number
}

const toasts = ref<ToastItem[]>([])
let nextId = 0

export function useToast() {
  function addToast(type: ToastItem['type'], message: string, duration = 3000) {
    const id = nextId++
    toasts.value.push({ id, type, message, duration })
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function clearAll() {
    toasts.value = []
  }

  return {
    toasts,
    addToast,
    removeToast,
    clearAll,
    success: (msg: string, duration?: number) => addToast('success', msg, duration),
    error: (msg: string, duration?: number) => addToast('error', msg, duration),
    warning: (msg: string, duration?: number) => addToast('warning', msg, duration),
    info: (msg: string, duration?: number) => addToast('info', msg, duration),
  }
}
