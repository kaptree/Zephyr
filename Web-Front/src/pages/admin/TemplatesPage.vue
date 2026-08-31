<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { fetchTemplates, createTemplate, updateTemplate, deleteTemplate } from '@/services/templates'
import { useAuthStore } from '@/stores/auth'
import MarkdownEditor from '@/components/common/MarkdownEditor.vue'
import { markdownToHtml } from '@/utils/markdown'
import type { Template } from '@/types'

const auth = useAuthStore()
const templates = ref<Template[]>([])
const loading = ref(true)
const keyword = ref('')
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const showDeleteConfirm = ref(false)
const deletingId = ref<string | null>(null)
const saving = ref(false)

const form = ref({
  name: '',
  type: 'default',
  description: '',
  content: '',
})

const TYPE_LABELS: Record<string, string> = {
  default: '通用任务',
  data_analysis: '数据分析',
  special_project: '专项行动',
  emergency_canvas: '紧急协查',
  collaborative_writing: '协同作战',
  custom: '自定义',
}

const TYPE_COLORS: Record<string, string> = {
  default: 'bg-blue-50 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300',
  data_analysis: 'bg-purple-50 text-purple-600 dark:bg-purple-900/40 dark:text-purple-300',
  special_project: 'bg-red-50 text-red-600 dark:bg-red-900/40 dark:text-red-300',
  emergency_canvas: 'bg-orange-50 text-orange-600 dark:bg-orange-900/40 dark:text-orange-300',
  collaborative_writing: 'bg-green-50 text-green-600 dark:bg-green-900/40 dark:text-green-300',
  custom: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
}

const isSuperAdmin = computed(() => auth.user?.role === 'super_admin')

// 模板可改删权限：系统管理员全部；其余用户仅可管理自己创建的模板
function canManage(t: Template): boolean {
  if (isSuperAdmin.value) return true
  return !!t.creator_id && t.creator_id === auth.user?.id
}

async function loadTemplates() {
  loading.value = true
  try {
    const res = await fetchTemplates({ keyword: keyword.value || undefined })
    templates.value = res.data || []
  } catch {
    templates.value = []
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  loadTemplates()
}

function handleClearSearch() {
  keyword.value = ''
  loadTemplates()
}

function openCreate() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', type: 'default', description: '', content: '' }
  showModal.value = true
}

function openEdit(t: Template) {
  isEditing.value = true
  editingId.value = t.id
  form.value = {
    name: t.name,
    type: t.type,
    description: t.description || '',
    content: t.content || '',
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name.trim()) return
  saving.value = true
  try {
    const payload = {
      name: form.value.name.trim(),
      type: form.value.type,
      description: form.value.description.trim(),
      content: form.value.content,
    }
    if (isEditing.value && editingId.value) {
      await updateTemplate(editingId.value, payload)
    } else {
      await createTemplate(payload)
    }
    showModal.value = false
    await loadTemplates()
  } finally {
    saving.value = false
  }
}

function confirmDelete(id: string) {
  deletingId.value = id
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!deletingId.value) return
  try {
    await deleteTemplate(deletingId.value)
    showDeleteConfirm.value = false
    deletingId.value = null
    await loadTemplates()
  } catch {
    // error handled by interceptor
  }
}

function creatorName(t: Template): string {
  return t.creator?.name || '系统'
}

// 内容预览：优先模板内容，兼容旧 JSON 字段模板
function contentPreview(t: Template): string {
  if (t.content?.trim()) return t.content.trim()
  try {
    const fields = JSON.parse(typeof t.fields === 'string' ? t.fields : JSON.stringify(t.fields ?? '[]'))
    if (Array.isArray(fields) && fields.length) {
      return fields.map((f: any) => `【${f.name}】`).join(' ')
    }
  } catch {
    /* ignore */
  }
  return ''
}

// 卡片内容预览：Markdown → HTML 渲染
function contentPreviewHtml(t: Template): string {
  return markdownToHtml(contentPreview(t))
}

