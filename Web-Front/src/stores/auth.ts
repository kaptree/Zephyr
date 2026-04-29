import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, Permission, LoginCredentials } from '@/types'
import { login as loginApi, getCurrentUser } from '@/services/admin'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(loadUser())
  const token = ref<string>(localStorage.getItem('auth_token') || '')
  const permissions = ref<Permission[]>(user.value?.permissions || [])

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

  async function login(credentials: LoginCredentials) {
    const res = await loginApi(credentials)
    applyUser(res.data.user, res.data.access_token)
  }

  async function fetchUserInfo() {
    try {
      const res = await getCurrentUser()
      user.value = res.data
      permissions.value = res.data.permissions || []
      localStorage.setItem('auth_user', JSON.stringify(res.data))
    } catch {
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
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
  }

  return {
    user,
    token,
    permissions,
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
