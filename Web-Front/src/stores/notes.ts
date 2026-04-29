import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Note, NoteFilters, CreateNotePayload, UpdateNotePayload, PaginatedData } from '@/types'
import * as noteService from '@/services/notes'
import {
  DEMO_NOTES,
  filterDemoNotes,
  createDemoNote,
  isDemoMode,
} from '@/services/demoData'

const BACKEND_DOWN_KEY = 'backend_down_at'

function skipBackendCall(): boolean {
  if (!isDemoMode()) return false
  const at = localStorage.getItem(BACKEND_DOWN_KEY)
  if (!at) return false
  return Date.now() - Number(at) < 30_000
}

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

  function isNetworkError(e: unknown): boolean {
    const err = e as { code?: string; response?: { status: number }; message?: string }
    if (!err.response) return true
    if (err.response.status === 404) return true
    if (err.response.status >= 500) return true
    const networkCodes = ['ECONNABORTED', 'ERR_NETWORK', 'ECONNREFUSED', 'ETIMEDOUT', 'ERR_CONNECTION_REFUSED', 'ERR_BAD_RESPONSE']
    if (err.code && networkCodes.includes(err.code)) return true
    if (err.message?.includes('timeout')) return true
    return false
  }

  async function fetchNotes(newFilters?: Partial<NoteFilters>) {
    loading.value = true
    error.value = null
    if (newFilters) {
      filters.value = { ...filters.value, ...newFilters }
    }

    // 演示模式 + 已知后端不通 → 直接走本地数据
    if (skipBackendCall()) {
      loadDemoActiveNotes()
      loading.value = false
      return
    }

    // 优先请求后端
    try {
      const res: PaginatedData<Note> = (await noteService.fetchNotes(filters.value)).data as unknown as PaginatedData<Note>
      activeNotes.value = res.data
      totalCount.value = res.total
      currentPage.value = res.page
      loading.value = false
      return
    } catch (e) {
      if (!isNetworkError(e)) {
        error.value = '加载便签失败'
        loading.value = false
        throw e
      }
    }

    // 降级到演示数据
    loadDemoActiveNotes()
    loading.value = false
  }

  function loadDemoActiveNotes() {
    if (!isDemoMode()) return
    const allActive = filterDemoNotes('active')
    let filtered = allActive
    if (filters.value.keyword) {
      const kw = filters.value.keyword.toLowerCase()
      filtered = filtered.filter(n => n.title.toLowerCase().includes(kw) || n.content.toLowerCase().includes(kw))
    }
    if (filters.value.owner_id) {
      filtered = filtered.filter(n => n.owner_id === filters.value.owner_id)
    }
    activeNotes.value = filtered
    totalCount.value = filtered.length
    currentPage.value = 1
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return
    // 演示模式下不支持分页（一次性加载全部）
    loading.value = false
    return
  }

  async function createNote(payload: CreateNotePayload) {
    // 优先请求后端
    try {
      const res = await noteService.createNote(payload)
      const newNote = res.data as unknown as Note
      activeNotes.value.unshift(newNote)
      return newNote
    } catch (e) {
      if (!isNetworkError(e)) throw e
    }

    // 降级到演示数据
    if (isDemoMode()) {
      const demoNote = createDemoNote({
        title: payload.title,
        content: payload.content,
        tags: payload.tags,
        source_type: payload.source_type,
        owner_id: payload.owner_id,
      })
      activeNotes.value.unshift(demoNote)
      return demoNote
    }
    throw new Error('创建失败')
  }

  async function updateNoteLocally(id: string, payload: UpdateNotePayload) {
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index === -1) return
    const original = { ...activeNotes.value[index] }

    // 乐观更新
    activeNotes.value[index] = { ...activeNotes.value[index], ...payload }

    try {
      const res = await noteService.updateNote(id, payload)
      activeNotes.value[index] = res.data as unknown as Note
    } catch (e) {
      if (!isNetworkError(e)) {
        activeNotes.value[index] = original
        throw new Error('更新失败')
      }
      // 演示模式下接受乐观更新
    }
  }

  async function completeNote(id: string) {
    try {
      await noteService.completeNote(id)
    } catch (e) {
      if (!isNetworkError(e)) throw e
    }
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      const [completed] = activeNotes.value.splice(index, 1)
      completed.status = 'completed'
      completed.completed_at = new Date().toISOString()
      archivedNotes.value.unshift(completed)
    }
  }

  async function remindNote(id: string, message?: string) {
    try {
      const res = await noteService.remindNote(id, { message, remind_type: 'urgent' })
      const index = activeNotes.value.findIndex(n => n.id === id)
      if (index !== -1) {
        activeNotes.value[index] = res.data as unknown as Note
      }
    } catch (e) {
      if (!isNetworkError(e)) throw e
      // 演示模式下本地修改
      const index = activeNotes.value.findIndex(n => n.id === id)
      if (index !== -1) {
        activeNotes.value[index] = { ...activeNotes.value[index], priority: 'urgent' }
      }
    }
  }

  async function archiveNote(id: string) {
    try {
      await noteService.archiveNote(id)
    } catch (e) {
      if (!isNetworkError(e)) throw e
    }
    const index = activeNotes.value.findIndex(n => n.id === id)
    if (index !== -1) {
      const [archived] = activeNotes.value.splice(index, 1)
      archived.status = 'archived'
      archived.archived_at = new Date().toISOString()
      archivedNotes.value.unshift(archived)
    }
  }

  async function restoreNote(id: string) {
    try {
      const res = await noteService.restoreNote(id)
      const note = res.data as unknown as Note
      activeNotes.value.unshift(note)
      archivedNotes.value = archivedNotes.value.filter(n => n.id !== id)
    } catch (e) {
      if (!isNetworkError(e)) throw e
      // 演示模式下本地恢复
      const index = archivedNotes.value.findIndex(n => n.id === id)
      if (index !== -1) {
        const [restored] = archivedNotes.value.splice(index, 1)
        restored.status = 'active'
        activeNotes.value.unshift(restored)
      }
    }
  }

  async function fetchArchivedNotes(archiveFilters: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await noteService.fetchNotes({ ...archiveFilters, status: 'archived' } as NoteFilters)
      const paginated = res.data as unknown as PaginatedData<Note>
      archivedNotes.value = paginated.data
      totalCount.value = paginated.total
    } catch (e) {
      if (!isNetworkError(e)) {
        loading.value = false
        throw e
      }
      // 降级到演示数据
      if (isDemoMode()) {
        const allArchived = filterDemoNotes('archived')
        const kw = archiveFilters.keyword as string | undefined
        let filtered = allArchived
        if (kw) {
          const lower = kw.toLowerCase()
          filtered = filtered.filter(n => n.title.toLowerCase().includes(lower))
        }
        archivedNotes.value = filtered
        totalCount.value = filtered.length
      }
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
