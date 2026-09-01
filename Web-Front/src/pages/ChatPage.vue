<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import { useAuthStore } from '@/stores/auth';
import { getVisibleUsers } from '@/services/admin';
import { uploadChatFile } from '@/services/notification';
import EmojiPicker from '@/components/chat/EmojiPicker.vue';
import type { ChatMessageItem, User } from '@/types';
import { renderNoteContent } from '@/utils/richText';
import { matchPinyin } from '@/utils/pinyin';

const store = useNotificationStore();
const auth = useAuthStore();
const route = useRoute();

const leftTab = ref<'conv' | 'contact'>('conv');
const keyword = ref('');
const currentPeer = ref<string | null>(null);
const input = ref('');
const users = ref<User[]>([]);
const messagesEl = ref<HTMLDivElement | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const uploading = ref(false);
const showEmoji = ref(false);
// 需求33：表情面板元素引用，用于「点击页面其他地方关闭」
const emojiBtnEl = ref<HTMLElement | null>(null);
const emojiPanelEl = ref<HTMLElement | null>(null);
const previewUrl = ref('');
const loadingOlder = ref(false);
const allLoaded = ref(false);
const page = ref(1);
const PAGE_SIZE = 30;
// 用户是否已上滑查看历史消息（不在底部）：此时收到新消息不自动下拉，显示浮动按钮
const scrolledUp = ref(false);
// 上滑查看历史期间收到的新消息数（浮动按钮上的角标）
const newMsgCount = ref(0);

// 好友姓名映射（会话中对方姓名兜底）
const nameMap = computed(() => {
  const map: Record<string, string> = {};
  users.value.forEach((u) => {
    map[u.id] = u.name || u.username;
  });
  return map;
});

const filteredUsers = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  const me = auth.user?.id;
  let list = users.value.filter((u) => u.id !== me && u.is_active !== false);
  if (kw) {
    // 需求36：支持拼音全拼 / 首字母搜索（无视大小写）
    list = list.filter((u) => matchPinyin(kw, u.name || '', u.username || '', u.dept_name || ''));
  }
  // 在线用户优先展示（仿微信通讯录）
  return [...list].sort((a, b) => Number(store.isOnline(b.id)) - Number(store.isOnline(a.id)));
});

const currentPeerName = computed(() => {
  if (!currentPeer.value) return '';
  const conv = store.conversations.find((c) => c.peer_id === currentPeer.value);
  if (conv?.peer_name) return conv.peer_name;
  return nameMap.value[currentPeer.value] || `用户 ${currentPeer.value.slice(0, 6)}`;
});

const currentMessages = computed(() => store.messages[currentPeer.value || ''] || []);

async function loadUsers() {
  try {
    const res = await getVisibleUsers();
    users.value = res.data || [];
  } catch {
    /* ignore */
  }
}

function peerName(id: string) {
  return nameMap.value[id] || `用户 ${id.slice(0, 6)}`;
}

async function openConversation(peerId: string, name?: string) {
  currentPeer.value = peerId;
  page.value = 1;
  allLoaded.value = false;
  showEmoji.value = false;
  scrolledUp.value = false;
  newMsgCount.value = 0;
  store.setViewingPeer(peerId);
  store.fetchOnlineUsers();
  await store.loadMessages(peerId);
  await store.markConversationRead(peerId);
  scrollToBottom();
}

function scrollToBottom(smooth = true) {
  const el = messagesEl.value;
  if (!el) return;
  const scroll = () => {
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' });
  };
  nextTick(scroll);
  // 图片等异步内容加载会撑高容器，加载完成后补一次，确保停在最新消息处
  setTimeout(scroll, 200);
  el.querySelectorAll('img').forEach((img) => {
    if (!img.complete) img.addEventListener('load', scroll, { once: true });
  });
}

// 消息区滚动：记录是否在底部；滚动到顶部时加载更早消息
function handleMessagesScroll(e: Event) {
  const el = e.target as HTMLElement;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  scrolledUp.value = !atBottom;
  if (atBottom) newMsgCount.value = 0;
  if (el.scrollTop < 40) loadOlder();
}

// 浮动按钮：一键回到最新消息
function goToLatest() {
  scrolledUp.value = false;
  newMsgCount.value = 0;
  scrollToBottom();
}

