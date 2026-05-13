# 任务11：大屏端适配

## 任务目标

实现应急大屏端页面，包含深色模式、分辨率自适应、只读同步展示。

## 依赖关系

- 依赖任务02（核心基础设施）完成
- 依赖任务05（核心组件）完成

## 技术要求

- 遵循文档第10节大屏端特殊适配
- 支持 1920×1080 至 5760×2160 分辨率
- 只读模式，禁止编辑
- 深色底（`#0F172A`），与Web端白底区分

## 具体步骤

### 11.1 大屏端路由与布局

- 路由路径：`/screen/:id`
- 独立布局：`src/layouts/ScreenLayout.vue`
- 全屏无边框（`100vw × 100vh`）
- 无侧边栏、无顶部通栏

### 11.2 大屏端页面 (`src/pages/ScreenPage.vue`)

#### 布局

```
┌──────────────────────────────────────────────────┐
│  [实时时钟]  轻燕工作台 · 应急指挥大屏  [信号]  │
├──────────────┬──────────────┬────────────────────┤
│              │              │                    │
│   分栏 1     │   分栏 2     │    分栏 3          │
│  (人员任务)  │  (人员任务)  │   (人员任务)       │
│              │              │                    │
│              │              │                    │
├──────────────┴──────────────┴────────────────────┤
│           [系统消息滚动条 / 指令通知]              │
└──────────────────────────────────────────────────┘
```

- 分栏数：4栏或6栏可配（URL参数 `?cols=4`）
- 每栏显示一个人员的任务内容
- CSS Grid 自适应分栏

### 11.3 深色模式样式

```css
/* 大屏端专属样式覆盖 */
.screen-layout {
  background: #0f172a; /* slate-900 */
  color: #e2e8f0; /* slate-200 */
}

/* 任务卡片反色适配 */
.screen-layout .sticky-note-card {
  background: #1e293b; /* slate-800 */
  border-color: #334155; /* slate-700 */
  color: #f1f5f9; /* slate-100 */
}

.screen-layout .sticky-note-card.note-yellow {
  background: #422006; /* 深色黄 */
  border-color: #d97706;
}

.screen-layout .sticky-note-card.note-green {
  background: #052e16; /* 深色绿 */
  border-color: #16a34a;
}

.screen-layout .sticky-note-card.note-red {
  background: #450a0a; /* 深色红 */
  border-color: #dc2626;
}
```

### 11.4 分辨率自适应

```css
/* 使用 vw/vh 单位确保跨分辨率兼容 */
.screen-layout {
  font-size: clamp(12px, 1.2vw, 18px);
}

.screen-layout .sticky-note-card {
  width: 100%;
  height: auto;
  padding: 1.5vw 2vw;
  border-radius: 1vw;
}

.screen-layout .note-title {
  font-size: clamp(14px, 1.4vw, 22px);
}

.screen-layout .note-content {
  font-size: clamp(12px, 1.1vw, 16px);
}
```

### 11.5 实时数据同步

- WebSocket 推送任务更新
- 心跳检测 30s，断线自动重连
- 页面右上角显示连接状态指示器：
  - 绿灯：已连接
  - 黄灯：连接中
  - 红灯：已断开
- 断开时显示"连接中断，正在重连..."浮层

### 11.6 只读模式

- 禁止编辑操作
- 禁止点击交互
- 仅接收同步展示
- 领导指令项：通过独立 WebSocket 频道接收，以醒目样式展示

### 11.7 入场动画

- 页面加载时任务卡片以交错动画入场
- 每个卡片延迟 100ms，依次出现

## 验收标准

1. 深色模式背景正常（`#0F172A`）
2. 任务卡片反色适配正确
3. 分栏布局自适应正常
4. WebSocket 实时同步正常
5. 连接状态指示器正常
6. 只读模式禁止编辑
7. 多分辨率下字体比例合理
8. 入场动画流畅

## 预计工时：3小时

## 交付物

- `src/layouts/ScreenLayout.vue`
- `src/pages/ScreenPage.vue`
- `src/components/screen/ScreenNoteCard.vue`
- `src/components/screen/ConnectionStatus.vue`
- `src/composables/useScreenSync.ts`
