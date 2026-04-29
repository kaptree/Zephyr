import type { Note, Tag, UserBrief, User } from '@/types'

export const mockTags: Tag[] = [
  { id: 'tag-1', name: '紧急', color: '#DC2626', scope: 'system', category: '优先级', usage_count: 15 },
  { id: 'tag-2', name: '情报研判', color: '#3B82F6', scope: 'system', category: '业务类型', usage_count: 8 },
  { id: 'tag-3', name: '会议纪要', color: '#22C55E', scope: 'system', category: '文档类型', usage_count: 5 },
  { id: 'tag-4', name: '个人备忘', color: '#8B5CF6', scope: 'personal', category: '其他', usage_count: 3 },
]

export const mockUserBriefs: UserBrief[] = [
  { id: 'user-1', name: '张三', avatar: '', dept_name: '刑警支队', role: 'group_leader' },
  { id: 'user-2', name: '李四', avatar: '', dept_name: '刑警支队', role: 'user' },
  { id: 'user-3', name: '王五', avatar: '', dept_name: '治安支队', role: 'user' },
]

export const mockUser: User = {
  id: 'user-1',
  name: '张三',
  avatar: '',
  email: 'zhangsan@police.gov.cn',
  phone: '13800001111',
  dept_id: 'dept-1',
  dept_name: '刑警支队',
  role: 'group_leader',
  permissions: [
    'create_note_self',
    'create_note_assigned',
    'edit_others_note',
    'delete_note',
    'remind',
    'view_dept_archive',
    'view_group_archive',
    'manage_tags',
    'access_screen',
  ],
}

export const mockAdminUser: User = {
  ...mockUser,
  id: 'admin-1',
  name: '管理员',
  dept_name: '公安局',
  role: 'super_admin',
  permissions: [
    'create_note_self',
    'create_note_assigned',
    'edit_others_note',
    'delete_note',
    'remind',
    'view_all_archive',
    'manage_departments',
    'manage_users',
    'manage_tags',
    'manage_templates',
    'access_screen',
    'send_command',
  ],
}

export function createMockNote(overrides: Partial<Note> = {}): Note {
  return {
    id: `note-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    title: '测试便签',
    content: '这是测试便签的内容，用于单元测试。',
    status: 'active',
    source_type: 'self',
    priority: 'normal',
    owner_id: 'user-1',
    creator_id: 'user-1',
    tags: [mockTags[0]],
    assignees: [],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
    ...overrides,
  }
}

export function createMockNotes(count: number): Note[] {
  return Array.from({ length: count }, (_, i) => createMockNote({
    id: `note-${i + 1}`,
    title: `便签 ${i + 1}`,
    content: `第${i + 1}条便签内容`,
  }))
}
