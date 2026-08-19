# 轻燕工作台 —— Web前端开发文档

> 版本：v2.0  
> 定位：面向公司系统的轻量化智能协同办公支撑解决方案 · Web管理端  
> 核心载体：桌面智能任务协同办公系统的Web管理端、专项工作组管理、数据大屏、工作成效分析

---

## 一、项目概述

Web前端作为系统的管理中枢与信息汇聚端，承载任务云端管理、专项工作组创建与任务分发、组织人员库维护（含人员岗位/技能标签/能力图谱）、标签模板配置、数据大屏实时展示、智能报告生成等功能。页面以纯白为底、极简为形，使用 `DaisyUI + Tailwind CSS` 构建高质感现代界面。

---

## 二、技术栈

| 层级        | 技术选型                     | 说明                                 |
| ----------- | ---------------------------- | ------------------------------------ |
| 框架        | Vue 3.4+ (Composition API)   | `<script setup lang="ts">` 标准写法  |
| 语言        | TypeScript                   | 全量类型定义                         |
| UI          | DaisyUI 4 + Tailwind CSS 3.4 | 原子化样式 + 组件库                  |
| 状态管理    | Pinia                        | 任务状态、用户权限、协同数据         |
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
│   ├── WorkbenchPage.vue      # ★ 核心工作台（任务墙 + 专项工作组列表 + 一键创建）
│   ├── ArchivePage.vue        # 归档查询
│   ├── CollaborationPage.vue  # 协同编辑室
│   ├── AnalyticsPage.vue      # 工作成效分析（含智能报告+报告模板编辑）
│   ├── ScreenPage.vue         # 数据大屏（平滑趋势图 + 实时动态）
│   ├── GroupDetailPage.vue    # ★ 专项行动详情（成员分组 + 专属任务卡片）
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
│       └── StickyNoteCard.vue  # 任务卡片（长方体、三色状态、hover动效）
├── router/
│   └── index.ts               # 路由表 + 权限守卫
├── services/
│   ├── api.ts                  # Axios 实例（BaseURL/拦截器）
│   ├── admin.ts                # 认证 + 部门 + 人员 API
│   ├── analytics.ts            # 分析统计 + 报告模板 API
│   ├── groupNotes.ts           # 专项工作组专属任务 API
│   ├── notes.ts                # 任务 CRUD + 盯办/归档 API
│   ├── system.ts               # 系统配置 API
│   ├── tags.ts                 # 标签管理 API
│   └── workgroup.ts            # 专项工作组 CRUD + 搜索 API
├── stores/
│   ├── auth.ts                 # 用户状态、Token、权限
│   ├── collaboration.ts        # 协同状态
│   └── notes.ts                # 任务状态
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
| `/workbench`                   | Workbench       | WorkbenchPage      | 已登录             | ★ 核心页：任务墙 + 专项行动 |
| `/workbench/archive`           | Archive         | ArchivePage        | 已登录             | 归档查询                    |
| `/workbench/inspect`           | InspectWorkbench | InspectWorkbenchPage | inspect_user_workbench | ★ 用户工作台（公司领导/超管） |
| `/workbench/collaboration/:id` | Collaboration   | CollaborationPage  | 已登录             | 协同编辑室                  |
| `/workbench/groups/:id`        | WorkGroupDetail | GroupDetailPage    | 已登录             | ★ 专项行动详情              |
| `/analytics`                   | Analytics       | AnalyticsPage      | 已登录             | 工作成效分析                |
| `/issues`                      | Issues          | IssuesPage         | 已登录             | ★ Bug 反馈列表（GitHub Issues 风格） |
| `/issues/:id`                  | IssueDetail     | IssueDetailPage    | 已登录             | ★ Bug 反馈详情 + 评论区     |
| `/chat`                        | Chat            | ChatPage           | 已登录             | ★ 聊天（仿 QQ/微信：会话/通讯录 + emoji + 文件） |
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

### 4.3 角色与菜单权限（★ 公司领导）

| 角色 | 说明 | 数据范围 | 系统管理/操作日志 | 用户工作台 |
| ---- | ---- | -------- | ----------------- | ---------- |
| `super_admin`    | 系统管理员 | 全局 | ✅ | ✅ |
| `company_leader` | 公司领导（★ 新增） | 全局（权限大于部门管理员） | ❌ | ✅ |
| `dept_admin`     | 部门管理员 | 本部门（含子部门） | ❌ | ❌ |
| `group_leader`   | 组长 | 组内/本人相关 | ❌ | ❌ |
| `user`           | 普通员工 | 本人相关 | ❌ | ❌ |
| `screen_role`    | 大屏角色 | 本人相关 | ❌ | ❌ |

