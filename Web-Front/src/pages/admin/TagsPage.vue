<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchTags } from '@/services/tags'
import type { Tag } from '@/types'

const tags = ref<Tag[]>([])
const loading = ref(false)
const loadError = ref('')
const showNewModal = ref(false)
const newTagName = ref('')
const newTagColor = ref('#3B82F6')

const colorOptions = ['#EF4444', '#F97316', '#EAB308', '#22C55E', '#14B8A6', '#3B82F6', '#8B5CF6', '#EC4899', '#78716C', '#64748B', '#94A3B8', '#475569']

onMounted(async () => {
  loading.value = true
  try {
    const res = await fetchTags()
    tags.value = res.data as unknown as Tag[]
  } catch {
    loadError.value = '加载标签失败'
  } finally {
    loading.value = false
  }
})

function addTag() {
  if (!newTagName.value.trim()) return
  tags.value.unshift({
    id: 'tag-new-' + Date.now(),
    name: newTagName.value.trim(),
    color: newTagColor.value,
    scope: 'system',
    category: '自定义',
    usage_count: 0,
  })
  newTagName.value = ''
  showNewModal.value = false
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">标签库管理</h2>
      <button class="btn-primary text-sm" @click="showNewModal = true">新建标签</button>
    </div>

    <div v-if="loading" class="grid grid-cols-4 gap-4">
      <div v-for="n in 8" :key="n" class="skeleton h-16 rounded-card" />
    </div>

    <div v-else-if="loadError" class="text-center py-8 text-sm text-red-400">{{ loadError }}</div>

    <div v-else class="grid grid-cols-4 gap-4">
      <div
        v-for="tag in tags"
        :key="tag.id"
        class="bg-white rounded-card border border-slate-100 p-4 flex items-center gap-3 hover:shadow-note transition-smooth"
      >
        <span class="w-4 h-4 rounded-full shrink-0" :style="{ backgroundColor: tag.color }" />
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-slate-900 truncate">{{ tag.name }}</div>
          <div class="text-xs text-slate-400">{{ tag.category }} · {{ tag.scope === 'system' ? '系统' : '个人' }}</div>
        </div>
        <span class="text-xs text-slate-400 shrink-0">{{ tag.usage_count }}次</span>
      </div>
    </div>

    <!-- 新建标签模态框 -->
    <Teleport to="body">
      <div v-if="showNewModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showNewModal = false" />
        <div class="relative bg-white rounded-card shadow-modal w-full max-w-sm mx-4 p-6 animate-fade-in">
          <h3 class="text-base font-semibold text-slate-900 mb-4">新建标签</h3>
          <form @submit.prevent="addTag" class="space-y-4">
            <input v-model="newTagName" class="input-field" placeholder="标签名称" autofocus />
            <div>
              <span class="text-xs text-slate-500 mb-2 block">颜色</span>
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
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" class="btn-secondary text-xs !py-1.5 !px-4" @click="showNewModal = false">取消</button>
              <button type="submit" class="btn-primary text-xs !py-1.5 !px-4">创建</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
