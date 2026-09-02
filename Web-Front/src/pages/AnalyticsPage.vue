<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useRouter } from 'vue-router';
import { onTableFlowScroll } from '@/utils/tableFlow';
import AnimatedNumber from '@/components/common/AnimatedNumber.vue';
import FormCheckbox from '@/components/common/FormCheckbox.vue';
import {
  fetchPersonalStats,
  generateAIReport,
  listReports,
  deleteReport,
  fetchReportTemplate,
  saveReportTemplate,
  fetchTeamStats,
  generateTeamReport,
  listReportAIConfigs,
} from '@/services/analytics';
import type {
  PersonalStatsData,
  WorkReportItem,
  ReportTemplateData,
  TeamStatsData,
  TeamMemberStat,
  AIConfigBrief,
} from '@/services/analytics';

type Period = 'week' | 'month' | 'year';

const router = useRouter();
const activePeriod = ref<Period>('week');
const stats = ref<PersonalStatsData | null>(null);
const loading = ref(false);
const error = ref('');

const reportLoading = ref(false);
const reportContent = ref('');
const reportError = ref('');
const reportGenerated = ref(false);
const reportType = ref('');

const viewTab = ref<'stats' | 'history' | 'team'>('stats');

// ===== Tab 滑动指示条（美化工程 · 空间感知） =====
// 测量激活按钮的几何位置，指示条以 left/width 平滑滑动跟随
const tabBarRef = ref<HTMLElement | null>(null);
const indicatorStyle = ref<{ left: string; width: string }>({ left: '0px', width: '0px' });
function updateIndicator() {
  const el = tabBarRef.value?.querySelector<HTMLElement>('button[data-active="true"]');
  if (el) {
    indicatorStyle.value = { left: `${el.offsetLeft}px`, width: `${el.offsetWidth}px` };
  }
}
watch(viewTab, () => nextTick(updateIndicator));
onMounted(() => {
  nextTick(updateIndicator);
  window.addEventListener('resize', updateIndicator);
});
onUnmounted(() => window.removeEventListener('resize', updateIndicator));

const reports = ref<WorkReportItem[]>([]);
const reportsLoading = ref(false);
const reportsTotal = ref(0);
const reportsPage = ref(1);
const reportsPageSize = 20;
const reportsFilter = ref({
  keyword: '',
  period: '',
  date_from: '',
  date_to: '',
});

// ===== 团队报告模块（★ 新增） =====
type TeamPeriod = 'week' | 'month' | 'custom';
const teamPeriod = ref<TeamPeriod>('week');
const teamDateFrom = ref('');
const teamDateTo = ref('');
const teamStats = ref<TeamStatsData | null>(null);
const teamLoading = ref(false);
const teamError = ref('');
const selectedUserIds = ref<string[]>([]); // 自定义勾选成员（空 = 全部）
const teamMembersExpanded = ref(false); // 成员成效明细默认收起，点击展开
const aiModels = ref<AIConfigBrief[]>([]);
const selectedModelId = ref('');
const teamReportLoading = ref(false);
const teamReportContent = ref('');
const teamReportType = ref('');
const teamReportGenerated = ref(false);
const teamReportError = ref('');

function teamRangeForPeriod(p: TeamPeriod): { from: string; to: string } {
  const now = new Date();
  const fmt = (d: Date) => d.toISOString().slice(0, 10);
  if (p === 'month') {
    const first = new Date(now.getFullYear(), now.getMonth(), 1);
    return { from: fmt(first), to: fmt(now) };
  }
  // week：本周一至今天
  const day = now.getDay() || 7;
  const monday = new Date(now.getFullYear(), now.getMonth(), now.getDate() - day + 1);
  return { from: fmt(monday), to: fmt(now) };
}

async function loadTeamStats() {
  teamLoading.value = true;
  teamError.value = '';
  try {
    const params: { date_from?: string; date_to?: string } = {};
    if (teamDateFrom.value) params.date_from = teamDateFrom.value;
    if (teamDateTo.value) params.date_to = teamDateTo.value;
    // 始终拉取全部可见成员，勾选由前端本地过滤（保证可自由连续勾选多人）
    const res = await fetchTeamStats(params);
    teamStats.value = res.data as TeamStatsData;
    if (!teamDateFrom.value && !teamDateTo.value && teamStats.value) {
      teamDateFrom.value = teamStats.value.date_from;
      teamDateTo.value = teamStats.value.date_to;
    }
  } catch {
    teamError.value = '加载团队统计数据失败';
    teamStats.value = null;
  } finally {
    teamLoading.value = false;
  }
}

// ===== 成员勾选（组建自定义团队） =====
const teamMemberList = computed(() => teamStats.value?.members || []);

function toggleTeamMember(id: string) {
  const idx = selectedUserIds.value.indexOf(id);
  if (idx >= 0) selectedUserIds.value.splice(idx, 1);
  else selectedUserIds.value.push(id);
  teamReportGenerated.value = false;
  teamReportContent.value = '';
}

function selectAllMembers() {
  selectedUserIds.value = teamMemberList.value.map((m) => m.user_id);
  teamReportGenerated.value = false;
  teamReportContent.value = '';
}

function clearSelectedMembers() {
  selectedUserIds.value = [];
  teamReportGenerated.value = false;
  teamReportContent.value = '';
}

