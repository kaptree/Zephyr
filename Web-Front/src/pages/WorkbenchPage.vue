<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useNoteStore } from '@/stores/notes';
import { useAuthStore } from '@/stores/auth';
import { useNotificationStore } from '@/stores/notification';
import type { Note } from '@/types';
import TagSelector from '@/components/common/TagSelector.vue';
import StickyNoteCard from '@/components/note/StickyNoteCard.vue';
import UserPicker from '@/components/common/UserPicker.vue';
import MarkdownEditor from '@/components/common/MarkdownEditor.vue';
import { markdownToHtml, htmlToMarkdown } from '@/utils/markdown';
import FeedbackModal from '@/components/notification/FeedbackModal.vue';
import { createWorkGroup, searchGroups, deleteWorkGroup } from '@/services/workgroup';
import type { WorkGroupData } from '@/services/workgroup';
import { recommendUsers, getWorkTypeOptions } from '@/services/admin';
import type { WorkTypeOption } from '@/types';
import { getPresets } from '@/services/presets';
import type { PresetGroup } from '@/types/preset';
import { fetchTemplates } from '@/services/templates';
import type { Template } from '@/types';

const router = useRouter();
const route = useRoute();
const noteStore = useNoteStore();
const auth = useAuthStore();
const notifStore = useNotificationStore();
const showCreateModal = ref(false);
const showDetailPanel = ref(false);
const selectedNote = ref<Note | null>(null);

const newTitle = ref('');
const newContent = ref('');
const selectedTagIds = ref<string[]>([]);
const sourceType = ref<'self' | 'assigned' | 'collaboration'>('self');
const selectedAssigneeIds = ref<string[]>([]);
const selectedCcIds = ref<string[]>([]);
const creating = ref(false);
const createError = ref('');

// 工作时间选项（秒）：指派任务时限，后端据此自动计算截止时间
const workTimeOptions = [
  { label: '不限制', value: 0 },
  { label: '1 小时', value: 3600 },
  { label: '2 小时', value: 7200 },
  { label: '4 小时', value: 14400 },
  { label: '8 小时（1天班）', value: 28800 },
  { label: '1 天', value: 86400 },
  { label: '2 天', value: 172800 },
  { label: '3 天', value: 259200 },
  { label: '7 天', value: 604800 },
];
const workTimeSeconds = ref(0);

const editingTitle = ref('');
const editingContent = ref('');
const selectedEditingTagIds = ref<string[]>([]);
const saving = ref(false);
const completing = ref(false);
const tagSaving = ref(false);
const tagError = ref('');
const feedbackVisible = ref(false);
const feedbackNote = ref<Note | null>(null);

const activeTab = ref('all');

const showWorkGroupModal = ref(false);
const workGroups = ref<WorkGroupData[]>([]);
const wgLoading = ref(false);
const groupsTotal = ref(0);
const groupsPage = ref(1);
const groupsPageSize = 20;
const groupsFilter = ref({ keyword: '', date_from: '', date_to: '' });
const wgName = ref('');
const wgDescription = ref('');
const wgTemplate = ref('default');
const wgDueDate = ref('');
const wgSubGroups = ref<
  { name: string; members: { user_id: string; role: string; sub_group_name: string }[] }[]
>([{ name: '', members: [] }]);
const wgCreating = ref(false);
const wgError = ref('');

const selectedWorkType = ref('');
const workTypeOptions = ref<WorkTypeOption[]>([]);
const recommending = ref(false);
const recommendError = ref('');
const recommendResult = ref<
  Array<{
    id: string;
    name: string;
    dept_name: string;
    work_type_stats?: { work_type: string; group_count: number }[];
  }>
>([]);
const selectedWGUserIds = ref<string[][]>([[]]);

const availablePresets = ref<PresetGroup[]>([]);
const selectedPresetId = ref('');

const userTemplates = ref<Template[]>([]);
const selectedTemplateId = ref('');

const displayedNotes = computed(() => {
  if (activeTab.value === 'red')
    return noteStore.activeNotes.filter((n) => n.color_status === 'red');
  if (activeTab.value === 'blue')
    return noteStore.activeNotes.filter((n) => n.color_status === 'blue');
  return noteStore.activeNotes;
});

// 详情中的指派任务是否为「发起者视角」（创建人 = 当前账号）
const selectedIsAssigner = computed(
  () =>
    selectedNote.value?.source_type === 'assigned' &&
    !!auth.user &&
    selectedNote.value.creator_id === auth.user.id
);

// 签收体系（需求19）：当前账号在详情任务中的被指派人行，及是否可签收
const selectedAssigneeMe = computed(() =>
  (selectedNote.value?.assignees || []).find(
    (a: any) => (a.user_id || (a as any).id) === auth.user?.id
  )
);
const canSignSelected = computed(
  () =>
    !!selectedNote.value &&
    selectedNote.value.source_type === 'assigned' &&
    !!selectedAssigneeMe.value &&
    selectedAssigneeMe.value.role_in_note !== 'initiator' &&
    selectedAssigneeMe.value.sign_status !== 'signed'
);

// 抄送视角（需求20）：当前账号仅被抄送（非被指派人/发起者）→ 详情只读
const selectedIsCcOnly = computed(
  () =>
    !!selectedNote.value &&
    (selectedNote.value.ccs || []).some((c: any) => (c.user_id || c.id) === auth.user?.id) &&
    !(selectedNote.value.assignees || []).some((a: any) => (a.user_id || a.id) === auth.user?.id)
);

// 需求23：详情中指派任务的被指派人本人完成进度（全部完成后发起者才可归档）
const selectedAssigneeMembers = computed(() =>
  (selectedNote.value?.assignees || []).filter((a: any) => a.role_in_note !== 'initiator')
);
const selectedCompletedCount = computed(
  () => selectedAssigneeMembers.value.filter((a: any) => a.is_completed).length
);
const selectedAllCompleted = computed(
  () =>
    selectedAssigneeMembers.value.length > 0 &&
    selectedCompletedCount.value === selectedAssigneeMembers.value.length
);
const selectedViewerIsMemberAssignee = computed(
  () =>
    !!selectedNote.value &&
    selectedNote.value.source_type === 'assigned' &&
    !!selectedAssigneeMe.value &&
    selectedAssigneeMe.value.role_in_note !== 'initiator'
);
const selectedMyCompleted = computed(() => !!selectedAssigneeMe.value?.is_completed);
const selectedIsAdminViewer = computed(
  () => auth.user?.role === 'super_admin' || auth.user?.role === 'dept_admin'
);
const selectedCompleteLabel = computed(() => {
  const note = selectedNote.value;
  if (!note) return '完成并归档';
  if (note.source_type === 'assigned') {
    if (selectedViewerIsMemberAssignee.value)
      return selectedMyCompleted.value ? '已完成' : '提交完成';
    if (selectedIsAssigner.value || selectedIsAdminViewer.value) {
      return selectedAllCompleted.value
        ? '归档任务'
        : `归档（${selectedCompletedCount.value}/${selectedAssigneeMembers.value.length}）`;
    }
    return ''; // 抄送等无完成权限：隐藏按钮
  }
  return '完成并归档';
});
const selectedCompleteDisabled = computed(() => {
  const note = selectedNote.value;
  if (!note || note.source_type !== 'assigned') return false;
  if (selectedViewerIsMemberAssignee.value) return false; // 已完成仍可再次提交反馈
  if (selectedIsAssigner.value || selectedIsAdminViewer.value) return !selectedAllCompleted.value;
  return true;
});

