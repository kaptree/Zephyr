<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import type { NotificationItem } from '@/types';
import { renderNoteContent } from '@/utils/richText';

const router = useRouter();
const store = useNotificationStore();

const open = ref(false);
const loading = ref(false);
const selected = ref<NotificationItem | null>(null);

const TYPE_LABEL: Record<string, string> = {
  task_assigned: '任务指派',
  task_completed: '任务完成',
  task_feedback: '任务反馈',
  task_remind: '催办提醒',
  system: '系统通知',
};

const TYPE_COLOR: Record<string, string> = {
  task_assigned: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300',
  task_completed: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-300',
  task_feedback: 'bg-violet-100 text-violet-600 dark:bg-violet-900/40 dark:text-violet-300',
  task_remind: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-300',
  system: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
};

const displayList = computed(() => store.notifications.slice(0, 20));

function toggle() {
  open.value = !open.value;
  if (open.value) load();
}

function load() {
  loading.value = true;
  store
    .fetchList({ page: 1, page_size: 20 })
    .finally(() => {
      loading.value = false;
    });
}

async function handleClickItem(n: NotificationItem) {
  if (!n.is_read) await store.markRead(n.id);
  selected.value = n;
}

async function handleMarkAll() {
  await store.markAllRead();
}

async function handleDelete(n: NotificationItem, e: Event) {
  e.stopPropagation();
  await store.remove(n.id);
  if (selected.value?.id === n.id) selected.value = null;
}

function goToNote(n: NotificationItem) {
  open.value = false;
  selected.value = null;
  if (n.note_id) {
    router.push({ path: '/workbench', query: { note: n.note_id } });
  }
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

function closeOnOutside(e: MouseEvent) {
  const el = document.querySelector('[data-notif-bell]');
  if (el && !el.contains(e.target as Node)) {
    open.value = false;
  }
}

onMounted(() => {
  store.fetchUnreadCount();
  document.addEventListener('click', closeOnOutside);
});
onUnmounted(() => {
  document.removeEventListener('click', closeOnOutside);
});
</script>

<template>
  <div class="relative" data-notif-bell>
    <!-- 铃铛按钮 -->
    <button
      class="relative p-2 rounded-lg text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth"
      title="通知"
      @click.stop="toggle"
    >
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
      <span
        v-if="store.unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold flex items-center justify-center"
      >
        {{ store.unreadCount > 99 ? '99+' : store.unreadCount }}
      </span>
    </button>

    <!-- 下拉面板 -->
    <div
      v-if="open"
      class="absolute right-0 mt-2 w-[380px] max-w-[calc(100vw-2rem)] bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl shadow-xl overflow-hidden z-50"
    >
      <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700">
        <span class="text-sm font-semibold text-slate-800 dark:text-slate-100">通知中心</span>
        <div class="flex items-center gap-2">
          <button
            v-if="store.unreadCount > 0"
            class="text-xs text-blue-500 hover:text-blue-600 transition-smooth"
            @click="handleMarkAll"
          >
            全部已读
          </button>
          <button class="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-300" @click="open = false">
            关闭
          </button>
        </div>
      </div>

      <div class="max-h-[380px] overflow-y-auto scrollbar-thin">
        <div v-if="loading" class="py-10 text-center text-sm text-slate-400">加载中...</div>
        <div v-else-if="displayList.length === 0" class="py-10 text-center">
          <svg class="w-10 h-10 mx-auto mb-2 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <p class="text-sm text-slate-400">暂无通知</p>
        </div>
        <div
          v-for="n in displayList"
          :key="n.id"
          class="px-4 py-3 border-b border-slate-50 dark:border-slate-700/50 hover:bg-slate-50 dark:hover:bg-slate-700/40 cursor-pointer transition-smooth"
          :class="{ 'bg-blue-50/60 dark:bg-blue-900/10': !n.is_read }"
          @click="handleClickItem(n)"
        >
          <div class="flex items-start gap-2.5">
            <span
              class="mt-0.5 shrink-0 px-1.5 py-0.5 rounded text-[10px] leading-4 font-medium"
              :class="TYPE_COLOR[n.type] || TYPE_COLOR.system"
            >
              {{ TYPE_LABEL[n.type] || '通知' }}
            </span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between gap-2">
                <p class="text-sm font-medium text-slate-800 dark:text-slate-100 truncate">{{ n.title }}</p>
                <span class="shrink-0 text-[10px] text-slate-400">{{ formatTime(n.created_at) }}</span>
              </div>
              <p class="mt-0.5 text-xs text-slate-500 dark:text-slate-400 line-clamp-2 break-all">
                {{ n.content }}
              </p>
            </div>
            <button
              class="shrink-0 p-1 text-slate-300 hover:text-red-400 transition-smooth"
              title="删除"
              @click="handleDelete(n, $event)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
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
            <span
              class="px-2 py-0.5 rounded text-xs font-medium"
              :class="TYPE_COLOR[selected.type] || TYPE_COLOR.system"
            >
              {{ TYPE_LABEL[selected.type] || '通知' }}
            </span>
            <h3 class="text-base font-semibold text-slate-800 dark:text-slate-100">{{ selected.title }}</h3>
          </div>
          <button
            class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
            @click="selected = null"
          >
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
          <div
            class="text-sm leading-relaxed text-slate-700 dark:text-slate-200 whitespace-pre-wrap break-all [&_*]:break-all"
            v-html="renderNoteContent(selected.content)"
          />
        </div>
        <div class="flex justify-end gap-2 px-5 py-3 border-t border-slate-100 dark:border-slate-700">
          <button
            class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth"
            @click="selected = null"
          >
            关闭
          </button>
          <button
            v-if="selected.note_id"
            class="px-3 py-1.5 text-sm rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-smooth"
            @click="goToNote(selected)"
          >
            查看任务
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