// 勾选后的团队汇总（本地过滤计算，供汇总卡片展示）
const teamSummaryStats = computed(() => {
  const all = teamStats.value;
  if (!all) return null;
  let members = all.members;
  if (selectedUserIds.value.length > 0) {
    const set = new Set(selectedUserIds.value);
    members = all.members.filter((m) => set.has(m.user_id));
  }
  let totalCreated = 0;
  let totalCompleted = 0;
  for (const m of members) {
    totalCreated += m.total_created;
    totalCompleted += m.total_completed;
  }
  const rate = totalCreated > 0 ? (totalCompleted / totalCreated) * 100 : 0;
  return {
    memberCount: members.length,
    totalCreated,
    totalCompleted,
    completionRate: rate,
  };
});

function switchTeamPeriod(p: TeamPeriod) {
  teamPeriod.value = p;
  teamReportGenerated.value = false;
  teamReportContent.value = '';
  if (p === 'custom') return; // 自定义：等待日期输入后点击查询
  const range = teamRangeForPeriod(p);
  teamDateFrom.value = range.from;
  teamDateTo.value = range.to;
  loadTeamStats();
}

function applyTeamRange() {
  if (!teamDateFrom.value || !teamDateTo.value) {
    showToast('error', '请选择完整的时间范围');
    return;
  }
  teamReportGenerated.value = false;
  teamReportContent.value = '';
  loadTeamStats();
}

async function loadAIModels() {
  try {
    const res = await listReportAIConfigs();
    aiModels.value = (res.data as unknown as AIConfigBrief[]) || [];
  } catch {
    aiModels.value = [];
  }
}

async function handleGenerateTeamReport() {
  teamReportLoading.value = true;
  teamReportError.value = '';
  teamReportContent.value = '';
  try {
    const res = await generateTeamReport({
      period: teamPeriod.value === 'custom' ? 'custom' : teamPeriod.value,
      date_from: teamDateFrom.value || undefined,
      date_to: teamDateTo.value || undefined,
      ai_config_id: selectedModelId.value || undefined,
      user_ids: selectedUserIds.value.length > 0 ? selectedUserIds.value : undefined,
    });
    const data = res.data as unknown as {
      report: string;
      report_type: string;
      report_id: string;
    };
    teamReportContent.value = data.report || '';
    teamReportType.value = data.report_type || '';
    teamReportGenerated.value = true;
    showToast('success', '团队报告生成成功，已保存到历史记录');
    loadReports();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '团队报告生成失败';
    teamReportError.value = msg;
    showToast('error', msg);
  } finally {
    teamReportLoading.value = false;
  }
}

function formatTeamHours(v: number): string {
  if (v <= 0) return '-';
  if (v < 1) return (v * 60).toFixed(0) + ' 分钟';
  return v.toFixed(1) + ' 小时';
}

function copyTeamReport() {
  if (!teamReportContent.value) return;
  navigator.clipboard.writeText(teamReportContent.value);
  showToast('success', '报告已复制到剪贴板');
}

