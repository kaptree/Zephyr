import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
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
  /** 需求26：issue 评论通知：关联问题 id，点击跳转问题详情 */
  issueId?: string;
  createdAt?: string;
}

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref(0);
  const notifications = ref<NotificationItem[]>([]);
  // 需求30：工作台任务变化事件计数（note:updated 时自增，工作台 watch 后自动刷新）
  const noteUpdateTick = ref(0);
  const lastNoteUpdate = ref<{ note_id?: string; action?: string }>({});
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
        window.setTimeout(() => dismissPopup(item.id), 4000);
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
          issueId: n.issue_id,
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

  // 解码本地 JWT，判断 token 是否已过期（无 token 视为已过期）
  function isTokenExpired(): boolean {
    const t = localStorage.getItem('auth_token');
    if (!t) return true;
    try {
      const payload = JSON.parse(atob(t.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')));
      return typeof payload.exp === 'number' && payload.exp * 1000 < Date.now();
    } catch {
      return false;
    }
  }

  // token 过期处理：清理登录态、断开 WebSocket、跳转登录页
  function handleAuthExpired() {
    const auth = useAuthStore();
    auth.logout();
    if (window.location.pathname !== '/login') {
      window.location.href = '/login';
    }
  }

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
      // 兜底：服务端因 token 过期强制断开连接时（auth:expired 消息可能因连接异常未送达），
      // 检测本地 token 已过期则清理登录态并跳转登录页
      if (isTokenExpired()) {
        handleAuthExpired();
      }
    };
    ws.onerror = () => {
      connected.value = false;
    };
    ws.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data);
        if (data.event === 'auth:expired') {
          // 服务端检测到 token 过期并主动断开：清理登录态并跳转登录页
          handleAuthExpired();
        } else if (data.event === 'notification:new' && data.notification) {
          handleNewNotification(data.notification as NotificationItem);
        } else if (data.event === 'chat:message' && data.message) {
          handleNewMessage(data.message as ChatMessageItem);
        } else if (data.event === 'presence:update') {
          handlePresence(data);
        } else if (data.event === 'chat:read') {
          handleChatRead(data);
        } else if (data.event === 'note:updated') {
          // 需求30：别人指派/抄送任务 → 工作台动态刷新
          lastNoteUpdate.value = { note_id: data.note_id, action: data.action };
          noteUpdateTick.value += 1;
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
    // 需求31：无音频提醒——页面在后台时用系统级通知（首次请求权限）
    ensureNotificationPermission();
    showSystemNotification(n.title, (n.content || '').slice(0, 100), navigateTargetForNotification(n));
    // 需求24：右上角弹窗，点击跳转任务详情 / issue 详情
    enqueuePopup({
      kind: 'notification',
      title: n.title,
      content: (n.content || '').slice(0, 120),
      noteId: n.note_id,
      issueId: n.issue_id,
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
    // 需求31：无音频提醒——页面在后台时系统级通知聊天消息
    if (m.sender_id !== me && viewingPeerId.value !== m.sender_id) {
      ensureNotificationPermission();
      showSystemNotification(
        `${peerDisplayName(m.sender_id)} 发来消息`,
        chatPreview(m),
        `/chat?peer=${m.sender_id}`
      );
      // 需求24：右上角弹窗（自己发送的不弹；正在查看该会话的不弹，避免打扰）
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

  // ---------------- 无音频提醒（需求31）：页面标题闪烁 + favicon 未读红点 + 系统级通知 ----------------

  const BASE_TITLE = '轻燕工作台';
  const DEFAULT_FAVICON = '/favicon.svg';
  let titleFlashTimer: number | null = null;
  let flashOn = false;
  let systemNotifyReady = false;

  // 动态绘制带红色未读角标的 favicon（canvas 生成 data URL）
  function drawBadgeFavicon(count: number) {
    try {
      const size = 64;
      const canvas = document.createElement('canvas');
      canvas.width = size;
      canvas.height = size;
      const ctx = canvas.getContext('2d');
      if (!ctx) return;
      // 底色：蓝色圆角方块 + 白字
      ctx.fillStyle = '#3B82F6';
      if (typeof ctx.roundRect === 'function') {
        ctx.beginPath();
        ctx.roundRect(6, 6, size - 12, size - 12, 14);
        ctx.fill();
      } else {
        ctx.fillRect(6, 6, size - 12, size - 12);
      }
      ctx.fillStyle = '#ffffff';
      ctx.font = 'bold 32px "PingFang SC","Microsoft YaHei",sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText('燕', size / 2, size / 2 - 4);
      // 红色未读角标
      const badge = count > 99 ? '99+' : String(count);
      const bx = size - 13;
      const by = 14;
      ctx.beginPath();
      ctx.arc(bx, by, 15, 0, Math.PI * 2);
      ctx.fillStyle = '#EF4444';
      ctx.fill();
      ctx.strokeStyle = '#ffffff';
      ctx.lineWidth = 3;
      ctx.stroke();
      ctx.fillStyle = '#ffffff';
      ctx.font = 'bold 13px sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(badge, bx, by + 0.5);
      const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
      if (link) link.href = canvas.toDataURL('image/png');
    } catch {
      /* ignore */
    }
  }

  function restoreFavicon() {
    const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (link && link.href !== DEFAULT_FAVICON) link.href = DEFAULT_FAVICON;
  }

  function stopTitleFlash() {
    if (titleFlashTimer !== null) {
      window.clearInterval(titleFlashTimer);
      titleFlashTimer = null;
    }
    flashOn = false;
    document.title = BASE_TITLE;
    restoreFavicon();
  }

  // 未读数变化 → 控制标题闪烁与 favicon 角标
  watch(
    unreadCount,
    (n) => {
      if (n > 0) {
        drawBadgeFavicon(n);
        if (titleFlashTimer === null) {
          document.title = `【${n}条新消息】${BASE_TITLE}`;
          titleFlashTimer = window.setInterval(() => {
            flashOn = !flashOn;
            document.title = flashOn
              ? `【${unreadCount.value}条新消息】${BASE_TITLE}`
              : BASE_TITLE;
          }, 1200);
        }
      } else {
        stopTitleFlash();
      }
    },
    { immediate: true }
  );

  // 请求系统通知权限（首次收到消息时调用，避免一进页面就弹授权框）
  function ensureNotificationPermission() {
    if (!('Notification' in window) || systemNotifyReady) return;
    const p = Notification.permission;
    if (p === 'granted') {
      systemNotifyReady = true;
    } else if (p === 'default') {
      Notification.requestPermission()
        .then((r) => {
          systemNotifyReady = r === 'granted';
        })
        .catch(() => {
          /* ignore */
        });
    }
  }

  // 页面在后台/最小化时，通过系统级通知提醒（不依赖音频；页面可见时用右上角弹窗）
  function showSystemNotification(title: string, body: string, target: string) {
    if (!('Notification' in window)) return;
    if (Notification.permission !== 'granted') return;
    if (!document.hidden) return;
    try {
      const n = new Notification(title, {
        body,
        icon: '/logo.jpg',
        tag: 'zephyr-notify',
      });
      n.onclick = () => {
        window.focus();
        if (target) window.location.assign(target);
        n.close();
      };
    } catch {
      /* ignore */
    }
  }

  function navigateTargetForNotification(n: NotificationItem): string {
    if (n.note_id) return `/workbench?note=${n.note_id}`;
    if (n.issue_id) return `/issues/${n.issue_id}`;
    return '/notifications';
  }

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
    noteUpdateTick,
    lastNoteUpdate,
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
