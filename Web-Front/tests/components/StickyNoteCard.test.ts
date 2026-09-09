import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import StickyNoteCard from '@/components/note/StickyNoteCard.vue';
import { createMockNote } from '../mocks/data';

describe('StickyNoteCard', () => {
  const createWrapper = (overrides = {}, extraProps: Record<string, unknown> = {}) => {
    setActivePinia(createPinia());
    const note = createMockNote(overrides);
    return {
      wrapper: mount(StickyNoteCard, {
        props: { note, mode: 'web', archived: false, ...extraProps },
      }),
      note,
    };
  };

  it('待办任务应有黄色背景样式', () => {
    const { wrapper } = createWrapper({ color_status: 'yellow' });
    expect(wrapper.classes()).toContain('bg-amber-100');
  });

  it('盯办任务应有红色背景样式', () => {
    const { wrapper } = createWrapper({ color_status: 'red' });
    expect(wrapper.classes()).toContain('bg-red-100');
  });

  it('应显示任务标题', () => {
    const { wrapper, note } = createWrapper();
    expect(wrapper.text()).toContain(note.title);
  });

  it('应显示任务内容', () => {
    const { wrapper, note } = createWrapper();
    expect(wrapper.text()).toContain(note.content);
  });

  it('被指派任务应显示盯办徽章', () => {
    const { wrapper } = createWrapper({ color_status: 'red', source_type: 'assigned' });
    expect(wrapper.text()).toContain('盯办');
  });

  it('应显示完成并归档按钮', () => {
    const { wrapper } = createWrapper();
    expect(wrapper.text()).toContain('完成并归档');
  });

  it('待办任务应显示重要按钮', () => {
    const { wrapper } = createWrapper({ color_status: 'yellow' }, { extraActions: true });
    expect(wrapper.text()).toContain('重要');
  });

  it('已归档应显示已归档水印', () => {
    setActivePinia(createPinia());
    const note = createMockNote();
    const wrapper = mount(StickyNoteCard, {
      props: { note, mode: 'web', archived: true },
    });
    expect(wrapper.text()).toContain('已归档');
  });

  it('点击应触发 click 事件', async () => {
    const { wrapper } = createWrapper();
    await wrapper.trigger('click');
    expect(wrapper.emitted('click')).toBeTruthy();
  });

  it('右键应触发 context-menu 事件', async () => {
    const { wrapper } = createWrapper();
    await wrapper.trigger('contextmenu');
    expect(wrapper.emitted('context-menu')).toBeTruthy();
  });
});
