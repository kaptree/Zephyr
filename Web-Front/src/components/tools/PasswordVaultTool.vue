<script setup lang="ts">
/* 密码本工具（原生渲染，非沙箱 iframe）：
   - 我的密码本：本人数据增删改查，密码默认掩码，一键复制
   - 密码广场：所有用户共享的条目，密码需单独查看（30 秒后自动隐藏）
   密码在服务端以 AES-GCM 加密落库；共享操作即时生效，可随时取消。 */
import { computed, onMounted, ref } from 'vue'
import { watch } from 'vue'
import {
  fetchVaultEntries,
  createVaultEntry,
  updateVaultEntry,
  deleteVaultEntry,
  setVaultEntryShared,
  fetchVaultPlaza,
  revealPlazaPassword,
} from '@/services/vault'
import type { VaultEntry, PlazaEntry } from '@/services/vault'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'

const toast = useToast()
const { confirm } = useConfirm()

const tab = ref<'mine' | 'plaza'>('mine')
const entries = ref<VaultEntry[]>([])
const plaza = ref<PlazaEntry[]>([])
const loading = ref(true)
const keyword = ref('')

/** 我的密码本：密码明文显示状态（数据本身在内存中，仅控制 UI） */
const visible = ref<Record<string, boolean>>({})
/** 密码广场：id → 已查看的明文（30 秒后自动清除） */
const plazaPwd = ref<Record<string, string>>({})
const plazaPwdTimer = ref<Record<string, ReturnType<typeof setTimeout>>>({})

// ===== 表单弹窗 =====
const showForm = ref(false)
const editing = ref<VaultEntry | null>(null)
const saving = ref(false)
const form = ref({ platform: '', website: '', account: '', password: '', notes: '' })

const filteredMine = computed(() => filterList(entries.value))
const filteredPlaza = computed(() =>
  filterList(plaza.value).map((p) => ({ ...p, _key: p.id })),
)

function filterList<T extends { platform: string; account: string; website?: string; notes?: string }>(
  list: T[],
): T[] {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return list
  return list.filter((e) =>
    [e.platform, e.account, e.website, e.notes].some((f) => (f || '').toLowerCase().includes(kw)),
  )
}

function errMsg(e: unknown, fallback: string) {
  return (e as { friendlyMessage?: string })?.friendlyMessage || fallback
}

async function loadAll() {
  loading.value = true
  try {
    const [mine, plazaRes] = await Promise.allSettled([fetchVaultEntries(), fetchVaultPlaza()])
    if (mine.status === 'fulfilled') entries.value = mine.value.data?.list || []
    if (plazaRes.status === 'fulfilled') plaza.value = plazaRes.value.data?.list || []
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
watch(tab, () => {
  keyword.value = ''
})

// ===== 复制 =====
async function copyText(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast.success(`${label}已复制`)
  } catch {
    toast.error('复制失败')
  }
}

// ===== 我的密码本 =====
function maskPwd(p: string) {
  return '•'.repeat(Math.min(Math.max(p.length, 6), 12))
}

function openCreate() {
  editing.value = null
  form.value = { platform: '', website: '', account: '', password: '', notes: '' }
  showForm.value = true
}

function openEdit(e: VaultEntry) {
  editing.value = e
  // 编辑时密码留空 = 保持原密码不变
  form.value = { platform: e.platform, website: e.website, account: e.account, password: '', notes: e.notes }
  showForm.value = true
}

async function submitForm() {
  const f = form.value
  if (!f.platform.trim() || !f.account.trim() || (!editing.value && !f.password)) {
    toast.warning('平台名称、账号、密码为必填项')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateVaultEntry(editing.value.id, { ...f })
      toast.success('条目已更新')
    } else {
      await createVaultEntry({ ...f })
      toast.success('条目已添加')
    }
    showForm.value = false
    await loadAll()
  } catch (e: unknown) {
    toast.error(errMsg(e, '保存失败'))
  } finally {
    saving.value = false
  }
}

