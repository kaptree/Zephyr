export interface NotificationItem {
  id: string
  recipient_id: string
  sender_id?: string
  sender?: {
    id: string
    username: string
    real_name?: string
    avatar?: string
  }
  note_id?: string
  type: 'task_assigned' | 'task_completed' | 'task_feedback' | 'task_remind' | 'system'
  title: string
  content: string
  is_read: boolean
  is_deleted: boolean
  read_at?: string
  created_at: string
}

export interface ChatMessageItem {
  id: string
  sender_id: string
  receiver_id: string
  note_id?: string
  type: 'text' | 'image' | 'file'
  content: string
  file_name?: string
  file_path?: string
  file_size?: number
  mime_type?: string
  is_read: boolean
  read_at?: string
  created_at: string
}

export interface ConversationItem {
  peer_id: string
  peer_name?: string
  peer_avatar?: string
  last_msg: string
  last_type?: string
  last_at: string
  unread: number
}

export interface ReminderItem {
  id: string
  note_id: string
  reminder_id: string
  target_id: string
  message: string
  remind_type: string
  is_acknowledged: boolean
  created_at: string
}
