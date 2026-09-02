<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import type { NotificationItem } from '@/types';
import { renderNoteContent } from '@/utils/richText';
import { playDeleteOut } from '@/utils/exitAnimations';
import { AppIcon, EmptyIllustration } from '@/components/icons';

const router = useRouter();
const store = useNotificationStore();

const open = ref(false);
const loading = ref(false);
const selected = ref<NotificationItem | null>(null);

/* ===== 「生命力衰减」生命周期状态 ===== */
const readingIds = reactive(new Set<string>()); // 单条已读三步过渡中
const tideIds = reactive(new Set<string>()); // 全部已读潮汐退场中
const tideDelays = ref<Record<string, number>>({});
const markAllRunning = ref(false);
const shaking = ref(false); // 新消息时铃铛晃动
const popping = ref(false); // 新消息时角标弹跳 +1
const displayUnread = ref(store.unreadCount); // 角标数字（全部已读时滚动递减）
let shakeTimer: number | undefined;
let popTimer: number | undefined;
let countRaf = 0;

const sleep = (ms: number) => new Promise<void>((r) => setTimeout(r, ms));

const TYPE_LABEL: Record<string, string> = {
  task_assigned: '任务指派',
  task_completed: '任务完成',
  task_feedback: '任务反馈',
  task_remind: '催办提醒',
  task_cc: '任务抄送',
  task_signed: '任务签收',
  issue_comment: '问题评论',
  issue_new: '新建问题',
  system: '系统通知',
};

const TYPE_COLOR: Record<string, string> = {
  task_assigned: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300',
  task_completed: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-300',
  task_feedback: 'bg-violet-100 text-violet-600 dark:bg-violet-900/40 dark:text-violet-300',
  task_remind: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-300',
  task_cc: 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-300',
  task_signed: 'bg-teal-100 text-teal-600 dark:bg-teal-900/40 dark:text-teal-300',
  issue_comment: 'bg-pink-100 text-pink-600 dark:bg-pink-900/40 dark:text-pink-300',
  issue_new: 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/40 dark:text-cyan-300',
  system: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
};

/* 优先级 → 能量色：紧急=红 / 重要=橙 / 普通=蓝 */
const PRIORITY_CLASS: Record<string, string> = {
  task_remind: 'notif-p-urgent',
  task_assigned: 'notif-p-important',
  task_feedback: 'notif-p-important',
};
function priorityClass(n: NotificationItem) {
  return PRIORITY_CLASS[n.type] || 'notif-p-normal';
}

function tideDelayStyle(n: NotificationItem) {
  const d = tideDelays.value[n.id];
  return d === undefined ? undefined : { '--tide-delay': `${d}ms` };
}

/* 新消息入场联动：铃铛晃动 600ms + 角标弹跳 +1 */
function triggerShake() {
  shaking.value = false;
  requestAnimationFrame(() => {
    shaking.value = true;
    if (shakeTimer) clearTimeout(shakeTimer);
    shakeTimer = window.setTimeout(() => (shaking.value = false), 700);
  });
}

function triggerPop() {
  popping.value = false;
  requestAnimationFrame(() => {
    popping.value = true;
    if (popTimer) clearTimeout(popTimer);
    popTimer = window.setTimeout(() => (popping.value = false), 500);
  });
}

function animateCountDown(from: number, duration: number) {
  const start = performance.now();
  const tick = (t: number) => {
    const p = Math.min(1, (t - start) / duration);
    displayUnread.value = Math.round(from * (1 - p));
    if (p < 1) countRaf = requestAnimationFrame(tick);
  };
  countRaf = requestAnimationFrame(tick);
}

/* WebSocket 推送新通知（列表头部 id 变化）→ 晃动 + 弹跳 */
watch(
  () => store.notifications,
  (nv, ov) => {
    const first = nv[0];
    if (!first || first.is_read) return;
    if (ov && ov.length > 0 && ov[0]?.id === first.id) return;
    triggerShake();
    triggerPop();
  }
);

/* 角标数字：全部已读（>0 → 0）时滚动递减归零，其余直接同步 */
watch(
  () => store.unreadCount,
  (nv, ov) => {
    if (ov > 0 && nv === 0) {
      animateCountDown(ov, 600);
    } else {
      displayUnread.value = nv;
    }
  }
);

const displayList = computed(() => store.notifications.slice(0, 10));

function toggle() {
  open.value = !open.value;
  if (open.value) load();
}

