import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Participant, CommandMessage } from '@/types'
import { sendRoomCommand, fetchRoomCommands } from '@/services/collaboration'

// 需求29：协同房间改为原生 WebSocket（与后端 gorilla 协议一致，?token= 鉴权，JSON event 驱动）
export const useCollaborationStore = defineStore('collaboration', () => {
  const roomId = ref<string>('')
  const noteTitle = ref('')
  const participants = ref<Participant[]>([])
  const canvasData = ref<Record<number, string>>({})
  const syncStatus = ref<'connected' | 'connecting' | 'disconnected'>('disconnected')
  const typingUsers = ref<Set<string>>(new Set())
  const commands = ref<CommandMessage[]>([])
  const columns = ref(4)

  let ws: WebSocket | null = null

  const typingStatusText = computed(() => {
    const users = Array.from(typingUsers.value)
    if (users.length === 0) return ''
    if (users.length === 1) return `${users[0]}正在输入...`
    return `${users.length}人正在输入...`
  })

  function buildWsUrl(id: string): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = localStorage.getItem('auth_token') || ''
    return `${protocol}//${window.location.host}/ws/${encodeURIComponent(id)}?token=${encodeURIComponent(token)}`
  }

  function sendRaw(obj: Record<string, unknown>) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(obj))
    }
  }

  function handleServerEvent(data: Record<string, unknown>) {
    switch (data.event) {
      case 'canvas:sync':
        if (data.column_id !== undefined) {
          canvasData.value[data.column_id as number] = (data.content as string) || ''
        }
        break
      // 需求29：别人下发的指令实时广播 → 动态刷新
      case 'command:broadcast':
        if (data.command) {
          const cmd = data.command as CommandMessage
          if (!commands.value.some((c) => c.id === cmd.id)) {
            commands.value.push(cmd)
          }
        }
        break
      default:
        break
    }
  }

  function joinRoom(id: string) {
    roomId.value = id
    syncStatus.value = 'connecting'

    ws = new WebSocket(buildWsUrl(id))
    ws.onopen = () => {
      syncStatus.value = 'connected'
      // 通知房间更新在场状态
      sendRaw({ event: 'room:join' })
    }
    ws.onclose = () => {
      syncStatus.value = 'disconnected'
    }
    ws.onerror = () => {
      syncStatus.value = 'disconnected'
    }
    ws.onmessage = (ev) => {
      try {
        handleServerEvent(JSON.parse(ev.data) as Record<string, unknown>)
      } catch {
        /* ignore */
      }
    }
  }

  function leaveRoom() {
    if (ws) {
      ws.close()
      ws = null
    }
    roomId.value = ''
    participants.value = []
    canvasData.value = {}
    typingUsers.value.clear()
    commands.value = []
    syncStatus.value = 'disconnected'
  }

  function pushLocalChange(columnId: number, content: string) {
    sendRaw({
      event: 'canvas:update',
      column_id: columnId,
      content,
      user_id: JSON.parse(localStorage.getItem('auth_user') || '{}')?.id,
    })
  }

  function sendTypingStatus(isTyping: boolean) {
    const user = JSON.parse(localStorage.getItem('auth_user') || '{}')
    sendRaw({
      event: isTyping ? 'typing:start' : 'typing:stop',
      user_id: user.id,
      name: user.name,
    })
  }

  // 需求29：发送指令（REST 持久化 + 后端广播，返回后本地也立即展示）
  async function sendCommand(message: string) {
    if (!roomId.value) return
    const res = await sendRoomCommand(roomId.value, message)
    const cmd = res.data as CommandMessage
    if (cmd && !commands.value.some((c) => c.id === cmd.id)) {
      commands.value.push(cmd)
    }
    return cmd
  }

  // 需求29：加载房间指令历史（进入房间时）
  async function fetchCommands() {
    if (!roomId.value) return
    try {
      const res = await fetchRoomCommands(roomId.value)
      commands.value = (res.data as CommandMessage[]) || []
    } catch {
      /* ignore */
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
    fetchCommands,
    setColumns,
  }
})
