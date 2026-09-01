<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  getReportDetail,
  deleteReport,
  type WorkReportItem,
  type ReportDetailData,
} from '@/services/analytics';
import { renderNoteContent } from '@/utils/richText';
import { useConfirm } from '@/composables/useConfirm';

// 全局确认对话框（轻量级通知美学）
const { confirm: appConfirm } = useConfirm();

const route = useRoute();
const router = useRouter();

const loading = ref(true);
const error = ref('');
const report = ref<WorkReportItem | null>(null);
const detail = ref<ReportDetailData>({ range: '', members: [] });
const deleting = ref(false);

const PERIOD_LABELS: Record<string, string> = {
  day: '日报',
  week: '周报',
  month: '月报',
  year: '年度报告',
  custom: '自定义',
};

const categoryLabel = computed(() => {
  const c = report.value?.category;
  if (c === 'team') return '团队报告';
  if (c === 'group') return '工作组报告';
  return '个人报告';
});
const categoryColor = computed(() => {
  const c = report.value?.category;
  if (c === 'team') return 'text-purple-500';
  if (c === 'group') return 'text-cyan-500';
  return 'text-blue-500';
});

async function load() {
  const id = route.params.id as string;
  loading.value = true;
  error.value = '';
  try {
    const res = await getReportDetail(id);
    report.value = res.data.report;
    // 兼容旧报告（无 Detail 字段）与 members 为 null 的情况
    detail.value = {
      range: res.data.detail?.range || '',
      members: Array.isArray(res.data.detail?.members) ? res.data.detail.members : [],
    };
  } catch {
    error.value = '报告加载失败';
  } finally {
    loading.value = false;
  }
}

