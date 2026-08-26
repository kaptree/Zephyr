<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { listIssues, createIssue, getIssueWatching, watchIssues, unwatchIssues } from '@/services/issues';
import type { IssueItem } from '@/services/issues';

const router = useRouter();

const issues = ref<IssueItem[]>([]);
const loading = ref(false);
const error = ref('');
const page = ref(1);
const total = ref(0);
const pageSize = 20;

const statusFilter = ref('open'); // 默认展示开放的 issue（需求27）
const typeFilter = ref(''); // '' | bug | feature
const keyword = ref('');

// 需求28：全局订阅（收到所有新 issue 通知）
const watching = ref(false);
const watchingLoading = ref(false);
const watchingError = ref('');

const showCreateModal = ref(false);
const formTitle = ref('');
const formContent = ref('');
const formType = ref<'bug' | 'feature'>('bug');
const creating = ref(false);
const createError = ref('');

const typeBadge: Record<string, { cls: string; icon: string; label: string }> = {
  bug: { cls: 'bg-red-50 dark:bg-red-900/40 text-red-600 dark:text-red-300', icon: '🐛', label: 'Bug 缺陷' },
  feature: { cls: 'bg-green-50 dark:bg-green-900/40 text-green-600 dark:text-green-300', icon: '✨', label: '预期功能' },
};