onMounted(loadTemplates)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">模板库管理</h2>
      <button
        class="text-sm px-4 py-2 bg-blue-600 text-white rounded-btn hover:bg-blue-700 transition-smooth"
        @click="openCreate"
      >
        + 添加模版
      </button>
    </div>

    <!-- 搜索栏 -->
    <div
      class="bg-white dark:bg-slate-800 rounded-card p-4 mb-6 flex flex-wrap items-center gap-3 border border-slate-100 dark:border-slate-700 transition-colors duration-300"
    >
      <input
        v-model="keyword"
        class="input-field !w-56"
        placeholder="🔍 搜索模板名称 / 简介 / 内容..."
        @keyup.enter="handleSearch"
      />
      <button class="btn-primary text-sm !py-2" @click="handleSearch">搜索</button>
      <button
        v-if="keyword"
        class="px-4 py-2 text-sm text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 transition-smooth"
        @click="handleClearSearch"
      >
        清空
      </button>
    </div>

    <div v-if="loading" class="text-center text-slate-400 dark:text-slate-500 py-20">加载中...</div>

    <div v-else-if="templates.length === 0" class="text-center text-slate-400 dark:text-slate-500 py-20">
      暂无模板，点击上方按钮创建
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div
        v-for="t in templates"
        :key="t.id"
        class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-5 hover:shadow-note transition-smooth"
      >
        <div class="flex items-start justify-between mb-2">
          <div class="flex items-center gap-2 min-w-0">
            <h4 class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">{{ t.name }}</h4>
            <span v-if="t.is_system" class="text-xs px-1.5 py-0.5 bg-slate-100 dark:bg-slate-700 text-slate-400 dark:text-slate-300 rounded shrink-0">系统</span>
          </div>
          <span class="text-xs px-2 py-0.5 rounded-tag shrink-0" :class="TYPE_COLORS[t.type] || TYPE_COLORS.default">
            {{ TYPE_LABELS[t.type] || t.type }}
          </span>
        </div>

        <!-- 创建人 + 简介 -->
        <div class="flex items-center gap-1.5 text-xs text-slate-400 dark:text-slate-500 mb-2">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          <span class="font-medium text-slate-500 dark:text-slate-400">{{ creatorName(t) }}</span>
          <span class="text-slate-300 dark:text-slate-600">·</span>
          <span class="truncate">{{ t.description || '暂无简介' }}</span>
        </div>

        <!-- 内容预览（Markdown 渲染） -->
        <div
          v-if="contentPreview(t)"
          class="text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/50 rounded-lg px-3 py-2 line-clamp-3 prose prose-sm dark:prose-invert max-w-none [&_h1]:text-sm [&_h1]:font-semibold [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:text-xs [&_p]:my-0.5 [&_ul]:my-0.5 [&_ol]:my-0.5 [&_pre]:my-0.5 [&_pre]:p-1.5 [&_code]:text-[10px] [&_li]:text-xs"
          v-html="contentPreviewHtml(t)"
        ></div>

        <div class="flex items-center justify-between mt-3 pt-3 border-t border-slate-50 dark:border-slate-700">
          <span class="text-xs text-slate-300 dark:text-slate-500">{{ t.created_at?.slice(0, 10) }}</span>
          <div v-if="canManage(t)" class="flex gap-2">
            <button
              class="text-xs px-2.5 py-1 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 rounded hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth"
              @click="openEdit(t)"
            >
              编辑
            </button>
            <button
              class="text-xs px-2.5 py-1 bg-red-50 dark:bg-red-900/40 text-red-600 dark:text-red-300 rounded hover:bg-red-100 dark:hover:bg-red-900/60 transition-smooth"
              @click="confirmDelete(t.id)"
            >
              删除
            </button>
          </div>
          <span v-else class="text-xs text-slate-300 dark:text-slate-500">仅创建者/管理员可编辑</span>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center" @click.self="showModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-lg w-[560px] max-h-[85vh] overflow-y-auto p-6 shadow-xl">
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">
          {{ isEditing ? '编辑模板' : '新建模板' }}
        </h3>
        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">模板名称 *</label>
              <input
                v-model="form.name"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-btn bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-blue-400"
                placeholder="输入模板名称"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">模板类型</label>
              <select v-model="form.type" class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-btn bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-blue-400">
                <option v-for="(label, key) in TYPE_LABELS" :key="key" :value="key">{{ label }}</option>
              </select>
            </div>
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">
              模板简介
              <span class="text-slate-400 dark:text-slate-500 font-normal">— 简要说明模板用途，展示在卡片上</span>
            </label>
            <input
              v-model="form.description"
              class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-btn bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-blue-400"
              placeholder="如：针对专项行动的任务记录模板"
              maxlength="500"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-600 dark:text-slate-300 mb-1">
              模板内容
              <span class="text-slate-400 dark:text-slate-500 font-normal">— 支持 Markdown 语法，可全屏编写；创建任务选用后自动填入</span>
            </label>
            <MarkdownEditor v-model="form.content" :min-height="260" placeholder="支持 Markdown 语法，例如：
# 任务背景
描述任务背景与来源...

## 任务目标
- 目标一
- 目标二

## 工作措施
1. 措施一
2. 措施二

## 完成标准
> 明确验收标准

## 备注" />
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <button
            class="text-sm px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 rounded-btn hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth"
            @click="showModal = false"
          >
            取消
          </button>
          <button
            class="text-sm px-4 py-2 bg-blue-600 text-white rounded-btn hover:bg-blue-700 transition-smooth disabled:opacity-50"
            :disabled="!form.name.trim() || saving"
            @click="handleSave"
          >
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center" @click.self="showDeleteConfirm = false">
      <div class="bg-white dark:bg-slate-800 rounded-lg w-[360px] p-6 shadow-xl">
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-2">确认删除</h3>
        <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">删除后不可恢复，确定要删除此模板吗？</p>
        <div class="flex justify-end gap-2">
          <button
            class="text-sm px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 rounded-btn hover:bg-slate-200 dark:hover:bg-slate-600 transition-smooth"
            @click="showDeleteConfirm = false"
          >
            取消
          </button>
          <button
            class="text-sm px-4 py-2 bg-red-600 text-white rounded-btn hover:bg-red-700 transition-smooth"
            @click="handleDelete"
          >
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
