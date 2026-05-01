import { get, post, put, del } from './api'
import type { ApiResponse, PaginatedData } from '@/types/api'
import type {
  AIConfigItem,
  AIConfigForm,
  ConfigFileItem,
  ConfigFileContent,
  ConfigFileHistoryItem,
  AdminLogItem,
  OperationLogItem,
} from '@/types/system'

export function getSystemConfig(): Promise<ApiResponse<Record<string, unknown>>> {
  return get('/api/v1/system/config')
}

export function updateSystemConfig(data: Record<string, unknown>): Promise<ApiResponse<Record<string, unknown>>> {
  return put('/api/v1/system/config', data)
}

export function listAIConfigs(): Promise<ApiResponse<AIConfigItem[]>> {
  return get('/api/v1/system/ai-configs')
}

export function createAIConfig(data: AIConfigForm): Promise<ApiResponse<AIConfigItem>> {
  return post('/api/v1/system/ai-configs', data)
}

export function updateAIConfig(id: string, data: Partial<AIConfigForm>): Promise<ApiResponse<null>> {
  return put(`/api/v1/system/ai-configs/${id}`, data)
}

export function deleteAIConfig(id: string): Promise<ApiResponse<null>> {
  return del(`/api/v1/system/ai-configs/${id}`)
}

export function listConfigFiles(): Promise<ApiResponse<ConfigFileItem[]>> {
  return get('/api/v1/system/config-files')
}

export function getConfigFile(name: string): Promise<ApiResponse<ConfigFileContent>> {
  return get(`/api/v1/system/config-files/${name}`)
}

export function updateConfigFile(name: string, content: string, changeSummary: string): Promise<ApiResponse<null>> {
  return put(`/api/v1/system/config-files/${name}`, { content, change_summary: changeSummary })
}

export function getConfigFileHistory(name: string): Promise<ApiResponse<ConfigFileHistoryItem[]>> {
  return get(`/api/v1/system/config-files/${name}/history`)
}

export function listAdminLogs(page: number, pageSize: number): Promise<ApiResponse<PaginatedData<AdminLogItem>>> {
  return get('/api/v1/system/logs', { page, page_size: pageSize })
}

export interface OperationLogQuery {
  page: number
  page_size: number
  user_id?: string
  user_name?: string
  action?: string
  method?: string
  date_from?: string
  date_to?: string
}

export function listOperationLogs(query: OperationLogQuery): Promise<ApiResponse<PaginatedData<OperationLogItem>>> {
  return get('/api/v1/system/operations', query as Record<string, unknown>)
}

export function getOperationActions(): Promise<ApiResponse<string[]>> {
  return get('/api/v1/system/operations/actions')
}
