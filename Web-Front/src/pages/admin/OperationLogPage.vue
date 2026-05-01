<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listOperationLogs, getOperationActions } from '@/services/system'
import type { OperationLogItem } from '@/types/system'

const logs = ref<OperationLogItem[]>([])
const loading = ref(false)
const page = ref(1)
const total = ref(0)
const pageSize = 20
const toast = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const filter = ref({
  user_name: '',
  action: '',
  method: '',
  date_from: '',
  date_to: '',
})

const availableActions = ref<string[]>([])

function showToast(type: 'success' | 'error', message: string) {
  toast.value = { type, message }
  setTimeout(() => { toast.value = null }, 3000)
}

async function loadAvailableActions() {
  try {
    const res = await getOperationActions()
    availableActions.value = res.data || []
  } catch { /* ignore */ }
}

async function loadLogs() {
  loading.value = true
  try {
    const res = await listOperationLogs({
      page: page.value,
      page_size: pageSize,
      user_name: filter.value.user_name || undefined,
      action: filter.value.action || undefined,
      method: filter.value.method || undefined,
      date_from: filter.value.date_from || undefined,
      date_to: filter.value.date_to || undefined,
    })
    const data = res.data as unknown as { data: OperationLogItem[]; total: number }
    logs.value = data.data || []
    total.value = data.total || 0
  } catch {
    showToast('error', '加载操作日志失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  page.value = 1
  loadLogs()
}

function resetFilters() {
  filter.value = { user_name: '', action: '', method: '', date_from: '', date_to: '' }
  page.value = 1
  loadLogs()
}

function prevPage() {
  if (page.value > 1) { page.value--; loadLogs() }
}

function nextPage() {
  if (page.value * pageSize < total.value) { page.value++; loadLogs() }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const roleLabels: Record<string, string> = {
  super_admin: '超级管理员',
  dept_admin: '部门管理员',
  group_leader: '组长',
  member: '普通用户',
}

const actionLabels: Record<string, string> = {
  login: '用户登录',
  logout: '用户登出',
  refresh_token: '刷新令牌',
  create_note: '创建便签',
  update_note: '更新便签',
  complete_note: '完成任务',
  remind_note: '盯办提醒',
  restore_note: '恢复便签',
  delete_note: '删除便签',
  create_tag: '创建标签',
  update_tag: '更新标签',
  delete_tag: '删除标签',
  create_department: '创建部门',
  update_department: '更新部门',
  delete_department: '删除部门',
  create_user: '创建用户',
  update_user: '更新用户',
  delete_user: '删除用户',
  create_group: '创建工作组',
  update_group: '更新工作组',
  create_template: '创建模板',
  update_template: '更新模板',
  update_system_config: '更新系统配置',
  create_ai_config: '创建AI配置',
  update_ai_config: '更新AI配置',
  delete_ai_config: '删除AI配置',
  update_config_file: '编辑配置文件',
  send_command: '下发指令',
}

function statusClass(code: number): string {
  if (code >= 200 && code < 300) return 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-400'
  if (code >= 400 && code < 500) return 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-700 dark:text-yellow-400'
  if (code >= 500) return 'bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-400'
  return 'bg-slate-100 dark:bg-slate-800 text-slate-500'
}

onMounted(() => {
  loadAvailableActions()
  loadLogs()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-slate-900 transition-colors duration-300">
    <div class="shrink-0 px-6 py-4 border-b border-slate-200 dark:border-slate-700">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">操作日志</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">查看所有用户的操作记录，支持多维度筛选</p>
    </div>

    <div class="flex-1 overflow-auto p-6">
      <div
        v-if="toast"
        :class="[
          'mb-4 px-4 py-3 rounded-lg text-sm flex items-center gap-2',
          toast.type === 'success' ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800' : 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
        ]"
      >
        <span>{{ toast.type === 'success' ? '✅' : '❌' }}</span>
        <span>{{ toast.message }}</span>
      </div>

      <div class="mb-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
          <div>
            <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">用户名</label>
            <input
              v-model="filter.user_name"
              type="text"
              class="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
              placeholder="输入用户名搜索..."
              @keyup.enter="applyFilters()"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">操作类型</label>
            <select
              v-model="filter.action"
              class="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
              @change="applyFilters()"
            >
              <option value="">全部类型</option>
              <option v-for="a in availableActions" :key="a" :value="a">{{ actionLabels[a] || a }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">请求方法</label>
            <select
              v-model="filter.method"
              class="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
              @change="applyFilters()"
            >
              <option value="">全部</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">开始日期</label>
            <input
              v-model="filter.date_from"
              type="date"
              class="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
              @change="applyFilters()"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">截止日期</label>
            <div class="flex gap-1.5">
              <input
                v-model="filter.date_to"
                type="date"
                class="flex-1 px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                @change="applyFilters()"
              />
              <button
                class="px-2.5 py-1.5 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-smooth shrink-0"
                title="重置筛选"
                @click="resetFilters()"
              >🔄</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
      </div>

      <div v-else-if="logs.length === 0" class="text-center py-16 text-slate-400 dark:text-slate-500">
        <p class="text-2xl mb-2">📋</p>
        <p class="text-sm">暂无操作记录</p>
        <p class="text-xs mt-1">用户操作记录将自动出现在这里</p>
      </div>

      <div v-else>
        <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-slate-50 dark:bg-slate-800/50">
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">操作时间</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">用户</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">角色</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">操作类型</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">方法</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">状态</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">IP地址</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr
                v-for="log in logs"
                :key="log.id"
                class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-smooth"
              >
                <td class="px-4 py-2.5 text-xs text-slate-600 dark:text-slate-400 whitespace-nowrap">
                  {{ formatTime(log.created_at) }}
                </td>
                <td class="px-4 py-2.5">
                  <span class="text-xs font-medium text-slate-900 dark:text-slate-100">{{ log.user_name }}</span>
                </td>
                <td class="px-4 py-2.5">
                  <span class="text-[10px] font-medium text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded">
                    {{ roleLabels[log.role] || log.role }}
                  </span>
                </td>
                <td class="px-4 py-2.5">
                  <span class="text-xs text-slate-700 dark:text-slate-300">{{ actionLabels[log.action] || log.action }}</span>
                </td>
                <td class="px-4 py-2.5">
                  <span :class="[
                    'text-[10px] font-mono font-bold px-1.5 py-0.5 rounded',
                    log.method === 'POST' ? 'bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-400' :
                    log.method === 'PUT' ? 'bg-amber-100 dark:bg-amber-900/50 text-amber-700 dark:text-amber-400' :
                    log.method === 'DELETE' ? 'bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-400' :
                    'bg-slate-100 dark:bg-slate-800 text-slate-500'
                  ]">{{ log.method }}</span>
                </td>
                <td class="px-4 py-2.5">
                  <span :class="['text-[10px] font-mono px-1.5 py-0.5 rounded', statusClass(log.status_code)]">
                    {{ log.status_code }}
                  </span>
                </td>
                <td class="px-4 py-2.5 text-xs text-slate-400 dark:text-slate-500 font-mono">
                  {{ log.ip_address }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex items-center justify-between mt-4">
          <span class="text-xs text-slate-400 dark:text-slate-500">
            共 {{ total }} 条记录，第 {{ page }} 页
          </span>
          <div class="flex items-center gap-2">
            <button
              class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
              :disabled="page <= 1"
              @click="prevPage()"
            >上一页</button>
            <button
              class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
              :disabled="page * pageSize >= total"
              @click="nextPage()"
            >下一页</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
