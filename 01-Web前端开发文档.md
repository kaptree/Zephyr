# 轻燕工作台 —— Web前端开发文档

> 版本：v2.0  
> 定位：面向公司系统的轻量化智能协同办公支撑解决方案 · Web管理端  
> 核心载体：桌面智能便签协同办公系统的Web管理端、专项工作组管理、数据大屏、工作成效分析

---

## 一、项目概述

Web前端作为系统的管理中枢与信息汇聚端，承载便签云端管理、专项工作组创建与任务分发、组织人员库维护、标签模板配置、数据大屏实时展示、智能报告生成等功能。页面以纯白为底、极简为形，使用 `DaisyUI + Tailwind CSS` 构建高质感现代界面。

---

## 二、技术栈

| 层级        | 技术选型                     | 说明                                 |
| ----------- | ---------------------------- | ------------------------------------ |
| 框架        | Vue 3.4+ (Composition API)   | `<script setup lang="ts">` 标准写法  |
| 语言        | TypeScript                   | 全量类型定义                         |
| UI          | DaisyUI 4 + Tailwind CSS 3.4 | 原子化样式 + 组件库                  |
| 状态管理    | Pinia                        | 便签状态、用户权限、协同数据         |
| 路由        | Vue Router 4                 | 懒加载、路由守卫鉴权                 |
| HTTP 客户端 | Axios                        | 统一拦截器、自动注入 Bearer Token    |
| 实时通信    | Socket.io-client             | Web端协同编辑（可选）                |
| 构建工具    | Vite 8                       | 极速 HMR                             |
| 图表        | 内联 SVG                     | 数据大屏趋势图（三次贝塞尔平滑曲线） |
| 测试        | Vitest + Vue Test Utils      | 单元测试                             |

---

## 三、项目目录结构

```
Web-Front/src/
├── App.vue                    # 根组件
├── main.ts                    # 入口：注册 Pinia / Router / 全局错误处理
├── style.css                  # Tailwind 指令 + 全局样式
├── composables/               # 组合式 API（暗色模式、离线检测、WebSocket、Toast 等）
├── directives/
│   └── permission.ts          # v-permission 权限指令
├── pages/
│   ├── LoginPage.vue          # 登录页
│   ├── WorkbenchPage.vue      # ★ 核心工作台（便签墙 + 专项工作组列表 + 一键创建）
│   ├── ArchivePage.vue        # 归档查询
│   ├── CollaborationPage.vue  # 协同编辑室
│   ├── AnalyticsPage.vue      # 工作成效分析（含智能报告+报告模板编辑）
│   ├── ScreenPage.vue         # 数据大屏（平滑趋势图 + 实时动态）
│   ├── GroupDetailPage.vue    # ★ 专项行动详情（成员分组 + 专属便签卡片）
│   ├── ProfilePage.vue        # 个人中心
│   ├── ForbiddenPage.vue      # 403
│   ├── NotFoundPage.vue       # 404
│   └── admin/
│       ├── DepartmentsPage.vue # 部门库管理
│       ├── UsersPage.vue       # 人员库管理
│       ├── TagsPage.vue        # 标签库管理
│       ├── TemplatesPage.vue   # 模板库管理
│       ├── SystemSettingsPage.vue  # 系统设置
│       └── OperationLogPage.vue    # 操作日志
├── components/
│   ├── common/
│   │   ├── TagSelector.vue     # 标签选择器（搜索+创建+多选）
│   │   └── UserPicker.vue      # 人员选择器（部门树+搜索+多选）
│   └── note/
│       └── StickyNoteCard.vue  # 便签卡片（长方体、三色状态、hover动效）
├── router/
│   └── index.ts               # 路由表 + 权限守卫
├── services/
│   ├── api.ts                  # Axios 实例（BaseURL/拦截器）
│   ├── admin.ts                # 认证 + 部门 + 人员 API
│   ├── analytics.ts            # 分析统计 + 报告模板 API
│   ├── groupNotes.ts           # 专项工作组专属便签 API
│   ├── notes.ts                # 便签 CRUD + 盯办/归档 API
│   ├── system.ts               # 系统配置 API
│   ├── tags.ts                 # 标签管理 API
│   └── workgroup.ts            # 专项工作组 CRUD + 搜索 API
├── stores/
│   ├── auth.ts                 # 用户状态、Token、权限
│   ├── collaboration.ts        # 协同状态
│   └── notes.ts                # 便签状态
├── types/
│   ├── api.ts                  # ApiResponse / PaginatedData 泛型
│   ├── collaboration.ts        # 协同类型
│   ├── index.ts                # 统一导出
│   ├── note.ts                 # Note / CreateNotePayload 等
│   ├── system.ts               # 系统配置类型
│   ├── tag.ts                  # Tag 类型
│   ├── user.ts                 # User / UserBrief / Department / TreeNode
│   └── workbench.ts            # Template / Group / SerialNumber
└── utils/
    └── errorHandler.ts         # 全局错误处理
```

