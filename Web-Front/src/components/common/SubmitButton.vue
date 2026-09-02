<script setup lang="ts">
import { computed } from 'vue'

/**
 * 状态机提交按钮（美化工程 · 第 6 项「沉浸式书写」）
 * - 空闲：细微"呼吸光泽"持续微动，吸引点击
 * - 加载：文字变"提交中..." + 旋转 spinner，背景略微变暗，禁用点击
 * - 成功：短暂变绿显示"✓ 成功"，"缩小 → 正常"弹动（500ms）
 * - 失败：变红显示"✗ 失败"，整体抖动 2 次，随后由父级恢复可提交状态
 */
export type SubmitState = 'idle' | 'loading' | 'success' | 'fail'

interface Props {
  state?: SubmitState
  label: string
  loadingLabel?: string
  successLabel?: string
  failLabel?: string
  disabled?: boolean
  type?: 'button' | 'submit'
  variant?: 'primary' | 'secondary'
}

const props = withDefaults(defineProps<Props>(), {
  state: 'idle',
  loadingLabel: '提交中...',
  successLabel: '✓ 成功',
  failLabel: '✗ 失败',
  disabled: false,
  type: 'button',
  variant: 'primary',
})

const emit = defineEmits<{ (e: 'click', event: MouseEvent): void }>()

const busy = computed(() => props.state === 'loading')
const isSuccess = computed(() => props.state === 'success')
const isFail = computed(() => props.state === 'fail')

const btnText = computed(() => {
  if (isSuccess.value) return props.successLabel
  if (isFail.value) return props.failLabel
  if (busy.value) return props.loadingLabel
  return props.label
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || busy"
    class="relative inline-flex items-center justify-center gap-2 rounded-btn px-5 py-2.5 font-medium overflow-hidden transition-smooth active:scale-[0.98]"
    :class="[
      isSuccess
        ? 'bg-green-500 hover:bg-green-500 text-white animate-btn-success-pop'
        : isFail
          ? 'bg-red-500 hover:bg-red-500 text-white animate-btn-fail-shake'
          : busy
            ? 'bg-blue-600/85 dark:bg-blue-500/85 text-white cursor-not-allowed'
            : variant === 'primary'
              ? 'bg-[#3B82F6] hover:bg-blue-600 text-white'
              : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600',
      disabled && !busy ? 'opacity-60 cursor-not-allowed' : '',
    ]"
    @click="emit('click', $event)"
  >
    <!-- 空闲状态：呼吸光泽（柔和白色高光带缓慢扫过） -->
    <span
      v-if="state === 'idle' && !disabled"
      class="submit-shimmer absolute inset-0 pointer-events-none animate-btn-shimmer"
    ></span>
    <!-- 加载状态：旋转 spinner -->
    <svg v-if="busy" class="w-4 h-4 shrink-0 animate-spin" viewBox="0 0 24 24" fill="none">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
      <path class="opacity-90" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
    </svg>
    <span>{{ btnText }}</span>
  </button>
</template>

<style scoped>
/* 呼吸光泽：高光带以 2.5s 周期从左向右扫过（配合 animate-btn-shimmer 移动 background-position） */
.submit-shimmer {
  background: linear-gradient(105deg, transparent 30%, rgba(255, 255, 255, 0.32) 50%, transparent 70%);
  background-size: 200% 100%;
}
</style>
