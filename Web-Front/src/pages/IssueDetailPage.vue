<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { getIssue, addIssueComment, updateIssueStatus, subscribeIssue, unsubscribeIssue } from '@/services/issues';
import type { IssueDetail, IssueCommentItem } from '@/services/issues';
import { renderNoteContent } from '@/utils/richText';
import { useConfirm } from '@/composables/useConfirm';

// 全局确认对话框（轻量级通知美学）
const { confirm: appConfirm } = useConfirm();

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const detail = ref<IssueDetail | null>(null);
const loading = ref(false);
const error = ref('');

const commentText = ref('');
const submitting = ref(false);
const commentError = ref('');
const updatingStatus = ref(false);
// 需求26：订阅状态与人数
const subscribed = ref(false);
const subscriberCount = ref(0);
const subscribing = ref(false);

const typeBadge: Record<string, { cls: string; icon: string; label: string }> = {
  bug: { cls: 'bg-red-50 dark:bg-red-900/40 text-red-600 dark:text-red-300', icon: '🐛', label: 'Bug 缺陷' },
  feature: { cls: 'bg-green-50 dark:bg-green-900/40 text-green-600 dark:text-green-300', icon: '✨', label: '预期功能' },
};

const canManage = computed(() => {
  if (!detail.value) return false;
  const u = auth.user;
  if (!u) return false;
  const isCreator = u.id === detail.value.issue.user_id;
  const isAdmin = u.role === 'super_admin' || u.role === 'dept_admin';
  return isCreator || isAdmin;
});

const isClosed = computed(() => detail.value?.issue.status === 'closed');

function fmtTime(ts: string): string {
  if (!ts) return '';
  return new Date(ts).toLocaleString('zh-CN');
}

async function load() {
  loading.value = true;
  error.value = '';
  try {
    const res = await getIssue(route.params.id as string);
    detail.value = res.data as unknown as IssueDetail;
    subscribed.value = !!detail.value.subscribed;
    subscriberCount.value = detail.value.subscriber_count || 0;
  } catch {
    error.value = '加载问题详情失败';
  } finally {
    loading.value = false;
  }
}

// 需求26：订阅 / 取消订阅
async function handleToggleSubscribe() {
  if (!detail.value || subscribing.value) return;
  subscribing.value = true;
  try {
    if (subscribed.value) {
      const res = await unsubscribeIssue(detail.value.issue.id);
      subscribed.value = false;
      subscriberCount.value = res.data.subscriber_count || 0;
    } else {
      const res = await subscribeIssue(detail.value.issue.id);
      subscribed.value = true;
      subscriberCount.value = res.data.subscriber_count || 0;
    }
  } catch {
    commentError.value = '订阅操作失败，请重试';
  } finally {
    subscribing.value = false;
  }
}

async function handleComment() {
  if (!commentText.value.trim()) {
    commentError.value = '请输入评论内容';
    return;
  }
  if (!detail.value) return;
  submitting.value = true;
  commentError.value = '';
  try {
    await addIssueComment(detail.value.issue.id, commentText.value.trim());
    commentText.value = '';
    await load();
  } catch {
    commentError.value = '评论失败，请重试';
  } finally {
    submitting.value = false;
  }
}

