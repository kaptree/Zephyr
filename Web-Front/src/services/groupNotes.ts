import { get, post } from './api';
import type { ApiResponse, PaginatedData } from '@/types/api';
import type { Note } from '@/types';

export function getGroupNotes(
  groupId: string,
  params?: Record<string, unknown>
): Promise<ApiResponse<PaginatedData<Note>>> {
  return get(`/api/v1/groups/${groupId}/notes`, params);
}

export function createGroupNote(
  groupId: string,
  payload: {
    title: string;
    content: string;
    owner_id?: string;
    due_time?: string;
    tag_ids?: string[];
  }
): Promise<ApiResponse<Note>> {
  return post(`/api/v1/groups/${groupId}/notes`, payload);
}