onMounted(() => {
  noteStore.fetchNotes();
  loadWorkGroups();
  loadWorkTypeOptions();
  loadPresets();
  loadUserTemplates();
  // 支持从通知中心跳转到指定任务（首次挂载时）
  const noteId = route.query.note as string | undefined;
  if (noteId) openNoteFromQuery(noteId);
});

// 通知中心「查看任务」跳转：监听 query.note 变化（同页面再次点击也能弹出便签）
watch(
  () => route.query.note,
  (noteId) => {
    if (noteId && typeof noteId === 'string') openNoteFromQuery(noteId);
  }
);

// 需求30：别人指派/抄送的任务 → 工作台动态刷新（无需手动刷新页面）
watch(
  () => notifStore.noteUpdateTick,
  () => {
    const last = notifStore.lastNoteUpdate;
    // 若新任务已在列表（刚通过其他途径加载过），仅刷新不重复弹出
    if (last?.note_id && noteStore.activeNotes.some((n) => n.id === last.note_id)) return;
    noteStore.fetchNotes();
  }
);

function openNoteFromQuery(noteId: string) {
  const target = noteStore.activeNotes.find((n) => n.id === noteId);
  if (target) {
    openDetail(target);
    return;
  }
  import('@/services/notes').then(({ fetchNoteById }) => {
    fetchNoteById(noteId)
      .then((res) => {
        openDetail(normalizeRawNote(res.data as unknown as Record<string, unknown>));
      })
      .catch(() => {
        /* 任务不存在 */
      });
  });
}

// 后端 Note 原始结构归一化为前端 Note
function normalizeRawNote(raw: Record<string, unknown>): Note {
  return {
    id: (raw.id as string) || '',
    title: (raw.title as string) || '',
    content: (raw.content as string) || '',
    color_status: (raw.color_status as Note['color_status']) || 'yellow',
    source_type: (raw.source_type as Note['source_type']) || 'self',
    owner_id: (raw.owner_id as string) || '',
    creator_id: (raw.creator_id as string) || '',
    creator: raw.creator as Note['creator'],
    owner: raw.owner as Note['owner'],
    is_archived: !!raw.is_archived,
    tags: (raw.tags || []) as Note['tags'],
    assignees: (raw.assignees || []) as Note['assignees'],
    ccs: raw.ccs as Note['ccs'],
    group_id: raw.group_id as string | undefined,
    dept_id: raw.dept_id as string | undefined,
    template_type: raw.template_type as string | undefined,
    due_time: raw.due_time as string | undefined,
    completed_at: raw.completed_at as string | undefined,
    archive_time: raw.archive_time as string | undefined,
    remind_count: (raw.remind_count as number) || 0,
    serial_no: raw.serial_no as string | undefined,
    created_at: (raw.created_at as string) || new Date().toISOString(),
    updated_at:
      (raw.updated_at as string) || (raw.created_at as string) || new Date().toISOString(),
  };
}

function handleTabClick(tab: string) {
  activeTab.value = tab;
  if (tab === 'groups') {
    loadWorkGroups();
    return;
  }
  if (tab === 'all') noteStore.fetchNotes({ status: undefined });
  else if (tab === 'red')
    noteStore.fetchNotes({ status: undefined }).then(() => {
      noteStore.activeNotes = noteStore.activeNotes.filter((n) => n.color_status === 'red');
    });
  else if (tab === 'assigned')
    noteStore.fetchNotes({ status: undefined }).then(() => {
      noteStore.activeNotes = noteStore.activeNotes.filter(
        (n) => n.source_type === 'assigned' || n.source_type === 'collaboration'
      );
    });
  else noteStore.fetchNotes({ status: tab });
}

watch(sourceType, (val) => {
  if (val === 'self') selectedAssigneeIds.value = [];
});

function openCreateModal() {
  newTitle.value = '';
  newContent.value = '';
  selectedTagIds.value = [];
  selectedAssigneeIds.value = [];
  selectedCcIds.value = [];
  sourceType.value = 'self';
  workTimeSeconds.value = 0;
  createError.value = '';
  showCreateModal.value = true;
}
function openDetail(note: Note) {
  selectedNote.value = note;
  editingTitle.value = note.title || '';
  editingContent.value = htmlToMarkdown(note.content || '');
  selectedEditingTagIds.value = (note.tags || []).map((t) => t.id);
  tagError.value = '';
  showDetailPanel.value = true;
}
function closeDetail() {
  showDetailPanel.value = false;
  selectedNote.value = null;
  completing.value = false;
}

async function handleSubmit() {
  if (creating.value) return;
  if (!newTitle.value.trim()) {
    createError.value = '请输入任务标题';
    return;
  }
  if (sourceType.value !== 'self' && selectedAssigneeIds.value.length === 0) {
    createError.value = '请选择指派人员';
    return;
  }
  creating.value = true;
  createError.value = '';
  try {
    const payload: any = {
      title: newTitle.value.trim(),
      content: markdownToHtml(newContent.value),
      tags: selectedTagIds.value,
      source_type: sourceType.value,
    };
    // 抄送人（需求20）：无论是否指派都可多选抄送
    if (selectedCcIds.value.length > 0) payload.cc_user_ids = selectedCcIds.value;
    if (sourceType.value !== 'self' && workTimeSeconds.value > 0)
      payload.work_time_seconds = workTimeSeconds.value;
    if (sourceType.value !== 'self')
      payload.assignees = selectedAssigneeIds.value.map((id) => ({ user_id: id }));
    if (sourceType.value === 'assigned' && selectedAssigneeIds.value.length > 0)
      payload.owner_id = selectedAssigneeIds.value[0];
    const created = await noteStore.createNote(payload);
    // 指派/协作任务：向所有被指派人员发送盯办提醒（排除发起人自己）
    if (sourceType.value !== 'self' && created) {
      const myId = auth.user?.id;
      const targets = (created.assignees || []).map(assigneeId).filter((id) => id && id !== myId);
      for (const tid of targets) {
        try {
          await noteStore.remindNote(
            created.id,
            tid,
            `【任务指派】${auth.user?.name || '管理员'} 指派您处理：${newTitle.value.trim()}`
          );
        } catch {
          /* ignore */
        }
      }
    }
    showCreateModal.value = false;
    // 局部刷新任务面板，立即展示最新任务（含倒计时）
    noteStore.fetchNotes();
  } catch (e: unknown) {
    createError.value =
      (e as { response?: { data?: { message?: string } } })?.response?.data?.message ||
      '创建任务失败';
  } finally {
    creating.value = false;
  }
}

