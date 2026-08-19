<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import type { Note } from '@/types';
import { renderNoteContent } from '@/utils/richText';

const props = withDefaults(
  defineProps<{
    note: Note;
    mode?: 'desktop' | 'web';
    archived?: boolean;
    editingBy?: string | null;
  }>(),
  {
    mode: 'web',
    archived: false,
    editingBy: null,
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

const isRed = computed(() => props.note.color_status === 'red');
const isBlue = computed(() => props.note.color_status === 'blue');
const isGreen = computed(() => props.note.color_status === 'green');
const isArchived = computed(() => props.archived || props.note.is_archived);

// 被指派人的反馈填报内容（指派便签上同步展示，指派人与被指派人可见）
const feedbackList = computed(() =>
  (props.note.assignees || [])
    .filter((a: any) => a.feedback_content)
    .map((a: any) => ({
      user_id: a.user_id || a.id || a.user?.id || '',
      user_name: a.user?.name || a.name || a.user?.username || '被指派人',
      content: a.feedback_content,
    }))
);

const displayTags = computed(() => {
  const max = 2;
  const tags = (props.note.tags || []).map((t: any) => {
    if (typeof t === 'string') return { id: t, name: t, sub_tag: '', color: '#64748B' };
    return t;
  });
  const visible = tags.slice(0, max);
  const remaining = tags.length - max;
  return { visible, remaining };
});

function tagLabel(tag: any): string {
  return tag.sub_tag ? `${tag.name} › ${tag.sub_tag}` : tag.name;
}

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

// ===== 左上角倒计时：任务下发（设定了工作时间/截止时间）后实时倒计时 =====
const now = ref(Date.now());
let timer: ReturnType<typeof setInterval> | null = null;
onMounted(() => {
  timer = setInterval(() => {
    now.value = Date.now();
  }, 30000);
});
onUnmounted(() => {
  if (timer) clearInterval(timer);
});

const dueTs = computed(() => {
  if (!props.note.due_time || props.note.completed_at || props.note.is_archived) return 0;
  const t = new Date(props.note.due_time).getTime();
  return Number.isFinite(t) ? t : 0;
});

// 剩余毫秒：>0 未到期，<0 已超时
const remainMs = computed(() => dueTs.value - now.value);

// 倒计时文本（左上角显示）
const countdownText = computed(() => {
  if (!dueTs.value) return '';
  const ms = remainMs.value;
  const abs = Math.abs(ms);
  const totalMin = Math.floor(abs / 60000);
  const days = Math.floor(totalMin / 1440);
  const hours = Math.floor((totalMin % 1440) / 60);
  const mins = totalMin % 60;
  let text: string;
  if (totalMin < 1) text = '不足1分钟';
  else if (days > 0) text = `${days}天${hours}小时`;
  else if (hours > 0) text = `${hours}小时${mins}分钟`;
  else text = `${mins}分钟`;
  return ms < 0 ? `已超时 ${text}` : `剩余 ${text}`;
});

// 是否紧急：已超时或剩余不足1小时
const isDueUrgent = computed(() => {
  if (!dueTs.value) return false;
  return remainMs.value < 3600_000;
});
</script>

<template>
  <div
    class="relative rounded-card p-5 transition-smooth cursor-pointer select-none"
    :class="{
      'opacity-80': isArchived,
      'ring-2 ring-purple-400 ring-offset-2 ring-offset-white dark:ring-offset-slate-950 shadow-lg shadow-purple-200/50':
        !!props.editingBy,
      'bg-red-100 dark:bg-red-900/60 border border-red-200 dark:border-red-900 border-l-4 border-l-red-600 dark:border-l-red-400':
        isRed,
      'bg-blue-100 dark:bg-blue-900/60 border border-blue-200 dark:border-blue-900 border-l-4 border-l-blue-600 dark:border-l-blue-400':
        isBlue,
      'bg-green-100 dark:bg-green-900/60 border border-green-200 dark:border-green-900 border-l-4 border-l-green-600 dark:border-l-green-400':
        isGreen,
      'bg-amber-100 dark:bg-amber-900/60 border border-amber-100 dark:border-amber-900 border-l-4 border-l-amber-600 dark:border-l-amber-400':
        !isRed && !isBlue && !isGreen,
    }"
    :style="{ animation: isRed ? 'pulse-alert 2s ease-in-out infinite' : 'none' }"
    @click="handleClick"
    @contextmenu="handleContextMenu"
    draggable="true"
  >
    <div
      v-if="props.editingBy"
      class="absolute -top-3 left-3 flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-medium animate-fade-in z-10"
      style="
        background: linear-gradient(135deg, #8b5cf6, #3b82f6);
        color: #fff;
        box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
      "
    >
      <span class="w-3.5 h-3.5 rounded-full bg-white/30 flex items-center justify-center text-[8px]"
        >✎</span
      >
      <span>{{ props.editingBy }}</span>
      <span class="inline-block w-1 h-3 bg-white/60 rounded-sm animate-pulse ml-0.5"></span>
    </div>

    <!-- 盯办徽章 -->
    <span v-if="isRed && !isArchived" class="badge-corner bg-red-600 text-white">
      盯办{{ note.remind_count > 0 ? note.remind_count : '' }}
    </span>
    <!-- 协作标识 -->
    <span v-if="isBlue && !isArchived" class="badge-corner bg-blue-500 text-white"> 协作 </span>
    <!-- 已归档水印 -->
    <span v-if="isArchived" class="watermark-archived">已归档</span>

    <!-- 左上角倒计时：设定了工作时间/截止时间的任务实时显示剩余时间 -->
    <span
      v-if="countdownText && !isArchived"
      class="absolute top-2 left-2 px-2 py-0.5 rounded-full text-[11px] font-semibold z-10 animate-fade-in"
      :class="isDueUrgent ? 'bg-red-600 text-white' : 'bg-amber-700 text-white'"
      :title="props.note.due_time"
    >
      ⏱ {{ countdownText }}
    </span>

    <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-2 line-clamp-1">
      {{ note.title || '无标题' }}
    </h3>

    <div
      :class="[
        'text-sm text-slate-500 dark:text-slate-300 transition-all duration-300 overflow-hidden rich-content-display',
        expanded ? 'note-content-expanded' : 'note-content-mask',
        expanded ? '' : 'max-h-[72px]',
      ]"
    >
      <span v-if="!note.content" class="text-slate-300 dark:text-slate-500">暂无内容</span>
      <span v-else v-html="renderNoteContent(note.content)"></span>
    </div>

    <button
      v-if="(note.content?.length || 0) > 100"
      class="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 mt-1 transition-smooth"
      @click.stop="toggleExpand"
    >
      {{ expanded ? '收起' : '展开全文' }}
    </button>

    <!-- 任务反馈区（指派便签同步显示） -->
    <div
      v-if="feedbackList.length"
      class="mt-3 rounded-lg bg-white/70 dark:bg-slate-900/40 p-2.5 border border-green-300/50 dark:border-green-800/60"
    >
      <p
        class="text-[11px] font-semibold text-green-700 dark:text-green-400 mb-1 flex items-center gap-1"
      >
        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        任务反馈
      </p>
      <div
        v-for="f in feedbackList"
        :key="f.user_id"
        class="text-xs text-slate-600 dark:text-slate-300 rich-content-display mb-1 last:mb-0"
      >
        <span class="font-medium text-slate-700 dark:text-slate-200">{{ f.user_name }}：</span>
        <span v-html="renderNoteContent(f.content)" />
      </div>
    </div>

    <!-- 标签区 -->
    <div v-if="(note.tags || []).length" class="flex items-center gap-1.5 mt-3 flex-wrap">
      <span
        v-for="tag in displayTags.visible"
        :key="tag.id"
        class="tag-capsule text-white text-[11px]"
        :style="{ backgroundColor: tag.color || '#64748B' }"
      >
        {{ tagLabel(tag) }}
      </span>
      <span v-if="displayTags.remaining > 0" class="text-xs text-slate-400">
        +{{ displayTags.remaining }}
      </span>
    </div>

    <!-- 底部信息 -->
    <div
      class="flex items-center justify-between mt-4 pt-3 border-t border-slate-200/50 dark:border-slate-700/50"
    >
      <span class="text-xs text-slate-400 dark:text-slate-500">{{
        note.created_at?.slice(0, 10)
      }}</span>
      <span v-if="note.due_time && !isArchived" class="text-xs text-slate-400 dark:text-slate-500">
        截止 {{ note.due_time.slice(0, 10) }}
      </span>
      <span
        v-else-if="isArchived && note.archive_time"
        class="text-xs text-slate-300 dark:text-slate-500"
      >
        归档于 {{ note.archive_time.slice(0, 10) }}
      </span>
    </div>

    <!-- 操作栏：黄/红状态 → 完成并归档 + 盯办 -->
    <div
      v-if="!isArchived"
      class="flex gap-2 mt-3 pt-3 border-t border-slate-200/50 dark:border-slate-700/50"
    >
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-green-100 text-green-700 hover:bg-green-200 dark:bg-green-900/50 dark:text-green-300 dark:hover:bg-green-900 transition-smooth"
        @click.stop="$emit('complete', note)"
      >
        完成并归档
      </button>
      <button
        v-if="!isRed"
        class="text-xs px-2.5 py-1 rounded-btn bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/50 dark:text-red-300 dark:hover:bg-red-900 transition-smooth"
        @click.stop="$emit('remind', note)"
      >
        盯办
      </button>
    </div>

    <div
      v-if="isArchived"
      class="flex gap-2 mt-3 pt-3 border-t border-slate-200/50 dark:border-slate-700/50"
    >
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/50 dark:text-blue-300 dark:hover:bg-blue-900 transition-smooth"
        @click.stop="$emit('restore', note)"
      >
        恢复
      </button>
      <button
        class="text-xs px-2.5 py-1 rounded-btn bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600 transition-smooth"
        @click.stop="$emit('export', note)"
      >
        导出
      </button>
    </div>
  </div>
</template>
