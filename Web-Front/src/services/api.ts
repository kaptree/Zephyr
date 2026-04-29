import axios from 'axios'
import type { ApiResponse } from '@/types'
import { isDemoMode } from './demoData'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '')

const api = axios.create({
  baseURL: BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    const { response } = error
    
    if (response?.status === 401 && !isDemoMode()) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
      window.location.href = '/login'
    }
    
    const errorMsg = response?.data?.message || response?.statusText || '网络请求异常'
    console.warn('[API Error]', errorMsg, response?.status)
    
    return Promise.reject(error)
  }
)

export async function get<T>(url: string, params?: Record<string, unknown>): Promise<ApiResponse<T>> {
  const response = await api.get<ApiResponse<T>>(url, { params })
  return response.data
}

export async function post<T>(url: string, data?: unknown): Promise<ApiResponse<T>> {
  const response = await api.post<ApiResponse<T>>(url, data)
  return response.data
}

export async function put<T>(url: string, data?: unknown): Promise<ApiResponse<T>> {
  const response = await api.put<ApiResponse<T>>(url, data)
  return response.data
}

export async function del<T>(url: string): Promise<ApiResponse<T>> {
  const response = await api.delete<ApiResponse<T>>(url)
  return response.data
}

export default api
