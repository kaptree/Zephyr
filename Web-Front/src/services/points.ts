import { get } from './api'

/** 排行榜单条（含人员基础信息与头像） */
export interface LeaderboardEntry {
  rank: number
  user_id: string
  name: string
  username: string
  department: string
  position: string
  role: string
  avatar_url: string
  points: number
  /** 与上一名的分差（第一名恒为 0） */
  delta_to_prev: number
  /** 积分段位：bronze / silver / gold / platinum */
  tier: 'bronze' | 'silver' | 'gold' | 'platinum'
}

/** 每日积分趋势项 */
export interface PointTrendItem {
  date: string
  points: number
}

/** 积分流水 */
export interface PointEventItem {
  id: string
  rule_code: string
  ref_type: string
  ref_id: string
  points: number
  note: string
  event_date: string
  created_at: string
}

/** 我的积分概览 */
export interface MePoints {
  total_points: number
  week_points: number
  month_points: number
  today_points: number
  rank: number
  total_users: number
  tier: string
  recent_events: PointEventItem[]
  trend: PointTrendItem[]
}

/** 积分规则 */
export interface PointRule {
  code: string
  name: string
  category: string
  points: number
  daily_cap: number
  enabled: boolean
  sort_order: number
}

export type LeaderboardPeriod = 'total' | 'week' | 'month'

/** 积分排行榜 */
export function fetchLeaderboard(period: LeaderboardPeriod, limit = 50) {
  return get<{ period: string; list: LeaderboardEntry[] }>('/api/v1/points/leaderboard', {
    period,
    limit,
  })
}

/** 我的积分概览（总分/周/月/今日/排名/流水/近14天趋势） */
export function fetchMyPoints() {
  return get<MePoints>('/api/v1/points/me')
}

/** 积分规则列表 */
export function fetchPointRules() {
  return get<{ list: PointRule[] }>('/api/v1/points/rules')
}
