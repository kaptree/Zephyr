<script setup lang="ts">
import { ref, computed, onMounted, h, defineComponent, type PropType } from 'vue';
import type { UserBrief, Department } from '@/types';
import { getDepartments, getUsers } from '@/services/admin';

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    multiple?: boolean;
    max?: number;
    dropUp?: boolean;
  }>(),
  {
    multiple: true,
    max: 20,
    dropUp: false,
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string[]];
}>();

const open = ref(false);
const searchText = ref('');
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

onMounted(loadData);

function toggleUser(userId: string) {
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

function getDirectDeptUsers(deptId: string): UserBrief[] {
  return users.value.filter((u) => (u as any).dept_id === deptId);
}

const DeptTreeItem = defineComponent({
  name: 'DeptTreeItem',
  props: {
    departments: { type: Array as PropType<Department[]>, required: true },
    expandedSet: { type: Object as PropType<Set<string>>, required: true },
    userList: { type: Array as PropType<UserBrief[]>, required: true },
    selectedIds: { type: Array as PropType<string[]>, required: true },
  },
  emits: ['toggle-dept', 'toggle-user'],
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

    return () => {
      return h(
        'div',
        { class: 'space-y-1' },
        props.departments.map((dept) => {
          const isExpanded = props.expandedSet.has(dept.id);
          const directUsers = getDirect(dept.id);
          const hasChildren = dept.children && dept.children.length > 0;

          return h('div', { key: dept.id }, [
            h(
              'button',
              {
                type: 'button',
                class:
                  'w-full flex items-center gap-2 px-3 py-2.5 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50',
                onClick: () => onToggleDept(dept.id),
              },
              [
                h(
                  'svg',
                  {
                    class: `w-3.5 h-3.5 text-slate-400 transition-transform ${isExpanded ? 'rotate-90' : ''}`,
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
                h('span', { class: 'font-medium text-slate-700' }, dept.name),
                h(
                  'span',
                  { class: 'text-xs text-slate-400 ml-auto' },
                  String(dept.member_count || 0)
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
                        'onToggle-dept': onToggleDept,
                        'onToggle-user': onToggleUser,
                      })
                    : null,

                  ...directUsers.map((user) =>
                    h(
                      'button',
                      {
                        type: 'button',
                        class: `w-full flex items-center gap-3 px-3 py-2 rounded-btn text-sm text-left transition-smooth ${isSel(user.id) ? 'bg-blue-50' : 'hover:bg-slate-50'}`,
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
                        isSel(user.id)
                          ? h('span', { class: 'text-xs text-[#3B82F6] ml-auto' }, '✓')
                          : null,
                      ]
                    )
                  ),

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
  <div class="relative">
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
              isSelected(user.id) ? 'bg-blue-50' : 'hover:bg-slate-50',
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
            <span v-if="isSelected(user.id)" class="text-xs text-[#3B82F6]">✓</span>
          </button>
        </div>
        <DeptTreeItem
          v-else
          :departments="departments"
          :expanded-set="expandedDepts"
          :user-list="users"
          :selected-ids="modelValue"
          @toggle-dept="toggleDept"
          @toggle-user="toggleUser"
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
