import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Participant, RemoteChangeData, CommandMessage } from '@/types'
import { io, Socket } from 'socket.io-client'

export const useCollaborationStore = defineStore('collaboration', () => {
  const roomId = ref<string>('')
  const noteTitle = ref('')
  const participants = ref<Participant[]>([])
  const canvasData = ref<Record<number, string>>({})
  const syncStatus = ref<'connected' | 'connecting' | 'disconnected'>('disconnected')
  const typingUsers = ref<Set<string>>(new Set())
  const commands = ref<CommandMessage[]>([])
  const columns = ref(4)

  let socket: Socket | null = null

  const typingStatusText = computed(() => {
    const users = Array.from(typingUsers.value)
    if (users.length === 0) return ''
    if (users.length === 1) return `${users[0]}正在输入...`
    return `${users.length}人正在输入...`
  })

  function joinRoom(id: string) {
    roomId.value = id
    syncStatus.value = 'connecting'

    const wsUrl = import.meta.env.VITE_WS_URL || 'http://localhost:8080'
    socket = io(`${wsUrl}/ws/notes/${id}`, {
      auth: { token: localStorage.getItem('auth_token') },
      transports: ['websocket', 'polling'],
    })

    socket.on('connect', () => {
      syncStatus.value = 'connected'
    })

    socket.on('disconnect', () => {
      syncStatus.value = 'disconnected'
    })

    socket.on('canvas:sync', (data: RemoteChangeData) => {
      if (data.column_id !== undefined) {
        canvasData.value[data.column_id] = data.content
      }
    })

    socket.on('participant:join', (p: Participant) => {
      const exists = participants.value.find(u => u.user_id === p.user_id)
      if (!exists) {
        participants.value.push({ ...p, is_online: true })
      } else {
        exists.is_online = true
      }
    })

    socket.on('participant:leave', (userId: string) => {
      const p = participants.value.find(u => u.user_id === userId)
      if (p) p.is_online = false
    })

    socket.on('typing:status', ({ user_id, name, isTyping }: { user_id: string; name: string; isTyping: boolean }) => {
      if (isTyping) {
        typingUsers.value.add(name)
      } else {
        typingUsers.value.delete(name)
      }
    })

    socket.on('command:broadcast', (cmd: CommandMessage) => {
      commands.value.push(cmd)
    })
  }

  function leaveRoom() {
    if (socket) {
      socket.disconnect()
      socket = null
    }
    roomId.value = ''
    participants.value = []
    canvasData.value = {}
    typingUsers.value.clear()
    commands.value = []
    syncStatus.value = 'disconnected'
  }

  function pushLocalChange(columnId: number, content: string) {
    if (socket && syncStatus.value === 'connected') {
      socket.emit('canvas:update', {
        column_id: columnId,
        content,
        user_id: JSON.parse(localStorage.getItem('auth_user') || '{}')?.id,
      })
    }
  }

  function sendTypingStatus(isTyping: boolean) {
    if (socket) {
      const user = JSON.parse(localStorage.getItem('auth_user') || '{}')
      socket.emit(isTyping ? 'typing:start' : 'typing:stop', {
        user_id: user.id,
        name: user.name,
      })
    }
  }

  function sendCommand(message: string) {
    if (socket) {
      socket.emit('command:broadcast', { message, timestamp: new Date().toISOString() })
    }
  }

  function setColumns(n: number) {
    columns.value = n
  }

  return {
    roomId,
    noteTitle,
    participants,
    canvasData,
    syncStatus,
    typingUsers,
    commands,
    columns,
    typingStatusText,
    joinRoom,
    leaveRoom,
    pushLocalChange,
    sendTypingStatus,
    sendCommand,
    setColumns,
  }
})
