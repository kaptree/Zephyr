export interface PresetMember {
  user_id: string
  user?: {
    id: string
    name: string
    avatar: string
    dept_name: string
  }
  role: string
  sub_group_name: string
}

export interface PresetGroup {
  id: string
  name: string
  description: string
  template_type: string
  creator_id: string
  members: PresetMember[]
  created_at: string
  updated_at: string
}

export interface CreatePresetPayload {
  name: string
  description?: string
  template_type?: string
  members: { user_id: string; role?: string; sub_group_name?: string }[]
}