<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { fetchPersonalStats, generateAIReport } from '@/services/analytics'
import type { PersonalStatsData } from '@/services/analytics'

type Period = 'week' | 'month' | 'year'

const activePeriod = ref<Period>('week')
const stats = ref<PersonalStatsData | null>(null)
const loading = ref(false)
const error = ref('')

const reportLoading = ref(false)
const reportContent = ref('')
const reportError = ref('')
const reportGenerated = ref(false)

const periodOptions: { key: Period; label: string; icon: string }[] = [
  { key: 'week', label: '本周', icon: '📅' },
  { key: 'month', label: '本月', icon: '📆' },
  { key: 'year', label: '本年度', icon: '📊' },
]

const maxTrendCount = computed(() => {
  if (!stats.value?.daily_trend?.length) return 1
  return Math.max(...stats.value.daily_trend.map(d => d.count), 1)
})

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetchPersonalStats(activePeriod.value)
    stats.value = res.data as PersonalStatsData
  } catch {
    error.value = '加载统计数据失败'
    stats.value = null
  } finally {
    loading.value = false
  }
}

function switchPeriod(period: Period) {
  activePeriod.value = period
  reportGenerated.value = false
  reportContent.value = ''
  loadStats()
}

async function handleGenerateReport() {
  reportLoading.value = true
  reportError.value = ''
  reportContent.value = ''
  try {
    const res = await generateAIReport(activePeriod.value)
    const data = res.data as { report: string }
    reportContent.value = data.report || ''
    reportGenerated.value = true
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'AI报告生成失败'
    reportError.value = msg
  } finally {
    reportLoading.value = false
  }
}

function formatPercent(v: number): string {
  return v.toFixed(1) + '%'
}

function formatHours(v: number): string {
  if (v < 1) return (v * 60).toFixed(0) + ' 分钟'
  return v.toFixed(1) + ' 小时'
}

function copyReport() {
  if (!reportContent.value) return
  navigator.clipboard.writeText(reportContent.value)
}