async function handleSaveDetail() {
  if (!selectedNote.value) return;
  saving.value = true;
  try {
    await noteStore.updateNoteLocally(selectedNote.value.id, {
      title: editingTitle.value.trim(),
      content: markdownToHtml(editingContent.value),
      tags: selectedEditingTagIds.value,
    } as any);
    closeDetail();
  } catch {
    /* ignore */
  } finally {
    saving.value = false;
  }
}
async function handleUpdateTags(tagIds: string[]) {
  if (!selectedNote.value) return;
  selectedEditingTagIds.value = tagIds;
  tagSaving.value = true;
  tagError.value = '';
  try {
    await noteStore.updateNoteTags(selectedNote.value.id, tagIds);
    selectedNote.value = {
      ...selectedNote.value,
      tags:
        noteStore.activeNotes.find((n) => n.id === selectedNote.value!.id)?.tags ||
        selectedNote.value.tags,
    };
  } catch {
    tagError.value = '标签更新失败，请重试';
    selectedEditingTagIds.value = (selectedNote.value.tags || []).map((t) => t.id);
  } finally {
    tagSaving.value = false;
  }
}
async function handleComplete(note: Note) {
  // 抄送人（仅被抄送、非被指派人）直接完成归档，不弹反馈填报
  const isCcOnlyNote =
    (note.ccs || []).some((c: any) => (c.user_id || c.id) === auth.user?.id) &&
    !(note.assignees || []).some((a: any) => (a.user_id || a.id) === auth.user?.id);
  if (isCcOnlyNote) {
    completing.value = true;
    try {
      await noteStore.completeNote(note.id, {});
      noteStore.fetchNotes();
      if (showDetailPanel.value && selectedNote.value?.id === note.id) closeDetail();
    } finally {
      completing.value = false;
    }
    return;
  }
  // 需求23：指派任务 —— 发起者/管理员直接归档（后端校验所有被指派人已完成）
  const myAssignee = (note.assignees || []).find((a: any) => (a.user_id || a.id) === auth.user?.id);
  const viewerIsMemberAssignee =
    note.source_type === 'assigned' && !!myAssignee && myAssignee.role_in_note !== 'initiator';
  const viewerIsAssigner =
    note.source_type === 'assigned' && !!auth.user && note.creator_id === auth.user.id;
  if (
    note.source_type === 'assigned' &&
    !viewerIsMemberAssignee &&
    (viewerIsAssigner || auth.user?.role === 'super_admin' || auth.user?.role === 'dept_admin')
  ) {
    completing.value = true;
    try {
      await noteStore.completeNote(note.id, {});
      noteStore.fetchNotes();
      if (showDetailPanel.value && selectedNote.value?.id === note.id) closeDetail();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      alert(err?.response?.data?.message || '归档失败');
    } finally {
      completing.value = false;
    }
    return;
  }
  // 被指派人提交完成 / 非指派任务：先弹反馈填报
  feedbackNote.value = note;
  feedbackVisible.value = true;
}
async function submitFeedback(content: string) {
  if (!feedbackNote.value) return;
  const note = feedbackNote.value;
  completing.value = true;
  try {
    await noteStore.completeNote(note.id, { feedback_content: content });
    // 完成任务后局部刷新，任务面板同步最新状态（完成/归档/倒计时移除）
    noteStore.fetchNotes();
    if (showDetailPanel.value && selectedNote.value?.id === note.id) closeDetail();
  } finally {
    completing.value = false;
    feedbackNote.value = null;
  }
}
// 标记为重要：任务卡片变红（color_status='red'）
async function handleImportant(note: Note) {
  try {
    await noteStore.updateNoteLocally(note.id, { color_status: 'red' });
    if (showDetailPanel.value && selectedNote.value?.id === note.id)
      selectedNote.value = { ...selectedNote.value, color_status: 'red' as const };
  } catch {
    /* ignore */
  }
}
// 删除任务：确认后软删除，从工作台移除（可在归档中恢复）
async function handleDelete(note: Note) {
  const title = note.title || '无标题';
  if (!window.confirm(`确定删除任务「${title}」吗？删除后将从工作台移除，可在归档中恢复。`)) return;
  try {
    await noteStore.archiveNote(note.id);
    noteStore.fetchNotes();
    if (showDetailPanel.value && selectedNote.value?.id === note.id) closeDetail();
  } catch {
    /* ignore */
  }
}

// 签收任务：被指派人签收，签收后卡片右上角显示「已签收」
async function handleSign() {
  if (!selectedNote.value) return;
  try {
    const updated = await noteStore.signNote(selectedNote.value.id);
    if (updated) selectedNote.value = updated;
    noteStore.fetchNotes();
  } catch {
    /* ignore */
  }
}

async function loadWorkGroups() {
  wgLoading.value = true;
  try {
    const res = await searchGroups({
      page: groupsPage.value,
      page_size: groupsPageSize,
      keyword: groupsFilter.value.keyword || undefined,
      date_from: groupsFilter.value.date_from || undefined,
      date_to: groupsFilter.value.date_to || undefined,
    });
    const d = res.data as unknown as { data: WorkGroupData[]; total: number };
    workGroups.value = d.data || [];
    groupsTotal.value = d.total || 0;
  } catch {
    /* ignore */
  } finally {
    wgLoading.value = false;
  }
}
function applyGroupsFilter() {
  groupsPage.value = 1;
  loadWorkGroups();
}
function resetGroupsFilter() {
  groupsFilter.value = { keyword: '', date_from: '', date_to: '' };
  groupsPage.value = 1;
  loadWorkGroups();
}
function goToGroup(id: string) {
  router.push(`/workbench/groups/${id}`);
}

