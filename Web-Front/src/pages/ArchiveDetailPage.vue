<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useToast } from '@/composables/useToast';
import { fetchNoteById, submitFeedback, restoreNote } from '@/services/notes';
import type { Note, NoteAssignee } from '@/types';
import FeedbackModal from '@/components/notification/FeedbackModal.vue';
import { renderNoteContent } from '@/utils/richText';

const route = useRoute();
const router = useRouter();
const toast = useToast();

const note = ref<Note | null>(null);
const loading = ref(false);
const error = ref('');
const restoring = ref(false);
const feedbackVisible = ref(false);
const submittingFeedback = ref(false);

function fmtTime(ts?: string): string {
  if (!ts) return '—';
  return ts.slice(0, 16).replace('T', ' ');
}

/** 被指派人反馈列表 */
const feedbackList = computed(() => {
  if (!note.value?.assignees) return [];
  return note.value.assignees
    .filter((a) => a.feedback_content)
    .map((a) => ({
      name: a.user?.name || '被指派人',
      content: a.feedback_content || '',
      time: a.feedback_at,
      completed: a.is_completed,
      completed_at: a.completed_at,
    }));
});

const sourceTypeLabel = computed(() => {
  const t = note.value?.source_type;
  return t === 'self' ? '自己创建' : t === 'assigned' ? '上级指派' : '协同任务';
});

async function load() {
  loading.value = true;
  error.value = '';
  try {
    const res = await fetchNoteById(route.params.id as string);
    note.value = res.data as unknown as Note;
  } catch {
    error.value = '加载归档详情失败';
  } finally {
    loading.value = false;
  }
}

async function handleRestore() {
  if (!note.value || restoring.value) return;
  restoring.value = true;
  try {
    await restoreNote(note.value.id);
    toast.success('任务已恢复至工作台');
    router.push('/workbench/archive');
  } catch {
    toast.error('恢复失败，请重试');
  } finally {
    restoring.value = false;
  }
}

function openFeedback() {
  feedbackVisible.value = true;
}

async function handleSubmitFeedback(content: string) {
  if (!note.value) return;
  submittingFeedback.value = true;
  try {
    await submitFeedback(note.value.id, content);
    toast.success('反馈提交成功，已通知任务发起人');
    feedbackVisible.value = false;
    await load(); // 刷新反馈列表
  } catch (e: unknown) {
    const err = e as { friendlyMessage?: string };
    toast.error(err.friendlyMessage || '反馈提交失败');
  } finally {
    submittingFeedback.value = false;
  }
}

/** 从指派对象取用户姓名 */
function assigneeName(a: NoteAssignee): string {
  return a.user?.name || '被指派人';
}

onMounted(load);
</script>