function downloadTeamReport() {
  if (!teamReportContent.value) return;
  const blob = new Blob([teamReportContent.value], { type: 'text/markdown;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `团队报告_${new Date().toISOString().slice(0, 10)}.md`;
  a.click();
  URL.revokeObjectURL(url);
}

const renderedTeamReport = computed(() => {
  const content = teamReportContent.value;
  if (!content) return '';
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(
      /^## (.+)$/gm,
      '<h3 class="text-base font-semibold mt-4 mb-2 text-slate-800 dark:text-slate-200">$1</h3>'
    )
    .replace(
      /^### (.+)$/gm,
      '<h4 class="text-sm font-semibold mt-3 mb-1 text-slate-700 dark:text-slate-300">$1</h4>'
    )
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold">$1</strong>')
    .replace(
      /^\| (.+) \|$/gm,
      '<div class="text-xs text-slate-600 dark:text-slate-400 my-0.5">$1</div>'
    )
    .replace(/^- (.+)$/gm, '<li class="ml-4 text-sm text-slate-600 dark:text-slate-400">$1</li>')
    .replace(/\n/g, '<br>');
});

const showTemplateModal = ref(false);
const templateLoading = ref(false);
const templateContent = ref('');
const templateSaving = ref(false);

const toastMsg = ref('');
const toastType = ref<'success' | 'error'>('success');
function showToast(type: 'success' | 'error', msg: string) {
  toastMsg.value = msg;
  toastType.value = type;
  setTimeout(() => {
    toastMsg.value = '';
  }, 3000);
}

const periodOptions: { key: Period; label: string; icon: string }[] = [
  { key: 'week', label: '本周', icon: '📅' },
  { key: 'month', label: '本月', icon: '📆' },
  { key: 'year', label: '本年度', icon: '📊' },
];

const periodLabels: Record<string, string> = {
  week: '本周',
  month: '本月',
  year: '本年度',
};

const maxTrendCount = computed(() => {
  if (!stats.value?.daily_trend?.length) return 1;
  return Math.max(...stats.value.daily_trend.map((d) => d.count), 1);
});

async function loadStats() {
  loading.value = true;
  error.value = '';
  try {
    const res = await fetchPersonalStats(activePeriod.value);
    stats.value = res.data as PersonalStatsData;
  } catch {
    error.value = '加载统计数据失败';
    stats.value = null;
  } finally {
    loading.value = false;
  }
}

function switchPeriod(period: Period) {
  activePeriod.value = period;
  reportGenerated.value = false;
  reportContent.value = '';
  loadStats();
}

async function handleGenerateReport() {
  reportLoading.value = true;
  reportError.value = '';
  reportContent.value = '';
  try {
    const res = await generateAIReport(activePeriod.value);
    const data = res.data as { report: string; report_type: string };
    reportContent.value = data.report || '';
    reportType.value = data.report_type || '';
    reportGenerated.value = true;
    showToast('success', '报告生成成功，已保存到历史记录');
    loadReports();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'AI报告生成失败';
    reportError.value = msg;
    showToast('error', msg);
  } finally {
    reportLoading.value = false;
  }
}

async function loadReports() {
  reportsLoading.value = true;
  try {
    const res = await listReports({
      page: reportsPage.value,
      page_size: reportsPageSize,
      keyword: reportsFilter.value.keyword || undefined,
      period: reportsFilter.value.period || undefined,
      date_from: reportsFilter.value.date_from || undefined,
      date_to: reportsFilter.value.date_to || undefined,
    });
    const data = res.data as unknown as { data: WorkReportItem[]; total: number };
    reports.value = data.data || [];
    reportsTotal.value = data.total || 0;
  } catch {
    showToast('error', '加载报告列表失败');
  } finally {
    reportsLoading.value = false;
  }
}

function applyReportFilters() {
  reportsPage.value = 1;
  loadReports();
}

function resetReportFilters() {
  reportsFilter.value = { keyword: '', period: '', date_from: '', date_to: '' };
  reportsPage.value = 1;
  loadReports();
}

function viewReport(report: WorkReportItem) {
  // 需求25：点击报告跳转到详情子页面
  router.push(`/analytics/reports/${report.id}`);
}

async function openTemplateEditor() {
  showTemplateModal.value = true;
  templateLoading.value = true;
  try {
    const res = await fetchReportTemplate();
    templateContent.value = (res.data as ReportTemplateData).content || '';
  } catch {
    showToast('error', '加载模板失败');
  } finally {
    templateLoading.value = false;
  }
}

async function handleSaveTemplate() {
  if (!templateContent.value.trim()) {
    showToast('error', '模板内容不能为空');
    return;
  }
  templateSaving.value = true;
  try {
    await saveReportTemplate(templateContent.value);
    showToast('success', '模板保存成功');
    showTemplateModal.value = false;
  } catch {
    showToast('error', '保存模板失败');
  } finally {
    templateSaving.value = false;
  }
}

async function handleDeleteReport(id: string) {
  try {
    await deleteReport(id);
    showToast('success', '报告已删除');
    loadReports();
  } catch {
    showToast('error', '删除报告失败');
  }
}

function formatPercent(v: number): string {
  return v.toFixed(1) + '%';
}

function formatHours(v: number): string {
  if (v < 1) return (v * 60).toFixed(0) + ' 分钟';
  return v.toFixed(1) + ' 小时';
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
}

function copyReport() {
  if (!reportContent.value) return;
  navigator.clipboard.writeText(reportContent.value);
  showToast('success', '报告已复制到剪贴板');
}

const renderedReport = computed(() => {
  const content = reportContent.value;
  if (!content) return '';
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(
      /^## (.+)$/gm,
      '<h3 class="text-base font-semibold mt-4 mb-2 text-slate-800 dark:text-slate-200">$1</h3>'
    )
    .replace(
      /^### (.+)$/gm,
      '<h4 class="text-sm font-semibold mt-3 mb-1 text-slate-700 dark:text-slate-300">$1</h4>'
    )
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold">$1</strong>')
    .replace(/^- (.+)$/gm, '<li class="ml-4 text-sm text-slate-600 dark:text-slate-400">$1</li>')
    .replace(/\n/g, '<br>');
});

function downloadReport() {
  if (!reportContent.value) return;
  const blob = new Blob([reportContent.value], { type: 'text/markdown;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `工作成效报告_${new Date().toISOString().slice(0, 10)}.md`;
  a.click();
  URL.revokeObjectURL(url);
}

onMounted(() => {
  loadStats();
  loadReports();
  loadTeamStats();
  loadAIModels();
});
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-slate-900 transition-colors duration-300">
    <!-- Header -->
    <div class="shrink-0 px-6 py-4 border-b border-slate-200 dark:border-slate-700">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">工作成效分析</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
        个人工作数据统计、趋势分析与AI智能报告
      </p>
    </div>

    <!-- Sub tabs（滑动指示条） -->
    <div class="shrink-0 px-6 pt-4">
      <div ref="tabBarRef" class="relative flex gap-2">
        <div
          class="absolute top-0 h-full rounded-lg bg-blue-50 dark:bg-blue-900/40 shadow-sm transition-[left,width] duration-300 ease-out pointer-events-none"
          :style="indicatorStyle"
          aria-hidden="true"
        ></div>
        <button
          v-for="tab in [
            { key: 'stats', label: '📊 数据统计' },
            { key: 'team', label: '👥 团队报告' },
            { key: 'history', label: '📋 报告历史' },
          ]"
          :key="tab.key"
          :data-active="viewTab === tab.key"
          :class="[
            'relative z-10 px-4 py-2 rounded-lg text-sm font-medium transition-smooth',
            viewTab === tab.key
              ? 'text-blue-600 dark:text-blue-400'
              : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800',
          ]"
          @click="viewTab = tab.key as 'stats' | 'history' | 'team'"
        >
          {{ tab.label }}
        </button>
        <button
          class="px-3 py-2 rounded-lg text-sm font-medium text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth ml-2 border border-dashed border-slate-300 dark:border-slate-600"
          title="编辑报告模板"
          @click="openTemplateEditor()"
        >
          📝 编辑模板
        </button>
      </div>
    </div>

    <!-- Toast -->
    <div
      v-if="toastMsg"
      :class="[
        'fixed top-4 right-4 z-50 px-4 py-3 rounded-lg text-sm shadow-lg transition-all',
        toastType === 'success' ? 'bg-green-600 text-white' : 'bg-red-600 text-white',
      ]"
    >
      {{ toastMsg }}
    </div>

    <div class="flex-1 overflow-auto p-6">
      <!-- Tab 内容切换：淡入淡出 + 轻微上移 -->
      <transition name="tab-swap" mode="out-in">
        <div :key="viewTab">
          <!-- ===================== 数据统计 Tab ===================== -->
          <template v-if="viewTab === 'stats'">
            <!-- Period Tabs -->
            <div class="flex gap-2 mb-4">
              <button
                v-for="opt in periodOptions"
                :key="opt.key"
                :class="[
                  'flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-smooth',
                  activePeriod === opt.key
                    ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 shadow-sm'
                    : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800',
                ]"
                @click="switchPeriod(opt.key)"
              >
                <span>{{ opt.icon }}</span>
                <span>{{ opt.label }}</span>
              </button>
            </div>

            <div
              v-if="error"
              class="mb-4 px-4 py-3 rounded-lg text-sm bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800"
            >
              {{ error }}
            </div>

            <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-6">
              <div
                v-for="i in 5"
                :key="i"
                class="p-4 rounded-xl border border-slate-100 dark:border-slate-800"
              >
                <div class="skeleton h-3 w-16 mb-3"></div>
                <div class="skeleton h-7 w-20"></div>
              </div>
            </div>

            <template v-else-if="stats">
              <div class="grid grid-cols-2 lg:grid-cols-5 gap-3 mb-6">
                <div
                  class="p-4 rounded-xl bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800"
                >
                  <p
                    class="text-[11px] font-medium text-blue-500 dark:text-blue-400 uppercase tracking-wide mb-1"
                  >
                    创建任务
                  </p>
                  <p class="text-2xl font-bold text-blue-700 dark:text-blue-300">
                    <AnimatedNumber :value="stats.total_created" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800"
                >
                  <p
                    class="text-[11px] font-medium text-green-500 dark:text-green-400 uppercase tracking-wide mb-1"
                  >
                    已完成
                  </p>
                  <p class="text-2xl font-bold text-green-700 dark:text-green-300">
                    <AnimatedNumber :value="stats.total_completed" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800"
                >
                  <p
                    class="text-[11px] font-medium text-amber-500 dark:text-amber-400 uppercase tracking-wide mb-1"
                  >
                    完成率
                  </p>
                  <p class="text-2xl font-bold text-amber-700 dark:text-amber-300">
                    <AnimatedNumber :value="stats.completion_rate" :format="formatPercent" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800"
                >
                  <p
                    class="text-[11px] font-medium text-red-500 dark:text-red-400 uppercase tracking-wide mb-1"
                  >
                    被盯办
                  </p>
                  <p class="text-2xl font-bold text-red-700 dark:text-red-300">
                    <AnimatedNumber :value="stats.remind_received" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-purple-50 dark:bg-purple-900/20 border border-purple-100 dark:border-purple-800"
                >
                  <p
                    class="text-[11px] font-medium text-purple-500 dark:text-purple-400 uppercase tracking-wide mb-1"
                  >
                    平均耗时
                  </p>
                  <p class="text-2xl font-bold text-purple-700 dark:text-purple-300">
                    <AnimatedNumber :value="stats.avg_completion_hours" :format="formatHours" />
                  </p>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
                <div
                  class="p-5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
                >
                  <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
                    每日任务趋势
                  </h3>
                  <div
                    v-if="stats.daily_trend.length === 0"
                    class="text-center py-8 text-slate-400 text-sm"
                  >
                    暂无数据
                  </div>
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
                      <span
                        class="text-[9px] text-slate-400 dark:text-slate-500 truncate w-full text-center"
                        >{{ d.date.slice(5) }}</span
                      >
                      <div
                        class="absolute -top-6 opacity-0 group-hover:opacity-100 bg-slate-800 text-white text-[10px] px-1.5 py-0.5 rounded whitespace-nowrap transition-opacity pointer-events-none"
                      >
                        {{ d.date }}：{{ d.count }} 条
                      </div>
                    </div>
                  </div>
                </div>

                <div
                  class="p-5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
                >
                  <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
                    标签使用分布
                  </h3>
                  <div
                    v-if="stats.tag_breakdown.length === 0"
                    class="text-center py-8 text-slate-400 text-sm"
                  >
                    暂无数据
                  </div>
                  <div v-else class="space-y-2.5">
                    <div
                      v-for="t in stats.tag_breakdown"
                      :key="t.tag_name"
                      class="flex items-center gap-3"
                    >
                      <span
                        class="text-xs text-slate-600 dark:text-slate-400 w-20 truncate"
                        :title="t.tag_name"
                        >{{ t.tag_name }}</span
                      >
                      <div
                        class="flex-1 h-5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden"
                      >
                        <div
                          class="h-full bg-gradient-to-r from-blue-400 to-blue-500 rounded-full transition-all"
                          :style="{
                            width:
                              (t.count / Math.max(...stats.tag_breakdown.map((x) => x.count), 1)) *
                                100 +
                              '%',
                          }"
                        ></div>
                      </div>
                      <span
                        class="text-xs font-mono text-slate-500 dark:text-slate-400 w-8 text-right"
                        >{{ t.count }}</span
                      >
                    </div>
                  </div>
                </div>
              </div>

              <div
                class="rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40 p-5"
              >
                <div class="flex items-center justify-between mb-4">
                  <div>
                    <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">
                      🤖 AI 智能报告
                    </h3>
                    <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                      基于当前统计数据自动生成结构化工作报告
                      <span class="text-blue-500">（无AI配置时自动使用模板生成）</span>
                    </p>
                  </div>
                  <div class="flex items-center gap-2">
                    <button
                      v-if="!reportGenerated"
                      class="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 rounded-lg transition-smooth disabled:opacity-50 flex items-center gap-2"
                      :disabled="reportLoading"
                      @click="handleGenerateReport()"
                    >
                      <span
                        v-if="reportLoading"
                        class="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"
                      ></span>
                      <span>{{ reportLoading ? '生成中...' : '生成报告' }}</span>
                    </button>
                    <template v-if="reportGenerated && reportContent">
                      <span
                        v-if="reportType === 'ai'"
                        class="text-[10px] px-1.5 py-0.5 rounded bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400 font-medium"
                        >AI生成</span
                      >
                      <span
                        v-else
                        class="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 dark:bg-amber-900/50 text-amber-600 dark:text-amber-400 font-medium"
                        >模板生成</span
                      >
                      <button
                        class="px-3 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-lg border border-slate-200 dark:border-slate-600 transition-smooth"
                        @click="copyReport()"
                      >
                        📋 复制
                      </button>
                      <button
                        class="px-3 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-lg border border-slate-200 dark:border-slate-600 transition-smooth"
                        @click="downloadReport()"
                      >
                        ⬇️ 下载
                      </button>
                    </template>
                  </div>
                </div>

                <div
                  v-if="reportError"
                  class="px-4 py-3 rounded-lg text-sm bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800 mb-3"
                >
                  {{ reportError }}
                </div>

                <div
                  v-if="!reportGenerated && !reportLoading"
                  class="text-center py-8 text-slate-400 dark:text-slate-500"
                >
                  <p class="text-2xl mb-2">🤖</p>
                  <p class="text-sm">点击"生成报告"按钮</p>
                  <p class="text-xs mt-1">AI将基于当前数据自动分析并生成工作报告</p>
                </div>

                <div
                  v-if="reportContent"
                  class="prose prose-sm dark:prose-invert max-w-none bg-white dark:bg-slate-900 rounded-lg p-5 border border-slate-200 dark:border-slate-700"
                >
                  <div v-html="renderedReport"></div>
                </div>
              </div>
            </template>
          </template>

          <!-- ===================== 团队报告 Tab（★ 新增） ===================== -->
          <template v-if="viewTab === 'team'">
            <!-- 时间范围 + 生成控制 -->
            <div
              class="mb-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
            >
              <div class="flex flex-wrap items-center gap-3">
                <div
                  class="flex items-center gap-1 bg-white dark:bg-slate-700 rounded-lg p-1 border border-slate-200 dark:border-slate-600"
                >
                  <button
                    v-for="opt in [
                      { key: 'week', label: '本周' },
                      { key: 'month', label: '本月' },
                      { key: 'custom', label: '自定义' },
                    ]"
                    :key="opt.key"
                    :class="[
                      'px-3 py-1.5 rounded-md text-sm font-medium transition-smooth',
                      teamPeriod === opt.key
                        ? 'bg-blue-500 text-white shadow-sm'
                        : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-600',
                    ]"
                    @click="switchTeamPeriod(opt.key as TeamPeriod)"
                  >
                    {{ opt.label }}
                  </button>
                </div>
                <div class="flex items-center gap-2">
                  <label class="text-xs text-slate-500">从</label>
                  <input
                    v-model="teamDateFrom"
                    type="date"
                    class="input-field !py-1.5 !text-xs !w-auto"
                  />
                  <label class="text-xs text-slate-500">至</label>
                  <input
                    v-model="teamDateTo"
                    type="date"
                    class="input-field !py-1.5 !text-xs !w-auto"
                  />
                  <button
                    class="text-xs px-3 py-1.5 rounded-btn bg-slate-200 dark:bg-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-slate-500 transition-smooth"
                    @click="applyTeamRange()"
                  >
                    查询
                  </button>
                </div>
              </div>
              <div
                class="flex flex-wrap items-center gap-3 mt-3 pt-3 border-t border-slate-200 dark:border-slate-600/60"
              >
                <label class="text-xs text-slate-500 shrink-0">🤖 生成模型</label>
                <select v-model="selectedModelId" class="input-field !py-1.5 !text-xs !w-auto">
                  <option value="">不使用AI（模板生成）</option>
                  <option v-for="m in aiModels" :key="m.id" :value="m.id">
                    {{ m.provider_name }}{{ m.model_name ? ' · ' + m.model_name : '' }}
                  </option>
                </select>
                <button
                  class="text-sm px-4 py-2 rounded-btn bg-gradient-to-r from-blue-500 to-purple-500 text-white hover:from-blue-600 hover:to-purple-600 transition-smooth disabled:opacity-50 flex items-center gap-1.5"
                  :disabled="teamReportLoading"
                  @click="handleGenerateTeamReport()"
                >
                  <span
                    v-if="teamReportLoading"
                    class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin"
                  ></span>
                  {{
                    teamReportLoading
                      ? '生成中...'
                      : '📄 生成' + (teamPeriod === 'month' ? '月报' : '周报')
                  }}
                </button>
                <span class="text-[11px] text-slate-400">
                  {{ selectedModelId ? '使用所选模型智能生成' : '未选择模型，将使用模板生成' }}
                </span>
              </div>
              <p v-if="teamReportError" class="text-xs text-red-500 mt-2">{{ teamReportError }}</p>
            </div>

            <!-- 成员成效统计表 -->
            <div v-if="teamLoading" class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
              <div
                v-for="i in 4"
                :key="i"
                class="p-4 rounded-xl border border-slate-100 dark:border-slate-800"
              >
                <div class="skeleton h-3 w-16 mb-3"></div>
                <div class="skeleton h-7 w-20"></div>
              </div>
            </div>
            <template v-else-if="teamStats">
              <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
                <div
                  class="p-4 rounded-xl bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-800"
                >
                  <p class="text-[11px] font-medium text-blue-500 dark:text-blue-400 mb-1">
                    团队成员
                  </p>
                  <p class="text-2xl font-bold text-blue-700 dark:text-blue-300">
                    <AnimatedNumber :value="teamSummaryStats?.memberCount ?? 0" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-800"
                >
                  <p class="text-[11px] font-medium text-indigo-500 dark:text-indigo-400 mb-1">
                    创建任务
                  </p>
                  <p class="text-2xl font-bold text-indigo-700 dark:text-indigo-300">
                    <AnimatedNumber :value="teamSummaryStats?.totalCreated ?? 0" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800"
                >
                  <p class="text-[11px] font-medium text-green-500 dark:text-green-400 mb-1">
                    完成任务
                  </p>
                  <p class="text-2xl font-bold text-green-700 dark:text-green-300">
                    <AnimatedNumber :value="teamSummaryStats?.totalCompleted ?? 0" />
                  </p>
                </div>
                <div
                  class="p-4 rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800"
                >
                  <p class="text-[11px] font-medium text-amber-500 dark:text-amber-400 mb-1">
                    整体完成率
                  </p>
                  <p class="text-2xl font-bold text-amber-700 dark:text-amber-300">
                    <AnimatedNumber
                      :value="teamSummaryStats?.completionRate ?? 0"
                      :format="(n: number) => n.toFixed(1) + '%'"
                    />
                  </p>
                </div>
              </div>

              <div
                class="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800/40 overflow-hidden"
              >
                <div
                  class="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-smooth"
                  @click="teamMembersExpanded = !teamMembersExpanded"
                >
                  <span
                    class="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2"
                  >
                    <span
                      class="inline-flex items-center justify-center w-5 h-5 rounded-md bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 transition-transform"
                      :class="teamMembersExpanded ? 'rotate-90' : ''"
                    >
                      ▶
                    </span>
                    👥 成员成效明细
                    <span
                      v-if="!teamMembersExpanded"
                      class="text-[11px] font-normal text-slate-400"
                    >
                      {{
                        selectedUserIds.length > 0
                          ? `（已选 ${selectedUserIds.length} 人组建团队）`
                          : '（全部成员）'
                      }}
                      · 点击展开
                    </span>
                  </span>
                  <div class="flex items-center gap-2">
                    <span class="text-xs text-slate-400">
                      {{ teamStats.date_from }} ~ {{ teamStats.date_to }}
                    </span>
                    <span class="text-[11px] text-blue-500 dark:text-blue-400">
                      {{ teamMembersExpanded ? '收起 ▲' : '展开 ▼' }}
                    </span>
                  </div>
                </div>
                <div
                  v-show="teamMembersExpanded"
                  class="px-4 py-2.5 border-b border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/40 flex flex-wrap items-center gap-2"
                >
                  <span class="text-xs font-medium text-slate-600 dark:text-slate-300">
                    {{
                      selectedUserIds.length === 0
                        ? '👥 当前统计全部成员'
                        : `✅ 已勾选 ${selectedUserIds.length} 人组建团队`
                    }}
                  </span>
                  <div class="flex gap-1.5 ml-auto">
                    <button
                      class="text-[11px] px-2 py-1 rounded-md bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400 hover:bg-blue-200 transition-smooth"
                      @click="selectAllMembers()"
                    >
                      全选
                    </button>
                    <button
                      class="text-[11px] px-2 py-1 rounded-md bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 transition-smooth"
                      @click="clearSelectedMembers()"
                    >
                      清空（全部成员）
                    </button>
                  </div>
                </div>
                <div
                  v-show="teamMembersExpanded"
                  class="table-flow overflow-x-auto"
                  @scroll="onTableFlowScroll"
                >
                  <table class="w-full text-sm">
                    <thead>
                      <tr
                        class="text-left text-xs text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/60"
                      >
                        <th class="px-3 py-2.5 w-10"></th>
                        <th class="px-4 py-2.5 font-medium">成员</th>
                        <th class="px-4 py-2.5 font-medium">部门</th>
                        <th class="px-4 py-2.5 font-medium text-center">创建任务</th>
                        <th class="px-4 py-2.5 font-medium text-center">完成任务</th>
                        <th class="px-4 py-2.5 font-medium text-center">完成率</th>
                        <th class="px-4 py-2.5 font-medium text-center">平均完成耗时</th>
                        <th class="px-4 py-2.5 font-medium text-center">被盯办</th>
                      </tr>
                    </thead>
                    <TransitionGroup
                      tag="tbody"
                      enter-active-class="animate-table-row-enter"
                      enter-from-class="opacity-0"
                      move-class="transition-transform duration-300"
                    >
                      <tr v-if="teamStats.members.length === 0" key="empty-members">
                        <td
                          colspan="8"
                          class="text-center py-12 text-slate-400 dark:text-slate-500"
                        >
                          <div class="animate-breathe inline-block">
                            <p class="text-4xl mb-3">👥</p>
                          </div>
                          <p class="text-sm">暂无成员数据</p>
                          <p class="text-xs mt-1">选择时间范围后，成员工作成效将自动统计在这里</p>
                        </td>
                      </tr>
                      <tr
                        v-for="m in teamStats.members"
                        :key="m.user_id"
                        :class="[
                          'border-b border-slate-100 dark:border-slate-700/60 cursor-pointer',
                          selectedUserIds.includes(m.user_id)
                            ? 'bg-blue-50/60 dark:bg-blue-900/20'
                            : '',
                        ]"
                        @click="toggleTeamMember(m.user_id)"
                      >
                        <td class="px-3 py-2.5 text-center" @click.stop>
                          <FormCheckbox
                            :model-value="selectedUserIds.includes(m.user_id)"
                            @update:model-value="toggleTeamMember(m.user_id)"
                          />
                        </td>
                        <td class="px-4 py-2.5 text-slate-700 dark:text-slate-300 font-medium">
                          {{ m.user_name }}
                          <span class="text-[10px] text-slate-400 ml-1">@{{ m.username }}</span>
                        </td>
                        <td class="px-4 py-2.5 text-slate-500 dark:text-slate-400">
                          {{ m.dept_name || '-' }}
                        </td>
                        <td class="px-4 py-2.5 text-center text-slate-600 dark:text-slate-300">
                          {{ m.total_created }}
                        </td>
                        <td class="px-4 py-2.5 text-center text-slate-600 dark:text-slate-300">
                          {{ m.total_completed }}
                        </td>
                        <td class="px-4 py-2.5 text-center">
                          <span
                            :key="
                              m.total_created === 0
                                ? 'na'
                                : m.completion_rate >= 80
                                  ? 'hi'
                                  : m.completion_rate >= 50
                                    ? 'mid'
                                    : 'low'
                            "
                            class="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
                            :class="
                              m.total_created === 0
                                ? 'bg-slate-100 dark:bg-slate-700 text-slate-500'
                                : m.completion_rate >= 80
                                  ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400 animate-status-pulse-green'
                                  : m.completion_rate >= 50
                                    ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 animate-status-pulse-yellow'
                                    : 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-400 animate-status-pulse-red'
                            "
                          >
                            {{ m.total_created === 0 ? '-' : m.completion_rate.toFixed(1) + '%' }}
                          </span>
                        </td>
                        <td class="px-4 py-2.5 text-center text-slate-600 dark:text-slate-300">
                          {{ formatTeamHours(m.avg_completion_hours) }}
                        </td>
                        <td class="px-4 py-2.5 text-center text-slate-600 dark:text-slate-300">
                          {{ m.remind_received }}
                        </td>
                      </tr>
                    </TransitionGroup>
                  </table>
                </div>
              </div>

              <!-- 团队报告展示 -->
              <div
                v-if="teamReportGenerated && teamReportContent"
                class="mt-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800/40 overflow-hidden"
              >
                <div
                  class="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between"
                >
                  <span
                    class="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2"
                  >
                    📄 {{ teamReportType === 'ai' ? 'AI 智能生成' : '模板生成' }}团队报告
                    <span
                      class="text-[10px] px-1.5 py-0.5 rounded-full"
                      :class="
                        teamReportType === 'ai'
                          ? 'bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400'
                          : 'bg-slate-100 dark:bg-slate-700 text-slate-500'
                      "
                    >
                      {{ teamReportType === 'ai' ? 'AI' : '模板' }}
                    </span>
                  </span>
                  <div class="flex gap-2">
                    <button
                      class="text-xs px-2.5 py-1 rounded-btn bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 transition-smooth"
                      @click="copyTeamReport()"
                    >
                      复制
                    </button>
                    <button
                      class="text-xs px-2.5 py-1 rounded-btn bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400 hover:bg-blue-200 transition-smooth"
                      @click="downloadTeamReport()"
                    >
                      下载
                    </button>
                  </div>
                </div>
                <div class="p-4 max-h-[420px] overflow-y-auto markdown-body">
                  <div v-html="renderedTeamReport"></div>
                </div>
              </div>
            </template>
            <p v-else-if="teamError" class="text-sm text-red-500">{{ teamError }}</p>
          </template>

          <!-- ===================== 报告历史 Tab ===================== -->
          <template v-if="viewTab === 'history'">
            <!-- Report filters -->
            <div
              class="mb-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
            >
              <div class="flex flex-wrap items-center gap-3">
                <input
                  v-model="reportsFilter.keyword"
                  type="text"
                  class="px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400 placeholder-slate-400 w-48"
                  placeholder="🔍 关键词搜索..."
                  @keyup.enter="applyReportFilters()"
                />
                <select
                  v-model="reportsFilter.period"
                  class="px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                  @change="applyReportFilters()"
                >
                  <option value="">全部周期</option>
                  <option value="week">本周</option>
                  <option value="month">本月</option>
                  <option value="year">本年度</option>
                </select>
                <input
                  v-model="reportsFilter.date_from"
                  type="date"
                  class="px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                  @change="applyReportFilters()"
                />
                <span class="text-xs text-slate-400">至</span>
                <input
                  v-model="reportsFilter.date_to"
                  type="date"
                  class="px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                  @change="applyReportFilters()"
                />
                <button
                  class="px-3 py-1.5 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-smooth"
                  @click="resetReportFilters()"
                >
                  🔄 重置
                </button>
              </div>
            </div>

            <div v-if="reportsLoading" class="flex items-center justify-center py-12">
              <div
                class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"
              ></div>
            </div>

            <div
              v-else-if="reports.length === 0"
              class="text-center py-16 text-slate-400 dark:text-slate-500"
            >
              <p class="text-2xl mb-2">📋</p>
              <p class="text-sm">暂无报告记录</p>
              <p class="text-xs mt-1">在"数据统计"中生成报告后，将自动出现在这里</p>
            </div>

            <!-- Report cards -->
            <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="report in reports"
                :key="report.id"
                class="p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:shadow-md transition-smooth cursor-pointer group"
                @click="viewReport(report)"
              >
                <div class="flex items-start justify-between mb-2">
                  <div class="flex-1 min-w-0">
                    <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-200 truncate">
                      {{ report.title }}
                    </h4>
                    <p class="text-[11px] text-slate-400 dark:text-slate-500 mt-0.5">
                      {{ formatTime(report.created_at) }}
                    </p>
                  </div>
                  <div class="flex items-center gap-1 shrink-0 ml-2">
                    <span
                      :class="[
                        'text-[10px] px-1.5 py-0.5 rounded font-medium',
                        report.category === 'team'
                          ? 'bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400'
                          : report.category === 'group'
                            ? 'bg-cyan-100 dark:bg-cyan-900/50 text-cyan-600 dark:text-cyan-400'
                            : 'bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400',
                      ]"
                    >
                      {{
                        report.category === 'team'
                          ? '团队'
                          : report.category === 'group'
                            ? '工作组'
                            : '个人'
                      }}
                    </span>
                    <span
                      :class="[
                        'text-[10px] px-1.5 py-0.5 rounded font-medium',
                        report.report_type === 'ai'
                          ? 'bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400'
                          : 'bg-amber-100 dark:bg-amber-900/50 text-amber-600 dark:text-amber-400',
                      ]"
                      >{{ report.report_type === 'ai' ? 'AI' : '模板' }}</span
                    >
                  </div>
                </div>
                <div
                  class="flex items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400 mb-2"
                >
                  <span>{{ periodLabels[report.period] || report.period_label }}</span>
                  <span>·</span>
                  <span>{{ report.user_name }}</span>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 line-clamp-3">
                  {{ report.content.slice(0, 150) }}...
                </p>
                <div
                  class="flex items-center justify-end gap-2 mt-3 pt-2 border-t border-slate-100 dark:border-slate-700"
                >
                  <button
                    class="text-[11px] text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-smooth opacity-0 group-hover:opacity-100"
                    @click.stop="handleDeleteReport(report.id)"
                  >
                    🗑 删除
                  </button>
                </div>
              </div>
            </div>

            <!-- Pagination -->
            <div
              v-if="reportsTotal > reportsPageSize"
              class="flex items-center justify-between mt-6"
            >
              <span class="text-xs text-slate-400 dark:text-slate-500"
                >共 {{ reportsTotal }} 份报告</span
              >
              <div class="flex items-center gap-2">
                <button
                  class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
                  :disabled="reportsPage <= 1"
                  @click="
                    reportsPage--;
                    loadReports();
                  "
                >
                  上一页
                </button>
                <button
                  class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
                  :disabled="reportsPage * reportsPageSize >= reportsTotal"
                  @click="
                    reportsPage++;
                    loadReports();
                  "
                >
                  下一页
                </button>
              </div>
            </div>
          </template>
        </div>
      </transition>
    </div>

    <!-- Template editor modal -->
    <Teleport to="body">
      <transition name="shrink-out">
        <div
          v-if="showTemplateModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          @click.self="showTemplateModal = false"
        >
          <div
            class="bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[85vh] flex flex-col mx-4 overflow-hidden"
          >
            <div
              class="shrink-0 flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700"
            >
              <div>
                <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">
                  📝 编辑报告模板
                </h3>
                <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                  使用
                  <code v-pre class="text-blue-500 bg-blue-50 dark:bg-blue-900/30 px-1 rounded">{{
                    变量名
                  }}</code>
                  作为占位符，生成报告时自动替换
                </p>
              </div>
              <button
                class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 text-xl leading-none"
                @click="showTemplateModal = false"
              >
                ✕
              </button>
            </div>
            <div
              class="shrink-0 px-6 py-2 border-b border-slate-100 dark:border-slate-800 flex flex-wrap gap-1.5"
            >
              <span class="text-[10px] text-slate-400 dark:text-slate-500">可用变量：</span>
              <code
                v-for="v in [
                  '{{userName}}',
                  '{{periodLabel}}',
                  '{{totalCreated}}',
                  '{{totalCompleted}}',
                  '{{completionRate}}',
                  '{{completionDesc}}',
                  '{{remindDesc}}',
                  '{{remindReceived}}',
                  '{{avgCompletionHours}}',
                  '{{tagList}}',
                  '{{dailyTrend}}',
                  '{{activeTagMsg}}',
                ]"
                :key="v"
                class="text-[10px] bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 px-1.5 py-0.5 rounded cursor-pointer hover:bg-blue-100 dark:hover:bg-blue-800/50"
                @click="templateContent += v"
                >{{ v }}</code
              >
            </div>
            <div v-if="templateLoading" class="flex items-center justify-center py-16">
              <div
                class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"
              ></div>
            </div>
            <div v-else class="flex-1 overflow-hidden p-4">
              <textarea
                v-model="templateContent"
                class="w-full h-full min-h-[300px] p-4 text-sm font-mono border border-slate-200 dark:border-slate-700 rounded-xl bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 resize-none focus:outline-none focus:ring-2 focus:ring-blue-400 placeholder-slate-400"
                placeholder="输入 Markdown 格式的报告模板..."
              ></textarea>
            </div>
            <div
              class="shrink-0 flex items-center justify-between px-6 py-3 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
            >
              <span class="text-[11px] text-slate-400 dark:text-slate-500"
                >💡 模板使用 Markdown 格式，点击上方变量快速插入</span
              >
              <div class="flex items-center gap-2">
                <button
                  class="px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-smooth"
                  @click="showTemplateModal = false"
                >
                  取消
                </button>
                <button
                  class="px-4 py-1.5 text-xs font-medium text-white bg-blue-500 hover:bg-blue-600 rounded-lg transition-smooth disabled:opacity-50"
                  :disabled="templateSaving"
                  @click="handleSaveTemplate()"
                >
                  <span
                    v-if="templateSaving"
                    class="animate-spin rounded-full h-3 w-3 border-2 border-white border-t-transparent inline-block mr-1"
                  ></span>
                  保存
                </button>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>
