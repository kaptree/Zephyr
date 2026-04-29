import type { User, Note, Tag, Department } from '@/types'

// ============================================================
//  演示账号：前后端接口不可用时，使用前端演示数据
//  策略：优先请求后端 API，网络不通 / 鉴权失败时降级到此处数据
// ============================================================

// ---------------------- 演示账号 ----------------------
export const DEMO_ACCOUNTS: Record<string, { password: string; user: User }> = {
  admin: {
    password: 'admin123',
    user: {
      id: 'u-admin-001',
      name: '李建国',
      avatar: '',
      email: 'lijg@police.gov.cn',
      phone: '13800001001',
      dept_id: 'dept-001',
      dept_name: '市公安局',
      role: 'super_admin',
      permissions: [
        'create_note_self', 'create_note_assigned', 'edit_others_note',
        'delete_note', 'remind', 'view_all_archive',
        'manage_departments', 'manage_users', 'manage_tags', 'manage_templates',
        'access_screen', 'send_command',
      ],
    },
  },
  zhang: {
    password: '123456',
    user: {
      id: 'u-dept-001',
      name: '张振华',
      avatar: '',
      email: 'zhangzh@police.gov.cn',
      phone: '13800001002',
      dept_id: 'dept-002',
      dept_name: '刑警支队',
      role: 'dept_admin',
      permissions: [
        'create_note_self', 'create_note_assigned', 'edit_others_note',
        'delete_note', 'remind', 'view_dept_archive',
        'manage_departments', 'manage_users', 'manage_tags',
        'access_screen',
      ],
    },
  },
  wang: {
    password: '123456',
    user: {
      id: 'u-leader-001',
      name: '王明辉',
      avatar: '',
      email: 'wangmh@police.gov.cn',
      phone: '13800001003',
      dept_id: 'dept-002',
      dept_name: '刑警支队',
      role: 'group_leader',
      permissions: [
        'create_note_self', 'edit_others_note', 'delete_note',
        'remind', 'view_group_archive',
        'access_screen',
      ],
    },
  },
  li: {
    password: '123456',
    user: {
      id: 'u-user-001',
      name: '李思远',
      avatar: '',
      email: 'lisy@police.gov.cn',
      phone: '13800001004',
      dept_id: 'dept-002',
      dept_name: '刑警支队',
      role: 'user',
      permissions: ['create_note_self'],
    },
  },
}

export function getDemoUser(username: string): (typeof DEMO_ACCOUNTS)[string] | undefined {
  return DEMO_ACCOUNTS[username]
}

// ---------------------- 演示标签 ----------------------
export const DEMO_TAGS: Tag[] = [
  { id: 'tag-001', name: '紧急',        color: '#DC2626', scope: 'system',   category: '优先级',   usage_count: 8 },
  { id: 'tag-002', name: '重要',        color: '#F97316', scope: 'system',   category: '优先级',   usage_count: 12 },
  { id: 'tag-003', name: '普通',        color: '#64748B', scope: 'system',   category: '优先级',   usage_count: 20 },
  { id: 'tag-004', name: '情报研判',    color: '#3B82F6', scope: 'system',   category: '业务类型', usage_count: 6 },
  { id: 'tag-005', name: '案件协查',    color: '#8B5CF6', scope: 'system',   category: '业务类型', usage_count: 4 },
  { id: 'tag-006', name: '会议纪要',    color: '#22C55E', scope: 'system',   category: '文档类型', usage_count: 7 },
  { id: 'tag-007', name: '工作日报',    color: '#14B8A6', scope: 'system',   category: '文档类型', usage_count: 15 },
  { id: 'tag-008', name: '盯办预警',    color: '#EF4444', scope: 'system',   category: '状态标记', usage_count: 3 },
  { id: 'tag-009', name: '待反馈',      color: '#EAB308', scope: 'system',   category: '状态标记', usage_count: 9 },
  { id: 'tag-010', name: '专项工作',    color: '#EC4899', scope: 'system',   category: '业务类型', usage_count: 5 },
  { id: 'tag-011', name: '个人备忘',    color: '#94A3B8', scope: 'personal', category: '其他',     usage_count: 3 },
  { id: 'tag-012', name: '重点关注',    color: '#78716C', scope: 'personal', category: '其他',     usage_count: 2 },
]

