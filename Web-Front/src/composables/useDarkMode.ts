import { ref, watch, onMounted } from 'vue'

const IS_DARK_KEY = 'labelpro_dark_mode'

const isDark = ref(false)

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

  watch(isDark, (val) => {
    localStorage.setItem(IS_DARK_KEY, String(val))
    applyTheme(val)
  })

  function toggle() {
    isDark.value = !isDark.value
  }

  return { isDark, toggle }
}
