# 资警数智·轻燕 —— Web前端开发文档

> 版本：v1.0  
> 定位：轻量化情指行一体化支撑解决方案 · Web管理端  
> 核心载体：桌面智能便签协同办公系统的Web管理端与即时协同大屏端

---

## 一、项目概述

Web前端作为系统的管理中枢与可视化展示端，承接桌面端便签的云端归档、跨终端检索、组织人员管理、模板配置及应急大屏协同等功能。页面以**纯白为底、极简为形、高级交互为魂**，打造零学习成本的公司办公新体验。

---

## 二、技术栈

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 框架 | Vue 3 (Composition API) | 响应式架构，适配复杂状态流转 |
| UI组件库 | DaisyUI + Tailwind CSS | 原子化样式，快速构建高质感界面 |
| 状态管理 | Pinia | 跨组件便签状态、用户权限、协同数据同步 |
| 路由 | Vue Router 4 | 按需加载，适配大屏与普通管理端双模式 |
| HTTP客户端 | Axios | RESTful API统一拦截，自动注入Token与权限头 |
| 实时通信 | Socket.io-client | 毫秒级协同编辑与大屏同步 |
| 富文本编辑 | TipTap / Slate | 便签内容轻量级富文本，支持图文混排 |
| 构建工具 | Vite | 极速HMR，优化打包体积 |

---

## 三、设计规范（Design System）

### 3.1 色彩体系
- **主背景**：`#FFFFFF`（纯白） + `#F8FAFC`（极浅灰，用于区分区块）
- **主文字**：`#0F172A`（ slate-900，标题） / `#475569`（slate-600，正文）
- **便签色彩语义**：
  - 待办（黄）：`#FEF3C7` 背景 + `#D97706` 边框/标识
  - 完成（绿）：`#DCFCE7` 背景 + `#16A34A` 标识
  - 盯办预警（红）：`#FEE2E2` 背景 + `#DC2626` 强警示，配合脉冲动画
- **交互蓝**：`#3B82F6`（仅用于主按钮、链接、选中态）
- **禁用态**：`#E2E8F0` + `#94A3B8`

### 3.2 字体与排版
- 字体族：`"Inter", "Noto Sans SC", -apple-system, sans-serif`
- 标题：20-24px，font-weight 600，letter-spacing -0.02em
- 正文：14px，line-height 1.6
- 标签/辅助：12px，slate-500

### 3.3 间距与圆角
- 卡片圆角：16px（大圆角，柔和现代感）
- 按钮圆角：10px
- 标签圆角：6px
- 基础间距：4px网格系统，卡片内边距 20px-24px

### 3.4 阴影与层级
- 便签卡片：`0 4px 24px -4px rgba(0,0,0,0.08)`，hover时上浮 `translateY(-2px)` + 阴影加深
- 弹窗/模态框：`0 24px 48px -12px rgba(0,0,0,0.18)`
- 盯办红色便签：附加 `box-shadow: 0 0 0 4px rgba(220,38,38,0.2)` 脉冲动画

### 3.5 动效规范
- 所有交互过渡：`transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1)`
- 便签归档消失：`opacity 0 → 1` + `scale 0.95 → 1` + `translateY(20px → 0)`，持续时间 400ms
- 盯办脉冲：2s infinite，关键帧 `scale(1)` → `scale(1.02)` + 阴影扩散

---

## 四、页面结构与路由设计

### 4.1 路由表

| 路由 | 页面名称 | 权限要求 | 说明 |
|------|---------|---------|------|
| `/login` | 登录页 | 公开 | 极简白底，公司蓝点缀，扫码/账密双模式 |
| `/workbench` | 个人工作台 | 已登录 | 核心页面，便签墙布局 |
| `/workbench/archive` | 归档查询 | 已登录 | 按标签/人员/时间检索历史便签 |
| `/workbench/collaboration/:id` | 协同编辑室 | 协同成员 | 多人同屏画布，分栏布局 |
| `/admin/departments` | 部门库管理 | dept_admin+ | 组织架构维护 |
| `/admin/users` | 人员库管理 | dept_admin+ | 人员增删改查、角色分配 |
| `/admin/tags` | 标签库管理 | dept_admin+ | 自定义属性标签维护 |
| `/admin/templates` | 模板库管理 | super_admin | 工作模板配置 |
| `/screen/:id` | 应急大屏端 | screen_role | 只读大屏，4/6栏实时同步展示 |
| `/profile` | 个人中心 | 已登录 | 我的便签统计、偏好设置 |

### 4.2 布局架构
- **管理端布局**：左侧极简侧边栏（icon+文字，可折叠） + 顶部通栏（面包屑+搜索+通知中心） + 主内容区纯白底
- **大屏端布局**：全屏无边框，深色底（仅大屏场景反色，管理端仍为白底），分栏自适应网格

