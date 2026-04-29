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
  name: string
  avatar: string
  email: string
  phone: string
  dept_id: string
  dept_name: string
  role: Role
  permissions: Permission[]
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
  children?: Department[]
  member_count: number
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}
