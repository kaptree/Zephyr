<script setup lang="ts">
import {
  ref,
  computed,
  onMounted,
  onUnmounted,
  h,
  defineComponent,
  type PropType,
  type DefineComponent,
} from 'vue';
import type { UserBrief, Department } from '@/types';
import { getDepartments, getUsers } from '@/services/admin';
import { useToast } from '@/composables/useToast';

const { warning: toastWarning } = useToast();

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    multiple?: boolean;
    max?: number;
    dropUp?: boolean;
    /** 不可选的用户（已在另一处被选择，如已在指派/抄送中），点击将提示且不会勾选 */
    disabledIds?: string[];
    /** 禁用用户的提示文案，如「已在抄送人员中」 */
    disabledNote?: string;
  }>(),
  {
    multiple: true,
    max: 20,
    dropUp: false,
    disabledIds: () => [],
    disabledNote: '已被选择',
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string[]];
}>();

const open = ref(false);
const searchText = ref('');
const rootEl = ref<HTMLElement | null>(null);
const departments = ref<Department[]>([]);
const users = ref<UserBrief[]>([]);
const loading = ref(false);
const loadError = ref('');
const expandedDepts = ref<Set<string>>(new Set());

const selectedUsers = computed(() => users.value.filter((u) => props.modelValue.includes(u.id)));

const filteredUsers = computed(() => {
  if (!searchText.value) return users.value;
  const q = searchText.value.toLowerCase();
  return users.value.filter(
    (u) => u.name.toLowerCase().includes(q) || u.dept_name.toLowerCase().includes(q)
  );
});

async function loadData() {
  loading.value = true;
  loadError.value = '';
  try {
    const [deptRes, userRes] = await Promise.all([
      getDepartments(false),
      getUsers({ page: 1, page_size: 100 }),
    ]);
    departments.value = (deptRes.data as unknown as Department[]) || [];

    const rawData = (userRes.data as unknown as { data: any[] }).data || [];
    users.value = rawData.map((u: any) => ({
      id: u.id || '',
      name: u.name || '',
      avatar: u.avatar || '',
      dept_id: u.dept_id || u.department?.id || '',
      dept_name: u.department?.name || u.dept_name || '',
      role: u.role || 'user',
    })) as UserBrief[];
  } catch {
    loadError.value = '加载组织架构失败，请检查网络连接后重试';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadData();
  // 点击组件外部空白处关闭下拉列表
  document.addEventListener('click', onDocClick);
});

onUnmounted(() => {
  document.removeEventListener('click', onDocClick);
});

function onDocClick(e: MouseEvent) {
  if (rootEl.value && !rootEl.value.contains(e.target as Node)) {
    open.value = false;
  }
}

function isDisabledUser(userId: string): boolean {
  return props.disabledIds.includes(userId);
}

function toggleUser(userId: string) {
  // 已在另一处选择（如已选为指派/抄送人员）的用户不允许重复选择
  if (isDisabledUser(userId)) {
    toastWarning(`该用户${props.disabledNote}，不能重复选择`);
    return;
  }
  const current = [...props.modelValue];
  const idx = current.indexOf(userId);
  if (idx >= 0) {
    current.splice(idx, 1);
  } else if (props.multiple) {
    if (current.length < props.max) {
      current.push(userId);
    }
  } else {
    emit('update:modelValue', [userId]);
    open.value = false;
    return;
  }
  emit('update:modelValue', current);
}

function removeUser(userId: string) {
  const current = props.modelValue.filter((id) => id !== userId);
  emit('update:modelValue', current);
}

function isSelected(userId: string): boolean {
  return props.modelValue.includes(userId);
}

function toggleDept(deptId: string) {
  const current = new Set(expandedDepts.value);
  if (current.has(deptId)) {
    current.delete(deptId);
  } else {
    current.add(deptId);
  }
  expandedDepts.value = current;
}

