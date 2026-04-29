import { ref, watch } from 'vue'

interface QueuedAction {
  id: string
  type: string
  payload: unknown
  timestamp: number
}

export function useOffline() {
  const isOnline = ref(navigator.onLine)
  const pendingActions = ref<QueuedAction[]>(
    JSON.parse(localStorage.getItem('offline_queue') || '[]')
  )

  function handleOnline() {
    isOnline.value = true
    syncPendingActions()
  }

  function handleOffline() {
    isOnline.value = false
  }

  function enqueueAction(action: Omit<QueuedAction, 'timestamp'>) {
    pendingActions.value.push({
      ...action,
      timestamp: Date.now(),
    })
    persistQueue()
  }

  function persistQueue() {
    localStorage.setItem('offline_queue', JSON.stringify(pendingActions.value))
  }

  async function syncPendingActions() {
    const actions = [...pendingActions.value]
    for (const action of actions) {
      try {
        // 重放操作 - 由调用方实现具体逻辑
        pendingActions.value = pendingActions.value.filter(a => a.id !== action.id)
      } catch {
        // 保留失败的操作下次重试
        break
      }
    }
    persistQueue()
  }

  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)

  return {
    isOnline,
    pendingActions,
    enqueueAction,
    syncPendingActions,
  }
}
