import { get, post } from './api'
import type { ApiResponse } from '@/types/api'

export interface PersonalStatsData {
  total_created: number
  total_completed: number
  completion_rate: number
  remind_received: number
  avg_completion_hours: number
  daily_trend: { date: string; count: number }[]
  tag_breakdown: { tag_name: string; count: number }[]
}

export interface AIReportData {
  period: string
  period_label: string
  stats: PersonalStatsData
  report: string
  generated_at: string
}

export function fetchPersonalStats(period: 'week' | 'month' | 'year'): Promise<ApiResponse<PersonalStatsData>> {
  return get('/api/v1/analytics/personal-stats', { period })
}

export function generateAIReport(period: 'week' | 'month' | 'year'): Promise<ApiResponse<AIReportData>> {
  return post('/api/v1/analytics/ai-report', { period })
}
