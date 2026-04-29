import { type App } from 'vue'

export function setupErrorHandler(app: App) {
  app.config.errorHandler = (err: unknown, instance, info) => {
    console.error('[全局错误]', err, info)
    showErrorToast('发生未知错误，请刷新页面重试')
    if (import.meta.env.PROD) {
      reportError(err, { component: (instance as { $options?: { name?: string } })?.$options?.name, info })
    }
  }

  window.addEventListener('unhandledrejection', (event: PromiseRejectionEvent) => {
    console.error('[未处理的Promise异常]', event.reason)
    showErrorToast('网络请求异常，请检查网络连接')
  })
}

function showErrorToast(message: string) {
  // 兜底：如果 Toast 组件未加载，使用 alert
  const event = new CustomEvent('toast:error', { detail: { message } })
  window.dispatchEvent(event)
}

function reportError(err: unknown, meta: Record<string, unknown>) {
  console.warn('[ErrorReport]', err, meta)
}
