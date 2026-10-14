import { describe, it, expect, vi, beforeAll } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import type { Note, NoteFeedbackItem, FeedbackParticipantStat } from '@/types';
import FeedbackTimeline from '@/components/note/FeedbackTimeline.vue';
import { listNoteFeedbacks, addNoteFeedback } from '@/services/notes';
import { createMockNote } from '../mocks/data';

vi.mock('@/services/notes', () => ({
  listNoteFeedbacks: vi.fn(),
  addNoteFeedback: vi.fn(),
}));

// 测试数据工厂
const fb = (
  id: string,
  user_id: string,
  user_name: string,
  content: string,
  extra: Partial<NoteFeedbackItem> = {}
): NoteFeedbackItem => ({
  id,
  note_id: 'note-1',
  user_id,
  user_name,
  content,
  created_at: new Date().toISOString(),
  ...extra,
});

const participant = (user_id: string, feedback_count = 1): FeedbackParticipantStat => ({
  user_id,
  feedback_count,
  last_feedback_at: new Date().toISOString(),
});

/** 指派任务：创建人 user-1 张三、交办人 user-5、被指派人 user-2 李四 / user-3 王五、抄送 user-4 赵六 */
const assignedNoteOverrides = {
  source_type: 'assigned',
  assigner_id: 'user-5',
  assignees: [
    { user_id: 'user-2', user: { id: 'user-2', name: '李四' } },
    { user_id: 'user-3', user: { id: 'user-3', name: '王五' } },
  ],
  ccs: [{ user_id: 'user-4', user: { id: 'user-4', name: '赵六' } }],
} as unknown as Partial<Note>;

/** 模拟后端倒序返回：最新发布在前（fb-3 最新 → fb-2 → fb-1 最早） */
const mixedFeedbacks = [
  fb('fb-3', 'user-4', '赵六', '同步一下风险点'),
  fb('fb-2', 'user-1', '张三', '收到，继续推进'),
  fb('fb-1', 'user-2', '李四', '第一阶段进展汇报'),
];
const mixedParticipants = [
  participant('user-2'),
  participant('user-1'),
  participant('user-4'),
];

const mountTimeline = async (
  noteOverrides: Partial<Note> = {},
  currentUserId = 'user-2',
  listData: { feedbacks: NoteFeedbackItem[]; participants: FeedbackParticipantStat[] } = {
    feedbacks: mixedFeedbacks,
    participants: mixedParticipants,
  }
) => {
  // 克隆副本，避免组件内 push/计数更新污染模块级共享数据
  vi.mocked(listNoteFeedbacks).mockResolvedValue({
    code: 0,
    message: 'ok',
    data: {
      feedbacks: listData.feedbacks.map((f) => ({ ...f })),
      participants: listData.participants.map((p) => ({ ...p })),
    },
  } as never);
  setActivePinia(createPinia());
  const note = createMockNote({ ...assignedNoteOverrides, ...noteOverrides } as Partial<Note>);
  const wrapper = mount(FeedbackTimeline, {
    props: { note, currentUserId },
  });
  await flushPromises();
  return { wrapper, note };
};

