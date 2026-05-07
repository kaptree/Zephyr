<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listGroupReports, deleteGroupReport, exportGroupReport } from '@/services/workgroup'
import type { WorkGroupReport } from '@/services/workgroup'

const route = useRoute()
const router = useRouter()
const groupId = route.params.id as string

const reports = ref<WorkGroupReport[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const pageSize = 20

async function loadReports() {
  loading.value = true
  try {
    const res = await listGroupReports(groupId, { page: page.value, page_size: pageSize })
    reports.value = (res.data as unknown as { data: WorkGroupReport[] }).data || []
    total.value = (res.data as unknown as { total: number }).total || 0
  } finally {
    loading.value = false
  }
}

async function handleDelete(reportId: string) {
  if (!confirm('确定删除此报告？')) return
  await deleteGroupReport(groupId, reportId)
  loadReports()
}

function handleExport(reportId: string, format: string) {
  exportGroupReport(groupId, reportId, format).then(blob => {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `report.${format === 'docx' ? 'docx' : format === 'pdf' ? 'pdf' : 'html'}`
    a.click()
    URL.revokeObjectURL(url)
  })
}

onMounted(loadReports)
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 p-6">
    <div class="max-w-4xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 transition-smooth text-sm flex items-center gap-1"
          @click="router.push(`/workbench/groups/${groupId}`)"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
          返回工作组
        </button>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">工作报告</h1>
      </div>

      <div v-if="loading" class="text-center py-12">
        <div class="animate-spin h-8 w-8 border-2 border-purple-500 border-t-transparent rounded-full mx-auto mb-3"></div>
        <p class="text-slate-500 text-sm">加载报告列表...</p>
      </div>

      <div v-else-if="reports.length === 0" class="text-center py-12 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800">
        <svg class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
        <p class="text-slate-500 text-sm">暂未生成报告</p>
        <p class="text-slate-400 text-xs mt-1">返回工作组页面，点击「生成报告」开始</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="r in reports"
          :key="r.id"
          class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 hover:border-purple-300 dark:hover:border-purple-700 transition-smooth"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="flex-1 min-w-0 cursor-pointer" @click="router.push(`/workbench/groups/${groupId}/reports/${r.id}`)">
              <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">{{ r.title }}</h3>
              <div class="flex items-center gap-3 mt-2 text-xs text-slate-500">
                <span>{{ r.user_name }} 生成</span>
                <span v-if="r.report_type === 'ai'" class="px-1.5 py-0.5 rounded bg-purple-50 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 text-[10px] font-medium">AI生成</span>
                <span v-else class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 text-[10px] font-medium">模板生成</span>
                <span>{{ r.created_at?.slice(0, 10) }}</span>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <button class="px-3 py-1.5 text-[10px] font-medium rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-smooth" @click="handleExport(r.id, 'html')">HTML</button>
              <button class="px-3 py-1.5 text-[10px] font-medium rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 transition-smooth" @click="handleExport(r.id, 'docx')">Word</button>
              <button class="px-3 py-1.5 text-[10px] font-medium rounded-lg bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 hover:bg-red-200 dark:hover:bg-red-900/50 transition-smooth" @click="handleDelete(r.id)">删除</button>
            </div>
          </div>
        </div>

        <div v-if="total > pageSize" class="flex items-center justify-center gap-4 mt-6">
          <button class="px-4 py-2 text-sm bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg disabled:opacity-40" :disabled="page <= 1" @click="page--; loadReports()">上一页</button>
          <span class="text-sm text-slate-500">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
          <button class="px-4 py-2 text-sm bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg disabled:opacity-40" :disabled="page >= Math.ceil(total / pageSize)" @click="page++; loadReports()">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>
