export interface AIConfigItem {
  id: string
  provider_name: string
  api_endpoint: string
  api_key_masked: string
  model_name: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AIConfigForm {
  provider_name: string
  api_endpoint: string
  api_key: string
  model_name: string
  description: string
  is_active?: boolean
}

export interface ConfigFileItem {
  name: string
  path: string
  description: string
}

export interface ConfigFileContent {
  name: string
  path: string
  content: string
  parsed: Record<string, unknown>
  size: number
}

export interface ConfigFileHistoryItem {
  id: string
  file_name: string
  content_before: string
  content_after: string
  changed_by: string
  changed_by_id: string
  change_summary: string
  created_at: string
}

export interface AdminLogItem {
  id: string
  admin_id: string
  admin_name: string
  action: string
  resource: string
  resource_id: string
  detail: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface OperationLogItem {
  id: string
  user_id: string
  user_name: string
  role: string
  action: string
  method: string
  path: string
  resource: string
  resource_id: string
  detail: string
  status_code: number
  ip_address: string
  created_at: string
}
