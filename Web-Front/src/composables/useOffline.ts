import { ref } from 'vue'

interface QueuedAction {
  id: string
  type: string
  payload: unknown
  timestamp: number
}

// 模块级单例：后端连通状态全局共享，轮询定时器只启动一次（初始假定可达，首次探测纠正）
const backendOnline = ref(true)
let pingTimer: ReturnType<typeof setInterval> | undefined

// 校验后端是否可连接：探测公开的 /api/v1/ping（经同源代理，5s 超时）
async function pingBackend() {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), 5000)
  try {
    const res = await fetch('/api/v1/ping', { cache: 'no-store', signal: controller.signal })
    backendOnline.value = res.ok
  } catch {
    backendOnline.value = false
  } finally {
    clearTimeout(timeout)
  }
}

export function useOffline() {
  const isOnline = ref(navigator.onLine)
  const pendingActions = ref<QueuedAction[]>(
    JSON.parse(localStorage.getItem('offline_queue') || '[]')
  )

  function handleOnline() {
    isOnline.value = true
    // 网络恢复立即探测后端连通性
    pingBackend()
    syncPendingActions()
  }

  function handleOffline() {
    isOnline.value = false
    // 网络断开后端必然不可达，立即置为离线并触发顶部横幅
    backendOnline.value = false
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

  // 首次探测 + 15s 周期轮询后端连通性（后端不可达 → 「离线缓存提醒」横幅）
  if (pingTimer === undefined) {
    pingBackend()
    pingTimer = setInterval(pingBackend, 15000)
  }

  return {
    isOnline,
    backendOnline,
    pendingActions,
    enqueueAction,
    syncPendingActions,
  }
}
