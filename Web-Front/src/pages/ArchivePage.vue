<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useToast } from '@/composables/useToast';
import { useNoteStore } from '@/stores/notes';
import type { Note } from '@/types';
import StickyNoteCard from '@/components/note/StickyNoteCard.vue';
import DateRangePicker from '@/components/common/DateRangePicker.vue';
import FeedbackModal from '@/components/notification/FeedbackModal.vue';
import { renderNoteContent } from '@/utils/richText';
import { submitFeedback } from '@/services/notes';

const noteStore = useNoteStore();
const toast = useToast();
const viewMode = ref<'timeline' | 'card'>('timeline'); // Bug8：默认时间轴
const keyword = ref('');
const dateFrom = ref('');
const dateTo = ref('');
const showDetailPanel = ref(false);
const selectedNote = ref<Note | null>(null);
const restoring = ref(false);
const feedbackVisible = ref(false);
const feedbackNote = ref<Note | null>(null);
const submittingFeedback = ref(false);

onMounted(() => {
  noteStore.fetchArchivedNotes({});
});

function handleSearch() {
  noteStore.fetchArchivedNotes({
    keyword: keyword.value || undefined,
    date_from: dateFrom.value || undefined,
    date_to: dateTo.value || undefined,
  });
}

function handleClear() {
  keyword.value = '';
  dateFrom.value = '';
  dateTo.value = '';
  noteStore.fetchArchivedNotes({});
}

function openDetail(note: Note) {
  selectedNote.value = note;
  showDetailPanel.value = true;
}

function closeDetail() {
  showDetailPanel.value = false;
  selectedNote.value = null;
}

async function handleRestore(note: Note) {
  restoring.value = true;
  try {
    await noteStore.restoreNote(note.id);
    if (showDetailPanel.value && selectedNote.value?.id === note.id) {
      closeDetail();
    }
  } catch {
    // ignore
  } finally {
    restoring.value = false;
  }
}

function openFeedback(note: Note) {
  feedbackNote.value = note;
  feedbackVisible.value = true;
}

async function handleSubmitFeedback(content: string) {
  if (!feedbackNote.value) return;
  submittingFeedback.value = true;
  try {
    await submitFeedback(feedbackNote.value.id, content);
    toast.success('反馈提交成功，已通知任务发起人');
    feedbackVisible.value = false;
    feedbackNote.value = null;
  } catch (e: unknown) {
    const err = e as { friendlyMessage?: string };
    toast.error(err.friendlyMessage || '反馈提交失败');
  } finally {
    submittingFeedback.value = false;
  }
}

function groupNotesByMonth(notes: Note[]) {
  const groups: { month: string; notes: Note[] }[] = [];
  const sorted = [...notes].sort(
    (a, b) =>
      new Date(b.archive_time || b.created_at).getTime() -
      new Date(a.archive_time || a.created_at).getTime()
  );
  for (const note of sorted) {
    const d = new Date(note.archive_time || note.created_at);
    const key = `${d.getFullYear()}年${d.getMonth() + 1}月`;
    let group = groups.find((g) => g.month === key);
    if (!group) {
      group = { month: key, notes: [] };
      groups.push(group);
    }
    group.notes.push(note);
  }
  return groups;
}

// Bug8：时间轴卡片展示反馈信息（被指派人反馈内容）
function feedbackList(note: Note) {
  return (note.assignees || [])
    .filter((a) => a.feedback_content)
    .map((a) => ({
      name: a.user?.name || '被指派人',
      content: a.feedback_content || '',
      time: a.feedback_at,
    }));
}
</script>

