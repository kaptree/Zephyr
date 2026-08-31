<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { getUsers } from '@/services/admin';
import { inspectUserWorkbench } from '@/services/notes';
import type { UserBrief } from '@/types';
import { matchPinyin } from '@/utils/pinyin';

const users = ref<UserBrief[]>([]);
const searchText = ref('');
const selectedUser = ref<UserBrief | null>(null);
const userListOpen = ref(false);
const notes = ref<any[]>([]);
const status = ref<'active' | 'completed' | 'archived'>('active');
const loading = ref(false);
const error = ref('');
const total = ref(0);

const filteredUsers = computed(() => {
  const q = searchText.value.trim();
  if (!q) return users.value;
  // 需求36：支持拼音全拼 / 首字母搜索（无视大小写）
  return users.value.filter((u) => matchPinyin(q, u.name, u.dept_name));
});

async function loadUsers() {
  try {
    const res = await getUsers({ page: 1, page_size: 100 });
    const raw = (res.data as unknown as { data: any[] }).data || [];
    users.value = raw.map((u: any) => ({
      id: u.id,
      name: u.name || '',
      avatar: u.avatar || '',
      dept_id: u.dept_id || u.department?.id || '',
      dept_name: u.department?.name || u.dept_name || '',
      role: u.role || 'user',
    })) as UserBrief[];
  } catch {
    /* ignore */
  }
}

function pickUser(u: UserBrief) {
  selectedUser.value = u;
  userListOpen.value = false;
  searchText.value = '';
}

function sourceLabel(note: any): string {
  if ((note.ccs || []).some((c: any) => c.user_id === selectedUser.value?.id)) return '任务抄送';
  if (note.source_type === 'assigned') {
    return note.creator_id === selectedUser.value?.id ? '自己指派' : '上级指派';
  }
  if (note.source_type === 'collaboration') return '协同任务';
  return '自己创建';
}

function sourceChipClass(label: string): string {
  switch (label) {
    case '任务抄送':
      return 'bg-purple-100 text-purple-700 dark:bg-purple-900/50 dark:text-purple-300';
    case '自己指派':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300';
    case '上级指派':
      return 'bg-orange-100 text-orange-700 dark:bg-orange-900/50 dark:text-orange-300';
    case '协同任务':
      return 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/50 dark:text-cyan-300';
    default:
      return 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300';
  }
}

async function loadNotes() {
  if (!selectedUser.value) {
    notes.value = [];
    total.value = 0;
    return;
  }
  loading.value = true;
  error.value = '';
  try {
    const res = await inspectUserWorkbench(selectedUser.value.id, status.value);
    const d = res.data as unknown as { data: any[]; total: number };
    notes.value = d.data || [];
    total.value = d.total || 0;
  } catch (e: any) {
    error.value = e?.response?.data?.message || '加载失败';
    notes.value = [];
  } finally {
    loading.value = false;
  }
}

watch(selectedUser, loadNotes);
watch(status, loadNotes);

onMounted(loadUsers);
</script>

