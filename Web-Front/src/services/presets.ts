import { get, post, put, del } from './api'
import type { PresetGroup, CreatePresetPayload } from '@/types/preset'

export function getPresets(templateType?: string) {
  return get<PresetGroup[]>('/api/v1/presets', templateType ? { template_type: templateType } : {})
}

export function getPreset(id: string) {
  return get<PresetGroup>(`/api/v1/presets/${id}`)
}

export function createPreset(payload: CreatePresetPayload) {
  return post<PresetGroup>('/api/v1/presets', payload)
}

export function updatePreset(id: string, payload: CreatePresetPayload) {
  return put<PresetGroup>(`/api/v1/presets/${id}`, payload)
}

export function deletePreset(id: string) {
  return del<{ success: boolean }>(`/api/v1/presets/${id}`)
}