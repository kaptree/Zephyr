<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, type CSSProperties } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import type { NotificationItem } from '@/types';
import { renderNoteContent } from '@/utils/richText';
import { playDeleteOut } from '@/utils/exitAnimations';

const router = useRouter();
const store = useNotificationStore();

const loading = ref(false);
const list = ref<NotificationItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;
const selected = ref<NotificationItem | null>(null);
const unreadOnly = ref(false);

/* ===== 「生命力衰减」生命周期状态 ===== */
const readingIds = reactive(new Set<string>()); // 单条已读三步过渡中
const tideIds = reactive(new Set<string>()); // 全部已读「潮汐」退场中
const tideDelays = ref<Record<string, number>>({}); // 潮汐 stagger 延迟（ms）
const animatedChecks = reactive(new Set<string>()); // 已播过 ✓ 弹出动画的消息
const tideRunning = ref(false);
const toastVisible = ref(false);
const toastMessage = ref('');
const displayUnread = ref(0); // 角标数字（潮汐时滚动递减）
const collapsedGroups = reactive(new Set<string>());
let toastTimer: number | undefined;

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

/* 优先级 → 能量色（呼吸圆点与竖条共用）：紧急=红 / 重要=橙 / 普通=蓝 */
const PRIORITY_CLASS: Record<string, string> = {
  task_remind: 'notif-p-urgent',
  task_assigned: 'notif-p-important',
  task_feedback: 'notif-p-important',
};
function priorityClass(n: NotificationItem) {
  return PRIORITY_CLASS[n.type] || 'notif-p-normal';
}

/* ===== 日期分组：今天 / 昨天 / 更早 ===== */
type GroupKey = 'today' | 'yesterday' | 'earlier';
const GROUP_LABEL: Record<GroupKey, string> = { today: '今天', yesterday: '昨天', earlier: '更早' };

function dateGroupOf(t?: string): GroupKey {
  if (!t) return 'earlier';
  const d = new Date(t);
  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const ts = d.getTime();
  if (ts >= startOfToday) return 'today';
  if (ts >= startOfToday - 24 * 3600 * 1000) return 'yesterday';
  return 'earlier';
}

const groups = computed(() => {
  const map: Record<GroupKey, NotificationItem[]> = { today: [], yesterday: [], earlier: [] };
  for (const n of list.value) map[dateGroupOf(n.created_at)].push(n);
  return (Object.keys(map) as GroupKey[])
    .filter((k) => map[k].length > 0)
    .map((k) => ({ key: k, label: GROUP_LABEL[k], items: map[k] }));
});

function toggleGroup(key: string) {
  if (collapsedGroups.has(key)) collapsedGroups.delete(key);
  else collapsedGroups.add(key);
}

function tideDelayStyle(n: NotificationItem): CSSProperties | undefined {
  const d = tideDelays.value[n.id];
  return d === undefined ? undefined : ({ '--tide-delay': `${d}ms` } as CSSProperties);
}

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
    list.value =
      page.value === 1
        ? data.data
        : [...list.value, ...data.data.filter((x) => !list.value.some((y) => y.id === x.id))];
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

/* ===== 单条「标为已读」三步过渡（总时长 500ms） =====
   第一步 200ms 竖条收缩变灰 → 第二步 300ms 光晕径向褪色 → 第三步 400ms 圆点淡出 + ✓ 弹出 */
async function markOneRead(n: NotificationItem) {
  if (n.is_read || readingIds.has(n.id) || tideIds.has(n.id)) return;
  readingIds.add(n.id);
  await sleep(500);
  readingIds.delete(n.id);
  if (n.is_read) return; // 过渡期间可能已被潮汐处理
  await store.markRead(n.id);
  n.is_read = true; // 页面列表与 store 为不同代理，需本地同步以触发渲染
  animatedChecks.add(n.id);
}

/* 点击卡片：未读先播三步过渡再打开详情，已读直接打开 */
async function onItemClick(n: NotificationItem) {
  if (!n.is_read) {
    await markOneRead(n);
  }
  selected.value = n;
}

/* ===== 「全部已读」信息潮汐 ===== */
async function handleMarkAllRead() {
  if (tideRunning.value) return;
  const unread = list.value.filter((n) => !n.is_read);
  if (unread.length === 0) {
    await store.markAllRead();
    return;
  }
  tideRunning.value = true;
  // 竖条以 50ms 间隔依次收缩（条数多时压缩间隔，总传播 ≤800ms）
  const step = Math.min(50, Math.max(15, Math.floor(800 / unread.length)));
  unread.forEach((n, i) => {
    tideIds.add(n.id);
    tideDelays.value[n.id] = i * step;
  });
  // 角标数字滚动递减归零
  animateCountDown(store.unreadCount, (unread.length - 1) * step + 600);
  await sleep((unread.length - 1) * step + 600);
  tideIds.clear();
  tideDelays.value = {};
  await store.markAllRead();
  unread.forEach((n) => animatedChecks.add(n.id));
  tideRunning.value = false;
  showToast('✨ 已清空所有未读');
}

