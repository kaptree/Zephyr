import { get, post, del } from './api';
import type { ApiResponse } from '@/types/api';

export interface WorkGroupData {
  id: string;
  name: string;
  description: string;
  initiator_id: string;
  initiator?: { id: string; name: string; username: string };
  template_type: string;
  status: string;
  due_time?: string;
  members: WorkGroupMemberData[];
  created_at: string;
  updated_at: string;
}

export interface WorkGroupMemberData {
  group_id: string;
  user_id: string;
  user?: {
    id: string;
    name: string;
    username: string;
    avatar: string;
    role: string;
    department?: { name: string };
  };
  role: string;
  sub_group_name: string;
}

export interface CreateWorkGroupPayload {
  name: string;
  description: string;
  template_type: string;
  due_time?: string;
  members: { user_id: string; role: string; sub_group_name: string }[];
  tags?: string[];
}

export function listWorkGroups(): Promise<ApiResponse<WorkGroupData[]>> {
  return get('/api/v1/groups');
}

export function getMyGroups(): Promise<ApiResponse<WorkGroupData[]>> {
  return get('/api/v1/groups/mine');
}

export function getWorkGroupDetail(id: string): Promise<ApiResponse<WorkGroupData>> {
  return get(`/api/v1/groups/${id}`);
}

export function createWorkGroup(
  payload: CreateWorkGroupPayload
): Promise<ApiResponse<WorkGroupData>> {
  return post('/api/v1/groups', payload);
}

export function getWorkGroupMembers(id: string): Promise<ApiResponse<WorkGroupMemberData[]>> {
  return get(`/api/v1/groups/${id}/members`);
}

export function deleteWorkGroup(id: string): Promise<ApiResponse<null>> {
  return del(`/api/v1/groups/${id}`);
}
