import { http, HttpResponse } from 'msw';
import { mockTags, mockUser, mockAdminUser, createMockNotes } from './data';
import type { Note } from '@/types';

// 内存存储
let notesStore: Note[] = createMockNotes(10);

export const handlers = [
  // 认证
  http.post('/api/v1/auth/login', async ({ request }) => {
    const body = (await request.json()) as { username: string; password: string };
    if (body.username === 'admin' && body.password === 'admin') {
      return HttpResponse.json({
        code: 0,
        message: 'success',
        data: { token: 'mock-token-admin', user: mockAdminUser },
      });
    }
    if (body.username === 'test' && body.password === 'test') {
      return HttpResponse.json({
        code: 0,
        message: 'success',
        data: { token: 'mock-token-test', user: mockUser },
      });
    }
    return HttpResponse.json(
      { code: 401, message: '用户名或密码错误', data: null },
      { status: 401 }
    );
  }),

  http.get('/api/v1/auth/me', () => {
    const token = ''; // 简化处理
    return HttpResponse.json({
      code: 0,
      data: token === 'mock-token-admin' ? mockAdminUser : mockUser,
    });
  }),

  // 任务
  http.get('/api/v1/notes', () => {
    return HttpResponse.json({
      code: 0,
      data: {
        data: notesStore.filter((n) => n.status === 'active'),
        total: notesStore.filter((n) => n.status === 'active').length,
        page: 1,
        page_size: 20,
      },
    });
  }),

  http.post('/api/v1/notes', async ({ request }) => {
    const body = (await request.json()) as Partial<Note>;
    const newNote: Note = {
      id: `note-new-${Date.now()}`,
      title: body.title || '新任务',
      content: body.content || '',
      status: 'active',
      source_type: 'self',
      priority: 'normal',
      owner_id: 'user-1',
      creator_id: 'user-1',
      tags: [],
      assignees: [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      allowed_actions: ['edit', 'delete', 'complete', 'remind'],
    };
    notesStore.unshift(newNote);
    return HttpResponse.json({ code: 0, data: newNote });
  }),

  http.put('/api/v1/notes/:id', async ({ params, request }) => {
    const body = (await request.json()) as Partial<Note>;
    const index = notesStore.findIndex((n) => n.id === params.id);
    if (index >= 0) {
      notesStore[index] = { ...notesStore[index], ...body, updated_at: new Date().toISOString() };
      return HttpResponse.json({ code: 0, data: notesStore[index] });
    }
    return HttpResponse.json({ code: 404, message: '任务不存在', data: null }, { status: 404 });
  }),

  http.post('/api/v1/notes/:id/complete', ({ params }) => {
    const index = notesStore.findIndex((n) => n.id === params.id);
    if (index >= 0) {
      notesStore[index].status = 'completed';
      notesStore[index].completed_at = new Date().toISOString();
      return HttpResponse.json({ code: 0, data: notesStore[index] });
    }
    return HttpResponse.json({ code: 404, message: '任务不存在', data: null }, { status: 404 });
  }),

  http.post('/api/v1/notes/:id/remind', ({ params }) => {
    const index = notesStore.findIndex((n) => n.id === params.id);
    if (index >= 0) {
      notesStore[index].priority = 'urgent';
      return HttpResponse.json({ code: 0, data: notesStore[index] });
    }
    return HttpResponse.json({ code: 404, message: '任务不存在', data: null }, { status: 404 });
  }),

  http.delete('/api/v1/notes/:id', ({ params }) => {
    const index = notesStore.findIndex((n) => n.id === params.id);
    if (index >= 0) {
      notesStore[index].status = 'archived';
      return HttpResponse.json({ code: 0, data: { success: true } });
    }
    return HttpResponse.json({ code: 404, message: '任务不存在', data: null }, { status: 404 });
  }),

  http.post('/api/v1/notes/:id/restore', ({ params }) => {
    const index = notesStore.findIndex((n) => n.id === params.id);
    if (index >= 0) {
      notesStore[index].status = 'active';
      return HttpResponse.json({ code: 0, data: notesStore[index] });
    }
    return HttpResponse.json({ code: 404, message: '任务不存在', data: null }, { status: 404 });
  }),

  // 标签
  http.get('/api/v1/tags', () => {
    return HttpResponse.json({ code: 0, data: mockTags });
  }),

  // 部门
  http.get('/api/v1/departments', () => {
    return HttpResponse.json({
      code: 0,
      data: [
        { id: 'dept-1', label: '刑警支队', children: [{ id: 'dept-1-1', label: '侦查一队' }] },
        { id: 'dept-2', label: '治安支队', children: [] },
      ],
    });
  }),

  // 用户
  http.get('/api/v1/users', () => {
    return HttpResponse.json({
      code: 0,
      data: { data: [mockUser, mockAdminUser], total: 2, page: 1 },
    });
  }),
];
