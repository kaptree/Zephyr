<script setup lang="ts">
import { ref, watch } from 'vue';
import RichTextEditor from '@/components/common/RichTextEditor.vue';
import type { Note } from '@/types';

const props = defineProps<{
  visible: boolean;
  note: Note | null;
  mode?: 'complete' | 'feedback';
}>();

const emit = defineEmits<{
  'update:visible': [value: boolean];
  submit: [content: string];
}>();

const content = ref('');
const submitting = ref(false);

watch(
  () => props.visible,
  (v) => {
    if (v) {
      content.value = '';
      submitting.value = false;
    }
  }
);

function cancel() {
  emit('update:visible', false);
}

async function submit() {
  if (!content.value.trim()) return;
  submitting.value = true;
  try {
    emit('submit', content.value);
    emit('update:visible', false);
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div
    v-if="visible && note"
    class="fixed inset-0 z-[80] flex items-center justify-center bg-black/40 p-4"
    @click.self="cancel"
  >
    <div class="w-full max-w-2xl bg-white dark:bg-slate-800 rounded-xl shadow-2xl overflow-hidden">
      <div class="flex items-center justify-between px-5 py-4 border-b border-slate-100 dark:border-slate-700">
        <div>
          <h3 class="text-base font-semibold text-slate-800 dark:text-slate-100">
            {{ mode === 'feedback' ? '补充反馈填报' : '任务反馈填报' }}
          </h3>
          <p class="mt-0.5 text-xs text-slate-400 truncate max-w-md">{{ note.title }}</p>
        </div>
        <button
          class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
          @click="cancel"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="px-5 py-4">
        <p class="mb-2 text-sm text-slate-500 dark:text-slate-400">
          请填写本次任务的工作成果与完成情况，提交后将通知任务发起人。
        </p>
        <RichTextEditor v-model="content" :min-height="200" placeholder="请输入反馈内容..." />
      </div>

      <div class="flex justify-end gap-2 px-5 py-3 border-t border-slate-100 dark:border-slate-700">
        <button
          class="px-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth"
          @click="cancel"
        >
          取消
        </button>
        <button
          class="px-4 py-2 text-sm rounded-lg bg-blue-500 text-white font-medium hover:bg-blue-600 transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="!content.trim() || submitting"
          @click="submit"
        >
          {{ submitting ? '提交中...' : mode === 'feedback' ? '提交反馈' : '提交反馈并完成' }}
        </button>
      </div>
    </div>
  </div>
</template>