function openWGModal() {
  wgName.value = '';
  wgDescription.value = '';
  wgTemplate.value = 'default';
  wgDueDate.value = '';
  wgSubGroups.value = [{ name: '', members: [] }];
  selectedWGUserIds.value = [[]];
  wgError.value = '';
  showWorkGroupModal.value = true;
}
function addSubGroup() {
  wgSubGroups.value.push({ name: '', members: [] });
  selectedWGUserIds.value.push([]);
}
function removeSubGroup(idx: number) {
  wgSubGroups.value.splice(idx, 1);
  selectedWGUserIds.value.splice(idx, 1);
}
function onWGUserSelect(idx: number, userIds: string[]) {
  selectedWGUserIds.value[idx] = userIds;
  wgSubGroups.value[idx].members = userIds.map((uid) => ({
    user_id: uid,
    role: idx === 0 ? 'leader' : 'member',
    sub_group_name: wgSubGroups.value[idx].name || `小组${idx + 1}`,
  }));
}
async function handleCreateWorkGroup() {
  if (wgCreating.value) return;
  if (!wgName.value.trim()) {
    wgError.value = '请输入工作组名称';
    return;
  }
  const allMembers = wgSubGroups.value.flatMap(
    (sg, idx) =>
      selectedWGUserIds.value[idx]?.map((uid) => ({
        user_id: uid,
        role: idx === 0 ? 'leader' : 'member',
        sub_group_name: sg.name || `小组${idx + 1}`,
      })) || []
  );
  if (allMembers.length === 0) {
    wgError.value = '请至少选择一个成员';
    return;
  }
  wgCreating.value = true;
  wgError.value = '';
  try {
    await createWorkGroup({
      name: wgName.value.trim(),
      description: wgDescription.value,
      template_type: wgTemplate.value,
      due_time: wgDueDate.value ? new Date(wgDueDate.value).toISOString() : undefined,
      preset_id: selectedPresetId.value || undefined,
      members: allMembers,
    });
    showWorkGroupModal.value = false;
    await Promise.all([noteStore.fetchNotes(), loadWorkGroups()]);
  } catch (e: unknown) {
    wgError.value =
      (e as { response?: { data?: { message?: string } } })?.response?.data?.message ||
      '创建工作组失败';
  } finally {
    wgCreating.value = false;
  }
}
async function handleDeleteGroup(id: string) {
  try {
    await deleteWorkGroup(id);
    await loadWorkGroups();
  } catch {
    /* ignore */
  }
}

async function loadWorkTypeOptions() {
  try {
    const res = await getWorkTypeOptions();
    workTypeOptions.value = (res.data as unknown as WorkTypeOption[]) || [];
  } catch {
    workTypeOptions.value = [];
  }
}

async function loadPresets() {
  try {
    const res = await getPresets();
    availablePresets.value = (res.data as unknown as PresetGroup[]) || [];
  } catch {
    availablePresets.value = [];
  }
}

async function loadUserTemplates() {
  try {
    const res = await fetchTemplates();
    userTemplates.value = res.data || [];
  } catch {
    userTemplates.value = [];
  }
}

function onTemplateSelect() {
  if (!selectedTemplateId.value) return;
  const tpl = userTemplates.value.find((t) => t.id === selectedTemplateId.value);
  if (!tpl) return;
  // 需求23/34：模板内容为 Markdown/纯文本模板，直接填入任务 Markdown 编辑器（旧 HTML 模板先转回 Markdown）
  let content = (tpl.content || '').trim();
  if (!content) {
    // 兼容旧 JSON 字段模板：把字段名转为占位正文
    try {
      const fields = JSON.parse(
        typeof tpl.fields === 'string' ? tpl.fields : JSON.stringify(tpl.fields ?? [])
      );
      if (Array.isArray(fields) && fields.length) {
        content = fields.map((f: any) => `【${f.name}】`).join('\n');
      }
    } catch {}
  }
  if (content) {
    content = htmlToMarkdown(content);
    const block = `> 📋 模板：${tpl.name}\n\n${content}`;
    newContent.value = newContent.value.trim() ? `${block}\n\n${newContent.value}` : block;
  }
}

function handlePresetSelect() {
  if (!selectedPresetId.value) return;
  const preset = availablePresets.value.find((p) => p.id === selectedPresetId.value);
  if (!preset || !preset.members) return;
  selectedAssigneeIds.value = preset.members.map((m) => m.user_id);
}

async function handleRecommend() {
  if (!selectedWorkType.value) return;
  recommending.value = true;
  recommendError.value = '';
  recommendResult.value = [];
  try {
    const res = await recommendUsers({ work_type: selectedWorkType.value, limit: 10 });
    const users = (res.data as unknown as any[]) || [];
    recommendResult.value = users.map((u: any) => ({
      id: u.id,
      name: u.name,
      dept_name: u.department?.name || u.dept_name || '',
      work_type_stats: [],
    }));
    if (users.length === 0) {
      recommendError.value = '暂无匹配的推荐人员';
    }
  } catch {
    recommendError.value = '推荐失败，请重试';
  } finally {
    recommending.value = false;
  }
}

function selectRecommendUser(userId: string) {
  if (!selectedAssigneeIds.value.includes(userId)) {
    selectedAssigneeIds.value.push(userId);
  }
}

/** 从指派对象中取用户 ID（兼容 user_id / id / 嵌套 user 三种形态） */
function assigneeId(a: any): string {
  return a?.id || a?.user_id || a?.user?.id || '';
}
/** 从指派对象中取用户姓名（兼容扁平与嵌套形态） */
function assigneeName(a: any): string {
  return a?.name || a?.user?.name || a?.user_id || '';
}

function getMemberCount(g: WorkGroupData): number {
  return g.members?.length || 0;
}
function getMemberNames(g: WorkGroupData): string {
  return (
    g.members
      ?.slice(0, 4)
      .map((m) => m.user?.name || m.user_id)
      .join('、') || ''
  );
}
function formatTime(d: string) {
  return d ? new Date(d).toLocaleString('zh-CN') : '-';
}
function statusLabel(s: string) {
  const m: Record<string, string> = { active: '进行中', completed: '已完成', archived: '已归档' };
  return m[s] || s;
}
const templateLabels: Record<string, string> = {
  default: '日常任务',
  data_analysis: '数据分析',
  special_project: '专项行动',
  emergency_canvas: '紧急协查',
  collaborative_writing: '协同作战',
};
</script>