- 侧边栏菜单按 `permission` 过滤：公司领导可见「用户工作台」；「系统管理/操作日志」菜单为 `manage_system` 权限（仅 super_admin），公司领导天然不可见
- 公司领导拥有任务全操作、部门/用户/标签/模板增改、数据大屏与指令权限，但**无删除权限**（删除用户/部门/模板仍为 super_admin 独占）

---

## 五、核心页面设计

### 5.1 WorkbenchPage — 工作台（核心入口）

**Tab 栏**：`全部 | 待办 | 指派 | 盯办 | 已完成 | 🏢专项工作组`

**专项工作组 Tab**：

- 搜索栏：关键词（ILIKE）、日期范围（date_from / date_to）组合筛选，Enter 触发搜索
- 专项行动列表：卡片式，展示标题/状态/模板类型/发起人/成员/截止时间，点击进入详情页
- 「一键创建」按钮：弹出模态框 → 填写名称/模板/描述/截止日期 → 多小组设置（每小组独立 UserPicker）→ 一键创建工作组并自动为每位成员创建任务任务

**任务 Tab（全部/待办/指派/盯办/已完成）**：

- 任务墙：`grid-cols-[repeat(auto-fill,minmax(280px,1fr))]` 自适应网格
- 点击任务 → 居中模态框编辑详情（标题/富文本内容/标签/完成归档/重要/删除）
- 任务卡片操作栏：`完成并归档` + `重要`（非红色卡片，点击变红）+ `删除`（确认后从工作台移除）（★ 新增，已移除「盯办」按钮）
- 悬浮 FAB 按钮 → 居中弹窗创建任务（标题/富文本内容/标签/类型/指派人员）
  - 「指派他人」支持**多人勾选**（最多 50 人）：部门树每行带「全选/全不选」按钮，一键选中该部门（含子部门）全部人员；被指派的所有人员都会收到盯办提醒
  - 「指派他人」模式下支持选择工作类型，「一键推荐」基于历史参与数据智能推荐合适人员（可多选）
  - 「指派他人/协作」模式下新增 **「⏱ 工作时间」选项**（★ 新增）：不限制 / 1小时 / 2小时 / 4小时 / 8小时 / 1天 / 2天 / 3天 / 7天，提交时携带 `work_time_seconds`，后端自动计算截止时间；任务下发后即开始倒计时
- **局部刷新**（★ 新增）：指派任务创建成功、任务完成归档（反馈提交）后自动 `fetchNotes()` 刷新任务面板，最新任务/状态/倒计时立即展示
- **富文本编辑**：创建/编辑任务内容使用 `RichTextEditor`（contenteditable），支持回车换行、加粗/斜体/下划线/删除线、标题、有序/无序列表、引用、代码块、撤销/重做、清除格式；旧纯文本内容自动兼容渲染（换行转 `<br>`）
- **辅助模块**：为不同的业务场景提供了各自独立的管理页面：
  - **[WorkbenchPage]** 便签 + 任务云端管理，团队任务统筹分发
  - **[GroupDetailPage]** 专项工作组详情页，支持新建任务时从预设组快速加载成员
  - **[UsersPage]** 人员库管理 → 含人员标签（岗位/技能）和「组预设」管理，可创建命名预设组供新建工作组时选用

### 5.2 GroupDetailPage — 专项行动详情

- 顶部信息栏：状态/模板/发起人/成员数/截止时间 → 成员按 `sub_group_name` 分组卡片展示
- **专属任务区**：仅展示 `group_id = 当前工作组` 的任务，以 `StickyNoteCard` 卡片形式呈现在自适应网格中
- 新建任务弹窗：支持选择负责人（从工作组成员中选）、截止日期、标签（TagSelector）
- 点击任务卡片 → 右侧滑出详情面板（编辑/归档/盯办）
- 所有表单使用 `@keydown.enter.prevent` 防止误提交

### 5.3 UsersPage — 人员库管理（增强）

