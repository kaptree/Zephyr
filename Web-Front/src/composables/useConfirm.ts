import { ref } from 'vue'

/**
 * 全局确认对话框（美化工程 · 轻量级通知美学）
 * 替代原生 window.confirm：危险操作警告红按钮、取消/确认按钮 200ms 错位依次出现。
 *
 * 用法：
 *   const { confirm } = useConfirm()
 *   if (!(await confirm({ message: '确定删除该标签吗？', danger: true, confirmText: '删除' }))) return
 */
export interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  /** 危险操作：确认按钮使用警告红 */
  danger?: boolean
}

interface ConfirmState {
  visible: boolean
  options: ConfirmOptions
  resolve: ((v: boolean) => void) | null
}

const state = ref<ConfirmState>({
  visible: false,
  options: { message: '' },
  resolve: null,
})

export function useConfirm() {
  function confirm(opts: ConfirmOptions | string): Promise<boolean> {
    const options = typeof opts === 'string' ? { message: opts } : opts
    return new Promise((resolve) => {
      state.value = { visible: true, options, resolve }
    })
  }

  function settle(result: boolean) {
    state.value.resolve?.(result)
    state.value = { visible: false, options: { message: '' }, resolve: null }
  }

  return { state, confirm, settle }
}