---

## 四、页面结构与路由

### 4.1 路由表

| 路由                           | 名称            | 组件               | 权限               | 说明                        |
| ------------------------------ | --------------- | ------------------ | ------------------ | --------------------------- |
| `/login`                       | Login           | LoginPage          | 公开               | 登录页                      |
| `/workbench`                   | Workbench       | WorkbenchPage      | 已登录             | ★ 核心页：便签墙 + 专项行动 |
| `/workbench/archive`           | Archive         | ArchivePage        | 已登录             | 归档查询                    |
| `/workbench/collaboration/:id` | Collaboration   | CollaborationPage  | 已登录             | 协同编辑室                  |
| `/workbench/groups/:id`        | WorkGroupDetail | GroupDetailPage    | 已登录             | ★ 专项行动详情              |
| `/analytics`                   | Analytics       | AnalyticsPage      | 已登录             | 工作成效分析                |
| `/screen/:id`                  | Screen          | ScreenPage         | 已登录             | 数据大屏                    |
| `/admin/departments`           | Departments     | DepartmentsPage    | manage_departments | 部门管理                    |
| `/admin/users`                 | Users           | UsersPage          | manage_users       | 人员管理                    |
| `/admin/tags`                  | Tags            | TagsPage           | manage_tags        | 标签管理                    |
| `/admin/templates`             | Templates       | TemplatesPage      | manage_templates   | 模板管理                    |
| `/admin/system`                | SystemSettings  | SystemSettingsPage | manage_system      | 系统设置                    |
| `/admin/operation-log`         | OperationLog    | OperationLogPage   | manage_system      | 操作日志                    |
| `/profile`                     | Profile         | ProfilePage        | 已登录             | 个人中心                    |
| `/403`                         | Forbidden       | ForbiddenPage      | -                  | 无权访问                    |
| `/:pathMatch(.*)*`             | NotFound        | NotFoundPage       | -                  | 404                         |

### 4.2 路由守卫

- 检查 `localStorage` 中 `auth_token` 是否存在
- `super_admin` 角色跳过权限检查
- 其他角色需检查路由 `meta.permissions` 与用户权限交集

---

## 五、核心页面设计

### 5.1 WorkbenchPage — 工作台（核心入口）

**Tab 栏**：`全部 | 待办 | 指派 | 盯办 | 已完成 | 🏢专项工作组`

**专项工作组 Tab**：

- 搜索栏：关键词（ILIKE）、日期范围（date_from / date_to）组合筛选，Enter 触发搜索
- 专项行动列表：卡片式，展示标题/状态/模板类型/发起人/成员/截止时间，点击进入详情页
- 「一键创建」按钮：弹出模态框 → 填写名称/模板/描述/截止日期 → 多小组设置（每小组独立 UserPicker）→ 一键创建工作组并自动为每位成员创建任务便签

**便签 Tab（全部/待办/指派/盯办/已完成）**：

- 便签墙：`grid-cols-[repeat(auto-fill,minmax(280px,1fr))]` 自适应网格
- 点击便签 → 右侧滑出详情面板（编辑标题/内容/查看标签/完成归档/盯办）
- 悬浮 FAB 按钮 → 弹窗创建便签（标题/内容/标签/类型/指派人员）

### 5.2 GroupDetailPage — 专项行动详情

- 顶部信息栏：状态/模板/发起人/成员数/截止时间 → 成员按 `sub_group_name` 分组卡片展示
- **专属便签区**：仅展示 `group_id = 当前工作组` 的便签，以 `StickyNoteCard` 卡片形式呈现在自适应网格中
- 新建便签弹窗：支持选择负责人（从工作组成员中选）、截止日期、标签（TagSelector）
- 点击便签卡片 → 右侧滑出详情面板（编辑/归档/盯办）
- 所有表单使用 `@keydown.enter.prevent` 防止误提交

### 5.3 ScreenPage — 数据大屏

