import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type {
  Note,
  NoteFilters,
  CreateNotePayload,
  UpdateNotePayload,
  CompleteNotePayload,
  PaginatedData,
  NotePositionInput,
} from '@/types';
import * as noteService from '@/services/notes';

type BackendNote = Note & Record<string, unknown>;

function normalizeNote(raw: BackendNote): Note {
  return {
    id: raw.id,
    title: raw.title || '',
    content: raw.content || '',
    color_status: (raw.color_status as Note['color_status']) || 'yellow',
    source_type: (raw.source_type as Note['source_type']) || 'self',
    owner_id: raw.owner_id || '',
    creator_id: raw.creator_id || '',
    creator: raw.creator as Note['creator'],
    owner: raw.owner as Note['owner'],
    is_archived: !!raw.is_archived,
    is_pinned: !!raw.is_pinned,
    pinned_at: (raw.pinned_at as string | undefined) || null,
    pos_x: (raw.pos_x as number | null | undefined) ?? null,
    pos_y: (raw.pos_y as number | null | undefined) ?? null,
    tags: (raw.tags || []) as Note['tags'],
    assignees: (raw.assignees || []) as Note['assignees'],
    ccs: raw.ccs as Note['ccs'],
    group_id: raw.group_id as string | undefined,
    dept_id: raw.dept_id as string | undefined,
    template_type: raw.template_type as string | undefined,
    due_time: raw.due_time as string | undefined,
    completed_at: raw.completed_at as string | undefined,
    archive_time: raw.archive_time as string | undefined,
    remind_count: raw.remind_count || 0,
    serial_no: raw.serial_no as string | undefined,
    created_at: raw.created_at || new Date().toISOString(),
    updated_at: raw.updated_at || raw.created_at || new Date().toISOString(),
  };
}