/** 收集某部门下（含所有子部门）的全部用户 ID */
function collectDeptUserIds(deptId: string): string[] {
  const ids: string[] = [];
  const walk = (deps: Department[]) => {
    for (const d of deps) {
      users.value.filter((u) => (u as any).dept_id === d.id).forEach((u) => ids.push(u.id));
      if (d.children?.length) walk(d.children);
    }
  };
  const find = (deps: Department[]): boolean => {
    for (const d of deps) {
      if (d.id === deptId) {
        walk([d]);
        return true;
      }
      if (d.children?.length && find(d.children)) return true;
    }
    return false;
  };
  find(departments.value);
  return [...new Set(ids)];
}

/** 一键全选 / 取消全选某部门下的所有人（排除已在另一处选择的禁用用户） */
function toggleSelectAllDept(deptId: string) {
  const ids = collectDeptUserIds(deptId).filter((id) => !isDisabledUser(id));
  if (ids.length === 0) return;
  const current = new Set(props.modelValue);
  const allSelected = ids.every((id) => current.has(id));
  if (allSelected) {
    ids.forEach((id) => current.delete(id));
  } else {
    ids.forEach((id) => current.add(id));
  }
  let result = Array.from(current);
  if (!props.multiple) {
    result = result.slice(0, 1);
  } else if (result.length > props.max) {
    result = result.slice(0, props.max);
  }
  emit('update:modelValue', result);
}

function getDirectDeptUsers(deptId: string): UserBrief[] {
  return users.value.filter((u) => (u as any).dept_id === deptId);
}