async function handleToggleStatus() {
  if (!detail.value) return;
  if (!(await appConfirm({ message: isClosed.value ? '确定重新打开这个问题吗？' : '确定关闭这个问题吗？关闭后问题标记为已解决。' }))) return;
  updatingStatus.value = true;
  try {
    await updateIssueStatus(detail.value.issue.id, isClosed.value ? 'open' : 'closed');
    await load();
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } };
    alert(err?.response?.data?.message || '操作失败');
  } finally {
    updatingStatus.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div>
    <button
      class="text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 mb-4 transition-smooth"
      @click="router.push('/issues')"
    >
      ← 返回问题列表
    </button>

    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-20 rounded-card" />
      <div class="skeleton h-64 rounded-card" />
    </div>

    <div v-else-if="error" class="text-center py-16 text-sm text-red-400">
      {{ error }}
      <button class="block mx-auto mt-2 text-blue-500 hover:underline" @click="load">重试</button>
    </div>

    <template v-else-if="detail">
      <div class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300">
        <!-- 头部 -->
        <div class="flex items-start justify-between gap-4 mb-4">
          <div class="flex items-start gap-3 min-w-0">
            <span class="text-xl shrink-0 mt-0.5">{{ typeBadge[detail.issue.type]?.icon }}</span>
            <div class="min-w-0">
              <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 leading-snug">
                {{ detail.issue.title }}
              </h2>
              <div class="flex items-center gap-2 mt-1.5 flex-wrap">
                <span class="text-xs text-slate-400 dark:text-slate-500">#{{ detail.issue.issue_no }}</span>
                <span class="text-[10px] px-1.5 py-0.5 rounded font-medium" :class="typeBadge[detail.issue.type]?.cls">
                  {{ typeBadge[detail.issue.type]?.label }}
                </span>
                <span
                  class="text-[10px] px-1.5 py-0.5 rounded font-medium"
                  :class="isClosed ? 'bg-purple-100 dark:bg-purple-900/50 text-purple-700 dark:text-purple-400' : 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-400'"
                >{{ isClosed ? '🟣 已关闭' : '🟢 开放' }}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <!-- 需求26：订阅按钮 -->
            <button
              class="shrink-0 text-xs px-3 py-1.5 rounded-btn transition-smooth"
              :class="subscribed
                ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 hover:bg-amber-200 dark:hover:bg-amber-900'
                : 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 hover:bg-blue-200 dark:hover:bg-blue-900'"
              :disabled="subscribing"
              :title="subscribed ? '取消订阅后，该问题的新评论不再提醒' : '订阅后，该问题的新评论会提醒你'"
              @click="handleToggleSubscribe"
            >
              {{ subscribing ? '处理中...' : subscribed ? '🔕 已订阅' : '🔔 订阅' }}
              <span v-if="subscriberCount > 0" class="ml-1 opacity-80">({{ subscriberCount }})</span>
            </button>
            <button
              v-if="canManage"
              class="shrink-0 text-xs px-3 py-1.5 rounded-btn transition-smooth"
              :class="isClosed
                ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300 hover:bg-green-200 dark:hover:bg-green-900'
                : 'bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-300 hover:bg-purple-200 dark:hover:bg-purple-900'"
              :disabled="updatingStatus"
              @click="handleToggleStatus"
            >
              {{ updatingStatus ? '处理中...' : isClosed ? '重新打开' : '关闭 Issue' }}
            </button>
          </div>
        </div>

        <!-- 创建人信息条 -->
        <div class="flex items-center gap-2 px-3 py-2 bg-slate-50 dark:bg-slate-900 rounded-lg mb-5">
          <div class="w-7 h-7 rounded-full bg-blue-500 flex items-center justify-center text-white text-[10px] font-medium shrink-0">
            {{ (detail.issue.creator?.name || detail.issue.user_name).charAt(0) }}
          </div>
          <div class="text-xs text-slate-600 dark:text-slate-300">
            <span class="font-medium">{{ detail.issue.creator?.name || detail.issue.user_name }}</span>
            <span class="text-slate-400 dark:text-slate-500"> 于 {{ fmtTime(detail.issue.created_at) }} 提交</span>
          </div>
        </div>

        <!-- 正文内容 -->
        <div
          class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed rich-content-display whitespace-pre-wrap break-words"
          v-html="renderNoteContent(detail.issue.content)"
        ></div>
      </div>

      <!-- 评论区 -->
      <div class="mt-5 bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300">
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">
          💬 讨论（{{ detail.comments.length }}）
        </h3>

        <div v-if="detail.comments.length === 0" class="text-center py-8 text-slate-400 dark:text-slate-500 text-sm">
          暂无评论，快来发表第一条反馈
        </div>

        <div v-else class="space-y-4">
          <div v-for="(c, idx) in detail.comments" :key="c.id">
            <div class="flex items-start gap-3">
              <div class="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-medium shrink-0">
                {{ (c.user?.name || c.user_name).charAt(0) }}
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-xs font-medium text-slate-900 dark:text-slate-100">{{ c.user?.name || c.user_name }}</span>
                  <span class="text-[10px] text-slate-400 dark:text-slate-500">{{ fmtTime(c.created_at) }}</span>
                </div>
                <div class="bg-slate-50 dark:bg-slate-900 rounded-lg px-4 py-3">
                  <div class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed rich-content-display whitespace-pre-wrap break-words" v-html="renderNoteContent(c.content)"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 评论输入 -->
        <div class="mt-5 pt-5 border-t border-slate-100 dark:border-slate-700">
          <textarea
            v-model="commentText"
            rows="3"
            class="input-field resize-none"
            placeholder="写下你的反馈或建议..."
            :disabled="isClosed"
          ></textarea>
          <p v-if="commentError" class="text-xs text-red-500 dark:text-red-400 mt-1">{{ commentError }}</p>
          <p v-if="isClosed" class="text-xs text-slate-400 dark:text-slate-500 mt-1">该问题已关闭，如需继续讨论请重新打开。</p>
          <div class="flex justify-end mt-2">
            <button
              class="btn-primary text-xs !py-1.5 !px-4 disabled:opacity-50"
              :disabled="submitting || isClosed"
              @click="handleComment"
            >
              {{ submitting ? '提交中...' : '发表评论' }}
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
