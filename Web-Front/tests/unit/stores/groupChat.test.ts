import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useNotificationStore } from '@/stores/notification';

// ---------------- mock 服务层 ----------------

// vi.mock 工厂会被提升到文件顶部，需用 vi.hoisted 让 mock 数据可被工厂引用
const { groupConvMock } = vi.hoisted(() => ({
  groupConvMock: {
    id: 'g1',
    name: '测试群',
    owner_id: 'user-me',
    member_count: 3,
    unread: 0,
    last_msg: '',
    created_at: '2026-09-07T00:00:00Z',
  },
}));

vi.mock('@/services/groupChat', () => ({
  fetchGroups: vi.fn().mockResolvedValue({ code: 0, data: [groupConvMock] }),
  createGroup: vi.fn(),
  fetchGroupDetail: vi.fn(),
  fetchGroupMembers: vi.fn().mockResolvedValue({ code: 0, data: [] }),
  addGroupMembers: vi.fn(),
  removeGroupMember: vi.fn(),
  renameGroup: vi.fn(),
  dissolveGroup: vi.fn(),
  quitGroup: vi.fn(),
  fetchGroupMessages: vi.fn().mockResolvedValue({
    code: 0,
    data: { data: [], total: 0, page: 1, page_size: 20 },
  }),
  sendGroupMessage: vi.fn(),
  markGroupRead: vi.fn().mockResolvedValue({ code: 0, data: { success: true } }),
}));

vi.mock('@/services/notification', () => ({
  fetchNotifications: vi.fn().mockResolvedValue({ code: 0, data: { data: [] } }),
  fetchConversations: vi.fn().mockResolvedValue({ code: 0, data: [] }),
  fetchChatMessages: vi.fn().mockResolvedValue({ code: 0, data: { data: [] } }),
  sendChatMessage: vi.fn(),
  markConversationRead: vi.fn(),
  markNotificationRead: vi.fn(),
  markAllNotificationsRead: vi.fn(),
  deleteNotification: vi.fn(),
  fetchUnreadCount: vi.fn().mockResolvedValue({ code: 0, data: { count: 0 } }),
  fetchChatOnline: vi.fn().mockResolvedValue({ code: 0, data: [] }),
}));

vi.mock('@/utils/sound', () => ({
  playNotificationSound: vi.fn(),
  playChatSound: vi.fn(),
}));

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { id: 'user-me', username: 'admin', name: '管理员' },
  }),
}));

import * as groupChatService from '@/services/groupChat';

// ---------------- mock WebSocket ----------------

class MockWebSocket {
  static instances: MockWebSocket[] = [];
  onopen: ((ev?: unknown) => void) | null = null;
  onclose: ((ev?: unknown) => void) | null = null;
  onerror: ((ev?: unknown) => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  constructor(public url: string) {
    MockWebSocket.instances.push(this);
  }
  close() {
    /* noop */
  }
}

function triggerWs(payload: object) {
  const ws = MockWebSocket.instances[MockWebSocket.instances.length - 1];
  if (!ws) throw new Error('WebSocket 未连接');
  ws.onmessage?.({ data: JSON.stringify(payload) });
}

async function flushTimers(ms = 30) {
  await new Promise((r) => setTimeout(r, ms));
}

// 测试用假 token（exp 为 1 小时后）
const fakeToken = `x.${btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 3600 }))}.y`;

function groupMsg(overrides: Record<string, unknown> = {}) {
  return {
    id: 'm1',
    group_id: 'g1',
    sender_id: 'user-other',
    sender_name: '张三',
    type: 'text',
    content: '大家好',
    created_at: '2026-09-07T10:00:00Z',
    ...overrides,
  };
}

describe('notification store - 群聊', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.setItem('auth_token', fakeToken);
    vi.stubGlobal('WebSocket', MockWebSocket);
    MockWebSocket.instances = [];
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('收到他人群消息：入缓存 + 右上角弹窗 + 刷新群列表', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    // 先建立该群的消息缓存（实时消息仅追加到已缓存群）
    await store.loadGroupMessages('g1');
    const fetchGroupsCalls = vi.mocked(groupChatService.fetchGroups).mock.calls.length;

    store.connectSocket();
    triggerWs({ event: 'chat:group_message', message: groupMsg() });
    await flushTimers();

    expect(store.groupMessages['g1']).toHaveLength(1);
    expect(store.groupMessages['g1'][0].content).toBe('大家好');

    // 弹窗标题为群名，内容带发送者姓名（sender_name 优先）
    const chatPopups = store.popups.filter((p) => p.kind === 'chat');
    expect(chatPopups).toHaveLength(1);
    expect(chatPopups[0].title).toBe('测试群');
    expect(chatPopups[0].content).toContain('张三');
    expect(chatPopups[0].groupId).toBe('g1');

    // 非当前查看群：触发群列表刷新
    expect(vi.mocked(groupChatService.fetchGroups).mock.calls.length).toBeGreaterThan(
      fetchGroupsCalls
    );
  });

