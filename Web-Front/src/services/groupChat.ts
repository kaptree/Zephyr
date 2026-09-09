import httpApi, { get, post, put, del, postForm } from './api'
import type {
  GroupConversationItem,
  GroupMessageItem,
  GroupMemberItem,
  GroupFileItem,
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
  /** 被 @ 的成员 ID 列表（后端过滤为有效群成员并随 WS 广播带回） */
  mentions?: string[]
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

// ---------------- 群公告 ----------------

// 设置群公告（仅群主；content 传空字符串表示清除公告）
export function setGroupAnnouncement(groupId: string, content: string) {
  return put<{ success: boolean }>(`/api/v1/chat/groups/${groupId}/announcement`, { content })
}

// 获取群公告
export function fetchGroupAnnouncement(groupId: string) {
  return get<{ announcement: string; announcement_updated_at: string | null }>(
    `/api/v1/chat/groups/${groupId}/announcement`
  )
}

// ---------------- 群文件夹 ----------------

// 成员上传群文件（multipart/form-data）
export function uploadGroupFile(groupId: string, file: File) {
  const form = new FormData()
  form.append('file', file)
  return postForm<GroupFileItem>(`/api/v1/chat/groups/${groupId}/files`, form)
}

// 群文件列表（按时间倒序）
export function fetchGroupFiles(groupId: string) {
  return get<GroupFileItem[]>(`/api/v1/chat/groups/${groupId}/files`)
}

// 删除群文件（上传者或群主）
export function deleteGroupFile(groupId: string, fileId: string) {
  return del<{ success: boolean }>(`/api/v1/chat/groups/${groupId}/files/${fileId}`)
}

// 下载群文件：带鉴权头请求 blob，再用 <a download> 触发保存（兼容中文文件名）
export async function downloadGroupFile(groupId: string, file: GroupFileItem) {
  const response = await httpApi.get<Blob>(
    `/api/v1/chat/groups/${groupId}/files/${file.id}/download`,
    { responseType: 'blob' }
  )
  const blob = response.data
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = file.file_name
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
