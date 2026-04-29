import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import StickyNoteCard from '@/components/note/StickyNoteCard.vue'
import { createMockNote } from '../mocks/data'

describe('StickyNoteCard', () => {
  const createWrapper = (overrides = {}) => {
    setActivePinia(createPinia())
    const note = createMockNote(overrides)
    return {
      wrapper: mount(StickyNoteCard, {
        props: { note, mode: 'web', archived: false },
      }),
      note,
    }
  }

  it('待办便签应有黄色背景样式', () => {
    const { wrapper } = createWrapper({ priority: 'normal', status: 'active' })
    const el = wrapper.element as HTMLElement
    const bg = el.style.background || el.style.backgroundColor || ''
    expect(bg).toMatch(/\(254,\s*243,\s*199\)/)
  })

  it('已完成便签应有绿色背景样式', () => {
    const { wrapper } = createWrapper({ status: 'completed' })
    const el = wrapper.element as HTMLElement
    const bg = el.style.background || el.style.backgroundColor || ''
    expect(bg).toMatch(/\(220,\s*252,\s*231\)/)
  })

  it('盯办便签应有红色背景样式', () => {
    const { wrapper } = createWrapper({ priority: 'urgent' })
    const el = wrapper.element as HTMLElement
    const bg = el.style.background || el.style.backgroundColor || ''
    expect(bg).toMatch(/\(254,\s*226,\s*226\)/)
  })

  it('应显示便签标题', () => {
    const { wrapper, note } = createWrapper()
    expect(wrapper.text()).toContain(note.title)
  })

  it('应显示便签内容', () => {
    const { wrapper, note } = createWrapper()
    expect(wrapper.text()).toContain(note.content)
  })

  it('盯办便签应显示盯办徽章', () => {
    const { wrapper } = createWrapper({ priority: 'urgent' })
    expect(wrapper.text()).toContain('盯办')
  })

  it('已完成便签应显示已完成角标', () => {
    const { wrapper } = createWrapper({ status: 'completed' })
    expect(wrapper.text()).toContain('已完成')
  })

  it('已归档应显示已归档水印', () => {
    setActivePinia(createPinia())
    const note = createMockNote()
    const wrapper = mount(StickyNoteCard, {
      props: { note, mode: 'web', archived: true },
    })
    expect(wrapper.text()).toContain('已归档')
  })

  it('点击应触发 click 事件', async () => {
    const { wrapper, note } = createWrapper()
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
  })

  it('右键应触发 context-menu 事件', async () => {
    const { wrapper } = createWrapper()
    await wrapper.trigger('contextmenu')
    expect(wrapper.emitted('context-menu')).toBeTruthy()
  })
})