const renderedReport = computed(() => {
  if (!reportContent.value) return ''
  return reportContent.value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/## (.+)/g, '<h3 class="text-base font-semibold mt-4 mb-2 text-slate-800 dark:text-slate-200">$1</h3>')
    .replace(/### (.+)/g, '<h4 class="text-sm font-semibold mt-3 mb-1 text-slate-700 dark:text-slate-300">$1</h4>')
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold">$1</strong>')
    .replace(/- (.+)/g, '<li class="ml-4 text-sm text-slate-600 dark:text-slate-400">$1</li>')
    .replace(/\n/g, '<br>')
})

function downloadReport() {
  if (!reportContent.value) return
  const blob = new Blob([reportContent.value], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `工作成效报告_${new Date().toISOString().slice(0, 10)}.md`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => {
  loadStats()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-slate-900 transition-colors duration-300">
    <!-- Header -->
    <div class="shrink-0 px-6 py-4 border-b border-slate-200 dark:border-slate-700">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">工作成效分析</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">个人工作数据统计、趋势分析与AI智能报告</p>
    </div>

    <!-- Period Tabs -->
    <div class="shrink-0 px-6 pt-4">
      <div class="flex gap-2">
        <button
          v-for="opt in periodOptions"
          :key="opt.key"
          :class="[
            'flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-smooth',
            activePeriod === opt.key
              ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 shadow-sm'
              : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
          ]"
          @click="switchPeriod(opt.key)"
        >
          <span>{{ opt.icon }}</span>
          <span>{{ opt.label }}</span>
        </button>
      </div>
    </div>

    <div class="flex-1 overflow-auto p-6">
      <!-- Error -->
      <div
        v-if="error"
        class="mb-4 px-4 py-3 rounded-lg text-sm bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800"
      >
        {{ error }}
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
      </div>

      <template v-else-if="stats">
        <!-- KPI Cards -->
        <div class="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-6">
          <div class="p-4 rounded-xl bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800">
            <p class="text-[11px] font-medium text-blue-500 dark:text-blue-400 uppercase tracking-wide mb-1">创建任务</p>
            <p class="text-2xl font-bold text-blue-700 dark:text-blue-300">{{ stats.total_created }}</p>
          </div>
          <div class="p-4 rounded-xl bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800">
            <p class="text-[11px] font-medium text-green-500 dark:text-green-400 uppercase tracking-wide mb-1">已完成</p>
            <p class="text-2xl font-bold text-green-700 dark:text-green-300">{{ stats.total_completed }}</p>
          </div>
          <div class="p-4 rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800">
            <p class="text-[11px] font-medium text-amber-500 dark:text-amber-400 uppercase tracking-wide mb-1">完成率</p>
            <p class="text-2xl font-bold text-amber-700 dark:text-amber-300">{{ formatPercent(stats.completion_rate) }}</p>
          </div>
          <div class="p-4 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800">
            <p class="text-[11px] font-medium text-red-500 dark:text-red-400 uppercase tracking-wide mb-1">被盯办</p>
            <p class="text-2xl font-bold text-red-700 dark:text-red-300">{{ stats.remind_received }}</p>
          </div>
          <div class="p-4 rounded-xl bg-purple-50 dark:bg-purple-900/20 border border-purple-100 dark:border-purple-800">
            <p class="text-[11px] font-medium text-purple-500 dark:text-purple-400 uppercase tracking-wide mb-1">平均耗时</p>
            <p class="text-2xl font-bold text-purple-700 dark:text-purple-300">{{ formatHours(stats.avg_completion_hours) }}</p>
          </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <!-- Trend Chart -->
          <div class="p-5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40">
            <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">每日任务趋势</h3>
            <div v-if="stats.daily_trend.length === 0" class="text-center py-8 text-slate-400 text-sm">暂无数据</div>
            <div v-else class="flex items-end gap-1 h-32">
              <div
                v-for="(d, i) in stats.daily_trend"
                :key="i"
                class="flex-1 flex flex-col items-center gap-1 group relative"
              >
                <div
                  class="w-full bg-blue-400 dark:bg-blue-500 rounded-t-sm transition-all hover:bg-blue-500 dark:hover:bg-blue-400 min-h-[4px]"
                  :style="{ height: Math.max((d.count / maxTrendCount) * 100, 4) + '%' }"
                ></div>
                <span class="text-[9px] text-slate-400 dark:text-slate-500 truncate w-full text-center">
                  {{ d.date.slice(5) }}
                </span>
                <div class="absolute -top-6 opacity-0 group-hover:opacity-100 bg-slate-800 text-white text-[10px] px-1.5 py-0.5 rounded whitespace-nowrap transition-opacity pointer-events-none">
                  {{ d.date }}：{{ d.count }} 条
                </div>
              </div>
            </div>
          </div>

          <!-- Tag Breakdown -->
          <div class="p-5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40">
            <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">标签使用分布</h3>
            <div v-if="stats.tag_breakdown.length === 0" class="text-center py-8 text-slate-400 text-sm">暂无数据</div>
            <div v-else class="space-y-2.5">
              <div v-for="t in stats.tag_breakdown" :key="t.tag_name" class="flex items-center gap-3">
                <span class="text-xs text-slate-600 dark:text-slate-400 w-20 truncate" :title="t.tag_name">{{ t.tag_name }}</span>
                <div class="flex-1 h-5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-gradient-to-r from-blue-400 to-blue-500 rounded-full transition-all"
                    :style="{ width: (t.count / Math.max(...stats.tag_breakdown.map(x => x.count), 1) * 100) + '%' }"
                  ></div>
                </div>
                <span class="text-xs font-mono text-slate-500 dark:text-slate-400 w-8 text-right">{{ t.count }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- AI Report Section -->
        <div class="rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40 p-5">
          <div class="flex items-center justify-between mb-4">
            <div>
              <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">🤖 AI 智能报告</h3>
              <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">基于当前统计数据自动生成结构化工作报告</p>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="!reportGenerated"
                class="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 rounded-lg transition-smooth disabled:opacity-50 flex items-center gap-2"
                :disabled="reportLoading"
                @click="handleGenerateReport()"
              >
                <span v-if="reportLoading" class="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"></span>
                <span>{{ reportLoading ? '生成中...' : '生成报告' }}</span>
              </button>
              <template v-if="reportGenerated && reportContent">
                <button
                  class="px-3 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-lg border border-slate-200 dark:border-slate-600 transition-smooth"
                  @click="copyReport()"
                  title="复制报告"
                >📋 复制</button>
                <button
                  class="px-3 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-lg border border-slate-200 dark:border-slate-600 transition-smooth"
                  @click="downloadReport()"
                  title="下载报告"
                >⬇️ 下载</button>
              </template>
            </div>
          </div>

          <div
            v-if="reportError"
            class="px-4 py-3 rounded-lg text-sm bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800 mb-3"
          >
            {{ reportError }}
          </div>

          <div v-if="!reportGenerated && !reportLoading" class="text-center py-8 text-slate-400 dark:text-slate-500">
            <p class="text-2xl mb-2">🤖</p>
            <p class="text-sm">点击"生成报告"按钮</p>
            <p class="text-xs mt-1">AI将基于当前数据自动分析并生成工作报告</p>
            <p class="text-xs text-slate-400 dark:text-slate-600 mt-2">需要先在系统管理 → AI服务配置中设置可用的AI服务</p>
          </div>

          <div v-if="reportContent" class="prose prose-sm dark:prose-invert max-w-none bg-white dark:bg-slate-900 rounded-lg p-5 border border-slate-200 dark:border-slate-700">
            <div v-html="renderedReport"></div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
