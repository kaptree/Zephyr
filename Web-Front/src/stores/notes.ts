import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Note, NoteFilters, CreateNotePayload, UpdateNotePayload, PaginatedData } from '@/types'
import * as noteService from '@/services/notes'

export const useNoteStore = defineStore('notes', () => {
  const activeNotes = ref<Note[]>([])
  const archivedNotes = ref<Note[]>([])
  const currentNote = ref<Note | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<NoteFilters>({ status: undefined, page: 1, page_size: 20 })
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
      activeNotes.value = paginated.data
      totalCount.value = paginated.total
      currentPage.value = paginated.page
    } catch (e: unknown) {
      const err = e as { response?: { status: number } }
      error.value = `加载便签失败（${err.response?.status || '网络错误'}）`
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
      activeNotes.value = [...activeNotes.value, ...paginated.data]
      currentPage.value = paginated.page
      totalCount.value = paginated.total
    } catch (e: unknown) {
      const err = e as { response?: { status: number } }
      error.value = `加载更多失败（${err.response?.status || '网络错误'}）`
    } finally {
      loading.value = false
    }
  }

  async function createNote(payload: CreateNotePayload) {
    const res = await noteService.createNote(payload)
    const newNote = res.data as unknown as Note
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
      activeNotes.value[index] = res.data as unknown as Note
    } catch {
      activeNotes.value[index] = original
      throw new Error('更新失败')
    }
  }

  async function completeNote(id: string) {
    await noteService.completeNote(id)
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      const [completed] = activeNotes.value.splice(index, 1)
      completed.status = 'completed'
      completed.completed_at = new Date().toISOString()
      archivedNotes.value.unshift(completed)
    }
  }

  async function remindNote(id: string, message?: string) {
    const res = await noteService.remindNote(id, { message, remind_type: 'urgent' })
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      activeNotes.value[index] = res.data as unknown as Note
    }
  }

  async function archiveNote(id: string) {
    await noteService.archiveNote(id)
    activeNotes.value = activeNotes.value.filter(n => n.id !== id)
  }

  async function restoreNote(id: string) {
    const res = await noteService.restoreNote(id)
    const note = res.data as unknown as Note
    activeNotes.value.unshift(note)
    archivedNotes.value = archivedNotes.value.filter(n => n.id !== id)
  }

  async function fetchArchivedNotes(archiveFilters: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await noteService.fetchNotes({ ...archiveFilters, status: 'archived' } as NoteFilters)
      const paginated = res.data as unknown as PaginatedData<Note>
      archivedNotes.value = paginated.data
      totalCount.value = paginated.total
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
