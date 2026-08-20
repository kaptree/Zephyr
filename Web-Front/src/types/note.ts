import type { Tag } from './tag'
import type { UserBrief } from './user'

export type NoteStatus = 'active' | 'completed' | 'archived'
export type NoteSourceType = 'self' | 'assigned' | 'collaboration'

/** 任务被指派人（含签收与反馈填报信息） */
export interface NoteAssignee {
  note_id: string
  user_id: string
  user?: UserBrief
  role_in_note: string
  /** 签收状态：unsigned 未签收 / signed 已签收（默认 unsigned） */
  sign_status?: 'unsigned' | 'signed'
  signed_at?: string
  feedback_content?: string
  feedback_at?: string
  /** 被指派人本人是否已完成本人部分（需求23：全部完成后发起者才可归档） */
  is_completed?: boolean
  completed_at?: string
  is_read: boolean
}

/** 任务抄送人：创建任务时多选抄送，抄送人仅查看（紫色卡片 +「抄送」徽章） */
export interface NoteCc {
  note_id: string
  user_id: string
  user?: UserBrief
  created_at?: string
}

export interface Note {
  id: string
  title: string
  content: string
  color_status: 'yellow' | 'red' | 'green' | 'blue'
  source_type: NoteSourceType
  owner_id: string
  creator_id: string
  /** 创建人（后端 List/FindByID 已预加载，json: creator） */
  creator?: { id: string; name: string }
  /** 负责人（后端 List/FindByID 已预加载，json: owner） */
  owner?: { id: string; name: string }
  is_archived: boolean
  tags: Tag[]
  assignees: NoteAssignee[]
  /** 抄送人列表（需求20） */
  ccs?: NoteCc[]
  group_id?: string
  dept_id?: string
  template_type?: string
  due_time?: string
  work_time_seconds?: number
  due_remind_at?: string
  completed_at?: string
  archive_time?: string
  remind_count: number
  serial_no?: string
  created_at: string
  updated_at: string
}

export const NOTE_STATUS_LABEL: Record<string, string> = {
  active: '待办',
  completed: '已完成',
  archived: '已归档',
}

export const NOTE_COLOR_STATUS: Record<string, string> = {
  yellow: '#FEF3C7',
  red: '#FEE2E2',
  green: '#DCFCE7',
  blue: '#DBEAFE',
}

export const NOTE_COLOR_BORDER: Record<string, string> = {
  yellow: '#D97706',
  red: '#DC2626',
  green: '#16A34A',
  blue: '#2563EB',
}

export interface CreateNotePayload {
  title: string
  content: string
  tags: string[]
  source_type: NoteSourceType
  due_time?: string
  work_time_seconds?: number
  owner_id?: string
  template_type?: string
  assignees?: string[]
}

export interface UpdateNotePayload {
  title?: string
  content?: string
  tags?: string[]
  due_time?: string
  color_status?: string
  owner_id?: string
}

export interface CompleteNotePayload {
  feedback_content?: string
  attachments?: string[]
}

export interface RemindPayload {
  remind_type?: 'urgent' | 'normal'
  message?: string
  target_id: string
}

export interface NoteFilters {
  status?: string
  tag_ids?: string[]
  department_id?: string
  owner_id?: string
  keyword?: string
  /** 起始日期 YYYY-MM-DD（按创建时间筛选） */
  date_from?: string
  /** 截止日期 YYYY-MM-DD（含当天） */
  date_to?: string
  page?: number
  page_size?: number
}

export interface ArchiveFilters {
  date_from?: string
  date_to?: string
  tag_ids?: string[]
  user_id?: string
  department_id?: string
  keyword?: string
  page?: number
  page_size?: number
}
