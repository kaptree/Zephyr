import { get, post } from './api'
import type { User, Department, LoginCredentials, LoginResponse, TreeNode } from '@/types'

export function login(credentials: LoginCredentials) {
  return post<LoginResponse>('/api/v1/auth/login', credentials)
}

export function getCurrentUser() {
  return get<User>('/api/v1/auth/me')
}

export function getDepartments(flat?: boolean) {
  return get<TreeNode[]>('/api/v1/departments', { flat: flat ? 'true' : 'false' })
}

export function getUsers(params: { dept_id?: string; role?: string; keyword?: string; page?: number }) {
  return get<{ data: User[]; total: number; page: number }>('/api/v1/users', params as Record<string, unknown>)
}

export function getVisibleUsers() {
  return get<User[]>('/api/v1/users/visible')
}

export function createUser(payload: Partial<User>) {
  return post<User>('/api/v1/users', payload)
}

export function updateUser(id: string, payload: Partial<User>) {
  return post<User>(`/api/v1/users/${id}`, payload)
}

export function deleteUser(id: string) {
  return post<{ success: boolean }>(`/api/v1/users/${id}/delete`)
}
