import type { Tag } from './tag'
import type { UserBrief } from './user'

export type NoteStatus = 'active' | 'completed' | 'archived'
export type NoteSourceType = 'self' | 'assigned' | 'collaboration'
export type NotePriority = 'normal' | 'urgent'

export interface Note {
  id: string
  title: string
  content: string
  status: NoteStatus
  source_type: NoteSourceType
  priority: NotePriority
  owner_id: string
  creator_id: string
  tags: Tag[]
  assignees: UserBrief[]
  group_id?: string
  template_type?: string
  due_time?: string
  completed_at?: string
  archived_at?: string
  created_at: string
  updated_at: string
  allowed_actions: string[]
}

export interface CreateNotePayload {
  title: string
  content: string
  tags: string[]
  source_type: NoteSourceType
  due_time?: string
  owner_id?: string
  template_type?: string
  group_id?: string
  assignees?: string[]
}

export interface UpdateNotePayload {
  title?: string
  content?: string
  tags?: string[]
  status?: NoteStatus
  due_time?: string
}

export interface CompleteNotePayload {
  feedback_content?: string
  attachments?: string[]
}

export interface RemindPayload {
  remind_type?: 'urgent' | 'normal'
  message?: string
}

export interface NoteFilters {
  status?: NoteStatus
  tag_id?: string
  dept_id?: string
  owner_id?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface ArchiveFilters {
  date_from?: string
  date_to?: string
  tag_ids?: string[]
  user_id?: string
  dept_id?: string
  keyword?: string
  page?: number
  page_size?: number
}
