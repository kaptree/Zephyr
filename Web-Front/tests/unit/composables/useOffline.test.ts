import { describe, it, expect, vi } from 'vitest'
import { useOffline } from '@/composables/useOffline'

describe('useOffline', () => {
  it('初始状态应为在线', () => {
    const { isOnline } = useOffline()
    expect(isOnline.value).toBe(navigator.onLine)
  })

  it('pendingActions 应从 localStorage 恢复数据', () => {
    localStorage.setItem('offline_queue', JSON.stringify([
      { id: 'action-1', type: 'create', payload: {}, timestamp: Date.now() }
    ]))
    const { pendingActions } = useOffline()
    expect(pendingActions.value).toHaveLength(1)
    expect(pendingActions.value[0].id).toBe('action-1')
    localStorage.clear()
  })

  it('enqueueAction 应添加操作到队列', () => {
    const { pendingActions, enqueueAction } = useOffline()
    enqueueAction({ id: 'action-2', type: 'update', payload: {} })
    expect(pendingActions.value).toHaveLength(1)
    expect(pendingActions.value[0].type).toBe('update')
  })
})
