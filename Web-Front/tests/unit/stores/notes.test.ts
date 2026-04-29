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
    data: { id: 'note-new', title: '新建便签', content: '内容', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: new Date().toISOString(), updated_at: new Date().toISOString(), allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
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

describe('useNoteStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('fetchNotes', () => {
    it('应正确获取活跃便签列表', async () => {
      const store = useNoteStore()
      await store.fetchNotes()
      expect(store.activeNotes).toHaveLength(2)
      expect(store.totalCount).toBe(2)
    })
  })

  describe('createNote', () => {
    it('创建成功后应插入列表顶部', async () => {
      const store = useNoteStore()
      await store.createNote({ title: '新建便签', content: '内容', tags: [], source_type: 'self' })
      expect(store.activeNotes[0].title).toBe('新建便签')
    })
  })

  describe('completeNote', () => {
    it('完成后应从活跃列表移除', async () => {
      const store = useNoteStore()
      await store.fetchNotes()
      const count = store.activeNotes.length
      await store.completeNote('note-1')
      expect(store.activeNotes.length).toBe(count - 1)
    })
  })

  describe('remindNote', () => {
    it('盯办后便签优先级应变 urgent', async () => {
      const store = useNoteStore()
      // 直接 patch 状态模拟已有便签
      store.$patch({
        activeNotes: [
          { id: 'note-1', title: '便签1', content: '内容1', status: 'active', source_type: 'self', priority: 'normal', owner_id: 'user-1', creator_id: 'user-1', tags: [], assignees: [], created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z', allowed_actions: ['edit', 'delete', 'complete', 'remind'] },
        ],
      })
      await store.remindNote('note-1')
      expect(store.activeNotes[0].priority).toBe('urgent')
    })
  })

  describe('archiveNote', () => {
    it('归档后应从活跃列表移除', async () => {
      const store = useNoteStore()
      await store.fetchNotes()
      await store.archiveNote('note-1')
      expect(store.activeNotes).toHaveLength(1)
    })
  })
})
