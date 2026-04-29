import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/services/admin', () => ({
  login: vi.fn().mockResolvedValue({
    code: 0,
    data: { access_token: 'test-token', refresh_token: '', expires_in: 7200, user: { id: 'user-1', username: 'zhangsan', name: '张三', avatar: '', email: '', phone: '', rank: '', dept_id: 'dept-1', dept_name: '刑警支队', role: 'group_leader', is_active: true, permissions: ['create_note_self', 'create_note_assigned', 'edit_others_note', 'delete_note', 'remind', 'view_dept_archive', 'view_group_archive', 'manage_tags', 'access_screen'] } },
  }),
  getCurrentUser: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'user-1', username: 'zhangsan', name: '张三', avatar: '', email: '', phone: '', rank: '', dept_id: 'dept-1', dept_name: '刑警支队', role: 'group_leader', is_active: true, permissions: ['create_note_self', 'create_note_assigned', 'edit_others_note', 'delete_note', 'remind', 'view_dept_archive', 'view_group_archive', 'manage_tags', 'access_screen'] },
  }),
}))

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('初始状态应为未登录', () => {
    const store = useAuthStore()
    expect(store.isLoggedIn).toBe(false)
    expect(store.user).toBeNull()
    expect(store.token).toBe('')
  })

  it('login 应正确更新 token 和 user', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isLoggedIn).toBe(true)
    expect(store.token).toBe('test-token')
    expect(store.user?.name).toBe('张三')
    expect(localStorage.getItem('auth_token')).toBe('test-token')
  })

  it('logout 应清除所有状态和 localStorage', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    store.logout()
    expect(store.isLoggedIn).toBe(false)
    expect(store.user).toBeNull()
    expect(store.token).toBe('')
    expect(localStorage.getItem('auth_token')).toBeFalsy()
  })

  it('isAdmin getter 应正确判断角色', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isAdmin).toBe(false)
  })

  it('isDeptAdmin getter 应正确判断角色', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isDeptAdmin).toBe(false)
  })

  it('isGroupLeader getter 应正确判断角色', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isGroupLeader).toBe(true)
  })

  it('canCreateForOthers 应正确判断 permission', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.canCreateForOthers).toBe(true)
  })

  it('updatePermissions 应正确更新权限数组', () => {
    const store = useAuthStore()
    store.updatePermissions(['manage_templates'])
    expect(store.permissions).toContain('manage_templates')
  })
})
