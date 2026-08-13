import { defineStore } from 'pinia';
import { ref } from 'vue';
import * as notifService from '@/services/notification';
import type { NotificationItem, ChatMessageItem, ConversationItem } from '@/types';
import { playNotificationSound, playChatSound } from '@/utils/sound';
import { useAuthStore } from './auth';

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref(0);
  const notifications = ref<NotificationItem[]>([]);
  const conversations = ref<ConversationItem[]>([]);
  const messages = ref<Record<string, ChatMessageItem[]>>({});
  const connected = ref(false);
  const socketEnabled = ref(false);
  const chatOpen = ref(false);
  const chatPeerId = ref<string | null>(null);

  let ws: WebSocket | null = null;

  function openChat(peerId?: string) {
    if (peerId) chatPeerId.value = peerId;
    chatOpen.value = true;
  }

  function closeChat() {
    chatOpen.value = false;
    chatPeerId.value = null;
  }

  // ---------------- WebSocket 全局通道 ----------------

  function connectSocket() {
    if (socketEnabled.value) return;
    const auth = useAuthStore();
    if (!auth.user?.id) return;

    socketEnabled.value = true;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const token = localStorage.getItem('auth_token') || '';
    const url = `${protocol}//${host}/ws/user/${encodeURIComponent(auth.user.id)}?token=${encodeURIComponent(token)}`;

    ws = new WebSocket(url);
    ws.onopen = () => {
      connected.value = true;
    };
    ws.onclose = () => {
      connected.value = false;
    };
    ws.onerror = () => {
      connected.value = false;
    };
    ws.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data);
        if (data.event === 'notification:new' && data.notification) {
          handleNewNotification(data.notification as NotificationItem);
        } else if (data.event === 'chat:message' && data.message) {
          handleNewMessage(data.message as ChatMessageItem);
        }
      } catch {
        /* ignore */
      }
    };
  }

  function handleNewNotification(n: NotificationItem) {
    // 未读数 +1，列表头部插入（去重）
    if (!n.is_read) unreadCount.value += 1;
    notifications.value = [n, ...notifications.value.filter((x) => x.id !== n.id)];
    playNotificationSound();
  }

  function handleNewMessage(m: ChatMessageItem) {
    playChatSound();
    // 若当前会话即接收方，直接追加
    const list = messages.value[m.sender_id];
    if (list) {
      list.push(m);
    }
    refreshConversations();
  }

  // ---------------- 通知 ----------------

  async function fetchUnreadCount() {
    try {
      const res = await notifService.fetchUnreadCount();
      unreadCount.value = res.data?.count ?? 0;
    } catch {
      /* ignore */
    }
  }

  async function fetchList(params?: { page?: number; page_size?: number; unread_only?: boolean }) {
    const res = await notifService.fetchNotifications(params);
    notifications.value = (res.data as unknown as { data: NotificationItem[] }).data || [];
    unreadCount.value =
      (res.data as unknown as { unread_count?: number }).unread_count ?? unreadCount.value;
    return res.data as unknown as { total: number; unread_count: number };
  }

  async function markRead(id: string) {
    await notifService.markNotificationRead(id);
    const item = notifications.value.find((n) => n.id === id);
    if (item && !item.is_read) {
      item.is_read = true;
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  }

  async function markAllRead() {
    await notifService.markAllNotificationsRead();
    notifications.value.forEach((n) => (n.is_read = true));
    unreadCount.value = 0;
  }

  async function remove(id: string) {
    await notifService.deleteNotification(id);
    const item = notifications.value.find((n) => n.id === id);
    notifications.value = notifications.value.filter((n) => n.id !== id);
    if (item && !item.is_read) {
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  }

  // ---------------- 聊天 ----------------

  async function refreshConversations() {
    try {
      const res = await notifService.fetchConversations();
      const list = (res.data as unknown as ConversationItem[]) || [];
      // 补充 peer 名称
      const auth = useAuthStore();
      list.forEach((c) => {
        c.peer_name = c.peer_id === auth.user?.id ? auth.user?.name || auth.user?.username : `用户 ${c.peer_id.slice(0, 6)}`;
      });
      conversations.value = list;
    } catch {
      /* ignore */
    }
  }

  async function loadMessages(peerId: string) {
    const res = await notifService.fetchChatMessages(peerId);
    messages.value[peerId] =
      (res.data as unknown as { data: ChatMessageItem[] }).data || [];
  }

  async function sendMessage(peerId: string, content: string, noteId?: string) {
    const msg = await notifService.sendChatMessage(peerId, content, noteId);
    const list = messages.value[peerId] || [];
    list.push(msg.data as ChatMessageItem);
    messages.value[peerId] = list;
    await refreshConversations();
    return msg.data as ChatMessageItem;
  }

  async function markConversationRead(peerId: string) {
    await notifService.markConversationRead(peerId);
    const conv = conversations.value.find((c) => c.peer_id === peerId);
    if (conv) conv.unread = 0;
  }

  function disconnect() {
    ws?.close();
    ws = null;
    socketEnabled.value = false;
    connected.value = false;
  }

  return {
    unreadCount,
    notifications,
    conversations,
    messages,
    connected,
    chatOpen,
    chatPeerId,
    openChat,
    closeChat,
    connectSocket,
    disconnect,
    fetchUnreadCount,
    fetchList,
    markRead,
    markAllRead,
    remove,
    refreshConversations,
    loadMessages,
    sendMessage,
    markConversationRead,
  };
});
