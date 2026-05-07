<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getGroupReport, exportGroupReport, type WorkGroupReport } from '@/services/workgroup'

const route = useRoute()
const router = useRouter()
const groupId = route.params.id as string
const reportId = route.params.reportId as string

const report = ref<WorkGroupReport | null>(null)
const loading = ref(true)
const error = ref('')

async function loadReport() {
  try {
    const res = await getGroupReport(groupId, reportId)
    report.value = res.data as unknown as WorkGroupReport
  } catch {
    error.value = '报告加载失败'
  } finally { loading.value = false }
}

function handleExport(format: string) {
  exportGroupReport(groupId, reportId, format).then(blob => {
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const ext = format === 'docx' ? 'docx' : format === 'pdf' ? 'pdf' : 'html'
    a.download = `report.${ext}`
    a.click()
    URL.revokeObjectURL(url)
  })
}

const renderedContent = computed(() => {
  if (!report.value) return ''
  const content = report.value.content
  const lines = content.split('\n')
  const result: string[] = []
  let inTable = false
  for (const line of lines) {
    const t = line.trim()
    if (t.startsWith('|') && t.includes('|')) {
      if (!inTable) { result.push('<table class="w-full border-collapse my-3 text-sm">'); inTable = true }
      if (t.includes('---')) continue
      const cells = t.split('|').map(c => c.trim()).filter(Boolean)
      result.push('<tr>' + cells.map(c => `<td class="border border-slate-300 dark:border-slate-600 px-3 py-1.5">${c}</td>`).join('') + '</tr>')
    } else {
      if (inTable) { result.push('</table>'); inTable = false }
      if (/^#{1,2}\s/.test(t)) result.push(`<h2 class="text-lg font-bold text-slate-900 dark:text-slate-100 mt-6 mb-2 pb-1.5 border-b border-indigo-400 dark:border-indigo-500">${t.replace(/^#+\s*/, '')}</h2>`)
      else if (/^#{3}\s/.test(t)) result.push(`<h3 class="text-base font-semibold text-slate-800 dark:text-slate-200 mt-4 mb-1">${t.replace(/^#+\s*/, '')}</h3>`)
      else if (t === '') result.push('<br>')
      else if (t.startsWith('- ') || t.startsWith('* ')) result.push(`<li class="ml-5 my-0.5">${t.slice(2)}</li>`)
      else if (t.startsWith('> ')) result.push(`<blockquote class="border-l-3 border-indigo-400 pl-3 py-1 my-2 text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800/50 rounded-r">${t.slice(2)}</blockquote>`)
      else if (t.startsWith('---')) result.push('<hr class="my-4 border-slate-200 dark:border-slate-700">')
      else if (t.match(/^\d+\.\s/)) result.push(`<p class="my-1"><span class="font-medium">${t.match(/^\d+\.\s/)![0]}</span>${t.replace(/^\d+\.\s*/, '')}</p>`)
      else result.push(`<p class="my-1.5 leading-relaxed">${t}</p>`)
    }
  }
  if (inTable) result.push('</table>')
  return result.join('\n')
})

onMounted(loadReport)
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950">
    <!-- Top bar -->
    <div class="sticky top-0 z-10 shrink-0 flex items-center justify-between px-6 py-3 border-b border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-950/90 backdrop-blur">
      <div class="flex items-center gap-4">
        <button class="text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 transition-smooth text-sm flex items-center gap-1" @click="router.push(`/workbench/groups/${groupId}/reports`)">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
          返回报告列表
        </button>
        <h1 class="text-lg font-bold text-slate-900 dark:text-slate-100 truncate max-w-md" v-if="report">{{ report.title }}</h1>
      </div>
      <div class="flex items-center gap-2">
        <button class="px-3 py-1.5 rounded-lg text-xs font-medium bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-smooth" @click="handleExport('html')">导出HTML</button>
        <button class="px-3 py-1.5 rounded-lg text-xs font-medium bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 transition-smooth" @click="handleExport('docx')">导出Word</button>
        <button class="px-3 py-1.5 rounded-lg text-xs font-medium bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-smooth" @click="handleExport('pdf')">导出PDF</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="animate-spin h-8 w-8 border-2 border-purple-500 border-t-transparent rounded-full"></div>
    </div>
    <div v-else-if="error" class="flex items-center justify-center py-24"><p class="text-red-500">{{ error }}</p></div>

    <!-- Content -->
    <div v-else-if="report" class="max-w-4xl mx-auto px-6 py-8">
      <div v-if="report.report_type === 'ai'" class="inline-block px-2 py-1 rounded bg-purple-50 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 text-[10px] font-medium mb-4">🤖 AI 生成</div>
      <div v-else class="inline-block px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 text-[10px] font-medium mb-4">📋 模板生成</div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-8 shadow-sm" v-html="renderedContent"></div>
    </div>
  </div>
</template>
