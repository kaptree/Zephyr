<script setup lang="ts">
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
    <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
      <svg class="w-8 h-8 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
      </svg>
    </div>
    <p class="text-sm text-slate-500 mb-4">{{ message || '出错了' }}</p>
    <button
      v-if="action"
      class="btn-primary !text-sm !py-2"
      @click="$emit('retry')"
    >
      {{ action }}
    </button>
  </div>
</template>