const DeptTreeItem: DefineComponent<{
  departments: Department[];
  expandedSet: Set<string>;
  userList: UserBrief[];
  selectedIds: string[];
  disabledIds: string[];
  disabledNote: string;
}> = defineComponent({
  name: 'DeptTreeItem',
  props: {
    departments: { type: Array as PropType<Department[]>, required: true },
    expandedSet: { type: Object as PropType<Set<string>>, required: true },
    userList: { type: Array as PropType<UserBrief[]>, required: true },
    selectedIds: { type: Array as PropType<string[]>, required: true },
    disabledIds: { type: Array as PropType<string[]>, required: true },
    disabledNote: { type: String, required: true },
  },
  emits: ['toggle-dept', 'toggle-user', 'select-all'],
  setup(props, { emit }) {
    function getDirect(deptId: string): UserBrief[] {
      return props.userList.filter((u: any) => u.dept_id === deptId);
    }
    function isSel(id: string): boolean {
      return props.selectedIds.includes(id);
    }
    function onToggleDept(id: string) {
      emit('toggle-dept', id);
    }
    function onToggleUser(id: string) {
      emit('toggle-user', id);
    }
    function onSelectAll(id: string) {
      emit('select-all', id);
    }
    /** 收集部门（含子部门）下所有用户 ID */
    function collectDeptIds(dept: Department): string[] {
      const ids: string[] = [];
      const walk = (d: Department) => {
        getDirect(d.id).forEach((u) => ids.push(u.id));
        if (d.children) d.children.forEach(walk);
      };
      walk(dept);
      return ids;
    }

    return () => {
      return h(
        'div',
        { class: 'space-y-1' },
        props.departments.map((dept) => {
          const isExpanded = props.expandedSet.has(dept.id);
          const directUsers = getDirect(dept.id);
          const hasChildren = dept.children && dept.children.length > 0;
          const deptUserIds = collectDeptIds(dept).filter((id) => !props.disabledIds.includes(id));
          const allSel = deptUserIds.length > 0 && deptUserIds.every((id) => isSel(id));

          return h('div', { key: dept.id }, [
            h(
              'div',
              {
                class:
                  'w-full flex items-center gap-2 px-3 py-2.5 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50',
              },
              [
                h(
                  'button',
                  {
                    type: 'button',
                    class: 'flex items-center gap-2 flex-1 min-w-0 text-left',
                    onClick: () => onToggleDept(dept.id),
                  },
                  [
                    h(
                      'svg',
                      {
                        class: `w-3.5 h-3.5 text-slate-400 transition-transform shrink-0 ${isExpanded ? 'rotate-90' : ''}`,
                        fill: 'none',
                        viewBox: '0 0 24 24',
                        stroke: 'currentColor',
                      },
                      [
                        h('path', {
                          'stroke-linecap': 'round',
                          'stroke-linejoin': 'round',
                          'stroke-width': '2',
                          d: 'M9 5l7 7-7 7',
                        }),
                      ]
                    ),
                    h('span', { class: 'font-medium text-slate-700 truncate' }, dept.name),
                    h('span', { class: 'text-xs text-slate-400' }, String(dept.member_count || 0)),
                  ]
                ),
                h(
                  'button',
                  {
                    type: 'button',
                    class: `text-[10px] px-1.5 py-0.5 rounded shrink-0 transition-smooth ${
                      allSel
                        ? 'bg-blue-500 text-white hover:bg-blue-600'
                        : 'bg-blue-50 text-blue-600 hover:bg-blue-100'
                    }`,
                    onClick: (e: MouseEvent) => {
                      e.stopPropagation();
                      onSelectAll(dept.id);
                    },
                  },
                  allSel ? '全不选' : '全选'
                ),
              ]
            ),

            isExpanded
              ? h('div', { class: 'ml-6 space-y-1' }, [
                  hasChildren
                    ? h(DeptTreeItem, {
                        departments: dept.children!,
                        expandedSet: props.expandedSet,
                        userList: props.userList,
                        selectedIds: props.selectedIds,
                        disabledIds: props.disabledIds,
                        disabledNote: props.disabledNote,
                        'onToggle-dept': onToggleDept,
                        'onToggle-user': onToggleUser,
                        'onSelect-all': onSelectAll,
                      })
                    : null,

                  ...directUsers.map((user) => {
                    const isDisabled = props.disabledIds.includes(user.id);
                    return h(
                      'button',
                      {
                        type: 'button',
                        class: `w-full flex items-center gap-3 px-3 py-2 rounded-btn text-sm text-left transition-smooth ${
                          isDisabled
                            ? 'opacity-40 cursor-not-allowed'
                            : isSel(user.id)
                              ? 'bg-blue-50'
                              : 'hover:bg-slate-50'
                        }`,
                        onClick: () => onToggleUser(user.id),
                      },
                      [
                        h(
                          'div',
                          {
                            class:
                              'w-6 h-6 rounded-full bg-slate-200 flex items-center justify-center text-[10px] font-medium text-slate-600 shrink-0',
                          },
                          user.name.charAt(0)
                        ),
                        h('span', { class: 'text-sm text-slate-900 truncate' }, user.name),
                        user.role === 'group_leader'
                          ? h(
                              'span',
                              { class: 'text-[9px] px-1 bg-amber-100 text-amber-700 rounded' },
                              '组长'
                            )
                          : null,
                        isDisabled
                          ? h(
                              'span',
                              {
                                class:
                                  'text-[9px] px-1 bg-slate-200 text-slate-500 rounded shrink-0 ml-auto',
                              },
                              props.disabledNote
                            )
                          : isSel(user.id)
                            ? h('span', { class: 'text-xs text-[#3B82F6] ml-auto' }, '✓')
                            : null,
                      ]
                    );
                  }),

                  !hasChildren && directUsers.length === 0
                    ? h('div', { class: 'px-3 py-2 text-xs text-slate-400' }, '暂无人员')
                    : null,
                ])
              : null,
          ]);
        })
      );
    };
  },
});
</script>

