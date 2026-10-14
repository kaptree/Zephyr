import { get, post } from './api'
import type { AchievementListResp, NewlyUnlocked } from '@/types/achievement'

/** 获取全部成就（定义 + 解锁状态 + 进度） */
export function fetchAchievements() {
  return get<AchievementListResp>('/api/v1/achievements')
}

/** 触发一轮成就检查，返回本轮新解锁（前端据此弹窗） */
export function checkAchievements() {
  return post<{ newly_unlocked: NewlyUnlocked[] }>('/api/v1/achievements/check')
}