// ---------------------- 演示部门 ----------------------
export const DEMO_DEPARTMENTS: Department[] = [
  {
    id: 'dept-001', name: '市公安局', parent_id: null, member_count: 4,
    children: [
      { id: 'dept-002', name: '刑警支队', parent_id: 'dept-001', member_count: 12,
        children: [
          { id: 'dept-003', name: '侦查一队', parent_id: 'dept-002', member_count: 5 },
          { id: 'dept-004', name: '侦查二队', parent_id: 'dept-002', member_count: 4 },
          { id: 'dept-005', name: '技术中队', parent_id: 'dept-002', member_count: 3 },
        ],
      },
      { id: 'dept-006', name: '治安支队', parent_id: 'dept-001', member_count: 8 },
      { id: 'dept-007', name: '网安支队', parent_id: 'dept-001', member_count: 6 },
    ],
  },
]

// ---------------------- 演示人员 ----------------------
export interface DemoUserBrief {
  id: string; name: string; avatar: string; dept_id: string; dept_name: string; role: string
}

export const DEMO_USERS: DemoUserBrief[] = [
  { id: 'u-admin-001', name: '李建国', avatar: '', dept_id: 'dept-001', dept_name: '市公安局',   role: 'super_admin' },
  { id: 'u-dept-001',  name: '张振华', avatar: '', dept_id: 'dept-002', dept_name: '刑警支队',   role: 'dept_admin' },
  { id: 'u-leader-001',name: '王明辉', avatar: '', dept_id: 'dept-002', dept_name: '刑警支队',   role: 'group_leader' },
  { id: 'u-user-001',  name: '李思远', avatar: '', dept_id: 'dept-002', dept_name: '刑警支队',   role: 'user' },
  { id: 'u-user-002',  name: '赵志强', avatar: '', dept_id: 'dept-003', dept_name: '侦查一队',   role: 'user' },
  { id: 'u-user-003',  name: '陈晓东', avatar: '', dept_id: 'dept-003', dept_name: '侦查一队',   role: 'user' },
  { id: 'u-user-004',  name: '刘建华', avatar: '', dept_id: 'dept-004', dept_name: '侦查二队',   role: 'user' },
  { id: 'u-user-005',  name: '周文博', avatar: '', dept_id: 'dept-005', dept_name: '技术中队',   role: 'user' },
  { id: 'u-user-006',  name: '吴国栋', avatar: '', dept_id: 'dept-006', dept_name: '治安支队',   role: 'user' },
  { id: 'u-user-007',  name: '郑安全', avatar: '', dept_id: 'dept-007', dept_name: '网安支队',   role: 'user' },
]

// ---------------------- 便签内容模板 ----------------------
const now = new Date()
function daysAgo(n: number) {
  const d = new Date(now)
  d.setDate(d.getDate() - n)
  return d.toISOString()
}

