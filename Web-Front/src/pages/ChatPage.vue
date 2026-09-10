<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import { useAuthStore } from '@/stores/auth';
import { getVisibleUsers } from '@/services/admin';
import { uploadChatFile } from '@/services/notification';
import * as groupChatService from '@/services/groupChat';
import EmojiPicker from '@/components/chat/EmojiPicker.vue';
import MentionPicker from '@/components/chat/MentionPicker.vue';
import type { MentionOption } from '@/components/chat/MentionPicker.vue';
import UserPicker from '@/components/common/UserPicker.vue';
import type { ChatMessageItem, GroupConversationItem, GroupMessageItem, GroupMemberItem, GroupFileItem, User } from '@/types';
import { renderNoteContent } from '@/utils/richText';
import { matchPinyin } from '@/utils/pinyin';
import {
  memberDisplayName,
  buildMentionInsert,
  findActiveMentionToken,
  splitMentionSegments,
  extractMentionedNames,
  renderMessageContent,
} from '@/utils/mention';
import { useToast } from '@/composables/useToast';

const store = useNotificationStore();
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const toast = useToast();

const leftTab = ref<'conv' | 'group' | 'contact'>('conv');
const keyword = ref('');
const currentPeer = ref<string | null>(null);
const currentGroup = ref<string | null>(null);
const input = ref('');
const users = ref<User[]>([]);
const messagesEl = ref<HTMLDivElement | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const uploading = ref(false);
const showEmoji = ref(false);
// 需求33：表情面板元素引用，用于「点击页面其他地方关闭」
const emojiBtnEl = ref<HTMLElement | null>(null);
const emojiPanelEl = ref<HTMLElement | null>(null);
const previewUrl = ref('');
const loadingOlder = ref(false);
const allLoaded = ref(false);
const page = ref(1);
const PAGE_SIZE = 30;
// 用户是否已上滑查看历史消息（不在底部）：此时收到新消息不自动下拉，显示浮动按钮
const scrolledUp = ref(false);
// 上滑查看历史期间收到的新消息数（浮动按钮上的角标）
const newMsgCount = ref(0);

// 群聊成员（群设置弹窗 / 消息发送者名解析）
const groupMembers = ref<GroupMemberItem[]>([]);
const membersLoading = ref(false);
// 创建群聊弹窗
const showCreateGroup = ref(false);
const newGroupName = ref('');
const newGroupMembers = ref<string[]>([]);
const creatingGroup = ref(false);
// 群设置弹窗
const showGroupSettings = ref(false);
const renameValue = ref('');
const addMemberIds = ref<string[]>([]);
const savingGroup = ref(false);
// 群公告（横幅展开态 + 群设置中的编辑态）
const announcementExpanded = ref(false);
const announcementEdit = ref('');
const savingAnnouncement = ref(false);
// 群文件侧滑面板
const showFilesPanel = ref(false);
const fileSearch = ref('');
const filePanelDragOver = ref(false);
const filePanelUploading = ref(false);
const downloadingId = ref('');
const filePanelInput = ref<HTMLInputElement | null>(null);

// ===== @ 提及（群聊，参考微信） =====
const inputEl = ref<HTMLTextAreaElement | null>(null);
const mentionPickerEl = ref<InstanceType<typeof MentionPicker> | null>(null);
const mentionPanelEl = ref<HTMLElement | null>(null);
const mirrorEl = ref<HTMLElement | null>(null);
// 当前活动的 @ 记号（@ 下标 + 过滤词），null 表示浮层关闭
const mentionToken = ref<{ at: number; keyword: string } | null>(null);

// 好友姓名映射（会话中对方姓名兜底）
const nameMap = computed(() => {
  const map: Record<string, string> = {};
  users.value.forEach((u) => {
    map[u.id] = u.name || u.username;
  });
  return map;
});

const filteredUsers = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  const me = auth.user?.id;
  let list = users.value.filter((u) => u.id !== me && u.is_active !== false);
  if (kw) {
    // 需求36：支持拼音全拼 / 首字母搜索（无视大小写）
    list = list.filter((u) => matchPinyin(kw, u.name || '', u.username || '', u.dept_name || ''));
  }
  // 在线用户优先展示（仿微信通讯录）
  return [...list].sort((a, b) => Number(store.isOnline(b.id)) - Number(store.isOnline(a.id)));
});

// 群聊列表（支持按群名搜索）
const filteredGroups = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  let list = store.groupConversations;
  if (kw) {
    list = list.filter((g) => matchPinyin(kw, g.name));
  }
  return list;
});

const currentPeerName = computed(() => {
  if (!currentPeer.value) return '';
  const conv = store.conversations.find((c) => c.peer_id === currentPeer.value);
  if (conv?.peer_name) return conv.peer_name;
  return nameMap.value[currentPeer.value] || `用户 ${currentPeer.value.slice(0, 6)}`;
});

const currentGroupInfo = computed(() =>
  currentGroup.value ? store.groupConversations.find((g) => g.id === currentGroup.value) : undefined
);

const isGroupOwner = computed(
  () => !!currentGroupInfo.value && currentGroupInfo.value.owner_id === auth.user?.id
);

// 群公告（进入群会话后由 store 缓存）
const currentAnnouncement = computed(() =>
  currentGroup.value ? store.groupAnnouncements[currentGroup.value] : undefined
);

// 群文件列表（带搜索过滤：文件名 / 上传者名）
const currentGroupFiles = computed<GroupFileItem[]>(() => {
  const gid = currentGroup.value;
  if (!gid) return [];
  const list = store.groupFiles[gid] || [];
  const kw = fileSearch.value.trim().toLowerCase();
  if (!kw) return list;
  return list.filter(
    (f) =>
      f.file_name.toLowerCase().includes(kw) ||
      (f.uploader_name || '').toLowerCase().includes(kw)
  );
});

// 当前展示的是否为群聊会话（群聊优先于私聊）
const activeIsGroup = computed(() => !!currentGroup.value);

// 群成员 id 列表（群设置弹窗添加成员时禁用已在群中的用户）
const existingMemberIds = computed(() => groupMembers.value.map((m) => m.user_id));

// 群消息发送者名：推送附带的 sender_name → sender 关联 → 群成员表 → 好友映射兜底
function senderNameOf(m: GroupMessageItem): string {
  if (m.sender_name) return m.sender_name;
  if (m.sender?.name) return m.sender.name;
  if (m.sender?.username) return m.sender.username;
  const mem = groupMembers.value.find((x) => x.user_id === m.sender_id);
  if (mem?.user?.name) return mem.user.name;
  if (mem?.user?.username) return mem.user.username;
  return nameMap.value[m.sender_id] || `用户 ${m.sender_id.slice(0, 6)}`;
}

// 统一渲染模型：私聊 / 群聊消息归一化后展示
interface DisplayMessage {
  id: string;
  senderId: string;
  type: string;
  content: string;
  fileName?: string;
  filePath?: string;
  fileSize?: number;
  createdAt: string;
  isRead?: boolean;
  senderName: string;
}

const displayMessages = computed<DisplayMessage[]>(() => {
  if (activeIsGroup.value) {
    return (store.groupMessages[currentGroup.value || ''] || []).map((m) => ({
      id: m.id,
      senderId: m.sender_id,
      type: m.type,
      content: m.content,
      fileName: m.file_name,
      filePath: m.file_path,
      fileSize: m.file_size,
      createdAt: m.created_at,
      senderName: senderNameOf(m),
    }));
  }
  return (store.messages[currentPeer.value || ''] || []).map((m) => ({
    id: m.id,
    senderId: m.sender_id,
    type: m.type,
    content: m.content,
    fileName: m.file_name,
    filePath: m.file_path,
    fileSize: m.file_size,
    createdAt: m.created_at,
    isRead: m.is_read,
    senderName: peerName(m.sender_id),
  }));
});