- **人员列表**：展示姓名/用户名/部门/岗位标签/技能标签/角色/工作类型参与统计/状态
- **人员档案弹窗**：点击「档案」查看完整人员档案，包括岗位、技能特长、参与过的工作类型及次数统计
- **编辑表单**：新增「岗位」和「技能特长」字段，支持通过逗号分隔输入多项技能
- **智能推荐集成**：人员标签数据为「指派他人」中的「一键推荐」提供数据支撑
- **组预设管理**：
  - 底部「组预设管理」区域展示所有预设组卡片（名称/成员/适用类型）
  - 「新建预设组」弹窗：输入名称、描述、选择适用工作类型、勾选成员列表
  - 预设组卡片可一键删除
  - 创建的预设组可在 WorkbenchPage 新建专项工作组时直接选用，自动填充成员列表

### 5.4 ScreenPage — 数据大屏

- 顶部时钟 + 日期 + "轻燕工作台 · 数据大屏" + 实时标识 + 「← 返回工作台」按钮
- 四指标卡片：任务总数 / 活跃任务 / 归档任务 / 人员总数
- 趋势图：内联 SVG，三次贝塞尔平滑曲线 + 渐变填充区域 + 数据点 + 坐标轴刻度
- 时间切换：近一周 / 近一月 / 近一季
- 最新动态：垂直时间轴，按操作类型着色

### 5.5 AnalyticsPage — 工作成效分析

- 数据统计 Tab：个人统计面板（创建/完成/完成率/盯办/趋势/标签分布）
- **团队报告 Tab（★ 新增）**：
  - 时间范围选择：本周 / 本月 / 自定义（起止日期输入 + 查询）
  - 汇总卡片：团队成员数 / 创建任务 / 完成任务 / 整体完成率
  - 成员成效明细表：成员 / 部门 / 创建任务 / 完成任务 / 完成率（颜色分级徽标）/ 平均完成耗时 / 被盯办次数
  - 「🤖 生成模型」下拉（★ 新增）：读取启用的 AI 配置（`GET /analytics/ai-configs`，无需超管权限）；未选模型 → 模板生成，选择模型 → 调用所选模型智能生成（失败自动回退模板）
  - **成员勾选组建团队（★ 新增）**：成员明细表每行复选框，可连续勾选多人（表格始终显示全部成员供勾选、勾选行高亮），支持「全选 / 清空（全部成员）」；汇总卡片按勾选成员实时本地过滤；生成周报/月报携带 `user_ids`，报告只统计勾选成员，归属当前用户
  - **成员明细默认收起（★ 新增）**：「👥 成员成效明细」区域默认折叠，仅显示标题栏（含已选摘要与展开指示），点击展开/收起成员表格，保持页面简洁
  - **报告历史类别 tag（★ 新增）**：报告卡片以 tag 区分 个人（蓝）/ 团队（紫）/ 工作组（青），与 AI/模板 tag 并列；详情弹窗同步显示「个人报告 / 团队报告 / 工作组报告」
  - 生成按钮：📄 生成周报/月报，报告支持复制/下载，自动保存至报告历史
- 报告历史 Tab：AI 生成的报告列表
- 「📝 编辑模板」按钮：Markdown 模板编辑器，支持变量占位符（如 `{{userName}}`、`{{completionRate}}`），变量高亮显示

### 5.6 TemplatesPage — 模板库管理（★ 用户自定义模板）

- **网格展示**：每个模板卡片展示名称、类型徽章（颜色编码）、字段数量、系统/用户标识
- **「+ 添加模版」按钮**：右上角主操作入口，打开创建/编辑模态框
- **模态表单**：
  - 模板名称（必填）
  - 模板类型下拉选择：通用任务 / 数据分析 / 专项行动 / 紧急协查 / 协同作战 / 自定义
  - 布局样式下拉选择：单栏 / 双栏 / 四宫格 / 六宫格
  - 字段定义 JSON 编辑区（textarea，带示例提示）
- **编辑功能**：所有模板均可编辑，点击「编辑」按钮打开相同模态框回填数据
- **删除保护**：系统内置模板（`is_system: true`）不显示删除按钮；用户自定义模板点击「删除」弹出二次确认
- **任务创建联动**：WorkbenchPage 创建任务时，可选择已有模板自动预填字段占位符到内容区

### 5.7 SystemSettingsPage — 系统设置（含 AI 服务配置，★ 新增 dify）

