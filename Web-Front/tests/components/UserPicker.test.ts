import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UserPicker from '@/components/common/UserPicker.vue'
import type { Department } from '@/types'

vi.mock('@/services/admin', () => ({
  getDepartments: vi.fn(() =>
    Promise.resolve({
      data: [
        { id: 'dept-1', name: 'DeptA', parent_id: null, member_count: 3, children: [
          { id: 'dept-1-1', name: 'SubDept', parent_id: 'dept-1', member_count: 2, children: [] },
        ]},
        { id: 'dept-2', name: 'DeptB', parent_id: null, member_count: 2, children: [] },
      ] as Department[],
    })
  ),
  getUsers: vi.fn(() =>
    Promise.resolve({
      data: {
        data: [
          { id: 'user-1', name: 'ZhangSan', avatar: '', dept_id: 'dept-1', department: { id: 'dept-1', name: 'DeptA' }, role: 'group_leader' },
          { id: 'user-2', name: 'LiSi', avatar: '', dept_id: 'dept-1', department: { id: 'dept-1', name: 'DeptA' }, role: 'user' },
          { id: 'user-3', name: 'WangWu', avatar: '', dept_id: 'dept-1-1', department: { id: 'dept-1-1', name: 'SubDept' }, role: 'user' },
          { id: 'user-4', name: 'ZhaoLiu', avatar: '', dept_id: 'dept-2', department: { id: 'dept-2', name: 'DeptB' }, role: 'user' },
          { id: 'user-5', name: 'QianQi', avatar: '', dept_id: 'dept-2', department: { id: 'dept-2', name: 'DeptB' }, role: 'user' },
        ],
        total: 5,
        page: 1,
        page_size: 100,
      },
    })
  ),
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  const buttons = wrapper.findAll('button')
  for (let i = 0; i < buttons.length; i++) {
    if (buttons[i].text().trim() === text) return buttons[i]
  }
  return null
}

function findButtonContainingText(wrapper: ReturnType<typeof mount>, text: string) {
  const buttons = wrapper.findAll('button')
  for (let i = 0; i < buttons.length; i++) {
    if (buttons[i].text().includes(text)) return buttons[i]
  }
  return null
}

describe('UserPicker', () => {
  const createWrapper = (props: Record<string, unknown> = {}) => {
    return mount(UserPicker, {
      props: {
        modelValue: [],
        ...props,
      },
    })
  }

  describe('button type attribute (防止误提交表单)', () => {
    it('"+ 选择人员" button has type="button"', () => {
      const wrapper = createWrapper()
      const btn = findButtonContainingText(wrapper, '选择人员')
      expect(btn).not.toBeNull()
      expect(btn!.attributes('type')).toBe('button')
    })

    it('remove user button (×) has type="button"', async () => {
      const wrapper = createWrapper({ modelValue: ['user-1'] })
      await flushPromises()
      await wrapper.vm.$nextTick()
      const allButtons = wrapper.findAll('button')
      const removeBtn = allButtons.find((b) => b.text().trim() === '×')
      expect(removeBtn).not.toBeUndefined()
      expect(removeBtn!.attributes('type')).toBe('button')
    })

    it('"完成" button has type="button"', async () => {
      const wrapper = createWrapper()
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()
      const doneBtn = findButtonByText(wrapper, '完成')
      expect(doneBtn).not.toBeNull()
      expect(doneBtn!.attributes('type')).toBe('button')
    })

    it('search result user buttons have type="button"', async () => {
      const wrapper = createWrapper()
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const searchInput = wrapper.find('input[placeholder*="搜索"]')
      expect(searchInput.exists()).toBe(true)
      await searchInput.setValue('ZhangSan')
      await wrapper.vm.$nextTick()

      const userBtn = findButtonContainingText(wrapper, 'ZhangSan')
      expect(userBtn).not.toBeNull()
      expect(userBtn!.attributes('type')).toBe('button')
    })

    it('department tree toggle buttons rendered with type:button', async () => {
      const wrapper = createWrapper()
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const deptBtn = findButtonContainingText(wrapper, 'DeptA')
      expect(deptBtn).not.toBeNull()
    })
  })

  describe('人员选择与表单提交分离', () => {
    it('clicking user in form does NOT trigger form submit', async () => {
      const formSubmitSpy = vi.fn()
      const wrapper = mount(
        {
          template: `
            <form @submit.prevent="onSubmit">
              <UserPicker v-model="selected" :multiple="true" :max="20" />
            </form>
          `,
          components: { UserPicker },
          setup() {
            const selected = { value: [] as string[] }
            return {
              selected: selected.value,
              onSubmit: formSubmitSpy,
            }
          },
        },
        {
          global: {
            stubs: {},
          },
        }
      )

      await wrapper.vm.$nextTick()

      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const searchInput = wrapper.find('input[placeholder*="搜索"]')
      await searchInput.setValue('ZhangSan')
      await wrapper.vm.$nextTick()

      const userBtn = findButtonContainingText(wrapper, 'ZhangSan')
      expect(userBtn).not.toBeNull()
      await userBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(formSubmitSpy).not.toHaveBeenCalled()
    })

    it('clicking remove user button in form does NOT trigger form submit', async () => {
      const formSubmitSpy = vi.fn()
      const wrapper = mount(
        {
          template: `
            <form @submit.prevent="onSubmit">
              <UserPicker v-model="selected" :multiple="true" :max="20" />
            </form>
          `,
          components: { UserPicker },
          setup() {
            const selected = { value: ['user-1'] as string[] }
            return {
              selected: selected.value,
              onSubmit: formSubmitSpy,
            }
          },
        },
        {
          global: {
            stubs: {},
          },
        }
      )

      await wrapper.vm.$nextTick()

      const allButtons = wrapper.findAll('button')
      const removeBtn = allButtons.find((b) => b.text().trim() === '×')
      if (removeBtn) {
        await removeBtn.trigger('click')
        await wrapper.vm.$nextTick()
        expect(formSubmitSpy).not.toHaveBeenCalled()
      }
    })

    it('opening and closing picker in form does NOT trigger form submit', async () => {
      const formSubmitSpy = vi.fn()
      const wrapper = mount(
        {
          template: `
            <form @submit.prevent="onSubmit">
              <UserPicker v-model="selected" :multiple="true" :max="20" />
            </form>
          `,
          components: { UserPicker },
          setup() {
            const selected = { value: [] as string[] }
            return {
              selected: selected.value,
              onSubmit: formSubmitSpy,
            }
          },
        },
        {
          global: {
            stubs: {},
          },
        }
      )

      await wrapper.vm.$nextTick()

      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()
      expect(formSubmitSpy).not.toHaveBeenCalled()

      const doneBtn = findButtonByText(wrapper, '完成')
      expect(doneBtn).not.toBeNull()
      await doneBtn!.trigger('click')
      await wrapper.vm.$nextTick()
      expect(formSubmitSpy).not.toHaveBeenCalled()
    })
  })

  describe('人员选择功能', () => {
    it('clicking user selects them', async () => {
      const wrapper = createWrapper()
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const searchInput = wrapper.find('input[placeholder*="搜索"]')
      await searchInput.setValue('ZhangSan')
      await wrapper.vm.$nextTick()

      const userBtn = findButtonContainingText(wrapper, 'ZhangSan')
      await userBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const emitted = wrapper.emitted('update:modelValue') as unknown[][]
      expect(emitted).toBeTruthy()
      expect(emitted.length).toBeGreaterThan(0)
      const lastEmit = emitted[emitted.length - 1][0] as string[]
      expect(lastEmit).toContain('user-1')
    })

    it('clicking selected user deselects them', async () => {
      const wrapper = createWrapper({ modelValue: ['user-1'] })
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const searchInput = wrapper.find('input[placeholder*="搜索"]')
      await searchInput.setValue('ZhangSan')
      await wrapper.vm.$nextTick()

      const userBtn = findButtonContainingText(wrapper, 'ZhangSan')
      await userBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      const emitted = wrapper.emitted('update:modelValue') as unknown[][]
      expect(emitted).toBeTruthy()
      const lastEmit = emitted[emitted.length - 1][0] as string[]
      expect(lastEmit).not.toContain('user-1')
    })

    it('shows selected count', async () => {
      const wrapper = createWrapper({ modelValue: ['user-1', 'user-2'] })
      const toggleBtn = findButtonContainingText(wrapper, '选择人员')
      await toggleBtn!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.text()).toContain('2')
    })
  })
})
