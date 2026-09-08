<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { useNotificationStore } from '@/stores/notification';
import { useAuthStore } from '@/stores/auth';
import { getVisibleUsers } from '@/services/admin';
import { uploadChatFile } from '@/services/notification';
import * as groupChatService from '@/services/groupChat';
import EmojiPicker from '@/components/chat/EmojiPicker.vue';
import UserPicker from '@/components/common/UserPicker.vue';
import type { ChatMessageItem, GroupConversationItem, GroupMessageItem, GroupMemberItem, User } from '@/types';
import { renderNoteContent } from '@/utils/richText';
import { matchPinyin } from '@/utils/pinyin';
import { useToast } from '@/composables/useToast';

const store = useNotificationStore();
const auth = useAuthStore();
const route = useRoute();
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
  try {
    await store.loadGroupMessages(groupId);
    await store.markGroupRead(groupId);
  } catch {
    /* ignore */
  }
  scrollToBottom();
  // 群成员用于发送者名解析与群设置弹窗，后台加载不阻塞消息展示
  loadGroupMembers(groupId);
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
  input.value = '';
  showEmoji.value = false;
  if (activeIsGroup.value && currentGroup.value) {
    await store.sendGroupMessage(currentGroup.value, { content });
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
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate mt-0.5">{{ g.last_msg || `${g.member_count} 名成员` }}</p>
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
                v-html="renderNoteContent(m.content)"
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

          <div class="flex items-end gap-2">
            <textarea
              v-model="input"
              rows="2"
              class="flex-1 resize-none rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3.5 py-2 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/50 scrollbar-thin"
              :placeholder="activeIsGroup ? `发消息到「${currentGroupInfo?.name || '群聊'}」，Enter 发送` : '输入消息，Enter 发送，Shift+Enter 换行'"
              @keydown.enter.exact.prevent="sendText"
              @focus="showEmoji = false"
            />
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
