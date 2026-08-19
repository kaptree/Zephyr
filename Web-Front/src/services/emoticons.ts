import { get, postForm, del } from './api';
import type { ApiResponse } from '@/types';

export interface Emoticon {
  id: string
  name: string
  category: string
  path: string
  uploader_id: string | null
  is_system: boolean
  created_at: string
}

export interface EmoticonListData {
  categories: string[]
  list: Emoticon[]
}

export function listEmoticons(category = '') {
  return get<EmoticonListData>('/api/v1/emoticons', { category } as Record<string, unknown>);
}

// 单文件上传（自己的表情包）
export function uploadEmoticon(file: File) {
  const fd = new FormData();
  fd.append('file', file);
  return postForm<Emoticon>('/api/v1/emoticons', fd);
}

// 批量上传（系统管理员）：files 可来自多选文件或文件夹（webkitdirectory）
export function batchUploadEmoticons(files: File[], category = '') {
  const fd = new FormData();
  if (category) fd.append('category', category);
  for (const f of files) fd.append('files', f, f.webkitRelativePath || f.name);
  return postForm<{ success: number; skipped: number }>('/api/v1/emoticons/batch', fd);
}

export function deleteEmoticon(id: string): Promise<ApiResponse<{ success: boolean }>> {
  return del<{ success: boolean }>(`/api/v1/emoticons/${id}`);
}