// ---------------------- 演示便签数据 ----------------------
export const DEMO_NOTES: Note[] = [
  // ---- 待办便签（黄色）----
  {
    id: 'note-001', title: '1215抢劫案线索核查', content: '根据市局通报，12月15日凌晨发生的抢劫案嫌疑人身份已初步确认。\n\n需要核查的线索：\n1. 嫌疑人A - 张某某，户籍地：XX区XX街道\n2. 嫌疑人A的手机基站数据已调取，案发时间段出现在案发地周边\n3. 请侦查一队前往嫌疑人住处蹲守，技术中队配合做DNA比对\n\n截止时间：本周五前完成初步核查报告。',
    status: 'active', source_type: 'assigned', priority: 'normal',
    owner_id: 'u-leader-001', creator_id: 'u-dept-001',
    tags: [DEMO_TAGS[0], DEMO_TAGS[3], DEMO_TAGS[8]],
    assignees: [{ id: 'u-user-002', name: '赵志强', avatar: '', dept_name: '侦查一队', role: 'user' }],
    created_at: daysAgo(1), updated_at: daysAgo(0),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },
  {
    id: 'note-002', title: '跨区域电信诈骗案协同研判', content: '省厅来文：近期省内有组织系列电信诈骗案涉及我市多名受害者。\n\n要求：\n1. 梳理本市受害人名单及被骗金额\n2. 联合网安支队追踪涉案IP地址\n3. 配合银行调取涉案账户流水\n4. 整理研判报告，报送省厅专案组\n\n协办单位：网安支队、经侦支队',
    status: 'active', source_type: 'assigned', priority: 'normal',
    owner_id: 'u-dept-001', creator_id: 'u-admin-001',
    tags: [DEMO_TAGS[0], DEMO_TAGS[3], DEMO_TAGS[5], DEMO_TAGS[10]],
    assignees: [
      { id: 'u-dept-001', name: '张振华', avatar: '', dept_name: '刑警支队', role: 'dept_admin' },
      { id: 'u-user-007', name: '郑安全', avatar: '', dept_name: '网安支队', role: 'user' },
    ],
    created_at: daysAgo(2), updated_at: daysAgo(1),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },
  {
    id: 'note-003', title: '今日工作安排', content: '- 上午9:00 支队周例会（会议室302）\n- 10:30 听取侦查一队1215案件进展汇报\n- 14:00 去市局送案件材料\n- 16:00 整理上周案件台账',
    status: 'active', source_type: 'self', priority: 'normal',
    owner_id: 'u-dept-001', creator_id: 'u-dept-001',
    tags: [DEMO_TAGS[6]],
    assignees: [],
    created_at: daysAgo(0), updated_at: daysAgo(0),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },
  {
    id: 'note-004', title: '走访调查安排', content: '针对"平安社区"专项行动，本周需要完成以下走访：\n\n1. 幸福小区 - 核查出租屋信息\n2. 阳光花园 - 走访重点人员3户\n3. 翠苑新村 - 治安隐患排查\n\n每走访一个小区需填写《走访登记表》，拍照留档。',
    status: 'active', source_type: 'self', priority: 'normal',
    owner_id: 'u-leader-001', creator_id: 'u-leader-001',
    tags: [DEMO_TAGS[2], DEMO_TAGS[10]],
    assignees: [
      { id: 'u-user-003', name: '陈晓东', avatar: '', dept_name: '侦查一队', role: 'user' },
      { id: 'u-user-004', name: '刘建华', avatar: '', dept_name: '侦查二队', role: 'user' },
    ],
    created_at: daysAgo(3), updated_at: daysAgo(2),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },
  {
    id: 'note-005', title: '下周一专案会材料准备', content: '专案汇报需要准备的资料清单：\n\n□ 案件基本情况综述\n□ 侦查过程时间轴\n□ 证据链梳理\n□ 嫌疑人社会关系图谱\n□ 下一步工作计划\n□ PPT汇报材料\n\n提醒：周日前完成初稿，发给支队长审核。',
    status: 'active', source_type: 'self', priority: 'normal',
    owner_id: 'u-user-001', creator_id: 'u-user-001',
    tags: [DEMO_TAGS[1], DEMO_TAGS[5], DEMO_TAGS[10]],
    assignees: [],
    due_time: new Date(now.getFullYear(), now.getMonth(), now.getDate() + 3).toISOString(),
    created_at: daysAgo(4), updated_at: daysAgo(3),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },

  // ---- 盯办便签（红色）----
  {
    id: 'note-006', title: '⚠ 在逃人员轨迹追踪（上级督办）', content: '【省厅督办件】\n\n在逃人员：王某，男，35岁，涉嫌故意伤害罪\n最后出现地点：我市XX区\n\n要求24小时内反馈追踪进展：\n- 已调取监控，锁定最后出现区域\n- 已布控火车站、长途汽车站\n- 正在追查其社会关系人\n- 请求技侦协助手机定位',
    status: 'active', source_type: 'assigned', priority: 'urgent',
    owner_id: 'u-leader-001', creator_id: 'u-admin-001',
    tags: [DEMO_TAGS[0], DEMO_TAGS[7]],
    assignees: [
      { id: 'u-leader-001', name: '王明辉', avatar: '', dept_name: '刑警支队', role: 'group_leader' },
      { id: 'u-user-002', name: '赵志强', avatar: '', dept_name: '侦查一队', role: 'user' },
    ],
    created_at: daysAgo(1), updated_at: daysAgo(0),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },
  {
    id: 'note-007', title: '⚠ 网络谣言扩散舆情监控', content: '网安支队通报：关于XX事件的网络谣言在本地微信群扩散。\n\n当前状态：\n- 监测到32个微信群转发\n- 已锁定谣言源头账号3个\n- 舆情呈上升趋势\n\n需要：立即启动舆情应对预案，协调宣传部门发布辟谣信息。',
    status: 'active', source_type: 'assigned', priority: 'urgent',
    owner_id: 'u-dept-001', creator_id: 'u-user-007',
    tags: [DEMO_TAGS[0], DEMO_TAGS[7], DEMO_TAGS[4]],
    assignees: [
      { id: 'u-dept-001', name: '张振华', avatar: '', dept_name: '刑警支队', role: 'dept_admin' },
      { id: 'u-user-007', name: '郑安全', avatar: '', dept_name: '网安支队', role: 'user' },
    ],
    created_at: daysAgo(0), updated_at: daysAgo(0),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  },

  // ---- 已完成便签（绿色）----
  {
    id: 'note-008', title: '1208盗窃案件侦查终结', content: '案件已侦查终结，嫌疑人供认不讳。\n\n处理结果：\n1. 追回被盗财物价值约15万元\n2. 嫌疑人移交检察院\n3. 案件卷宗已整理归档\n\n反馈意见：侦查过程中监控调取及时，走访工作到位。',
    status: 'completed', source_type: 'assigned', priority: 'normal',
    owner_id: 'u-user-002', creator_id: 'u-leader-001',
    tags: [DEMO_TAGS[3], DEMO_TAGS[5]],
    assignees: [{ id: 'u-user-002', name: '赵志强', avatar: '', dept_name: '侦查一队', role: 'user' }],
    created_at: daysAgo(10), updated_at: daysAgo(2),
    completed_at: daysAgo(2),
    allowed_actions: ['view'],
  },
  {
    id: 'note-009', title: '11月工作台账汇总', content: '11月支队工作台账已完成汇总：\n\n- 接处警：127起\n- 刑事案件：18起（破获14起）\n- 治安案件：36起（查处32起）\n- 抓获嫌疑人：22人\n- 刑事拘留：9人\n- 移送起诉：7人\n\n台账已上报市局。',
    status: 'completed', source_type: 'self', priority: 'normal',
    owner_id: 'u-dept-001', creator_id: 'u-dept-001',
    tags: [DEMO_TAGS[5], DEMO_TAGS[6]],
    assignees: [],
    created_at: daysAgo(28), updated_at: daysAgo(25),
    completed_at: daysAgo(25),
    allowed_actions: ['view'],
  },

  // ---- 已归档便签 ----
  {
    id: 'note-010', title: '1005绑架案侦破总结', content: '案件侦破全过程总结，已归档。\n\n关键节点：\n10/05 22:00 - 接报案\n10/06 01:00 - 锁定嫌疑人\n10/06 06:00 - 成功解救人质\n10/06 08:00 - 嫌疑人落网\n\n经验总结：多警种协作高效，技术手段支撑有力。',
    status: 'archived', source_type: 'assigned', priority: 'normal',
    owner_id: 'u-leader-001', creator_id: 'u-admin-001',
    tags: [DEMO_TAGS[3], DEMO_TAGS[4], DEMO_TAGS[5]],
    assignees: [],
    created_at: daysAgo(60), updated_at: daysAgo(30),
    completed_at: daysAgo(55), archived_at: daysAgo(30),
    allowed_actions: ['view', 'export'],
  },
  {
    id: 'note-011', title: '第三季度绩效考核材料', content: '已归档的绩效材料：\n\n1. 各中队考核评分表\n2. 个人绩效总结\n3. 领导评语\n4. 加分项明细\n\n所有材料已签字确认，存入档案。',
    status: 'archived', source_type: 'self', priority: 'normal',
    owner_id: 'u-dept-001', creator_id: 'u-dept-001',
    tags: [DEMO_TAGS[2], DEMO_TAGS[6]],
    assignees: [],
    created_at: daysAgo(45), updated_at: daysAgo(40),
    completed_at: daysAgo(40), archived_at: daysAgo(35),
    allowed_actions: ['view', 'export'],
  },
  {
    id: 'note-012', title: '治安巡逻部署方案', content: '国庆期间巡逻方案（已归档）：\n\n重点区域：火车站、商业街、景区\n巡逻力量：每日3组，每组4人\n巡逻时间：08:00-22:00\n应急备勤：24小时1组待命',
    status: 'archived', source_type: 'self', priority: 'normal',
    owner_id: 'u-dept-001', creator_id: 'u-dept-001',
    tags: [DEMO_TAGS[2]],
    assignees: [],
    created_at: daysAgo(90), updated_at: daysAgo(70),
    completed_at: daysAgo(80), archived_at: daysAgo(70),
    allowed_actions: ['view', 'export'],
  },
]