// ===== @ 提及（群聊，参考微信） =====

// @ 提及候选名：群成员显示名 + 全体成员（气泡高亮与发送反推共用）
const mentionCandidates = computed(() => {
  if (!activeIsGroup.value) return [];
  const names = groupMembers.value.map((m) => memberDisplayName(m));
  names.push('全体成员');
  return names;
});

// 输入框镜像分段：@名字 高亮，其余透明
const mirrorSegments = computed(() =>
  activeIsGroup.value ? splitMentionSegments(input.value, mentionCandidates.value) : []
);

// 输入/光标变化时刷新活动 @ 记号（决定浮层开关与过滤词）
function refreshMentionToken() {
  if (!activeIsGroup.value || !inputEl.value) {
    mentionToken.value = null;
    return;
  }
  mentionToken.value = findActiveMentionToken(input.value, inputEl.value.selectionStart ?? input.value.length);
  // 弹出 @ 浮层时收起表情面板，避免遮挡
  if (mentionToken.value) showEmoji.value = false;
}

// 选中浮层候选项：把「@过滤词」替换为「@名字 」，光标落在插入文本之后
function onMentionSelect(opt: MentionOption) {
  const token = mentionToken.value;
  const el = inputEl.value;
  if (!token || !el) return;
  const caret = el.selectionStart ?? input.value.length;
  const insert =
    opt.type === 'all'
      ? buildMentionInsert('全体成员')
      : buildMentionInsert(memberDisplayName(opt.member!));
  input.value = input.value.slice(0, token.at) + insert + input.value.slice(caret);
  mentionToken.value = null;
  // 同步把光标移到插入文本末尾：避免后续 keyup 等事件用旧光标位置重新激活浮层
  el.focus();
  el.setSelectionRange(token.at + insert.length, token.at + insert.length);
  nextTick(() => {
    el.focus();
    const pos = token.at + insert.length;
    el.setSelectionRange(pos, pos);
  });
}

// 浮层打开时的 ↑↓ 导航（输入框 keydown 转发；未打开时保持光标移动的默认行为）
function onMentionNavKey(e: KeyboardEvent, dir: 'up' | 'down') {
  if (!mentionToken.value) return;
  e.preventDefault();
  if (dir === 'up') mentionPickerEl.value?.moveUp();
  else mentionPickerEl.value?.moveDown();
}

// 仅光标移动/删除类按键的 keyup 才刷新 @ 记号：
// 文本变更由 @input 覆盖，Enter 的 keyup 若刷新会在 DOM 尚未同步时用旧光标重开浮层
const MENTION_CURSOR_KEYS = new Set(['ArrowLeft', 'ArrowRight', 'Home', 'End', 'Delete', 'Backspace']);
function onMentionKeyup(e: KeyboardEvent) {
  if (MENTION_CURSOR_KEYS.has(e.key)) refreshMentionToken();
}

// Enter：浮层打开时选中高亮项，否则发送消息
function onEnterKeydown() {
  if (mentionToken.value) {
    mentionPickerEl.value?.pickActive();
    return;
  }
  sendText();
}

// 发送时从内容反推被 @ 的成员 ID（删除 @文本 自动失效；@全体成员 = 除自己外全部成员）
function collectMentionIds(content: string): string[] {
  const names = new Set(extractMentionedNames(content, mentionCandidates.value));
  if (names.size === 0) return [];
  const me = auth.user?.id;
  if (names.has('全体成员')) {
    return groupMembers.value.filter((m) => m.user_id !== me).map((m) => m.user_id);
  }
  return groupMembers.value
    .filter((m) => m.user_id !== me && names.has(memberDisplayName(m)))
    .map((m) => m.user_id);
}

// 群聊文本气泡渲染：@名字 高亮（mention-tag），私聊沿用原转义
function renderGroupBubble(content: string): string {
  if (!activeIsGroup.value) return renderNoteContent(content);
  return renderMessageContent(content, mentionCandidates.value);
}

// 镜像层与输入框滚动同步（内容超过两行时保持高亮对齐）
function syncMirrorScroll() {
  if (mirrorEl.value && inputEl.value) mirrorEl.value.scrollTop = inputEl.value.scrollTop;
}

async function loadUsers() {
  try {
    const res = await getVisibleUsers();
    users.value = res.data || [];
  } catch {
    /* ignore */
  }
}

function peerName(id: string) {
  return nameMap.value[id] || `用户 ${id.slice(0, 6)}`;
}

async function openConversation(peerId: string, name?: string) {
  currentPeer.value = peerId;
  // 切换到私聊：退出群聊查看状态
  currentGroup.value = null;
  store.setViewingGroup(null);
  page.value = 1;
  allLoaded.value = false;
  showEmoji.value = false;
  scrolledUp.value = false;
  newMsgCount.value = 0;
  store.setViewingPeer(peerId);
  store.fetchOnlineUsers();
  await store.loadMessages(peerId);
  await store.markConversationRead(peerId);
  scrollToBottom();
}

async function openGroup(groupId: string) {
  // 群可能已被解散或不存在（如浏览器残留失效的 ?group= 参数）：
  // 先校验群详情，404 时提示并回退空态，避免渲染出「文件上传/下载全部 404」的空壳视图
  try {
    await groupChatService.fetchGroupDetail(groupId);
  } catch (err) {
    const status = (err as { response?: { status?: number } })?.response?.status;
    if (status === 404) {
      toast.error('该群聊不存在或已被解散');
      currentGroup.value = null;
      store.setViewingGroup(null);
      router.replace({ path: '/chat' });
      return;
    }
    // 其他错误（网络抖动等）不阻断，按原逻辑继续加载
  }
  currentGroup.value = groupId;
  // 切换到群聊：退出私聊查看状态
  currentPeer.value = null;
  store.setViewingPeer(null);
  page.value = 1;
  allLoaded.value = false;
  showEmoji.value = false;
  scrolledUp.value = false;
  newMsgCount.value = 0;
  store.setViewingGroup(groupId);
  announcementExpanded.value = false;
  try {
    await store.loadGroupMessages(groupId);
    await store.markGroupRead(groupId);
  } catch {
    /* ignore */
  }
  scrollToBottom();
  // 群成员用于发送者名解析与群设置弹窗，后台加载不阻塞消息展示
  loadGroupMembers(groupId);
  // 群公告：横幅展示（后台加载）
  store.loadGroupAnnouncement(groupId);
}

async function loadGroupMembers(groupId: string) {
  membersLoading.value = true;
  try {
    const res = await groupChatService.fetchGroupMembers(groupId);
    groupMembers.value = (res.data as unknown as GroupMemberItem[]) || [];
  } catch {
    /* ignore */
  } finally {
    membersLoading.value = false;
  }
}

function scrollToBottom(smooth = true) {
  const el = messagesEl.value;
  if (!el) return;
  const scroll = () => {
    el.scrollTo({ top: el.scrollHeight, behavior: smooth ? 'smooth' : 'auto' });
  };
  nextTick(scroll);
  // 图片等异步内容加载会撑高容器，加载完成后补一次，确保停在最新消息处
  setTimeout(scroll, 200);
  el.querySelectorAll('img').forEach((img) => {
    if (!img.complete) img.addEventListener('load', scroll, { once: true });
  });
}

// 消息区滚动：记录是否在底部；滚动到顶部时加载更早消息
function handleMessagesScroll(e: Event) {
  const el = e.target as HTMLElement;
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  scrolledUp.value = !atBottom;
  if (atBottom) newMsgCount.value = 0;
  if (el.scrollTop < 40) loadOlder();
}

// 浮动按钮：一键回到最新消息
function goToLatest() {
  scrolledUp.value = false;
  newMsgCount.value = 0;
  scrollToBottom();
}

