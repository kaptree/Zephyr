<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDepartments } from '@/services/admin'
import type { Department } from '@/types'

const departments = ref<Department[]>([])
const loading = ref(false)
const loadError = ref('')
const expanded = ref<Set<string>>(new Set())
const selectedDept = ref<Department | null>(null)

onMounted(async () => {
  loading.value = true
  try {
    const res = await getDepartments(false)
    departments.value = res.data as unknown as Department[]
    if (departments.value.length > 0) {
      expanded.value.add(departments.value[0].id)
    }
  } catch {
    loadError.value = '加载部门架构失败'
  } finally {
    loading.value = false
  }
})

function toggleExpand(id: string) {
  const s = new Set(expanded.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  expanded.value = s
}

function selectDept(dept: Department) {
  selectedDept.value = dept
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">部门库管理</h2>
      <button class="btn-primary text-sm">新建部门</button>
    </div>

    <div class="grid grid-cols-2 gap-6">
      <!-- 部门树 -->
      <div class="bg-white rounded-card border border-slate-100 p-4">
        <h4 class="text-sm font-semibold text-slate-700 mb-3">组织架构</h4>
        <div v-if="loading" class="skeleton h-40 rounded-lg" />
        <div v-else-if="loadError" class="text-sm text-red-400 py-8 text-center">{{ loadError }}</div>
        <div v-else class="space-y-1">
          <template v-for="dept in departments" :key="dept.id">
            <button
              class="w-full flex items-center gap-2 px-3 py-2 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50"
              :class="{ 'bg-blue-50 text-blue-700 font-medium': selectedDept?.id === dept.id }"
              @click="toggleExpand(dept.id); selectDept(dept)"
            >
              <svg :class="['w-3.5 h-3.5 text-slate-400 transition-transform', expanded.has(dept.id) ? 'rotate-90' : '']" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              <span class="flex-1 truncate">{{ dept.name }}</span>
              <span class="text-xs text-slate-400">{{ dept.member_count }}人</span>
            </button>
            <div v-if="expanded.has(dept.id) && dept.children" class="ml-6 space-y-1 border-l border-slate-100 pl-4">
              <template v-for="child in dept.children" :key="child.id">
                <button
                  class="w-full flex items-center gap-2 px-3 py-2 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50"
                  :class="{ 'bg-blue-50 text-blue-700 font-medium': selectedDept?.id === child.id }"
                  @click="selectDept(child)"
                >
                  <span class="flex-1 truncate">{{ child.name }}</span>
                  <span class="text-xs text-slate-400">{{ child.member_count }}人</span>
                </button>
              </template>
            </div>
          </template>
        </div>
      </div>

      <!-- 部门详情 -->
      <div class="bg-white rounded-card border border-slate-100 p-4">
        <h4 class="text-sm font-semibold text-slate-700 mb-3">部门详情</h4>
        <div v-if="!selectedDept" class="text-sm text-slate-400 py-8 text-center">
          请在左侧选择一个部门
        </div>
        <div v-else class="space-y-3">
          <div class="flex items-center gap-2">
            <span class="text-xs text-slate-400 w-16">名称</span>
            <input class="input-field !py-1.5 !text-sm" :value="selectedDept.name" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-slate-400 w-16">上级</span>
            <span class="text-sm text-slate-700">{{ selectedDept.parent_id || '—' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-slate-400 w-16">人数</span>
            <span class="text-sm text-slate-700">{{ selectedDept.member_count }} 人</span>
          </div>
          <div class="flex gap-2 pt-3 border-t border-slate-100">
            <button class="btn-primary text-xs !py-1.5 !px-3">保存</button>
            <button class="px-3 py-1.5 text-xs bg-red-50 text-red-600 rounded-btn hover:bg-red-100 transition-smooth">删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