const statusBadge: Record<string, { cls: string; label: string }> = {
  open: { cls: 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-400', label: '🟢 开放' },
  closed: { cls: 'bg-purple-100 dark:bg-purple-900/50 text-purple-700 dark:text-purple-400', label: '🟣 已关闭' },
};

const openCount = computed(() => issues.value.filter((i) => i.status === 'open').length);

function relTime(ts: string): string {
  if (!ts) return '';
  const diff = Date.now() - new Date(ts).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return '刚刚';
  if (m < 60) return `${m} 分钟前`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h} 小时前`;
  const d = Math.floor(h / 24);
  if (d < 30) return `${d} 天前`;
  return new Date(ts).toLocaleDateString('zh-CN');
}

async function load() {
  loading.value = true;
  error.value = '';
  try {
    const res = await listIssues({
      page: page.value,
      page_size: pageSize,
      status: statusFilter.value || undefined,
      type: typeFilter.value || undefined,
      keyword: keyword.value.trim() || undefined,
    });
    const data = res.data as unknown as { data: IssueItem[]; total: number };
    issues.value = data.data || [];
    total.value = data.total || 0;
  } catch {
    error.value = '加载问题列表失败';
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  load();
}

function resetFilters() {
  statusFilter.value = '';
  typeFilter.value = '';
  keyword.value = '';
  page.value = 1;
  load();
}

function openCreate() {
  formTitle.value = '';
  formContent.value = '';
  formType.value = 'bug';
  createError.value = '';
  showCreateModal.value = true;
}

async function handleCreate() {
  if (!formTitle.value.trim()) {
    createError.value = '请输入问题标题';
    return;
  }
  if (!formContent.value.trim()) {
    createError.value = '请输入问题描述';
    return;
  }
  creating.value = true;
  createError.value = '';
  try {
    const res = await createIssue({
      title: formTitle.value.trim(),
      content: formContent.value.trim(),
      type: formType.value,
    });
    const created = res.data as unknown as IssueItem;
    showCreateModal.value = false;
    router.push(`/issues/${created.id}`);
  } catch {
    createError.value = '创建问题失败，请重试';
  } finally {
    creating.value = false;
  }
}

function goDetail(id: string) {
  router.push(`/issues/${id}`);
}

// 需求28：加载全局订阅状态
async function loadWatching() {
  try {
    const res = await getIssueWatching();
    watching.value = !!res.data?.watching;
  } catch {
    watching.value = false;
  }
}

// 需求28：切换全局订阅
async function handleToggleWatching() {
  if (watchingLoading.value) return;
  watchingLoading.value = true;
  watchingError.value = '';
  try {
    if (watching.value) {
      const res = await unwatchIssues();
      watching.value = !!res.data?.watching;
    } else {
      const res = await watchIssues();
      watching.value = !!res.data?.watching;
    }
  } catch {
    watchingError.value = '订阅操作失败，请重试';
  } finally {
    watchingLoading.value = false;
  }
}

function prevPage() {
  if (page.value > 1) {
    page.value--;
    load();
  }
}

function nextPage() {
  if (page.value * pageSize < total.value) {
    page.value++;
    load();
  }
}

onMounted(() => {
  load();
  loadWatching();
});
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-5">
      <div>
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">🐛 Bug 反馈</h2>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
          提交系统缺陷或功能建议，与大家一同跟进解决
        </p>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <!-- 需求28：全局订阅按钮（新建 issue 左侧） -->
        <button
          class="text-sm px-4 py-2 rounded-btn transition-smooth disabled:opacity-50"
          :class="watching
            ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 hover:bg-amber-200 dark:hover:bg-amber-900'
            : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'"
          :disabled="watchingLoading"
          :title="watching ? '已订阅全部 issue，任何新建 issue 都会提醒你' : '订阅后，所有新建 issue 都会提醒你'"
          @click="handleToggleWatching"
        >
          {{ watchingLoading ? '处理中...' : watching ? '🔕 已订阅' : '🔔 订阅' }}
        </button>
        <button class="btn-primary text-sm !py-2" @click="openCreate">新建 Issue</button>
      </div>
    </div>
    <p v-if="watchingError" class="text-xs text-red-500 dark:text-red-400 mb-2">{{ watchingError }}</p>

    <!-- 筛选栏 -->
    <div class="flex flex-wrap items-center gap-3 mb-4">
      <div class="flex bg-slate-100 dark:bg-slate-800 rounded-btn p-0.5">
        <button
          :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-smooth', statusFilter === '' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="statusFilter = ''; applyFilters()"
        >全部</button>
        <button
          :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-smooth', statusFilter === 'open' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="statusFilter = 'open'; applyFilters()"
        >🟢 开放中</button>
        <button
          :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-smooth', statusFilter === 'closed' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="statusFilter = 'closed'; applyFilters()"
        >🟣 已关闭</button>
      </div>

      <div class="flex bg-slate-100 dark:bg-slate-800 rounded-btn p-0.5">
        <button
          :class="['px-3 py-1.5 rounded-md text-sm font-medium transition-smooth', typeFilter === '' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="typeFilter = ''; applyFilters()"
        >全部类型</button>
        <button
          :class="['px-3 py-1.5 rounded-md text-sm font-medium transition-smooth', typeFilter === 'bug' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="typeFilter = 'bug'; applyFilters()"
        >🐛 Bug</button>
        <button
          :class="['px-3 py-1.5 rounded-md text-sm font-medium transition-smooth', typeFilter === 'feature' ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400']"
          @click="typeFilter = 'feature'; applyFilters()"
        >✨ 预期功能</button>
      </div>

      <div class="flex-1 min-w-[200px] flex gap-2">
        <input
          v-model="keyword"
          type="text"
          class="flex-1 px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
          placeholder="搜索标题 / 内容..."
          @keyup.enter="applyFilters()"
        />
        <button class="px-3 py-1.5 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-smooth" @click="resetFilters()">重置</button>
      </div>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
    </div>

    <!-- 空态 -->
    <div v-else-if="issues.length === 0" class="text-center py-16 text-slate-400 dark:text-slate-500">
      <p class="text-3xl mb-3">🐛</p>
      <p class="text-sm">暂无问题反馈</p>
      <p class="text-xs mt-1">点击右上角「新建 Issue」提交问题或建议</p>
    </div>

    <!-- 列表 -->
    <div v-else class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 overflow-hidden">
      <div v-for="(issue, idx) in issues" :key="issue.id">
        <div
          :class="[
            'px-5 py-3.5 flex items-start gap-3 cursor-pointer transition-smooth',
            issue.status === 'closed'
              ? 'bg-slate-100 dark:bg-slate-700/40 hover:bg-slate-200 dark:hover:bg-slate-700/60'
              : 'hover:bg-slate-50 dark:hover:bg-slate-700/40',
          ]"
          @click="goDetail(issue.id)"
        >
          <span class="text-lg leading-6 shrink-0 mt-0.5">{{ typeBadge[issue.type]?.icon || '📌' }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">{{ issue.title }}</span>
              <span class="text-xs text-slate-400 dark:text-slate-500 shrink-0">#{{ issue.issue_no }}</span>
            </div>
            <div class="flex items-center gap-2 mt-1 flex-wrap">
              <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="typeBadge[issue.type]?.cls">
                {{ typeBadge[issue.type]?.label }}
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="statusBadge[issue.status]?.cls">
                {{ statusBadge[issue.status]?.label }}
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded font-medium bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400">
                💬 {{ issue.comment_count }}
              </span>
              <span class="text-xs text-slate-400 dark:text-slate-500">
                {{ issue.creator?.name || issue.user_name }} 于 {{ relTime(issue.created_at) }} 提交
              </span>
            </div>
          </div>
        </div>
        <div v-if="idx < issues.length - 1" class="mx-5 border-b border-slate-100 dark:border-slate-700" />
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > pageSize" class="flex items-center justify-between mt-4">
      <span class="text-xs text-slate-400 dark:text-slate-500">共 {{ total }} 条，第 {{ page }} 页</span>
      <div class="flex items-center gap-2">
        <button class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40" :disabled="page <= 1" @click="prevPage()">上一页</button>
        <button class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40" :disabled="page * pageSize >= total" @click="nextPage()">下一页</button>
      </div>
    </div>

    <!-- 新建 Issue 弹窗 -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showCreateModal = false" />
        <div class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-xl mx-4 p-6 animate-fade-in">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">新建 Issue</h3>
            <button class="p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth" @click="showCreateModal = false">
              <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </button>
          </div>
          <form @submit.prevent="handleCreate" class="space-y-4">
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">问题类型</span>
              <div class="grid grid-cols-2 gap-3">
                <label :class="['flex items-center gap-2 px-4 py-2.5 rounded-btn border-2 cursor-pointer transition-smooth', formType === 'bug' ? 'border-red-500 bg-red-50 dark:bg-red-900/40 text-red-700 dark:text-red-300' : 'border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-400']">
                  <input v-model="formType" type="radio" value="bug" class="sr-only" />
                  <span>🐛 Bug 缺陷</span>
                </label>
                <label :class="['flex items-center gap-2 px-4 py-2.5 rounded-btn border-2 cursor-pointer transition-smooth', formType === 'feature' ? 'border-green-500 bg-green-50 dark:bg-green-900/40 text-green-700 dark:text-green-300' : 'border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-400']">
                  <input v-model="formType" type="radio" value="feature" class="sr-only" />
                  <span>✨ 预期功能</span>
                </label>
              </div>
            </div>
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">标题</span>
              <input v-model="formTitle" class="input-field" placeholder="简要描述问题或建议" autofocus />
            </div>
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">详细描述</span>
              <textarea v-model="formContent" rows="5" class="input-field resize-none" placeholder="请详细描述问题现象、复现步骤，或预期功能的使用场景..."></textarea>
            </div>
            <p v-if="createError" class="text-xs text-red-500 dark:text-red-400">{{ createError }}</p>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" class="btn-secondary text-xs !py-1.5 !px-4" @click="showCreateModal = false">取消</button>
              <button type="submit" class="btn-primary text-xs !py-1.5 !px-4" :disabled="creating">
                {{ creating ? '提交中...' : '提交 Issue' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
