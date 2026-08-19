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
  const onlineIds = ref<string[]>([]);
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
        } else if (data.event === 'presence:update') {
          handlePresence(data);
        } else if (data.event === 'chat:read') {
          handleChatRead(data);
        }
      } catch {
        /* ignore */
      }
    };
  }

  // 在线用户列表更新（聊天页显示对方在线/离线状态）
  function handlePresence(data: { online_ids?: string[] }) {
    onlineIds.value = data.online_ids || [];
  }

  // 已读回执：对方读了我的消息 → 将我发出的消息标记为已读
  function handleChatRead(data: { reader_id: string }) {
    const list = messages.value[data.reader_id];
    if (!list) return;
    const me = useAuthStore().user?.id;
    list.forEach((m) => {
      if (m.sender_id === me) m.is_read = true;
    });
  }

  function handleNewNotification(n: NotificationItem) {
    // 未读数 +1，列表头部插入（去重）
    if (!n.is_read) unreadCount.value += 1;
    notifications.value = [n, ...notifications.value.filter((x) => x.id !== n.id)];
    playNotificationSound();
  }

  function handleNewMessage(m: ChatMessageItem) {
    playChatSound();
    // 若当前会话列表已加载，追加并去重
    const list = messages.value[m.sender_id];
    if (list) {
      if (!list.some((x) => x.id === m.id)) list.push(m);
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

  // 拉取当前在线用户列表（进入聊天页时调用）
  async function fetchOnlineUsers() {
    try {
      const res = await notifService.fetchChatOnline();
      onlineIds.value = (res.data as unknown as { online_ids: string[] }).online_ids || [];
    } catch {
      /* ignore */
    }
  }

  function isOnline(userId: string) {
    return onlineIds.value.includes(userId);
  }

  async function refreshConversations() {
    try {
      const res = await notifService.fetchConversations();
      const list = (res.data as unknown as ConversationItem[]) || [];
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

  // 分页加载更早的消息（用于会话历史翻页）
  async function fetchMessagesPage(peerId: string, page: number, pageSize: number) {
    const res = await notifService.fetchChatMessages(peerId, { page, page_size: pageSize });
    return res.data;
  }

  async function sendMessage(peerId: string, payload: notifService.SendChatPayload) {
    const msg = await notifService.sendChatMessage(peerId, payload);
    const list = messages.value[peerId] || [];
    if (!list.some((x) => x.id === (msg.data as ChatMessageItem).id)) {
      list.push(msg.data as ChatMessageItem);
    }
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
    onlineIds,
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
    fetchOnlineUsers,
    isOnline,
    refreshConversations,
    loadMessages,
    fetchMessagesPage,
    sendMessage,
    markConversationRead,
  };
});
