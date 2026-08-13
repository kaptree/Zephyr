import { ref, onMounted, onUnmounted } from 'vue';
import type { NotificationItem, ChatMessageItem } from '@/types';

/**
 * 用户个人通知 WebSocket 通道
 * 后端路由：GET /ws/user/:user_id?token=xxx
 * 事件：notification:new / chat:message
 */
export function useUserSocket(userId: string) {
  const ws = ref<WebSocket | null>(null);
  const connected = ref(false);

  const onNotification = ref<((n: NotificationItem) => void) | null>(null);
  const onChatMessage = ref<((m: ChatMessageItem) => void) | null>(null);

  function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const token = localStorage.getItem('auth_token') || '';
    const url = `${protocol}//${host}/ws/user/${encodeURIComponent(userId)}?token=${encodeURIComponent(token)}`;

    const socket = new WebSocket(url);
    ws.value = socket;

    socket.onopen = () => {
      connected.value = true;
    };

    socket.onclose = () => {
      connected.value = false;
    };

    socket.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data);
        if (data.event === 'notification:new' && data.notification) {
          onNotification.value?.(data.notification as NotificationItem);
        } else if (data.event === 'chat:message' && data.message) {
          onChatMessage.value?.(data.message as ChatMessageItem);
        }
      } catch {
        /* ignore malformed messages */
      }
    };
  }

  onMounted(() => {
    if (userId) connect();
  });

  onUnmounted(() => {
    ws.value?.close();
    ws.value = null;
  });

  return { connected, onNotification, onChatMessage };
}