- 顶部时钟 + 日期 + "轻燕工作台 · 数据大屏" + 实时标识 + 「← 返回工作台」按钮
- 四指标卡片：便签总数 / 活跃便签 / 归档便签 / 人员总数
- 趋势图：内联 SVG，三次贝塞尔平滑曲线 + 渐变填充区域 + 数据点 + 坐标轴刻度
- 时间切换：近一周 / 近一月 / 近一季
- 最新动态：垂直时间轴，按操作类型着色

### 5.4 AnalyticsPage — 工作成效分析

- 数据统计 Tab：个人统计面板（创建/完成/完成率/盯办/趋势/标签分布）
- 报告历史 Tab：AI 生成的报告列表
- 「📝 编辑模板」按钮：Markdown 模板编辑器，支持变量占位符（如 `{{userName}}`、`{{completionRate}}`），变量高亮显示

---

## 六、核心组件设计

### 6.1 StickyNoteCard

- **Props**：`note: Note`、`mode: 'desktop' | 'web'`、`archived: boolean`
- **Events**：`click`、`complete`、`remind`、`restore`
- **状态视觉**：
  - 黄色（默认待办）：`#FEF3C7` 底色
  - 红色（盯办预警）：`#FEE2E2` 底色 + 脉冲动画
  - 绿色（已完成）：`#DCFCE7` 底色
- **内容展开**：超过 100 字显示"展开全文"
- **标签展示**：最多 2 个胶囊，超出显示 `+N`；`tags` 为 `undefined` 时安全降级

### 6.2 TagSelector

- 下拉浮层，支持多选/搜索/Enter 创建新标签
- 已选标签以彩色胶囊展示
- 所有 `<input>` 禁止 Enter 冒泡触发外层表单提交（`@keydown.enter.prevent`）

### 6.3 UserPicker

- 部门树浏览 + 搜索双模式
- 多选/单选，已选人员头像+姓名胶囊
- 搜索框加 `@keydown.enter.prevent` 防止触发外层表单提交

---

## 七、状态管理（Pinia Stores）

### 7.1 `useNoteStore`

- **State**：`activeNotes`、`selectedNote`、`loading`
- **Actions**：`fetchNotes`、`createNote`、`completeNote`、`remindNote`、`updateNoteLocally`

### 7.2 `useAuthStore`

- **State**：`user`、`token`
- 登录后存储 token 到 localStorage，自动注入请求头

---

## 八、服务层（Services）

| 服务文件        | 主要函数                                                                                                                     | 说明                   |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| `api.ts`        | `get / post / put / del`                                                                                                     | Axios 封装，统一拦截器 |
| `admin.ts`      | `login / getDepartments / getUsers / createUser / updateUser / deleteUser / getVisibleUsers`                                 | 认证+组织+人员         |
| `notes.ts`      | `fetchNotes / createNote / updateNote / completeNote / remindNote / deleteNote / restoreNote / fetchNoteStats / exportNotes` | 便签 CRUD + 统计       |
| `tags.ts`       | `fetchTags / createTag / updateTag / deleteTag`                                                                              | 标签管理               |
| `workgroup.ts`  | `searchGroups / getMyGroups / getWorkGroupDetail / createWorkGroup / deleteWorkGroup / getWorkGroupMembers`                  | 工作组 CRUD + 搜索     |
| `groupNotes.ts` | `getGroupNotes / createGroupNote`                                                                                            | 专属便签               |
| `analytics.ts`  | `fetchPersonalStats / generateAIReport / fetchReports / fetchReportTemplate / saveReportTemplate`                            | 分析+报告模板          |
| `system.ts`     | `fetchConfig / updateConfig / fetchOperations`                                                                               | 系统配置+日志          |

---

## 九、设计规范

### 9.1 色彩体系

- **主背景**：`#FFFFFF` / `#F8FAFC`
- **主文字**：`#0F172A`（标题）/ `#475569`（正文）
- **交互蓝**：`#3B82F6`
- **专项紫蓝渐变**：`from-purple-500 to-blue-500`
- **便签三色**：黄 `#FEF3C7` / 绿 `#DCFCE7` / 红 `#FEE2E2`

### 9.2 圆角与阴影

- 卡片：`rounded-xl`（12px），hover 时 `shadow-md`
- 按钮：`rounded-btn`（10px）
- 模态框：`rounded-card`（16px）+ `shadow-modal`

### 9.3 动效

- 全局过渡：`transition-smooth`（all 0.3s cubic-bezier）
- 便签插入：`animate-spring-enter`
- 盯办脉冲：2s infinite pulse

---

_文档结束_
