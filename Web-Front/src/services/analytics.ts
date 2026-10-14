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
  /** 后台生成状态：generating 生成中 / completed 已完成 / failed 失败 */
  status?: string;
  stats: PersonalStatsData;
  report: string;
  generated_at: string;
}

export interface WorkReportItem {
  id: string;
  user_id: string;
  user_name: string;
  period: string;
  period_label: string;
  report_type: string;
  category?: string;
  /** 后台生成状态：generating 生成中 / completed 已完成 / failed 失败（旧数据默认 completed） */
  status?: string;
  title: string;
  content: string;
  created_at: string;
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
  period: 'week' | 'month' | 'year',
  aiConfigId?: string
): Promise<ApiResponse<AIReportData>> {
  // aiConfigId：空 = 智能生成（AI 优先，失败回退模板）；"template" = 模板生成；其他 = 指定 AI 配置 ID
  // AI 模式后端已改为后台异步生成，接口立即返回 generating 状态，无需长超时
  return post(
    '/api/v1/analytics/ai-report',
    { period, ai_config_id: aiConfigId || '' },
    { timeout: 60000 }
  );
}

/** 查询当前用户仍在后台生成中的报告（切回页面时恢复轮询用） */
export function fetchPendingReports(): Promise<ApiResponse<WorkReportItem[]>> {
  return get('/api/v1/analytics/reports/pending');
}

export function listReports(
  query: ReportListQuery
): Promise<ApiResponse<PaginatedData<WorkReportItem>>> {
  return get('/api/v1/analytics/reports', query as unknown as Record<string, unknown>);
}

export function getReport(id: string): Promise<ApiResponse<WorkReportItem>> {
  return get(`/api/v1/analytics/reports/${id}`);
}

/** 需求25：报告详情（含周期内各成员任务/反馈明细） */
export interface ReportTaskDetail {
  id: string;
  title: string;
  content: string;
  status: 'active' | 'completed';
  created_at: string;
  completed_at?: string;
  feedback: string;
}
export interface ReportMemberDetail {
  user_id: string;
  user_name: string;
  tasks: ReportTaskDetail[];
}
export interface ReportDetailData {
  range: string;
  members: ReportMemberDetail[];
}
export function getReportDetail(
  id: string
): Promise<ApiResponse<{ report: WorkReportItem; detail: ReportDetailData }>> {
  return get(`/api/v1/analytics/reports/${id}/detail`);
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
  /** 后台生成状态：generating 生成中 / completed 已完成 / failed 失败 */
  status?: string;
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
  // AI 模式后端已改为后台异步生成，接口立即返回 generating 状态，无需长超时
  return post('/api/v1/analytics/team-report', data, { timeout: 60000 });
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

// ==================== AI 报告提示词（★ 新增） ====================

export interface AIPromptData {
  content: string;
  is_customized: boolean;
  default: string;
}

export function fetchAIPrompt(): Promise<ApiResponse<AIPromptData>> {
  return get('/api/v1/analytics/ai-prompt');
}

export function saveAIPrompt(content: string): Promise<ApiResponse<AIPromptData>> {
  return put('/api/v1/analytics/ai-prompt', { content });
}