<template>
  <div>
    <!-- 返回按钮 -->
    <button
      class="text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 mb-4 transition-smooth"
      @click="router.push('/workbench/archive')"
    >
      ← 返回归档列表
    </button>

    <!-- 加载 -->
    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-20 rounded-card" />
      <div class="skeleton h-64 rounded-card" />
    </div>

    <!-- 错误 -->
    <div v-else-if="error" class="text-center py-16 text-sm text-red-400">
      {{ error }}
      <button class="block mx-auto mt-2 text-blue-500 hover:underline" @click="load">重试</button>
    </div>

    <template v-else-if="note">
      <!-- 主卡片：标题 + 内容 -->
      <div
        class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300"
      >
        <!-- 头部 -->
        <div class="flex items-start justify-between gap-4 mb-4">
          <div class="min-w-0">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 leading-snug">
              {{ note.title || '无标题' }}
            </h2>
            <div class="flex items-center gap-2 mt-1.5 flex-wrap">
              <span
                class="text-xs px-2 py-0.5 bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-300 rounded-tag"
                >已归档</span
              >
              <span
                class="text-xs px-2 py-0.5 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 rounded-tag"
                >{{ sourceTypeLabel }}</span
              >
              <span v-if="note.serial_no" class="text-xs text-slate-400">{{ note.serial_no }}</span>
            </div>
          </div>
        </div>

        <!-- 创建人信息条 -->
        <div class="flex items-center gap-2 px-3 py-2 bg-slate-50 dark:bg-slate-900 rounded-lg mb-5">
          <div
            class="w-7 h-7 rounded-full bg-blue-500 flex items-center justify-center text-white text-[10px] font-medium shrink-0"
          >
            {{ (note.creator?.name || '未').charAt(0) }}
          </div>
          <div class="text-xs text-slate-600 dark:text-slate-300">
            <span class="font-medium">{{ note.creator?.name || '未知用户' }}</span>
            <span class="text-slate-400 dark:text-slate-500"> 于 {{ fmtTime(note.created_at) }} 创建</span>
          </div>
        </div>

        <!-- 正文内容 -->
        <div>
          <span class="text-xs text-slate-400 dark:text-slate-500 mb-1.5 block">任务内容</span>
          <div
            v-if="note.content"
            class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed rich-content-display prose prose-sm dark:prose-invert max-w-none [&_p]:my-1 [&_ul]:my-1 [&_ol]:my-1 [&_pre]:my-2 [&_h1]:text-base [&_h2]:text-base [&_h3]:text-sm"
            v-html="renderNoteContent(note.content)"
          ></div>
          <span v-else class="text-sm text-slate-300 dark:text-slate-600">暂无内容</span>
        </div>

        <!-- 标签 -->
        <div v-if="note.tags?.length" class="flex flex-wrap gap-2 mt-4">
          <span
            v-for="tag in note.tags"
            :key="tag.id"
            class="tag-capsule text-white text-[10px]"
            :style="{ backgroundColor: tag.color || '#64748B' }"
            >{{ tag.sub_tag ? tag.name + ' › ' + tag.sub_tag : tag.name }}</span
          >
        </div>
      </div>

      <!-- 元数据卡片 -->
      <div
        class="mt-5 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300"
      >
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">📋 任务信息</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
          <div class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">来源类型</span>
            <span class="text-slate-700 dark:text-slate-300">{{ sourceTypeLabel }}</span>
          </div>
          <div class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">创建时间</span>
            <span class="text-slate-700 dark:text-slate-300">{{ fmtTime(note.created_at) }}</span>
          </div>
          <div class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">完成时间</span>
            <span class="text-slate-700 dark:text-slate-300">{{ fmtTime(note.completed_at) }}</span>
          </div>
          <div class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">归档时间</span>
            <span class="text-slate-700 dark:text-slate-300">{{ fmtTime(note.archive_time) }}</span>
          </div>
          <div v-if="note.due_time" class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">截止时间</span>
            <span class="text-red-500">{{ fmtTime(note.due_time) }}</span>
          </div>
          <div v-if="note.template_type" class="flex justify-between p-2.5 bg-slate-50 dark:bg-slate-900 rounded-lg">
            <span class="text-slate-400 dark:text-slate-500">模板类型</span>
            <span class="text-slate-700 dark:text-slate-300">{{ note.template_type }}</span>
          </div>
        </div>

        <!-- 被指派人 -->
        <div v-if="note.assignees?.length" class="mt-4">
          <span class="text-xs text-slate-400 dark:text-slate-500 mb-2 block">负责人 / 被指派人</span>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="a in note.assignees"
              :key="a.user_id"
              class="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300"
            >
              {{ assigneeName(a) }}
              <span
                v-if="a.is_completed"
                class="text-[10px] px-1 py-0.5 rounded-full bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300"
                >已完成</span
              >
            </span>
          </div>
        </div>

        <!-- 抄送人 -->
        <div v-if="note.ccs?.length" class="mt-4">
          <span class="text-xs text-slate-400 dark:text-slate-500 mb-2 block">抄送人</span>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="c in note.ccs"
              :key="c.user_id"
              class="inline-flex items-center text-xs px-2.5 py-1 rounded-full bg-purple-50 dark:bg-purple-900/40 text-purple-600 dark:text-purple-300"
            >
              {{ c.user?.name || '抄送人' }}
            </span>
          </div>
        </div>
      </div>

      <!-- 任务反馈卡片 -->
      <div
        class="mt-5 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300"
      >
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">
            💬 任务反馈（{{ feedbackList.length }}）
          </h3>
          <button
            class="text-xs px-3 py-1.5 bg-violet-50 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 rounded-btn hover:bg-violet-100 dark:hover:bg-violet-900/60 transition-smooth"
            @click="openFeedback"
          >
            补充反馈
          </button>
        </div>

        <div v-if="feedbackList.length === 0" class="text-center py-8 text-sm text-slate-400 dark:text-slate-500">
          暂无反馈
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(fb, i) in feedbackList"
            :key="i"
            class="rounded-lg bg-green-50 dark:bg-green-900/30 border-l-2 border-green-500 px-4 py-3"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs font-semibold text-green-700 dark:text-green-300">💬 {{ fb.name }}</span>
              <span v-if="fb.time" class="text-[10px] text-slate-400 dark:text-slate-500">{{
                fb.time.slice(0, 16).replace('T', ' ')
              }}</span>
            </div>
            <div
              class="text-sm text-green-800 dark:text-green-200 rich-content-display"
              v-html="renderNoteContent(fb.content)"
            ></div>
            <div v-if="fb.completed" class="mt-1.5 text-[10px] text-green-600 dark:text-green-400">
              <span v-if="fb.completed_at">完成于 {{ fb.completed_at.slice(0, 16).replace('T', ' ') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 操作栏 -->
      <div class="flex gap-3 mt-5">
        <button
          class="flex-1 py-2.5 btn-primary text-sm disabled:opacity-50"
          :disabled="restoring"
          @click="handleRestore"
        >
          {{ restoring ? '恢复中...' : '恢复任务至工作台' }}
        </button>
        <button class="flex-1 py-2.5 btn-secondary text-sm" @click="router.push('/workbench/archive')">
          返回列表
        </button>
      </div>
    </template>

    <!-- 补充反馈填报弹窗 -->
    <FeedbackModal
      :visible="feedbackVisible"
      :note="note"
      mode="feedback"
      @update:visible="feedbackVisible = $event"
      @submit="handleSubmitFeedback"
    />
  </div>
</template>