export const useNoteStore = defineStore('notes', () => {
  const activeNotes = ref<Note[]>([]);
  const archivedNotes = ref<Note[]>([]);
  const currentNote = ref<Note | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const filters = ref<NoteFilters>({ status: 'active', page: 1, page_size: 20 });
  const totalCount = ref(0);
  const currentPage = ref(1);

  const hasMore = computed(() => activeNotes.value.length < totalCount.value);

  async function fetchNotes(newFilters?: Partial<NoteFilters>) {
    loading.value = true;
    error.value = null;
    if (newFilters) {
      filters.value = { ...filters.value, ...newFilters };
    }
    try {
      const res = await noteService.fetchNotes(filters.value);
      const paginated = res.data as unknown as PaginatedData<Note>;
      activeNotes.value = (paginated.data || []).map((raw) =>
        normalizeNote(raw as unknown as BackendNote)
      );
      totalCount.value = paginated.total || 0;
      currentPage.value = paginated.page || 1;
    } catch (e: unknown) {
      const err = e as { response?: { status: number; data?: { message?: string } } };
      error.value =
        err.response?.data?.message || `加载失败（${err.response?.status || '网络错误'}）`;
    } finally {
      loading.value = false;
    }
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return;
    loading.value = true;
    try {
      const res = await noteService.fetchNotes({
        ...filters.value,
        page: currentPage.value + 1,
      });
      const paginated = res.data as unknown as PaginatedData<Note>;
      activeNotes.value = [
        ...activeNotes.value,
        ...(paginated.data || []).map((raw) => normalizeNote(raw as unknown as BackendNote)),
      ];
      currentPage.value = paginated.page || currentPage.value;
      totalCount.value = paginated.total || totalCount.value;
    } catch (e: unknown) {
      const err = e as { response?: { status: number } };
      error.value = `加载更多失败（${err.response?.status || '网络错误'}）`;
    } finally {
      loading.value = false;
    }
  }

  async function createNote(payload: CreateNotePayload) {
    const res = await noteService.createNote(payload);
    const newNote = normalizeNote(res.data as unknown as BackendNote);
    activeNotes.value.unshift(newNote);
    return newNote;
  }

  async function updateNoteLocally(id: string, payload: UpdateNotePayload) {
    const index = activeNotes.value.findIndex((n) => n.id === id);
    if (index === -1) {
      await noteService.updateNote(id, payload);
      return;
    }
    const original = { ...activeNotes.value[index] };
    const optimistic: Record<string, unknown> = { ...payload };
    activeNotes.value[index] = { ...activeNotes.value[index], ...optimistic } as Note;
    try {
      const res = await noteService.updateNote(id, payload);
      activeNotes.value[index] = normalizeNote(res.data as unknown as BackendNote);
    } catch {
      activeNotes.value[index] = original;
      throw new Error('更新失败');
    }
  }

  async function updateNoteTags(id: string, tagIds: string[]) {
    const index = activeNotes.value.findIndex((n) => n.id === id);
    if (index === -1) {
      await noteService.updateNote(id, { tags: tagIds });
      return;
    }
    const originalTags = activeNotes.value[index].tags;
    const syntheticTags = tagIds.map((tid) => {
      const existing = (originalTags || []).find((t) => t.id === tid);
      return (
        existing || {
          id: tid,
          name: '',
          sub_tag: '',
          color: '#64748B',
          scope: 'personal' as const,
          category: '',
          usage_count: 0,
        }
      );
    });
    activeNotes.value[index] = { ...activeNotes.value[index], tags: syntheticTags };
    try {
      const res = await noteService.updateNote(id, { tags: tagIds });
      const updated = normalizeNote(res.data as unknown as BackendNote);
      activeNotes.value[index] = { ...activeNotes.value[index], tags: updated.tags };
    } catch {
      activeNotes.value[index] = { ...activeNotes.value[index], tags: originalTags };
      throw new Error('标签更新失败');
    }
  }

  /** 批量更新任务画布位置：乐观更新本地，保存失败时抛错由调用方处理 */
  async function updatePositions(items: NotePositionInput[]) {
    if (!items.length) return;
    const posMap = new Map(items.map((i) => [i.id, i]));
    const original = activeNotes.value;
    activeNotes.value = activeNotes.value.map((n) => {
      const p = posMap.get(n.id);
      return p ? { ...n, pos_x: p.pos_x, pos_y: p.pos_y } : n;
    });
    try {
      await noteService.updateNotePositions(items);
    } catch {
      activeNotes.value = original;
      throw new Error('位置保存失败');
    }
  }

  async function completeNote(id: string, payload?: CompleteNotePayload) {
    await noteService.completeNote(id, payload);
    const index = activeNotes.value.findIndex((n) => n.id === id);
    if (index !== -1) {
      activeNotes.value.splice(index, 1);
    }
  }

  async function remindNote(id: string, targetId: string, message?: string) {
    const res = await noteService.remindNote(id, {
      message,
      target_id: targetId,
      remind_type: 'urgent',
    });
    const index = activeNotes.value.findIndex((n) => n.id === id);
    if (index !== -1) {
      activeNotes.value[index] = normalizeNote(res.data as unknown as BackendNote);
    }
  }

  async function signNote(id: string) {
    const res = await noteService.signNote(id);
    const note = normalizeNote(res.data as unknown as BackendNote);
    const index = activeNotes.value.findIndex((n) => n.id === id);
    if (index !== -1) {
      activeNotes.value[index] = note;
    }
    return note;
  }

  async function archiveNote(id: string) {
    await noteService.archiveNote(id);
    activeNotes.value = activeNotes.value.filter((n) => n.id !== id);
  }

  async function restoreNote(id: string) {
    const res = await noteService.restoreNote(id);
    const note = normalizeNote(res.data as unknown as BackendNote);
    activeNotes.value.unshift(note);
    archivedNotes.value = archivedNotes.value.filter((n) => n.id !== id);
  }

  async function fetchArchivedNotes(archiveFilters: Record<string, unknown>) {
    loading.value = true;
    try {
      const res = await noteService.fetchNotes({
        ...archiveFilters,
        status: 'archived',
      } as NoteFilters);
      const paginated = res.data as unknown as PaginatedData<Note>;
      archivedNotes.value = (paginated.data || []).map((raw) =>
        normalizeNote(raw as unknown as BackendNote)
      );
      totalCount.value = paginated.total || 0;
    } finally {
      loading.value = false;
    }
  }

  function setCurrentNote(note: Note | null) {
    currentNote.value = note;
  }

  return {
    activeNotes,
    archivedNotes,
    currentNote,
    loading,
    error,
    filters,
    totalCount,
    currentPage,
    hasMore,
    fetchNotes,
    loadMore,
    createNote,
    updateNoteLocally,
    updateNoteTags,
    updatePositions,
    completeNote,
    remindNote,
    signNote,
    archiveNote,
    restoreNote,
    fetchArchivedNotes,
    setCurrentNote,
  };
});
