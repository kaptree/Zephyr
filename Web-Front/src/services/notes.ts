import { get, post, put, del } from './api'
import type { Note, CreateNotePayload, UpdateNotePayload, NoteFilters, PaginatedData, CompleteNotePayload, RemindPayload } from '@/types'

export function fetchNotes(filters: NoteFilters) {
  const params: Record<string, unknown> = {}
  if (filters.status) params.status = filters.status
  if (filters.tag_ids?.length) params.tag_ids = filters.tag_ids
  if (filters.department_id) params.department_id = filters.department_id
  if (filters.owner_id) params.owner_id = filters.owner_id
  if (filters.keyword) params.keyword = filters.keyword
  if (filters.page) params.page = filters.page
  if (filters.page_size) params.page_size = filters.page_size
  return get<PaginatedData<Note>>('/api/v1/notes', params)
}

export function fetchNoteById(id: string) {
  return get<Note>(`/api/v1/notes/${id}`)
}

export function createNote(payload: CreateNotePayload) {
  return post<Note>('/api/v1/notes', payload)
}

export function updateNote(id: string, payload: UpdateNotePayload) {
  return put<Note>(`/api/v1/notes/${id}`, payload)
}

export function completeNote(id: string, payload?: CompleteNotePayload) {
  return post<Note>(`/api/v1/notes/${id}/complete`, payload)
}

export function remindNote(id: string, payload?: RemindPayload) {
  return post<Note>(`/api/v1/notes/${id}/remind`, payload)
}

export function archiveNote(id: string) {
  return del<{ success: boolean }>(`/api/v1/notes/${id}`)
}

export function restoreNote(id: string) {
  return post<Note>(`/api/v1/notes/${id}/restore`)
}

export function exportNote(id: string, templateId?: string) {
  return post<{ download_url: string }>(`/api/v1/notes/${id}/export`, {
    format: 'word',
    template_id: templateId,
  })
}

export function fetchNoteStats(params?: { days?: number; dept_id?: string; status?: string }) {
  return get<{ total_notes: number; active_notes: number; trend: { date: string; count: number }[] }>('/api/v1/notes/stats', params as Record<string, unknown>)
}

export function fetchHeatmap(year: number) {
  return get<{ total_archived: number; year: number; daily: { date: string; count: number }[] }>('/api/v1/notes/heatmap', { year } as Record<string, unknown>)
}