// 加载更早消息（私聊 / 群聊统一入口）
async function loadOlder() {
  if (loadingOlder.value || allLoaded.value) return;
  if (!activeIsGroup.value && !currentPeer.value) return;
  loadingOlder.value = true;
  const next = page.value + 1;
  try {
    if (activeIsGroup.value && currentGroup.value) {
      const res = await store.fetchGroupMessagesPage(currentGroup.value, next, PAGE_SIZE);
      const list = (res as unknown as { data: GroupMessageItem[] }) || { data: [] };
      const older = list.data || [];
      if (older.length < PAGE_SIZE) allLoaded.value = true;
      if (older.length > 0) {
        const prev = store.groupMessages[currentGroup.value] || [];
        store.groupMessages[currentGroup.value] = [...older, ...prev];
        page.value = next;
      } else {
        allLoaded.value = true;
      }
    } else if (currentPeer.value) {
      const res = await store.fetchMessagesPage(currentPeer.value, next, PAGE_SIZE);
      const list = (res as unknown as { data: ChatMessageItem[] }) || { data: [] };
      const older = list.data || [];
      if (older.length < PAGE_SIZE) allLoaded.value = true;
      if (older.length > 0) {
        const prev = store.messages[currentPeer.value] || [];
        store.messages[currentPeer.value] = [...older, ...prev];
        page.value = next;
      } else {
        allLoaded.value = true;
      }
    }
  } catch {
    /* ignore */
  } finally {
    loadingOlder.value = false;
  }
}

async function sendText() {
  const content = input.value.trim();
  if (!content) return;
  // @ 提及：从内容反推被 @ 的成员 ID（删掉 @文本 自动失效）
  const mentions = activeIsGroup.value ? collectMentionIds(content) : [];
  input.value = '';
  mentionToken.value = null;
  showEmoji.value = false;
  if (activeIsGroup.value && currentGroup.value) {
    await store.sendGroupMessage(currentGroup.value, mentions.length ? { content, mentions } : { content });
  } else if (currentPeer.value) {
    await store.sendMessage(currentPeer.value, { content });
  }
  scrollToBottom();
}

function insertEmoji(e: string) {
  input.value += e;
}

// 发送图片表情（图片消息）
async function sendEmoticon(path: string) {
  if (!path) return;
  showEmoji.value = false;
  const payload = {
    type: 'image' as const,
    file_name: '表情',
    file_path: path,
    mime_type: 'image/png',
  };
  if (activeIsGroup.value && currentGroup.value) {
    await store.sendGroupMessage(currentGroup.value, payload);
  } else if (currentPeer.value) {
    await store.sendMessage(currentPeer.value, payload);
  }
  scrollToBottom();
}

async function onFileSelected(e: Event) {
  const inputEl = e.target as HTMLInputElement;
  const file = inputEl.files?.[0];
  inputEl.value = '';
  if (!file) return;
  if (!activeIsGroup.value && !currentPeer.value) return;

  if (file.size > 10 * 1024 * 1024) {
    toast.warning('文件大小超过限制，最大允许 10MB');
    return;
  }
  uploading.value = true;
  try {
    const res = await uploadChatFile(file);
    const meta = res.data as unknown as { file_name: string; file_path: string; file_size: number; mime_type: string };
    const isImage = (meta.mime_type || '').startsWith('image/');
    const payload = {
      type: isImage ? ('image' as const) : ('file' as const),
      file_name: meta.file_name,
      file_path: meta.file_path,
      file_size: meta.file_size,
      mime_type: meta.mime_type,
    };
    if (activeIsGroup.value && currentGroup.value) {
      await store.sendGroupMessage(currentGroup.value, payload);
    } else if (currentPeer.value) {
      await store.sendMessage(currentPeer.value, payload);
    }
    scrollToBottom();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '文件上传失败');
  } finally {
    uploading.value = false;
  }
}

// ---------------- 创建群聊 ----------------

function openCreateModal() {
  newGroupName.value = '';
  newGroupMembers.value = [];
  showCreateGroup.value = true;
}

async function submitCreateGroup() {
  const name = newGroupName.value.trim();
  if (!name) {
    toast.warning('请输入群聊名称');
    return;
  }
  if (newGroupMembers.value.length === 0) {
    toast.warning('请至少选择一名成员');
    return;
  }
  creatingGroup.value = true;
  try {
    const res = await groupChatService.createGroup({ name, members: newGroupMembers.value });
    showCreateGroup.value = false;
    newGroupName.value = '';
    newGroupMembers.value = [];
    toast.success('群聊创建成功');
    await store.refreshGroups();
    const gid = (res.data as unknown as GroupConversationItem | undefined)?.id;
    leftTab.value = 'group';
    if (gid) openGroup(gid);
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '创建群聊失败');
  } finally {
    creatingGroup.value = false;
  }
}

// ---------------- 群设置 ----------------

async function openGroupSettings() {
  if (!currentGroup.value) return;
  renameValue.value = currentGroupInfo.value?.name || '';
  addMemberIds.value = [];
  showGroupSettings.value = true;
  await loadGroupMembers(currentGroup.value);
  // 群公告编辑初始值（未加载过则先拉取）
  if (!store.groupAnnouncements[currentGroup.value]) {
    await store.loadGroupAnnouncement(currentGroup.value);
  }
  announcementEdit.value = store.groupAnnouncements[currentGroup.value]?.content || '';
}

async function submitRename() {
  if (!currentGroup.value || !isGroupOwner.value) return;
  const name = renameValue.value.trim();
  if (!name) {
    toast.warning('群聊名称不能为空');
    return;
  }
  savingGroup.value = true;
  try {
    await groupChatService.renameGroup(currentGroup.value, name);
    toast.success('群聊名称已更新');
    await store.refreshGroups();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '重命名失败');
  } finally {
    savingGroup.value = false;
  }
}

async function submitAddMembers() {
  if (!currentGroup.value) return;
  if (addMemberIds.value.length === 0) {
    toast.warning('请选择要添加的成员');
    return;
  }
  savingGroup.value = true;
  try {
    const res = await groupChatService.addGroupMembers(currentGroup.value, addMemberIds.value);
    const added = (res.data as unknown as { added?: number })?.added ?? addMemberIds.value.length;
    toast.success(`已添加 ${added} 名成员`);
    addMemberIds.value = [];
    await loadGroupMembers(currentGroup.value);
    await store.refreshGroups();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '添加成员失败');
  } finally {
    savingGroup.value = false;
  }
}

async function removeGroupMemberAct(member: GroupMemberItem) {
  if (!currentGroup.value || !member.user) return;
  const uname = member.user.name || member.user.username || member.user_id.slice(0, 6);
  if (!confirm(`确定将「${uname}」移出群聊？`)) return;
  try {
    await groupChatService.removeGroupMember(currentGroup.value, member.user_id);
    toast.success('已移出群聊');
    await loadGroupMembers(currentGroup.value);
    await store.refreshGroups();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '移除成员失败');
  }
}

async function dissolveGroupAct() {
  if (!currentGroup.value) return;
  if (!confirm('解散后所有成员将被移出且消息不可恢复，确定解散该群聊？')) return;
  savingGroup.value = true;
  try {
    await groupChatService.dissolveGroup(currentGroup.value);
    toast.success('群聊已解散');
    showGroupSettings.value = false;
    store.setViewingGroup(null);
    delete store.groupMessages[currentGroup.value];
    currentGroup.value = null;
    await store.refreshGroups();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '解散群聊失败');
  } finally {
    savingGroup.value = false;
  }
}

async function quitGroupAct() {
  if (!currentGroup.value) return;
  if (!confirm('退出后将不再接收该群消息，确定退出群聊？')) return;
  savingGroup.value = true;
  try {
    await groupChatService.quitGroup(currentGroup.value);
    toast.success('已退出群聊');
    showGroupSettings.value = false;
    store.setViewingGroup(null);
    delete store.groupMessages[currentGroup.value];
    currentGroup.value = null;
    await store.refreshGroups();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '退出群聊失败');
  } finally {
    savingGroup.value = false;
  }
}

