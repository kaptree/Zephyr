<template>
  <div
    class="rounded-card border border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-800/50 p-4"
  >
    <!-- 标题行：总数 + 参与人数 -->
    <div class="flex items-center gap-2 mb-3 flex-wrap">
      <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">💬 阶段性反馈</h3>
      <span v-if="feedbacks.length" class="text-xs text-slate-400 dark:text-slate-500">
        ({{ feedbacks.length }}) · {{ participants.length }} 人参与
      </span>
      <span
        v-if="note.is_archived"
        class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-400 dark:bg-slate-700 dark:text-slate-500"
        >已归档 · 只读</span
      >
    </div>

    <!-- 人员筛选 chips（多人参与时展示，按参与次数降序） -->
    <div v-if="participantChips.length > 1" class="flex flex-wrap gap-1.5 mb-4">
      <button
        class="px-2.5 py-0.5 rounded-full text-xs transition-smooth"
        :class="
          !filterUserId
            ? 'bg-blue-500 text-white'
            : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600'
        "
        @click="filterUserId = null"
      >
        全部·{{ feedbacks.length }}
      </button>
      <button
        v-for="p in participantChips"
        :key="p.user_id"
        class="px-2.5 py-0.5 rounded-full text-xs transition-smooth"
        :class="
          filterUserId === p.user_id
            ? 'bg-blue-500 text-white'
            : 'bg-slate-100 text-slate-600 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-300 dark:hover:bg-slate-600'
        "
        :title="p.last_feedback_at ? '最近反馈：' + fmtTime(p.last_feedback_at) : ''"
        @click="filterUserId = filterUserId === p.user_id ? null : p.user_id"
      >
        {{ p.name || '成员' }}·{{ p.feedback_count }}
      </button>
    </div>

    <!-- 输入区：置于时间轴上方，提交后新反馈紧随其下（最新在前），无需下拉；仅交办人/被指派人可见；归档任务隐藏 -->
    <div v-if="canWrite" class="mb-4 pb-4 border-b border-slate-100 dark:border-slate-700">
      <textarea
        v-model="draft"
        rows="2"
        class="input-field resize-none text-sm"
        placeholder="反馈当前进展、遇到的问题或下一步计划…"
        :maxlength="2000"
      ></textarea>
      <div class="flex items-center justify-between mt-2">
        <span class="text-[10px] text-slate-400 dark:text-slate-500"
          >{{ draft.trim().length }}/2000</span
        >
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="text-xs px-3 py-1.5 rounded-btn bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth disabled:opacity-50"
            :disabled="imageUploading"
            title="上传图片"
            @click="insertImage"
          >
            {{ imageUploading ? '上传中...' : '🖼️ 上传图片' }}
          </button>
          <button
            class="btn-primary text-xs !py-1.5 !px-4 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="submitting || !draft.trim()"
            @click="submit"
          >
            {{ submitting ? '提交中...' : '提交反馈' }}
          </button>
        </div>
      </div>
      <p v-if="submitError" class="text-xs text-red-500 dark:text-red-400 mt-1">
        {{ submitError }}
      </p>
    </div>
    <p
      v-else-if="!note.is_archived && myRole === 'cc'"
      class="mb-3 text-[10px] text-slate-400 dark:text-slate-500"
    >
      抄送人仅可查看反馈
    </p>

    <!-- 时间轴列表 -->
    <div v-if="loading" class="py-6 text-center text-xs text-slate-400 dark:text-slate-500">
      加载中...
    </div>
    <div v-else-if="visibleFeedbacks.length === 0" class="py-6 text-center">
      <p class="text-xs text-slate-400 dark:text-slate-500 leading-relaxed">
        暂无阶段性反馈。任务推进中的关键进展、问题与下一步计划都可在此记录，交办人与其他负责人会同步收到提醒。
      </p>
    </div>
    <div v-else class="relative">
      <!-- 滚动锚点：提交后新反馈出现在顶部，滚回列表顶端 -->
      <div ref="timelineTopRef" aria-hidden="true" />
      <div
        class="absolute left-[15px] top-3 bottom-3 w-px bg-slate-200 dark:bg-slate-700"
        aria-hidden="true"
      />
      <div
        v-for="fb in visibleFeedbacks"
        :key="fb.id"
        class="relative flex items-start gap-3 pb-4 last:pb-0"
      >
        <!-- 首字圆形头像（时间轴节点） -->
        <div
          class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-medium shrink-0 z-10"
          :class="avatarColor(fb.user_id)"
        >
          {{ displayName(fb).charAt(0) }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1 flex-wrap">
            <span
              class="text-xs font-medium"
              :class="
                fb.user_id === currentUserId
                  ? 'text-blue-600 dark:text-blue-400'
                  : 'text-slate-900 dark:text-slate-100'
              "
              >{{ displayName(fb) }}</span
            >
            <span
              v-if="roleBadge(fb.user_id)"
              class="text-[10px] px-1.5 py-0.5 rounded-full"
              :class="roleBadgeClass(fb.user_id)"
              >{{ roleBadge(fb.user_id) }}</span
            >
            <span class="text-[10px] text-slate-400 dark:text-slate-500">{{
              fmtTime(fb.created_at)
            }}</span>
            <span
              v-if="fb.id === latestVisibleId"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-blue-50 text-blue-500 dark:bg-blue-900/40 dark:text-blue-300"
              >⟨最新⟩</span
            >
          </div>
          <!-- 内容气泡：本人发言描边强调 -->
          <div
            class="bg-slate-50 dark:bg-slate-900 rounded-lg px-4 py-3"
            :class="fb.user_id === currentUserId ? 'ring-1 ring-blue-200 dark:ring-blue-800' : ''"
          >
            <div
              class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed rich-content-display whitespace-pre-wrap break-words"
              v-html="renderNoteContent(fb.content)"
            ></div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import type { Note, NoteFeedbackItem, FeedbackParticipantStat } from '@/types';
import { listNoteFeedbacks, addNoteFeedback } from '@/services/notes';
import { renderNoteContent } from '@/utils/richText';
import { useToast } from '@/composables/useToast';
import { useImageUpload } from '@/composables/useImageUpload';

const props = defineProps<{
  /** 当前任务（含 creator_id / assigner_id / assignees / ccs，用于角色判定与姓名解析） */
  note: Note;
  /** 当前登录用户 id（角色判定 + 本人发言强调） */
  currentUserId?: string;
}>();

const emit = defineEmits<{
  /** 提交成功后通知父组件（父组件可同步本地任务快照） */
  (e: 'submitted', fb: NoteFeedbackItem): void;
}>();

const { success: toastSuccess } = useToast();

// 图片上传：上传后追加 <img> 到反馈内容（内容以 HTML 渲染，图片可直接显示）
const { uploading: imageUploading, pickImage } = useImageUpload();
function insertImage() {
  pickImage((url) => {
    const img = `<img src="${url}" />`;
    draft.value = draft.value ? `${draft.value}\n${img}` : img;
  });
}

const loading = ref(false);
const feedbacks = ref<NoteFeedbackItem[]>([]);
const participants = ref<FeedbackParticipantStat[]>([]);
const filterUserId = ref<string | null>(null);
const draft = ref('');
const submitting = ref(false);
const submitError = ref('');
const timelineTopRef = ref<HTMLElement | null>(null);

// ===== 加载 =====
async function load() {
  loading.value = true;
  filterUserId.value = null;
  submitError.value = '';
  try {
    const res = await listNoteFeedbacks(props.note.id);
    feedbacks.value = res.data?.feedbacks || [];
    participants.value = res.data?.participants || [];
  } catch {
    feedbacks.value = [];
    participants.value = [];
  } finally {
    loading.value = false;
  }
}
watch(() => props.note.id, load, { immediate: true });

// ===== 展示辅助 =====
function displayName(fb: NoteFeedbackItem): string {
  return fb.user?.name || fb.user_name || '未知用户';
}

/** 用户姓名解析：优先任务关联关系，兜底反馈冗余姓名 */
function nameOfUserId(userId: string): string {
  if (userId === props.note.creator_id) return props.note.creator?.name || '';
  const a = props.note.assignees?.find((x) => (x.user_id || (x as any).id) === userId);
  if (a) return a.user?.name || (a as any).name || '';
  const c = props.note.ccs?.find((x) => x.user_id === userId);
  if (c) return c.user?.name || '';
  const fb = feedbacks.value.find((f) => f.user_id === userId);
  return fb ? displayName(fb) : '成员';
}

/** 角色判定（与后端 FeedbackRoleInNote 对齐）：creator/assigner → assigner 徽章 */
type Role = 'assigner' | 'assignee' | 'cc' | 'unknown';
function roleOf(userId: string): Role {
  if (userId === props.note.creator_id || props.note.assigner_id === userId) return 'assigner';
  if (props.note.assignees?.some((a) => (a.user_id || (a as any).id) === userId)) return 'assignee';
  if (props.note.ccs?.some((c) => c.user_id === userId)) return 'cc';
  return 'unknown';
}
const myRole = computed<Role>(() => (props.currentUserId ? roleOf(props.currentUserId) : 'unknown'));
const canWrite = computed(
  () => !props.note.is_archived && (myRole.value === 'assigner' || myRole.value === 'assignee')
);

/** 角色徽章文案：指派任务创建人显示"交办人"，协作任务显示"发起人" */
function roleBadge(userId: string): string {
  const r = roleOf(userId);
  if (r === 'assigner') return props.note.source_type === 'assigned' ? '交办人' : '发起人';
  if (r === 'assignee') return '被指派人';
  if (r === 'cc') return '抄送';
  return '';
}
function roleBadgeClass(userId: string): string {
  const r = roleOf(userId);
  if (r === 'assigner')
    return 'bg-blue-100 text-blue-700 dark:bg-blue-900/60 dark:text-blue-300';
  if (r === 'assignee')
    return 'bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300';
  return 'bg-purple-100 text-purple-700 dark:bg-purple-900/60 dark:text-purple-300';
}

/** 头像配色：按 user_id 哈希取色，同人同色 */
const AVATAR_COLORS = [
  'bg-blue-500',
  'bg-green-500',
  'bg-violet-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-cyan-500',
];
function avatarColor(userId: string): string {
  let h = 0;
  for (let i = 0; i < userId.length; i++) h = (h * 31 + userId.charCodeAt(i)) % 997;
  return AVATAR_COLORS[h % AVATAR_COLORS.length];
}

/** 时间格式：24h 内相对时间，超 24h 显示 MM-DD HH:mm */
function fmtTime(iso?: string | null): string {
  if (!iso) return '';
  const t = new Date(iso);
  if (Number.isNaN(t.getTime())) return '';
  const diff = Date.now() - t.getTime();
  if (diff < 60_000) return '刚刚';
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
  const p = (n: number) => String(n).padStart(2, '0');
  return `${p(t.getMonth() + 1)}-${p(t.getDate())} ${p(t.getHours())}:${p(t.getMinutes())}`;
}

// ===== 筛选与最新徽章 =====
const visibleFeedbacks = computed(() =>
  filterUserId.value
    ? feedbacks.value.filter((f) => f.user_id === filterUserId.value)
    : feedbacks.value
);
/** 当前筛选视图内第一条（最新发布）标 ⟨最新⟩ */
const latestVisibleId = computed(() => {
  const list = visibleFeedbacks.value;
  return list.length ? list[0].id : '';
});
const participantChips = computed(() =>
  participants.value.map((p) => ({ ...p, name: nameOfUserId(p.user_id) }))
);

// ===== 提交 =====
async function submit() {
  const content = draft.value.trim();
  if (!content || submitting.value) return;
  if (content.length > 2000) {
    submitError.value = '反馈内容不能超过 2000 字';
    return;
  }
  submitting.value = true;
  submitError.value = '';
  try {
    const res = await addNoteFeedback(props.note.id, content);
    const fb = res.data;
    // 若当前正在按他人筛选，重置筛选以免自己的新反馈被隐藏
    if (filterUserId.value && filterUserId.value !== fb.user_id) {
      filterUserId.value = null;
    }
    // 倒序展示：新反馈局部插入顶部（无需刷新页面），同时更新参与人计数
    feedbacks.value.unshift(fb);
    const p = participants.value.find((x) => x.user_id === fb.user_id);
    if (p) {
      p.feedback_count += 1;
      p.last_feedback_at = fb.created_at;
    } else {
      participants.value.push({
        user_id: fb.user_id,
        feedback_count: 1,
        last_feedback_at: fb.created_at,
      });
    }
    draft.value = '';
    toastSuccess('✓ 反馈已提交');
    emit('submitted', fb);
    await nextTick();
    timelineTopRef.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } };
    submitError.value = err?.response?.data?.message || '提交失败，请重试';
  } finally {
    submitting.value = false;
  }
}
</script>