// 轻量 Markdown → HTML（与工作成效页一致的口径）
function renderMarkdown(content: string): string {
  if (!content) return '';
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/^## (.+)$/gm, '<h3 class="report-h2">$1</h3>')
    .replace(/^### (.+)$/gm, '<h4 class="report-h3">$1</h4>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/^\| (.+) \|$/gm, '<div class="report-table-row">$1</div>')
    .replace(/^- (.+)$/gm, '<li class="report-li">$1</li>')
    .replace(/\n/g, '<br>');
}
const renderedContent = computed(() => renderMarkdown(report.value?.content || ''));

const totalTasks = computed(() =>
  detail.value.members.reduce((sum, m) => sum + (m.tasks || []).length, 0)
);

function formatTime(ts?: string): string {
  if (!ts) return '';
  const d = new Date(ts);
  return `${d.toLocaleDateString('zh-CN')} ${d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`;
}

function formatShortTime(ts?: string): string {
  if (!ts) return '';
  const d = new Date(ts);
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
}

// ---------- 导出：Word（HTML .doc）与 PDF（打印） ----------

function buildArticleHTML(): string {
  const r = report.value;
  if (!r) return '';
  const md = renderMarkdown(r.content);
  const membersHtml = (detail.value.members || [])
    .map((m) => {
      const tasksHtml = (m.tasks || []).length
        ? m.tasks
            .map((t) => {
              const status =
                t.status === 'completed'
                  ? '<span style="color:#16a34a">已完成</span>'
                  : '<span style="color:#d97706">进行中</span>';
              const contentHtml = t.content ? `<div style="margin:4px 0;color:#334155">${renderNoteContent(t.content)}</div>` : '';
              const feedbackHtml = t.feedback
                ? `<div style="margin:6px 0;padding:8px 12px;background:#f0fdf4;border-left:3px solid #22c55e;color:#166534;white-space:pre-line">反馈：${t.feedback}</div>`
                : '';
              return `<div style="margin:10px 0;padding:10px 14px;border:1px solid #e2e8f0;border-radius:8px">
                <div style="font-size:13px;color:#94a3b8">${formatShortTime(t.created_at)} · ${status}</div>
                <div style="font-size:14px;font-weight:600;color:#1e293b;margin-top:2px">${t.title || '无标题'}</div>
                ${contentHtml}
                ${feedbackHtml}
              </div>`;
            })
            .join('')
        : '<div style="color:#94a3b8;margin:6px 0">该成员本周期暂无任务记录</div>';
      return `<div style="margin:16px 0;padding:14px 16px;background:#f8fafc;border-radius:10px">
        <div style="font-size:15px;font-weight:700;color:#1e293b;margin-bottom:8px">${m.user_name || m.user_id} <span style="font-weight:400;font-size:12px;color:#94a3b8">（${(m.tasks || []).length} 项任务）</span></div>
        ${tasksHtml}
      </div>`;
    })
    .join('');

  const rangeHtml = detail.value.range ? `<p style="color:#64748b;font-size:13px">统计周期：${detail.value.range}</p>` : '';
  const membersSection =
    (detail.value.members || []).length > 0
      ? `<div style="margin-top:32px;padding-top:20px;border-top:2px solid #6366f1">
          <h2 style="color:#1e293b;border-bottom:2px solid #e2e8f0;padding-bottom:8px">📋 成员任务明细（${totalTasks.value} 项）</h2>
          ${rangeHtml}
          ${membersHtml}
        </div>`
      : '';

  return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>${r.title}</title>
<style>
body { font-family: 'PingFang SC','Microsoft YaHei','Noto Sans SC',sans-serif; max-width: 900px; margin: 0 auto; padding: 40px 24px; color: #1e293b; line-height: 1.7; }
h1 { font-size: 24px; color: #1e293b; }
.meta { color: #94a3b8; font-size: 13px; margin-bottom: 8px; }
hr { border: none; border-top: 1px solid #e2e8f0; margin: 20px 0; }
.report-h2 { font-size: 18px; margin: 20px 0 8px; }
.report-h3 { font-size: 15px; margin: 14px 0 6px; }
.report-li { margin: 2px 0 2px 20px; }
.report-table-row { font-size: 13px; color: #475569; margin: 2px 0; }
</style>
</head>
<body>
<h1>${r.title}</h1>
<p class="meta">生成人：${r.user_name} · ${categoryLabel.value} · ${PERIOD_LABELS[r.period] || r.period_label || r.period} · ${formatTime(r.created_at)}</p>
<hr>
${md}
${membersSection}
</body>
</html>`;
}

function downloadWord() {
  const r = report.value;
  if (!r) return;
  const blob = new Blob(['\ufeff' + buildArticleHTML()], { type: 'application/msword;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${(r.title || '工作报告').replace(/[\\/:*?"<>|]/g, '-')}.doc`;
  a.click();
  URL.revokeObjectURL(url);
}

function downloadPdf() {
  const r = report.value;
  if (!r) return;
  const win = window.open('', '_blank');
  if (!win) return;
  win.document.write(buildArticleHTML());
  win.document.close();
  win.focus();
  setTimeout(() => win.print(), 400);
}

async function handleDelete() {
  const r = report.value;
  if (!r || deleting.value) return;
  if (!(await appConfirm({ message: '确定删除该报告吗？删除后不可恢复。', danger: true, confirmText: '删除' }))) return;
  deleting.value = true;
  try {
    await deleteReport(r.id);
    router.push('/analytics');
  } catch {
    /* ignore */
  } finally {
    deleting.value = false;
  }
}

function goBack() {
  router.push('/analytics');
}

onMounted(load);
</script>

<template>
  <div class="max-w-4xl mx-auto">
    <!-- 头部 -->
    <div
      class="sticky top-0 z-10 bg-white/90 dark:bg-slate-900/90 backdrop-blur rounded-card border border-slate-100 dark:border-slate-700 p-4 mb-4 flex items-center justify-between gap-3 transition-colors duration-300"
    >
      <div class="flex items-center gap-3 min-w-0">
        <button
          class="shrink-0 text-slate-400 hover:text-blue-500 transition-smooth"
          title="返回报告历史"
          @click="goBack"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-slate-900 dark:text-slate-100 truncate">
            {{ report?.title || '工作报告' }}
          </h2>
          <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
            {{ report?.user_name }} · <span :class="categoryColor">{{ categoryLabel }}</span>
            <template v-if="report"> · {{ formatTime(report.created_at) }}</template>
          </p>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button
          class="px-3 py-1.5 text-xs font-medium rounded-lg bg-red-50 dark:bg-red-900/40 text-red-600 dark:text-red-300 hover:bg-red-100 dark:hover:bg-red-900/60 transition-smooth"
          :disabled="deleting"
          @click="handleDelete"
        >
          删除
        </button>
        <button
          class="px-3 py-1.5 text-xs font-medium rounded-lg bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/60 transition-smooth"
          @click="downloadWord"
        >
          📄 导出 Word
        </button>
        <button
          class="px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-50 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/60 transition-smooth"
          @click="downloadPdf"
        >
          🖨 导出 PDF
        </button>
      </div>
    </div>

    <div v-if="loading" class="py-20 text-center text-slate-400">加载中...</div>
    <div v-else-if="error" class="py-20 text-center text-red-500">{{ error }}</div>

    <template v-else-if="report">
      <!-- 报告正文（文章形式） -->
      <article
        class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-8 shadow-sm transition-colors duration-300"
      >
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">{{ report.title }}</h1>
        <div class="flex flex-wrap items-center gap-2 text-xs text-slate-400 dark:text-slate-500 mb-1">
          <span class="px-2 py-0.5 rounded-tag" :class="categoryColor">{{ categoryLabel }}</span>
          <span
            class="px-2 py-0.5 rounded-tag bg-amber-50 dark:bg-amber-900/40 text-amber-600 dark:text-amber-300"
          >
            {{ report.report_type === 'ai' ? 'AI生成' : '模板生成' }}
          </span>
          <span>{{ PERIOD_LABELS[report.period] || report.period_label || report.period }}</span>
          <span>·</span>
          <span>{{ formatTime(report.created_at) }}</span>
        </div>
        <div v-if="detail.range" class="text-xs text-slate-400 dark:text-slate-500 mb-4">
          统计周期：{{ detail.range }}
        </div>
        <hr class="border-slate-100 dark:border-slate-700 my-5" />
        <div class="report-article" v-html="renderedContent"></div>
      </article>

      <!-- 成员任务与反馈明细 -->
      <section
        v-if="detail.members.length"
        class="mt-5 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300"
      >
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-1">
          📋 成员任务明细
          <span class="text-xs font-normal text-slate-400 dark:text-slate-500">（共 {{ totalTasks }} 项，按时间顺序）</span>
        </h3>
        <div v-if="detail.range" class="text-xs text-slate-400 dark:text-slate-500 mb-4">
          统计周期：{{ detail.range }}
        </div>

        <div
          v-for="m in detail.members"
          :key="m.user_id"
          class="mb-5 last:mb-0 rounded-xl border border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/40 p-4"
        >
          <div class="flex items-center gap-2 mb-3">
            <div
              class="w-7 h-7 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-xs font-medium flex items-center justify-center shrink-0"
            >
              {{ (m.user_name || '?').slice(0, 1) }}
            </div>
            <span class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ m.user_name || m.user_id }}</span>
            <span class="text-xs text-slate-400 dark:text-slate-500">（{{ m.tasks.length }} 项任务）</span>
          </div>

          <div v-if="m.tasks.length === 0" class="text-xs text-slate-400 dark:text-slate-500">
            该成员本周期暂无任务记录
          </div>

          <div v-for="t in m.tasks" :key="t.id" class="mb-3 last:mb-0">
            <div
              class="bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-lg p-3.5"
            >
              <div class="flex items-center gap-2 text-[11px] text-slate-400 dark:text-slate-500">
                <span>{{ formatTime(t.created_at) }}</span>
                <span
                  class="px-1.5 py-0.5 rounded-full font-medium"
                  :class="
                    t.status === 'completed'
                      ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300'
                      : 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300'
                  "
                >
                  {{ t.status === 'completed' ? '已完成' : '进行中' }}
                </span>
                <span v-if="t.completed_at" class="text-[10px]">完成于 {{ formatTime(t.completed_at) }}</span>
              </div>
              <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mt-1.5">
                {{ t.title || '无标题' }}
              </h4>
              <div
                v-if="t.content"
                class="text-xs text-slate-500 dark:text-slate-400 mt-1 rich-content-display line-clamp-4"
                v-html="renderNoteContent(t.content)"
              ></div>
              <div
                v-if="t.feedback"
                class="mt-2 rounded-lg bg-green-50 dark:bg-green-900/30 border-l-2 border-green-500 px-3 py-2 text-xs text-green-700 dark:text-green-300 whitespace-pre-line"
              >
                <span class="font-semibold">任务反馈：</span>{{ t.feedback }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        v-else
        class="mt-5 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 text-center text-xs text-slate-400 transition-colors duration-300"
      >
        该报告暂无成员任务明细（新生成的报告将自动附带各成员任务与反馈）
      </section>
    </template>
  </div>
</template>

<style scoped>
.report-article :deep(.report-h2) {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--tw-slate-800);
  margin: 1.25rem 0 0.5rem;
}
.report-article :deep(.report-h3) {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0.875rem 0 0.375rem;
}
.report-article :deep(.report-li) {
  margin: 2px 0 2px 1.25rem;
}
.report-article :deep(.report-table-row) {
  font-size: 0.8125rem;
  color: var(--tw-slate-500);
  margin: 2px 0;
}
</style>
