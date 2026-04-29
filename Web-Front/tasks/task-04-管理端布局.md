# 任务04：管理端布局框架

## 任务目标
实现管理端通用布局框架：左侧可折叠侧边栏、顶部通栏、主内容区。

## 依赖关系
- 依赖任务03（认证系统）完成

## 技术要求
- 遵循文档第4.2节布局架构
- 遵循文档第3节设计规范
- 使用 Tailwind CSS + DaisyUI

## 具体步骤

### 4.1 管理端布局组件 (`src/layouts/AdminLayout.vue`)

采用 flex 布局，结构如下：
```
┌──────────┬────────────────────────────────────┐
│          │  顶部通栏（面包屑 + 搜索 + 通知）      │
│  侧边栏   ├────────────────────────────────────┤
│  (可折叠) │                                    │
│          │       主内容区（纯白底）              │
│          │     <router-view />                 │
│          │                                    │
└──────────┴────────────────────────────────────┘
```

### 4.2 侧边栏组件 (`src/components/layout/Sidebar.vue`)

#### 视觉规范
- 宽度：展开 240px / 折叠 64px
- 背景：`#F8FAFC`（slate-50），与主内容区 `#FFFFFF` 区分
- 选中态：交互蓝（`#3B82F6`）左侧 3px 竖条 + 背景 `#EFF6FF`
- 图标：使用 Heroicons（或 SVG inline）
- 底部：用户头像 + 名称 + 登出按钮

#### 菜单项
| 图标 | 文字 | 路由 | 权限要求 |
|------|------|------|---------|
| 📋 | 工作台 | `/workbench` | 已登录 |
| 📁 | 归档查询 | `/workbench/archive` | 已登录 |
| 👥 | 协同编辑 | `/workbench/collaboration/:id` | 协同成员 |
| 🏢 | 部门管理 | `/admin/departments` | dept_admin+ |
| 👤 | 人员管理 | `/admin/users` | dept_admin+ |
| 🏷️ | 标签管理 | `/admin/tags` | dept_admin+ |
| 📄 | 模板管理 | `/admin/templates` | super_admin |
| 📊 | 应急大屏 | `/screen/:id` | screen_role |
| ⚙️ | 个人中心 | `/profile` | 已登录 |

#### 交互
- 折叠/展开：侧边栏顶部 hamburger 按钮
- 折叠时仅显示图标，hover 显示 tooltip 文字
- 折叠状态记忆到 localStorage
- 无权限菜单项不渲染（或灰色禁用+tooltip）

### 4.3 顶部通栏组件 (`src/components/layout/TopBar.vue`)

- **面包屑导航**：根据当前路由自动生成
- **全局搜索框**：支持搜索便签、人员、标签（带下拉结果面板）
- **通知中心**：铃铛图标 + 未读红点，点击展开通知列表
- **高度**：56px
- **背景**：`#FFFFFF`，底部 1px 边框 `#E2E8F0`

### 4.4 主内容区

- 背景 `#F8FAFC`，内嵌白色内容卡片
- 使用 `<router-view />` 渲染子页面
- 过渡动画：`<transition name="fade">` 包裹 router-view

### 4.5 移动端响应式适配（预留）

- 768px 以下侧边栏变为全屏遮罩抽屉
- 顶部通栏简化

## 验收标准
1. 侧边栏菜单完整展示10个入口
2. 折叠/展开动画流畅（`transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1)`）
3. 折叠状态记忆正常工作
4. 面包屑正确反映当前路由层级
5. 权限不足的菜单项正确隐藏
6. 路由切换时内容区正常渲染

## 预计工时：3小时

## 交付物
- `src/layouts/AdminLayout.vue`
- `src/components/layout/Sidebar.vue`
- `src/components/layout/TopBar.vue`
- `src/components/layout/Breadcrumb.vue`
- `src/components/layout/NotificationCenter.vue`