<template>
  <div class="relative min-h-full">
    <!-- Tab bar -->
    <div class="flex items-center gap-3 mb-6">
      <button
        v-for="tab in [
          { label: '全部', value: 'all' },
          { label: '待办', value: 'active' },
          { label: '指派', value: 'assigned' },
          { label: '盯办', value: 'red' },
          { label: '已完成', value: 'completed' },
        ]"
        :key="tab.value"
        :class="[
          'px-4 py-1.5 rounded-btn text-sm font-medium transition-smooth',
          activeTab === tab.value
            ? 'bg-[#3B82F6] text-white'
            : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-600',
        ]"
        @click="handleTabClick(tab.value)"
      >
        {{ tab.label }}
      </button>
      <div class="w-px h-6 bg-slate-200 dark:bg-slate-700 mx-1" />
      <button
        :class="[
          'px-4 py-1.5 rounded-btn text-sm font-medium transition-smooth flex items-center gap-1.5',
          activeTab === 'groups'
            ? 'bg-gradient-to-r from-purple-500 to-blue-500 text-white'
            : 'bg-white dark:bg-slate-800 text-purple-600 dark:text-purple-400 hover:bg-purple-50 dark:hover:bg-slate-700 border border-purple-200 dark:border-purple-700',
        ]"
        @click="handleTabClick('groups')"
      >
        <span>🏢</span><span>专项工作组</span>
      </button>
      <button
        v-if="activeTab === 'groups'"
        class="ml-auto px-4 py-1.5 text-sm font-medium text-white bg-gradient-to-r from-purple-500 to-blue-500 hover:from-purple-600 hover:to-blue-600 rounded-lg transition-smooth shadow-sm flex items-center gap-1.5"
        @click="openWGModal()"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        一键创建
      </button>
    </div>

    <!-- Error -->
    <div
      v-if="noteStore.error"
      class="mb-6 px-4 py-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-card text-sm text-red-600 dark:text-red-300 flex items-center justify-between"
    >
      <span>{{ noteStore.error }}</span>
      <button
        class="text-xs text-red-500 dark:text-red-300 underline hover:text-red-700 dark:hover:text-red-200 ml-4"
        @click="noteStore.fetchNotes()"
      >
        重试
      </button>
    </div>

    <!-- ====== 专项行动列表 ====== -->
    <template v-if="activeTab === 'groups'">
      <div
        class="mb-4 p-4 rounded-xl border border-purple-100 dark:border-purple-800 bg-purple-50/30 dark:bg-purple-900/5"
      >
        <div class="flex flex-wrap items-center gap-3">
          <input
            v-model="groupsFilter.keyword"
            type="text"
            class="px-3 py-1.5 text-sm border border-purple-200 dark:border-purple-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-purple-400 placeholder-slate-400 w-48"
            placeholder="🔍 关键词搜索..."
            @keydown.enter.prevent="applyGroupsFilter()"
          />
          <input
            v-model="groupsFilter.date_from"
            type="date"
            class="px-3 py-1.5 text-sm border border-purple-200 dark:border-purple-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-purple-400"
            @change="applyGroupsFilter()"
          />
          <span class="text-xs text-slate-400">至</span>
          <input
            v-model="groupsFilter.date_to"
            type="date"
            class="px-3 py-1.5 text-sm border border-purple-200 dark:border-purple-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-purple-400"
            @change="applyGroupsFilter()"
          />
          <button
            class="px-3 py-1.5 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-smooth"
            @click="resetGroupsFilter()"
          >
            🔄 重置
          </button>
          <span class="ml-auto text-xs text-slate-400 dark:text-slate-500"
            >共 {{ groupsTotal }} 个专项行动</span
          >
        </div>
      </div>

      <div v-if="wgLoading" class="flex items-center justify-center py-16">
        <div
          class="animate-spin rounded-full h-8 w-8 border-2 border-purple-500 border-t-transparent"
        ></div>
      </div>

      <div
        v-else-if="workGroups.length === 0"
        class="text-center py-16 text-slate-400 dark:text-slate-500"
      >
        <p class="text-3xl mb-3">🏢</p>
        <p class="text-sm">暂无专项行动</p>
        <p class="text-xs mt-1">点击右上角「一键创建」发起跨部门专项协同工作</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="wg in workGroups"
          :key="wg.id"
          class="group p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:shadow-md hover:border-purple-200 dark:hover:border-purple-700 transition-smooth cursor-pointer"
          @click="goToGroup(wg.id)"
        >
          <div class="flex items-start justify-between mb-2">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-base font-semibold text-slate-800 dark:text-slate-200 truncate">{{
                  wg.name
                }}</span>
                <span
                  :class="[
                    'text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0',
                    wg.status === 'active'
                      ? 'bg-green-100 dark:bg-green-900/50 text-green-600 dark:text-green-400'
                      : 'bg-slate-100 dark:bg-slate-700 text-slate-500',
                  ]"
                  >{{ statusLabel(wg.status) }}</span
                >
                <span
                  class="text-[10px] px-1.5 py-0.5 rounded-full bg-purple-50 dark:bg-purple-900/30 text-purple-500 dark:text-purple-400 font-medium shrink-0"
                  >{{ templateLabels[wg.template_type] || wg.template_type }}</span
                >
              </div>
              <p
                v-if="wg.description"
                class="text-xs text-slate-500 dark:text-slate-400 line-clamp-2 mb-1.5"
              >
                {{ wg.description }}
              </p>
              <div class="flex items-center gap-4 text-[11px] text-slate-400 dark:text-slate-500">
                <span>👤 {{ wg.initiator?.name || '未知' }}</span>
                <span>👥 {{ getMemberCount(wg) }} 人</span>
                <span class="truncate max-w-[200px]">{{ getMemberNames(wg) }}</span>
                <span v-if="wg.due_time" class="text-red-400"
                  >📅 截止 {{ wg.due_time.slice(0, 10) }}</span
                >
                <span class="text-slate-300 dark:text-slate-600">{{
                  formatTime(wg.created_at)
                }}</span>
              </div>
            </div>
            <button
              class="text-[11px] text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-smooth opacity-0 group-hover:opacity-100 shrink-0 ml-3"
              @click.stop="handleDeleteGroup(wg.id)"
              title="删除"
            >
              🗑
            </button>
          </div>
        </div>
      </div>

      <div v-if="groupsTotal > groupsPageSize" class="flex items-center justify-between mt-6">
        <span class="text-xs text-slate-400">共 {{ groupsTotal }} 个专项行动</span>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
            :disabled="groupsPage <= 1"
            @click="
              groupsPage--;
              loadWorkGroups();
            "
          >
            上一页
          </button>
          <button
            class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
            :disabled="groupsPage * groupsPageSize >= groupsTotal"
            @click="
              groupsPage++;
              loadWorkGroups();
            "
          >
            下一页
          </button>
        </div>
      </div>
    </template>

    <!-- ====== 任务内容区 ====== -->
    <template v-else>
      <div
        v-if="noteStore.loading && noteStore.activeNotes.length === 0"
        class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5"
      >
        <div v-for="n in 6" :key="n" class="skeleton h-44 rounded-card" />
      </div>
      <div
        v-else-if="!noteStore.loading && displayedNotes.length === 0 && !noteStore.error"
        class="flex flex-col items-center justify-center py-24"
      >
        <div
          class="w-24 h-24 bg-slate-100 dark:bg-slate-800 rounded-3xl flex items-center justify-center mb-6"
        >
          <svg
            class="w-12 h-12 text-slate-300 dark:text-slate-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
        <p class="text-slate-400 dark:text-slate-500 text-sm">
          {{ activeTab === 'completed' ? '暂无已完成任务' : '暂无活跃任务' }}
        </p>
        <p class="text-slate-300 dark:text-slate-600 text-xs mt-1">点击右下角 '+' 新建任务</p>
      </div>
      <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
        <StickyNoteCard
          v-for="note in displayedNotes"
          :key="note.id"
          :note="note"
          mode="web"
          :archived="false"
          :extra-actions="true"
          class="animate-spring-enter"
          @click="openDetail(note)"
          @complete="handleComplete"
          @important="handleImportant"
          @delete="handleDelete"
        />
      </div>
    </template>

    <!-- FAB -->
    <button
      class="fixed right-8 bottom-8 w-14 h-14 rounded-full bg-[#3B82F6] text-white shadow-btn-float hover:bg-blue-600 active:scale-95 transition-smooth flex items-center justify-center z-30"
      @click="openCreateModal"
    >
      <svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2.5"
          d="M12 4v16m8-8H4"
        />
      </svg>
    </button>

    <!-- ====== 新建任务模态框 ====== -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showCreateModal = false" />
        <div
          class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-3xl mx-4 animate-fade-in max-h-[90vh] flex flex-col"
        >
          <div class="p-6 overflow-y-auto flex-1">
            <div class="flex items-center justify-between mb-6">
              <div>
                <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                  📝 新建任务
                </h2>
                <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                  支持 Markdown 语法，可全屏编写
                </p>
              </div>
              <button
                class="p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
                @click="showCreateModal = false"
              >
                <svg
                  class="w-5 h-5 text-slate-400 dark:text-slate-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <form class="space-y-4" @submit.prevent="handleSubmit">
              <input
                v-model="newTitle"
                class="input-field"
                placeholder="任务标题"
                autofocus
                @keydown.enter.prevent
              />
              <div>
                <label class="block text-xs font-medium text-slate-500 mb-1.5"
                  >任务内容（支持 Markdown，可全屏）</label
                >
                <MarkdownEditor
                  v-model="newContent"
                  placeholder="请输入任务内容..."
                  :min-height="200"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-500 mb-1"
                  >使用模板（可选）</label
                >
                <select
                  v-model="selectedTemplateId"
                  class="w-full text-sm border border-slate-200 dark:border-slate-600 rounded p-2 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100"
                  @change="onTemplateSelect"
                >
                  <option value="">不使用模板</option>
                  <option v-for="t in userTemplates" :key="t.id" :value="t.id">
                    {{ t.name }}{{ t.is_system ? ' (系统)' : '' }}
                  </option>
                </select>
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1.5 block">标签</span
                ><TagSelector v-model="selectedTagIds" :max="5" />
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-2 block">任务类型</span>
                <div class="flex gap-3">
                  <label
                    :class="[
                      'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-btn border-2 cursor-pointer transition-smooth',
                      sourceType === 'self'
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400'
                        : 'border-slate-200 dark:border-slate-600 text-slate-500',
                    ]"
                    ><input v-model="sourceType" type="radio" value="self" class="sr-only" /><span
                      class="text-sm font-medium"
                      >仅自己</span
                    ></label
                  >
                  <label
                    :class="[
                      'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-btn border-2 cursor-pointer transition-smooth',
                      sourceType === 'assigned'
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400'
                        : 'border-slate-200 dark:border-slate-600 text-slate-500',
                    ]"
                    ><input
                      v-model="sourceType"
                      type="radio"
                      value="assigned"
                      class="sr-only"
                    /><span class="text-sm font-medium">指派他人</span></label
                  >
                </div>
              </div>
              <div v-if="sourceType !== 'self'">
                <span class="text-xs text-slate-500 mb-1.5 block">{{
                  sourceType === 'assigned'
                    ? '选择指派人员（可多选，支持一键全选部门/小组）'
                    : '选择协作人员（可多选）'
                }}</span
                ><UserPicker
                  v-model="selectedAssigneeIds"
                  :multiple="true"
                  :max="50"
                  :drop-up="true"
                  :disabled-ids="selectedCcIds"
                  disabled-note="已在抄送人员中"
                />
                <div class="mt-3 flex items-center gap-3">
                  <label class="text-xs text-slate-500 shrink-0">⏱ 工作时间</label>
                  <select v-model="workTimeSeconds" class="input-field !py-1.5 !text-sm !w-auto">
                    <option v-for="opt in workTimeOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                  <span class="text-[11px] text-slate-400"
                    >任务下发后开始倒计时，到期前自动提醒</span
                  >
                </div>
                <div
                  class="mt-3 p-3 rounded-lg border border-dashed border-blue-200 dark:border-blue-700 bg-blue-50/50 dark:bg-blue-900/10"
                >
                  <div class="flex items-center gap-2 mb-2">
                    <span class="text-xs text-slate-500">🔍 工作类型</span>
                    <select
                      v-model="selectedWorkType"
                      class="input-field !py-1 !text-xs !w-auto"
                      @change="handleRecommend()"
                    >
                      <option value="">选择类型</option>
                      <option v-for="opt in workTypeOptions" :key="opt.value" :value="opt.value">
                        {{ opt.label }}
                      </option>
                    </select>
                    <button
                      type="button"
                      class="text-xs px-2.5 py-1 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded-btn hover:from-blue-600 hover:to-purple-600 transition-smooth disabled:opacity-50 flex items-center gap-1"
                      :disabled="!selectedWorkType || recommending"
                      @click="handleRecommend()"
                    >
                      <span
                        v-if="recommending"
                        class="inline-block w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin"
                      ></span>
                      🤖 一键推荐
                    </button>
                  </div>
                  <p v-if="recommendError" class="text-xs text-red-500 mb-2">
                    {{ recommendError }}
                  </p>
                  <div
                    v-if="recommendResult.length > 0"
                    class="flex flex-wrap gap-1 max-h-24 overflow-y-auto"
                  >
                    <button
                      v-for="u in recommendResult"
                      :key="u.id"
                      type="button"
                      :class="[
                        'text-xs px-2 py-1 rounded-full transition-smooth',
                        selectedAssigneeIds.includes(u.id)
                          ? 'bg-blue-500 text-white'
                          : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-blue-100 dark:hover:bg-blue-900/40 border border-slate-200 dark:border-slate-600',
                      ]"
                      @click="selectRecommendUser(u.id)"
                    >
                      {{ u.name }}
                      <span v-if="u.dept_name" class="text-[10px] opacity-60 ml-0.5">{{
                        u.dept_name
                      }}</span>
                      <span v-if="selectedAssigneeIds.includes(u.id)" class="ml-1">✓</span>
                    </button>
                  </div>
                </div>
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1.5 block"
                  >抄送人员（可多选，抄送人可查看该任务——紫色卡片 +「抄送」徽章）</span
                ><UserPicker
                  v-model="selectedCcIds"
                  :multiple="true"
                  :max="50"
                  :drop-up="true"
                  :disabled-ids="selectedAssigneeIds"
                  disabled-note="已在指派人员中"
                />
              </div>
              <p v-if="createError" class="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-btn">
                {{ createError }}
              </p>
              <div
                class="flex justify-end gap-3 pt-4 border-t border-slate-100 dark:border-slate-700"
              >
                <button
                  type="button"
                  class="px-5 py-2.5 text-sm text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 rounded-btn hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth"
                  @click="showCreateModal = false"
                  :disabled="creating"
                >
                  取消
                </button>
                <button
                  type="submit"
                  class="px-5 py-2.5 text-sm text-white bg-[#3B82F6] rounded-btn hover:bg-blue-600 transition-smooth disabled:opacity-50"
                  :disabled="creating"
                >
                  {{ creating ? '创建中...' : '创建任务' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ====== 详情编辑模态框（居中） ====== -->
    <Teleport to="body">
      <div
        v-if="showDetailPanel && selectedNote"
        class="fixed inset-0 z-50 flex items-center justify-center"
      >
        <div class="overlay-backdrop" @click="closeDetail" />
        <div
          class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-3xl mx-4 animate-fade-in max-h-[90vh] flex flex-col"
        >
          <div class="p-6 overflow-y-auto flex-1">
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-2">
                <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                  📝 任务详情
                </h2>
                <span
                  v-if="selectedIsCcOnly"
                  class="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded-tag"
                  >抄送</span
                >
                <span
                  v-if="selectedIsAssigner"
                  class="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded-tag"
                  >指派</span
                >
                <span
                  v-else-if="selectedNote.color_status === 'red'"
                  class="text-xs px-2 py-0.5 bg-red-100 text-red-700 rounded-tag"
                  >盯办中</span
                >
                <span
                  v-if="selectedNote.color_status === 'blue'"
                  class="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded-tag"
                  >协作</span
                >
              </div>
              <button
                class="p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
                @click="closeDetail"
              >
                <svg
                  class="w-5 h-5 text-slate-400 dark:text-slate-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <div class="space-y-5">
              <div>
                <span class="text-xs text-slate-400 mb-1 block">标题</span
                ><input
                  v-model="editingTitle"
                  class="input-field text-base font-semibold"
                  @keydown.enter.prevent
                />
              </div>
              <div>
                <span class="text-xs text-slate-400 mb-1 block">内容（支持 Markdown，可全屏）</span>
                <MarkdownEditor
                  v-model="editingContent"
                  placeholder="请输入任务内容..."
                  :min-height="220"
                />
              </div>
              <div>
                <span class="text-xs text-slate-400 mb-1 flex items-center gap-2">
                  标签
                  <span v-if="tagSaving" class="text-[10px] text-blue-400">保存中...</span>
                  <span v-if="tagError" class="text-[10px] text-red-400">{{ tagError }}</span>
                </span>
                <TagSelector
                  v-model="selectedEditingTagIds"
                  :max="10"
                  scope="all"
                  @update:model-value="handleUpdateTags"
                />
              </div>
              <div class="bg-slate-50 dark:bg-slate-900 rounded-card p-4 space-y-2">
                <div class="flex justify-between text-xs">
                  <span class="text-slate-400">来源类型</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.source_type === 'self'
                      ? '自己创建'
                      : selectedNote.source_type === 'assigned'
                        ? '上级指派'
                        : '协同任务'
                  }}</span>
                </div>
                <div class="flex justify-between text-xs" v-if="selectedNote.creator?.name">
                  <span class="text-slate-400">创建人</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.creator.name
                  }}</span>
                </div>
                <div class="flex justify-between text-xs" v-if="selectedNote.assignees?.length">
                  <span class="text-slate-400">负责人</span
                  ><span
                    class="text-slate-700 dark:text-slate-300 flex items-center gap-1.5 flex-wrap justify-end"
                  >
                    <template v-for="a in selectedNote.assignees" :key="a.user_id || (a as any).id">
                      <span class="flex items-center gap-1">
                        <span>{{ a.user?.name || (a as any).name }}</span>
                        <span
                          v-if="(a as any).sign_status === 'signed'"
                          class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300"
                          :title="
                            (a as any).signed_at
                              ? '签收于 ' + (a as any).signed_at.slice(0, 16).replace('T', ' ')
                              : ''
                          "
                          >已签收</span
                        >
                        <span
                          v-else-if="(a as any).role_in_note !== 'initiator'"
                          class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-400 dark:bg-slate-700 dark:text-slate-500"
                          >未签收</span
                        >
                        <span
                          v-if="(a as any).is_completed && (a as any).role_in_note !== 'initiator'"
                          class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300"
                          :title="
                            (a as any).completed_at
                              ? '完成于 ' + (a as any).completed_at.slice(0, 16).replace('T', ' ')
                              : ''
                          "
                          >已完成</span
                        >
                      </span>
                      <button
                        v-if="(a.user_id || (a as any).id) !== auth.user?.id"
                        class="text-blue-500 hover:text-blue-600 hover:underline shrink-0"
                        title="发起聊天"
                        @click="notifStore.openChat(a.user_id || (a as any).id)"
                      >
                        联系
                      </button>
                    </template>
                  </span>
                </div>
                <div class="flex justify-between text-xs">
                  <span class="text-slate-400">创建时间</span
                  ><span class="text-slate-700 dark:text-slate-300">{{
                    selectedNote.created_at?.slice(0, 16).replace('T', ' ')
                  }}</span>
                </div>
                <div class="flex justify-between text-xs" v-if="selectedNote.due_time">
                  <span class="text-slate-400">截止时间</span
                  ><span class="text-red-500">{{
                    selectedNote.due_time.slice(0, 16).replace('T', ' ')
                  }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="pt-4 border-t border-slate-100 dark:border-slate-700 p-6 space-y-3">
            <button
              v-if="canSignSelected"
              class="w-full py-2.5 text-sm bg-blue-500 text-white rounded-btn hover:bg-blue-600 active:scale-[0.99] transition-smooth"
              @click="handleSign"
            >
              签收任务
            </button>
            <div class="flex gap-2">
              <button
                v-if="!selectedIsCcOnly"
                class="flex-1 py-2.5 btn-primary text-sm disabled:opacity-50"
                :disabled="saving"
                @click="handleSaveDetail"
              >
                {{ saving ? '保存中...' : '保存' }}
              </button>
              <button
                v-if="selectedCompleteLabel"
                class="flex-1 py-2.5 text-sm bg-green-500 text-white rounded-btn hover:bg-green-600 transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="completing || selectedCompleteDisabled"
                :title="
                  selectedCompleteDisabled && !selectedAllCompleted
                    ? '尚有被指派人未完成任务，暂不能归档'
                    : ''
                "
                @click="handleComplete(selectedNote!)"
              >
                {{ completing ? '提交中...' : selectedCompleteLabel }}
              </button>
              <button
                v-if="!selectedIsCcOnly && selectedNote.color_status !== 'red'"
                class="flex-1 py-2.5 text-sm bg-red-50 text-red-600 rounded-btn hover:bg-red-100 transition-smooth"
                @click="handleImportant(selectedNote!)"
              >
                重要
              </button>
              <button
                class="flex-1 py-2.5 text-sm bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
                @click="handleDelete(selectedNote!)"
              >
                删除
              </button>
            </div>
            <button class="w-full py-2 btn-secondary text-sm" @click="closeDetail">关闭</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ====== 专项工作组创建模态框 ====== -->
    <Teleport to="body">
      <div
        v-if="showWorkGroupModal"
        class="fixed inset-0 z-50 flex items-start justify-center pt-[5vh]"
      >
        <div class="overlay-backdrop" @click="showWorkGroupModal = false" />
        <div
          class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-2xl mx-4 animate-fade-in max-h-[90vh] flex flex-col"
        >
          <div class="p-6 overflow-auto flex-1">
            <div class="flex items-center justify-between mb-6">
              <div>
                <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                  🏢 一键创建专项工作组
                </h2>
                <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                  创建工作组并自动将任务分发至每位成员
                </p>
              </div>
              <button
                class="p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
                @click="showWorkGroupModal = false"
              >
                <svg
                  class="w-5 h-5 text-slate-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <form class="space-y-4" @submit.prevent="handleCreateWorkGroup" @keydown.enter.prevent>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="text-xs text-slate-500 mb-1 block">工作组名称 *</label
                  ><input v-model="wgName" class="input-field" placeholder="如：雷霆2026专项行动" />
                </div>
                <div>
                  <label class="text-xs text-slate-500 mb-1 block">模板类型</label
                  ><select v-model="wgTemplate" class="input-field">
                    <option value="default">日常工作任务</option>
                    <option value="data_analysis">数据分析研判</option>
                    <option value="special_project">专项行动方案</option>
                    <option value="emergency_canvas">紧急协查通报</option>
                    <option value="collaborative_writing">协同作战方案</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="text-xs text-slate-500 mb-1 block">工作要求描述</label
                ><textarea
                  v-model="wgDescription"
                  class="input-field h-20 resize-none"
                  placeholder="填写专项工作的具体要求、目标、时间节点及交付标准..."
                />
              </div>
              <div>
                <label class="text-xs text-slate-500 mb-1 block">截止日期</label
                ><input v-model="wgDueDate" type="date" class="input-field" />
              </div>
              <div
                v-if="availablePresets.length > 0"
                class="p-3 rounded-lg border border-dashed border-orange-200 bg-orange-50/50"
              >
                <label class="text-xs text-slate-500 mb-1.5 flex items-center gap-1">
                  📋 选择人员预设组
                  <span class="text-[10px] text-orange-500">（自动填充成员）</span>
                </label>
                <select
                  v-model="selectedPresetId"
                  class="input-field !py-1.5 !text-sm"
                  @change="handlePresetSelect"
                >
                  <option value="">不使用预设</option>
                  <option v-for="preset in availablePresets" :key="preset.id" :value="preset.id">
                    {{ preset.name }} ({{ preset.members?.length || 0 }}人)
                  </option>
                </select>
              </div>
              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="text-xs text-slate-500">工作小组设置</label
                  ><button
                    type="button"
                    class="text-xs text-blue-500 hover:text-blue-600 font-medium"
                    @click="addSubGroup"
                  >
                    + 添加小组
                  </button>
                </div>
                <div class="space-y-3">
                  <div
                    v-for="(sg, idx) in wgSubGroups"
                    :key="idx"
                    class="p-3 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/40"
                  >
                    <div class="flex items-center gap-2 mb-2">
                      <span
                        class="text-[10px] px-1.5 py-0.5 rounded font-medium"
                        :class="
                          idx === 0
                            ? 'bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400'
                            : 'bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400'
                        "
                        >{{ idx === 0 ? '组长组' : `小组${idx + 1}` }}</span
                      >
                      <input
                        v-model="sg.name"
                        class="flex-1 text-xs px-2 py-1 border border-slate-200 dark:border-slate-600 rounded bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100"
                        placeholder="小组名称（可选）"
                      />
                      <button
                        v-if="wgSubGroups.length > 1"
                        type="button"
                        class="text-xs text-red-400 hover:text-red-600"
                        @click="removeSubGroup(idx)"
                      >
                        ✕
                      </button>
                    </div>
                    <UserPicker
                      :model-value="selectedWGUserIds[idx] || []"
                      :multiple="true"
                      :max="50"
                      @update:model-value="onWGUserSelect(idx, $event)"
                    />
                    <p class="text-[10px] text-slate-400 dark:text-slate-500 mt-1">
                      {{
                        idx === 0 ? '第一组为组长组，成员角色自动设为组长' : '成员角色自动设为组员'
                      }}
                    </p>
                  </div>
                </div>
              </div>
              <p v-if="wgError" class="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-btn">
                {{ wgError }}
              </p>
              <div
                class="flex justify-end gap-3 pt-4 border-t border-slate-100 dark:border-slate-700"
              >
                <button
                  type="button"
                  class="px-5 py-2.5 text-sm text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 rounded-btn hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth"
                  @click="showWorkGroupModal = false"
                  :disabled="wgCreating"
                >
                  取消
                </button>
                <button
                  type="submit"
                  class="px-5 py-2.5 text-sm text-white bg-gradient-to-r from-purple-500 to-blue-500 rounded-btn hover:from-purple-600 hover:to-blue-600 transition-smooth disabled:opacity-50"
                  :disabled="wgCreating"
                >
                  {{ wgCreating ? '创建中...' : '一键创建并分发任务' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 任务反馈填报弹窗 -->
    <FeedbackModal
      :visible="feedbackVisible"
      :note="feedbackNote"
      @update:visible="feedbackVisible = $event"
      @submit="submitFeedback"
    />
  </div>
</template>