<template>
  <div>
    <!-- 筛选栏 -->
    <div
      class="bg-white dark:bg-slate-800 rounded-card p-4 mb-6 flex flex-wrap items-center gap-3 border border-slate-100 dark:border-slate-700 transition-colors duration-300"
    >
      <DateRangePicker
        v-model:from="dateFrom"
        v-model:to="dateTo"
        placeholder="选择归档日期范围"
      />
      <input
        v-model="keyword"
        class="input-field !w-40"
        placeholder="关键词搜索"
        @keyup.enter="handleSearch"
      />
      <button class="btn-primary text-sm !py-2" @click="handleSearch">搜索</button>
      <button
        class="px-4 py-2 text-sm text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 transition-smooth"
        @click="handleClear"
      >
        清空
      </button>
    </div>

    <!-- 视图切换 -->
    <div class="flex items-center justify-between mb-6">
      <span class="text-xs text-slate-400 dark:text-slate-500">共 {{ noteStore.totalCount }} 条归档记录</span>
      <div class="flex bg-slate-100 dark:bg-slate-800 rounded-btn p-0.5">
        <button
          :class="[
            'px-4 py-1.5 rounded-md text-sm font-medium transition-smooth',
            viewMode === 'timeline' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400',
          ]"
          @click="viewMode = 'timeline'"
        >
          时间轴
        </button>
        <button
          :class="[
            'px-4 py-1.5 rounded-md text-sm font-medium transition-smooth',
            viewMode === 'card' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400',
          ]"
          @click="viewMode = 'card'"
        >
          卡片
        </button>
      </div>
    </div>

    <!-- 加载 -->
    <div
      v-if="noteStore.loading"
      class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5"
    >
      <div v-for="n in 6" :key="n" class="skeleton h-44 rounded-card" />
    </div>

    <!-- 空态 -->
    <div
      v-else-if="noteStore.archivedNotes.length === 0"
      class="flex flex-col items-center justify-center py-24"
    >
      <div class="w-24 h-24 bg-slate-100 dark:bg-slate-800 rounded-3xl flex items-center justify-center mb-6">
        <svg class="w-12 h-12 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
          />
        </svg>
      </div>
      <p class="text-slate-400 dark:text-slate-500 text-sm">暂无归档任务</p>
      <p class="text-slate-300 dark:text-slate-600 text-xs mt-1">在工作台完成任务后，会自动归档到这里</p>
    </div>

    <!-- 时间轴视图 -->
    <div v-else-if="viewMode === 'timeline'" class="relative pl-8">
      <div class="absolute left-3 top-0 bottom-0 w-0.5 bg-slate-200" />
      <div
        v-for="group in groupNotesByMonth(noteStore.archivedNotes)"
        :key="group.month"
        class="mb-8"
      >
        <div class="flex items-center gap-3 mb-4">
          <div class="w-2.5 h-2.5 rounded-full bg-slate-300 -ml-[32px] ring-4 ring-white dark:ring-slate-950" />
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ group.month }}</span>
          <span class="text-xs text-slate-400 dark:text-slate-500">{{ group.notes.length }}条</span>
        </div>
        <div class="space-y-3">
          <div v-for="note in group.notes" :key="note.id" class="flex items-start gap-4">
            <span class="text-xs text-slate-400 w-10 shrink-0 pt-1"
              >{{ (note.archive_time || note.created_at)?.slice(8, 10) }}日</span
            >
            <div
              class="flex-1 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-4 relative hover:shadow-note transition-smooth cursor-pointer"
              @click="openDetail(note)"
            >
              <div class="flex items-start justify-between gap-2">
                <h4 class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
                  {{ note.title || '无标题' }}
                </h4>
                <span class="watermark-archived shrink-0">已归档</span>
              </div>

              <!-- Bug8：任务创建人信息 -->
              <div class="flex items-center gap-2 mt-1.5 text-xs text-slate-400 dark:text-slate-500">
                <span class="flex items-center gap-1 min-w-0">
                  <svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  <span class="truncate">{{ note.creator?.name || note.owner?.name || '未知用户' }}</span>
                </span>
                <span class="text-slate-300 dark:text-slate-600">·</span>
                <span>{{ (note.archive_time || note.created_at)?.slice(0, 10) }} 归档</span>
              </div>

              <!-- Bug8：详细信息（内容展示更完整） -->
              <p class="text-xs text-slate-500 dark:text-slate-400 line-clamp-3 mt-2 leading-relaxed">
                <span v-if="!note.content" class="text-slate-300 dark:text-slate-600">暂无内容</span>
                <span v-else v-html="renderNoteContent(note.content)"></span>
              </p>

              <div v-if="note.tags?.length" class="flex items-center gap-2 mt-2 flex-wrap">
                <span
                  v-for="tag in note.tags.slice(0, 3)"
                  :key="tag.id"
                  class="tag-capsule text-white text-[10px]"
                  :style="{ backgroundColor: tag.color || '#94A3B8' }"
                  >{{ tag.sub_tag ? tag.name + ' › ' + tag.sub_tag : tag.name }}</span
                >
              </div>

              <!-- Bug8：反馈信息 -->
              <div v-if="feedbackList(note).length" class="mt-2.5 space-y-1.5">
                <div
                  v-for="(fb, i) in feedbackList(note)"
                  :key="i"
                  class="rounded-lg bg-green-50 dark:bg-green-900/30 border-l-2 border-green-500 px-2.5 py-1.5 text-xs text-green-700 dark:text-green-300"
                >
                  <span class="font-semibold">💬 {{ fb.name }}：</span>
                  <span class="line-clamp-2">{{ fb.content }}</span>
                </div>
              </div>
              <div v-else class="mt-2 text-[11px] text-slate-300 dark:text-slate-600">暂无反馈</div>
            </div>
            <div class="flex flex-col gap-1 shrink-0 pt-1">
              <button
                class="text-xs px-2 py-1 bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-300 rounded hover:bg-blue-100 dark:hover:bg-blue-900/60 transition-smooth"
                @click="handleRestore(note)"
              >
                恢复
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <StickyNoteCard
        v-for="note in noteStore.archivedNotes"
        :key="note.id"
        :note="note"
        mode="web"
        :archived="true"
        class="animate-spring-enter"
        @click="openDetail(note)"
        @restore="handleRestore"
      />
    </div>

    <!-- 详情侧滑面板 -->
    <Teleport to="body">
      <div v-if="showDetailPanel && selectedNote">
        <div class="overlay-backdrop" @click="closeDetail" />
        <div class="slide-panel">
          <div class="p-6 h-full flex flex-col">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-2">
                <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">归档详情</h2>
                <span class="text-xs px-2 py-0.5 bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-300 rounded-tag"
                  >已归档</span
                >
              </div>
              <button
                class="p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
                @click="closeDetail"
              >
                <svg
                  class="w-5 h-5 text-slate-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <div class="flex-1 overflow-auto space-y-5">
              <div>
                <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">标题</span>
                <p class="text-sm font-semibold text-slate-900 dark:text-slate-100">{{ selectedNote.title }}</p>
              </div>
              <div>
                <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">内容</span>
                <div class="text-sm text-slate-700 dark:text-slate-300 rich-content-display">
                  <span v-if="!selectedNote.content" class="text-slate-300 dark:text-slate-500">暂无内容</span>
                  <span v-else v-html="renderNoteContent(selectedNote.content)"></span>
                </div>
              </div>
              <div v-if="selectedNote.tags?.length" class="flex flex-wrap gap-2">
                <span
                  v-for="tag in selectedNote.tags"
                  :key="tag.id"
                  class="tag-capsule text-white"
                  :style="{ backgroundColor: tag.color || '#64748B' }"
                  >{{ tag.sub_tag ? tag.name + ' › ' + tag.sub_tag : tag.name }}</span
                >
              </div>
              <div
                class="bg-slate-50 dark:bg-slate-900 rounded-card p-4 space-y-2 text-xs transition-colors duration-300"
              >
                <div class="flex justify-between">
                  <span class="text-slate-400 dark:text-slate-500">创建时间</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.created_at?.slice(0, 16).replace('T', ' ')
                  }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-slate-400 dark:text-slate-500">完成时间</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.completed_at?.slice(0, 16).replace('T', ' ') || '—'
                  }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-slate-400 dark:text-slate-500">归档时间</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.archive_time?.slice(0, 16).replace('T', ' ') || '—'
                  }}</span>
                </div>
              </div>
            </div>

            <div class="flex gap-3 pt-4 border-t border-slate-100 dark:border-slate-700 mt-4">
              <button
                class="flex-1 py-2.5 btn-primary text-sm disabled:opacity-50"
                :disabled="restoring"
                @click="handleRestore(selectedNote!)"
              >
                {{ restoring ? '恢复中...' : '恢复任务' }}
              </button>
              <button
                class="flex-1 py-2.5 text-sm bg-violet-500 text-white rounded-btn hover:bg-violet-600 transition-smooth"
                @click="openFeedback(selectedNote!)"
              >
                反馈填报
              </button>
              <button class="flex-1 py-2.5 btn-secondary text-sm" @click="closeDetail">关闭</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 补充反馈填报弹窗 -->
    <FeedbackModal
      :visible="feedbackVisible"
      :note="feedbackNote"
      mode="feedback"
      @update:visible="feedbackVisible = $event"
      @submit="handleSubmitFeedback"
    />
  </div>
</template>
