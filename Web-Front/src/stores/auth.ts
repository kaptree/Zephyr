import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, Permission, LoginCredentials } from '@/types'
import { login as loginApi, getCurrentUser } from '@/services/admin'
import { getDemoUser, setDemoMode, isDemoMode } from '@/services/demoData'

const BACKEND_DOWN_KEY = 'backend_down_at'

function isBackendRecentlyDown(): boolean {
  const at = localStorage.getItem(BACKEND_DOWN_KEY)
  if (!at) return false
  // 30秒内后端不可达则跳过重试
  return Date.now() - Number(at) < 30_000
}

function markBackendDown() {
  localStorage.setItem(BACKEND_DOWN_KEY, String(Date.now()))
}

function markBackendUp() {
  localStorage.removeItem(BACKEND_DOWN_KEY)
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(loadUser())
  const token = ref<string>(localStorage.getItem('auth_token') || '')
  const permissions = ref<Permission[]>(user.value?.permissions || [])
  const demoMode = ref(isDemoMode())

  function loadUser(): User | null {
    const stored = localStorage.getItem('auth_user')
    if (stored) {
      try { return JSON.parse(stored) } catch { return null }
    }
    return null
  }

  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'super_admin')
  const isDeptAdmin = computed(() => user.value?.role === 'dept_admin' || user.value?.role === 'super_admin')
  const isGroupLeader = computed(() => user.value?.role === 'group_leader' || isDeptAdmin.value || isAdmin.value)
  const canCreateForOthers = computed(() => permissions.value.includes('create_note_assigned'))

  function applyUser(u: User, t: string) {
    token.value = t
    user.value = u
    permissions.value = u.permissions || []
    localStorage.setItem('auth_token', t)
    localStorage.setItem('auth_user', JSON.stringify(u))
  }

  function isAxiosNetworkError(e: unknown): boolean {
    const err = e as { code?: string; response?: { status: number }; message?: string }
    // 无 response → 网络不可达
    if (!err.response) return true
    // 404 → 后端未部署，降级到演示模式
    if (err.response.status === 404) return true
    // 5xx → 服务端异常，降级
    if (err.response.status >= 500) return true
    // axios 常见网络错误码
    const networkCodes = ['ECONNABORTED', 'ERR_NETWORK', 'ECONNREFUSED', 'ETIMEDOUT', 'ERR_CONNECTION_REFUSED', 'ERR_BAD_RESPONSE']
    if (err.code && networkCodes.includes(err.code)) return true
    // 请求超时消息
    if (err.message?.includes('timeout')) return true
    return false
  }

  async function login(credentials: LoginCredentials) {
    // 如果最近刚知道后端不通，直接走演示模式
    if (isBackendRecentlyDown()) {
      const demo = getDemoUser(credentials.username)
      if (demo && demo.password === credentials.password) {
        applyUser(demo.user, 'demo-token-' + credentials.username)
        setDemoMode(true)
        demoMode.value = true
        return
      }
      throw new Error('用户名或密码错误')
    }

    // 优先请求后端
    try {
      const res = await loginApi(credentials)
      markBackendUp()
      applyUser(res.data.user, res.data.token)
      setDemoMode(false)
      demoMode.value = false
      return
    } catch (e: unknown) {
      if (!isAxiosNetworkError(e)) {
        throw e
      }
      markBackendDown()
    }

    // 降级：走前端演示登录
    const demo = getDemoUser(credentials.username)
    if (!demo || demo.password !== credentials.password) {
      throw new Error('用户名或密码错误')
    }
    applyUser(demo.user, 'demo-token-' + credentials.username)
    setDemoMode(true)
    demoMode.value = true
  }

  async function fetchUserInfo() {
    try {
      const res = await getCurrentUser()
      user.value = res.data
      permissions.value = res.data.permissions || []
      localStorage.setItem('auth_user', JSON.stringify(res.data))
    } catch {
      if (isDemoMode()) return
      logout()
    }
  }

  function updatePermissions(newPermissions: Permission[]) {
    permissions.value = newPermissions
    if (user.value) {
      user.value.permissions = newPermissions
      localStorage.setItem('auth_user', JSON.stringify(user.value))
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    permissions.value = []
    demoMode.value = false
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    localStorage.removeItem(BACKEND_DOWN_KEY)
    setDemoMode(false)
  }

  return {
    user,
    token,
    permissions,
    demoMode,
    isLoggedIn,
    isAdmin,
    isDeptAdmin,
    isGroupLeader,
    canCreateForOthers,
    login,
    fetchUserInfo,
    updatePermissions,
    logout,
  }
})
