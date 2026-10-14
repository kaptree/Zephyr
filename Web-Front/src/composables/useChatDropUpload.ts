import { ref, computed } from 'vue'

/**
 * 聊天窗口拖拽上传状态机：
 * - dragenter/dragleave 在经过子元素时会成对触发，用深度计数避免高亮闪烁
 * - 仅响应文件拖拽（忽略页面内文本 / 元素拖动）
 * - drop 时提取文件列表并回调，由调用方负责上传发送
 */
export function useChatDropUpload(handleFiles: (files: File[]) => void) {
  const dragDepth = ref(0)
  const dragOver = computed(() => dragDepth.value > 0)

  function isFileDrag(e: DragEvent) {
    return Array.from(e.dataTransfer?.types || []).includes('Files')
  }

  function onDragEnter(e: DragEvent) {
    if (isFileDrag(e)) dragDepth.value += 1
  }

  function onDragLeave(e: DragEvent) {
    if (isFileDrag(e)) dragDepth.value = Math.max(0, dragDepth.value - 1)
  }

  function onDrop(e: DragEvent) {
    dragDepth.value = 0
    const files = Array.from(e.dataTransfer?.files || [])
    if (files.length) handleFiles(files)
  }

  function reset() {
    dragDepth.value = 0
  }

  return { dragOver, onDragEnter, onDragLeave, onDrop, reset }
}
