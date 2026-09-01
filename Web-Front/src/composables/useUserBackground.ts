import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { CSSProperties } from 'vue'
import type { BackgroundFill } from '@/types'

// 平台背景图（个人中心设置）→ AdminLayout 内容区背景层样式
// 填充方式映射：cover 适应裁剪 / contain 完整显示 / fill 拉伸填满 / tile 平铺
export function useUserBackground() {
  const auth = useAuthStore()

  const bgStyle = computed<CSSProperties | null>(() => {
    const u = auth.user
    const url = u?.background
    if (!url) return null
    const fill = (u.bg_fill || 'cover') as BackgroundFill
    return {
      backgroundImage: `url(${url})`,
      backgroundSize: fill === 'tile' ? 'auto' : fill === 'fill' ? '100% 100%' : fill,
      backgroundRepeat: fill === 'tile' ? 'repeat' : 'no-repeat',
      backgroundPosition: 'center',
      opacity: String(typeof u.bg_opacity === 'number' ? u.bg_opacity : 1),
    }
  })

  return { bgStyle }
}
