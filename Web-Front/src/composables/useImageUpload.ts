import { ref } from 'vue'
import { uploadImage } from '@/services/admin'
import { useToast } from '@/composables/useToast'

/** 允许上传的图片扩展名（含 bmp） */
const IMAGE_EXT_RE = /\.(png|jpe?g|gif|webp|bmp)$/i

/**
 * 通用「图片上传」能力：任务 / 反馈 / Issue 内嵌图片共用。
 * 动态创建文件选择框 → 校验格式 → 上传 → 成功后回调 onInsert(url) 由调用方插入富文本/文本。
 * 大小限制由后端复用聊天图片限制，超限错误文案统一含「请联系管理员修改聊天文件大小限制」。
 */
export function useImageUpload() {
  const toast = useToast()
  const uploading = ref(false)

  /** 打开文件选择框；选中后上传，成功回调 onInsert(图片URL) */
  function pickImage(onInsert: (url: string) => void) {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/png,image/jpeg,image/gif,image/webp,image/bmp'
    input.onchange = () => {
      const file = input.files?.[0]
      input.value = ''
      if (file) upload(file, onInsert)
    }
    input.click()
  }

  async function upload(file: File, onInsert: (url: string) => void) {
    if (uploading.value) return
    if (!file.type.startsWith('image/') && !IMAGE_EXT_RE.test(file.name)) {
      toast.warning('仅支持 png/jpg/jpeg/gif/webp/bmp 格式图片')
      return
    }
    uploading.value = true
    try {
      const res = await uploadImage(file)
      const url = (res.data as { url: string }).url
      if (url) onInsert(url)
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } }; friendlyMessage?: string }
      toast.error(err?.response?.data?.message || err?.friendlyMessage || '图片上传失败')
    } finally {
      uploading.value = false
    }
  }

  return { uploading, pickImage }
}
