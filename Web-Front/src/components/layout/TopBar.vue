<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(r => r.meta.title)
  return matched.map(r => ({
    title: r.meta.title as string,
    path: r.path,
  }))
})
</script>

<template>
  <header class="h-14 bg-white border-b border-slate-200 flex items-center px-6 shrink-0 gap-3">
    <!-- 面包屑 -->
    <div class="flex items-center gap-2 text-sm">
      <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
        <span v-if="index > 0" class="text-slate-300">/</span>
        <span
          :class="[
            'transition-smooth',
            index === breadcrumbs.length - 1
              ? 'text-slate-900 font-medium'
              : 'text-slate-400'
          ]"
        >
          {{ crumb.title }}
        </span>
      </template>
    </div>

    <!-- 演示模式标识 -->
    <div
      v-if="auth.demoMode"
      class="flex items-center gap-1.5 px-2.5 py-1 bg-amber-50 border border-amber-200 rounded-tag text-xs text-amber-700 font-medium"
    >
      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
      </svg>
      演示模式
    </div>

    <div class="flex-1" />

    <!-- 通知按钮 -->
    <button class="relative p-2 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-smooth">
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
      </svg>
      <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full" />
    </button>
  </header>
</template>