- **Tab 结构**：系统设置 / AI服务配置 / **聊天文件** / 配置文件管理 / 操作日志
- **聊天文件 Tab**（★ 新增）：聊天文件传输白名单/黑名单管理
  - 白名单/黑名单均为 textarea（逗号分隔扩展名），含示例 placeholder 与规则说明（黑名单优先；白名单非空时仅允许白名单；白名单为空时除黑名单外全部放行）
  - 「保存策略」调用 `PUT /system/chat-file-policy`，成功提示「保存成功，策略已即时生效（热加载）」并刷新显示最近更新时间；「重置」重新加载数据库当前已保存值
  - 仅 super_admin 可见（系统管理菜单本身即超管权限）
- **AI 服务配置**：
  - 服务商下拉：OpenAI / DeepSeek / 通义千问 / 智谱AI / **Dify** / 自定义（内置各服务商地址占位与名称自动填充）
  - 新增/编辑表单：服务商、名称、API 地址、API 密钥（已配置密钥脱敏展示）、模型名称、描述、启用状态
  - **「⚡ 一键测试连通」按钮**（★ 新增）：调用 `POST /system/ai-configs/test` 实时探测，成功显示绿色徽标「连通成功（xx ms）」，失败显示红色徽标与错误信息
  - **保存门禁**（★ 新增）：未通过连通性测试时保存按钮禁用并提示「请先测试连通」；后端同样强制校验（失败 400 拒绝保存）
- **操作日志 Tab**（★ 增强）：
  - 表格列：操作时间 / 用户 / 角色 / 操作类型 / 方法 / 状态 / **详情** / IP地址
  - 操作类型中文映射全覆盖（创建/编辑/删除任务、反馈、附件、标签、部门、用户、模板、工作组、预设组、报告生成、通知、聊天、AI 配置等，含历史数据 create/update/delete 兜底翻译）
  - **「详情」列**：展示后端生成的「动作描述 · 关键字段摘要」（如「编辑任务 · title：xxx」），超长自动截断、悬浮显示全文
  - 筛选：用户名 / 操作类型 / 请求方法 / 日期范围，分页展示

### 5.8 ProfilePage — 个人中心（★ 热力图优化）

- **归档活动热力图（GitHub 风格，重新设计）**：
  - 固定 7 行网格（周一 → 周日），每列一周：1月1日所在周的周一为起点、12月31日所在周的周日为终点，首尾跨年格子透明占位，列整齐对齐（53 列 × 7 格）
  - 日期键使用**本地时区** `localDateStr`（避免 `toISOString` 的 UTC 偏移错位），并 `String(t.date).slice(0,10)` 兼容后端 ISO 格式
  - **明暗双配色**（`useDarkMode` 联动）：浅色 `#EFF4F9→#2563EB`、暗色 `#1E293B→#93C5FD`，暗色下不刺眼
  - 月份标签按列定位（含首列偏移）、星期标签固定 7 行、图例「少→多」动态适配暗色、悬停显示「日期 · N条归档」
  - 年份切换（近 4 年）重新请求 `/notes/heatmap`

### 5.9 TagsPage — 标签库管理（★ Bug 修复）

- **新建标签真实入库**：`addTag` 调用 `createTag` API（一级名称 + 可选二级 + 颜色 + 分类 + scope），成功后重新加载列表——修复原先仅本地数组假添加、刷新即消失的问题
- **删除标签**：标签卡片新增删除按钮（hover 红色垃圾桶图标），confirm 二次确认后调 `deleteTag`；后端级联清理所有任务上的该标签关联（含子标签），保持标签一致性
- **随机配色（★ 新增）**：新建标签弹窗打开时默认颜色从 12 色候选集中随机挑选（倾向当前使用较少的颜色，避免同色标签堆积），用户仍可手动选择；任务处（TagSelector）创建标签同样随机配色，不再固定蓝色
- 空态提示、删除中禁用态、暗色模式可读性均已完善

### 5.10 IssuesPage — Bug 反馈列表（★ 新增，GitHub/Gitee Issues 风格）

- 侧边栏「🐛 Bug 反馈」入口（位于「工作成效分析」之后），所有登录用户可访问
- 列表项：类型图标（🐛 Bug / ✨ 预期功能）+ 标题 + `#编号` + 类型徽章（红/绿）+ 状态徽章（🟢开放/🟣已关闭）+ 💬 评论数 + 创建人 + 相对时间（刚刚/N分钟前/N天前）
- 筛选栏：状态三段切换（全部/开放中/已关闭）、类型三段切换（全部类型/Bug/预期功能）、关键词搜索（标题/内容，Enter 触发）+ 重置按钮
- 分页：每页 20 条，共 N 条 + 第 N 页，上一页/下一页
- 新建弹窗：类型单选卡片（🐛 Bug 缺陷 红边 / ✨ 预期功能 绿边）+ 标题 + 详细描述（textarea），校验标题/描述非空，提交成功跳转详情页
- 空态：🐛 图标 + 「暂无问题反馈」引导文案；加载态：居中旋转 spinner

