export interface Participant {
  user_id: string
  user_name: string
  avatar: string
  dept_name: string
  role: 'initiator' | 'member' | 'leader'
  column_id: number
  is_online: boolean
}

export interface RemoteChangeData {
  column_id: number
  content: string
  delta: unknown
  user_id: string
  updated_by: string
}

// 需求29：协同房间指令（服务端 CollaborationCommand 结构）
export interface CommandMessage {
  id: string
  note_id: string
  user_id: string
  user_name: string
  content: string
  created_at: string
}

export interface CollaborationRoom {
  id: string
  note_id: string
  note_title: string
  participants: Participant[]
  columns: number
}