async function removeEntry(e: VaultEntry) {
  if (
    !(await confirm({
      title: '删除密码条目',
      message: `确定删除「${e.platform}」的记录吗？删除后不可恢复。`,
      confirmText: '删除',
      danger: true,
    }))
  )
    return
  try {
    await deleteVaultEntry(e.id)
    entries.value = entries.value.filter((x) => x.id !== e.id)
    plaza.value = plaza.value.filter((x) => x.id !== e.id)
    delete visible.value[e.id]
    toast.success('已删除')
  } catch (err: unknown) {
    toast.error(errMsg(err, '删除失败'))
  }
}

async function toggleShare(e: VaultEntry) {
  const shared = !e.shared_at
  try {
    await setVaultEntryShared(e.id, shared)
    e.shared_at = shared ? new Date().toISOString() : null
    toast.success(shared ? '已共享到密码广场' : '已取消共享')
    // 广场含共享者信息，直接重新拉取保持一致
    if (shared) {
      const res = await fetchVaultPlaza()
      plaza.value = res.data?.list || []
    } else {
      plaza.value = plaza.value.filter((x) => x.id !== e.id)
    }
  } catch (err: unknown) {
    toast.error(errMsg(err, '共享状态更新失败'))
  }
}

// ===== 密码广场 =====
async function toggleReveal(p: PlazaEntry) {
  if (plazaPwd.value[p.id]) {
    delete plazaPwd.value[p.id]
    clearTimeout(plazaPwdTimer.value[p.id])
    return
  }
  try {
    const res = await revealPlazaPassword(p.id)
    plazaPwd.value[p.id] = res.data?.password || ''
    // 30 秒后自动隐藏，降低泄露风险
    clearTimeout(plazaPwdTimer.value[p.id])
    plazaPwdTimer.value[p.id] = setTimeout(() => delete plazaPwd.value[p.id], 30_000)
  } catch (e: unknown) {
    toast.error(errMsg(e, '查看密码失败'))
  }
}

