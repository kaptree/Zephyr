<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Note } from '@/types';

const props = withDefaults(
  defineProps<{
    note: Note;
    mode?: 'desktop' | 'web';
    archived?: boolean;
  }>(),
  {
    mode: 'web',
    archived: false,
  }
);

const emit = defineEmits<{
  click: [note: Note];
  'context-menu': [event: MouseEvent, note: Note];
  complete: [note: Note];
  remind: [note: Note];
  restore: [note: Note];
  export: [note: Note];
}>();

const expanded = ref(false);

const isUrgent = computed(() => props.note.color_status === 'red');
const isArchived = computed(() => props.archived || props.note.is_archived);

const displayTags = computed(() => {
  const max = 2;
  const tags = props.note.tags || [];
  const visible = tags.slice(0, max);
  const remaining = tags.length - max;
  return { visible, remaining };
});

function handleClick() {
  emit('click', props.note);
}

function handleContextMenu(e: MouseEvent) {
  e.preventDefault();
  emit('context-menu', e, props.note);
}

function toggleExpand() {
  expanded.value = !expanded.value;
}
</script>

<template>
  <div
    class="relative rounded-card p-5 transition-smooth cursor-pointer select-none"
    :class="{ 'opacity-80': isArchived }"
    :style="{
      background: isUrgent ? '#FEE2E2' : '#FEF3C7',
      borderLeft: isUrgent ? '1px solid #DC2626' : '4px solid #D97706',
      border: isUrgent ? '1px solid #DC2626' : '',
      animation: isUrgent ? 'pulse-alert 2s ease-in-out infinite' : 'none',
    }"
    @click="handleClick"
    @contextmenu="handleContextMenu"
    draggable="true"
  >
    <!-- 盯办徽章 -->
    <span v-if="isUrgent && !isArchived" class="badge-corner bg-red-500 text-white">
      盯办{{ note.remind_count > 0 ? note.remind_count : '' }}
    </span>
    <!-- 已归档水印 -->
    <span v-if="isArchived" class="watermark-archived">已归档</span>

    <h3 class="text-base font-semibold text-slate-900 mb-2 line-clamp-1">
      {{ note.title || '无标题' }}
    </h3>

    <div
      :class="[
        'text-sm text-slate-500 transition-all duration-300 overflow-hidden',
        expanded ? 'note-content-expanded' : 'note-content-mask',
        expanded ? '' : 'max-h-[72px]',
      ]"
    >
      {{ note.content || '暂无内容' }}
    </div>

    <button
      v-if="(note.content?.length || 0) > 100"
      class="text-xs text-slate-400 hover:text-slate-600 mt-1 transition-smooth"
      @click.stop="toggleExpand"
    >
      {{ expanded ? '收起' : '展开全文' }}
    </button>

    <!-- 标签区 -->
    <div v-if="(note.tags || []).length" class="flex items-center gap-1.5 mt-3 flex-wrap">
      <span
        v-for="tag in displayTags.visible"
        :key="tag.id"
        class="tag-capsule text-white text-[11px]"
        :style="{ backgroundColor: tag.color || '#64748B' }"
      >
        {{ tag.name }}
      </span>
      <span v-if="displayTags.remaining > 0" class="text-xs text-slate-400">
        +{{ displayTags.remaining }}
      </span>
    </div>

    <!-- 底部信息 -->
    <div class="flex items-center justify-between mt-4 pt-3 border-t border-slate-200/50">
      <span class="text-xs text-slate-400">{{ note.created_at?.slice(0, 10) }}</span>
      <span v-if="note.due_time && !isArchived" class="text-xs text-slate-400">
        截止 {{ note.due_time.slice(0, 10) }}
      </span>
      <span v-else-if="isArchived && note.archive_time" class="text-xs text-slate-300">
        归档于 {{ note.archive_time.slice(0, 10) }}
      </span>
    </div>

    <!-- 操作栏：黄/红状态 → 完成并归档 + 盯办 -->
    <div v-if="!isArchived" class="flex gap-2 mt-3 pt-3 border-t border-slate-200/50">
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-green-100 text-green-700 hover:bg-green-200 transition-smooth"
        @click.stop="$emit('complete', note)"
      >
        完成并归档
      </button>
      <button
        v-if="!isUrgent"
        class="text-xs px-2.5 py-1 rounded-btn bg-red-100 text-red-700 hover:bg-red-200 transition-smooth"
        @click.stop="$emit('remind', note)"
      >
        盯办
      </button>
    </div>

    <div v-if="isArchived" class="flex gap-2 mt-3 pt-3 border-t border-slate-200/50">
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-blue-100 text-blue-700 hover:bg-blue-200 transition-smooth"
        @click.stop="$emit('restore', note)"
      >
        恢复
      </button>
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-slate-100 text-slate-600 hover:bg-slate-200 transition-smooth"
        @click.stop="$emit('export', note)"
      >
        导出
      </button>
    </div>
  </div>
</template>
