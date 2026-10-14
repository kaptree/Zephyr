<script setup lang="ts">
/* 工具运行页 —— 沙箱 iframe 加载工具入口 HTML：
   iframe 不授予 allow-same-origin（不透明源，无法访问主应用会话/DOM），
   服务端响应附带专用 CSP；进入时上报 launch（会话去重），离开时尽力回填时长。 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { BASE_URL } from '@/services/api'
import { fetchToolDetail, launchTool, exitTool, toolRunUrl } from '@/services/tools'
import type { ToolItem } from '@/services/tools'
import AppIcon from '@/components/icons/AppIcon.vue'
import PasswordVaultTool from '@/components/tools/PasswordVaultTool.vue'

const route = useRoute()
const router = useRouter()

const tool = ref<ToolItem | null>(null)
const loading = ref(true)
const errorMsg = ref('')

const slug = computed(() => String(route.params.slug || ''))
const runSrc = computed(() => `${BASE_URL}${toolRunUrl(slug.value)}`)
/** 原生渲染工具：密码本（含敏感数据与登录态交互，不走沙箱 iframe） */
const isNativeTool = computed(() => slug.value === 'password-vault')

/** 会话 ID：同一次进入（含页面刷新）只计一次使用，存 sessionStorage 按 slug 区分 */
const SESSION_KEY = 'tool_session'
function sessionKey(sid: string) {
  return `${SESSION_KEY}:${sid}`
}

const enteredAt = ref(0)
let sessionId = ''

function back() {
  if (window.history.length > 1) router.back()
  else router.push('/tools')
}

onMounted(async () => {
  try {
    const res = await fetchToolDetail(slug.value)
    tool.value = res.data ?? null
  } catch (e: unknown) {
    errorMsg.value =
      (e as { friendlyMessage?: string })?.friendlyMessage || '工具不存在或未上架'
  } finally {
    loading.value = false
  }

  if (!tool.value) return

  // 上报启动：同一会话（sessionStorage 未清除，如刷新）后端自动去重
  enteredAt.value = Date.now()
  try {
    const cached = sessionStorage.getItem(sessionKey(slug.value))
    const res = await launchTool(slug.value, cached || '')
    sessionId = res.data?.session_id || ''
    if (sessionId) sessionStorage.setItem(sessionKey(slug.value), sessionId)
  } catch {
    /* 统计失败不阻塞使用 */
  }
})

onBeforeUnmount(() => {
  if (!sessionId) return
  const duration = Math.round((Date.now() - enteredAt.value) / 1000)
  // 尽力上报：页面卸载场景下用 keepalive fetch 兜底
  try {
    fetch(`${BASE_URL}/api/v1/tools/${slug.value}/exit`, {
      method: 'POST',
      keepalive: true,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('auth_token') || ''}`,
      },
      body: JSON.stringify({ session_id: sessionId, duration }),
    }).catch(() => {})
  } catch {
    /* ignore */
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors duration-300 flex flex-col">
    <!-- 顶部：返回 + 工具信息 -->
    <div
      class="shrink-0 bg-white/90 dark:bg-slate-800/90 backdrop-blur border-b border-slate-200 dark:border-slate-700"
    >
      <div class="max-w-6xl mx-auto px-4 sm:px-6 py-3 flex items-center gap-3">
        <button
          class="shrink-0 w-9 h-9 rounded-xl flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
          aria-label="返回"
          @click="back"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div class="min-w-0">
          <template v-if="tool">
            <h1 class="text-base font-semibold text-slate-800 dark:text-slate-100 truncate">
              {{ tool.name }}
            </h1>
            <p class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ tool.description }}</p>
          </template>
          <div v-else class="h-9" />
        </div>
        <div class="ml-auto shrink-0 flex items-center gap-1.5 text-xs text-slate-400 dark:text-slate-500">
          <AppIcon name="check-circle" :size="14" class="text-green-500" />
          {{ isNativeTool ? '本地安全运行' : '沙箱运行中' }}
        </div>
      </div>
    </div>

    <!-- 工具主体 -->
    <div class="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 py-5">
      <div v-if="loading" class="h-[70vh] rounded-2xl bg-white dark:bg-slate-800 animate-pulse" />
      <div
        v-else-if="errorMsg"
        class="h-[50vh] rounded-2xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 flex flex-col items-center justify-center gap-3"
      >
        <AppIcon name="warning" :size="40" class="text-amber-400" />
        <p class="text-sm text-slate-500 dark:text-slate-400">{{ errorMsg }}</p>
        <button
          class="px-4 py-2 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition-colors"
          @click="back"
        >返回工具超市</button>
      </div>
      <!-- 密码本：原生渲染（沙箱 iframe 内无法携带登录态，且不宜承载敏感数据界面） -->
      <PasswordVaultTool v-else-if="isNativeTool" />
      <iframe
        v-else
        :src="runSrc"
        class="w-full h-[calc(100vh-150px)] min-h-[480px] rounded-2xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-sm"
        sandbox="allow-scripts allow-forms allow-modals allow-popups"
        referrerpolicy="no-referrer"
        title="工具运行区域"
      />
    </div>
  </div>
</template>
