<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchTags, createTag, deleteTag } from '@/services/tags'
import type { Tag } from '@/types'

const tags = ref<Tag[]>([])
const loading = ref(false)
const loadError = ref('')
const showNewModal = ref(false)
const newTagName = ref('')
const newTagSubTag = ref('')
const newTagColor = ref('#3B82F6')
const creating = ref(false)
const createError = ref('')
const deletingId = ref<string | null>(null)

const colorOptions = ['#EF4444', '#F97316', '#EAB308', '#22C55E', '#14B8A6', '#3B82F6', '#8B5CF6', '#EC4899', '#78716C', '#64748B', '#94A3B8', '#475569']

// 随机挑选标签颜色：倾向当前出现次数较少的颜色，避免同色标签堆积
function randomTagColor(existingColors: string[] = []): string {
  const countMap = new Map<string, number>();
  for (const c of colorOptions) countMap.set(c, 0);
  for (const c of existingColors) {
    if (countMap.has(c)) countMap.set(c, (countMap.get(c) || 0) + 1);
  }
  let min = Infinity;
  for (const n of countMap.values()) min = Math.min(min, n);
  const candidates = colorOptions.filter((c) => countMap.get(c) === min);
  return candidates[Math.floor(Math.random() * candidates.length)];
}

function openNewModal() {
  newTagName.value = ''
  newTagSubTag.value = ''
  newTagColor.value = randomTagColor(tags.value.map((t) => t.color))
  createError.value = ''
  showNewModal.value = true
}

async function loadTags() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await fetchTags()
    tags.value = res.data as unknown as Tag[]
  } catch {
    loadError.value = '加载标签失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadTags)

async function addTag() {
  if (!newTagName.value.trim()) return
  creating.value = true
  createError.value = ''
  try {
    await createTag({
      name: newTagName.value.trim(),
      sub_tag: newTagSubTag.value.trim() || undefined,
      color: newTagColor.value,
      category: '自定义',
      scope: 'system',
    })
    newTagName.value = ''
    newTagSubTag.value = ''
    showNewModal.value = false
    await loadTags()
  } catch {
    createError.value = '创建标签失败，请重试'
  } finally {
    creating.value = false
  }
}

async function handleDelete(tag: Tag) {
  if (!confirm(`确定删除标签"${tag.name}${tag.sub_tag ? ' › ' + tag.sub_tag : ''}"吗？\n删除后所有任务上的该标签将一并移除。`)) return
  deletingId.value = tag.id
  try {
    await deleteTag(tag.id)
    await loadTags()
  } catch {
    alert('删除标签失败，请重试')
  } finally {
    deletingId.value = null
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">标签库管理</h2>
      <button class="btn-primary text-sm" @click="openNewModal">新建标签</button>
    </div>

    <div v-if="loading" class="grid grid-cols-4 gap-4">
      <div v-for="n in 8" :key="n" class="skeleton h-16 rounded-card" />
    </div>

    <div v-else-if="loadError" class="text-center py-8 text-sm text-red-400">{{ loadError }}</div>

    <div v-else-if="tags.length === 0" class="text-center py-16 text-slate-400 dark:text-slate-500">
      <p class="text-2xl mb-2">🏷️</p>
      <p class="text-sm">暂无标签，点击右上角「新建标签」创建</p>
    </div>

    <div v-else class="grid grid-cols-4 gap-4">
      <div
        v-for="tag in tags"
        :key="tag.id"
        class="group bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-4 flex items-center gap-3 hover:shadow-note transition-smooth"
      >
        <span class="w-4 h-4 rounded-full shrink-0" :style="{ backgroundColor: tag.color }" />
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
            {{ tag.name }}
            <span v-if="tag.sub_tag" class="text-xs text-slate-400 dark:text-slate-500 ml-1">› {{ tag.sub_tag }}</span>
          </div>
          <div class="text-xs text-slate-400 dark:text-slate-400">{{ tag.category }} · {{ tag.scope === 'system' ? '系统' : '个人' }}</div>
        </div>
        <span :class="['text-xs shrink-0 font-medium', tag.usage_count > 5 ? 'text-blue-600 dark:text-blue-400' : 'text-slate-400 dark:text-slate-400']">
          {{ tag.usage_count }}次
        </span>
        <button
          class="shrink-0 w-6 h-6 flex items-center justify-center rounded text-slate-300 dark:text-slate-600 hover:text-red-500 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/40 transition-smooth disabled:opacity-50"
          title="删除标签"
          :disabled="deletingId === tag.id"
          @click="handleDelete(tag)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 新建标签模态框 -->
    <Teleport to="body">
      <div v-if="showNewModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showNewModal = false" />
        <div class="relative z-50 bg-white dark:bg-slate-800 rounded-card shadow-modal w-full max-w-sm mx-4 p-6 animate-fade-in">
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">新建标签</h3>
          <form @submit.prevent="addTag" class="space-y-4">
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">一级标签名称</span>
              <input v-model="newTagName" class="input-field" placeholder="一级标签名称" autofocus />
            </div>
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">二级标签（可选）</span>
              <input v-model="newTagSubTag" class="input-field" placeholder="二级标签（可选）" />
            </div>
            <div>
              <span class="text-xs text-slate-500 dark:text-slate-400 mb-2 block">颜色</span>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="c in colorOptions"
                  :key="c"
                  type="button"
                  class="w-7 h-7 rounded-full transition-smooth"
                  :class="newTagColor === c ? 'ring-2 ring-offset-2 ring-blue-400 scale-110' : 'hover:scale-105'"
                  :style="{ backgroundColor: c }"
                  @click="newTagColor = c"
                />
              </div>
            </div>
            <p v-if="createError" class="text-xs text-red-500 dark:text-red-400">{{ createError }}</p>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" class="btn-secondary text-xs !py-1.5 !px-4" @click="showNewModal = false">取消</button>
              <button type="submit" class="btn-primary text-xs !py-1.5 !px-4" :disabled="creating || !newTagName.trim()">
                {{ creating ? '创建中...' : '创建' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