### 5.11 IssueDetailPage — Bug 反馈详情（★ 新增，GitHub Issues 风格）

- 头部：「← 返回问题列表」、类型图标 + 标题 + `#编号` + 类型徽章 + 状态徽章；右上角「关闭 Issue / 重新打开」按钮（仅创建人本人或 super_admin/dept_admin 可见）
- 创建人信息条：头像（首字母圆形）+ 姓名 + 「于 xx 提交」；正文用 `renderNoteContent` 渲染富文本
- 评论区：💬 讨论（N）、评论列表（头像 + 姓名 + 时间 + 内容气泡）、空态「暂无评论，快来发表第一条反馈」
- 评论输入：textarea + 「发表评论」；**问题关闭后输入框与按钮禁用**，提示「该问题已关闭，如需继续讨论请重新打开。」（后端同步拒绝 403）
- 关闭/重开：confirm 确认后调 `PUT /issues/:id/status`，成功后重新加载；关闭权限不足时 alert 后端返回的 message
- 暗色模式：所有卡片/徽章/输入框均有 `dark:` 配色（slate-800/900 背景、浅色文字）

### 5.12 ChatPage — 聊天（★ 新增，仿 QQ/微信在线聊天）

- 侧边栏「💬 聊天」入口（位于「工作成效分析」与「Bug 反馈」之间），路由 `/chat`，全高左右布局
- **左侧栏（300px）**：
  - 顶部：「💬 聊天」标题 + 在线状态徽标（WebSocket 连接状态）+ 搜索框
  - 「会话 / 通讯录」双 tab：
    - 会话：最近会话列表（渐变头像 + 姓名 + 最后消息摘要 + 时间 + 红色未读角标 99+），点击打开
    - 通讯录：全量可见用户列表（姓名 + 部门），支持按姓名/部门搜索，点击发起会话
- **右侧消息主区**：
  - 空态：无会话时提示「选择一个好友开始聊天」
  - 顶部信息条：对方头像 + 姓名 + 在线状态
  - 消息区：气泡左右分栏（自己蓝色右侧 / 对方白色左侧，圆角 + 小尾巴），顶部滚动自动加载更早历史（分页翻页）；图片消息直接渲染 `<img>`（点击放大预览遮罩）；文件消息渲染文件卡片（图标 + 文件名 + 大小 + 点击下载）
  - 输入区：工具栏（😊 表情按钮 / 📎 文件按钮）+ emoji 面板（`EmojiPicker`）+ textarea（Enter 发送、Shift+Enter 换行）+ 发送按钮
- **实时性**：复用全局 WebSocket（`/ws/user/:id`），`store.connectSocket()` 保证通道，收到 `chat:message` 事件自动追加消息 + 刷新会话 + 未读角标
- 暗色模式：深色容器（slate-800/900/950）+ 浅色文字 + 蓝色气泡不变，配色统一

### 5.13 InspectWorkbenchPage — 用户工作台（★ 新增，公司领导/超管）

- 路由 `/workbench/inspect`，权限 `inspect_user_workbench`（super_admin + company_leader）
- 顶部**用户搜索选择器**：输入姓名/部门关键词回车展开候选列表（点击外部关闭），选择目标用户后加载其工作台
- **状态筛选**：待办（含总数）/ 已完成 / 已归档，对应 `GET /notes/users/:userId/workbench?status=active|completed|archived`
- **任务卡片网格**：复用 StickyNoteCard 视图（来源标签、重要红标、创建人、负责人 + 已签收徽标），数据维度 = 目标用户作为创建人/负责人/被指派人/抄送人的任务
- 侧边栏入口「🔍 用户工作台」，公司领导可见、部门管理员等无此权限角色自动隐藏

---

## 六、核心组件设计

### 6.1 StickyNoteCard

