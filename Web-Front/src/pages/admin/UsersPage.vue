<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUsers } from '@/services/admin'

const users = ref<{ id: string; name: string; dept_name: string; role: string }[]>([])
const loading = ref(false)
const loadError = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const res = await getUsers({ page: 1 })
    users.value = (res.data as unknown as { data: { id: string; name: string; dept_name: string; role: string }[] }).data || []
  } catch {
    loadError.value = '加载人员列表失败'
  } finally {
    loading.value = false
  }
})

const roleMap: Record<string, string> = {
  super_admin: '系统管理员',
  dept_admin: '部门管理员',
  group_leader: '组长',
  user: '普通民警',
  screen_role: '大屏角色',
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">人员库管理</h2>
      <button class="btn-primary text-sm">新建人员</button>
    </div>

    <div class="bg-white rounded-card border border-slate-100 overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-sm text-slate-400">加载中...</div>
      <div v-else-if="loadError" class="p-8 text-center text-sm text-red-400">{{ loadError }}</div>
      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50">
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">姓名</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">部门</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">角色</th>
            <th class="text-right px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50/50 transition-smooth">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-xs font-medium text-blue-600">
                  {{ user.name.charAt(0) }}
                </div>
                <span class="text-sm font-medium text-slate-900">{{ user.name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-sm text-slate-600">{{ user.dept_name }}</td>
            <td class="px-4 py-3">
              <span class="text-xs px-2 py-0.5 rounded-tag font-medium" :class="{
                'bg-amber-100 text-amber-700': user.role === 'super_admin',
                'bg-blue-100 text-blue-700': user.role === 'dept_admin',
                'bg-green-100 text-green-700': user.role === 'group_leader',
                'bg-slate-100 text-slate-600': user.role === 'user',
              }">
                {{ roleMap[user.role] || user.role }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <button class="text-xs px-2 py-1 bg-slate-100 text-slate-600 rounded hover:bg-slate-200 transition-smooth">编辑</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