---

## 五、核心组件设计

### 5.1 StickyNoteCard（便签卡片组件）
- **Props**：`note`（便签对象）、`mode`（desktop/web）、`permissions`（当前用户权限）
- **状态视觉**：
  - 黄色：左上角4px色条标识，轻微闪烁（仅限桌面端接收前3分钟）
  - 绿色：整体边框变绿，出现"已完成"角标，操作区仅保留"撤销归档"
  - 红色：边框红色脉冲，右上角"盯办"徽章，强提醒样式
- **交互**：
  - 单击：展开详情侧滑面板（右侧滑出，不跳页）
  - 长按/右键：上下文菜单（编辑标签、盯办、转派、删除）
  - 拖拽：工作台内自由排序（本地记忆，可选同步到服务端）
- **内容区**：支持折叠/展开，默认展示前3行，超出渐变遮罩

### 5.2 TagSelector（标签选择器）
- 下拉浮层，白底大圆角，支持多选、搜索、新建标签
- 标签胶囊展示，自带语义色点（标签可配置颜色）
- 最近使用标签置顶

### 5.3 UserPicker（人员选择器）
- 树形组织架构 + 搜索双模式
- 部门库展开，人员库勾选，支持批量选择
- 已选人员以头像+姓名胶囊展示，组长身份特殊角标

### 5.4 CollaborationCanvas（协同画布）
- 网格分栏（2/4/6/8栏可选），每栏顶部固定人员信息条
- 每栏独立富文本区，但共享同一WebSocket Room
- 底部实时显示"XX正在输入..."
- 大屏投屏按钮：一键进入全屏无UI模式

### 5.5 ArchiveSearch（归档检索面板）
- 顶部筛选栏：时间范围、标签多选、人员选择、部门选择、关键词
- 结果区：时间轴/卡片双视图切换
- 操作：查看详情、导出Word、恢复至工作台

---

## 六、状态管理（Pinia Stores）

### 6.1 `useNoteStore`
- **State**：`activeNotes`（当前活跃便签列表）、`archivedNotes`（归档列表，分页）、`currentNote`（详情）、`loading`
- **Actions**：
  - `fetchNotes(filters)`：按权限获取便签列表
  - `createNote(payload)`：创建便签，成功后本地prepend
  - `completeNote(id)`：完成任务，本地移除+归档（动画后）
  - `archiveNote(id)`：手动归档
  - `remindNote(id)`：盯办，更新状态为红色
  - `syncFromDesktop(note)`：接收桌面端推送同步

### 6.2 `useAuthStore`
- **State**：`user`（当前用户）、`token`、 `permissions`（权限矩阵）
- **Getters**：`isAdmin`、`isDeptAdmin`、`canCreateForOthers`（是否可为他人创建便签）

### 6.3 `useCollaborationStore`
- **State**：`roomId`、`participants`、`canvasData`（分栏内容映射）、`syncStatus`
- **Actions**：`joinRoom`、`leaveRoom`、`pushLocalChange(columnId, content)`、`handleRemoteChange(data)`

---

## 七、核心业务逻辑与接口调用规范

### 7.1 便签生命周期（前端视角）

```
创建 → 编辑/贴标签 → （协同/指派） → 完成提交 → 归档动画 → 从活跃列表移除 → 进入归档库
```

**业务规则**：
1. **创建权限**：
   - 普通用户：仅能创建 `source_type = self` 的便签，owner_id 为自己
   - 部门管理员：可创建 `source_type = assigned` 的便签，owner_id 指定为本部门人员
   - 系统管理员：可创建给任何人，跨部门指派
   - 任何用户创建多人协同时，自动成为发起人（creator_id），可拉取任意可见人员

2. **标签贴附**：
   - 一个便签可贴 0-N 个标签
   - 标签来自系统标签库 + 个人自定义标签（个人标签仅自己可见，系统标签全员可见）
   - 标签在卡片上以胶囊形态展示，最多显示2个，超出显示 `+N`

3. **完成与归档**：
   - 用户点击"完成"按钮 → 前端先播放归档动画（卡片缩小+透明度降低+位移）→ 同时调用 `POST /api/v1/notes/{id}/complete`
   - 服务端返回成功后，该卡片从 `activeNotes` 移除，若用户在归档页则加入 `archivedNotes`
   - 归档后，桌面端对应便签同步消失（通过WebSocket或轮询同步）

4. **盯办预警**：
   - 发起人/组长在Web端点击"盯办" → 调用 `POST /api/v1/notes/{id}/remind`
   - 被盯办人桌面端便签立即变红（WebSocket推送），Web端卡片同步变红

### 7.2 接口调用规范

