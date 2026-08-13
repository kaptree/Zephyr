import { get, post, del } from './api'
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

export function fetchChatMessages(peerId: string, params?: { page?: number; page_size?: number }) {
  return get<PaginatedData<ChatMessageItem>>(`/api/v1/chat/${peerId}/messages`, params)
}

export function sendChatMessage(peerId: string, content: string, noteId?: string) {
  return post<ChatMessageItem>(`/api/v1/chat/${peerId}/messages`, {
    content,
    note_id: noteId,
  })
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
