export interface Template {
  id: string
  name: string
  type: string
  fields: TemplateField[]
  layout: string
  created_at: string
}

export interface TemplateField {
  id: string
  name: string
  type: 'text' | 'date' | 'select' | 'multi-select' | 'rich-text'
  required: boolean
  options?: string[]
  order: number
}

export interface Group {
  id: string
  name: string
  note_id?: string
  members: GroupMember[]
  created_at: string
}

export interface GroupMember {
  user_id: string
  user_name: string
  avatar: string
  role: 'leader' | 'member'
}

export interface SerialNumber {
  serial_no: string
  type: string
}

export interface LedgerEntry {
  date: string
  count: number
  dept_name: string
}

export interface LedgerSummary {
  data: LedgerEntry[]
  summary: {
    total: number
    by_dept: Record<string, number>
  }
}