// 加载更早消息
async function loadOlder() {
  if (loadingOlder.value || allLoaded.value || !currentPeer.value) return;
  loadingOlder.value = true;
  const next = page.value + 1;
  try {
    const res = await store.fetchMessagesPage(currentPeer.value, next, PAGE_SIZE);
    const list = (res as unknown as { data: ChatMessageItem[]; total: number }) || { data: [] };
    const older = list.data || [];
    if (older.length < PAGE_SIZE) allLoaded.value = true;
    if (older.length > 0) {
      const prev = store.messages[currentPeer.value] || [];
      store.messages[currentPeer.value] = [...older, ...prev];
      page.value = next;
    } else {
      allLoaded.value = true;
    }
  } catch {
    /* ignore */
  } finally {
    loadingOlder.value = false;
  }
}

async function sendText() {
  const content = input.value.trim();
  if (!content || !currentPeer.value) return;
  input.value = '';
  showEmoji.value = false;
  await store.sendMessage(currentPeer.value, { content });
  scrollToBottom();
}

function insertEmoji(e: string) {
  input.value += e;
}

// 发送图片表情（图片消息）
async function sendEmoticon(path: string) {
  if (!currentPeer.value || !path) return;
  showEmoji.value = false;
  await store.sendMessage(currentPeer.value, {
    type: 'image',
    file_name: '表情',
    file_path: path,
    mime_type: 'image/png',
  });
  scrollToBottom();
}

async function onFileSelected(e: Event) {
  const inputEl = e.target as HTMLInputElement;
  const file = inputEl.files?.[0];
  inputEl.value = '';
  if (!file || !currentPeer.value) return;

  if (file.size > 10 * 1024 * 1024) {
    alert('文件大小超过限制，最大允许 10MB');
    return;
  }
  uploading.value = true;
  try {
    const res = await uploadChatFile(file);
    const meta = res.data as unknown as { file_name: string; file_path: string; file_size: number; mime_type: string };
    const isImage = (meta.mime_type || '').startsWith('image/');
    await store.sendMessage(currentPeer.value, {
      type: isImage ? 'image' : 'file',
      file_name: meta.file_name,
      file_path: meta.file_path,
      file_size: meta.file_size,
      mime_type: meta.mime_type,
    });
    scrollToBottom();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    alert(e2?.response?.data?.message || '文件上传失败');
  } finally {
    uploading.value = false;
  }
}

function formatTime(ts: string): string {
  if (!ts) return '';
  const d = new Date(ts);
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  const hm = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  if (sameDay) return hm;
  const yesterday = new Date(now.getTime() - 86400000);
  if (d.toDateString() === yesterday.toDateString()) return `昨天 ${hm}`;
  return `${d.toLocaleDateString('zh-CN')} ${hm}`;
}

