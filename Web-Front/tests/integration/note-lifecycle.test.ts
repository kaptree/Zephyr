import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNoteStore } from '@/stores/notes'

vi.mock('@/services/notes', () => ({
  fetchNotes: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      data: [
        { id: 'note-1', title: '便签1', content: '内容1', color_status: 'yellow', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
        { id: 'note-2', title: '便签2', content: '内容2', color_status: 'yellow', source_type: 'assigned', owner_id: 'user-2', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-02T00:00:00Z', updated_at: '2024-01-02T00:00:00Z' },
      ],
      total: 2, page: 1, page_size: 20,
    },
  }),
  createNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '已更新', content: '内容1', color_status: 'yellow', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
  }),
  updateNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '已更新', content: '内容1', color_status: 'yellow', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
  }),
  completeNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '便签1', content: '内容1', color_status: 'green', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString(), completed_at: new Date().toISOString() },
  }),
  remindNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-1', title: '便签1', content: '内容1', color_status: 'red', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 1, created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString() },
  }),
  archiveNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { success: true },
  }),
  restoreNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { id: 'note-restored', title: '便签1', content: '内容1', color_status: 'yellow', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: new Date().toISOString() },
  }),
}))

describe('便签完整生命周期 - 集成测试', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('创建 → 编辑 → 完成归档 完整流程', async () => {
    const store = useNoteStore()
    await store.fetchNotes()
    expect(store.activeNotes).toHaveLength(2)
    await store.completeNote('note-1')
    expect(store.activeNotes).toHaveLength(1)
  })

  it('创建 → 盯办提醒 → 变红 → 完成', async () => {
    const store = useNoteStore()
    store.$patch({
      activeNotes: [
        { id: 'note-1', title: '便签1', content: '内容1', color_status: 'yellow', source_type: 'self', owner_id: 'user-1', creator_id: 'user-1', is_archived: false, tags: [], assignees: [], remind_count: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' } as any,
      ],
    })
    await store.remindNote('note-1', '请尽快处理')
    expect(store.activeNotes[0].color_status).toBe('red')
    await store.completeNote('note-1')
    expect(store.activeNotes).toHaveLength(0)
  })
})
