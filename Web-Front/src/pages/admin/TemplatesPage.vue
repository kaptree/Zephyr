<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchTemplates, createTemplate, updateTemplate, deleteTemplate } from '@/services/templates'
import type { Template } from '@/types'

const templates = ref<Template[]>([])
const loading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const showDeleteConfirm = ref(false)
const deletingId = ref<string | null>(null)
const saving = ref(false)

const form = ref({
  name: '',
  type: 'default',
  fields: '',
  layout: '1',
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
  default: 'bg-blue-50 text-blue-600',
  data_analysis: 'bg-purple-50 text-purple-600',
  special_project: 'bg-red-50 text-red-600',
  emergency_canvas: 'bg-orange-50 text-orange-600',
  collaborative_writing: 'bg-green-50 text-green-600',
  custom: 'bg-slate-100 text-slate-600',
}

async function loadTemplates() {
  loading.value = true
  try {
    const res = await fetchTemplates()
    templates.value = res.data || []
  } catch {
    templates.value = []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', type: 'default', fields: '', layout: '1' }
  showModal.value = true
}

function openEdit(t: Template) {
  isEditing.value = true
  editingId.value = t.id
  let fieldsStr = t.fields
  if (typeof fieldsStr === 'object') {
    fieldsStr = JSON.stringify(fieldsStr, null, 2)
  }
  form.value = {
    name: t.name,
    type: t.type,
    fields: fieldsStr,
    layout: t.layout || '1',
  }
  showModal.value = true
}

async function handleSave() {
  if (!form.value.name.trim()) return
  saving.value = true
  try {
    if (isEditing.value && editingId.value) {
      await updateTemplate(editingId.value, {
        name: form.value.name,
        type: form.value.type,
        fields: form.value.fields || undefined,
        layout: form.value.layout,
      })
    } else {
      await createTemplate({
        name: form.value.name,
        type: form.value.type,
        fields: form.value.fields || undefined,
        layout: form.value.layout,
      })
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

function parseFields(fieldsStr: string): any[] {
  try {
    return JSON.parse(fieldsStr)
  } catch {
    return []
  }
}

onMounted(loadTemplates)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">模板库管理</h2>
      <button
        class="text-sm px-4 py-2 bg-blue-600 text-white rounded-btn hover:bg-blue-700 transition-smooth"
        @click="openCreate"
      >
        + 添加模版
      </button>
    </div>

    <div v-if="loading" class="text-center text-slate-400 py-20">加载中...</div>

    <div v-else-if="templates.length === 0" class="text-center text-slate-400 py-20">
      暂无模板，点击上方按钮创建
    </div>

    <div v-else class="grid grid-cols-2 gap-4">
      <div
        v-for="t in templates"
        :key="t.id"
        class="bg-white rounded-card border border-slate-100 p-5 hover:shadow-note transition-smooth"
      >
        <div class="flex items-start justify-between mb-2">
          <div class="flex items-center gap-2">
            <h4 class="text-sm font-semibold text-slate-900">{{ t.name }}</h4>
            <span v-if="t.is_system" class="text-xs px-1.5 py-0.5 bg-slate-100 text-slate-400 rounded">系统</span>
          </div>
          <span class="text-xs px-2 py-0.5 rounded-tag" :class="TYPE_COLORS[t.type] || TYPE_COLORS.default">
            {{ TYPE_LABELS[t.type] || t.type }}
          </span>
        </div>
        <div class="flex items-center justify-between mt-3 pt-3 border-t border-slate-50">
          <span class="text-xs text-slate-400">{{ parseFields(t.fields).length }} 个字段</span>
          <div class="flex gap-2">
            <button
              class="text-xs px-2.5 py-1 bg-slate-100 text-slate-600 rounded hover:bg-slate-200 transition-smooth"
              @click="openEdit(t)"
            >
              编辑
            </button>
            <button
              v-if="!t.is_system"
              class="text-xs px-2.5 py-1 bg-red-50 text-red-600 rounded hover:bg-red-100 transition-smooth"
              @click="confirmDelete(t.id)"
            >
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center" @click.self="showModal = false">
      <div class="bg-white rounded-lg w-[480px] max-h-[80vh] overflow-y-auto p-6 shadow-xl">
        <h3 class="text-base font-semibold text-slate-900 mb-4">
          {{ isEditing ? '编辑模板' : '新建模板' }}
        </h3>
        <div class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-slate-600 mb-1">模板名称 *</label>
            <input
              v-model="form.name"
              class="w-full px-3 py-2 text-sm border border-slate-200 rounded-btn focus:outline-none focus:border-blue-400"
              placeholder="输入模板名称"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-600 mb-1">模板类型</label>
            <select v-model="form.type" class="w-full px-3 py-2 text-sm border border-slate-200 rounded-btn focus:outline-none focus:border-blue-400">
              <option v-for="(label, key) in TYPE_LABELS" :key="key" :value="key">{{ label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-600 mb-1">布局样式</label>
            <select v-model="form.layout" class="w-full px-3 py-2 text-sm border border-slate-200 rounded-btn focus:outline-none focus:border-blue-400">
              <option value="1">单栏</option>
              <option value="2">双栏</option>
              <option value="4">四宫格</option>
              <option value="6">六宫格</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium text-slate-600 mb-1">
              字段定义 (JSON)
              <span class="text-slate-400 font-normal">— 示例：[{"name":"任务描述","type":"textarea","required":true,"order":1}]</span>
            </label>
            <textarea
              v-model="form.fields"
              rows="6"
              class="w-full px-3 py-2 text-xs border border-slate-200 rounded-btn focus:outline-none focus:border-blue-400 font-mono"
              placeholder='[{"name":"字段名","type":"text","required":true,"order":1}]'
            ></textarea>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-6">
          <button
            class="text-sm px-4 py-2 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
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
      <div class="bg-white rounded-lg w-[360px] p-6 shadow-xl">
        <h3 class="text-base font-semibold text-slate-900 mb-2">确认删除</h3>
        <p class="text-sm text-slate-500 mb-4">删除后不可恢复，确定要删除此模板吗？</p>
        <div class="flex justify-end gap-2">
          <button
            class="text-sm px-4 py-2 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
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