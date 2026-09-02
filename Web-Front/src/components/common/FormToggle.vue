<script setup lang="ts">
/**
 * 弹簧开关（美化工程 · 第 6 项「沉浸式书写」）
 * 滑块以"滑动 + 弹簧弹性"移动（500ms，带轻微回弹），轨道背景色平滑渐变
 */
interface Props {
  modelValue: boolean
  label?: string
  disabled?: boolean
  size?: 'sm' | 'md'
}

const props = withDefaults(defineProps<Props>(), {
  label: '',
  disabled: false,
  size: 'md',
})

const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

function toggle() {
  if (!props.disabled) emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    class="inline-flex items-center gap-2 select-none"
    :class="disabled ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'"
    @click="toggle"
  >
    <span
      class="relative inline-flex shrink-0 rounded-full transition-colors duration-300"
      :class="[
        size === 'sm' ? 'w-8 h-[18px]' : 'w-10 h-[22px]',
        modelValue
          ? 'bg-[#3B82F6] dark:bg-[#60A5FA]'
          : 'bg-slate-300 dark:bg-slate-600',
      ]"
    >
      <span
        class="form-toggle-thumb absolute top-1/2 -translate-y-1/2 left-0.5 rounded-full bg-white shadow-sm"
        :class="[
          size === 'sm' ? 'w-[14px] h-[14px]' : 'w-[18px] h-[18px]',
          modelValue
            ? (size === 'sm' ? 'translate-x-[14px]' : 'translate-x-[18px]')
            : 'translate-x-0',
        ]"
      ></span>
    </span>
    <span v-if="label" class="text-sm text-slate-600 dark:text-slate-300">{{ label }}</span>
  </button>
</template>

<style scoped>
/* 滑块：弹簧弹性移动（500ms，超出后回弹），释放带轻微弹性感 */
.form-toggle-thumb {
  transition: transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  will-change: transform;
}
</style>
