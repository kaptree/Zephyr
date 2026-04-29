import { get, post, put, del } from './api'
import type { Note, CreateNotePayload, UpdateNotePayload, NoteFilters, PaginatedData, CompleteNotePayload, RemindPayload } from '@/types'

export function fetchNotes(filters: NoteFilters) {
  return get<PaginatedData<Note>>('/api/v1/notes', filters as Record<string, unknown>)
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
