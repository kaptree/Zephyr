# 任务05：核心组件开发

## 任务目标
开发5个核心可复用组件：StickyNoteCard、TagSelector、UserPicker、CollaborationCanvas、ArchiveSearch。

## 依赖关系
- 依赖任务02（核心基础设施）完成
- 可独立于页面开发（仅需类型定义和 Store 骨架）

## 技术要求
- 遵循文档第5节核心组件设计
- 遵循文档第3节设计规范（色彩、动效、圆角、阴影）
- 组件 props 使用 TypeScript 泛型定义
- 组件事件使用 `defineEmits` 声明
- 使用 Composition API (`<script setup lang="ts">`)

## 具体步骤

### 5.1 StickyNoteCard (`src/components/note/StickyNoteCard.vue`)

**Props**：
```typescript
interface Props {
  note: Note
  mode?: 'desktop' | 'web'
  permissions?: string[]
}
```

**视觉状态**：
- **待办（黄色）**：`bg-[#FEF3C7]` 背景，左上角 4px 宽 `#D97706` 色条
- **完成（绿色）**：`bg-[#DCFCE7]` 背景，`#16A34A` 边框，"已完成"角标
- **盯办（红色）**：`bg-[#FEE2E2]` 背景，`#DC2626` 边框 + 脉冲动画 + "盯办"徽章

**交互**：
- 单击 → emit `click` 事件（用于打开详情面板）
- 右键/长按 → 弹出上下文菜单（编辑标签、盯办、转派、删除）
- 拖拽 → emit `drag-start` / `drag-end` 事件

**内容区**：
- 默认展示前3行，超出显示渐变遮罩
- 点击可展开/折叠

**标签展示**：
- 胶囊形态，最多显示2个，超出显示 `+N`
- 标签自带语义色点

**CSS 动画**：
- 悬停：`translateY(-2px)` + 阴影加深
- 新建进入：弹簧动画 `scale(0.8) → scale(1.03) → scale(1)` 300ms
- 归档消失：`opacity(1→0)` + `scale(1→0.95)` + `translateY(0→20px)` 400ms
- 盯办脉冲：2s infinite keyframes

### 5.2 TagSelector (`src/components/common/TagSelector.vue`)

**Props**：
```typescript
interface Props {
  modelValue: string[]  // 已选标签ID数组
  max?: number          // 最大可选数量
  scope?: 'personal' | 'system' | 'all'
}
```

**功能**：
- 下拉浮层，白底大圆角（16px），阴影
- 搜索框 + 标签列表（勾选）
- 已选标签顶部胶囊展示（带 × 移除）
- "创建新标签"按钮（触发 emit `create-tag`）
- 最近使用标签置顶（localStorage 记忆）
- 标签带语义色点
- 多选支持

### 5.3 UserPicker (`src/components/common/UserPicker.vue`)

**Props**：
```typescript
interface Props {
  modelValue: string[]  // 已选用户ID数组
  multiple?: boolean
  max?: number
}
```

**功能**：
- 树形组织架构展开（部门 → 人员）
- 搜索框（支持姓名/拼音首字母）
- 已选人员以头像+姓名胶囊展示
- 组长身份特殊角标
- 批量选择支持

### 5.4 CollaborationCanvas (`src/components/note/CollaborationCanvas.vue`)

**Props**：
```typescript
interface Props {
  roomId: string
  participants: Participant[]
  columns: number  // 2/4/6/8
}
```

**功能**：
- 网格分栏布局（CSS Grid，columns 控制列数）
- 每栏顶部固定人员信息条（部门徽章 + 姓名 + 在线绿点）
- 每栏独立富文本区（预留 TipTap 集成）
- 底部"正在输入..."状态条
- 投屏按钮：一键切换全屏无UI模式

### 5.5 ArchiveSearch (`src/components/note/ArchiveSearch.vue`)

**Props**：
```typescript
interface Props {
  // 父组件传递的初始筛选条件
  initialFilters?: ArchiveFilters
}
```

**功能**：
- 顶部筛选栏：时间范围（DateRangePicker）、标签多选（TagSelector）、人员选择（UserPicker）、部门选择、关键词
- 结果区：时间轴视图 / 卡片视图 切换按钮
- 时间轴视图：以月分组，垂直时间线，左侧日期右侧卡片
- 卡片视图：与工作台一致，带"已归档"水印
- 操作按钮：查看详情、导出Word、恢复至工作台
- 分页加载

## 验收标准
1. StickyNoteCard 三种颜色状态正确渲染
2. StickyNoteCard 单击/右键交互正常
3. TagSelector 多选、搜索、新建标签流程完整
4. UserPicker 树形展开和搜索功能正常
5. CollaborationCanvas 分栏网格正常布局
6. ArchiveSearch 筛选和双视图切换正常
7. 所有组件符合设计规范中的色彩/圆角/阴影规范
8. TypeScript 类型检查通过

## 预计工时：6小时

## 交付物
- `src/components/note/StickyNoteCard.vue`
- `src/components/common/TagSelector.vue`
- `src/components/common/UserPicker.vue`
- `src/components/note/CollaborationCanvas.vue`
- `src/components/note/ArchiveSearch.vue`
