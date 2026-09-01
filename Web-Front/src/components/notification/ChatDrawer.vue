<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue';
import { useNotificationStore } from '@/stores/notification';
import { useAuthStore } from '@/stores/auth';
import { getVisibleUsers } from '@/services/admin';
import type { ConversationItem, ChatMessageItem, User } from '@/types';
import { renderNoteContent } from '@/utils/richText';

const props = defineProps<{ visible: boolean }>();
const emit = defineEmits<{ 'update:visible': [value: boolean] }>();

const store = useNotificationStore();
const auth = useAuthStore();

const loading = ref(false);
const currentPeer = ref<string | null>(null);
const input = ref('');
const nameMap = ref<Record<string, string>>({});
const messagesEl = ref<HTMLDivElement | null>(null);

async function loadUsers() {
  try {
    const res = await getVisibleUsers();
    const users = res.data;
    const map: Record<string, string> = {};
    users.forEach((u) => {
      map[u.id] = u.name || u.username;
    });
    nameMap.value = map;
  } catch {
    /* ignore */
  }
}

function peerName(id: string) {
  return nameMap.value[id] || `用户 ${id.slice(0, 6)}`;
}

async function openConversation(conv: ConversationItem) {
  currentPeer.value = conv.peer_id;
  store.setViewingPeer(conv.peer_id);
  store.loadMessages(conv.peer_id).finally(async () => {
    await store.markConversationRead(conv.peer_id);
    scrollToBottom();
  });
}

async function openConversationById(peerId: string) {
  const conv = store.conversations.find((c) => c.peer_id === peerId);
  if (conv) {
    await openConversation(conv);
  } else {
    currentPeer.value = peerId;
    store.setViewingPeer(peerId);
    await store.loadMessages(peerId);
    scrollToBottom();
  }
}

async function send() {
  const content = input.value.trim();
  if (!content || !currentPeer.value) return;
  input.value = '';
  await store.sendMessage(currentPeer.value, { content });
  await nextTick();
  scrollToBottom();
}

function scrollToBottom() {
  nextTick(() => {
    messagesEl.value?.scrollTo({ top: messagesEl.value.scrollHeight, behavior: 'smooth' });
  });
}

function formatTime(t?: string) {
  if (!t) return '';
  const d = new Date(t);
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  const hm = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  return sameDay ? hm : `${d.toLocaleDateString('zh-CN')} ${hm}`;
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
  return '📄';
}

// 实时新消息滚动
watch(
  () => store.messages[currentPeer.value || ''],
  () => {
    if (currentPeer.value) scrollToBottom();
  }
);

