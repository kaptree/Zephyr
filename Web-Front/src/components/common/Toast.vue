<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, removeToast } = useToast()

const typeCls: Record<string, string> = {
  success: 'bg-green-50/95 dark:bg-green-900/80 text-green-700 dark:text-green-200 border border-green-200 dark:border-green-700',
  error: 'bg-red-50/95 dark:bg-red-900/80 text-red-700 dark:text-red-200 border border-red-200 dark:border-red-700',
  warning: 'bg-yellow-50/95 dark:bg-yellow-900/80 text-yellow-700 dark:text-yellow-200 border border-yellow-200 dark:border-yellow-700',
  info: 'bg-blue-50/95 dark:bg-blue-900/80 text-blue-700 dark:text-blue-200 border border-blue-200 dark:border-blue-700',
}

const barCls: Record<string, string> = {
  success: 'bg-green-500',
  error: 'bg-red-500',
  warning: 'bg-yellow-500',
  info: 'bg-blue-500',
}

const icons: Record<string, string> = {
  success: '✓',
  error: '✗',
  warning: '⚠',
  info: 'ℹ',
}
</script>

<template>
  <!-- 底部居中 Toast（与右上角消息弹窗错位），点击可提前关闭 -->
  <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[100] flex flex-col items-center gap-2 pointer-events-none">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="[
          'pointer-events-auto relative overflow-hidden px-4 py-3 rounded-btn shadow-modal text-sm font-medium transition-smooth cursor-pointer flex items-center gap-2 backdrop-blur-sm min-w-[200px]',
          typeCls[toast.type],
        ]"
        @click="removeToast(toast.id)"
      >
        <span>{{ icons[toast.type] }}</span>
        {{ toast.message }}
        <!-- 进度条倒计时：线性走完后通知自动消失 -->
        <span
          class="absolute bottom-0 left-0 h-[3px] rounded-full toast-progress"
          :class="barCls[toast.type]"
          :style="{ animationDuration: `${toast.duration}ms` }"
        />
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
/* 从底部滑入（300ms ease-out），走完/点击后下滑淡出 */
.toast-enter-active {
  transition: all 0.3s cubic-bezier(0, 0, 0.2, 1);
}
.toast-leave-active {
  transition: all 0.2s ease-in;
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(24px) scale(0.96);
}
.toast-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.96);
}
.toast-move {
  transition: transform 0.25s ease;
}
</style>
