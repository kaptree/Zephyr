import { get, post, put } from './api'
import type { ApiResponse, PaginatedData } from '@/types'

export interface IssueUserBrief {
  id: string
  name: string
  dept_name?: string
}

export interface IssueItem {
  id: string
  issue_no: number
  title: string
  content: string
  type: 'bug' | 'feature'
  status: 'open' | 'closed'
  user_id: string
  user_name: string
  creator?: IssueUserBrief
  created_at: string
  updated_at: string
  comment_count: number
}

export interface IssueCommentItem {
  id: string
  issue_id: string
  user_id: string
  user_name: string
  user?: IssueUserBrief
  content: string
  created_at: string
}

export interface IssueDetail {
  issue: IssueItem
  comments: IssueCommentItem[]
}

export interface IssueQuery {
  page: number
  page_size: number
  status?: string
  type?: string
  keyword?: string
}

export function listIssues(query: IssueQuery): Promise<ApiResponse<PaginatedData<IssueItem>>> {
  return get('/api/v1/issues', query as unknown as Record<string, unknown>)
}

export function getIssue(id: string): Promise<ApiResponse<IssueDetail>> {
  return get(`/api/v1/issues/${id}`)
}

export function createIssue(payload: { title: string; content: string; type: string }): Promise<ApiResponse<IssueItem>> {
  return post('/api/v1/issues', payload)
}

export function addIssueComment(id: string, content: string): Promise<ApiResponse<IssueCommentItem>> {
  return post(`/api/v1/issues/${id}/comments`, { content })
}

export function updateIssueStatus(id: string, status: 'open' | 'closed'): Promise<ApiResponse<{ success: boolean; status: string }>> {
  return put(`/api/v1/issues/${id}/status`, { status })
}