- **Props**：`note: Note`、`mode: 'desktop' | 'web'`、`archived: boolean`、`editingBy?: string | null`、`extraActions?: boolean`（★ 新增，工作台开启后卡片显示「重要」「删除」按钮）
- **Events**：`click`、`complete`、`restore`、`export`、`important`（★ 新增）、`delete`（★ 新增）
- **状态视觉**：
  - 黄色（默认待办）：`#FEF3C7` 底色
  - 红色（重要/盯办预警）：`#FEE2E2` 底色 + 脉冲动画
  - 绿色（已完成）：`#DCFCE7` 底色
- **右上角徽章**（★ 新增，需求 18/19/20）：指派任务按视角区分——发起者（当前账号 = `creator_id`）看到浅蓝色卡片 + 蓝色「指派」徽章；接收者看到红色卡片 + 「盯办N」徽章（N 为盯办提醒次数），**签收后变为绿色「已签收」徽章**；**抄送人看到紫色卡片 + 紫色「抄送」徽章**；非指派任务被标记「重要」显示红色 + 「盯办N」；协作任务显示蓝色「协作」徽章
- **签收统计**（★ 新增，需求 19）：发起者视角的指派卡片底部信息区显示「已签收 x/y / 已全部签收」（绿色高亮，悬停 title 显示已签收人姓名，统计排除发起者 initiator）；接收者视角在详情面板操作区上方显示蓝色「签收任务」按钮（签收后隐藏），发起者不显示
- **来源标签**（★ 新增，需求 20）：每张卡片标题上方显示来源标签——自己创建（灰色）/ 自己指派（蓝色，发起者看指派任务）/ 上级指派（橙色，接收者看指派任务）/ 任务抄送（紫色，抄送人）/ 协同任务（青色）
- **操作栏**（★ 新增/调整）：移除「盯办」按钮；新增「重要」按钮（非红色卡片显示，点击置 `color_status='red'` 卡片变红）与「删除」按钮（点击弹出确认框，确认后软删除从工作台移除，可在归档恢复）；**抄送任务卡片仅显示「完成并归档」+「删除」（无「重要」）**；`extraActions` 默认 false，专项工作组详情页/归档页不显示新按钮
- **内容展开**：超过 100 字显示"展开全文"
- **标签展示**：最多 2 个胶囊，超出显示 `+N`；`tags` 为 `undefined` 时安全降级
- **左上角倒计时**（★ 新增）：设定了 `due_time` 的未完成任务，卡片左上角显示「⏱ 剩余 X天X小时 / X小时X分钟 / 已超时 X」，每 30 秒自动刷新；剩余不足 1 小时或已超时显示红色警示徽标；任务完成/归档后自动隐藏
- **任务反馈区块**（★ 新增）：指派任务被指派人提交反馈后，卡片显示绿色「任务反馈」区块（被指派人姓名 + 富文本反馈内容）

### 6.2 TagSelector

- 下拉浮层，支持多选/搜索/Enter 创建新标签
- 已选标签以彩色胶囊展示
- 所有 `<input>` 禁止 Enter 冒泡触发外层表单提交（`@keydown.enter.prevent`）

### 6.3 UserPicker

- 部门树浏览 + 搜索双模式
- 多选/单选，已选人员头像+姓名胶囊
- 搜索框加 `@keydown.enter.prevent` 防止触发外层表单提交
- 支持 `drop-up` 属性控制下拉面板向上/向下展开（弹窗内自动向上避免溢出）
- 部门树每行提供「全选/全不选」按钮：一键选中该部门及其所有子部门的全部人员（受 `max` 限制）
- ★ 支持 `disabledIds` + `disabledNote`（bug4 修复）：已在另一处选择（如指派/抄送互斥）的用户显示灰色半透明 + 提示文案，点击弹出警告且不勾选；部门「全选」自动排除禁用用户
- ★ 点击下拉列表外部空白处自动关闭（bug4 修复：document 级 click 监听，内部点击/「完成」按钮不受影响）

### 6.4 RichTextEditor

- **Props**：`modelValue`(HTML字符串)、`placeholder`、`minHeight`
- **Events**：`update:modelValue`
- 基于 `contenteditable` + `document.execCommand` 实现，无第三方依赖
- 工具栏：撤销/重做、标题、加粗、斜体、下划线、删除线、有序/无序列表、引用、代码块、清除格式
- 支持回车换行（默认内容块行为），placeholder 空态提示
- 与 `utils/richText.ts` 配合：旧纯文本内容渲染时自动转义并转 `<br>`，新富文本内容原样渲染

### 6.5 NotificationBell — 通知铃铛（★ 新增）

