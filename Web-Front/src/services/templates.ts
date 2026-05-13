import { get, post, put, del } from './api'
import type { Template } from '@/types'

export function fetchTemplates(params?: { type?: string }) {
  return get<Template[]>('/api/v1/templates', params as Record<string, unknown>)
}

export function fetchTemplateById(id: string) {
  return get<Template>(`/api/v1/templates/${id}`)
}

export function createTemplate(payload: {
  name: string
  type?: string
  fields?: string
  layout?: string
}) {
  return post<Template>('/api/v1/templates', payload)
}

export function updateTemplate(id: string, payload: {
  name?: string
  type?: string
  fields?: string
  layout?: string
}) {
  return put<Template>(`/api/v1/templates/${id}`, payload)
}

export function deleteTemplate(id: string) {
  return del<{ success: boolean }>(`/api/v1/templates/${id}`)
}