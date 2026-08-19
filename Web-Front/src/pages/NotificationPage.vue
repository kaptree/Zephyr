<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import type { NotificationItem } from '@/types';
import { renderNoteContent } from '@/utils/richText';

const router = useRouter();
const store = useNotificationStore();

const loading = ref(false);
const list = ref<NotificationItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;
const selected = ref<NotificationItem | null>(null);
const unreadOnly = ref(false);

const TYPE_LABEL: Record<string, string> = {
  task_assigned: '任务指派',
  task_completed: '任务完成',
  task_feedback: '任务反馈',
  task_remind: '催办提醒',
  task_cc: '任务抄送',
  task_signed: '任务签收',
  system: '系统通知',
};

const TYPE_COLOR: Record<string, string> = {
  task_assigned: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300',
  task_completed: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-300',
  task_feedback: 'bg-violet-100 text-violet-600 dark:bg-violet-900/40 dark:text-violet-300',
  task_remind: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-300',
  task_cc: 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-300',
  task_signed: 'bg-teal-100 text-teal-600 dark:bg-teal-900/40 dark:text-teal-300',
  system: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
};

const hasMore = computed(() => list.value.length < total.value);

async function load(reset = false) {
  loading.value = true;
  try {
    if (reset) {
      page.value = 1;
      list.value = [];
    }
    const res = await store.fetchList({
      page: page.value,
      page_size: pageSize,
      unread_only: unreadOnly.value || undefined,
    });
    const data = res as unknown as { data: NotificationItem[]; total: number };
    list.value = page.value === 1 ? data.data : [...list.value, ...data.data];
    total.value = data.total;
  } finally {
    loading.value = false;
  }
}

function loadMore() {
  if (loading.value || !hasMore.value) return;
  page.value += 1;
  load();
}

async function handleMarkRead(n: NotificationItem) {
  if (!n.is_read) {
    await store.markRead(n.id);
    n.is_read = true;
  }
  selected.value = n;
}

async function handleDelete(n: NotificationItem) {
  await store.remove(n.id);
  list.value = list.value.filter((x) => x.id !== n.id);
  total.value = Math.max(0, total.value - 1);
  if (selected.value?.id === n.id) selected.value = null;
}

function goToNote(n: NotificationItem) {
  if (n.note_id) {
    router.push({ path: '/workbench', query: { note: n.note_id } });
  }
}

function toggleUnreadOnly() {
  unreadOnly.value = !unreadOnly.value;
  load(true);
}

function formatTime(t?: string) {
  if (!t) return '';
  const d = new Date(t);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  if (diff < 60 * 1000) return '刚刚';
  if (diff < 60 * 60 * 1000) return `${Math.floor(diff / 60000)} 分钟前`;
  if (diff < 24 * 60 * 60 * 1000) return `${Math.floor(diff / 3600000)} 小时前`;
  return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
}

onMounted(() => {
  load(true);
  store.fetchUnreadCount();
});
</script>

<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">通知中心</h1>
        <p class="mt-0.5 text-xs text-slate-400">共 {{ total }} 条通知</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth"
          :class="{ 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-300': unreadOnly }"
          @click="toggleUnreadOnly"
        >
          {{ unreadOnly ? '显示全部' : '只看未读' }}
        </button>
        <button
          v-if="store.unreadCount > 0"
          class="px-3 py-1.5 text-sm rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-smooth"
          @click="store.markAllRead().then(() => load())"
        >
          全部已读
        </button>
        <button
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth"
          @click="router.push('/workbench')"
        >
          返回工作台
        </button>
      </div>
    </div>

    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden">
      <div v-if="loading && list.length === 0" class="py-20 text-center text-sm text-slate-400">
        加载中...
      </div>
      <div v-else-if="list.length === 0" class="py-20 text-center">
        <svg class="w-12 h-12 mx-auto mb-3 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        <p class="text-sm text-slate-400">暂无通知</p>
      </div>
      <div v-else>
        <div
          v-for="n in list"
          :key="n.id"
          class="flex items-start gap-3 px-4 py-3.5 border-b border-slate-50 dark:border-slate-700/50 hover:bg-slate-50 dark:hover:bg-slate-700/40 cursor-pointer transition-smooth"
          :class="{ 'bg-blue-50/60 dark:bg-blue-900/10': !n.is_read }"
          @click="handleMarkRead(n)"
        >
          <span
            class="mt-0.5 shrink-0 px-2 py-0.5 rounded text-[11px] leading-4 font-medium"
            :class="TYPE_COLOR[n.type] || TYPE_COLOR.system"
          >
            {{ TYPE_LABEL[n.type] || '通知' }}
          </span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-100 truncate">
                {{ n.title }}
                <span v-if="!n.is_read" class="ml-1 inline-block w-1.5 h-1.5 rounded-full bg-red-500 align-middle" />
              </p>
              <span class="shrink-0 text-[11px] text-slate-400">{{ formatTime(n.created_at) }}</span>
            </div>
            <p class="mt-0.5 text-xs text-slate-500 dark:text-slate-400 line-clamp-2 break-all">{{ n.content }}</p>
          </div>
          <button
            v-if="n.note_id"
            class="shrink-0 px-2 py-1 text-[11px] rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-300 hover:bg-blue-100 transition-smooth"
            @click.stop="goToNote(n)"
          >
            查看任务
          </button>
          <button
            class="shrink-0 p-1 text-slate-300 hover:text-red-400 transition-smooth"
            title="删除"
            @click.stop="handleDelete(n)"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
        <div v-if="hasMore" class="py-3 text-center">
          <button
            class="text-sm text-blue-500 hover:text-blue-600 disabled:opacity-50"
            :disabled="loading"
            @click="loadMore"
          >
            {{ loading ? '加载中...' : '加载更多' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 通知详情弹窗 -->
    <div
      v-if="selected"
      class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4"
      @click.self="selected = null"
    >
      <div class="w-full max-w-lg bg-white dark:bg-slate-800 rounded-xl shadow-2xl overflow-hidden">
        <div class="flex items-center justify-between px-5 py-4 border-b border-slate-100 dark:border-slate-700">
          <div class="flex items-center gap-2">
            <span class="px-2 py-0.5 rounded text-xs font-medium" :class="TYPE_COLOR[selected.type] || TYPE_COLOR.system">
              {{ TYPE_LABEL[selected.type] || '通知' }}
            </span>
            <h3 class="text-base font-semibold text-slate-800 dark:text-slate-100">{{ selected.title }}</h3>
          </div>
          <button class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth" @click="selected = null">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-5 py-4 max-h-[60vh] overflow-y-auto scrollbar-thin">
          <div class="flex items-center gap-2 mb-3 text-xs text-slate-400">
            <span v-if="selected.sender">来自：{{ selected.sender.username }}</span>
            <span v-if="selected.sender && selected.created_at">·</span>
            <span>{{ formatTime(selected.created_at) }}</span>
          </div>
          <div class="text-sm leading-relaxed text-slate-700 dark:text-slate-200 whitespace-pre-wrap break-all [&_*]:break-all" v-html="renderNoteContent(selected.content)" />
        </div>
        <div class="flex justify-end gap-2 px-5 py-3 border-t border-slate-100 dark:border-slate-700">
          <button class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth" @click="selected = null">关闭</button>
          <button v-if="selected.note_id" class="px-3 py-1.5 text-sm rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-smooth" @click="goToNote(selected)">查看任务</button>
        </div>
      </div>
    </div>
  </div>
</template>