<template>
  <div ref="rootEl" class="relative">
    <div class="flex flex-wrap gap-1.5 mb-1.5">
      <span
        v-for="user in selectedUsers"
        :key="user.id"
        class="inline-flex items-center gap-1 px-2.5 py-1 bg-blue-50 text-blue-700 rounded-tag text-xs font-medium"
      >
        <span
          class="w-4 h-4 rounded-full bg-blue-200 flex items-center justify-center text-[9px] text-blue-600 font-bold"
        >
          {{ user.name.charAt(0) }}
        </span>
        {{ user.name }}
        <span
          v-if="user.role === 'group_leader'"
          class="text-[9px] px-1 bg-amber-100 text-amber-700 rounded"
          >组长</span
        >
        <button
          type="button"
          class="ml-0.5 hover:text-blue-900 transition-smooth"
          @click="removeUser(user.id)"
        >
          &times;
        </button>
      </span>
      <button
        type="button"
        class="inline-flex items-center px-2.5 py-1 border border-dashed border-slate-300 text-slate-400 rounded-tag text-xs hover:border-slate-400 transition-smooth"
        @click="open = !open"
      >
        + 选择人员
      </button>
    </div>

    <div
      v-if="open"
      class="absolute left-0 w-80 bg-white rounded-card shadow-modal border border-slate-100 z-50 overflow-hidden"
      :class="dropUp ? 'bottom-full mb-1' : 'top-full mt-1'"
    >
      <div class="p-3 border-b border-slate-100">
        <input
          v-model="searchText"
          class="input-field !text-xs"
          placeholder="搜索人员（支持拼音首字母）"
          @keydown.enter.prevent
        />
      </div>

      <div class="max-h-72 overflow-y-auto scrollbar-thin p-2">
        <div v-if="loading" class="text-center py-4 text-xs text-slate-400">加载中...</div>
        <div v-else-if="loadError" class="text-center py-4">
          <p class="text-xs text-red-400 mb-2">{{ loadError }}</p>
          <button
            type="button"
            class="text-xs px-3 py-1 bg-red-50 text-red-600 rounded-btn hover:bg-red-100 transition-smooth"
            @click="loadData"
          >
            重新加载
          </button>
        </div>
        <div v-else-if="searchText">
          <button
            v-for="user in filteredUsers"
            :key="user.id"
            type="button"
            :class="[
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-btn text-sm text-left transition-smooth',
              isDisabledUser(user.id)
                ? 'opacity-40 cursor-not-allowed'
                : isSelected(user.id)
                  ? 'bg-blue-50'
                  : 'hover:bg-slate-50',
            ]"
            @click="toggleUser(user.id)"
          >
            <div
              class="w-7 h-7 rounded-full bg-slate-200 flex items-center justify-center text-xs font-medium text-slate-600 shrink-0"
            >
              {{ user.name.charAt(0) }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm text-slate-900 truncate">{{ user.name }}</div>
              <div class="text-xs text-slate-400 truncate">{{ user.dept_name }}</div>
            </div>
            <span
              v-if="isDisabledUser(user.id)"
              class="text-[9px] px-1 bg-slate-200 text-slate-500 rounded shrink-0"
              >{{ disabledNote }}</span
            >
            <span v-else-if="isSelected(user.id)" class="text-xs text-[#3B82F6]">✓</span>
          </button>
        </div>
        <DeptTreeItem
          v-else
          :departments="departments"
          :expanded-set="expandedDepts"
          :user-list="users"
          :selected-ids="modelValue"
          :disabled-ids="disabledIds"
          :disabled-note="disabledNote"
          @toggle-dept="toggleDept"
          @toggle-user="toggleUser"
          @select-all="toggleSelectAllDept"
        />
      </div>

      <div class="border-t border-slate-100 p-2 flex justify-between">
        <span class="text-xs text-slate-400">{{ props.modelValue.length }} 人已选</span>
        <button
          type="button"
          class="text-xs px-3 py-1.5 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
          @click="open = false"
        >
          完成
        </button>
      </div>
    </div>
  </div>
</template>
