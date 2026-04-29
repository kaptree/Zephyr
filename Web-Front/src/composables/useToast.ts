import { ref } from 'vue'

interface ToastItem {
  id: number
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
}

const toasts = ref<ToastItem[]>([])
let nextId = 0

export function useToast() {
  function addToast(type: ToastItem['type'], message: string, duration = 3000) {
    const id = nextId++
    toasts.value.push({ id, type, message })
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
    success: (msg: string) => addToast('success', msg),
    error: (msg: string) => addToast('error', msg),
    warning: (msg: string) => addToast('warning', msg),
    info: (msg: string) => addToast('info', msg),
  }
}