所有接口统一通过 Axios 实例调用，BaseURL 由环境变量配置。

#### 7.2.1 便签管理接口

| 接口 | 方法 | 路径 | 请求参数 | 响应数据 | 前端处理逻辑 |
|------|------|------|---------|---------|-------------|
| 获取便签列表 | GET | `/api/v1/notes` | `?status=active&tag_id=&dept_id=&owner_id=&keyword=&page=1&page_size=20` | `{data:[],total,page}` | 根据当前路由/标签页自动附加过滤条件；首次加载骨架屏，后续增量更新 |
| 创建便签 | POST | `/api/v1/notes` | `{title,content,tags:[],source_type,due_time,owner_id,template_type,group_id,assignees:[]}` | `{data:Note}` | 成功后本地插入列表顶部；若owner_id≠自己，调用推送接口通知服务端下发桌面 |
| 获取便签详情 | GET | `/api/v1/notes/{id}` | - | `{data:Note}` | 右侧滑出详情面板，富文本渲染 |
| 更新便签 | PUT | `/api/v1/notes/{id}` | `{title,content,tags,status,due_time}` | `{data:Note}` | 乐观更新：先改本地状态，后调接口，失败回滚 |
| 完成便签 | POST | `/api/v1/notes/{id}/complete` | `{feedback_content,attachments:[]}` | `{data:Note}` | 触发归档动画，400ms后从活跃列表移除；若处于协同模式，通知所有参与者 |
| 盯办提醒 | POST | `/api/v1/notes/{id}/remind` | `{remind_type='urgent',message}` | `{data:Note}` | 本地卡片变红，Toast提示"盯办已发送" |
| 删除/归档 | DELETE | `/api/v1/notes/{id}` | `{soft=true}` | `{success}` | 软删除，移至归档；物理删除仅管理员可操作 |
| 恢复便签 | POST | `/api/v1/notes/{id}/restore` | - | `{data:Note}` | 从归档恢复至活跃列表，插入顶部 |

#### 7.2.2 标签管理接口

| 接口 | 方法 | 路径 | 请求参数 | 响应数据 | 前端处理逻辑 |
|------|------|------|---------|---------|-------------|
| 获取标签库 | GET | `/api/v1/tags` | `?scope=personal/system/all` | `{data:[{id,name,color,scope,category}]}` | 创建便签时预加载，本地缓存 |
| 创建标签 | POST | `/api/v1/tags` | `{name,color,category,scope}` | `{data:Tag}` | 即时加入本地标签池，无需刷新 |
| 更新标签 | PUT | `/api/v1/tags/{id}` | `{name,color}` | `{data:Tag}` | 同步更新所有便签上的标签展示 |

#### 7.2.3 组织人员接口

| 接口 | 方法 | 路径 | 请求参数 | 响应数据 | 前端处理逻辑 |
|------|------|------|---------|---------|-------------|
| 获取部门树 | GET | `/api/v1/departments` | `?flat=false` | `{data:TreeNode[]}` | 人员选择器、分组配置器使用 |
| 获取人员列表 | GET | `/api/v1/users` | `?dept_id=&role=&keyword=&page=1` | `{data:[{id,name,dept_name,avatar,role}]}` | 支持拼音首字母搜索 |
| 获取当前用户可见人员 | GET | `/api/v1/users/visible` | - | `{data:[]}` | 创建协同时快速拉取，按最近协同频率排序 |

#### 7.2.4 协同与模板接口

| 接口 | 方法 | 路径 | 请求参数 | 响应数据 | 前端处理逻辑 |
|------|------|------|---------|---------|-------------|
| 获取模板列表 | GET | `/api/v1/templates` | `?type=` | `{data:[{id,name,fields,layout}]}` | 创建便签时弹窗选择 |
| 创建工作组 | POST | `/api/v1/groups` | `{name,note_id,members:[{user_id,role}]}` | `{data:Group}` | 专项工作模式核心接口 |
| 获取工作组成员 | GET | `/api/v1/groups/{id}/members` | - | `{data:[{user,role}]}` | 组长/组员权限区分展示 |
| 加入协同房间 | WS | `/ws/notes/{id}` | Socket.io handshake | Room events | 进入协同编辑室时连接，断线自动重连 |
| 推送画布变更 | WS emit | `canvas:update` | `{column_id,content,delta,user_id}` | broadcast | 本地TipTap编辑器onChange节流300ms发送 |
| 接收画布变更 | WS on | `canvas:sync` | `{column_id,content,updated_by}` | - | 非本栏更新时，合并至对应栏位 |

#### 7.2.5 文号与台账接口

