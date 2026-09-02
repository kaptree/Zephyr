<script setup lang="ts">
import { computed } from 'vue'

/**
 * 绘制勾号复选框（美化工程 · 第 6 项「沉浸式书写」）
 * 选中时勾号以 stroke-dasharray"绘制勾号"动画呈现（400ms），
 * 背景色从透明平滑过渡到主题色（200ms）
 * 支持两种用法：布尔 v-model，或配合 value 的数组多选 v-model
 */
interface Props {
  modelValue: boolean | (string | number)[]
  /** 数组多选时的选项值 */
  value?: string | number
  label?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  value: undefined,
  label: '',
  disabled: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean | (string | number)[]): void
}>()

const checked = computed(() =>
  Array.isArray(props.modelValue)
    ? props.value !== undefined && props.modelValue.includes(props.value)
    : !!props.modelValue
)

function toggle() {
  if (props.disabled) return
  if (Array.isArray(props.modelValue) && props.value !== undefined) {
    const arr = [...props.modelValue]
    const i = arr.indexOf(props.value)
    if (i >= 0) arr.splice(i, 1)
    else arr.push(props.value)
    emit('update:modelValue', arr)
  } else {
    emit('update:modelValue', !props.modelValue)
  }
}
</script>

<template>
  <button
    type="button"
    role="checkbox"
    :aria-checked="checked"
    :disabled="disabled"
    class="inline-flex items-center gap-2 select-none"
    :class="disabled ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'"
    @click="toggle"
  >
    <span
      class="grid place-items-center w-5 h-5 shrink-0 rounded-md border-2 transition-colors duration-200"
      :class="checked
        ? 'bg-[#3B82F6] border-[#3B82F6] dark:bg-[#60A5FA] dark:border-[#60A5FA]'
        : 'bg-transparent border-slate-300 dark:border-slate-500'"
    >
      <svg class="w-3.5 h-3.5 text-white" viewBox="0 0 24 24" fill="none">
        <polyline
          points="5 13 10 18 19 7"
          stroke="currentColor"
          stroke-width="3"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="form-checkbox-check"
          :class="{ 'on': checked }"
        />
      </svg>
    </span>
    <span v-if="label || $slots.default" class="text-sm text-slate-600 dark:text-slate-300">
      <slot>{{ label }}</slot>
    </span>
  </button>
</template>

<style scoped>
/* 勾号：未选中时隐藏（dashoffset = 路径长度），选中后"绘制"呈现（400ms） */
.form-checkbox-check {
  stroke-dasharray: 26;
  stroke-dashoffset: 26;
  opacity: 0;
  transition: opacity 0.15s ease;
}
.form-checkbox-check.on {
  opacity: 1;
  animation: ff-check-draw 0.4s ease-out forwards;
}
@keyframes ff-check-draw {
  from { stroke-dashoffset: 26; }
  to { stroke-dashoffset: 0; }
}
</style>