// ---------------------- 演示数据辅助函数 ----------------------
export function filterDemoNotes(
  status?: string,
  keyword?: string,
  ownerId?: string
): Note[] {
  let notes = [...DEMO_NOTES]

  if (status === 'active') {
    notes = notes.filter(n => n.status === 'active')
  } else if (status === 'completed') {
    notes = notes.filter(n => n.status === 'completed')
  } else if (status === 'archived') {
    notes = notes.filter(n => n.status === 'archived')
  }

  if (keyword) {
    const kw = keyword.toLowerCase()
    notes = notes.filter(
      n => n.title.toLowerCase().includes(kw) || n.content.toLowerCase().includes(kw)
    )
  }

  if (ownerId) {
    notes = notes.filter(n => n.owner_id === ownerId || n.assignees.some(a => a.id === ownerId))
  }

  return notes
}

let demoIdCounter = Date.now()
export function createDemoNote(payload: { title: string; content: string; tags: string[]; source_type: string; owner_id?: string }): Note {
  return {
    id: `note-demo-${++demoIdCounter}`,
    title: payload.title,
    content: payload.content || '',
    status: 'active',
    source_type: (payload.source_type as Note['source_type']) || 'self',
    priority: 'normal',
    owner_id: payload.owner_id || 'u-user-001',
    creator_id: payload.owner_id || 'u-user-001',
    tags: DEMO_TAGS.filter(t => payload.tags.includes(t.id)),
    assignees: [],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    allowed_actions: ['edit', 'delete', 'complete', 'remind'],
  }
}

// 当前演示模式（localStorage 持久化）
export function isDemoMode(): boolean {
  return localStorage.getItem('demo_mode') === 'true'
}

export function setDemoMode(on: boolean) {
  if (on) {
    localStorage.setItem('demo_mode', 'true')
  } else {
    localStorage.removeItem('demo_mode')
  }
}