function formatSize(bytes?: number): string {
  if (!bytes) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function fileIcon(name?: string): string {
  const ext = (name || '').split('.').pop()?.toLowerCase() || '';
  if (['png', 'jpg', 'jpeg', 'gif', 'webp'].includes(ext)) return '🖼';
  if (['pdf'].includes(ext)) return '📕';
  if (['doc', 'docx'].includes(ext)) return '📘';
  if (['xls', 'xlsx'].includes(ext)) return '📗';
  if (['zip', 'rar', '7z'].includes(ext)) return '🗜';
  return '📄';
}

// 实时新消息滚动：在底部自动下拉；查看历史时只累加角标（已读标记由 store 在收到消息时即时处理）
watch(
  () => store.messages[currentPeer.value || ''],
  (list, oldList) => {
    if (!currentPeer.value) return;
    // bug5：正在查看历史消息时不自动下拉，仅显示浮动按钮；在底部对话中才自动下拉
    if (!scrolledUp.value) {
      scrollToBottom();
    } else if (list && oldList && list.length > oldList.length) {
      // 上滑查看历史期间收到新消息 → 只累加角标，不打断阅读位置
      const last = list[list.length - 1];
      if (last && last.sender_id !== auth.user?.id) {
        newMsgCount.value++;
      }
    }
  }
);

onMounted(() => {
  store.connectSocket();
  loadUsers();
  store.refreshConversations();
  store.fetchOnlineUsers();
  // 需求24：弹窗点击跳转 /chat?peer=xxx 时直达对应会话
  const peer = route.query.peer as string | undefined;
  if (peer) openConversation(peer);
  // 需求33：点击表情面板之外的区域关闭面板
  document.addEventListener('mousedown', handleOutsideClick);
});

// 需求33：点击页面其他地方关闭表情面板（面板与触发按钮内部不关闭）
function handleOutsideClick(e: MouseEvent) {
  if (!showEmoji.value) return;
  const t = e.target as Node;
  if (emojiPanelEl.value?.contains(t)) return;
  if (emojiBtnEl.value?.contains(t)) return;
  showEmoji.value = false;
}

// 需求24：已在聊天页时，点击弹窗切换会话（peer 查询参数变化）
watch(
  () => route.query.peer,
  (peer) => {
    if (peer && peer !== currentPeer.value) openConversation(peer as string);
  }
);

onUnmounted(() => {
  // 离开聊天页：清除正在查看的会话，后续消息正常计入未读角标
  store.setViewingPeer(null);
  // 需求33：移除全局点击监听
  document.removeEventListener('mousedown', handleOutsideClick);
});
</script>

<template>
  <div class="h-full flex rounded-card overflow-hidden border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800">
    <!-- ============ 左侧栏 ============ -->
    <aside class="w-[300px] shrink-0 flex flex-col border-r border-slate-100 dark:border-slate-700 bg-slate-50/60 dark:bg-slate-900/40">
      <!-- 头部 -->
      <div class="px-4 py-3.5 border-b border-slate-100 dark:border-slate-700">
        <div class="flex items-center justify-between mb-2.5">
          <h2 class="text-base font-semibold text-slate-800 dark:text-slate-100">💬 聊天</h2>
          <span class="text-[10px] px-2 py-0.5 rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 font-medium">
            {{ store.connected ? '在线' : '连接中' }}
          </span>
        </div>
        <input
          v-model="keyword"
          type="text"
          class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-400"
          :placeholder="leftTab === 'conv' ? '搜索会话...' : '搜索姓名 / 部门...'"
        />
      </div>

      <!-- Tab 切换 -->
      <div class="flex border-b border-slate-100 dark:border-slate-700">
        <button
          class="flex-1 py-2.5 text-sm font-medium transition-smooth relative"
          :class="leftTab === 'conv' ? 'text-blue-500' : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="leftTab = 'conv'"
        >
          会话
          <span class="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full transition-smooth"
            :class="leftTab === 'conv' ? 'bg-blue-500' : 'bg-transparent'" />
        </button>
        <button
          class="flex-1 py-2.5 text-sm font-medium transition-smooth relative"
          :class="leftTab === 'contact' ? 'text-blue-500' : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="leftTab = 'contact'"
        >
          通讯录
          <span class="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full transition-smooth"
            :class="leftTab === 'contact' ? 'bg-blue-500' : 'bg-transparent'" />
        </button>
      </div>

      <!-- 会话列表 -->
      <div v-if="leftTab === 'conv'" class="flex-1 overflow-y-auto scrollbar-thin">
        <div v-if="store.conversations.length === 0" class="px-4 py-10 text-center">
          <p class="text-3xl mb-2">💬</p>
          <p class="text-xs text-slate-400 dark:text-slate-500">暂无会话</p>
          <p class="text-[10px] text-slate-300 dark:text-slate-600 mt-1">切换到「通讯录」选择一个好友开始聊天</p>
        </div>
        <div
          v-for="conv in store.conversations"
          :key="conv.peer_id"
          class="px-3 py-2.5 flex items-center gap-2.5 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/40 transition-smooth"
          :class="{ 'bg-blue-50 dark:bg-blue-900/10': currentPeer === conv.peer_id }"
          @click="openConversation(conv.peer_id, conv.peer_name)"
        >
          <div class="relative w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ (conv.peer_name || peerName(conv.peer_id)).slice(0, 1) }}
            <span
              v-if="store.isOnline(conv.peer_id)"
              class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full bg-green-500 border-2 border-white dark:border-slate-800"
            ></span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">{{ conv.peer_name || peerName(conv.peer_id) }}</p>
              <span class="text-[10px] text-slate-400 shrink-0">{{ formatTime(conv.last_at) }}</span>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate mt-0.5">{{ conv.last_msg }}</p>
          </div>
          <span v-if="conv.unread > 0" class="shrink-0 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold leading-[18px] text-center">
            {{ conv.unread > 99 ? '99+' : conv.unread }}
          </span>
        </div>
      </div>

      <!-- 通讯录 -->
      <div v-else class="flex-1 overflow-y-auto scrollbar-thin">
        <div v-if="filteredUsers.length === 0" class="px-4 py-10 text-center text-xs text-slate-400 dark:text-slate-500">
          {{ keyword ? '未找到匹配的好友' : '暂无其他用户' }}
        </div>
        <div
          v-for="u in filteredUsers"
          :key="u.id"
          class="px-3 py-2.5 flex items-center gap-2.5 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/40 transition-smooth"
          @click="openConversation(u.id, u.name || u.username)"
        >
          <div class="relative w-10 h-10 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ (u.name || u.username).slice(0, 1) }}
            <span
              class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-slate-800"
              :class="store.isOnline(u.id) ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-600'"
            ></span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">{{ u.name || u.username }}</p>
              <span class="text-[10px] shrink-0" :class="store.isOnline(u.id) ? 'text-green-500' : 'text-slate-400 dark:text-slate-500'">
                {{ store.isOnline(u.id) ? '在线' : '离线' }}
              </span>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ u.dept_name || u.role }}</p>
          </div>
        </div>
      </div>
    </aside>

    <!-- ============ 主区 ============ -->
    <main class="flex-1 flex flex-col min-w-0 bg-slate-50 dark:bg-slate-900/50">
      <!-- 空态 -->
      <div v-if="!currentPeer" class="flex-1 flex flex-col items-center justify-center">
        <div class="w-16 h-16 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center text-3xl mb-3">💬</div>
        <p class="text-sm text-slate-400 dark:text-slate-500">选择一个好友开始聊天</p>
        <p class="text-xs text-slate-300 dark:text-slate-600 mt-1">支持文本、emoji 表情、图片与文件传输</p>
      </div>

      <template v-else>
        <!-- 顶部信息条 -->
        <div class="h-14 shrink-0 flex items-center gap-3 px-5 border-b border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-800">
          <div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ currentPeerName.slice(0, 1) }}
          </div>
          <div class="min-w-0">
            <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ currentPeerName }}</p>
            <p class="text-[10px] flex items-center gap-1" :class="store.isOnline(currentPeer) ? 'text-green-500' : 'text-slate-400'">
              <span class="inline-block w-1.5 h-1.5 rounded-full" :class="store.isOnline(currentPeer) ? 'bg-green-500' : 'bg-slate-400'"></span>
              {{ store.isOnline(currentPeer) ? '在线' : '离线' }}
            </p>
          </div>
        </div>

        <!-- 消息区 -->
        <div class="relative flex-1 flex flex-col min-h-0">
          <div
            ref="messagesEl"
            class="flex-1 overflow-y-auto scrollbar-thin px-6 py-4 space-y-3"
            @scroll.passive="handleMessagesScroll"
          >
            <div v-if="loadingOlder" class="text-center text-xs text-slate-400 py-1">加载更早消息...</div>
            <div v-if="allLoaded && currentMessages.length > 0" class="text-center text-[10px] text-slate-300 dark:text-slate-600 py-1">—— 已显示全部消息 ——</div>

            <div
              v-for="m in currentMessages"
              :key="m.id"
              class="flex items-end gap-2"
              :class="m.sender_id === auth.user?.id ? 'justify-end' : 'justify-start'"
            >
            <!-- 对方头像 -->
            <div v-if="m.sender_id !== auth.user?.id" class="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 text-white text-xs font-medium flex items-center justify-center shrink-0 mb-1">
              {{ peerName(m.sender_id).slice(0, 1) }}
            </div>

            <div class="max-w-[70%] min-w-0">
              <!-- 文本 -->
              <div
                v-if="m.type === 'text'"
                class="px-3.5 py-2 rounded-2xl text-sm leading-relaxed break-words rich-content-display"
                :class="m.sender_id === auth.user?.id
                  ? 'bg-blue-500 text-white rounded-br-sm'
                  : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 rounded-bl-sm'"
                v-html="renderNoteContent(m.content)"
              ></div>

              <!-- 图片 -->
              <div v-else-if="m.type === 'image'" class="min-w-[120px]">
                <img
                  :src="m.file_path"
                  :alt="m.file_name || '图片'"
                  class="max-w-[240px] max-h-[300px] rounded-xl cursor-pointer object-cover shadow-sm border border-slate-200 dark:border-slate-600"
                  loading="lazy"
                  @click="previewUrl = m.file_path || ''"
                />
              </div>

              <!-- 文件 -->
              <div v-else-if="m.type === 'file'" class="min-w-[200px]">
                <a
                  :href="m.file_path"
                  download
                  target="_blank"
                  rel="noopener"
                  class="flex items-center gap-3 px-3.5 py-2.5 rounded-2xl bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 hover:shadow-md transition-smooth"
                  :class="m.sender_id === auth.user?.id ? 'rounded-br-sm' : 'rounded-bl-sm'"
                >
                  <span class="text-2xl shrink-0">{{ fileIcon(m.file_name) }}</span>
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-slate-800 dark:text-slate-100 truncate max-w-[180px]">{{ m.file_name }}</p>
                    <p class="text-[10px] text-slate-400">{{ formatSize(m.file_size) }} · 点击下载</p>
                  </div>
                </a>
              </div>

              <p class="mt-1 text-[10px] flex items-center gap-1.5 px-1" :class="m.sender_id === auth.user?.id ? 'justify-end' : ''">
                <span class="text-slate-400">{{ formatTime(m.created_at) }}</span>
                <span
                  v-if="m.sender_id === auth.user?.id"
                  :class="m.is_read ? 'text-blue-500 dark:text-blue-400' : 'text-slate-400 dark:text-slate-500'"
                >{{ m.is_read ? '已读' : '未读' }}</span>
              </p>
            </div>
          </div>
        </div>
          <!-- 浮动按钮：上滑查看历史消息时显示，点击回到最新消息 -->
          <button
            v-if="scrolledUp"
            class="absolute bottom-4 right-4 z-10 flex items-center gap-1.5 px-3 py-2 rounded-full bg-blue-500 text-white text-xs font-medium shadow-lg hover:bg-blue-600 active:scale-95 transition-smooth"
            @click="goToLatest"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7-7-7m14-8l-7 7-7-7" />
            </svg>
            <span v-if="newMsgCount > 0" class="min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center">{{ newMsgCount }}</span>
            <span>最新消息</span>
          </button>
        </div>

        <!-- 输入区 -->
        <div class="relative shrink-0 border-t border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-800 p-3">
          <!-- 工具栏 -->
          <div class="flex items-center gap-1 mb-2">
            <button
              ref="emojiBtnEl"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              title="表情"
              @click="showEmoji = !showEmoji"
            >😊</button>
            <button
              class="w-8 h-8 rounded-lg flex items-center justify-center text-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              title="发送文件（支持绝大部分格式，禁止可执行/脚本文件）"
              :disabled="uploading"
              @click="fileInput?.click()"
            >📎</button>
            <span v-if="uploading" class="text-xs text-slate-400 ml-1">上传中...</span>
            <input ref="fileInput" type="file" class="hidden" @change="onFileSelected" />
          </div>

          <!-- emoji 面板 -->
          <transition name="shrink-out">
          <div
            v-if="showEmoji"
            ref="emojiPanelEl"
            class="absolute bottom-full left-3 mb-2 z-20"
            @click.stop
          >
            <EmojiPicker @select="insertEmoji" @send-image="sendEmoticon" />
          </div>
          </transition>

          <div class="flex items-end gap-2">
            <textarea
              v-model="input"
              rows="2"
              class="flex-1 resize-none rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3.5 py-2 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 scrollbar-thin"
              placeholder="输入消息，Enter 发送，Shift+Enter 换行"
              @keydown.enter.exact.prevent="sendText"
              @focus="showEmoji = false"
            />
            <button
              class="shrink-0 px-5 py-2 rounded-xl bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-smooth disabled:opacity-50"
              :disabled="!input.trim()"
              @click="sendText"
            >
              发送
            </button>
          </div>
        </div>
      </template>
    </main>

    <!-- 图片预览 -->
    <Teleport to="body">
      <transition name="shrink-out">
      <div v-if="previewUrl" class="fixed inset-0 z-[90] flex items-center justify-center bg-black/70 p-6" @click="previewUrl = ''">
        <img :src="previewUrl" class="max-w-[90vw] max-h-[90vh] rounded-lg object-contain" />
        <button class="absolute top-4 right-4 w-10 h-10 rounded-full bg-black/40 text-white text-xl flex items-center justify-center hover:bg-black/60 transition-smooth">✕</button>
      </div>
      </transition>
    </Teleport>
  </div>
</template>
