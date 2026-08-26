<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import type { PopupItem } from '@/stores/notification';

const store = useNotificationStore();
const router = useRouter();

// 点击弹窗跳转：聊天消息 → 聊天会话页；任务通知 → 工作台任务详情；issue 评论 → 问题详情；其他 → 通知中心
function onPopupClick(p: PopupItem) {
  if (p.kind === 'chat' && p.peerId) {
    router.push({ path: '/chat', query: { peer: p.peerId } });
  } else if (p.kind === 'notification' && p.noteId) {
    router.push({ path: '/workbench', query: { note: p.noteId } });
  } else if (p.kind === 'notification' && p.issueId) {
    router.push(`/issues/${p.issueId}`);
  } else {
    router.push('/notifications');
  }
  store.dismissPopup(p.id);
}

function formatTime(ts?: string): string {
  if (!ts) return '';
  const d = new Date(ts);
  const now = new Date();
  const hm = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  if (d.toDateString() === now.toDateString()) return hm;
  return `${d.toLocaleDateString('zh-CN')} ${hm}`;
}
</script>

<template>
  <div class="fixed top-4 right-4 z-[120] flex flex-col items-end gap-2 pointer-events-none">
    <TransitionGroup name="float-pop">
      <div
        v-for="p in store.popups"
        :key="p.id"
        class="pointer-events-auto w-[320px] max-w-[calc(100vw-2rem)] bg-white dark:bg-slate-800 rounded-xl shadow-modal border border-slate-100 dark:border-slate-700 p-3.5 flex items-start gap-3 cursor-pointer hover:shadow-xl transition-smooth"
        @click="onPopupClick(p)"
      >
        <div
          class="w-9 h-9 rounded-full flex items-center justify-center text-base shrink-0"
          :class="
            p.kind === 'chat'
              ? 'bg-emerald-100 dark:bg-emerald-900/50'
              : 'bg-blue-100 dark:bg-blue-900/50'
          "
        >
          {{ p.kind === 'chat' ? '💬' : '🔔' }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between gap-2">
            <span class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{
              p.title
            }}</span>
            <span class="text-[10px] text-slate-400 shrink-0">{{ formatTime(p.createdAt) }}</span>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2 whitespace-pre-line">
            {{ p.content || '…' }}
          </p>
        </div>
        <button
          class="shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-slate-300 dark:text-slate-600 hover:text-slate-500 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
          title="关闭"
          @click.stop="store.dismissPopup(p.id)"
        >
          <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
/* 从右上角丝滑弹出：弹性缓动（带轻微回弹） */
.float-pop-enter-active {
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.float-pop-leave-active {
  transition: all 0.25s ease;
}
.float-pop-enter-from {
  opacity: 0;
  transform: translateX(120%) scale(0.9);
}
.float-pop-leave-to {
  opacity: 0;
  transform: translateX(120%) scale(0.9);
}
.float-pop-move {
  transition: transform 0.3s ease;
}
</style>
