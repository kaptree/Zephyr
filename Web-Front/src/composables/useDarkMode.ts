import { ref, watch, onMounted } from 'vue'

const IS_DARK_KEY = 'labelpro_dark_mode'

const isDark = ref(false)

function applyTheme(dark: boolean) {
  const root = document.documentElement
  if (dark) {
    root.classList.add('dark')
    root.setAttribute('data-theme', 'dark')
  } else {
    root.classList.remove('dark')
    root.setAttribute('data-theme', 'light')
  }
}

export function useDarkMode() {
  onMounted(() => {
    const stored = localStorage.getItem(IS_DARK_KEY)
    if (stored !== null) {
      isDark.value = stored === 'true'
    } else {
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    applyTheme(isDark.value)
  })

  watch(isDark, (val) => {
    localStorage.setItem(IS_DARK_KEY, String(val))
    applyTheme(val)
  })

  function toggle(ev?: MouseEvent) {
    const next = !isDark.value
    const doc = document as Document & {
      startViewTransition?: (cb: () => void) => { ready: Promise<void> }
    }

    // 降级路径：浏览器不支持 View Transition 或用户偏好减少动效时直接切换
    if (!doc.startViewTransition || window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      isDark.value = next
      return
    }

    // 圆形扩散过渡：从点击位置向外扩散新主题（500ms）
    const x = ev?.clientX ?? window.innerWidth / 2
    const y = ev?.clientY ?? 0
    const transition = doc.startViewTransition(() => {
      // 回调内同步应用主题，保证新旧快照正确
      applyTheme(next)
      isDark.value = next // watch 稍后写入 localStorage（applyTheme 幂等）
    })

    transition.ready.then(() => {
      const endRadius = Math.hypot(
        Math.max(x, window.innerWidth - x),
        Math.max(y, window.innerHeight - y),
      )
      document.documentElement.animate(
        {
          clipPath: [
            `circle(0px at ${x}px ${y}px)`,
            `circle(${endRadius}px at ${x}px ${y}px)`,
          ],
        },
        {
          duration: 500,
          easing: 'ease-in-out',
          pseudoElement: '::view-transition-new(root)',
        },
      )
    }).catch(() => {
      /* 过渡被打断时静默忽略 */
    })
  }

  return { isDark, toggle }
}
