<script setup lang="ts">
import { useDarkMode } from '@/composables/useDarkMode'
import { AppIcon } from '@/components/icons'

const { isDark, toggle } = useDarkMode()
</script>

<template>
  <button
    type="button"
    :aria-label="isDark ? '切换到日间模式' : '切换到夜间模式'"
    v-tooltip="isDark ? '切换到日间模式' : '切换到夜间模式'"
    role="switch"
    :aria-checked="isDark"
    class="relative w-10 h-10 rounded-xl flex items-center justify-center transition-all duration-300 hover:scale-110 active:scale-95 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    :class="isDark ? 'bg-slate-700 text-amber-400 hover:bg-slate-600' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'"
    @click="toggle($event)"
  >
    <!-- 太阳/月亮：solid 填色版，切换时 250ms 交叉淡入淡出 -->
    <AppIcon :name="isDark ? 'moon' : 'sun'" variant="solid" :size="20" enter />
  </button>
</template>
