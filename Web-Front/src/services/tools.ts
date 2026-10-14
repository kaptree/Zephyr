import { get, post, postForm, put, del } from './api'

/** 小工具（工具超市条目） */
export interface ToolItem {
  id: string
  slug: string
  name: string
  description: string
  /** letter / emoji / image */
  icon_type: 'letter' | 'emoji' | 'image'
  /** 首字母 / emoji 字符 / 图片路径或 data URI */
  icon_value: string
  category: string
  tags: string
  /** pending / published / rejected / offline */
  status: 'pending' | 'published' | 'rejected' | 'offline'
  author_name: string
  is_system: boolean
  launch_count: number
  week_launch_count: number
  created_at: string
}

/** 上传工具的表单载荷 */
export interface ToolUploadPayload {
  file: File
  icon?: File
  name?: string
  description?: string
  category?: string
  tags?: string
}

export type ToolSort = 'hot' | 'latest'

/** 工具超市列表（published + 本人上传，服务端支持关键词/分类/排序） */
export function fetchTools(params: {
  keyword?: string
  category?: string
  sort?: ToolSort
}) {
  return get<{ list: ToolItem[] }>('/api/v1/tools', params)
}

/** 分类列表 */
export function fetchToolCategories() {
  return get<{ list: string[] }>('/api/v1/tools/categories')
}

/** 我上传的工具（含待审核） */
export function fetchMyTools() {
  return get<{ list: ToolItem[] }>('/api/v1/tools/mine')
}

/** 待审核列表（管理员） */
export function fetchPendingTools() {
  return get<{ list: ToolItem[] }>('/api/v1/tools/pending')
}

/** 工具详情 */
export function fetchToolDetail(slug: string) {
  return get<ToolItem>(`/api/v1/tools/${slug}`)
}

/** 上传 HTML 工具 */
export function uploadTool(payload: ToolUploadPayload) {
  const fd = new FormData()
  fd.append('file', payload.file)
  if (payload.icon) fd.append('icon', payload.icon)
  if (payload.name) fd.append('name', payload.name)
  if (payload.description) fd.append('description', payload.description)
  if (payload.category) fd.append('category', payload.category)
  if (payload.tags) fd.append('tags', payload.tags)
  return postForm<ToolItem>('/api/v1/tools', fd)
}

/** 上报启动（返回会话 ID，同一会话刷新去重） */
export function launchTool(slug: string, sessionId = '') {
  return post<{ session_id: string }>(`/api/v1/tools/${slug}/launch`, {
    session_id: sessionId,
  })
}

/** 上报停留时长（离开运行页时尽力上报） */
export function exitTool(slug: string, sessionId: string, duration: number) {
  return post<null>(`/api/v1/tools/${slug}/exit`, {
    session_id: sessionId,
    duration: Math.max(0, Math.round(duration)),
  })
}

/** 审核 / 上下架（管理员） */
export function updateToolStatus(slug: string, status: 'published' | 'rejected' | 'offline') {
  return put<ToolItem>(`/api/v1/tools/${slug}/status`, { status })
}

/** 删除工具（管理员或作者本人） */
export function deleteTool(slug: string) {
  return del<null>(`/api/v1/tools/${slug}`)
}

/** 工具运行页 iframe 地址（后端返回带沙箱 CSP 的入口 HTML） */
export function toolRunUrl(slug: string) {
  return `/api/v1/tools/${slug}/run`
}

/** 状态中文标签 */
export const TOOL_STATUS_LABEL: Record<ToolItem['status'], string> = {
  pending: '待审核',
  published: '已上架',
  rejected: '已打回',
  offline: '已下线',
}
