import { get, post, del, put } from './api';
import type { ApiResponse, PaginatedData } from '@/types/api';

export interface PersonalStatsData {
  total_created: number;
  total_completed: number;
  completion_rate: number;
  remind_received: number;
  avg_completion_hours: number;
  daily_trend: { date: string; count: number }[];
  tag_breakdown: { tag_name: string; count: number }[];
}

export interface AIReportData {
  report_id: string;
  period: string;
  period_label: string;
  report_type: string;
  stats: PersonalStatsData;
  report: string;
  generated_at: string;
}

export interface WorkReportItem {
  id: string
  user_id: string
  user_name: string
  period: string
  period_label: string
  report_type: string
  category?: string
  title: string
  content: string
  created_at: string
}

export interface ReportListQuery {
  page: number;
  page_size: number;
  period?: string;
  keyword?: string;
  date_from?: string;
  date_to?: string;
}

export function fetchPersonalStats(
  period: 'week' | 'month' | 'year'
): Promise<ApiResponse<PersonalStatsData>> {
  return get('/api/v1/analytics/personal-stats', { period });
}

export function generateAIReport(
  period: 'week' | 'month' | 'year'
): Promise<ApiResponse<AIReportData>> {
  return post('/api/v1/analytics/ai-report', { period });
}

export function listReports(
  query: ReportListQuery
): Promise<ApiResponse<PaginatedData<WorkReportItem>>> {
  return get('/api/v1/analytics/reports', query as unknown as Record<string, unknown>);
}

export function getReport(id: string): Promise<ApiResponse<WorkReportItem>> {
  return get(`/api/v1/analytics/reports/${id}`);
}

export function deleteReport(id: string): Promise<ApiResponse<null>> {
  return del(`/api/v1/analytics/reports/${id}`);
}

export interface ReportTemplateData {
  id: string;
  name: string;
  content: string;
  updated_at: string;
}

export function fetchReportTemplate(): Promise<ApiResponse<ReportTemplateData>> {
  return get('/api/v1/analytics/report-template');
}

export function saveReportTemplate(content: string): Promise<ApiResponse<ReportTemplateData>> {
  return put('/api/v1/analytics/report-template', { content });
}

// ==================== 团队报告（★ 新增） ====================

export interface TeamMemberStat {
  user_id: string;
  user_name: string;
  username: string;
  dept_name: string;
  total_created: number;
  total_completed: number;
  completion_rate: number;
  avg_completion_hours: number;
  remind_received: number;
}

export interface TeamStatsData {
  date_from: string;
  date_to: string;
  members: TeamMemberStat[];
  total_created: number;
  total_completed: number;
  completion_rate: number;
  member_count: number;
}

export function fetchTeamStats(params: {
  date_from?: string;
  date_to?: string;
  user_ids?: string[];
}): Promise<ApiResponse<TeamStatsData>> {
  const query: Record<string, unknown> = {};
  if (params.date_from) query.date_from = params.date_from;
  if (params.date_to) query.date_to = params.date_to;
  if (params.user_ids && params.user_ids.length > 0) query.user_ids = params.user_ids.join(',');
  return get('/api/v1/analytics/team-stats', query);
}

export interface TeamReportData {
  report_id: string;
  period: string;
  period_label: string;
  report_type: string;
  stats: TeamStatsData;
  report: string;
  generated_at: string;
}

export function generateTeamReport(data: {
  period?: string;
  date_from?: string;
  date_to?: string;
  ai_config_id?: string;
  user_ids?: string[];
}): Promise<ApiResponse<TeamReportData>> {
  return post('/api/v1/analytics/team-report', data);
}

export interface AIConfigBrief {
  id: string;
  provider_type: string;
  provider_name: string;
  model_name: string;
}

export function listReportAIConfigs(): Promise<ApiResponse<AIConfigBrief[]>> {
  return get('/api/v1/analytics/ai-configs');
}
