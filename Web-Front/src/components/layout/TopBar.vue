<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import DarkToggle from '@/components/common/DarkToggle.vue'

const route = useRoute()
const currentTime = ref('')
let timer: ReturnType<typeof setInterval>

const breadcrumbs = computed(() => {
  timer = setInterval(() => {
    currentTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }, 1000)
  currentTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
})

onUnmounted(() => clearInterval(timer))
</script>

<template>
  <header class="h-14 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 flex items-center px-6 shrink-0 gap-3 transition-colors duration-300">
    <!-- 面包屑 -->
    <div class="flex items-center gap-2 text-sm">
      <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
        <span v-if="index > 0" class="text-slate-300 dark:text-slate-600">/</span>
        <span
          :class="[
            'transition-smooth',
            index === breadcrumbs.length - 1
              ? 'text-slate-900 dark:text-slate-100 font-medium'
              : 'text-slate-400 dark:text-slate-500'
          ]"
        >
          {{ crumb.title }}
        </span>
      </template>
    </div>

    <div class="flex-1" />

    <!-- 时间显示 -->
    <span class="text-xs text-slate-400 dark:text-slate-500 tabular-nums font-mono hidden sm:block">{{ currentTime }}</span>

    <!-- 通知按钮 -->
    <button class="relative p-2 rounded-lg text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth">
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
      </svg>
      <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full" />
    </button>

    <!-- 夜间模式切换 -->
    <DarkToggle />
  </header>
</template>
