import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import WorkbenchFilterBar from '@/components/note/WorkbenchFilterBar.vue';

/**
 * WorkbenchFilterBar（工作台筛选行）组件测试
 * 覆盖：渲染、关键词模糊查询交互、时间范围筛选联动、清空、筛选生效提示
 */

// 便于复用的挂载工厂
function mountBar(
  props: { keyword?: string; dateFrom?: string; dateTo?: string } = {}
) {
  return mount(WorkbenchFilterBar, {
    props: {
      keyword: '',
      dateFrom: '',
      dateTo: '',
      ...props,
    },
  });
}

describe('WorkbenchFilterBar 工作台筛选行', () => {
  it('应渲染时间范围选择、关键词输入框、搜索与清空按钮', () => {
    const wrapper = mountBar();
    // DateRangePicker 触发按钮（默认显示 placeholder 文案）
    expect(wrapper.text()).toContain('按创建时间筛选');
    // 关键词输入框
    const input = wrapper.find('input.input-field');
    expect(input.exists()).toBe(true);
    expect(input.attributes('placeholder')).toContain('搜索标题 / 内容关键词');
    // 搜索 / 清空按钮
    const buttons = wrapper.findAll('button').map((b) => b.text());
    expect(buttons).toContain('搜索');
    expect(buttons).toContain('清空');
  });

  it('输入关键词应同步 emit update:keyword', async () => {
    const wrapper = mountBar();
    await wrapper.find('input.input-field').setValue('诈骗 案件');
    expect(wrapper.emitted('update:keyword')).toBeTruthy();
    expect(wrapper.emitted('update:keyword')![0][0]).toBe('诈骗 案件');
  });

  it('关键词输入框回车应触发 search', async () => {
    const wrapper = mountBar();
    const input = wrapper.find('input.input-field');
    await input.setValue('电诈');
    await input.trigger('keyup.enter');
    expect(wrapper.emitted('search')).toHaveLength(1);
  });

  it('点击搜索按钮应触发 search', async () => {
    const wrapper = mountBar();
    await wrapper.find('[data-testid="filter-search-btn"]').trigger('click');
    expect(wrapper.emitted('search')).toHaveLength(1);
  });

  it('点击清空按钮应触发 reset', async () => {
    const wrapper = mountBar();
    await wrapper.find('[data-testid="filter-reset-btn"]').trigger('click');
    expect(wrapper.emitted('reset')).toHaveLength(1);
  });

  it('选择日期范围预设（近7天）应同步 update:dateFrom/dateTo 并触发 date-change', async () => {
    const wrapper = mountBar();
    // 打开 DateRangePicker 面板
    await wrapper.find('[data-testid="workbench-filter-bar"] .relative button').trigger('click');
    // 点击「近7天」快捷预设
    const preset = wrapper.findAll('button').find((b) => b.text() === '近7天');
    expect(preset).toBeTruthy();
    await preset!.trigger('click');

    const fromEvents = wrapper.emitted('update:dateFrom');
    const toEvents = wrapper.emitted('update:dateTo');
    expect(fromEvents).toHaveLength(1);
    expect(toEvents).toHaveLength(1);

    // 近7天：from = 今天-6 天，to = 今天（YYYY-MM-DD，与组件本地时区一致）
    const fmt = (d: Date) => {
      const p = (n: number) => String(n).padStart(2, '0');
      return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
    };
    const today = new Date();
    const from = new Date();
    from.setDate(from.getDate() - 6);
    expect(fromEvents![0][0]).toBe(fmt(from));
    expect(toEvents![0][0]).toBe(fmt(today));
    // 日期变化应触发即时查询事件
    expect(wrapper.emitted('date-change')).toHaveLength(1);
  });

  it('存在筛选条件时应显示「筛选生效中」提示，无筛选时隐藏', async () => {
    const wrapper = mountBar({ keyword: '电诈' });
    expect(wrapper.find('[data-testid="filter-active-hint"]').exists()).toBe(true);
    expect(wrapper.text()).toContain('筛选生效中');

    await wrapper.setProps({ keyword: '', dateFrom: '', dateTo: '' });
    expect(wrapper.find('[data-testid="filter-active-hint"]').exists()).toBe(false);

    // 仅时间范围也算筛选生效
    await wrapper.setProps({ dateFrom: '2026-10-01', dateTo: '2026-10-13' });
    expect(wrapper.find('[data-testid="filter-active-hint"]').exists()).toBe(true);
  });
});
