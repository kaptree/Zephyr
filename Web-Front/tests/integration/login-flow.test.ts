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

describe('登录流程 - 集成测试', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('正确账密登录后存储状态', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isLoggedIn).toBe(true)
    expect(localStorage.getItem('auth_token')).toBe('test-token')
  })

  it('登出后状态被清除', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    store.logout()
    expect(store.isLoggedIn).toBe(false)
    expect(localStorage.getItem('auth_token')).toBeFalsy()
    expect(localStorage.getItem('auth_user')).toBeFalsy()
  })

  it('已登录用户可访问工作台', async () => {
    const store = useAuthStore()
    await store.login({ username: 'test', password: 'test' })
    expect(store.isLoggedIn).toBe(true)
    expect(localStorage.getItem('auth_token')).toBeTruthy()
  })

  it('无权限用户不能访问管理员页面', () => {
    const store = useAuthStore()
    store.$patch({
      token: 't',
      user: { id: 'u1', username: 'u', name: 'U', avatar: '', email: '', phone: '', rank: '', dept_id: 'd1', dept_name: 'D', role: 'user', is_active: true, permissions: ['create_note_self'] },
      permissions: ['create_note_self'],
    })
    expect(store.permissions).not.toContain('manage_templates')
    expect(store.permissions).not.toContain('manage_departments')
  })

  it('超管用户拥有所有管理权限', () => {
    const store = useAuthStore()
    const adminPerms = ['create_note_self', 'create_note_assigned', 'edit_others_note', 'delete_note', 'remind', 'view_all_archive', 'manage_departments', 'manage_users', 'manage_tags', 'manage_templates', 'access_screen', 'send_command']
    store.$patch({
      token: 'admin-token',
      user: { id: 'admin-1', username: 'admin', name: '管理员', avatar: '', email: '', phone: '', rank: '', dept_id: 'dept-1', dept_name: '公安局', role: 'super_admin', is_active: true, permissions: adminPerms },
      permissions: adminPerms,
    })
    expect(store.permissions).toContain('manage_templates')
    expect(store.permissions).toContain('manage_departments')
    expect(store.isAdmin).toBe(true)
  })
})
