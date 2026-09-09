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
  /** 需求26：issue 评论通知关联的问题 id */
  issue_id?: string
  type: 'task_assigned' | 'task_completed' | 'task_feedback' | 'task_remind' | 'issue_comment' | 'issue_new' | 'system'
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

// ---------------- 群聊 ----------------

export interface GroupConversationItem {
  id: string
  name: string
  owner_id: string
  member_count: number
  unread: number
  last_msg: string
  last_type?: string
  last_at?: string | null
  created_at: string
}

export interface GroupMessageItem {
  id: string
  group_id: string
  sender_id: string
  /** 实时推送时附带的发送者姓名（历史消息通过 sender 关联查询） */
  sender_name?: string
  sender?: { id: string; username: string; name?: string; avatar?: string }
  type: 'text' | 'image' | 'file'
  content: string
  /** 被 @ 的成员 ID 列表（后端过滤为有效群成员；无 @ 时为空/缺省） */
  mentions?: string[]
  file_name?: string
  file_path?: string
  file_size?: number
  mime_type?: string
  created_at: string
}

export interface GroupMemberItem {
  id: string
  group_id: string
  user_id: string
  role: 'owner' | 'member'
  user?: {
    id: string
    username: string
    name?: string
    avatar?: string
    is_active?: boolean
    department?: { id: string; name: string }
  }
  joined_at: string
  last_read_at?: string | null
}

// 群文件夹文件（成员上传/下载，上传者或群主可删除）
export interface GroupFileItem {
  id: string
  group_id: string
  uploader_id: string
  uploader?: {
    id: string
    username: string
    name?: string
    avatar?: string
    department?: { id: string; name: string }
  }
  /** 实时推送时附带的上传者姓名（历史列表通过 uploader 关联查询） */
  uploader_name?: string
  file_name: string
  file_path: string
  file_size: number
  mime_type: string
  download_count: number
  created_at: string
}
