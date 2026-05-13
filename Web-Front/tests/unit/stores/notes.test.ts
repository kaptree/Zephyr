import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useNoteStore } from '@/stores/notes';

vi.mock('@/services/notes', () => ({
  fetchNotes: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      data: [
        {
          id: 'note-1',
          title: '任务1',
          content: '内容1',
          color_status: 'yellow',
          source_type: 'self',
          owner_id: 'user-1',
          creator_id: 'user-1',
          is_archived: false,
          tags: [],
          assignees: [],
          remind_count: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'note-2',
          title: '任务2',
          content: '内容2',
          color_status: 'yellow',
          source_type: 'assigned',
          owner_id: 'user-2',
          creator_id: 'user-1',
          is_archived: false,
          tags: [],
          assignees: [],
          remind_count: 0,
          created_at: '2024-01-02T00:00:00Z',
          updated_at: '2024-01-02T00:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    },
  }),
  createNote: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      id: 'note-new',
      title: '新建任务',
      content: '内容',
      color_status: 'yellow',
      source_type: 'self',
      owner_id: 'user-1',
      creator_id: 'user-1',
      is_archived: false,
      tags: [],
      assignees: [],
      remind_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  }),
  updateNote: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      id: 'note-1',
      title: '已更新',
      content: '内容1',
      color_status: 'yellow',
      source_type: 'self',
      owner_id: 'user-1',
      creator_id: 'user-1',
      is_archived: false,
      tags: [],
      assignees: [],
      remind_count: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  }),
  completeNote: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      id: 'note-1',
      title: '任务1',
      content: '内容1',
      color_status: 'green',
      source_type: 'self',
      owner_id: 'user-1',
      creator_id: 'user-1',
      is_archived: false,
      tags: [],
      assignees: [],
      remind_count: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: new Date().toISOString(),
      completed_at: new Date().toISOString(),
    },
  }),
  remindNote: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      id: 'note-1',
      title: '任务1',
      content: '内容1',
      color_status: 'red',
      source_type: 'self',
      owner_id: 'user-1',
      creator_id: 'user-1',
      is_archived: false,
      tags: [],
      assignees: [],
      remind_count: 1,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: new Date().toISOString(),
    },
  }),
  archiveNote: vi.fn().mockResolvedValue({
    code: 0,
    data: { success: true },
  }),
  restoreNote: vi.fn().mockResolvedValue({
    code: 0,
    data: {
      id: 'note-restored',
      title: '任务1',
      content: '内容1',
      color_status: 'yellow',
      source_type: 'self',
      owner_id: 'user-1',
      creator_id: 'user-1',
      is_archived: false,
      tags: [],
      assignees: [],
      remind_count: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: new Date().toISOString(),
    },
  }),
}));

describe('useNoteStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  describe('fetchNotes', () => {
    it('应正确获取活跃任务列表', async () => {
      const store = useNoteStore();
      await store.fetchNotes();
      expect(store.activeNotes).toHaveLength(2);
      expect(store.totalCount).toBe(2);
    });
  });

  describe('createNote', () => {
    it('创建成功后应插入列表顶部', async () => {
      const store = useNoteStore();
      await store.createNote({ title: '新建任务', content: '内容', tags: [], source_type: 'self' });
      expect(store.activeNotes[0].title).toBe('新建任务');
    });
  });

  describe('completeNote', () => {
    it('完成后应从活跃列表移除', async () => {
      const store = useNoteStore();
      await store.fetchNotes();
      const count = store.activeNotes.length;
      await store.completeNote('note-1');
      expect(store.activeNotes.length).toBe(count - 1);
    });
  });

  describe('remindNote', () => {
    it('盯办后任务状态应变红色', async () => {
      const store = useNoteStore();
      store.$patch({
        activeNotes: [
          {
            id: 'note-1',
            title: '任务1',
            content: '内容1',
            color_status: 'yellow',
            source_type: 'self',
            owner_id: 'user-1',
            creator_id: 'user-1',
            is_archived: false,
            tags: [],
            assignees: [],
            remind_count: 0,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          } as any,
        ],
      });
      await store.remindNote('note-1');
      expect(store.activeNotes[0].color_status).toBe('red');
    });
  });

  describe('archiveNote', () => {
    it('归档后应从活跃列表移除', async () => {
      const store = useNoteStore();
      await store.fetchNotes();
      await store.archiveNote('note-1');
      expect(store.activeNotes).toHaveLength(1);
    });
  });

  describe('updateNoteTags', () => {
    const mockTag = {
      id: 'tag-1',
      name: '紧急',
      color: '#DC2626',
      scope: 'system' as const,
      category: '优先级',
      usage_count: 15,
    };

    it('应成功更新任务标签', async () => {
      const { updateNote } = await import('@/services/notes');
      const store = useNoteStore();
      store.$patch({
        activeNotes: [
          {
            id: 'note-1',
            title: '任务1',
            content: '内容1',
            color_status: 'yellow',
            source_type: 'self',
            owner_id: 'user-1',
            creator_id: 'user-1',
            is_archived: false,
            tags: [],
            assignees: [],
            remind_count: 0,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          } as any,
        ],
      });

      await store.updateNoteTags('note-1', ['tag-1']);
      expect(updateNote).toHaveBeenCalledWith('note-1', { tags: ['tag-1'] });
    });

    it('乐观更新应反映在 tags 中', async () => {
      const store = useNoteStore();
      store.$patch({
        activeNotes: [
          {
            id: 'note-1',
            title: '任务1',
            content: '内容1',
            color_status: 'yellow',
            source_type: 'self',
            owner_id: 'user-1',
            creator_id: 'user-1',
            is_archived: false,
            tags: [mockTag],
            assignees: [],
            remind_count: 0,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          } as any,
        ],
      });

      const promise = store.updateNoteTags('note-1', ['tag-1']);
      expect(store.activeNotes[0].tags).toHaveLength(1);
      expect((store.activeNotes[0].tags as any)[0].id).toBe('tag-1');
      await promise;
    });

    it('空数组应清空所有标签', async () => {
      const { updateNote } = await import('@/services/notes');
      const store = useNoteStore();
      store.$patch({
        activeNotes: [
          {
            id: 'note-1',
            title: '任务1',
            content: '内容1',
            color_status: 'yellow',
            source_type: 'self',
            owner_id: 'user-1',
            creator_id: 'user-1',
            is_archived: false,
            tags: [mockTag],
            assignees: [],
            remind_count: 0,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          } as any,
        ],
      });

      await store.updateNoteTags('note-1', []);
      expect(updateNote).toHaveBeenCalledWith('note-1', { tags: [] });
    });

    it('更新失败时应回滚标签', async () => {
      const { updateNote } = await import('@/services/notes');
      vi.mocked(updateNote).mockRejectedValueOnce(new Error('网络错误'));
      const store = useNoteStore();
      store.$patch({
        activeNotes: [
          {
            id: 'note-1',
            title: '任务1',
            content: '内容1',
            color_status: 'yellow',
            source_type: 'self',
            owner_id: 'user-1',
            creator_id: 'user-1',
            is_archived: false,
            tags: [mockTag],
            assignees: [],
            remind_count: 0,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          } as any,
        ],
      });

      await expect(store.updateNoteTags('note-1', ['tag-2'])).rejects.toThrow('标签更新失败');
      expect(store.activeNotes[0].tags).toHaveLength(1);
      expect((store.activeNotes[0].tags as any)[0].id).toBe('tag-1');
    });
  });
});
