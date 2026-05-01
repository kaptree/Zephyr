<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  getSystemConfig,
  updateSystemConfig,
  listAIConfigs,
  createAIConfig,
  updateAIConfig,
  deleteAIConfig,
  listConfigFiles,
  getConfigFile,
  updateConfigFile,
  getConfigFileHistory,
  listAdminLogs,
} from '@/services/system'
import type {
  AIConfigItem,
  ConfigFileItem,
  ConfigFileContent,
  ConfigFileHistoryItem,
  AdminLogItem,
} from '@/types/system'

const safeObj = (v: unknown): Record<string, unknown> => {
  return (v && typeof v === 'object' && !Array.isArray(v)) ? v as Record<string, unknown> : {}
}

const activeTab = ref<'config' | 'ai' | 'files' | 'logs'>('config')

const tabs = [
  { key: 'config' as const, label: '系统设置', icon: '⚙️' },
  { key: 'ai' as const, label: 'AI服务配置', icon: '🤖' },
  { key: 'files' as const, label: '配置文件管理', icon: '📂' },
  { key: 'logs' as const, label: '操作日志', icon: '📋' },
]

// ===== 系统设置 Tab =====
const systemConfig = ref<Record<string, unknown>>({})
const configLoading = ref(false)
const configSaving = ref(false)
const configToast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const editingConfigKey = ref('')
const editingConfigValue = ref('')

function showConfigToast(type: 'success' | 'error', message: string) {
  configToast.value = { type, message }
  setTimeout(() => { configToast.value = null }, 3000)
}

async function loadSystemConfig() {
  configLoading.value = true
  try {
    const res = await getSystemConfig()
    systemConfig.value = safeObj(res.data)
  } catch {
    systemConfig.value = {}
    showConfigToast('error', '加载系统配置失败')
  } finally {
    configLoading.value = false
  }
}

function getConfigSections(): { key: string; label: string; fields: { key: string; value: unknown }[] }[] {
  const cfg = systemConfig.value
  const sections: { key: string; label: string; fields: { key: string; value: unknown }[] }[] = []
  const knownSections: Record<string, string> = {
    server: '服务器配置',
    database: '数据库配置',
    redis: 'Redis配置',
    jwt: 'JWT认证',
    log: '日志配置',
    storage: '存储配置',
    websocket: 'WebSocket',
    rate_limit: '接口限流',
    security: '安全配置',
    scheduler: '定时任务',
    features: '功能开关',
  }
  for (const [key, label] of Object.entries(knownSections)) {
    const sectionData = cfg[key]
    if (sectionData && typeof sectionData === 'object' && !Array.isArray(sectionData)) {
      const fields = Object.entries(sectionData as Record<string, unknown>).map(([k, v]) => ({
        key: `${key}.${k}`,
        value: v,
      }))
      sections.push({ key, label, fields })
    }
  }
  return sections
}

function startEditConfig(key: string, value: unknown) {
  editingConfigKey.value = key
  editingConfigValue.value = typeof value === 'string' ? value : JSON.stringify(value)
}

function cancelEditConfig() {
  editingConfigKey.value = ''
  editingConfigValue.value = ''
}

