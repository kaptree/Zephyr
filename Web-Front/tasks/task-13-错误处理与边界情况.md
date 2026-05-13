# 任务13：错误处理与边界情况

## 任务目标

实现全局错误处理机制、离线模式支持、协同冲突处理、虚拟滚动优化。

## 依赖关系

- 横向贯穿任务06-09（所有页面）

## 技术要求

- 遵循文档第11节错误处理与边界情况
- 全局异常兜底
- 优雅降级

## 具体步骤

### 13.1 全局错误边界

#### Vue 全局错误处理

```typescript
// src/utils/errorHandler.ts
import { App } from 'vue';

export function setupErrorHandler(app: App) {
  // 捕获未处理的组件错误
  app.config.errorHandler = (err, instance, info) => {
    console.error('[全局错误]', err, info);
    // 显示友好错误提示
    showErrorToast('发生未知错误，请刷新页面重试');
    // 上报错误日志（生产环境）
    if (import.meta.env.PROD) {
      reportError(err, { component: instance?.$options?.name, info });
    }
  };

  // 捕获未处理的 Promise 异常
  window.addEventListener('unhandledrejection', (event) => {
    console.error('[未处理的Promise异常]', event.reason);
    showErrorToast('网络请求异常，请检查网络连接');
  });
}
```

### 13.2 离线模式 (`src/composables/useOffline.ts`)

```typescript
export function useOffline() {
  const isOnline = ref(navigator.onLine);
  const pendingActions = ref<QueuedAction[]>([]);

  // 监听网络状态变化
  watchEffect(() => {
    window.addEventListener('online', () => {
      isOnline.value = true;
      syncPendingActions(); // 批量同步挂起的操作
    });
    window.addEventListener('offline', () => {
      isOnline.value = false;
      showToast('您已离线，操作将在恢复连接后同步', 'warning');
    });
  });

  // 写操作入队
  function enqueueAction(action: QueuedAction) {
    pendingActions.value.push({
      ...action,
      timestamp: Date.now(),
    });
    saveToLocalStorage(pendingActions.value);
  }

  // 恢复后批量同步
  async function syncPendingActions() {
    for (const action of pendingActions.value) {
      try {
        await executeAction(action);
        removeFromQueue(action.id);
      } catch {
        // 该次同步失败，保留在队列中下次重试
      }
    }
  }
}
```

#### 离线提示

- 顶部显示黄色横条提示："您当前处于离线模式"
- 网络恢复时绿色提示："网络已恢复"
- 所有写操作（创建、更新、完成、归档）前端正常响应，但显示"saving..."标记

### 13.3 协同冲突处理 (`src/composables/useCollisionResolver.ts`)

```typescript
export function useCollisionResolver() {
  // 冲突检测
  function detectConflict(localVersion: number, remoteVersion: number): boolean {
    return localVersion !== remoteVersion;
  }

  // 冲突解决策略：Last-Write-Wins + 用户提示
  function resolveConflict(local: string, remote: string, remoteUser: string): ConflictResolution {
    return {
      content: remote, // LWW策略，远程覆盖
      showWarning: true,
      message: `${remoteUser} 刚刚修改了此处，内容已更新`,
      conflict: true,
    };
  }
}
```

### 13.4 虚拟滚动 (`src/composables/useVirtualScroll.ts`)

- 当任务数量 > 100 时自动启用
- 仅渲染可视区域 + 上下缓冲区（各5个卡片）
- 使用 `IntersectionObserver` 或 `@tanstack/vue-virtual`

```typescript
export function useVirtualScroll<T>(items: Ref<T[]>, options: VirtualScrollOptions) {
  const { containerRef, visibleItems, totalHeight, offsetY } = useVirtualizer({
    items,
    itemHeight: 200,
    overscan: 5,
    containerRef,
  });

  return { containerRef, visibleItems, totalHeight, offsetY };
}
```

### 13.5 API 错误处理增强

在 Axios 响应拦截器中：

```typescript
// 统一错误码处理
const ERROR_MAP: Record<number, string> = {
  400: '请求参数错误',
  401: '登录已过期，请重新登录',
  403: '暂无操作权限',
  404: '请求的资源不存在',
  409: '数据冲突，请刷新后重试',
  422: '数据验证失败',
  429: '请求过于频繁，请稍后再试',
  500: '服务器内部错误，请稍后重试',
  502: '网关错误',
  503: '服务暂不可用',
};

// 网络超时：30s 超时后显示"请求超时，请检查网络"
// 网络错误：无 response 时显示"网络连接失败"
```

### 13.6 Toast 通知组件 (`src/components/common/Toast.vue`)

- 4种类型：success（绿）、error（红）、warning（黄）、info（蓝）
- 3秒自动消失
- 可手动关闭
- 堆叠显示（多个时不覆盖）
- 位置：右上角固定

### 13.7 加载状态统一管理

- 骨架屏组件：`src/components/common/Skeleton.vue`
- Loading 遮罩组件：`src/components/common/LoadingOverlay.vue`
- 空态组件：`src/components/common/EmptyState.vue`
- 错误重试组件：`src/components/common/ErrorRetry.vue`

## 验收标准

1. Vue 全局错误处理正确捕获组件异常
2. 离线模式队列写入和恢复同步正常
3. 协同冲突检测和提示正确显示
4. 虚拟滚动在100+任务时正常工作
5. API 错误码提示文案正确
6. Toast 通知类型/样式/消失逻辑正确
7. 骨架屏/Loading/Empty/Error 通用组件可用

## 预计工时：3小时

## 交付物

- `src/utils/errorHandler.ts`
- `src/composables/useOffline.ts`
- `src/composables/useCollisionResolver.ts`
- `src/composables/useVirtualScroll.ts`
- `src/components/common/Toast.vue`
- `src/components/common/Skeleton.vue`
- `src/components/common/LoadingOverlay.vue`
- `src/components/common/EmptyState.vue`
- `src/components/common/ErrorRetry.vue`