- 右上角铃铛按钮 + 未读数角标（>99 显示 99+），点击展开下拉通知面板
- 通知列表：类型徽标（指派/完成/反馈/催办）、标题、摘要、相对时间、未读高亮
- 支持「全部已读」、单条删除；点击消息 → 详情弹窗（富文本渲染），可「查看任务」跳转工作台对应任务
- 挂载时拉取未读数，通过 Pinia store 订阅 WebSocket 实时增量

### 6.6 ChatDrawer — 聊天面板（★ 新增）

- 右侧抽屉：会话列表（头像/最近消息/未读数）+ 消息区（气泡、时间、富文本渲染）+ 输入区（Enter 发送 / Shift+Enter 换行）
- 会话未读角标同步到顶部栏聊天图标
- 从任务详情「联系」入口可直达指定会话（`openChat(peerId)`）

### 6.7 FeedbackModal — 任务反馈填报（★ 新增）

- 完成并归档任务时弹出，复用 `RichTextEditor` 填写反馈内容
- 支持两种模式：`complete`（提交反馈并完成）/ `feedback`（归档后补充反馈）
- 归档详情面板提供「反馈填报」按钮补充提交，提交后通知任务发起人

### 6.8 工具函数 `utils/sound.ts`（★ 新增）

- 基于 Web Audio API 合成提示音（无外部资源）：新通知双音提示、新聊天单音提示

### 6.9 NotificationPage — 通知中心（★ 新增）

- 独立路由 `/notifications`：展示全部通知（分页 + 加载更多，每页 20 条）
- 功能：只看未读筛选、全部已读、单条删除、点击查看详情弹窗、行内「查看任务」跳转工作台并弹出对应任务便签、返回工作台
- 铃铛下拉面板仅展示最近 10 条，底部「查看全部通知 →」跳转本页

### 6.10 StickyNoteCard — 任务反馈区块（★ 新增）

- 便签卡片在 `assignees[].feedback_content` 存在时，显示绿色「任务反馈」区块（被指派人姓名 + 反馈内容，富文本渲染）
- 指派人与被指派人查看同一任务便签时内容与状态同步（后端同一数据行 + 列表预加载 assignees）

### 6.11 夜间模式（Dark Mode）适配（★ 新增）

- 机制：`composables/useDarkMode.ts` 通过 `localStorage`（`labelpro_dark_mode`）持久化，在 `<html>` 根节点添加/移除 `dark` class 与 `data-theme` 属性；Tailwind `darkMode: 'class'` 下所有 `dark:` 前缀类生效
- 便签卡片（StickyNoteCard）：背景由内联样式硬编码浅色改为 Tailwind 类 + `dark:` 深色变体
  - 浅色：`amber-100 / red-100 / blue-100 / green-100`（对应原 `#FEF3C7/#FEE2E2/#DBEAFE/#DCFCE7`）
  - 暗色：`dark:amber-900/60 / dark:red-900/60 / dark:blue-900/60 / dark:green-900/60` + 深色边框与 `border-l` 强调色
  - 标题 `dark:text-slate-100`、正文 `dark:text-slate-300`、按钮/分隔线/反馈区/水印均补暗色变体
- 倒计时徽章与盯办/协作徽章加深底色（`amber-700` / `red-600`）提升白字对比度
- 管理页面（UsersPage / DepartmentsPage / TagsPage / TemplatesPage / ArchivePage / WorkbenchPage）：表格、卡片、按钮、徽章、弹窗、输入框、下拉、文本域等全部补齐 `dark:` 配色

### 6.12 EmojiPicker — 表情选择器（★ 新增）

- 组件 `src/components/chat/EmojiPicker.vue`，`@select` 事件返回所选 emoji
- 9 个分类 tab：「表情包」+ 笑脸/手势/动物/食物/活动/旅行/物品/符号，`8 列网格`、悬停高亮、面板 300px 宽、分类区 176px 高可滚动
- 表情包为高频趣味 emoji 组合（😀😂🎉💪 等 40 个），纯 Unicode 实现，内网无外部资源依赖

---

## 七、状态管理（Pinia Stores）

### 7.1 `useNoteStore`

- **State**：`activeNotes`、`selectedNote`、`loading`
- **Actions**：`fetchNotes`、`createNote`、`completeNote`（支持携带反馈内容）、`remindNote`、`updateNoteLocally`

### 7.2 `useAuthStore`