function animateCountDown(from: number, duration: number) {
  const start = performance.now();
  const tick = (t: number) => {
    const p = Math.min(1, (t - start) / duration);
    displayUnread.value = Math.round(from * (1 - p));
    if (p < 1) requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
}

function showToast(msg: string) {
  toastMessage.value = msg;
  toastVisible.value = true;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => {
    toastVisible.value = false;
  }, 2400);
}

async function handleDelete(n: NotificationItem) {
  await playDeleteOut(document.querySelector(`[data-notification-id="${n.id}"]`));
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
  if (tideRunning.value) return;
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

/* ===== WebSocket 新消息实时入场：从右滑入列表顶部（TransitionGroup enter/move） ===== */
watch(
  () => store.notifications,
  (nv) => {
    if (loading.value || tideRunning.value) return;
    const first = nv[0];
    if (!first || list.value.some((x) => x.id === first.id)) return;
    if (unreadOnly.value && first.is_read) return;
    list.value = [first, ...list.value];
    total.value += 1;
  }
);

/* 角标数字平时与 store 同步（潮汐期间由滚动动画接管） */
watch(
  () => store.unreadCount,
  (v) => {
    if (!tideRunning.value) displayUnread.value = v;
  }
);

onMounted(() => {
  load(true);
  store.fetchUnreadCount();
  displayUnread.value = store.unreadCount;
});

onBeforeUnmount(() => {
  if (toastTimer) clearTimeout(toastTimer);
});
</script>

<template>
  <div class="max-w-4xl mx-auto">
    <!-- 顶部滑入 Toast（全部已读终结反馈） -->
    <transition name="notif-toast">
      <div
        v-if="toastVisible"
        class="fixed top-6 left-0 right-0 z-[70] flex justify-center pointer-events-none"
      >
        <div
          class="px-4 py-2 rounded-full bg-slate-900/90 dark:bg-white/90 text-white dark:text-slate-900 text-sm shadow-lg backdrop-blur-sm"
        >
          {{ toastMessage }}
        </div>
      </div>
    </transition>

    <div class="flex items-center justify-between mb-4">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">通知中心</h1>
          <span
            v-if="displayUnread > 0"
            class="inline-flex items-center h-5 px-2 rounded-full bg-blue-500 text-white text-[11px] font-semibold tabular-nums"
          >
            {{ displayUnread }}
          </span>
        </div>
        <p class="mt-0.5 text-xs text-slate-400">共 {{ total }} 条通知</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth disabled:opacity-60"
          :class="{ 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-300 border-blue-200 dark:border-blue-800': unreadOnly }"
          :disabled="tideRunning"
          @click="toggleUnreadOnly"
        >
          {{ unreadOnly ? '显示全部' : '只看未读' }}
        </button>
        <button
          v-if="store.unreadCount > 0 || tideRunning"
          class="px-3 py-1.5 text-sm rounded-lg bg-blue-500 text-white hover:bg-blue-600 transition-smooth inline-flex items-center gap-1.5 disabled:opacity-70"
          :disabled="tideRunning"
          @click="handleMarkAllRead"
        >
          <svg v-if="tideRunning" class="w-3.5 h-3.5 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
          </svg>
          {{ tideRunning ? '执行中...' : '全部已读' }}
        </button>
        <button
          class="px-3 py-1.5 text-sm rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-smooth"
          @click="router.push('/workbench')"
        >
          返回工作台
        </button>
      </div>
    </div>

    <div v-if="loading && list.length === 0" class="py-20 text-center text-sm text-slate-400">
      加载中...
    </div>

    <!-- 宁静空状态：缓慢呼吸插画 + 安心文案 -->
    <div v-else-if="list.length === 0" class="py-16 text-center">
      <div class="notif-empty-breathe inline-block">
        <svg class="w-20 h-20 mx-auto text-slate-300 dark:text-slate-600" viewBox="0 0 64 64" fill="none" aria-hidden="true">
          <path d="M14 26h28v14a10 10 0 01-10 10h-8a10 10 0 01-10-10V26z" fill="currentColor" opacity="0.3" />
          <path d="M14 26h28v14a10 10 0 01-10 10h-8a10 10 0 01-10-10V26z" stroke="currentColor" stroke-width="2.5" stroke-linejoin="round" />
          <path d="M42 30h4a6 6 0 010 12h-4" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" />
          <path d="M23 19c0-3 3-3 3-6M33 19c0-3 3-3 3-6" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" opacity="0.65" />
        </svg>
      </div>
      <p class="mt-3 text-sm text-slate-400 dark:text-slate-500">☕ 暂无新消息，安心工作吧</p>
    </div>

    <!-- 日期分组卡片流 -->
    <div v-else class="space-y-3">
      <div v-for="g in groups" :key="g.key">
        <!-- 粘性分组头：磨砂玻璃 + 装饰圆点 + 折叠箭头 -->
        <button
          class="notif-group-head w-full flex items-center gap-2 px-4 py-2.5 rounded-lg border border-slate-100 dark:border-slate-700/60 text-left"
          @click="toggleGroup(g.key)"
        >
          <svg
            class="notif-group-chevron w-3 h-3 text-slate-400 dark:text-slate-500"
            :class="{ collapsed: collapsedGroups.has(g.key) }"
            viewBox="0 0 12 12"
            fill="none"
            aria-hidden="true"
          >
            <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span class="w-1.5 h-1.5 rounded-full bg-blue-400/80 dark:bg-blue-400/60" aria-hidden="true" />
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-300 tracking-wide">{{ g.label }}</span>
          <span class="text-[11px] text-slate-400 dark:text-slate-500">{{ g.items.length }}</span>
          <span v-if="g.items.some((i) => !i.is_read)" class="text-[11px] text-blue-500 dark:text-blue-400">
            {{ g.items.filter((i) => !i.is_read).length }} 条未读
          </span>
        </button>

        <!-- 折叠容器（grid-template-rows 0fr ↔ 1fr） -->
        <div class="notif-collapse" :style="{ gridTemplateRows: collapsedGroups.has(g.key) ? '0fr' : '1fr' }">
          <div class="notif-collapse-inner">
            <TransitionGroup name="notif-list" tag="div" class="pt-2.5 space-y-2.5">
              <div
                v-for="n in g.items"
                :key="n.id"
                :data-notification-id="n.id"
                class="notif-item flex items-start gap-3 pl-5 pr-4 py-3.5 rounded-xl border cursor-pointer"
                :class="[
                  priorityClass(n),
                  n.is_read
                    ? 'notif-read bg-white dark:bg-slate-800 border-slate-100 dark:border-slate-700'
                    : 'notif-unread bg-white/75 dark:bg-slate-800/60 border-slate-100 dark:border-slate-700/60',
                  { 'notif-reading': readingIds.has(n.id), 'notif-tide': tideIds.has(n.id) },
                ]"
                :style="tideDelayStyle(n)"
                @click="onItemClick(n)"
              >
                <!-- 左侧 4px 能量竖条 -->
                <span class="notif-bar" aria-hidden="true" />
                <!-- 呼吸脉冲圆点（未读）/ 灰色对勾（已读） -->
                <span v-if="!n.is_read" class="notif-dot mt-1.5" aria-hidden="true" />
                <span v-else class="mt-1 flex items-center shrink-0" aria-hidden="true">
                  <svg
                    class="w-3.5 h-3.5 text-gray-300 dark:text-gray-500"
                    :class="{ 'notif-check': animatedChecks.has(n.id) }"
                    viewBox="0 0 16 16"
                    fill="none"
                  >
                    <path d="M3 8.5L6.5 12L13 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </span>

                <span
                  class="mt-0.5 shrink-0 px-2 py-0.5 rounded text-[11px] leading-4 font-medium"
                  :class="TYPE_COLOR[n.type] || TYPE_COLOR.system"
                >
                  {{ TYPE_LABEL[n.type] || '通知' }}
                </span>

                <div class="flex-1 min-w-0">
                  <div class="flex items-start justify-between gap-2">
                    <p class="notif-title text-sm font-medium text-slate-800 dark:text-slate-100 truncate">
                      {{ n.title }}
                    </p>
                    <span class="notif-time shrink-0 text-[11px] text-slate-400" :class="{ 'notif-time-unread': !n.is_read }">
                      {{ formatTime(n.created_at) }}
                    </span>
                  </div>
                  <p class="notif-content mt-0.5 text-xs text-slate-500 dark:text-slate-400 line-clamp-2 break-all">
                    {{ n.content }}
                  </p>
                </div>

                <div class="shrink-0 flex items-center gap-1.5">
                  <button
                    v-if="n.note_id"
                    class="px-2 py-1 text-[11px] rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-300 hover:bg-blue-100 transition-smooth"
                    @click.stop="goToNote(n)"
                  >
                    查看任务
                  </button>
                  <button
                    v-if="!n.is_read"
                    class="px-2 py-1 text-[11px] rounded-lg border border-slate-200 dark:border-slate-600 text-slate-400 hover:text-blue-500 hover:border-blue-300 transition-smooth"
                    @click.stop="markOneRead(n)"
                  >
                    标为已读
                  </button>
                  <button
                    class="p-1 text-slate-300 hover:text-red-400 transition-smooth"
                    title="删除"
                    @click.stop="handleDelete(n)"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </TransitionGroup>
          </div>
        </div>
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
    </transition>
  </div>
</template>
