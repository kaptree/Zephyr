import { describe, expect, it, vi } from 'vitest'
import { useChatDropUpload } from '@/composables/useChatDropUpload'

// jsdom 未实现 DragEvent，用普通对象模拟（组合式函数只读取 dataTransfer.types / files）
function fakeDragEvent(opts: { types?: string[]; files?: File[] } = {}): DragEvent {
  return {
    dataTransfer: {
      types: opts.types || [],
      files: (opts.files || []) as unknown as FileList,
    },
  } as unknown as DragEvent
}

function makeFile(name: string, size = 1024, type = 'text/plain') {
  return new File(['x'.repeat(size)], name, { type })
}

describe('useChatDropUpload', () => {
  it('非文件拖拽（如选中文本）不触发高亮，drop 也不回调', () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragEnter, onDragLeave, onDrop } = useChatDropUpload(handleFiles)

    onDragEnter(fakeDragEvent({ types: ['text/plain'] }))
    expect(dragOver.value).toBe(false)

    onDrop(fakeDragEvent({ types: ['text/plain'] }))
    expect(handleFiles).not.toHaveBeenCalled()
  })

  it('拖入文件时 dragOver 为 true，离开后恢复 false', () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragEnter, onDragLeave } = useChatDropUpload(handleFiles)

    onDragEnter(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(true)

    onDragLeave(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(false)
  })

  it('深度计数：经过子元素（成对 enter/leave）高亮不闪烁', () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragEnter, onDragLeave } = useChatDropUpload(handleFiles)

    // 进入主区 → 进入子元素 → 离开子元素（进入子元素会先触发一次主区 leave）
    onDragEnter(fakeDragEvent({ types: ['Files'] }))
    onDragEnter(fakeDragEvent({ types: ['Files'] }))
    onDragLeave(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(true) // 仍在主区内

    onDragLeave(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(false) // 全部离开
  })

  it('深度计数不会变为负数', () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragLeave } = useChatDropUpload(handleFiles)

    onDragLeave(fakeDragEvent({ types: ['Files'] }))
    onDragLeave(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(false)
  })

  it('drop 提取全部文件并回调，同时复位高亮', async () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragEnter, onDrop } = useChatDropUpload(handleFiles)

    onDragEnter(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(true)

    const files = [makeFile('a.pdf'), makeFile('b.png')]
    onDrop(fakeDragEvent({ types: ['Files'], files }))
    expect(dragOver.value).toBe(false)
    expect(handleFiles).toHaveBeenCalledTimes(1)
    expect(handleFiles).toHaveBeenCalledWith(files)
  })

  it('drop 无文件（如拖入目录或无效内容）不回调', () => {
    const handleFiles = vi.fn()
    const { onDrop } = useChatDropUpload(handleFiles)

    onDrop(fakeDragEvent({ types: ['Files'], files: [] }))
    expect(handleFiles).not.toHaveBeenCalled()
  })

  it('reset 手动复位高亮', () => {
    const handleFiles = vi.fn()
    const { dragOver, onDragEnter, reset } = useChatDropUpload(handleFiles)

    onDragEnter(fakeDragEvent({ types: ['Files'] }))
    expect(dragOver.value).toBe(true)

    reset()
    expect(dragOver.value).toBe(false)
  })
})