<template>
  <div class="space-y-5">
    <!-- 顶部：用户选择 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">用户工作台</h1>
        <p class="text-xs text-slate-400 mt-0.5">选择一位用户，查看其工作台任务内容</p>
      </div>
    </div>

    <!-- 用户选择器 -->
    <div class="relative max-w-xl">
      <div class="flex items-center gap-2">
        <div class="relative flex-1">
          <input
            v-model="searchText"
            class="input-field"
            placeholder="搜索用户姓名 / 部门..."
            @focus="userListOpen = true"
            @keydown.enter.prevent="userListOpen = !userListOpen"
          />
          <div
            v-if="userListOpen"
            class="absolute left-0 right-0 top-full mt-1 bg-white dark:bg-slate-800 rounded-card shadow-modal border border-slate-100 dark:border-slate-700 z-50 overflow-hidden"
          >
            <div class="max-h-72 overflow-y-auto scrollbar-thin p-1.5">
              <button
                v-for="u in filteredUsers"
                :key="u.id"
                type="button"
                class="w-full flex items-center gap-3 px-3 py-2 rounded-btn text-sm text-left transition-smooth hover:bg-slate-50 dark:hover:bg-slate-700"
                :class="selectedUser?.id === u.id ? 'bg-blue-50 dark:bg-blue-900/40' : ''"
                @click="pickUser(u)"
              >
                <div
                  class="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-600 flex items-center justify-center text-xs font-medium text-slate-600 dark:text-slate-300 shrink-0"
                >
                  {{ u.name.charAt(0) }}
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-sm text-slate-900 dark:text-slate-100 truncate">{{ u.name }}</div>
                  <div class="text-xs text-slate-400 truncate">{{ u.dept_name }}</div>
                </div>
                <span v-if="selectedUser?.id === u.id" class="text-xs text-[#3B82F6]">✓</span>
              </button>
              <div v-if="filteredUsers.length === 0" class="px-3 py-4 text-center text-xs text-slate-400">
                未找到匹配用户
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 状态筛选 + 用户信息 -->
    <div
      v-if="selectedUser"
      class="bg-white dark:bg-slate-800 rounded-card shadow-sm p-4 flex flex-wrap items-center justify-between gap-3"
    >
      <div class="flex items-center gap-3">
        <div
          class="w-10 h-10 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm font-medium"
        >
          {{ selectedUser.name.charAt(0) }}
        </div>
        <div>
          <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">
            {{ selectedUser.name }}
          </div>
          <div class="text-xs text-slate-400">{{ selectedUser.dept_name }}</div>
        </div>
      </div>
      <div class="flex gap-2">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-tag font-medium transition-smooth"
          :class="
            status === 'active'
              ? 'bg-blue-500 text-white'
              : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300 hover:bg-slate-200'
          "
          @click="status = 'active'"
        >
          待办（{{ total }}）
        </button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-tag font-medium transition-smooth"
          :class="
            status === 'completed'
              ? 'bg-green-500 text-white'
              : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300 hover:bg-slate-200'
          "
          @click="status = 'completed'"
        >
          已完成
        </button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-tag font-medium transition-smooth"
          :class="
            status === 'archived'
              ? 'bg-slate-700 text-white'
              : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300 hover:bg-slate-200'
          "
          @click="status = 'archived'"
        >
          已归档
        </button>
      </div>
    </div>

    <p v-if="error" class="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-btn">{{ error }}</p>

    <!-- 任务卡片 -->
    <div v-if="!selectedUser" class="text-center py-16">
      <div class="text-4xl mb-3">🔍</div>
      <p class="text-slate-400 text-sm">请先选择一位用户，查看其工作台内容</p>
    </div>
    <div
      v-else-if="loading"
      class="text-center py-16 text-xs text-slate-400"
    >
      加载中...
    </div>
    <div
      v-else-if="notes.length === 0"
      class="text-center py-16"
    >
      <div class="text-4xl mb-3">📭</div>
      <p class="text-slate-400 text-sm">该用户暂无任务</p>
    </div>
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <div
        v-for="note in notes"
        :key="note.id"
        class="bg-white dark:bg-slate-800 rounded-card p-5 shadow-note border border-slate-100 dark:border-slate-700 transition-smooth"
      >
        <div class="flex items-center gap-1.5 mb-2 flex-wrap">
          <span
            class="inline-flex items-center text-[11px] px-1.5 py-0.5 rounded-full font-medium"
            :class="sourceChipClass(sourceLabel(note))"
          >
            {{ sourceLabel(note) }}
          </span>
          <span
            v-if="note.color_status === 'red'"
            class="inline-flex items-center text-[11px] px-1.5 py-0.5 rounded-full bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300 font-medium"
            >重要</span
          >
        </div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-2 line-clamp-1">
          {{ note.title || '无标题' }}
        </h3>
        <p
          class="text-sm text-slate-500 dark:text-slate-300 line-clamp-3 mb-3 rich-content-display"
          v-html="(note.content || '暂无内容')"
        />
        <div v-if="(note.tags || []).length" class="flex items-center gap-1.5 mb-3 flex-wrap">
          <span
            v-for="tag in (note.tags || []).slice(0, 2)"
            :key="tag.id"
            class="tag-capsule text-white text-[11px]"
            :style="{ backgroundColor: tag.color || '#64748B' }"
          >
            {{ tag.name }}
          </span>
          <span v-if="(note.tags || []).length > 2" class="text-xs text-slate-400">
            +{{ (note.tags || []).length - 2 }}
          </span>
        </div>
        <div class="flex items-center justify-between pt-3 border-t border-slate-100 dark:border-slate-700 text-xs">
          <span class="text-slate-400">创建人：{{ note.creator?.name || '-' }}</span>
          <span class="text-slate-400">{{ (note.created_at || '').slice(0, 10) }}</span>
        </div>
        <div
          v-if="(note.assignees || []).length"
          class="mt-2 flex items-center gap-1.5 flex-wrap text-xs"
        >
          <span class="text-slate-400">负责人：</span>
          <span
            v-for="a in note.assignees"
            :key="a.user_id"
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300"
          >
            {{ a.user?.name || (a as any).name }}
            <span
              v-if="(a as any).sign_status === 'signed'"
              class="text-[10px] text-green-600 dark:text-green-400"
              >已签收</span
            >
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