function getDisplayValue(value: unknown): string {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

async function saveConfigField() {
  if (!editingConfigKey.value) return
  const parts = editingConfigKey.value.split('.')
  if (parts.length !== 2) return

  let parsedValue: unknown = editingConfigValue.value
  if (editingConfigValue.value === 'true') parsedValue = true
  else if (editingConfigValue.value === 'false') parsedValue = false
  else if (/^\d+$/.test(editingConfigValue.value)) parsedValue = parseInt(editingConfigValue.value)

  configSaving.value = true
  try {
    const updateData: Record<string, unknown> = {
      [parts[0]]: { [parts[1]]: parsedValue },
    }
    await updateSystemConfig(updateData)
    showConfigToast('success', '配置已更新并热加载生效')
    cancelEditConfig()
    await loadSystemConfig()
  } catch {
    showConfigToast('error', '更新配置失败')
  } finally {
    configSaving.value = false
  }
}

// ===== AI服务配置 Tab =====
const aiConfigs = ref<AIConfigItem[]>([])
const aiLoading = ref(false)
const aiToast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const showAIForm = ref(false)
const editingAIId = ref<string>('')
const aiForm = ref({
  provider_name: '',
  api_endpoint: '',
  api_key: '',
  model_name: '',
  description: '',
  is_active: true,
})
const deletingAIId = ref<string>('')

function showAIToast(type: 'success' | 'error', message: string) {
  aiToast.value = { type, message }
  setTimeout(() => { aiToast.value = null }, 3000)
}

async function loadAIConfigs() {
  aiLoading.value = true
  try {
    const res = await listAIConfigs()
    aiConfigs.value = res.data as AIConfigItem[]
  } catch {
    showAIToast('error', '加载AI配置失败')
  } finally {
    aiLoading.value = false
  }
}

function openCreateAIForm() {
  editingAIId.value = ''
  aiForm.value = {
    provider_name: '',
    api_endpoint: '',
    api_key: '',
    model_name: '',
    description: '',
    is_active: true,
  }
  showAIForm.value = true
}

function openEditAIForm(item: AIConfigItem) {
  editingAIId.value = item.id
  aiForm.value = {
    provider_name: item.provider_name,
    api_endpoint: item.api_endpoint,
    api_key: '',
    model_name: item.model_name || '',
    description: item.description || '',
    is_active: item.is_active,
  }
  showAIForm.value = true
}

function closeAIForm() {
  showAIForm.value = false
  editingAIId.value = ''
}

async function saveAIForm() {
  if (!aiForm.value.provider_name || !aiForm.value.api_endpoint) {
    showAIToast('error', '请填写服务商名称和API端点')
    return
  }
  if (!editingAIId.value && !aiForm.value.api_key) {
    showAIToast('error', '请填写API密钥')
    return
  }

  try {
    if (editingAIId.value) {
      await updateAIConfig(editingAIId.value, {
        provider_name: aiForm.value.provider_name,
        api_endpoint: aiForm.value.api_endpoint,
        api_key: aiForm.value.api_key || undefined,
        model_name: aiForm.value.model_name,
        description: aiForm.value.description,
        is_active: aiForm.value.is_active,
      })
      showAIToast('success', 'AI配置已更新')
    } else {
      await createAIConfig(aiForm.value)
      showAIToast('success', 'AI配置已创建')
    }
    closeAIForm()
    await loadAIConfigs()
  } catch {
    showAIToast('error', '保存AI配置失败')
  }
}

async function handleDeleteAI(id: string) {
  try {
    await deleteAIConfig(id)
    showAIToast('success', 'AI配置已删除')
    deletingAIId.value = ''
    await loadAIConfigs()
  } catch {
    showAIToast('error', '删除AI配置失败')
  }
}

// ===== 配置文件管理 Tab =====
const configFiles = ref<ConfigFileItem[]>([])
const filesLoading = ref(false)
const selectedFile = ref<ConfigFileContent | null>(null)
const fileContent = ref('')
const fileLoading = ref(false)
const fileSaving = ref(false)
const changeSummary = ref('')
const fileHistories = ref<ConfigFileHistoryItem[]>([])
const showHistory = ref(false)
const fileToast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const viewingHistoryContent = ref<{ before: string; after: string } | null>(null)

function showFileToast(type: 'success' | 'error', message: string) {
  fileToast.value = { type, message }
  setTimeout(() => { fileToast.value = null }, 3000)
}

async function loadConfigFiles() {
  filesLoading.value = true
  try {
    const res = await listConfigFiles()
    configFiles.value = res.data as ConfigFileItem[]
  } catch {
    showFileToast('error', '加载配置文件列表失败')
  } finally {
    filesLoading.value = false
  }
}

async function openConfigFile(file: ConfigFileItem) {
  fileLoading.value = true
  showHistory.value = false
  viewingHistoryContent.value = null
  try {
    const res = await getConfigFile(file.name)
    selectedFile.value = res.data as ConfigFileContent
    fileContent.value = selectedFile.value.content
    changeSummary.value = ''
  } catch {
    showFileToast('error', '加载配置文件失败')
  } finally {
    fileLoading.value = false
  }
}

async function saveConfigFile() {
  if (!selectedFile.value) return

  try {
    JSON.parse(fileContent.value)
  } catch {
    showFileToast('error', 'JSON格式校验失败，请检查语法')
    return
  }

  fileSaving.value = true
  try {
    await updateConfigFile(selectedFile.value.name, fileContent.value, changeSummary.value || '手动编辑配置')
    showFileToast('success', '配置文件已保存')
    changeSummary.value = ''
  } catch {
    showFileToast('error', '保存配置文件失败')
  } finally {
    fileSaving.value = false
  }
}

async function loadFileHistory(fileName: string) {
  try {
    const res = await getConfigFileHistory(fileName)
    fileHistories.value = res.data as ConfigFileHistoryItem[]
    showHistory.value = true
  } catch {
    showFileToast('error', '加载历史记录失败')
  }
}

function viewHistoryDiff(item: ConfigFileHistoryItem) {
  viewingHistoryContent.value = {
    before: item.content_before,
    after: item.content_after,
  }
}

// ===== 操作日志 Tab =====
const adminLogs = ref<AdminLogItem[]>([])
const logsLoading = ref(false)
const logsPage = ref(1)
const logsTotal = ref(0)
const logsPageSize = 20

async function loadAdminLogs() {
  logsLoading.value = true
  try {
    const res = await listAdminLogs(logsPage.value, logsPageSize)
    const data = res.data as unknown as { data: AdminLogItem[]; total: number }
    adminLogs.value = data.data || []
    logsTotal.value = data.total || 0
  } catch {
    // ignore
  } finally {
    logsLoading.value = false
  }
}

function prevLogsPage() {
  if (logsPage.value > 1) {
    logsPage.value--
    loadAdminLogs()
  }
}

function nextLogsPage() {
  if (logsPage.value * logsPageSize < logsTotal.value) {
    logsPage.value++
    loadAdminLogs()
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const actionLabels: Record<string, string> = {
  update_config: '更新系统配置',
  create_ai_config: '创建AI配置',
  update_ai_config: '更新AI配置',
  delete_ai_config: '删除AI配置',
  update_config_file: '编辑配置文件',
}

// ===== Init =====
onMounted(() => {
  loadSystemConfig()
  loadAIConfigs()
  loadConfigFiles()
  loadAdminLogs()
})
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-slate-900 transition-colors duration-300">
    <!-- Header -->
    <div class="shrink-0 px-6 py-4 border-b border-slate-200 dark:border-slate-700">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">系统管理</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">管理系统配置、AI服务与配置文件</p>
    </div>

    <!-- Tab Navigation -->
    <div class="shrink-0 px-6 border-b border-slate-200 dark:border-slate-700">
      <div class="flex gap-1 -mb-px">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="[
            'px-4 py-2.5 text-sm font-medium border-b-2 transition-smooth flex items-center gap-1.5',
            activeTab === tab.key
              ? 'border-blue-500 text-blue-600 dark:text-blue-400'
              : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 hover:border-slate-300'
          ]"
          @click="activeTab = tab.key"
        >
          <span>{{ tab.icon }}</span>
          <span>{{ tab.label }}</span>
        </button>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-auto p-6">

      <!-- ===== 系统设置 Tab ===== -->
      <div v-if="activeTab === 'config'" class="max-w-4xl">
        <!-- Toast -->
        <div
          v-if="configToast"
          :class="[
            'mb-4 px-4 py-3 rounded-lg text-sm flex items-center gap-2',
            configToast.type === 'success' ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800' : 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
          ]"
        >
          <span>{{ configToast.type === 'success' ? '✅' : '❌' }}</span>
          <span>{{ configToast.message }}</span>
        </div>

        <div v-if="configLoading" class="flex items-center justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
        </div>

        <div v-else>
          <div v-for="section in getConfigSections()" :key="section.key" class="mb-6">
            <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 pb-2 border-b border-slate-100 dark:border-slate-800">
              {{ section.label }}
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="field in section.fields"
                :key="field.key"
                class="flex flex-col gap-1 p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-700"
              >
                <label class="text-[11px] font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wide">
                  {{ field.key.split('.')[1] }}
                </label>
                <template v-if="editingConfigKey === field.key">
                  <div class="flex items-center gap-2">
                    <input
                      v-model="editingConfigValue"
                      type="text"
                      class="flex-1 px-2 py-1 text-sm border border-blue-300 dark:border-blue-600 rounded bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                      @keyup.enter="saveConfigField()"
                    />
                    <button
                      class="px-2.5 py-1 text-xs font-medium text-white bg-blue-500 hover:bg-blue-600 rounded transition-smooth disabled:opacity-50"
                      :disabled="configSaving"
                      @click="saveConfigField()"
                    >保存</button>
                    <button
                      class="px-2.5 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded transition-smooth"
                      @click="cancelEditConfig()"
                    >取消</button>
                  </div>
                </template>
                <template v-else>
                  <div class="flex items-center justify-between group">
                    <span class="text-sm text-slate-900 dark:text-slate-100 font-mono break-all">
                      {{ getDisplayValue(field.value) }}
                    </span>
                    <button
                      class="opacity-0 group-hover:opacity-100 px-2 py-0.5 text-xs text-blue-500 hover:text-blue-700 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded transition-smooth"
                      @click="startEditConfig(field.key, field.value)"
                    >编辑</button>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== AI服务配置 Tab ===== -->
      <div v-if="activeTab === 'ai'" class="max-w-4xl">
        <div
          v-if="aiToast"
          :class="[
            'mb-4 px-4 py-3 rounded-lg text-sm flex items-center gap-2',
            aiToast.type === 'success' ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800' : 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
          ]"
        >
          <span>{{ aiToast.type === 'success' ? '✅' : '❌' }}</span>
          <span>{{ aiToast.message }}</span>
        </div>

        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">AI服务配置列表</h3>
          <button
            v-if="!showAIForm"
            class="px-3 py-1.5 text-xs font-medium text-white bg-blue-500 hover:bg-blue-600 rounded-btn transition-smooth"
            @click="openCreateAIForm()"
          >+ 添加配置</button>
        </div>

        <!-- AI Form Modal -->
        <div v-if="showAIForm" class="mb-6 p-5 rounded-xl border-2 border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-900/20">
          <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4">
            {{ editingAIId ? '编辑AI配置' : '新增AI配置' }}
          </h4>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">服务商名称 *</label>
              <input
                v-model="aiForm.provider_name"
                type="text"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                placeholder="如：OpenAI / 阿里云通义千问"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">API端点 *</label>
              <input
                v-model="aiForm.api_endpoint"
                type="text"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                placeholder="https://api.openai.com/v1"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">
                API密钥 {{ editingAIId ? '(留空不修改)' : '*' }}
              </label>
              <input
                v-model="aiForm.api_key"
                type="password"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                :placeholder="editingAIId ? '留空则不修改密钥' : 'sk-...'"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">模型名称</label>
              <input
                v-model="aiForm.model_name"
                type="text"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                placeholder="gpt-4"
              />
            </div>
            <div class="sm:col-span-2">
              <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">描述</label>
              <input
                v-model="aiForm.description"
                type="text"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                placeholder="该配置的用途说明"
              />
            </div>
            <div class="sm:col-span-2 flex items-center gap-2">
              <label class="relative inline-flex items-center cursor-pointer">
                <input v-model="aiForm.is_active" type="checkbox" class="sr-only peer" />
                <div class="w-9 h-5 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-500"></div>
              </label>
              <span class="text-xs text-slate-500 dark:text-slate-400">启用</span>
            </div>
          </div>
          <div class="flex items-center gap-2 mt-4">
            <button
              class="px-4 py-1.5 text-xs font-medium text-white bg-blue-500 hover:bg-blue-600 rounded-btn transition-smooth"
              @click="saveAIForm()"
            >保存</button>
            <button
              class="px-4 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-btn transition-smooth"
              @click="closeAIForm()"
            >取消</button>
          </div>
        </div>

        <div v-if="aiLoading" class="flex items-center justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
        </div>

        <div v-else-if="aiConfigs.length === 0" class="text-center py-12 text-slate-400 dark:text-slate-500">
          <p class="text-lg mb-2">🤖</p>
          <p class="text-sm">暂无AI服务配置</p>
          <p class="text-xs mt-1">点击"添加配置"开始创建</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="item in aiConfigs"
            :key="item.id"
            :class="[
              'p-4 rounded-xl border transition-smooth',
              item.is_active
                ? 'border-green-200 dark:border-green-800 bg-green-50/50 dark:bg-green-900/10'
                : 'border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/30 opacity-50'
            ]"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-sm font-semibold text-slate-900 dark:text-slate-100">{{ item.provider_name }}</span>
                  <span :class="[
                    'px-1.5 py-0.5 text-[10px] font-medium rounded',
                    item.is_active ? 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-400' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                  ]">
                    {{ item.is_active ? '启用中' : '已禁用' }}
                  </span>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 mb-1">{{ item.api_endpoint }}</p>
                <div class="flex items-center gap-3 text-xs text-slate-400 dark:text-slate-500">
                  <span v-if="item.model_name">模型: {{ item.model_name }}</span>
                  <span>密钥: {{ item.api_key_masked || '••••••••' }}</span>
                  <span v-if="item.description">{{ item.description }}</span>
                </div>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-3">
                <button
                  class="px-2 py-1 text-xs text-blue-500 hover:text-blue-700 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded transition-smooth"
                  @click="openEditAIForm(item)"
                >编辑</button>
                <button
                  class="px-2 py-1 text-xs text-red-500 hover:text-red-700 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded transition-smooth"
                  @click="deletingAIId = item.id"
                >删除</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Delete Confirm Dialog -->
        <Teleport to="body">
          <div
            v-if="deletingAIId"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
            @click.self="deletingAIId = ''"
          >
            <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-6 max-w-sm w-full mx-4">
              <h4 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-2">确认删除</h4>
              <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">删除后无法恢复，确定要删除该AI配置吗？</p>
              <div class="flex justify-end gap-2">
                <button
                  class="px-4 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-btn transition-smooth"
                  @click="deletingAIId = ''"
                >取消</button>
                <button
                  class="px-4 py-1.5 text-sm font-medium text-white bg-red-500 hover:bg-red-600 rounded-btn transition-smooth"
                  @click="handleDeleteAI(deletingAIId)"
                >确认删除</button>
              </div>
            </div>
          </div>
        </Teleport>
      </div>

      <!-- ===== 配置文件管理 Tab ===== -->
      <div v-if="activeTab === 'files'" class="max-w-5xl">
        <div
          v-if="fileToast"
          :class="[
            'mb-4 px-4 py-3 rounded-lg text-sm flex items-center gap-2',
            fileToast.type === 'success' ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800' : 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
          ]"
        >
          <span>{{ fileToast.type === 'success' ? '✅' : '❌' }}</span>
          <span>{{ fileToast.message }}</span>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <!-- File List -->
          <div class="lg:col-span-1">
            <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">配置文件列表</h3>
            <div v-if="filesLoading" class="text-center py-8 text-slate-400">
              <div class="animate-spin rounded-full h-6 w-6 border-2 border-blue-500 border-t-transparent mx-auto mb-2"></div>
            </div>
            <div v-else class="space-y-1">
              <button
                v-for="file in configFiles"
                :key="file.name"
                :class="[
                  'w-full text-left px-3 py-2.5 rounded-lg text-sm transition-smooth',
                  selectedFile?.name === file.name
                    ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium'
                    : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
                ]"
                @click="openConfigFile(file)"
              >
                <div class="flex items-center gap-2">
                  <span>📄</span>
                  <span class="font-mono text-xs truncate">{{ file.name }}</span>
                </div>
              </button>
            </div>
          </div>

          <!-- File Content Editor -->
          <div class="lg:col-span-2">
            <template v-if="!selectedFile">
              <div class="text-center py-16 text-slate-400 dark:text-slate-500">
                <p class="text-2xl mb-2">📂</p>
                <p class="text-sm">选择左侧配置文件以查看和编辑</p>
              </div>
            </template>
            <template v-else>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 font-mono">
                  {{ selectedFile.name }}
                </h3>
                <div class="flex items-center gap-2">
                  <button
                    class="px-2.5 py-1 text-xs font-medium text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 rounded transition-smooth"
                    @click="loadFileHistory(selectedFile.name)"
                  >历史记录</button>
                </div>
              </div>

              <div v-if="fileLoading" class="flex items-center justify-center py-12">
                <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
              </div>

              <template v-else>
                <!-- History View -->
                <div v-if="showHistory && !viewingHistoryContent" class="space-y-3 mb-4">
                  <div class="flex items-center justify-between">
                    <h4 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase">修改历史</h4>
                    <button
                      class="text-xs text-blue-500 hover:text-blue-700 dark:text-blue-400"
                      @click="showHistory = false"
                    >返回编辑</button>
                  </div>
                  <div v-if="fileHistories.length === 0" class="text-center py-4 text-slate-400 text-sm">暂无修改记录</div>
                  <div
                    v-for="h in fileHistories"
                    :key="h.id"
                    class="p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 cursor-pointer hover:border-blue-300 dark:hover:border-blue-700 transition-smooth"
                    @click="viewHistoryDiff(h)"
                  >
                    <div class="flex items-center justify-between mb-1">
                      <span class="text-xs font-medium text-slate-700 dark:text-slate-300">{{ h.changed_by }}</span>
                      <span class="text-[10px] text-slate-400">{{ formatTime(h.created_at) }}</span>
                    </div>
                    <p class="text-xs text-slate-500 dark:text-slate-400">{{ h.change_summary || '无修改说明' }}</p>
                  </div>
                </div>

                <!-- History Diff View -->
                <div v-if="viewingHistoryContent" class="mb-4">
                  <button
                    class="text-xs text-blue-500 hover:text-blue-700 dark:text-blue-400 mb-2 inline-block"
                    @click="viewingHistoryContent = null"
                  >← 返回历史列表</button>
                  <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
                    <div>
                      <label class="text-[10px] font-semibold text-slate-400 uppercase mb-1 block">修改前</label>
                      <pre class="text-[11px] font-mono p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-slate-700 dark:text-slate-300 max-h-96 overflow-auto whitespace-pre-wrap">{{ viewingHistoryContent.before || '(空文件)' }}</pre>
                    </div>
                    <div>
                      <label class="text-[10px] font-semibold text-slate-400 uppercase mb-1 block">修改后</label>
                      <pre class="text-[11px] font-mono p-3 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-slate-700 dark:text-slate-300 max-h-96 overflow-auto whitespace-pre-wrap">{{ viewingHistoryContent.after }}</pre>
                    </div>
                  </div>
                </div>

                <!-- Editor -->
                <div v-if="!showHistory || viewingHistoryContent" :class="{ 'opacity-40 pointer-events-none': !!viewingHistoryContent }">
                  <div class="mb-2">
                    <label class="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">修改说明</label>
                    <input
                      v-model="changeSummary"
                      type="text"
                      class="w-full px-3 py-1.5 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400"
                      placeholder="简要说明本次修改内容..."
                    />
                  </div>
                  <textarea
                    v-model="fileContent"
                    class="w-full h-80 px-3 py-3 text-sm font-mono border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-1 focus:ring-blue-400 resize-none"
                    spellcheck="false"
                  ></textarea>
                  <button
                    class="mt-3 px-4 py-2 text-sm font-medium text-white bg-blue-500 hover:bg-blue-600 rounded-btn transition-smooth disabled:opacity-50"
                    :disabled="fileSaving"
                    @click="saveConfigFile()"
                  >
                    {{ fileSaving ? '保存中...' : '保存配置文件' }}
                  </button>
                </div>
              </template>
            </template>
          </div>
        </div>
      </div>

      <!-- ===== 操作日志 Tab ===== -->
      <div v-if="activeTab === 'logs'" class="max-w-5xl">
        <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">系统管理操作记录</h3>

        <div v-if="logsLoading" class="flex items-center justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
        </div>

        <div v-else-if="adminLogs.length === 0" class="text-center py-12 text-slate-400 dark:text-slate-500">
          <p class="text-lg mb-2">📋</p>
          <p class="text-sm">暂无操作记录</p>
        </div>

        <div v-else>
          <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-slate-50 dark:bg-slate-800/50">
                  <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">操作时间</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">管理员</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">操作类型</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">详情</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">IP地址</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                <tr v-for="log in adminLogs" :key="log.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-smooth">
                  <td class="px-4 py-3 text-xs text-slate-600 dark:text-slate-400 whitespace-nowrap">{{ formatTime(log.created_at) }}</td>
                  <td class="px-4 py-3 text-xs font-medium text-slate-900 dark:text-slate-100">{{ log.admin_name }}</td>
                  <td class="px-4 py-3">
                    <span class="px-1.5 py-0.5 text-[10px] font-medium rounded bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-400">
                      {{ actionLabels[log.action] || log.action }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-xs text-slate-500 dark:text-slate-400 max-w-xs truncate">{{ log.detail }}</td>
                  <td class="px-4 py-3 text-xs text-slate-400 dark:text-slate-500 font-mono">{{ log.ip_address }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div class="flex items-center justify-between mt-4">
            <span class="text-xs text-slate-400 dark:text-slate-500">
              共 {{ logsTotal }} 条记录，第 {{ logsPage }} 页
            </span>
            <div class="flex items-center gap-2">
              <button
                class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
                :disabled="logsPage <= 1"
                @click="prevLogsPage()"
              >上一页</button>
              <button
                class="px-3 py-1 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded transition-smooth disabled:opacity-40"
                :disabled="logsPage * logsPageSize >= logsTotal"
                @click="nextLogsPage()"
              >下一页</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
