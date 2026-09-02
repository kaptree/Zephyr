<script setup lang="ts">
import { computed, getCurrentInstance, ref, watch } from 'vue'

/**
 * 浮动标签输入框（美化工程 · 第 6 项「沉浸式书写」）
 * - 聚焦或输入后标签上浮至输入框左上角（200ms），颜色从灰变主色
 * - 底部下划线从中心向两端展开（0 → 100%，300ms）
 * - 校验通过：右侧绿色对勾"弹出 + 旋转"（300ms），边框短暂闪烁绿色光晕
 * - 校验失败：整框"水平抖动"（400ms 振幅 5px），错误提示"滑入 + 淡入"（250ms），
 *   修正后"收缩 + 淡出"（200ms）优雅消失
 */
interface Props {
  modelValue: string
  label: string
  type?: string
  placeholder?: string
  /** 校验错误信息：非空时抖动 + 红色提示 */
  error?: string
  /** 外部校验通过时为 true：显示绿色对勾 */
  success?: boolean
  disabled?: boolean
  autocomplete?: string
  maxlength?: number
  inputmode?: 'text' | 'numeric' | 'tel' | 'email' | 'decimal'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  placeholder: '',
  error: '',
  success: false,
  disabled: false,
  autocomplete: 'off',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'focus', event: FocusEvent): void
  (e: 'blur', event: FocusEvent): void
}>()

const inputId = `ff-field-${getCurrentInstance()?.uid ?? Math.random().toString(36).slice(2, 8)}`

const focused = ref(false)
const flashSuccess = ref(false)

const hasValue = computed(() => String(props.modelValue ?? '').length > 0)
const floatOn = computed(() => focused.value || hasValue.value)
const showSuccess = computed(() => props.success && hasValue.value && !props.error)

// 校验通过瞬间：边框闪烁一次绿色光晕
watch(showSuccess, (val) => {
  if (!val) return
  flashSuccess.value = true
  setTimeout(() => { flashSuccess.value = false }, 650)
})

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="relative">
    <!-- 校验失败：整框水平抖动 -->
    <div :class="{ 'animate-shake-x': !!error }">
      <div class="relative">
        <input
          :id="inputId"
          :type="type"
          :value="modelValue"
          :disabled="disabled"
          :autocomplete="autocomplete"
          :maxlength="maxlength"
          :inputmode="inputmode"
          class="float-input w-full rounded-btn border bg-white dark:bg-slate-800 px-4 pt-5 pb-1.5 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 disabled:opacity-60 disabled:cursor-not-allowed outline-none transition-colors duration-200"
          :class="[
            error
              ? 'err border-red-400 dark:border-red-500'
              : 'border-slate-200 dark:border-slate-600',
            flashSuccess ? 'flash-green' : '',
          ]"
          :placeholder="placeholder"
          @input="onInput"
          @focus="focused = true; emit('focus', $event)"
          @blur="focused = false; emit('blur', $event)"
        />
        <label
          :for="inputId"
          class="float-label"
          :class="[
            floatOn ? 'on' : '',
            error
              ? 'text-red-500 dark:text-red-400'
              : floatOn
                ? 'text-[#3B82F6] dark:text-[#60A5FA]'
                : 'text-slate-400 dark:text-slate-500',
          ]"
        >{{ label }}</label>
        <!-- 底部下划线：聚焦/输入后从中心向两端展开（300ms） -->
        <span class="float-underline" :class="{ 'on': floatOn }"></span>
        <!-- 校验通过：绿色对勾"弹出 + 旋转"（300ms） -->
        <span
          v-if="showSuccess"
          class="absolute right-3 top-1/2 -translate-y-1/2 grid place-items-center w-5 h-5 rounded-full bg-green-500 shadow-sm animate-check-pop"
        >
          <svg class="w-3 h-3 text-white" viewBox="0 0 12 12" fill="none">
            <path d="M2.5 6.5L5 9l4.5-6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </span>
      </div>
    </div>
    <!-- 错误提示：滑入 + 淡入 250ms；修正后收缩 + 淡出 200ms -->
    <transition name="error-slide">
      <p v-if="error" class="mt-1 flex items-center gap-1 text-xs text-red-500 dark:text-red-400">
        <svg class="w-3.5 h-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M18 10A8 8 0 112 10a8 8 0 0116 0zm-9-4a1 1 0 112 0v5a1 1 0 11-2 0V6zm1 8a1.25 1.25 0 100-2.5A1.25 1.25 0 0010 14z" clip-rule="evenodd" />
        </svg>
        {{ error }}
      </p>
    </transition>
  </div>
</template>

<style scoped>
/* 标签：默认垂直居中占位，聚焦/输入后上浮至左上角（200ms 平滑过渡） */
.float-label {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 14px;
  line-height: 1.2;
  pointer-events: none;
  transition:
    top 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    font-size 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.2s ease-out;
}
.float-label.on {
  top: 6px;
  transform: translateY(0);
  font-size: 11px;
}

/* 聚焦柔光：200ms 光晕扩散 */
.float-input:focus {
  border-color: #3B82F6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.12);
}
.dark .float-input:focus {
  border-color: #60A5FA;
  box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.16);
}
/* 错误态聚焦保持警示红 */
.float-input.err:focus {
  border-color: #ef4444;
  box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.12);
}
.dark .float-input.err:focus {
  border-color: #f87171;
}

/* 校验通过瞬间：绿色光晕闪烁一次 */
.float-input.flash-green {
  animation: ff-flash-green 0.65s ease-out;
}
@keyframes ff-flash-green {
  0% {
    border-color: #16A34A;
    box-shadow: 0 0 0 0 rgba(22, 163, 74, 0.45);
  }
  100% {
    border-color: rgba(22, 163, 74, 0.55);
    box-shadow: 0 0 0 7px rgba(22, 163, 74, 0);
  }
}

/* 底部下划线：从中心向两端展开 0 → 100%（300ms） */
.float-underline {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  height: 2px;
  width: 0;
  border-radius: 9999px;
  background: #3B82F6;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  pointer-events: none;
}
.dark .float-underline {
  background: #60A5FA;
}
.float-underline.on {
  width: calc(100% - 2px);
}
</style>
