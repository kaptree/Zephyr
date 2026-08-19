import { ref, onMounted, onUnmounted } from 'vue'
import type { Socket } from 'socket.io-client'
import { io } from 'socket.io-client'

export function useSocket(roomId: string) {
  const socket = ref<Socket | null>(null)
  const connected = ref(false)

  onMounted(() => {
    // 内网部署：VITE_WS_URL 留空时同源连接（避免回退到本机 8080 产生无效请求）
    const wsUrl = import.meta.env.VITE_WS_URL || ''
    socket.value = io(wsUrl ? `${wsUrl}/ws/notes/${roomId}` : `/ws/notes/${roomId}`, {
      auth: { token: localStorage.getItem('auth_token') },
      transports: ['websocket', 'polling'],
    })

    socket.value.on('connect', () => {
      connected.value = true
    })

    socket.value.on('disconnect', () => {
      connected.value = false
    })
  })

  onUnmounted(() => {
    socket.value?.disconnect()
    socket.value = null
  })

  return {
    socket,
    connected,
  }
}