function fmtTime(s: string) {
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div
    class="h-[calc(100vh-150px)] min-h-[480px] flex flex-col rounded-2xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-sm overflow-hidden"
  >
    <!-- 顶部：Tab + 搜索 + 新建 -->
    <div class="shrink-0 px-5 pt-4 pb-3 border-b border-slate-100 dark:border-slate-700">
      <div class="flex items-center gap-2 flex-wrap">
        <div class="flex rounded-xl bg-slate-100 dark:bg-slate-700/60 p-1">
          <button
            class="px-3.5 py-1.5 rounded-lg text-sm font-medium transition-colors"
            :class="tab === 'mine'
              ? 'bg-white dark:bg-slate-600 text-indigo-600 dark:text-indigo-300 shadow-sm'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'"
            @click="tab = 'mine'"
          >
            我的密码本
            <span v-if="entries.length" class="ml-0.5 text-xs opacity-70">{{ entries.length }}</span>
          </button>
          <button
            class="px-3.5 py-1.5 rounded-lg text-sm font-medium transition-colors"
            :class="tab === 'plaza'
              ? 'bg-white dark:bg-slate-600 text-indigo-600 dark:text-indigo-300 shadow-sm'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'"
            @click="tab = 'plaza'"
          >
            密码广场
            <span v-if="plaza.length" class="ml-0.5 text-xs opacity-70">{{ plaza.length }}</span>
          </button>
        </div>

        <div class="relative ml-auto w-44 sm:w-56">
          <svg
            class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"
            fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 10a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="keyword"
            type="text"
            placeholder="搜索平台 / 账号 / 备注"
            class="w-full pl-8 pr-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          />
        </div>

        <button
          v-if="tab === 'mine'"
          class="shrink-0 px-3.5 py-2 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition-colors"
          @click="openCreate"
        >
          + 新增条目
        </button>
      </div>
    </div>

    <!-- 列表区 -->
    <div class="flex-1 overflow-y-auto px-5 py-4">
      <div v-if="loading" class="grid sm:grid-cols-2 gap-3">
        <div v-for="i in 4" :key="i" class="h-40 rounded-xl bg-slate-100 dark:bg-slate-700/50 animate-pulse" />
      </div>

      <!-- 我的密码本 -->
      <template v-else-if="tab === 'mine'">
        <div v-if="!filteredMine.length" class="h-full flex flex-col items-center justify-center gap-2 text-center">
          <span class="text-4xl">🔐</span>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ keyword ? '没有匹配的条目' : '密码本还是空的，点击右上角「新增条目」开始记录' }}
          </p>
        </div>
        <div v-else class="grid sm:grid-cols-2 gap-3">
          <div
            v-for="e in filteredMine"
            :key="e.id"
            class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 hover:shadow-sm transition-shadow"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-semibold text-slate-800 dark:text-slate-100 truncate">{{ e.platform }}</span>
                  <span
                    v-if="e.shared_at"
                    class="shrink-0 px-1.5 py-0.5 rounded-full bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[10px]"
                  >已共享</span>
                </div>
                <a
                  v-if="e.website"
                  :href="e.website"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="block text-xs text-indigo-500 hover:underline truncate mt-0.5"
                >{{ e.website }}</a>
              </div>
              <button
                class="shrink-0 px-2 py-1 rounded-lg text-xs font-medium transition-colors"
                :class="e.shared_at
                  ? 'text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-500/10'
                  : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700'"
                @click="toggleShare(e)"
              >
                {{ e.shared_at ? '取消共享' : '共享' }}
              </button>
            </div>

            <div class="mt-3 space-y-1.5 text-sm">
              <div class="flex items-center gap-2">
                <span class="shrink-0 w-9 text-xs text-slate-400">账号</span>
                <span class="min-w-0 flex-1 truncate text-slate-700 dark:text-slate-200">{{ e.account }}</span>
                <button class="shrink-0 text-slate-400 hover:text-indigo-500" title="复制账号" @click="copyText(e.account, '账号')">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
              <div class="flex items-center gap-2">
                <span class="shrink-0 w-9 text-xs text-slate-400">密码</span>
                <span class="min-w-0 flex-1 truncate text-slate-700 dark:text-slate-200 font-mono">
                  {{ visible[e.id] ? e.password : maskPwd(e.password) }}
                </span>
                <button
                  class="shrink-0 text-slate-400 hover:text-indigo-500"
                  :title="visible[e.id] ? '隐藏密码' : '显示密码'"
                  @click="visible[e.id] = !visible[e.id]"
                >
                  <svg v-if="!visible[e.id]" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.542-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878l4.242 4.242M21 21l-3.759-3.759" />
                  </svg>
                </button>
                <button class="shrink-0 text-slate-400 hover:text-indigo-500" title="复制密码" @click="copyText(e.password, '密码')">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
              <p v-if="e.notes" class="text-xs text-slate-400 line-clamp-2">备注：{{ e.notes }}</p>
            </div>

            <div class="mt-3 pt-2 border-t border-slate-100 dark:border-slate-700 flex justify-end gap-1">
              <button
                class="px-2.5 py-1 rounded-lg text-xs text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                @click="openEdit(e)"
              >编辑</button>
              <button
                class="px-2.5 py-1 rounded-lg text-xs text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 transition-colors"
                @click="removeEntry(e)"
              >删除</button>
            </div>
          </div>
        </div>
      </template>

      <!-- 密码广场 -->
      <template v-else>
        <div v-if="!filteredPlaza.length" class="h-full flex flex-col items-center justify-center gap-2 text-center">
          <span class="text-4xl">🌐</span>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ keyword ? '没有匹配的条目' : '广场暂无共享条目，在自己的密码本中点击「共享」即可发布到这里' }}
          </p>
        </div>
        <div v-else class="grid sm:grid-cols-2 gap-3">
          <div
            v-for="p in filteredPlaza"
            :key="p._key"
            class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 hover:shadow-sm transition-shadow"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-semibold text-slate-800 dark:text-slate-100 truncate">{{ p.platform }}</span>
                  <span
                    class="shrink-0 px-1.5 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-500/10 text-indigo-500 dark:text-indigo-300 text-[10px]"
                  >{{ p.shared_by_name }} 共享</span>
                </div>
                <a
                  v-if="p.website"
                  :href="p.website"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="block text-xs text-indigo-500 hover:underline truncate mt-0.5"
                >{{ p.website }}</a>
              </div>
              <span class="shrink-0 text-[11px] text-slate-400">{{ fmtTime(p.shared_at) }}</span>
            </div>

            <div class="mt-3 space-y-1.5 text-sm">
              <div class="flex items-center gap-2">
                <span class="shrink-0 w-9 text-xs text-slate-400">账号</span>
                <span class="min-w-0 flex-1 truncate text-slate-700 dark:text-slate-200">{{ p.account }}</span>
                <button class="shrink-0 text-slate-400 hover:text-indigo-500" title="复制账号" @click="copyText(p.account, '账号')">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
              <div class="flex items-center gap-2">
                <span class="shrink-0 w-9 text-xs text-slate-400">密码</span>
                <span class="min-w-0 flex-1 truncate text-slate-700 dark:text-slate-200 font-mono">
                  {{ plazaPwd[p.id] ? plazaPwd[p.id] : '••••••' }}
                </span>
                <button
                  class="shrink-0 px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors"
                  :class="plazaPwd[p.id]
                    ? 'text-amber-600 dark:text-amber-400 hover:bg-amber-50 dark:hover:bg-amber-500/10'
                    : 'text-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-500/10'"
                  @click="toggleReveal(p)"
                >
                  {{ plazaPwd[p.id] ? '隐藏' : '查看密码' }}
                </button>
                <button v-if="plazaPwd[p.id]" class="shrink-0 text-slate-400 hover:text-indigo-500" title="复制密码" @click="copyText(plazaPwd[p.id], '密码')">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
              <p v-if="p.notes" class="text-xs text-slate-400 line-clamp-2">备注：{{ p.notes }}</p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- 底部安全提示 -->
    <div class="shrink-0 px-5 py-2.5 border-t border-slate-100 dark:border-slate-700 text-[11px] text-slate-400 dark:text-slate-500">
      🔐 密码在服务端以 AES-GCM 加密存储，仅本人可见；共享到广场后所有登录用户可见，请谨慎共享敏感账号。
    </div>
  </div>

  <!-- 新增 / 编辑弹窗 -->
  <div v-if="showForm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showForm = false" />
    <div class="relative w-full max-w-md rounded-2xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-xl p-5">
      <h3 class="text-base font-semibold text-slate-800 dark:text-slate-100">
        {{ editing ? '编辑条目' : '新增条目' }}
      </h3>
      <div class="mt-4 space-y-3">
        <div>
          <label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">平台名称 <span class="text-red-400">*</span></label>
          <input
            v-model="form.platform"
            type="text"
            placeholder="例如：GitHub"
            class="w-full px-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          />
        </div>
        <div>
          <label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">网址</label>
          <input
            v-model="form.website"
            type="text"
            placeholder="https://example.com"
            class="w-full px-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          />
        </div>
        <div>
          <label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">账号 <span class="text-red-400">*</span></label>
          <input
            v-model="form.account"
            type="text"
            placeholder="用户名 / 邮箱 / 手机号"
            class="w-full px-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          />
        </div>
        <div>
          <label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">
            密码 <span v-if="!editing" class="text-red-400">*</span>
            <span v-else class="text-slate-400">（留空表示不修改）</span>
          </label>
          <input
            v-model="form.password"
            type="password"
            :placeholder="editing ? '留空保持原密码' : '请输入密码'"
            autocomplete="new-password"
            class="w-full px-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50"
          />
        </div>
        <div>
          <label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">备注</label>
          <textarea
            v-model="form.notes"
            rows="2"
            placeholder="其他补充信息"
            class="w-full px-3 py-2 rounded-xl text-sm bg-slate-50 dark:bg-slate-700/60 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-400/50 resize-none"
          />
        </div>
      </div>
      <div class="mt-5 flex justify-end gap-2">
        <button
          class="px-4 py-2 rounded-xl text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
          @click="showForm = false"
        >取消</button>
        <button
          class="px-4 py-2 rounded-xl text-sm font-medium bg-indigo-500 hover:bg-indigo-600 disabled:opacity-60 text-white transition-colors"
          :disabled="saving"
          @click="submitForm"
        >
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </div>
  </div>
</template>
