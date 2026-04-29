import { ref, onMounted, onUnmounted } from 'vue'
import type { Socket } from 'socket.io-client'
import { io } from 'socket.io-client'

export function useSocket(roomId: string) {
  const socket = ref<Socket | null>(null)
  const connected = ref(false)

  onMounted(() => {
    const wsUrl = import.meta.env.VITE_WS_URL || 'http://localhost:8080'
    socket.value = io(`${wsUrl}/ws/notes/${roomId}`, {
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
