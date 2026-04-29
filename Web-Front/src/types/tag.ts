export interface Tag {
  id: string
  name: string
  color: string
  scope: 'personal' | 'system'
  category: string
  usage_count: number
}

export interface CreateTagPayload {
  name: string
  color: string
  category: string
  scope: 'personal' | 'system'
}
