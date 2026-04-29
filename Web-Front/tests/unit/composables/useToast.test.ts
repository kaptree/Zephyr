import { describe, it, expect, beforeEach } from 'vitest'
import { useToast } from '@/composables/useToast'

describe('useToast', () => {
  let toast: ReturnType<typeof useToast>

  beforeEach(() => {
    toast = useToast()
    toast.clearAll()
  })

  it('addToast 应添加提示到列表', () => {
    toast.success('操作成功')
    expect(toast.toasts.value).toHaveLength(1)
    expect(toast.toasts.value[0].message).toBe('操作成功')
    expect(toast.toasts.value[0].type).toBe('success')
  })

  it('removeToast 应移除对应提示', () => {
    toast.error('错误信息')
    const id = toast.toasts.value[0].id
    toast.removeToast(id)
    expect(toast.toasts.value).toHaveLength(0)
  })

  it('success 快捷方法应正确工作', () => {
    toast.success('快捷成功')
    expect(toast.toasts.value[0].type).toBe('success')
  })

  it('error 快捷方法应正确工作', () => {
    toast.error('错误信息')
    expect(toast.toasts.value[0].type).toBe('error')
  })

  it('warning 快捷方法应正确工作', () => {
    toast.warning('警告信息')
    expect(toast.toasts.value[0].type).toBe('warning')
  })

  it('info 快捷方法应正确工作', () => {
    toast.info('提示信息')
    expect(toast.toasts.value[0].type).toBe('info')
  })

  it('clearAll 应清除所有提示', () => {
    toast.success('A')
    toast.error('B')
    toast.clearAll()
    expect(toast.toasts.value).toHaveLength(0)
  })
})
