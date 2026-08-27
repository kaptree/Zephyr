import { get, post } from './api'
import type { CommandMessage } from '@/types'
import type { ApiResponse } from '@/types'

// 需求29：协同房间指令（领导下发指令 → 实时广播，成员动态刷新）

export function sendRoomCommand(noteId: string, commandText: string): Promise<ApiResponse<CommandMessage>> {
  return post(`/api/v1/rooms/${noteId}/command`, { command_text: commandText })
}

export function fetchRoomCommands(noteId: string, limit = 50): Promise<ApiResponse<CommandMessage[]>> {
  return get(`/api/v1/rooms/${noteId}/commands`, { limit })
}
