<script setup lang="ts">
import { EmptyIllustration } from '@/components/icons'
import type { EmptyKind } from '@/components/icons'

defineProps<{
  icon?: string
  message?: string
  /** 品牌插画种类（coffee/cat/inbox）：传入后以 3.2s 呼吸动画插画替代静态图标 */
  illustration?: EmptyKind
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16">
    <!-- 品牌插画：缓慢呼吸动画，明暗主题自适应 -->
    <EmptyIllustration v-if="illustration" :kind="illustration" :size="72" class="mb-6" />
    <div v-else class="w-20 h-20 bg-slate-50 dark:bg-slate-800 rounded-3xl flex items-center justify-center mb-6">
      <svg v-if="!icon" class="w-10 h-10 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
      <span v-else class="text-3xl">{{ icon }}</span>
    </div>
    <p class="text-sm text-slate-400 dark:text-slate-500">{{ message || '暂无数据' }}</p>
  </div>
</template>
