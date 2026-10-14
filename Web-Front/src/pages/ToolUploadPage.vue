<script setup lang="ts">
/* 上传工具 —— 用户自助上架 HTML 小工具：
   选择 .html 文件（≤2MB），可选填写名称/简介/分类/标签与图标（≤1MB）；
   未填写的字段由后端解析 HTML 元信息自动补全；提交后进入待审核。 */
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadTool } from '@/services/tools'
import { useToast } from '@/composables/useToast'
import SubmitButton from '@/components/common/SubmitButton.vue'
import AppIcon from '@/components/icons/AppIcon.vue'

const router = useRouter()
const toast = useToast()

const file = ref<File | null>(null)
const icon = ref<File | null>(null)
const name = ref('')
const description = ref('')
const category = ref('')
const tags = ref('')
const submitting = ref(false)

const CATEGORIES = ['开发工具', '文本处理', '换算工具', '效率工具', '其他']

const htmlInput = ref<HTMLInputElement | null>(null)
const iconInput = ref<HTMLInputElement | null>(null)

function pickFile(e: Event, target: 'html' | 'icon') {
  const input = e.target as HTMLInputElement
  const f = input.files?.[0] || null
  if (!f) return
  if (target === 'html') {
    const ext = f.name.toLowerCase().split('.').pop()
    if (ext !== 'html' && ext !== 'htm') {
      toast.error('仅支持 .html / .htm 文件')
      input.value = ''
      return
    }
    if (f.size > 2 * 1024 * 1024) {
      toast.error('HTML 文件最大允许 2MB')
      input.value = ''
      return
    }
    file.value = f
    // 名称留空时用文件名预填
    if (!name.value) name.value = f.name.replace(/\.(html?|htm)$/i, '')
  } else {
    if (f.size > 1024 * 1024) {
      toast.error('图标最大允许 1MB')
      input.value = ''
      return
    }
    icon.value = f
  }
}

function clearFile() {
  file.value = null
  if (htmlInput.value) htmlInput.value.value = ''
}

function clearIcon() {
  icon.value = null
  if (iconInput.value) iconInput.value.value = ''
}

