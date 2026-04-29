import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Note, NoteFilters, CreateNotePayload, UpdateNotePayload, PaginatedData } from '@/types'
import * as noteService from '@/services/notes'

type BackendNote = Note & Record<string, unknown>

function normalizeNote(raw: BackendNote): Note {
  return {
    id: raw.id,
    title: raw.title || '',
    content: raw.content || '',
    color_status: (raw.color_status as Note['color_status']) || 'yellow',
    source_type: (raw.source_type as Note['source_type']) || 'self',
    owner_id: raw.owner_id || '',
    creator_id: raw.creator_id || '',
    is_archived: !!raw.is_archived,
    tags: (raw.tags || []) as Note['tags'],
    assignees: (raw.assignees || []) as Note['assignees'],
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
  }
}

export const useNoteStore = defineStore('notes', () => {
  const activeNotes = ref<Note[]>([])
  const archivedNotes = ref<Note[]>([])
  const currentNote = ref<Note | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<NoteFilters>({ status: 'active', page: 1, page_size: 20 })
  const totalCount = ref(0)
  const currentPage = ref(1)

  const hasMore = computed(() => activeNotes.value.length < totalCount.value)

  async function fetchNotes(newFilters?: Partial<NoteFilters>) {
    loading.value = true
    error.value = null
    if (newFilters) {
      filters.value = { ...filters.value, ...newFilters }
    }
    try {
      const res = await noteService.fetchNotes(filters.value)
      const paginated = res.data as unknown as PaginatedData<Note>
      activeNotes.value = (paginated.data || []).map(normalizeNote)
      totalCount.value = paginated.total || 0
      currentPage.value = paginated.page || 1
    } catch (e: unknown) {
      const err = e as { response?: { status: number; data?: { message?: string } } }
      error.value = err.response?.data?.message || `加载失败（${err.response?.status || '网络错误'}）`
    } finally {
      loading.value = false
    }
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return
    loading.value = true
    try {
      const res = await noteService.fetchNotes({
        ...filters.value,
        page: currentPage.value + 1,
      })
      const paginated = res.data as unknown as PaginatedData<Note>
      activeNotes.value = [...activeNotes.value, ...(paginated.data || []).map(normalizeNote)]
      currentPage.value = paginated.page || currentPage.value
      totalCount.value = paginated.total || totalCount.value
    } catch (e: unknown) {
      const err = e as { response?: { status: number } }
      error.value = `加载更多失败（${err.response?.status || '网络错误'}）`
    } finally {
      loading.value = false
    }
  }

  async function createNote(payload: CreateNotePayload) {
    const res = await noteService.createNote(payload)
    const newNote = normalizeNote(res.data as unknown as BackendNote)
    activeNotes.value.unshift(newNote)
    return newNote
  }

  async function updateNoteLocally(id: string, payload: UpdateNotePayload) {
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index === -1) return
    const original = { ...activeNotes.value[index] }
    activeNotes.value[index] = { ...activeNotes.value[index], ...payload }
    try {
      const res = await noteService.updateNote(id, payload)
      activeNotes.value[index] = normalizeNote(res.data as unknown as BackendNote)
    } catch {
      activeNotes.value[index] = original
      throw new Error('更新失败')
    }
  }

  async function completeNote(id: string) {
    await noteService.completeNote(id)
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      activeNotes.value.splice(index, 1)
    }
  }

  async function remindNote(id: string, message?: string) {
    const res = await noteService.remindNote(id, { message, remind_type: 'urgent' })
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      activeNotes.value[index] = normalizeNote(res.data as unknown as BackendNote)
    }
  }

  async function archiveNote(id: string) {
    await noteService.archiveNote(id)
    activeNotes.value = activeNotes.value.filter(n => n.id !== id)
  }

  async function restoreNote(id: string) {
    const res = await noteService.restoreNote(id)
    const note = normalizeNote(res.data as unknown as BackendNote)
    activeNotes.value.unshift(note)
    archivedNotes.value = archivedNotes.value.filter(n => n.id !== id)
  }

  async function fetchArchivedNotes(archiveFilters: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await noteService.fetchNotes({ ...archiveFilters, status: 'archived' } as NoteFilters)
      const paginated = res.data as unknown as PaginatedData<Note>
      archivedNotes.value = (paginated.data || []).map(normalizeNote)
      totalCount.value = paginated.total || 0
    } finally {
      loading.value = false
    }
  }

  function setCurrentNote(note: Note | null) {
    currentNote.value = note
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
    completeNote,
    remindNote,
    archiveNote,
    restoreNote,
    fetchArchivedNotes,
    setCurrentNote,
  }
})
