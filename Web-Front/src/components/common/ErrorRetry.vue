<script setup lang="ts">
import { AppIcon } from '@/components/icons'

defineProps<{
  message?: string
  action?: string
}>()

defineEmits<{
  retry: []
}>()
</script>

<template>
  <div class="text-center py-12">
    <div class="w-16 h-16 bg-red-50 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
      <!-- 警告三角：快速闪烁 3 次（0.9s × 3）后停止，语义红 -->
      <AppIcon name="warning" :size="28" class="animate-icon-flash-warn text-red-500 dark:text-red-400" />
    </div>
    <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">{{ message || '出错了' }}</p>
    <button
      v-if="action"
      class="btn-primary !text-sm !py-2"
      @click="$emit('retry')"
    >
      {{ action }}
    </button>
  </div>
</template>