  it('正在查看该群时收到消息：自动标记已读且不弹窗', async () => {
    const conv = { ...groupConvMock, unread: 5 };
    vi.mocked(groupChatService.fetchGroups).mockResolvedValue({ code: 0, data: [conv] });
    const store = useNotificationStore();
    await store.refreshGroups();
    store.setViewingGroup('g1');

    store.connectSocket();
    triggerWs({ event: 'chat:group_message', message: groupMsg() });
    await flushTimers();

    expect(vi.mocked(groupChatService.markGroupRead)).toHaveBeenCalledWith('g1');
    expect(store.popups.filter((p) => p.kind === 'chat')).toHaveLength(0);
    expect(store.groupConversations[0].unread).toBe(0);
  });

  it('自己发送的消息：入缓存但不弹窗、不刷新列表', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');
    const fetchGroupsCalls = vi.mocked(groupChatService.fetchGroups).mock.calls.length;

    store.connectSocket();
    triggerWs({ event: 'chat:group_message', message: groupMsg({ sender_id: 'user-me' }) });
    await flushTimers();

    expect(store.groupMessages['g1']).toHaveLength(1);
    expect(store.popups.filter((p) => p.kind === 'chat')).toHaveLength(0);
    expect(vi.mocked(groupChatService.fetchGroups).mock.calls.length).toBe(fetchGroupsCalls);
  });

  it('sendGroupMessage：消息入缓存并刷新群列表', async () => {
    const sent = groupMsg({ id: 'm-send', sender_id: 'user-me' });
    vi.mocked(groupChatService.sendGroupMessage).mockResolvedValue({
      code: 0,
      data: sent as never,
    });
    const store = useNotificationStore();
    await store.refreshGroups();

    const result = await store.sendGroupMessage('g1', { content: 'hello', type: 'text' });

    expect(vi.mocked(groupChatService.sendGroupMessage)).toHaveBeenCalledWith('g1', {
      content: 'hello',
      type: 'text',
    });
    expect(result.id).toBe('m-send');
    expect(store.groupMessages['g1'].some((m) => m.id === 'm-send')).toBe(true);
  });

  it('markGroupRead：群未读清零', async () => {
    const conv = { ...groupConvMock, unread: 7 };
    vi.mocked(groupChatService.fetchGroups).mockResolvedValue({ code: 0, data: [conv] });
    const store = useNotificationStore();
    await store.refreshGroups();
    expect(store.groupConversations[0].unread).toBe(7);

    await store.markGroupRead('g1');

    expect(store.groupConversations[0].unread).toBe(0);
  });

  it('群被解散事件：清空该群消息缓存并刷新列表', async () => {
    const msg = groupMsg();
    vi.mocked(groupChatService.fetchGroupMessages).mockResolvedValue({
      code: 0,
      data: { data: [msg as never], total: 1, page: 1, page_size: 20 },
    });
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');
    expect(store.groupMessages['g1']).toHaveLength(1);

    store.connectSocket();
    triggerWs({ event: 'chat:group_updated', action: 'dissolved', group_id: 'g1' });
    await flushTimers();

    expect(store.groupMessages['g1']).toBeUndefined();
    expect(vi.mocked(groupChatService.fetchGroups)).toHaveBeenCalled();
  });

  it('收到 @我的群消息：标记有人@我且弹窗文案突出 @了你', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');

    store.connectSocket();
    triggerWs({
      event: 'chat:group_message',
      message: groupMsg({ content: '看这里', mentions: ['user-me'] }),
    });
    await flushTimers();

    expect(store.groupMentionedMe['g1']).toBe(true);
    const popup = store.popups.filter((p) => p.kind === 'chat')[0];
    expect(popup.content).toContain('在群聊中@了你');
    expect(popup.content).toContain('张三');
  });

  it('收到普通群消息：不标记有人@我', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');

    store.connectSocket();
    triggerWs({ event: 'chat:group_message', message: groupMsg() });
    await flushTimers();

    expect(store.groupMentionedMe['g1']).toBeUndefined();
  });

  it('正在查看该群时被 @：自动已读，不标记有人@我', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');
    store.setViewingGroup('g1');

    store.connectSocket();
    triggerWs({
      event: 'chat:group_message',
      message: groupMsg({ mentions: ['user-me'] }),
    });
    await flushTimers();

    expect(store.groupMentionedMe['g1']).toBeUndefined();
    expect(store.popups.filter((p) => p.kind === 'chat')).toHaveLength(0);
  });

  it('markGroupRead：清除有人@我标识', async () => {
    const store = useNotificationStore();
    await store.refreshGroups();
    await store.loadGroupMessages('g1');

    store.connectSocket();
    triggerWs({
      event: 'chat:group_message',
      message: groupMsg({ mentions: ['user-me'] }),
    });
    await flushTimers();
    expect(store.groupMentionedMe['g1']).toBe(true);

    await store.markGroupRead('g1');
    expect(store.groupMentionedMe['g1']).toBeUndefined();
  });
});