function load() {
  loading.value = true;
  store
    .fetchList({ page: 1, page_size: 10 })
    .finally(() => {
      loading.value = false;
    });
}

/* 单条已读三步过渡（500ms）后打开详情 */
async function markOneRead(n: NotificationItem) {
  if (n.is_read || readingIds.has(n.id) || tideIds.has(n.id)) return;
  readingIds.add(n.id);
  await sleep(500);
  readingIds.delete(n.id);
  if (n.is_read) return;
  await store.markRead(n.id);
}

async function handleClickItem(n: NotificationItem) {
  if (!n.is_read) await markOneRead(n);
  selected.value = n;
}

/* 面板版「信息潮汐」：竖条 50ms 间隔依次收缩变灰 + 圆点消散 */
async function handleMarkAll() {
  if (markAllRunning.value) return;
  const unread = displayList.value.filter((n) => !n.is_read);
  if (unread.length === 0) {
    await store.markAllRead();
    return;
  }
  markAllRunning.value = true;
  const step = Math.min(50, Math.max(15, Math.floor(800 / unread.length)));
  unread.forEach((n, i) => {
    tideIds.add(n.id);
    tideDelays.value[n.id] = i * step;
  });
  await sleep((unread.length - 1) * step + 600);
  tideIds.clear();
  tideDelays.value = {};
  await store.markAllRead();
  markAllRunning.value = false;
}

async function handleDelete(n: NotificationItem, e: Event) {
  e.stopPropagation();
  await playDeleteOut(document.querySelector(`[data-notification-id="${n.id}"]`));
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

function goToAll() {
  open.value = false;
  router.push('/notifications');
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
  if (shakeTimer) clearTimeout(shakeTimer);
  if (popTimer) clearTimeout(popTimer);
  if (countRaf) cancelAnimationFrame(countRaf);
});
</script>

<template>
  <div class="relative" data-notif-bell>
    <!-- 铃铛按钮 -->
    <button
      class="relative p-2 rounded-lg text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth"
      v-tooltip="'通知'"
      @click.stop="toggle"
    >
      <!-- 铃铛：未读 solid / 已读 outline（250ms 交叉淡入）；新消息时晃动 600ms -->
      <AppIcon
        name="bell"
        :size="20"
        :variant="store.unreadCount > 0 ? 'solid' : 'outline'"
        :class="shaking ? 'animate-bell-shake' : ''"
      />
      <span
        v-if="store.unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold flex items-center justify-center animate-breathe"
      >
        {{ store.unreadCount > 99 ? '99+' : store.unreadCount }}
      </span>
    </button>

    <!-- 下拉面板 -->
    <transition name="shrink-out">
      <div
        v-if="open"
        class="absolute right-0 mt-2 w-[380px] max-w-[calc(100vw-2rem)] bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl shadow-xl overflow-hidden z-50"
      >
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700">
          <span class="text-sm font-semibold text-slate-800 dark:text-slate-100">通知中心</span>
          <div class="flex items-center gap-2">
            <button
              v-if="store.unreadCount > 0"
              class="text-xs text-blue-500 hover:text-blue-600 transition-smooth inline-flex items-center gap-1 disabled:opacity-60"
              :disabled="markAllRunning"
              @click="handleMarkAll"
            >
              <svg v-if="markAllRunning" class="w-3 h-3 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
              </svg>
              {{ markAllRunning ? '执行中...' : '全部已读' }}
            </button>
            <button class="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-300" @click="open = false">
              关闭
            </button>
          </div>
        </div>

        <div class="max-h-[380px] overflow-y-auto scrollbar-thin">
          <div v-if="loading" class="py-10 text-center text-sm text-slate-400">加载中...</div>
          <div v-else-if="displayList.length === 0" class="py-8 text-center">
            <EmptyIllustration kind="inbox" :size="44" label="暂无通知" class="mx-auto mb-1" />
            <p class="text-sm text-slate-400 dark:text-slate-500">暂无通知</p>
          </div>
          <div v-else>
            <div
              v-for="n in displayList"
              :key="n.id"
              :data-notification-id="n.id"
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
                  v-tooltip="'删除'"
                  @click="handleDelete(n, $event)"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
            <!-- 查看全部 -->
            <div class="px-4 py-2.5 border-t border-slate-100 dark:border-slate-700">
              <button
                class="w-full py-1.5 text-xs text-blue-500 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-smooth"
                @click="goToAll"
              >
                查看全部通知 →
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <!-- 通知详情弹窗 -->
    <transition name="shrink-out">
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
    </transition>
  </div>
</template>