function formatSize(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)}MB`
  return `${Math.max(1, Math.round(bytes / 1024))}KB`
}

function previewUrl(f: File): string {
  return URL.createObjectURL(f)
}

const canSubmit = computed(() => !!file.value && !submitting.value)

async function submit() {
  if (!file.value) return
  submitting.value = true
  try {
    await uploadTool({
      file: file.value,
      icon: icon.value || undefined,
      name: name.value.trim() || undefined,
      description: description.value.trim() || undefined,
      category: category.value || undefined,
      tags: tags.value.trim() || undefined,
    })
    toast.success('上传成功，等待管理员审核')
    router.push('/tools')
  } catch (e: unknown) {
    toast.error((e as { friendlyMessage?: string })?.friendlyMessage || '上传失败')
  } finally {
    submitting.value = false
  }
}

function back() {
  if (window.history.length > 1) router.back()
  else router.push('/tools')
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors duration-300">
    <div class="max-w-2xl mx-auto px-4 sm:px-6 py-8">
      <!-- 顶部 -->
      <div class="flex items-center gap-3 mb-6">
        <button
          class="shrink-0 w-9 h-9 rounded-xl flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
          aria-label="返回"
          @click="back"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div>
          <h1 class="text-xl font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
            <AppIcon name="grid" :size="20" class="text-indigo-500 dark:text-indigo-400" />
            上传工具
          </h1>
          <p class="mt-0.5 text-xs text-slate-400 dark:text-slate-500">
            上传一个自包含的 HTML 页面，审核通过后即可在工具超市使用
          </p>
        </div>
      </div>

      <form
        class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 p-6 space-y-5"
        @submit.prevent="submit"
      >
        <!-- HTML 文件 -->
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">
            工具页面 <span class="text-red-400">*</span>
            <span class="ml-1 text-xs font-normal text-slate-400">.html 文件，≤2MB，样式与脚本需内联</span>
          </label>
          <div
            v-if="!file"
            class="border-2 border-dashed border-slate-300 dark:border-slate-600 rounded-xl py-8 flex flex-col items-center justify-center gap-2 cursor-pointer hover:border-indigo-400 dark:hover:border-indigo-500 transition-colors"
            @click="htmlInput?.click()"
          >
            <svg class="w-8 h-8 text-slate-300 dark:text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 16.5V9.75m0 0l3 3m-3-3l-3 3M6.75 19.5a4.5 4.5 0 01-1.41-8.775 5.25 5.25 0 0110.233-2.33 3 3 0 013.758 3.848A3.752 3.752 0 0118 19.5H6.75z" />
            </svg>
            <p class="text-sm text-slate-400">点击选择 HTML 文件</p>
          </div>
          <div
            v-else
            class="flex items-center gap-3 bg-slate-50 dark:bg-slate-700/50 rounded-xl px-4 py-3"
          >
            <AppIcon name="check-circle" :size="20" class="text-green-500 shrink-0" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-slate-700 dark:text-slate-200 truncate">{{ file.name }}</p>
              <p class="text-xs text-slate-400">{{ formatSize(file.size) }}</p>
            </div>
            <button
              type="button"
              class="text-xs text-slate-400 hover:text-red-500 transition-colors"
              @click="clearFile"
            >移除</button>
          </div>
          <input ref="htmlInput" type="file" accept=".html,.htm" class="hidden" @change="pickFile($event, 'html')" />
        </div>

        <!-- 名称 / 简介 -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">
              工具名称
              <span class="ml-1 text-xs font-normal text-slate-400">留空则取页面标题</span>
            </label>
            <input
              v-model="name"
              type="text"
              maxlength="100"
              placeholder="如：JSON 格式化"
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-50 dark:bg-slate-700/50 border border-slate-200 dark:border-slate-600 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">分类</label>
            <select
              v-model="category"
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-50 dark:bg-slate-700/50 border border-slate-200 dark:border-slate-600 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
            >
              <option value="">默认（其他）</option>
              <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
            </select>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">
            简介
            <span class="ml-1 text-xs font-normal text-slate-400">留空则取页面 meta 描述</span>
          </label>
          <textarea
            v-model="description"
            rows="2"
            maxlength="500"
            placeholder="一句话介绍这个工具的功能"
            class="w-full px-3.5 py-2.5 rounded-xl bg-slate-50 dark:bg-slate-700/50 border border-slate-200 dark:border-slate-600 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50 resize-none"
          />
        </div>

        <!-- 标签 / 图标 -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">
              标签
              <span class="ml-1 text-xs font-normal text-slate-400">逗号分隔，便于搜索</span>
            </label>
            <input
              v-model="tags"
              type="text"
              maxlength="200"
              placeholder="如：json,格式化"
              class="w-full px-3.5 py-2.5 rounded-xl bg-slate-50 dark:bg-slate-700/50 border border-slate-200 dark:border-slate-600 text-sm text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">
              图标
              <span class="ml-1 text-xs font-normal text-slate-400">图片 ≤1MB，留空取名称首字</span>
            </label>
            <div v-if="!icon" class="flex items-center gap-3">
              <button
                type="button"
                class="px-3.5 py-2.5 rounded-xl bg-slate-50 dark:bg-slate-700/50 border border-dashed border-slate-300 dark:border-slate-600 text-sm text-slate-500 dark:text-slate-400 hover:border-indigo-400 transition-colors"
                @click="iconInput?.click()"
              >选择图片</button>
            </div>
            <div v-else class="flex items-center gap-3 bg-slate-50 dark:bg-slate-700/50 rounded-xl px-3.5 py-2">
              <img :src="previewUrl(icon)" alt="图标预览" class="w-8 h-8 rounded-lg object-cover" />
              <span class="text-xs text-slate-500 dark:text-slate-400 truncate flex-1">{{ icon.name }}</span>
              <button
                type="button"
                class="text-xs text-slate-400 hover:text-red-500 transition-colors"
                @click="clearIcon"
              >移除</button>
            </div>
            <input ref="iconInput" type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="hidden" @change="pickFile($event, 'icon')" />
          </div>
        </div>

        <!-- 安全提示 -->
        <div class="flex items-start gap-2 text-xs text-slate-400 dark:text-slate-500 bg-slate-50 dark:bg-slate-700/40 rounded-xl px-4 py-3">
          <AppIcon name="info" :size="16" class="shrink-0 mt-0.5 text-slate-400" />
          <p>
            上传的工具将在隔离沙箱中运行：外部脚本、页面嵌套等危险内容会被自动移除，
            工具无法访问工作台的登录态与数据。审核通过后全局可见。
          </p>
        </div>

        <!-- 提交 -->
        <div class="flex justify-end gap-3">
          <button
            type="button"
            class="px-5 py-2.5 rounded-xl text-sm font-medium text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
            @click="back"
          >取消</button>
          <SubmitButton
            :state="submitting ? 'loading' : 'idle'"
            label="提交审核"
            loading-label="提交中..."
            :disabled="!canSubmit"
            type="submit"
            @click="submit"
          />
        </div>
      </form>
    </div>
  </div>
</template>