// ---------------- 群公告 ----------------

// 群主发布/更新群公告（内容为空视为清除）
async function saveAnnouncement() {
  if (!currentGroup.value || !isGroupOwner.value) return;
  const content = announcementEdit.value.trim();
  if (content.length > 500) {
    toast.warning('群公告最多 500 字');
    return;
  }
  savingAnnouncement.value = true;
  try {
    await groupChatService.setGroupAnnouncement(currentGroup.value, content);
    // 后端不推送给操作者本人，本地同步缓存
    store.groupAnnouncements[currentGroup.value] = {
      content,
      updated_at: content ? new Date().toISOString() : null,
    };
    toast.success(content ? '群公告已发布，成员将实时收到通知' : '群公告已清除');
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '设置群公告失败');
  } finally {
    savingAnnouncement.value = false;
  }
}

// 群主一键清除公告
async function clearAnnouncement() {
  if (!currentGroup.value || !currentAnnouncement.value?.content) return;
  announcementEdit.value = '';
  await saveAnnouncement();
}

// ---------------- 群文件夹 ----------------

function openFilesPanel() {
  if (!currentGroup.value) return;
  showFilesPanel.value = true;
  fileSearch.value = '';
  filePanelDragOver.value = false;
  store.loadGroupFiles(currentGroup.value);
}

// 上传群文件（按钮选择 + 拖拽共用）
async function uploadToGroupFiles(file: File) {
  if (!currentGroup.value || filePanelUploading.value) return;
  if (file.size > 10 * 1024 * 1024) {
    toast.warning('文件大小超过限制，最大允许 10MB');
    return;
  }
  filePanelUploading.value = true;
  try {
    await groupChatService.uploadGroupFile(currentGroup.value, file);
    toast.success(`「${file.name}」已上传到群文件`);
    // 后端不推送给上传者本人，手动刷新列表
    await store.loadGroupFiles(currentGroup.value);
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '文件上传失败');
  } finally {
    filePanelUploading.value = false;
  }
}

function onPanelFileSelected(e: Event) {
  const inputEl = e.target as HTMLInputElement;
  const file = inputEl.files?.[0];
  inputEl.value = '';
  if (file) uploadToGroupFiles(file);
}

// 拖拽上传：拖入高亮，松手取第一个文件上传
function onPanelDrop(e: DragEvent) {
  filePanelDragOver.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) uploadToGroupFiles(file);
}

// 下载群文件（带鉴权头拉取 blob；下载计数由后端自增并实时推送给其他成员）
async function downloadFileAct(f: GroupFileItem) {
  if (!currentGroup.value || downloadingId.value) return;
  downloadingId.value = f.id;
  try {
    await groupChatService.downloadGroupFile(currentGroup.value, f);
    // 本地同步 +1（后端不推送给下载者本人）
    f.download_count += 1;
  } catch (err) {
    // blob 响应中解析后端错误信息
    const resp = (err as { response?: { data?: Blob | { message?: string } } })?.response;
    let msg = '下载失败';
    if (resp?.data instanceof Blob) {
      try {
        msg = JSON.parse(await resp.data.text()).message || msg;
      } catch {
        /* ignore */
      }
    } else if (resp?.data?.message) {
      msg = resp.data.message;
    }
    toast.error(msg);
  } finally {
    downloadingId.value = '';
  }
}

// 删除群文件（上传者或群主）
async function deleteFileAct(f: GroupFileItem) {
  if (!currentGroup.value) return;
  if (!confirm(`确定删除文件「${f.file_name}」？删除后不可恢复。`)) return;
  try {
    await groupChatService.deleteGroupFile(currentGroup.value, f.id);
    toast.success('文件已删除');
    await store.loadGroupFiles(currentGroup.value);
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    toast.error(e2?.response?.data?.message || '删除文件失败');
  }
}

// 能否删除该文件：上传者本人或群主
function canDeleteFile(f: GroupFileItem): boolean {
  return isGroupOwner.value || f.uploader_id === auth.user?.id;
}

function isImageFile(f: GroupFileItem): boolean {
  return (f.mime_type || '').startsWith('image/');
}

function formatTime(ts: string): string {
  if (!ts) return '';
  const d = new Date(ts);
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  const hm = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  if (sameDay) return hm;
  const yesterday = new Date(now.getTime() - 86400000);
  if (d.toDateString() === yesterday.toDateString()) return `昨天 ${hm}`;
  return `${d.toLocaleDateString('zh-CN')} ${hm}`;
}