describe('FeedbackTimeline', () => {
  beforeAll(() => {
    // jsdom 未实现 scrollIntoView，提交流程会调用
    Element.prototype.scrollIntoView = vi.fn();
  });

  it('空反馈应显示占位文案', async () => {
    const { wrapper } = await mountTimeline(
      {},
      'user-2',
      { feedbacks: [], participants: [] }
    );
    expect(wrapper.text()).toContain('暂无阶段性反馈');
    expect(wrapper.text()).not.toContain('人参与');
  });

  it('应按最新发布在前（倒序）渲染反馈并显示姓名', async () => {
    const { wrapper } = await mountTimeline();
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items.length).toBe(3);
    expect(items[0].text()).toContain('赵六');
    expect(items[1].text()).toContain('张三');
    expect(items[2].text()).toContain('李四');
  });

  it('⟨最新⟩徽章应标在列表第一条（最新发布）', async () => {
    const { wrapper } = await mountTimeline();
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items[0].text()).toContain('⟨最新⟩');
    expect(items[1].text()).not.toContain('⟨最新⟩');
    expect(items[2].text()).not.toContain('⟨最新⟩');
  });

  it('标题行应显示总数与参与人数', async () => {
    const { wrapper } = await mountTimeline();
    expect(wrapper.text()).toContain('(3) · 3 人参与');
  });

  it('指派任务创建人应显示交办人徽章', async () => {
    const { wrapper } = await mountTimeline();
    const creatorItem = wrapper.findAll('.relative.flex.items-start')[1];
    expect(creatorItem.text()).toContain('交办人');
  });

  it('被指派人与抄送人应显示对应徽章', async () => {
    const { wrapper } = await mountTimeline();
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items[2].text()).toContain('被指派人');
    expect(items[0].text()).toContain('抄送');
  });

  it('协作任务创建人应显示发起人徽章', async () => {
    const { wrapper } = await mountTimeline({ source_type: 'collab' });
    const creatorItem = wrapper.findAll('.relative.flex.items-start')[1];
    expect(creatorItem.text()).toContain('发起人');
    expect(creatorItem.text()).not.toContain('交办人');
  });

  it('多人参与时应显示人员筛选 chips 并可过滤', async () => {
    const { wrapper } = await mountTimeline();
    expect(wrapper.text()).toContain('全部·3');
    expect(wrapper.text()).toContain('李四·1');

    // 点击"李四"chip 后仅显示李四的反馈，⟨最新⟩落在其最后一条
    const chip = wrapper.findAll('button').find((b) => b.text() === '李四·1');
    expect(chip).toBeTruthy();
    await chip!.trigger('click');
    await flushPromises();
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items.length).toBe(1);
    expect(items[0].text()).toContain('李四');
    expect(items[0].text()).toContain('⟨最新⟩');
    expect(wrapper.text()).toContain('全部·3'); // 总数徽标不变
  });

  it('单人不显示筛选 chips', async () => {
    const { wrapper } = await mountTimeline(
      {},
      'user-2',
      { feedbacks: [mixedFeedbacks[0]], participants: [participant('user-2')] }
    );
    expect(wrapper.text()).not.toContain('全部·');
  });

  it('归档任务应显示只读标识且隐藏输入区', async () => {
    const { wrapper } = await mountTimeline({ is_archived: true });
    expect(wrapper.text()).toContain('已归档 · 只读');
    expect(wrapper.find('textarea').exists()).toBe(false);
  });

  it('被指派人应可输入并提交反馈', async () => {
    const { wrapper, note } = await mountTimeline({}, 'user-2');
    const textarea = wrapper.find('textarea');
    expect(textarea.exists()).toBe(true);

    vi.mocked(addNoteFeedback).mockResolvedValue({
      code: 0,
      message: 'ok',
      data: fb('fb-9', 'user-2', '李四', '今天推进到一半，明天继续'),
    } as never);

    await textarea.setValue('今天推进到一半，明天继续');
    await wrapper.find('button.btn-primary').trigger('click');
    await flushPromises();

    expect(addNoteFeedback).toHaveBeenCalledWith(note.id, '今天推进到一半，明天继续');
    // 局部更新（不整页刷新）：新反馈插入时间轴顶部 + 输入清空 + 通知父组件
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items.length).toBe(4);
    expect(items[0].text()).toContain('今天推进到一半，明天继续');
    expect(wrapper.text()).toContain('全部·4');
    expect((textarea.element as HTMLTextAreaElement).value).toBe('');
    const emitted = wrapper.emitted('submitted');
    expect(emitted).toBeTruthy();
    expect((emitted![0][0] as NoteFeedbackItem).id).toBe('fb-9');
  });

  it('提交失败应显示后端错误信息且不追加', async () => {
    const { wrapper } = await mountTimeline({}, 'user-2');
    vi.mocked(addNoteFeedback).mockRejectedValue({
      response: { data: { message: '任务已归档，禁止反馈' } },
    });

    await wrapper.find('textarea').setValue('测试失败场景');
    await wrapper.find('button.btn-primary').trigger('click');
    await flushPromises();

    expect(wrapper.text()).toContain('任务已归档，禁止反馈');
    // reject 不应追加条目
    expect(wrapper.findAll('.relative.flex.items-start').length).toBe(3);
  });

  it('抄送人应只读并看到提示', async () => {
    const { wrapper } = await mountTimeline({}, 'user-4');
    expect(wrapper.find('textarea').exists()).toBe(false);
    expect(wrapper.text()).toContain('抄送人仅可查看反馈');
  });

  it('本人发言应有描边强调样式', async () => {
    const { wrapper } = await mountTimeline({}, 'user-2');
    const items = wrapper.findAll('.relative.flex.items-start');
    expect(items[2].find('.ring-blue-200').exists()).toBe(true);
    expect(items[0].find('.ring-blue-200').exists()).toBe(false);
  });

  it('最近反馈时间应显示相对时间', async () => {
    const { wrapper } = await mountTimeline();
    expect(wrapper.text()).toContain('刚刚');
  });
});
