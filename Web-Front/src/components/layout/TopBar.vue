<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import DarkToggle from '@/components/common/DarkToggle.vue'
import NotificationBell from '@/components/notification/NotificationBell.vue'
import ChatDrawer from '@/components/notification/ChatDrawer.vue'

const route = useRoute()
const auth = useAuthStore()
const notifStore = useNotificationStore()

const currentTime = ref('')
let timer: ReturnType<typeof setInterval>

const breadcrumbs = computed(() => {
  timer = setInterval(() => {
    currentTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }, 1000)
  currentTime.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  return route.matched
    .filter((r) => r.meta?.title)
    .map((r) => ({ path: r.path, title: r.meta.title as string }))
})

onMounted(() => {
  // 登录后建立用户个人通知 WebSocket 通道
  notifStore.connectSocket()
})

onUnmounted(() => clearInterval(timer))
</script>

<template>
  <header class="h-14 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 flex items-center px-6 shrink-0 gap-3 transition-colors duration-300">
    <!-- 面包屑 -->
    <div class="flex items-center gap-2 text-sm">
      <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
        <span v-if="index > 0" class="text-slate-300 dark:text-slate-600">/</span>
        <span
          :class="[
            'transition-smooth',
            index === breadcrumbs.length - 1
              ? 'text-slate-900 dark:text-slate-100 font-medium'
              : 'text-slate-400 dark:text-slate-500'
          ]"
        >
          {{ crumb.title }}
        </span>
      </template>
    </div>

    <div class="flex-1" />

    <!-- 时间显示 -->
    <span class="text-xs text-slate-400 dark:text-slate-500 tabular-nums font-mono hidden sm:block">{{ currentTime }}</span>

    <!-- 聊天入口 -->
    <button
      class="relative p-2 rounded-lg text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth"
      v-tooltip="'消息聊天'"
      @click="notifStore.openChat()"
    >
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
        />
      </svg>
      <span
        v-if="notifStore.conversations.reduce((sum, c) => sum + (c.unread || 0), 0) > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-amber-500 text-white text-[10px] font-semibold flex items-center justify-center animate-breathe"
      >
        {{ notifStore.conversations.reduce((sum, c) => sum + (c.unread || 0), 0) > 99 ? '99+' : notifStore.conversations.reduce((sum, c) => sum + (c.unread || 0), 0) }}
      </span>
    </button>

    <!-- 通知铃铛 -->
    <NotificationBell />

    <!-- 夜间模式切换 -->
    <DarkToggle />

    <!-- 聊天抽屉 -->
    <ChatDrawer v-model:visible="notifStore.chatOpen" />
  </header>
</template>