| 接口 | 方法 | 路径 | 请求参数 | 响应数据 | 前端处理逻辑 |
|------|------|------|---------|---------|-------------|
| 获取文号 | GET | `/api/v1/serial-numbers/next` | `?type=work_report` | `{data:{serial_no}}` | 生成Word报告时自动填充 |
| 生成报告 | POST | `/api/v1/notes/{id}/export` | `{format='word',template_id}` | `{data:{download_url}}` | 显示生成进度，完成后新窗口下载 |
| 获取台账 | GET | `/api/v1/ledger` | `?date_from=&date_to=&dept_id=` | `{data:[],summary}` | 数据可视化图表展示 |

---

## 八、权限控制矩阵（前端路由与按钮级）

| 功能 | 系统管理员 | 部门管理员 | 组长 | 普通用户 | 说明 |
|------|-----------|-----------|------|---------|------|
| 创建便签（给自己） | ✓ | ✓ | ✓ | ✓ | 基础功能 |
| 创建便签（给他人） | ✓ | ✓（限本部门） | ✗ | ✗ | 需 `can_assign` 权限位 |
| 编辑他人便签 | ✓ | ✓（限本部门） | ✓（限本组） | ✗ | 仅便签创建者/上级/组长可编辑 |
| 删除便签 | ✓ | ✓（限本部门） | ✗ | ✗（仅自己创建的） | 软删除，管理员可恢复 |
| 盯办操作 | ✓ | ✓ | ✓ | ✗ | 仅发起人、组长、管理员 |
| 查看归档（全部） | ✓ | ✓（限本部门） | ✓（限本组） | ✓（仅自己） | 数据权限隔离 |
| 管理部门/人员库 | ✓ | ✓（限本部门） | ✗ | ✗ | 组织架构维护 |
| 配置系统模板 | ✓ | ✗ | ✗ | ✗ | 仅超管 |
| 进入大屏端 | ✓ | ✓ | ✓ | ✗ | 需分配大屏角色 |

**实现方式**：
- 路由守卫：进入 `/admin/*` 前校验 `useAuthStore.permissions`
- 按钮级：`v-permission="['remind']"` 自定义指令，无权限时按钮隐藏或禁用并Tooltip提示
- 数据级：即使前端隐藏按钮，后端仍严格校验，前端以服务端返回的 `allowed_actions` 字段为准动态渲染操作栏

---

## 九、关键交互细节

### 9.1 工作台便签墙（核心页面）
- **布局**：CSS Grid 自适应，`grid-template-columns: repeat(auto-fill, minmax(280px, 1fr))`，gap 20px
- **空态**：纯白底中央插画 + "点击右下角 '+' 新建便签"，无冗余文字
- **新建流程**：
  1. 点击悬浮按钮（右下角，56px圆形，蓝色，带阴影）
  2. 弹出创建模态框：标题输入框（自动聚焦）+ 内容区 + 标签选择器 + 底部操作栏
  3. 可选择"仅自己"/"指派他人"/"开启协同"，切换时动态显示人员选择器
  4. 确认创建后，新卡片以 `scale(0.8) opacity(0)` → 正常状态的弹簧动画插入首位

### 9.2 协同编辑室
- 进入时检测权限：非成员跳转403页
- 每栏顶部：部门徽章 + 姓名 + 在线状态绿点
- 内容区：TipTap编辑器，菜单栏悬浮（加粗、列表、链接）
- 领导指挥输入框：底部固定区域，输入后点击"下发指令"，以系统消息形式插入所有栏位顶部（红色边框消息条）
- 投屏模式：隐藏所有UI控件，仅展示画布内容，URL带 `?mode=projection`

### 9.3 归档与追溯
- 时间轴视图：以月分组，垂直时间线，左侧日期右侧卡片
- 卡片视图：与工作台一致，但带"已归档"水印角标
- 操作：支持批量导出Excel/Word，导出时自动附加文号

---

## 十、大屏端特殊适配

- **分辨率适配**：支持 1920×1080 至 5760×2160 自适应，字体使用 vw/vh 单位
- **数据刷新**：WebSocket推送为主，心跳检测30s，断线自动重连并显示"连接中..."浮层
- **只读模式**：大屏端禁止编辑，仅接收同步，领导指令通过独立WebSocket频道下发
- **夜间模式**：大屏默认深色底（`#0F172A`），与Web端管理端白底区分，便签卡片反色适配

---

## 十一、错误处理与边界情况

- **网络断开**：所有写操作进入本地队列，恢复后批量同步；界面显示"离线模式"轻提示
- **权限变更**：WebSocket推送权限更新，即时生效，无刷新
- **并发编辑**：协同场景采用 Last-Write-Wins + 操作者提示，冲突时显示"XX刚刚修改了此处，是否合并？"
- **大数据量**：便签墙虚拟滚动（>100张），归档查询服务端分页

---

*文档结束*
