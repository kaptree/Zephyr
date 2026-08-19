import { get, post, postForm, del } from './api'
import type {
  NotificationItem,
  ChatMessageItem,
  ConversationItem,
  ReminderItem,
  PaginatedData,
} from '@/types'

// ---------------- 通知 ----------------

export function fetchNotifications(params?: {
  page?: number
  page_size?: number
  unread_only?: boolean
}) {
  return get<PaginatedData<NotificationItem>>('/api/v1/notifications', params)
}

export function fetchUnreadCount() {
  return get<{ count: number }>('/api/v1/notifications/unread-count')
}

export function markNotificationRead(id: string) {
  return post<{ success: boolean }>(`/api/v1/notifications/${id}/read`)
}

export function markAllNotificationsRead() {
  return post<{ success: boolean }>('/api/v1/notifications/read-all')
}

export function deleteNotification(id: string) {
  return del<{ success: boolean }>(`/api/v1/notifications/${id}`)
}

// ---------------- 聊天 ----------------

export function fetchConversations() {
  return get<ConversationItem[]>('/api/v1/chat/conversations')
}

// 当前在线用户 ID 列表
export function fetchChatOnline() {
  return get<{ online_ids: string[] }>('/api/v1/chat/online')
}

export function fetchChatMessages(peerId: string, params?: { page?: number; page_size?: number }) {
  return get<PaginatedData<ChatMessageItem>>(`/api/v1/chat/${peerId}/messages`, params)
}

// 聊天消息载荷（text / image / file）
export interface SendChatPayload {
  content?: string
  type?: 'text' | 'image' | 'file'
  note_id?: string
  file_name?: string
  file_path?: string
  file_size?: number
  mime_type?: string
}

export function sendChatMessage(peerId: string, payload: SendChatPayload) {
  return post<ChatMessageItem>(`/api/v1/chat/${peerId}/messages`, payload)
}

// 聊天文件上传（返回文件元数据，之后随消息发送）
export function uploadChatFile(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return postForm<{ file_name: string; file_path: string; file_size: number; mime_type: string }>(
    '/api/v1/chat/attachments',
    formData
  );
}

export function markConversationRead(peerId: string) {
  return post<{ success: boolean }>(`/api/v1/chat/${peerId}/read`)
}

// ---------------- 盯办提醒 ----------------

export function fetchReminders() {
  return get<ReminderItem[]>('/api/v1/reminders')
}

export function acknowledgeReminder(id: string) {
  return post<{ success: boolean }>(`/api/v1/reminders/${id}/acknowledge`)
}
