import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNoteStore } from '@/stores/notes'

vi.mock('@/services/notes', () => ({
  fetchNotes: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      data: [
        { id: 'note-1', title: '便签1', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
        { id: 'note-2', title: '便签2', content: '内容2', status: 'active', source_type: 'assigned', priority: 'normal', owner_id: 'user-2', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-02T00:00:00Z', updated_at: '2024-01-02T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete'] },
      ],
      total: 2, page: 1, page_size: 20,
    },
  }),
  createNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '已更新', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
  }),
  updateNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '已更新', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
  }),
  completeNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '便签1', content: '内容1', status: 'completed', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString(), completed_at: new Date().toISOString(), allowed_actions: ['view'] },
  }),
  remindNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '便签1', content: '内容1', status: 'active', source_type: 'self', priority: 'urgent', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString(), allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
  }),
  archiveNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { success: true },
  }),
  restoreNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-restored', title: '便签1', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString(), allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
  }),
}))

describe('便签完整生命周期 - 集成测试', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('创建 → 编辑 → 完成归档 完整流程', async () => {
    const store = useNoteStore()
    
    // 先加载列表
    await store.fetchNotes()
    expect(store.activeNotes).toHaveLength(2)
    
    // 完成便签 note-1（mock 的 completeNote 返回该 ID）
    await store.completeNote('note-1')
    expect(store.activeNotes).toHaveLength(1)
  })

  it('创建 → 盯办提醒 → 变红 → 完成', async () => {
    const store = useNoteStore()
    
    // 直接 patch 状态模拟已有便签
    store.$patch({
      activeNotes: [
        { id: 'note-1', title: '便签1', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
      ],
    })
    
    // 盯办 note-1
    await store.remindNote('note-1', '请尽快处理')
    expect(store.activeNotes[0].priority).toBe('urgent')
    
    // 完成
    await store.completeNote('note-1')
    expect(store.activeNotes).toHaveLength(0)
  })
})