function formatSize(bytes?: number): string {
  if (!bytes) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function fileIcon(name?: string): string {
  const ext = (name || '').split('.').pop()?.toLowerCase() || '';
  if (['png', 'jpg', 'jpeg', 'gif', 'webp'].includes(ext)) return '🖼';
  if (['pdf'].includes(ext)) return '📕';
  if (['doc', 'docx'].includes(ext)) return '📘';
  if (['xls', 'xlsx'].includes(ext)) return '📗';
  if (['zip', 'rar', '7z'].includes(ext)) return '🗜';
  return '📄';
}

// 实时新消息滚动：在底部自动下拉；查看历史时只累加角标（已读标记由 store 在收到消息时即时处理）
watch(
  () => displayMessages.value,
  (list, oldList) => {
    if (!currentPeer.value && !currentGroup.value) return;
    // bug5：正在查看历史消息时不自动下拉，仅显示浮动按钮；在底部对话中才自动下拉
    if (!scrolledUp.value) {
      scrollToBottom();
    } else if (list && oldList && list.length > oldList.length) {
      // 上滑查看历史期间收到新消息 → 只累加角标，不打断阅读位置
      const last = list[list.length - 1];
      if (last && last.senderId !== auth.user?.id) {
        newMsgCount.value++;
      }
    }
  }
);

onMounted(() => {
  store.connectSocket();
  loadUsers();
  store.refreshConversations();
  store.refreshGroups();
  store.fetchOnlineUsers();
  // 需求24：弹窗点击跳转 /chat?peer=xxx 时直达对应会话
  const peer = route.query.peer as string | undefined;
  if (peer) openConversation(peer);
  // 群聊：弹窗点击跳转 /chat?group=xxx 时直达对应群聊
  const group = route.query.group as string | undefined;
  if (group) {
    leftTab.value = 'group';
    openGroup(group);
  }
  // 需求33：点击表情面板之外的区域关闭面板
  document.addEventListener('mousedown', handleOutsideClick);
});

// 需求33：点击页面其他地方关闭表情面板（面板与触发按钮内部不关闭）
function handleOutsideClick(e: MouseEvent) {
  if (!showEmoji.value) return;
  const t = e.target as Node;
  if (emojiPanelEl.value?.contains(t)) return;
  if (emojiBtnEl.value?.contains(t)) return;
  showEmoji.value = false;
}

// 需求24：已在聊天页时，点击弹窗切换会话（peer 查询参数变化）
watch(
  () => route.query.peer,
  (peer) => {
    if (peer && peer !== currentPeer.value) openConversation(peer as string);
  }
);

// 群聊：已在聊天页时，点击弹窗切换群聊（group 查询参数变化）
watch(
  () => route.query.group,
  (group) => {
    if (group && group !== currentGroup.value) {
      leftTab.value = 'group';
      openGroup(group as string);
    }
  }
);

// 群聊被解散（WebSocket dissolved 推送）：store 退出查看态，这里同步回退到空态视图
watch(
  () => store.viewingGroupId,
  (gid) => {
    if (gid === null && currentGroup.value) {
      currentGroup.value = null;
    }
  }
);

onUnmounted(() => {
  // 离开聊天页：清除正在查看的会话，后续消息正常计入未读角标
  store.setViewingPeer(null);
  store.setViewingGroup(null);
  // 需求33：移除全局点击监听
  document.removeEventListener('mousedown', handleOutsideClick);
});
</script>

<template>
  <div class="h-full flex rounded-card overflow-hidden border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800">
    <!-- ============ 左侧栏 ============ -->
    <aside class="w-[300px] shrink-0 flex flex-col border-r border-slate-100 dark:border-slate-700 bg-slate-50/60 dark:bg-slate-900/40">
      <!-- 头部 -->
      <div class="px-4 py-3.5 border-b border-slate-100 dark:border-slate-700">
        <div class="flex items-center justify-between mb-2.5">
          <h2 class="text-base font-semibold text-slate-800 dark:text-slate-100">💬 聊天</h2>
          <div class="flex items-center gap-1.5">
            <button
              v-if="leftTab === 'group'"
              class="w-6 h-6 rounded-lg bg-blue-500 text-white text-sm leading-none flex items-center justify-center hover:bg-blue-600 transition-smooth"
              title="创建群聊"
              @click="openCreateModal"
            >+</button>
            <span class="text-[10px] px-2 py-0.5 rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 font-medium">
              {{ store.connected ? '在线' : '连接中' }}
            </span>
          </div>
        </div>
        <input
          v-model="keyword"
          type="text"
          class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-400"
          :placeholder="leftTab === 'conv' ? '搜索会话...' : leftTab === 'group' ? '搜索群聊...' : '搜索姓名 / 部门...'"
        />
      </div>

      <!-- Tab 切换：会话 / 群聊 / 通讯录 -->
      <div class="flex border-b border-slate-100 dark:border-slate-700">
        <button
          class="flex-1 py-2.5 text-sm font-medium transition-smooth relative"
          :class="leftTab === 'conv' ? 'text-blue-500' : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="leftTab = 'conv'"
        >
          会话
          <span class="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full transition-smooth"
            :class="leftTab === 'conv' ? 'bg-blue-500' : 'bg-transparent'" />
        </button>
        <button
          class="flex-1 py-2.5 text-sm font-medium transition-smooth relative"
          :class="leftTab === 'group' ? 'text-blue-500' : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="leftTab = 'group'"
        >
          群聊
          <span class="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full transition-smooth"
            :class="leftTab === 'group' ? 'bg-blue-500' : 'bg-transparent'" />
        </button>
        <button
          class="flex-1 py-2.5 text-sm font-medium transition-smooth relative"
          :class="leftTab === 'contact' ? 'text-blue-500' : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300'"
          @click="leftTab = 'contact'"
        >
          通讯录
          <span class="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 rounded-full transition-smooth"
            :class="leftTab === 'contact' ? 'bg-blue-500' : 'bg-transparent'" />
        </button>
      </div>

      <!-- 会话列表 -->
      <div v-if="leftTab === 'conv'" class="flex-1 overflow-y-auto scrollbar-thin">
        <div v-if="store.conversations.length === 0" class="px-4 py-10 text-center">
          <p class="text-3xl mb-2">💬</p>
          <p class="text-xs text-slate-400 dark:text-slate-500">暂无会话</p>
          <p class="text-[10px] text-slate-300 dark:text-slate-600 mt-1">切换到「通讯录」选择一个好友开始聊天</p>
        </div>
        <div
          v-for="conv in store.conversations"
          :key="conv.peer_id"
          class="px-3 py-2.5 flex items-center gap-2.5 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/40 transition-smooth"
          :class="{ 'bg-blue-50 dark:bg-blue-900/10': currentPeer === conv.peer_id }"
          @click="openConversation(conv.peer_id, conv.peer_name)"
        >
          <div class="relative w-10 h-10 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ (conv.peer_name || peerName(conv.peer_id)).slice(0, 1) }}
            <span
              v-if="store.isOnline(conv.peer_id)"
              class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full bg-green-500 border-2 border-white dark:border-slate-800"
            ></span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">{{ conv.peer_name || peerName(conv.peer_id) }}</p>
              <span class="text-[10px] text-slate-400 shrink-0">{{ formatTime(conv.last_at) }}</span>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate mt-0.5">{{ conv.last_msg }}</p>
          </div>
          <span v-if="conv.unread > 0" class="shrink-0 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold leading-[18px] text-center">
            {{ conv.unread > 99 ? '99+' : conv.unread }}
          </span>
        </div>
      </div>

      <!-- 群聊列表 -->
      <div v-else-if="leftTab === 'group'" class="flex-1 overflow-y-auto scrollbar-thin">
        <div v-if="filteredGroups.length === 0" class="px-4 py-10 text-center">
          <p class="text-3xl mb-2">👥</p>
          <p class="text-xs text-slate-400 dark:text-slate-500">{{ keyword ? '未找到匹配的群聊' : '暂无群聊' }}</p>
          <button
            v-if="!keyword"
            class="mt-3 text-xs px-3 py-1.5 rounded-btn bg-blue-500 text-white hover:bg-blue-600 transition-smooth"
            @click="openCreateModal"
          >+ 创建群聊</button>
        </div>
        <div
          v-for="g in filteredGroups"
          :key="g.id"
          class="px-3 py-2.5 flex items-center gap-2.5 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/40 transition-smooth"
          :class="{ 'bg-blue-50 dark:bg-blue-900/10': currentGroup === g.id }"
          @click="openGroup(g.id)"
        >
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-400 to-purple-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ (g.name || '群').slice(0, 1) }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">{{ g.name }}</p>
              <span class="text-[10px] text-slate-400 shrink-0">{{ formatTime(g.last_at || '') }}</span>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate mt-0.5">
              <span v-if="store.groupMentionedMe[g.id]" class="text-red-500 font-medium">[有人@我] </span>{{ g.last_msg || `${g.member_count} 名成员` }}
            </p>
          </div>
          <span v-if="g.unread > 0" class="shrink-0 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold leading-[18px] text-center">
            {{ g.unread > 99 ? '99+' : g.unread }}
          </span>
        </div>
      </div>

      <!-- 通讯录 -->
      <div v-else class="flex-1 overflow-y-auto scrollbar-thin">
        <div v-if="filteredUsers.length === 0" class="px-4 py-10 text-center text-xs text-slate-400 dark:text-slate-500">
          {{ keyword ? '未找到匹配的好友' : '暂无其他用户' }}
        </div>
        <div
          v-for="u in filteredUsers"
          :key="u.id"
          class="px-3 py-2.5 flex items-center gap-2.5 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/40 transition-smooth"
          @click="openConversation(u.id, u.name || u.username)"
        >
          <div class="relative w-10 h-10 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
            {{ (u.name || u.username).slice(0, 1) }}
            <span
              class="absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-slate-800"
              :class="store.isOnline(u.id) ? 'bg-green-500' : 'bg-slate-300 dark:bg-slate-600'"
            ></span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between gap-2">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">{{ u.name || u.username }}</p>
              <span class="text-[10px] shrink-0" :class="store.isOnline(u.id) ? 'text-green-500' : 'text-slate-400 dark:text-slate-500'">
                {{ store.isOnline(u.id) ? '在线' : '离线' }}
              </span>
            </div>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ u.dept_name || u.role }}</p>
          </div>
        </div>
      </div>
    </aside>

    <!-- ============ 主区 ============ -->
    <main class="flex-1 flex flex-col min-w-0 bg-slate-50 dark:bg-slate-900/50">
      <!-- 空态 -->
      <div v-if="!currentPeer && !currentGroup" class="flex-1 flex flex-col items-center justify-center">
        <div class="w-16 h-16 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center text-3xl mb-3">💬</div>
        <p class="text-sm text-slate-400 dark:text-slate-500">选择一个好友或群聊开始聊天</p>
        <p class="text-xs text-slate-300 dark:text-slate-600 mt-1">支持文本、emoji 表情、图片与文件传输</p>
      </div>

      <template v-else>
        <!-- 顶部信息条 -->
        <div class="h-14 shrink-0 flex items-center gap-3 px-5 border-b border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-800">
          <!-- 群聊信息条 -->
          <template v-if="activeIsGroup">
            <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-violet-400 to-purple-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
              {{ (currentGroupInfo?.name || '群').slice(0, 1) }}
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ currentGroupInfo?.name || '群聊' }}</p>
              <p class="text-[10px] text-slate-400">{{ currentGroupInfo?.member_count ?? groupMembers.length }} 名成员</p>
            </div>
            <button
              class="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-500 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              title="群文件：成员可上传 / 下载"
              @click="openFilesPanel"
            >📁 文件</button>
            <button
              class="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-500 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              @click="openGroupSettings"
            >⚙ 群设置</button>
          </template>
          <!-- 私聊信息条 -->
          <template v-else>
            <div class="w-9 h-9 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 text-white text-sm font-medium flex items-center justify-center shrink-0">
              {{ currentPeerName.slice(0, 1) }}
            </div>
            <div class="min-w-0">
              <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ currentPeerName }}</p>
              <p class="text-[10px] flex items-center gap-1" :class="store.isOnline(currentPeer!) ? 'text-green-500' : 'text-slate-400'">
                <span class="inline-block w-1.5 h-1.5 rounded-full" :class="store.isOnline(currentPeer!) ? 'bg-green-500' : 'bg-slate-400'"></span>
                {{ store.isOnline(currentPeer!) ? '在线' : '离线' }}
              </p>
            </div>
          </template>
        </div>

        <!-- 群公告横幅：默认一行截断，点击展开全文 -->
        <div v-if="activeIsGroup && currentAnnouncement?.content" class="shrink-0 px-5 py-2 border-b border-amber-200/60 dark:border-amber-800/40 bg-gradient-to-r from-amber-50 via-orange-50 to-amber-50 dark:from-amber-900/20 dark:via-orange-900/10 dark:to-amber-900/20">
          <div
            class="flex items-start gap-2 cursor-pointer select-none"
            @click="announcementExpanded = !announcementExpanded"
          >
            <span class="text-sm leading-5 mt-0.5 shrink-0">📢</span>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-amber-200/70 dark:bg-amber-800/50 text-amber-700 dark:text-amber-300 shrink-0">群公告</span>
                <p
                  class="text-xs text-amber-800 dark:text-amber-200 flex-1 min-w-0"
                  :class="announcementExpanded ? 'whitespace-pre-wrap break-words' : 'truncate'"
                >{{ currentAnnouncement.content }}</p>
                <svg
                  class="w-3.5 h-3.5 text-amber-500 shrink-0 transition-transform duration-200"
                  :class="announcementExpanded ? 'rotate-180' : ''"
                  fill="none" viewBox="0 0 24 24" stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
              <p v-if="announcementExpanded && currentAnnouncement.updated_at" class="text-[10px] text-amber-500 dark:text-amber-400/70 mt-1">
                群主更新于 {{ formatTime(currentAnnouncement.updated_at) }}
              </p>
            </div>
          </div>
        </div>

        <!-- 消息区 -->
        <div class="relative flex-1 flex flex-col min-h-0">
          <div
            ref="messagesEl"
            class="flex-1 overflow-y-auto scrollbar-thin px-6 py-4 space-y-3"
            @scroll.passive="handleMessagesScroll"
          >
            <div v-if="loadingOlder" class="text-center text-xs text-slate-400 py-1">加载更早消息...</div>
            <div v-if="allLoaded && displayMessages.length > 0" class="text-center text-[10px] text-slate-300 dark:text-slate-600 py-1">—— 已显示全部消息 ——</div>

            <div
              v-for="m in displayMessages"
              :key="m.id"
              class="flex items-end gap-2"
              :class="m.senderId === auth.user?.id ? 'justify-end' : 'justify-start'"
            >
            <!-- 对方头像 -->
            <div v-if="m.senderId !== auth.user?.id" class="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 text-white text-xs font-medium flex items-center justify-center shrink-0 mb-1">
              {{ m.senderName.slice(0, 1) }}
            </div>

            <div class="max-w-[70%] min-w-0">
              <!-- 群聊显示发送者名 -->
              <p v-if="activeIsGroup && m.senderId !== auth.user?.id" class="text-[10px] text-slate-400 dark:text-slate-500 mb-0.5 px-1">{{ m.senderName }}</p>

              <!-- 文本 -->
              <div
                v-if="m.type === 'text'"
                class="px-3.5 py-2 rounded-2xl text-sm leading-relaxed break-words rich-content-display"
                :class="m.senderId === auth.user?.id
                  ? 'bg-blue-500 text-white rounded-br-sm'
                  : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 rounded-bl-sm'"
                v-html="renderGroupBubble(m.content)"
              ></div>

              <!-- 图片 -->
              <div v-else-if="m.type === 'image'" class="min-w-[120px]">
                <img
                  :src="m.filePath"
                  :alt="m.fileName || '图片'"
                  class="max-w-[240px] max-h-[300px] rounded-xl cursor-pointer object-cover shadow-sm border border-slate-200 dark:border-slate-600"
                  loading="lazy"
                  @click="previewUrl = m.filePath || ''"
                />
              </div>

              <!-- 文件 -->
              <div v-else-if="m.type === 'file'" class="min-w-[200px]">
                <a
                  :href="m.filePath"
                  download
                  target="_blank"
                  rel="noopener"
                  class="flex items-center gap-3 px-3.5 py-2.5 rounded-2xl bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 hover:shadow-md transition-smooth"
                  :class="m.senderId === auth.user?.id ? 'rounded-br-sm' : 'rounded-bl-sm'"
                >
                  <span class="text-2xl shrink-0">{{ fileIcon(m.fileName) }}</span>
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-slate-800 dark:text-slate-100 truncate max-w-[180px]">{{ m.fileName }}</p>
                    <p class="text-[10px] text-slate-400">{{ formatSize(m.fileSize) }} · 点击下载</p>
                  </div>
                </a>
              </div>

              <p class="mt-1 text-[10px] flex items-center gap-1.5 px-1" :class="m.senderId === auth.user?.id ? 'justify-end' : ''">
                <span class="text-slate-400">{{ formatTime(m.createdAt) }}</span>
                <span
                  v-if="!activeIsGroup && m.senderId === auth.user?.id"
                  :class="m.isRead ? 'text-blue-500 dark:text-blue-400' : 'text-slate-400 dark:text-slate-500'"
                >{{ m.isRead ? '已读' : '未读' }}</span>
              </p>
            </div>
          </div>
        </div>
          <!-- 浮动按钮：上滑查看历史消息时显示，点击回到最新消息 -->
          <button
            v-if="scrolledUp"
            class="absolute bottom-4 right-4 z-10 flex items-center gap-1.5 px-3 py-2 rounded-full bg-blue-500 text-white text-xs font-medium shadow-lg hover:bg-blue-600 active:scale-95 transition-smooth"
            @click="goToLatest"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7-7-7m14-8l-7 7-7-7" />
            </svg>
            <span v-if="newMsgCount > 0" class="min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center">{{ newMsgCount }}</span>
            <span>最新消息</span>
          </button>
        </div>

        <!-- 输入区 -->
        <div class="relative shrink-0 border-t border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-800 p-3">
          <!-- 工具栏 -->
          <div class="flex items-center gap-1 mb-2">
            <button
              ref="emojiBtnEl"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              title="表情"
              @click="showEmoji = !showEmoji"
            >😊</button>
            <button
              class="w-8 h-8 rounded-lg flex items-center justify-center text-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
              title="发送文件（支持绝大部分格式，禁止可执行/脚本文件）"
              :disabled="uploading"
              @click="fileInput?.click()"
            >📎</button>
            <span v-if="uploading" class="text-xs text-slate-400 ml-1">上传中...</span>
            <input ref="fileInput" type="file" class="hidden" @change="onFileSelected" />
          </div>

          <!-- emoji 面板 -->
          <transition name="shrink-out">
          <div
            v-if="showEmoji"
            ref="emojiPanelEl"
            class="absolute bottom-full left-3 mb-2 z-20"
            @click.stop
          >
            <EmojiPicker @select="insertEmoji" @send-image="sendEmoticon" />
          </div>
          </transition>

          <!-- @ 成员选择浮层（微信交互：输入 @ 弹出，Enter 选中） -->
          <transition name="shrink-out">
          <div
            v-if="mentionToken"
            ref="mentionPanelEl"
            class="absolute bottom-full left-3 mb-2 z-20"
            @click.stop
          >
            <MentionPicker
              ref="mentionPickerEl"
              :members="groupMembers"
              :keyword="mentionToken.keyword"
              :show-all="isGroupOwner"
              @select="onMentionSelect"
            />
          </div>
          </transition>

          <div class="flex items-end gap-2">
            <div class="relative flex-1 min-w-0">
              <textarea
                ref="inputEl"
                v-model="input"
                rows="2"
                class="relative w-full resize-none rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3.5 py-2 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 scrollbar-thin"
                :placeholder="activeIsGroup ? `发消息到「${currentGroupInfo?.name || '群聊'}」，Enter 发送，@ 提醒成员` : '输入消息，Enter 发送，Shift+Enter 换行'"
                @input="refreshMentionToken"
                @click="refreshMentionToken"
                @keyup="onMentionKeyup"
                @scroll.passive="syncMirrorScroll"
                @keydown.enter.exact.prevent="onEnterKeydown"
                @keydown.up="onMentionNavKey($event, 'up')"
                @keydown.down="onMentionNavKey($event, 'down')"
                @keydown.esc="mentionToken = null"
                @focus="showEmoji = false"
              />
              <!-- @ 提及镜像高亮层：与输入框同排版（文字透明），为 @名字 垫蓝色底衬 -->
              <div
                v-if="activeIsGroup && mirrorSegments.length"
                ref="mirrorEl"
                aria-hidden="true"
                class="mention-mirror absolute inset-0 rounded-xl border border-transparent px-3.5 py-2 text-sm leading-relaxed break-words whitespace-pre-wrap overflow-hidden pointer-events-none"
              ><template v-for="(seg, i) in mirrorSegments" :key="i"><span v-if="seg.mentioned" class="mention-flag">{{ seg.text }}</span><template v-else>{{ seg.text }}</template></template></div>
            </div>
            <button
              class="shrink-0 px-5 py-2 rounded-xl bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 transition-smooth disabled:opacity-50"
              :disabled="!input.trim()"
              @click="sendText"
            >
              发送
            </button>
          </div>
        </div>
      </template>
    </main>

    <!-- 创建群聊弹窗 -->
    <Teleport to="body">
      <transition name="shrink-out">
      <div v-if="showCreateGroup" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showCreateGroup = false" />
        <div class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-md mx-4 p-6 animate-fade-in">
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">创建群聊</h3>
          <div class="space-y-4">
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">群聊名称</span>
              <input v-model="newGroupName" class="input-field" placeholder="请输入群聊名称（最多 50 字）" maxlength="50" autofocus @keydown.enter.prevent />
            </div>
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">添加成员（可按部门一键全选）</span>
              <UserPicker v-model="newGroupMembers" :max="200" />
            </div>
          </div>
          <div class="flex justify-end gap-2 mt-6">
            <button class="btn-secondary" @click="showCreateGroup = false">取消</button>
            <button
              class="btn-primary"
              :disabled="creatingGroup || !newGroupName.trim() || newGroupMembers.length === 0"
              @click="submitCreateGroup"
            >{{ creatingGroup ? '创建中...' : '创建群聊' }}</button>
          </div>
        </div>
      </div>
      </transition>
    </Teleport>

    <!-- 群设置弹窗 -->
    <Teleport to="body">
      <transition name="shrink-out">
      <div v-if="showGroupSettings" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showGroupSettings = false" />
        <div class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-md mx-4 p-6 animate-fade-in flex flex-col max-h-[80vh]">
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4 shrink-0">群设置</h3>

          <!-- 群名 / 重命名 -->
          <div class="shrink-0 mb-4">
            <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">群聊名称{{ isGroupOwner ? '' : '（仅群主可修改）' }}</span>
            <div class="flex gap-2">
              <input v-model="renameValue" class="input-field flex-1" :disabled="!isGroupOwner" maxlength="50" @keydown.enter.prevent="submitRename" />
              <button
                v-if="isGroupOwner"
                class="btn-secondary shrink-0"
                :disabled="savingGroup || renameValue.trim() === (currentGroupInfo?.name || '')"
                @click="submitRename"
              >保存</button>
            </div>
          </div>

          <!-- 群公告（群主可编辑，实时推送给成员） -->
          <div class="shrink-0 mb-4">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs text-slate-500 dark:text-slate-400">群公告{{ isGroupOwner ? '' : '（仅群主可设置）' }}</span>
              <span class="text-[10px] text-slate-300 dark:text-slate-600">{{ announcementEdit.length }}/500</span>
            </div>
            <textarea
              v-model="announcementEdit"
              rows="3"
              :disabled="!isGroupOwner"
              maxlength="500"
              class="input-field resize-none text-xs"
              :placeholder="isGroupOwner ? '写下群公告，保存后所有成员将实时收到通知...' : '群主还没有发布公告'"
            ></textarea>
            <div v-if="isGroupOwner" class="flex justify-end gap-2 mt-1.5">
              <button
                class="text-xs px-3 py-1.5 rounded-btn text-slate-500 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth disabled:opacity-40 disabled:cursor-not-allowed"
                :disabled="savingAnnouncement || !currentAnnouncement?.content"
                @click="clearAnnouncement"
              >清除</button>
              <button
                class="btn-primary !px-4 !py-1.5 !text-xs"
                :disabled="savingAnnouncement || announcementEdit.trim() === (currentAnnouncement?.content || '')"
                @click="saveAnnouncement"
              >{{ savingAnnouncement ? '保存中...' : '发布' }}</button>
            </div>
          </div>

          <!-- 成员列表 -->
          <div class="shrink-0 mb-1.5 flex items-center justify-between">
            <span class="text-xs text-slate-500 dark:text-slate-400">群成员（{{ groupMembers.length }}）</span>
            <span class="text-[10px] text-slate-300 dark:text-slate-600">按加入时间排序</span>
          </div>
          <div class="flex-1 min-h-0 overflow-y-auto scrollbar-thin border border-slate-100 dark:border-slate-700 rounded-xl p-2 mb-4 space-y-0.5">
            <div v-if="membersLoading" class="text-center text-xs text-slate-400 py-4">加载中...</div>
            <template v-else>
              <div
                v-for="mem in groupMembers"
                :key="mem.id"
                class="flex items-center gap-2.5 px-2 py-1.5 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-smooth"
              >
                <div class="w-7 h-7 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 text-white text-[10px] font-medium flex items-center justify-center shrink-0">
                  {{ (mem.user?.name || mem.user?.username || '？').slice(0, 1) }}
                </div>
                <span class="text-sm text-slate-700 dark:text-slate-200 truncate flex-1">
                  {{ mem.user?.name || mem.user?.username || mem.user_id.slice(0, 6) }}
                  <span v-if="mem.user_id === auth.user?.id" class="text-[10px] text-slate-400">（我）</span>
                </span>
                <span v-if="mem.role === 'owner'" class="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-400 shrink-0">群主</span>
                <button
                  v-else-if="isGroupOwner"
                  class="text-[10px] px-1.5 py-0.5 rounded text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-smooth shrink-0"
                  @click="removeGroupMemberAct(mem)"
                >移出</button>
              </div>
            </template>
          </div>

          <!-- 群主：添加成员 -->
          <div v-if="isGroupOwner" class="shrink-0 mb-4">
            <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">添加新成员</span>
            <div class="flex items-start gap-2">
              <div class="flex-1 min-w-0">
                <UserPicker v-model="addMemberIds" :max="200" :disabled-ids="existingMemberIds" disabled-note="已在群中" />
              </div>
              <button class="btn-secondary shrink-0" :disabled="savingGroup || addMemberIds.length === 0" @click="submitAddMembers">添加</button>
            </div>
          </div>

          <!-- 底部操作 -->
          <div class="shrink-0 flex justify-between">
            <button
              v-if="isGroupOwner"
              class="text-xs px-3 py-2 rounded-btn text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-smooth"
              :disabled="savingGroup"
              @click="dissolveGroupAct"
            >解散群聊</button>
            <button
              v-else
              class="text-xs px-3 py-2 rounded-btn text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-smooth"
              :disabled="savingGroup"
              @click="quitGroupAct"
            >退出群聊</button>
            <button class="btn-secondary" @click="showGroupSettings = false">关闭</button>
          </div>
        </div>
      </div>
      </transition>
    </Teleport>

    <!-- 群文件侧滑面板 -->
    <Teleport to="body">
      <transition name="slide">
      <div v-if="showFilesPanel" class="fixed inset-0 z-50">
        <div class="overlay-backdrop" @click="showFilesPanel = false" />
        <aside class="absolute top-0 right-0 z-50 h-full w-[380px] max-w-[92vw] bg-white dark:bg-slate-800 shadow-modal flex flex-col border-l border-slate-200 dark:border-slate-700">
          <!-- 头部 -->
          <div class="shrink-0 px-4 py-3.5 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between">
            <div>
              <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">📁 群文件</h3>
              <p class="text-[10px] text-slate-400 mt-0.5">共 {{ currentGroupFiles.length }} 个文件 · 拖入文件即可上传</p>
            </div>
            <button class="w-7 h-7 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth flex items-center justify-center" @click="showFilesPanel = false">✕</button>
          </div>

          <!-- 搜索 + 上传 -->
          <div class="shrink-0 px-4 py-2.5 flex items-center gap-2 border-b border-slate-100 dark:border-slate-700">
            <input
              v-model="fileSearch"
              type="text"
              class="flex-1 px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-blue-400"
              placeholder="搜索文件名 / 上传者..."
            />
            <button
              class="shrink-0 px-3 py-1.5 rounded-lg bg-blue-500 text-white text-xs font-medium hover:bg-blue-600 transition-smooth disabled:opacity-50"
              :disabled="filePanelUploading"
              @click="filePanelInput?.click()"
            >{{ filePanelUploading ? '上传中...' : '＋ 上传' }}</button>
            <input ref="filePanelInput" type="file" class="hidden" @change="onPanelFileSelected" />
          </div>

          <!-- 文件列表（整区接受拖拽上传） -->
          <div
            class="flex-1 min-h-0 overflow-y-auto scrollbar-thin relative transition-colors"
            :class="filePanelDragOver ? 'bg-blue-50/70 dark:bg-blue-900/20' : ''"
            @dragover.prevent="filePanelDragOver = true"
            @dragleave.self="filePanelDragOver = false"
            @drop.prevent="onPanelDrop"
          >
            <!-- 拖拽提示 -->
            <div
              v-if="filePanelDragOver"
              class="absolute inset-2 z-10 rounded-xl border-2 border-dashed border-blue-400 bg-blue-50/80 dark:bg-blue-900/40 flex flex-col items-center justify-center pointer-events-none"
            >
              <span class="text-3xl mb-1">📂</span>
              <p class="text-xs text-blue-500 font-medium">松手上传到群文件</p>
            </div>

            <div v-if="currentGroupFiles.length === 0" class="px-6 py-14 text-center">
              <p class="text-4xl mb-2">🗂</p>
              <p class="text-xs text-slate-400 dark:text-slate-500">{{ fileSearch ? '未找到匹配的文件' : '群文件空空如也' }}</p>
              <p class="text-[10px] text-slate-300 dark:text-slate-600 mt-1">上传文件后，所有成员都可以在此下载</p>
            </div>

            <div
              v-for="f in currentGroupFiles"
              :key="f.id"
              class="flex items-center gap-2.5 px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-smooth group"
            >
              <!-- 缩略图 / 类型图标（图片可点击放大预览） -->
              <div class="w-10 h-10 rounded-lg overflow-hidden shrink-0 bg-slate-100 dark:bg-slate-700 flex items-center justify-center">
                <img
                  v-if="isImageFile(f)"
                  :src="f.file_path"
                  :alt="f.file_name"
                  class="w-full h-full object-cover cursor-pointer"
                  loading="lazy"
                  @click="previewUrl = f.file_path"
                />
                <span v-else class="text-xl">{{ fileIcon(f.file_name) }}</span>
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-xs font-medium text-slate-700 dark:text-slate-200 truncate" :title="f.file_name">{{ f.file_name }}</p>
                <p class="text-[10px] text-slate-400 truncate">
                  {{ formatSize(f.file_size) }} · {{ f.uploader_name || '群成员' }} · {{ formatTime(f.created_at) }}
                </p>
                <p class="text-[10px] text-slate-300 dark:text-slate-500 mt-0.5">📥 已被下载 {{ f.download_count }} 次</p>
              </div>
              <div class="flex items-center gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  class="w-7 h-7 rounded-lg text-sm flex items-center justify-center text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition-smooth disabled:opacity-40"
                  title="下载"
                  :disabled="downloadingId === f.id"
                  @click="downloadFileAct(f)"
                >{{ downloadingId === f.id ? '⏳' : '⬇' }}</button>
                <button
                  v-if="canDeleteFile(f)"
                  class="w-7 h-7 rounded-lg text-sm flex items-center justify-center text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-smooth"
                  title="删除（上传者或群主）"
                  @click="deleteFileAct(f)"
                >🗑</button>
              </div>
            </div>
          </div>
        </aside>
      </div>
      </transition>
    </Teleport>

    <!-- 图片预览 -->
    <Teleport to="body">
      <transition name="shrink-out">
      <div v-if="previewUrl" class="fixed inset-0 z-[90] flex items-center justify-center bg-black/70 p-6" @click="previewUrl = ''">
        <img :src="previewUrl" class="max-w-[90vw] max-h-[90vh] rounded-lg object-contain" />
        <button class="absolute top-4 right-4 w-10 h-10 rounded-full bg-black/40 text-white text-xl flex items-center justify-center hover:bg-black/60 transition-smooth">✕</button>
      </div>
      </transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* 群文件面板滑入/滑出：遮罩淡入淡出 + 面板右侧平移 */
.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.2s ease;
}
.slide-enter-active aside,
.slide-leave-active aside {
  transition: transform 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
}
.slide-enter-from aside,
.slide-leave-to aside {
  transform: translateX(100%);
}

/* ===== @ 提及 ===== */
/* 输入框镜像高亮层：文字透明，仅为「@名字」提供蓝底衬（不干扰 IME 与选区） */
.mention-mirror {
  color: transparent;
  z-index: 1;
}
.mention-mirror .mention-flag {
  background: rgba(59, 130, 246, 0.2);
  border-radius: 4px;
}
/* 消息气泡中的「@名字」高亮（微信蓝） */
.rich-content-display :deep(.mention-tag) {
  color: #2563eb;
  background: rgba(59, 130, 246, 0.12);
  border-radius: 4px;
  padding: 0 2px;
}
/* 自己的气泡（蓝底）中的「@名字」：白字 + 半透明白衬 */
.mention-own :deep(.mention-tag) {
  color: #fff;
  background: rgba(255, 255, 255, 0.24);
}
</style>
