import { defineStore } from 'pinia';
import { ref } from 'vue';
import * as notifService from '@/services/notification';
import type { NotificationItem, ChatMessageItem, ConversationItem } from '@/types';
import { playNotificationSound, playChatSound } from '@/utils/sound';
import { useAuthStore } from './auth';

/** 需求24：右上角消息弹窗条目 */
export interface PopupItem {
  id: number;
  kind: 'chat' | 'notification';
  title: string;
  content: string;
  /** 聊天消息：对端用户 id，点击跳转聊天会话 */
  peerId?: string;
  /** 任务通知：关联任务 id，点击跳转任务详情 */
  noteId?: string;
  createdAt?: string;
}

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
  // 当前正在查看的会话（聊天页/聊天抽屉打开会话时设置）：收到该会话新消息立即标记已读
  const viewingPeerId = ref<string | null>(null);
  // 需求24：右上角消息弹窗（当前展示 + 等待队列，离线补推时逐步弹出）
  const popups = ref<PopupItem[]>([]);
  const popupQueue = ref<PopupItem[]>([]);
  let popupId = 0;
  let popupPumping = false;

  let ws: WebSocket | null = null;

  function openChat(peerId?: string) {
    if (peerId) chatPeerId.value = peerId;
    chatOpen.value = true;
  }

  function closeChat() {
    chatOpen.value = false;
    chatPeerId.value = null;
  }

  // 设置/清除当前正在查看的会话（聊天页打开会话 / 聊天抽屉打开会话 / 离开页面时）
  function setViewingPeer(peerId: string | null) {
    viewingPeerId.value = peerId;
  }

  // ---------------- 需求24：右上角消息弹窗 ----------------

  function chatPreview(m: ChatMessageItem): string {
    if (m.type === 'image') return '[图片]';
    if (m.type === 'file') return `[文件] ${m.file_name || ''}`.trim();
    return (m.content || '').slice(0, 120);
  }

  function peerDisplayName(peerId: string): string {
    const conv = conversations.value.find((c) => c.peer_id === peerId);
    if (conv?.peer_name) return conv.peer_name;
    return `用户 ${peerId.slice(0, 6)}`;
  }

  // 入队一条弹窗：同时最多展示 2 条，队列中的按间隔逐步弹出（离线补推时逐条浮现）
  function enqueuePopup(item: Omit<PopupItem, 'id'>) {
    popupQueue.value.push({ id: ++popupId, ...item });
    pumpPopups();
  }

  function dismissPopup(id: number) {
    popups.value = popups.value.filter((p) => p.id !== id);
  }

  function pumpPopups() {
    if (popupPumping) return;
    popupPumping = true;
    const step = () => {
      while (popups.value.length < 2 && popupQueue.value.length > 0) {
        const item = popupQueue.value.shift()!;
        popups.value.push(item);
        window.setTimeout(() => dismissPopup(item.id), 5000);
      }
      if (popupQueue.value.length > 0) {
        window.setTimeout(step, 700);
      } else {
        popupPumping = false;
      }
    };
    window.setTimeout(step, 0);
  }

  // 需求24：离线期间错过的未读通知 / 未读聊天，上线后逐步从右上角弹出
  async function replayMissed() {
    try {
      const res = await notifService.fetchNotifications({ page: 1, page_size: 20, unread_only: true });
      const items = (res.data as unknown as { data: NotificationItem[] }).data || [];
      // 从旧到新弹出，最多 5 条
      for (const n of items.slice(-5)) {
        enqueuePopup({
          kind: 'notification',
          title: n.title,
          content: (n.content || '').slice(0, 120),
          noteId: n.note_id,
          createdAt: n.created_at,
        });
      }
    } catch {
      /* ignore */
    }
    try {
      const convRes = await notifService.fetchConversations();
      const convs = (convRes.data as unknown as ConversationItem[]) || [];
      const me = useAuthStore().user?.id;
      let count = 0;
      for (const conv of convs.filter((c) => (c.unread || 0) > 0)) {
        if (count >= 5) break;
        const res = await notifService.fetchChatMessages(conv.peer_id, { page: 1, page_size: 20 });
        const list = (res.data as unknown as { data: ChatMessageItem[] }).data || [];
        const unreadMsgs = list.filter((m) => m.sender_id !== me && !m.is_read);
        for (const m of unreadMsgs.slice(-Math.min(3, 5 - count))) {
          enqueuePopup({
            kind: 'chat',
            title: conv.peer_name || peerDisplayName(conv.peer_id),
            content: chatPreview(m),
            peerId: conv.peer_id,
            createdAt: m.created_at,
          });
          count++;
        }
      }
    } catch {
      /* ignore */
    }
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
      // 需求24：离线期间错过的消息/通知，上线后逐步从右上角弹出
      replayMissed();
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
    // 需求24：右上角弹窗，点击跳转任务详情
    enqueuePopup({
      kind: 'notification',
      title: n.title,
      content: (n.content || '').slice(0, 120),
      noteId: n.note_id,
      createdAt: n.created_at,
    });
  }

  function handleNewMessage(m: ChatMessageItem) {
    playChatSound();
    const list = messages.value[m.sender_id];
    if (list) {
      if (!list.some((x) => x.id === m.id)) {
        // 用新数组引用替换，确保聊天页/抽屉对消息的 watch 能触发（原地 push 不会触发 watch）
        messages.value[m.sender_id] = [...list, m];
      }
    }
    const me = useAuthStore().user?.id;
    // 需求24：右上角弹窗（自己发送的不弹；正在查看该会话的不弹，避免打扰）
    if (m.sender_id !== me && viewingPeerId.value !== m.sender_id) {
      enqueuePopup({
        kind: 'chat',
        title: peerDisplayName(m.sender_id),
        content: chatPreview(m),
        peerId: m.sender_id,
        createdAt: m.created_at,
      });
    }
    // 正在查看该会话：立即标记已读并清零角标，不刷新会话列表（避免并发用旧未读数覆盖已读状态）
    if (viewingPeerId.value && viewingPeerId.value === m.sender_id) {
      markConversationRead(m.sender_id);
      return;
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
    // 本地同步：将该会话中对方发来的消息标记为已读（退出聊天后角标立即清零）
    const list = messages.value[peerId];
    if (list) {
      const me = useAuthStore().user?.id;
      list.forEach((m) => {
        if (m.sender_id !== me) m.is_read = true;
      });
    }
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
    popups,
    dismissPopup,
    openChat,
    closeChat,
    setViewingPeer,
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