watch(
  () => props.visible,
  (v) => {
    if (v) {
      loadUsers();
      store.refreshConversations().then(() => {
        // 从任务详情跳转过来的指定会话
        if (store.chatPeerId && store.chatPeerId !== currentPeer.value) {
          openConversationById(store.chatPeerId);
        }
      });
    } else {
      // 关闭抽屉：清除正在查看的会话，后续消息正常计入未读角标
      store.setViewingPeer(null);
    }
  }
);

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <transition name="drawer-fade">
    <div v-if="visible" class="fixed inset-0 z-[70] bg-black/30" @click="emit('update:visible', false)" />
  </transition>
  <transition name="drawer-slide">
    <div
      v-if="visible"
      class="fixed top-0 right-0 h-full w-[420px] max-w-[100vw] bg-white dark:bg-slate-800 shadow-2xl z-[71] flex flex-col"
    >
      <!-- 头部 -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700">
        <h3 class="text-base font-semibold text-slate-800 dark:text-slate-100">消息聊天</h3>
        <button
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
          @click="emit('update:visible', false)"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- 会话列表 -->
      <div class="border-b border-slate-100 dark:border-slate-700 max-h-[200px] overflow-y-auto scrollbar-thin">
        <div v-if="store.conversations.length === 0" class="px-4 py-6 text-center text-sm text-slate-400">
          暂无会话，点击任务中的成员头像即可发起聊天
        </div>
        <div
          v-for="conv in store.conversations"
          :key="conv.peer_id"
          class="px-4 py-2.5 flex items-center gap-3 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-smooth"
          :class="{ 'bg-blue-50/70 dark:bg-blue-900/10': currentPeer === conv.peer_id }"
          @click="openConversation(conv)"
        >
          <div
            class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-sm font-medium flex items-center justify-center shrink-0"
          >
            {{ peerName(conv.peer_id).slice(0, 1) }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-slate-700 dark:text-slate-200 truncate">{{ peerName(conv.peer_id) }}</p>
            <p class="text-xs text-slate-400 truncate">{{ conv.last_msg }}</p>
          </div>
          <div class="shrink-0 text-right">
            <span v-if="conv.unread > 0" class="inline-block min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold leading-[18px] text-center">
              {{ conv.unread > 99 ? '99+' : conv.unread }}
            </span>
          </div>
        </div>
      </div>

      <!-- 消息区 -->
      <div ref="messagesEl" class="flex-1 overflow-y-auto scrollbar-thin px-4 py-3 space-y-3 bg-slate-50 dark:bg-slate-900/40">
        <div v-if="!currentPeer" class="h-full flex flex-col items-center justify-center text-sm text-slate-400">
          <svg class="w-12 h-12 mb-2 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <p>选择一个会话开始聊天</p>
        </div>
        <template v-else>
          <div
            v-for="m in store.messages[currentPeer] || []"
            :key="m.id"
            class="flex"
            :class="m.sender_id === auth.user?.id ? 'justify-end' : 'justify-start'"
          >
            <div class="max-w-[75%]">
              <!-- 文本 -->
              <div
                v-if="m.type !== 'image' && m.type !== 'file'"
                class="px-3 py-2 rounded-2xl text-sm leading-relaxed break-words"
                :class="
                  m.sender_id === auth.user?.id
                    ? 'bg-blue-500 text-white rounded-br-sm'
                    : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 rounded-bl-sm'
                "
                v-html="renderNoteContent(m.content)"
              />
              <!-- 图片 -->
              <img
                v-else-if="m.type === 'image'"
                :src="m.file_path"
                :alt="m.file_name || '图片'"
                class="max-w-[200px] max-h-[260px] rounded-xl object-cover border border-slate-200 dark:border-slate-600"
                loading="lazy"
              />
              <!-- 文件 -->
              <a
                v-else-if="m.type === 'file'"
                :href="m.file_path"
                download
                target="_blank"
                rel="noopener"
                class="flex items-center gap-2 px-3 py-2 rounded-2xl text-sm bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 hover:shadow-md transition-smooth"
                :class="m.sender_id === auth.user?.id ? 'rounded-br-sm' : 'rounded-bl-sm'"
              >
                <span class="text-xl shrink-0">{{ fileIcon(m.file_name) }}</span>
                <div class="min-w-0">
                  <p class="text-xs font-medium text-slate-800 dark:text-slate-100 truncate max-w-[130px]">{{ m.file_name }}</p>
                  <p class="text-[10px] text-slate-400">{{ formatSize(m.file_size) }}</p>
                </div>
              </a>
              <p class="mt-1 text-[10px] flex items-center gap-1.5 px-1" :class="m.sender_id === auth.user?.id ? 'justify-end' : ''">
                <span class="text-slate-400">{{ formatTime(m.created_at) }}</span>
                <span
                  v-if="m.sender_id === auth.user?.id"
                  :class="m.is_read ? 'text-blue-500 dark:text-blue-400' : 'text-slate-400 dark:text-slate-500'"
                >{{ m.is_read ? '已读' : '未读' }}</span>
              </p>
            </div>
          </div>
        </template>
      </div>

      <!-- 输入区 -->
      <div v-if="currentPeer" class="border-t border-slate-100 dark:border-slate-700 p-3 bg-white dark:bg-slate-800">
        <div class="flex items-end gap-2">
          <textarea
            v-model="input"
            rows="2"
            class="flex-1 resize-none rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50"
            placeholder="输入消息内容，Enter 发送，Shift+Enter 换行"
            @keydown.enter.exact.prevent="send"
          />
          <button
            class="shrink-0 px-4 py-2 rounded-lg bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-smooth disabled:opacity-50"
            :disabled="!input.trim()"
            @click="send"
          >
            发送
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}
.drawer-slide-enter-active {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
/* 会话关闭：从右下角收缩为一个小点（300ms），模拟“收起对话”的物理感 */
.drawer-slide-leave-active {
  transition:
    transform 0.3s cubic-bezier(0.4, 0, 1, 1),
    opacity 0.3s ease-in;
  transform-origin: bottom right;
}
.drawer-slide-enter-from {
  transform: translateX(100%);
}
.drawer-slide-leave-to {
  opacity: 0;
  transform: translate(38%, 38%) scale(0);
}
</style>
