import { get, post, put, del } from './api'
import type { Tag, PaginatedData, CreateTagPayload } from '@/types'

export function fetchTags(scope?: 'personal' | 'system' | 'all') {
  return get<PaginatedData<Tag>>('/api/v1/tags', { scope })
}

export function createTag(payload: { name: string; sub_tag?: string; color?: string; category?: string; scope?: string }) {
  return post<Tag>('/api/v1/tags', payload)
}

export function updateTag(id: string, payload: Partial<CreateTagPayload>) {
  return put<Tag>(`/api/v1/tags/${id}`, payload)
}

export function deleteTag(id: string) {
  return del<{ success: boolean }>(`/api/v1/tags/${id}`)
}
