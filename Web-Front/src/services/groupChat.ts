import { get, post, put, del } from './api'
import type {
  GroupConversationItem,
  GroupMessageItem,
  GroupMemberItem,
  PaginatedData,
} from '@/types'

// ---------------- 群聊 ----------------

// 我所在的群聊列表（含成员数 / 未读数 / 最后一条消息）
export function fetchGroups() {
  return get<GroupConversationItem[]>('/api/v1/chat/groups')
}

export interface CreateGroupPayload {
  name: string
  members: string[]
}

// 创建群聊（创建者为群主，members 为初始成员用户 ID）
export function createGroup(payload: CreateGroupPayload) {
  return post<GroupConversationItem>('/api/v1/chat/groups', payload)
}

export function fetchGroupDetail(groupId: string) {
  return get<GroupConversationItem>(`/api/v1/chat/groups/${groupId}`)
}

export function fetchGroupMembers(groupId: string) {
  return get<GroupMemberItem[]>(`/api/v1/chat/groups/${groupId}/members`)
}

// 批量添加成员（群主）
export function addGroupMembers(groupId: string, userIds: string[]) {
  return post<{ added: number }>(`/api/v1/chat/groups/${groupId}/members`, { user_ids: userIds })
}

// 移除成员（群主）
export function removeGroupMember(groupId: string, userId: string) {
  return del<{ success: boolean }>(`/api/v1/chat/groups/${groupId}/members/${userId}`)
}

// 重命名群聊（群主）
export function renameGroup(groupId: string, name: string) {
  return put<{ success: boolean }>(`/api/v1/chat/groups/${groupId}`, { name })
}

// 解散群聊（群主）
export function dissolveGroup(groupId: string) {
  return del<{ success: boolean }>(`/api/v1/chat/groups/${groupId}`)
}

// 退出群聊（非群主）
export function quitGroup(groupId: string) {
  return post<{ success: boolean }>(`/api/v1/chat/groups/${groupId}/quit`)
}

export function fetchGroupMessages(groupId: string, params?: { page?: number; page_size?: number }) {
  return get<PaginatedData<GroupMessageItem>>(`/api/v1/chat/groups/${groupId}/messages`, params)
}

// 群消息载荷（text / image / file）
export interface SendGroupPayload {
  content?: string
  type?: 'text' | 'image' | 'file'
  file_name?: string
  file_path?: string
  file_size?: number
  mime_type?: string
}

export function sendGroupMessage(groupId: string, payload: SendGroupPayload) {
  return post<GroupMessageItem>(`/api/v1/chat/groups/${groupId}/messages`, payload)
}

export function markGroupRead(groupId: string) {
  return post<{ success: boolean }>(`/api/v1/chat/groups/${groupId}/read`)
}