- **State**：`user`、`token`
- 登录后存储 token 到 localStorage，自动注入请求头

### 7.3 `useNotificationStore`（★ 新增）

- **State**：`unreadCount`、`notifications`、`conversations`、`messages`、`connected`、`chatOpen`、`chatPeerId`
- **Actions**：`connectSocket`（建立 `/ws/user/:user_id` 个人通知通道）、`fetchUnreadCount`、`fetchList`、`markRead`、`markAllRead`、`remove`、`refreshConversations`、`loadMessages`、`sendMessage`、`markConversationRead`、`openChat`、`closeChat`
- WebSocket 事件：`notification:new`（未读数+1、插入列表、播放提示音）、`chat:message`（追加消息、刷新会话、播放提示音）

---

## 八、服务层（Services）

| 服务文件        | 主要函数                                                                                                                     | 说明                   |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| `api.ts`        | `get / post / put / del`                                                                                                     | Axios 封装，统一拦截器 |
| `admin.ts`      | `login / getDepartments / getUsers / createUser / updateUser / deleteUser / getVisibleUsers / getUserProfile / getUsersWithStats / recommendUsers / getWorkTypeOptions` | 认证+组织+人员+推荐   |
| `notes.ts`      | `fetchNotes / createNote / updateNote / completeNote / remindNote / deleteNote / restoreNote / fetchNoteStats / exportNotes / fetchHeatmap` | 任务 CRUD + 统计 + 热力图 |
| `tags.ts`       | `fetchTags / createTag / updateTag / deleteTag`                                                                              | 标签管理               |
| `templates.ts`  | `fetchTemplates / fetchTemplateById / createTemplate / updateTemplate / deleteTemplate`                                      | ★ 模板管理 CRUD        |
| `workgroup.ts`  | `searchGroups / getMyGroups / getWorkGroupDetail / createWorkGroup / deleteWorkGroup / getWorkGroupMembers`                  | 工作组 CRUD + 搜索     |
| `groupNotes.ts` | `getGroupNotes / createGroupNote`                                                                                            | 专属任务               |
| `analytics.ts`  | `fetchPersonalStats / generateAIReport / fetchReports / fetchReportTemplate / saveReportTemplate`                            | 分析+报告模板          |
| `system.ts`     | `fetchConfig / updateConfig / fetchOperations`                                                                               | 系统配置+日志          |
| `notification.ts` | `fetchNotifications / fetchUnreadCount / markNotificationRead / markAllNotificationsRead / deleteNotification / fetchConversations / fetchChatMessages / sendChatMessage / markConversationRead / fetchReminders / acknowledgeReminder` | ★ 通知+聊天+提醒       |

---

## 九、设计规范

### 9.1 色彩体系

- **主背景**：`#FFFFFF` / `#F8FAFC`（暗色：`#0F172A` / `#1E293B`）
- **主文字**：`#0F172A`（标题）/ `#475569`（正文）（暗色：`#F1F5F9` 标题 / `#CBD5E1` 正文）
- **交互蓝**：`#3B82F6`（暗色 `#60A5FA`）
- **专项紫蓝渐变**：`from-purple-500 to-blue-500`
- **任务三色**：黄 `#FEF3C7` / 绿 `#DCFCE7` / 红 `#FEE2E2`（暗色：`amber-900/60` / `green-900/60` / `red-900/60`，配合浅色文字保证对比度）

### 9.2 圆角与阴影

- 卡片：`rounded-xl`（12px），hover 时 `shadow-md`
- 按钮：`rounded-btn`（10px）
- 模态框：`rounded-card`（16px）+ `shadow-modal`

### 9.3 动效

- 全局过渡：`transition-smooth`（all 0.3s cubic-bezier）
- 任务插入：`animate-spring-enter`
- 盯办脉冲：2s infinite pulse

### 9.4 字体（★ 内网本地化）

- **无互联网依赖**：字体完全本地化，页面加载 0 个外网请求，适配内网/离线部署
- 拉丁字体：`@fontsource/inter`（400/500/600/700），经 `main.ts` import 后由 Vite 打包为本地 woff2，替代原 Google Fonts 在线加载（已移除 `style.css` 中的 `@import url('https://fonts.googleapis.com/...')`）
- 中文字体：回退本地系统字体栈 `PingFang SC`（macOS）/ `Microsoft YaHei`（Windows）/ `Noto Sans CJK SC`，不下载任何中文字体文件

---

_文档结束_
