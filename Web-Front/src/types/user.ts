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

export interface User {
  id: string
  username: string
  name: string
  avatar: string
  email: string
  phone: string
  rank: string
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
