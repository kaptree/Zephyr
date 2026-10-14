<script setup lang="ts">
/* 工具超市 —— 小工具板块首页：
   搜索（服务端支持拼音首字母）+ 分类筛选 + 热门/最新排序 + 网格卡片；
   卡片含图标/名称/简介/使用次数，点击即进入沙箱运行页；
   底部附「我的上传」与（管理员）「待审核」管理区。 */
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { BASE_URL } from '@/services/api'
import {
  fetchTools,
  fetchMyTools,
  fetchPendingTools,
  deleteTool,
  updateToolStatus,
  TOOL_STATUS_LABEL,
} from '@/services/tools'
import type { ToolItem, ToolSort } from '@/services/tools'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import AppIcon from '@/components/icons/AppIcon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Skeleton from '@/components/common/Skeleton.vue'

const router = useRouter()
const auth = useAuthStore()
const toast = useToast()
const { confirm } = useConfirm()

const loading = ref(true)
const errorMsg = ref('')
const keyword = ref('')
const category = ref('全部')
const sort = ref<ToolSort>('hot')
const list = ref<ToolItem[]>([])
const categories = ref<string[]>([])
const mine = ref<ToolItem[]>([])
const pending = ref<ToolItem[]>([])

async function load() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await fetchTools({
      keyword: keyword.value.trim() || undefined,
      category: category.value === '全部' ? undefined : category.value,
      sort: sort.value,
    })
    list.value = res.data?.list ?? []
  } catch (e: unknown) {
    errorMsg.value = (e as { friendlyMessage?: string })?.friendlyMessage || '加载工具列表失败'
  } finally {
    loading.value = false
  }
}

async function loadMine() {
  try {
    mine.value = (await fetchMyTools()).data?.list ?? []
  } catch {
    mine.value = []
  }
}

async function loadPending() {
  if (!auth.isAdmin) return
  try {
    pending.value = (await fetchPendingTools()).data?.list ?? []
  } catch {
    pending.value = []
  }
}

// 搜索防抖（服务端过滤，400ms 防抖减轻请求压力）
let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(load, 400)
}

function switchCategory(c: string) {
  if (category.value === c) return
  category.value = c
  load()
}

function switchSort(s: ToolSort) {
  if (sort.value === s) return
  sort.value = s
  load()
}

const SORT_TABS: { key: ToolSort; label: string }[] = [
  { key: 'hot', label: '热门' },
  { key: 'latest', label: '最新' },
]

const STATUS_STYLE: Record<ToolItem['status'], string> = {
  pending:
    'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 ring-amber-300/60 dark:ring-amber-600/40',
  published:
    'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300 ring-green-300/60 dark:ring-green-600/40',
  rejected:
    'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-300 ring-red-300/60 dark:ring-red-600/40',
  offline:
    'bg-slate-200 dark:bg-slate-700/60 text-slate-600 dark:text-slate-300 ring-slate-300/60 dark:ring-slate-600/40',
}

function openTool(t: ToolItem) {
  router.push(`/tools/${t.slug}`)
}

function goUpload() {
  router.push('/tools/upload')
}

/** 图标地址：上传图片需拼 API 域名（data URI 直接使用） */
function iconSrc(t: ToolItem): string | null {
  if (t.icon_type !== 'image' || !t.icon_value) return null
  return t.icon_value.startsWith('data:') ? t.icon_value : `${BASE_URL}${t.icon_value}`
}

