<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { UserBrief, Department } from '@/types'
import { getDepartments, getUsers } from '@/services/admin'

const props = withDefaults(defineProps<{
  modelValue: string[]
  multiple?: boolean
  max?: number
}>(), {
  multiple: true,
  max: 20,
})

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const open = ref(false)
const searchText = ref('')
const departments = ref<Department[]>([])
const users = ref<UserBrief[]>([])
const loading = ref(false)
const loadError = ref('')
const expandedDepts = ref<Set<string>>(new Set())

const selectedUsers = computed(() =>
  users.value.filter(u => props.modelValue.includes(u.id))
)

const filteredUsers = computed(() => {
  if (!searchText.value) return users.value
  const q = searchText.value.toLowerCase()
  return users.value.filter(u =>
    u.name.toLowerCase().includes(q) ||
    u.dept_name.toLowerCase().includes(q)
  )
})

onMounted(async () => {
  loading.value = true
  try {
    const [deptRes, userRes] = await Promise.all([
      getDepartments(false),
      getUsers({ page: 1 }),
    ])
    departments.value = deptRes.data as unknown as Department[]
    users.value = (userRes.data as unknown as { data: UserBrief[] }).data || []
  } catch {
    loadError.value = '加载组织架构失败'
  } finally {
    loading.value = false
  }
})

function toggleUser(userId: string) {
  const current = [...props.modelValue]
  const idx = current.indexOf(userId)
  if (idx >= 0) {
    current.splice(idx, 1)
  } else if (props.multiple) {
    if (current.length < props.max) {
      current.push(userId)
    }
  } else {
    emit('update:modelValue', [userId])
    open.value = false
    return
  }
  emit('update:modelValue', current)
}

function removeUser(userId: string) {
  const current = props.modelValue.filter(id => id !== userId)
  emit('update:modelValue', current)
}

function isSelected(userId: string): boolean {
  return props.modelValue.includes(userId)
}

function toggleDept(deptId: string) {
  const current = new Set(expandedDepts.value)
  if (current.has(deptId)) {
    current.delete(deptId)
  } else {
    current.add(deptId)
  }
  expandedDepts.value = current
}

function getDeptUsers(deptId: string): UserBrief[] {
  return users.value.filter(u => {
    const dept = departments.value.find(d => d.id === deptId)
    return dept && u.dept_name === dept.name
  })
}
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
        <span v-if="user.role === 'group_leader'" class="text-[9px] px-1 bg-amber-100 text-amber-700 rounded">组长</span>
        <button class="ml-0.5 hover:text-blue-900 transition-smooth" @click="removeUser(user.id)">&times;</button>
      </span>
      <button
        class="inline-flex items-center px-2.5 py-1 border border-dashed border-slate-300 text-slate-400 rounded-tag text-xs hover:border-slate-400 transition-smooth"
        @click="open = !open"
      >
        + 选择人员
      </button>
    </div>

    <div
      v-if="open"
      class="absolute top-full left-0 mt-1 w-80 bg-white rounded-card shadow-modal border border-slate-100 z-50 overflow-hidden"
    >
      <div class="p-3 border-b border-slate-100">
        <input
          v-model="searchText"
          class="input-field !text-xs"
          placeholder="搜索人员（支持拼音首字母）"
        />
      </div>

      <div class="max-h-72 overflow-y-auto scrollbar-thin p-2">
        <div v-if="loading" class="text-center py-4 text-xs text-slate-400">加载中...</div>
        <div v-else-if="loadError" class="text-center py-4 text-xs text-red-400">{{ loadError }}</div>
        <div v-else-if="searchText">
          <button
            v-for="user in filteredUsers"
            :key="user.id"
            :class="[
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-btn text-sm text-left transition-smooth',
              isSelected(user.id) ? 'bg-blue-50' : 'hover:bg-slate-50'
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
        <div v-else class="space-y-1">
          <div v-for="dept in departments" :key="dept.id">
            <button
              class="w-full flex items-center gap-2 px-3 py-2.5 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50"
              @click="toggleDept(dept.id)"
            >
              <svg
                :class="['w-3.5 h-3.5 text-slate-400 transition-transform', expandedDepts.has(dept.id) ? 'rotate-90' : '']"
                fill="none" viewBox="0 0 24 24" stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <span class="font-medium text-slate-700">{{ dept.name }}</span>
              <span class="text-xs text-slate-400 ml-auto">{{ dept.member_count || getDeptUsers(dept.id).length }}</span>
            </button>

            <div v-if="expandedDepts.has(dept.id)" class="ml-6 space-y-1">
              <button
                v-for="user in getDeptUsers(dept.id)"
                :key="user.id"
                :class="[
                  'w-full flex items-center gap-3 px-3 py-2 rounded-btn text-sm text-left transition-smooth',
                  isSelected(user.id) ? 'bg-blue-50' : 'hover:bg-slate-50'
                ]"
                @click="toggleUser(user.id)"
              >
                <div
                  class="w-6 h-6 rounded-full bg-slate-200 flex items-center justify-center text-[10px] font-medium text-slate-600 shrink-0"
                >
                  {{ user.name.charAt(0) }}
                </div>
                <span class="text-sm text-slate-900 truncate">{{ user.name }}</span>
                <span v-if="user.role === 'group_leader'" class="text-[9px] px-1 bg-amber-100 text-amber-700 rounded">组长</span>
                <span v-if="isSelected(user.id)" class="text-xs text-[#3B82F6] ml-auto">✓</span>
              </button>
              <div v-if="getDeptUsers(dept.id).length === 0" class="px-3 py-2 text-xs text-slate-400">
                暂无人员
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="border-t border-slate-100 p-2 flex justify-between">
        <span class="text-xs text-slate-400">{{ props.modelValue.length }} 人已选</span>
        <button
          class="text-xs px-3 py-1.5 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
          @click="open = false"
        >
          完成
        </button>
      </div>
    </div>
  </div>
</template>
