# 任务08：协同编辑室

## 任务目标
实现协同编辑室页面，包括多人同屏协同画布、WebSocket 实时同步、投屏模式。

## 依赖关系
- 依赖任务04（管理端布局）完成
- 依赖任务05（CollaborationCanvas 组件）完成

## 技术要求
- 遵循文档第9.2节协同编辑室设计
- 遵循文档第7.2.4节协同与模板接口
- 使用 Socket.io-client 实现实时通信

## 具体步骤

### 8.1 页面路由与权限
- 路由路径：`/workbench/collaboration/:id`
- 进入时检测权限（`beforeEnter` 守卫）：非成员跳转 403

### 8.2 页面结构 (`src/pages/CollaborationPage.vue`)

```
┌─────────────────────────────────────────────────────┐
│  ← 返回  协同编辑室：[房间名称]    [2] [4] [6] [8]栏 │
│                                               [投屏] │
├─────────────────────────────────────────────────────┤
│                                                       │
│  ┌─────────────┐  ┌─────────────┐                    │
│  │ 张三 (组长)  │  │ 李四 (组员)  │  ...              │
│  │ 🟢 在线      │  │ 🟢 在线      │                    │
│  ├─────────────┤  ├─────────────┤                    │
│  │             │  │             │                    │
│  │  富文本编辑  │  │  富文本编辑  │                    │
│  │  区域       │  │  区域       │                    │
│  │             │  │             │                    │
│  └─────────────┘  └─────────────┘                    │
│                                                       │
├─────────────────────────────────────────────────────┤
│  李四正在输入...                        [下发指令 ▼]  │
└─────────────────────────────────────────────────────┘
```

### 8.3 WebSocket 连接管理

#### 连接
- 页面加载时连接 Socket.io：`io(\`${WS_URL}/ws/notes/${roomId}\`)`
- 携带认证 token 进行握手
- 连接成功后加入 room

#### 事件处理
| 事件 | 方向 | 处理逻辑 |
|------|------|---------|
| `canvas:update` | 发送 | 本地编辑器 onChange → 节流300ms → emit |
| `canvas:sync` | 接收 | 非本栏更新 → 合并至对应栏位 |
| `participant:join` | 接收 | 新增参与者信息条 |
| `participant:leave` | 接收 | 参与者变灰/离线 |
| `command:broadcast` | 接收 | 领导下发的指令 → 红色边框消息条插入所有栏位顶部 |
| `typing:start` | 发送 | 本地开始输入 → emit |
| `typing:stop` | 发送 | 本地停止输入 → emit |
| `typing:status` | 接收 | 更新"XX正在输入..." |

#### 断线重连
- Socket.io 自动重连机制
- 连接中断时显示"连接断开，正在重连..."浮层
- 重连成功后增量同步数据

### 8.4 分栏编辑区

- 复用 CollaborationCanvas 组件
- 每栏独立富文本区（使用 TipTap 编辑器）
- 菜单栏悬浮（加粗、列表、链接）
- 仅当前用户栏位可编辑，他人栏位只读（灰色背景）
- 栏位数切换（2/4/6/8），通过按钮切换

### 8.5 领导下发指令

- 底部固定区域：输入框 + "下发指令" 按钮
- 点击后：emit `command:broadcast`，内容以红色边框系统消息形式插入所有栏位顶部
- 仅组长/管理员可见此功能

### 8.6 投屏模式

- 点击"投屏"按钮 → 全屏无UI模式
- URL 追加 `?mode=projection`
- 隐藏所有控件（侧边栏、顶栏、工具栏）
- 仅展示画布内容
- ESC 或双击退出投屏模式

### 8.7 CollaborationStore 完善

```typescript
// 完整实现：
joinRoom(roomId: string): void
leaveRoom(): void
pushLocalChange(columnId: string, content: string): void
handleRemoteChange(data: RemoteChangeData): void
handleParticipantJoin(user: UserBrief): void
handleParticipantLeave(userId: string): void
handleCommand(message: string, from: string): void
updateTypingStatus(userId: string, isTyping: boolean): void
```

### 8.8 并发冲突处理

- 采用 Last-Write-Wins 策略
- 冲突时显示 Toast："XX刚刚修改了此处，内容已更新"
- 提供"查看历史版本"入口

## 验收标准
1. WebSocket 连接/断开正常
2. 分栏编辑区域正确渲染
3. 多栏同时编辑实时同步
4. "正在输入..."状态实时更新
5. 领导下发指令功能正常（红色边框消息条）
6. 投屏模式正确隐藏所有控件
7. 断线重连后数据恢复
8. 栏位数切换正常
9. 非成员无法进入

## 预计工时：5小时

## 交付物
- `src/pages/CollaborationPage.vue`
- 完善 `src/components/note/CollaborationCanvas.vue`
- 完善 `src/stores/collaboration.ts`
- `src/composables/useSocket.ts`
- `src/components/note/CommandInput.vue`
