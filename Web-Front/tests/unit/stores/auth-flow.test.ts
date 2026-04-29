import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { mockUser } from '../../mocks/data'

describe('useAuthStore - API 登录流程', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('登录成功应设置 token 到 localStorage', async () => {
    const store = useAuthStore()
    store.$patch({
      token: 'my-test-token',
      user: mockUser,
      permissions: mockUser.permissions,
    })
    localStorage.setItem('auth_token', 'my-test-token')
    localStorage.setItem('auth_user', JSON.stringify(mockUser))
    
    expect(localStorage.getItem('auth_token')).toBe('my-test-token')
    expect(store.isLoggedIn).toBe(true)
  })

  it('登出应清除 token', () => {
    const store = useAuthStore()
    store.$patch({ token: 'token', user: mockUser, permissions: mockUser.permissions })
    store.logout()
    expect(store.isLoggedIn).toBe(false)
    expect(store.token).toBe('')
    expect(localStorage.getItem('auth_token')).toBeFalsy()
  })

  it('getter isAdmin 应对 super_admin 返回 true', () => {
    const store = useAuthStore()
    store.$patch({
      token: 't',
      user: { ...mockUser, role: 'super_admin' },
      permissions: (mockUser as any).permissions,
    })
    expect(store.isAdmin).toBe(true)
    expect(store.isDeptAdmin).toBe(true)
  })

  it('getter isDeptAdmin 应对 dept_admin 返回 true', () => {
    const store = useAuthStore()
    store.$patch({
      token: 't',
      user: { ...mockUser, role: 'dept_admin' },
      permissions: (mockUser as any).permissions,
    })
    expect(store.isDeptAdmin).toBe(true)
    expect(store.isAdmin).toBe(false)
  })

  it('普通用户 isAdmin 应为 false', () => {
    const store = useAuthStore()
    store.$patch({
      token: 't',
      user: { ...mockUser, role: 'user' },
      permissions: (mockUser as any).permissions,
    })
    expect(store.isAdmin).toBe(false)
    expect(store.isDeptAdmin).toBe(false)
    expect(store.isGroupLeader).toBe(false)
  })

  it('canCreateForOthers 在无对应权限时应返回 false', () => {
    const store = useAuthStore()
    store.$patch({
      token: 't',
      user: { ...mockUser, role: 'user' },
      permissions: ['create_note_self'],
    })
    expect(store.canCreateForOthers).toBe(false)
  })
})
