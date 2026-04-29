# 任务12：个人中心页面

## 任务目标
实现个人中心页面，展示用户便签统计数据、个人偏好设置。

## 依赖关系
- 依赖任务04（管理端布局）完成

## 技术要求
- 遵循文档第4.1节路由设计（`/profile`）
- 集成于管理端布局内

## 具体步骤

### 12.1 页面路由
- 路由路径：`/profile`

### 12.2 页面结构 (`src/pages/ProfilePage.vue`)

```
┌──────────────────────────────────────────────────────────┐
│  个人中心                                                  │
├──────────────────────┬───────────────────────────────────┤
│                      │                                    │
│  [头像]              │  基本信息                          │
│  张三                │  姓名：张三                        │
│  刑警支队            │  部门：刑警支队 · 侦查一队          │
│  组长                │  角色：组长                        │
│                      │  邮箱：zhangsan@police.gov.cn      │
│                      │                                    │
├──────────────────────┼───────────────────────────────────┤
│                      │                                    │
│  我的统计             │  便签统计                         │
│                      │  ┌─────┬─────┬─────┬─────┐       │
│                      │  │活跃 │已完成│盯办 │归档 │       │
│                      │  │ 12  │ 45  │ 3   │ 89  │       │
│                      │  └─────┴─────┴─────┴─────┘       │
│                      │                                    │
│                      │  本周趋势（柱状图）                 │
│                      │  [▄] [▆] [▃] [▅] [▇] [▂] [█]     │
│                      │  周一 周二 周三 周四 周五 周六 周日   │
│                      │                                    │
├──────────────────────┼───────────────────────────────────┤
│                      │                                    │
│  [保存] [退出登录]    │  偏好设置                          │
│                      │  [开关] 桌面端便签同步提醒           │
│                      │  [开关] 完成时播放音效               │
│                      │  [开关] 盯办强提醒                   │
│                      │  [下拉] 默认便签排序：按创建时间     │
│                      │  [下拉] 工作台便签墙列宽：自适应     │
│                      │                                    │
└──────────────────────┴───────────────────────────────────┘
```

### 12.3 便签统计

#### API
```typescript
// 调用 GET /api/v1/notes/stats（或复用 notes 接口计算）
interface NoteStats {
  active: number
  completed: number
  reminded: number
  archived: number
  weekly_trend: number[]  // [7] 周一至周日每日数量
}
```

#### 统计卡片
- 4个统计数字卡片，使用 DaisyUI `stats` 组件
- 便签趋势：使用简易柱状图（CSS绘制或轻量图表库）

### 12.4 偏好设置

- 使用 DaisyUI `form-control` + `toggle` 开关组件
- 设置项存储至 localStorage，key: `user_preferences`
- 桌面端同步提醒/音效/强提醒等开关
- 默认便签排序方式选择
- 工作台列宽偏好

```typescript
interface UserPreferences {
  desktopSyncNotify: boolean
  completeSound: boolean
  urgentAlert: boolean
  defaultSort: 'created_at' | 'updated_at' | 'priority'
  gridWidth: 'auto' | 'compact' | 'wide'
}
```

### 12.5 退出登录

- 点击"退出登录" → 确认对话框
- 清除 AuthStore + localStorage
- 跳转至 `/login`
- 清除所有本地状态

## 验收标准
1. 用户基本信息正确展示
2. 便签统计数据展示正确
3. 简易柱状图渲染正常
4. 偏好设置开关正常工作
5. 设置持久化到 localStorage
6. 退出登录流程完整

## 预计工时：2小时

## 交付物
- `src/pages/ProfilePage.vue`
- `src/components/profile/StatsCard.vue`
- `src/components/profile/WeeklyChart.vue`
- `src/components/profile/PreferencePanel.vue`
- `src/stores/preferences.ts`