function formatCount(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}w`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

async function review(t: ToolItem, status: 'published' | 'rejected') {
  try {
    await updateToolStatus(t.slug, status)
    toast.success(status === 'published' ? `「${t.name}」已上架` : `「${t.name}」已打回`)
    pending.value = pending.value.filter((p) => p.slug !== t.slug)
    load()
  } catch (e: unknown) {
    toast.error((e as { friendlyMessage?: string })?.friendlyMessage || '操作失败')
  }
}

async function removeMine(t: ToolItem) {
  if (
    !(await confirm({
      title: '删除工具',
      message: `确定删除「${t.name}」吗？删除后不可恢复。`,
      confirmText: '删除',
      danger: true,
    }))
  )
    return
  try {
    await deleteTool(t.slug)
    toast.success('工具已删除')
    mine.value = mine.value.filter((m) => m.slug !== t.slug)
    load()
  } catch (e: unknown) {
    toast.error((e as { friendlyMessage?: string })?.friendlyMessage || '删除失败')
  }
}

const hasMine = computed(() => mine.value.length > 0)
const hasPending = computed(() => auth.isAdmin && pending.value.length > 0)

onMounted(() => {
  load()
  loadMine()
  loadPending()
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors duration-300">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 py-8">
      <!-- 顶部：标题 + 上传入口 -->
      <div class="flex items-start justify-between gap-4 mb-6">
        <div>
          <h1 class="text-2xl font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
            <AppIcon name="grid" :size="24" class="text-indigo-500 dark:text-indigo-400" />
            工具超市
          </h1>
          <p class="mt-1 text-sm text-slate-400 dark:text-slate-500">
            实用小工具一站式入口，点击即用；也支持上传自己的 HTML 工具
          </p>
        </div>
        <button
          class="shrink-0 inline-flex items-center gap-1.5 px-4 py-2.5 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-semibold shadow-sm transition-all duration-200 hover:-translate-y-0.5 active:translate-y-0"
          @click="goUpload"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          上传工具
        </button>
      </div>

      <!-- 搜索 + 排序 -->
      <div class="flex flex-col sm:flex-row gap-3 sm:items-center mb-4">
        <div class="relative flex-1">
          <svg
            class="absolute left-3.5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400"
            fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 10.5a6.5 6.5 0 11-13 0 6.5 6.5 0 0113 0z" />
          </svg>
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索工具名称、简介、标签（支持拼音首字母）"
            class="w-full pl-10 pr-4 py-2.5 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-sm text-slate-700 dark:text-slate-200 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-400/50 focus:border-indigo-400 transition-shadow"
            @input="onSearch"
          />
        </div>
        <div class="flex rounded-xl bg-slate-100 dark:bg-slate-800 p-1 shrink-0 self-start">
          <button
            v-for="tab in SORT_TABS"
            :key="tab.key"
            :class="[
              'px-4 py-1.5 rounded-lg text-sm font-medium transition-all duration-200',
              sort === tab.key
                ? 'bg-white dark:bg-slate-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200',
            ]"
            @click="switchSort(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <!-- 分类筛选 -->
      <div class="flex gap-2 flex-wrap mb-6">
        <button
          v-for="c in ['全部', ...categories]"
          :key="c"
          :class="[
            'px-3.5 py-1.5 rounded-full text-xs font-medium ring-1 transition-all duration-200',
            category === c
              ? 'bg-indigo-500 text-white ring-indigo-500 shadow-sm'
              : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 ring-slate-200 dark:ring-slate-700 hover:ring-indigo-300 dark:hover:ring-indigo-600',
          ]"
          @click="switchCategory(c)"
        >
          {{ c }}
        </button>
      </div>

      <!-- 错误 / 加载 / 空态 -->
      <div v-if="errorMsg" class="text-center py-16 text-sm text-red-500">{{ errorMsg }}</div>
      <div v-else-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <Skeleton v-for="i in 6" :key="i" class="h-28 rounded-2xl" />
      </div>
      <EmptyState
        v-else-if="list.length === 0"
        illustration="coffee"
        message="没有找到匹配的工具，换个关键词试试，或上传一个新工具"
      />

      <!-- 工具网格 -->
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <button
          v-for="t in list"
          :key="t.slug"
          class="group text-left bg-white dark:bg-slate-800 rounded-2xl border border-slate-200/80 dark:border-slate-700 p-5 transition-all duration-200 hover:-translate-y-1 hover:shadow-lg hover:border-indigo-300 dark:hover:border-indigo-600 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          @click="openTool(t)"
        >
          <div class="flex items-start gap-3.5">
            <!-- 图标 -->
            <div
              class="shrink-0 w-12 h-12 rounded-xl bg-gradient-to-br from-indigo-100 to-sky-100 dark:from-indigo-900/60 dark:to-sky-900/60 flex items-center justify-center overflow-hidden"
            >
              <img
                v-if="iconSrc(t)"
                :src="iconSrc(t)!"
                :alt="t.name"
                class="w-full h-full object-cover"
              />
              <span
                v-else
                :class="[
                  t.icon_type === 'emoji' ? 'text-2xl leading-none' : 'text-lg font-bold text-indigo-500 dark:text-indigo-300',
                ]"
              >{{ t.icon_value || t.name.slice(0, 1) }}</span>
            </div>
            <!-- 名称 + 简介 -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-1.5">
                <h3 class="font-semibold text-slate-800 dark:text-slate-100 truncate">{{ t.name }}</h3>
                <span
                  v-if="t.is_system"
                  class="shrink-0 text-[10px] px-1.5 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-900/50 text-indigo-500 dark:text-indigo-300 ring-1 ring-indigo-200 dark:ring-indigo-700"
                >内置</span>
              </div>
              <p class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400 line-clamp-2">
                {{ t.description }}
              </p>
            </div>
            <!-- 使用箭头 -->
            <svg
              class="shrink-0 w-4 h-4 text-slate-300 dark:text-slate-600 transition-all duration-200 group-hover:text-indigo-500 group-hover:translate-x-0.5"
              fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
          <!-- 底部：使用次数 + 分类 -->
          <div class="mt-3.5 flex items-center gap-2 text-xs">
            <span class="inline-flex items-center gap-1 text-slate-400 dark:text-slate-500">
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.042 21.672L13.684 16.6m0 0l-2.51 2.225.569-9.47 5.227 7.917-3.286-.672zM12 2.25V4.5m5.834.166l-1.591 1.591M20.25 10.5H18M7.757 14.743l-1.59 1.59M6 10.5H3.75m4.007-4.243l-1.59-1.59" />
              </svg>
              {{ formatCount(t.launch_count) }} 次使用
            </span>
            <span class="ml-auto px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-700/60 text-slate-500 dark:text-slate-400">
              {{ t.category }}
            </span>
          </div>
        </button>
      </div>

      <!-- 待审核（管理员） -->
      <div v-if="hasPending" class="mt-10">
        <h2 class="text-base font-semibold text-slate-700 dark:text-slate-200 mb-3 flex items-center gap-2">
          <AppIcon name="check-circle" :size="18" class="text-amber-500" />
          待审核工具
          <span class="text-xs font-normal text-slate-400">{{ pending.length }} 个</span>
        </h2>
        <div class="space-y-3">
          <div
            v-for="t in pending"
            :key="t.slug"
            class="flex flex-wrap items-center gap-3 bg-white dark:bg-slate-800 rounded-xl border border-amber-200/70 dark:border-amber-700/40 px-4 py-3"
          >
            <div class="min-w-0 flex-1">
              <div class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ t.name }}</div>
              <div class="text-xs text-slate-400 truncate">{{ t.description }} · 上传者：{{ t.author_name || '未知' }}</div>
            </div>
            <button
              class="px-3 py-1.5 rounded-lg bg-green-500 hover:bg-green-600 text-white text-xs font-medium transition-colors"
              @click="review(t, 'published')"
            >上架</button>
            <button
              class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 text-xs font-medium hover:bg-red-100 dark:hover:bg-red-900/40 hover:text-red-600 dark:hover:text-red-300 transition-colors"
              @click="review(t, 'rejected')"
            >打回</button>
          </div>
        </div>
      </div>

      <!-- 我的上传 -->
      <div v-if="hasMine" class="mt-10">
        <h2 class="text-base font-semibold text-slate-700 dark:text-slate-200 mb-3 flex items-center gap-2">
          <AppIcon name="pencil" :size="18" class="text-indigo-500" />
          我的上传
          <span class="text-xs font-normal text-slate-400">{{ mine.length }} 个</span>
        </h2>
        <div class="space-y-3">
          <div
            v-for="t in mine"
            :key="t.slug"
            class="flex flex-wrap items-center gap-3 bg-white dark:bg-slate-800 rounded-xl border border-slate-200/80 dark:border-slate-700 px-4 py-3"
          >
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ t.name }}</span>
                <span
                  :class="['text-[10px] px-1.5 py-0.5 rounded-full ring-1', STATUS_STYLE[t.status]]"
                >{{ TOOL_STATUS_LABEL[t.status] }}</span>
              </div>
              <div class="text-xs text-slate-400 truncate">{{ t.description }} · {{ t.launch_count }} 次使用</div>
            </div>
            <button
              v-if="t.status === 'published'"
              class="px-3 py-1.5 rounded-lg bg-indigo-50 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-300 text-xs font-medium hover:bg-indigo-100 dark:hover:bg-indigo-900/70 transition-colors"
              @click="openTool(t)"
            >打开</button>
            <button
              class="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-500 dark:text-slate-400 hover:bg-red-50 dark:hover:bg-red-900/30 hover:text-red-500 transition-colors"
              @click="removeMine(t)"
            >删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
