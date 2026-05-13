export type Role = 'super_admin' | 'dept_admin' | 'group_leader' | 'user' | 'screen_role'

export type Permission =
  | 'create_note_self'
  | 'create_note_assigned'
  | 'edit_others_note'
  | 'delete_note'
  | 'remind'
  | 'view_all_archive'
  | 'view_dept_archive'
  | 'view_group_archive'
  | 'manage_departments'
  | 'manage_users'
  | 'manage_tags'
  | 'manage_templates'
  | 'access_screen'
  | 'send_command'
  | 'manage_system'

export interface User {
  id: string
  username: string
  name: string
  avatar: string
  email: string
  phone: string
  rank: string
  position: string
  skills: string
  dept_id: string
  dept_name: string
  role: Role
  permissions: Permission[]
  is_active: boolean
}

export interface UserBrief {
  id: string
  name: string
  avatar: string
  dept_name: string
  role: Role
}

export interface WorkTypeStat {
  work_type: string
  group_count: number
}

export interface UserProfile extends User {
  work_type_stats: WorkTypeStat[]
}

export interface WorkTypeOption {
  value: string
  label: string
}

export const WORK_TYPE_LABELS: Record<string, string> = {
  default: '日常任务',
  data_analysis: '数据分析',
  special_project: '专项行动',
  emergency_canvas: '紧急协查',
  collaborative_writing: '协同作战',
}

export interface Department {
  id: string
  name: string
  parent_id: string | null
  level?: number
  children?: Department[]
  member_count: number
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}
