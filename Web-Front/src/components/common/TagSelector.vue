<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Tag } from '@/types'
import { fetchTags } from '@/services/tags'

const props = withDefaults(defineProps<{
  modelValue: string[]
  max?: number
  scope?: 'personal' | 'system' | 'all'
}>(), {
  max: 10,
  scope: 'all',
})

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
  'create-tag': [name: string]
}>()

const open = ref(false)
const searchText = ref('')
const allTags = ref<Tag[]>([])
const loading = ref(false)
const loadError = ref('')

const selectedTags = computed(() =>
  allTags.value.filter(t => props.modelValue.includes(t.id))
)

const recentTags = computed(() => {
  const recent = JSON.parse(localStorage.getItem('recent_tags') || '[]') as string[]
  return recent.map((id: string) => allTags.value.find(t => t.id === id)).filter(Boolean) as Tag[]
})

const filteredTags = computed(() => {
  if (!searchText.value) return allTags.value
  const q = searchText.value.toLowerCase()
  return allTags.value.filter(t => t.name.toLowerCase().includes(q))
})

onMounted(async () => {
  loading.value = true
  try {
    const res = await fetchTags(props.scope)
    allTags.value = res.data as unknown as Tag[]
  } catch {
    loadError.value = '加载标签失败'
  } finally {
    loading.value = false
  }
})

function toggleTag(tagId: string) {
  const current = [...props.modelValue]
  const idx = current.indexOf(tagId)
  if (idx >= 0) {
    current.splice(idx, 1)
  } else if (current.length < props.max) {
    current.push(tagId)
  }
  emit('update:modelValue', current)
}

function removeTag(tagId: string) {
  const current = props.modelValue.filter(id => id !== tagId)
  emit('update:modelValue', current)
}

function isSelected(tagId: string): boolean {
  return props.modelValue.includes(tagId)
}

function handleCreateTag() {
  const name = searchText.value.trim()
  if (name) {
    emit('create-tag', name)
    searchText.value = ''
  }
}
</script>

<template>
  <div class="relative">
    <div class="flex flex-wrap gap-1.5 mb-1.5">
      <span
        v-for="tag in selectedTags"
        :key="tag.id"
        class="tag-capsule text-white"
        :style="{ backgroundColor: tag.color || '#64748B' }"
      >
        {{ tag.name }}
        <button class="ml-1 hover:opacity-70" @click="removeTag(tag.id)">&times;</button>
      </span>
      <button
        class="tag-capsule border border-dashed border-slate-300 text-slate-400 hover:border-slate-400 transition-smooth text-xs"
        @click="open = !open"
      >
        + 添加标签
      </button>
    </div>

    <div
      v-if="open"
      class="absolute top-full left-0 mt-1 w-72 bg-white rounded-card shadow-modal border border-slate-100 z-50 overflow-hidden"
    >
      <div class="p-3 border-b border-slate-100">
        <input
          v-model="searchText"
          class="input-field !text-xs"
          placeholder="搜索标签..."
          @keyup.enter="handleCreateTag"
        />
      </div>

      <div v-if="recentTags.length && !searchText" class="px-3 pt-2">
        <span class="text-[10px] text-slate-400 uppercase">最近使用</span>
        <div class="flex flex-wrap gap-1 mt-1">
          <span
            v-for="tag in recentTags"
            :key="tag.id"
            :class="[
              'tag-capsule cursor-pointer text-xs transition-smooth',
              isSelected(tag.id) ? 'ring-2 ring-offset-1 ring-blue-400' : ''
            ]"
            :style="{ backgroundColor: isSelected(tag.id) ? (tag.color || '#64748B') : '#F1F5F9', color: isSelected(tag.id) ? '#fff' : '#475569' }"
            @click="toggleTag(tag.id)"
          >
            {{ tag.name }}
          </span>
        </div>
      </div>

      <div class="max-h-48 overflow-y-auto scrollbar-thin px-3 py-2">
        <div v-if="loading" class="text-center py-4 text-xs text-slate-400">加载中...</div>
        <div v-else-if="loadError" class="text-center py-4 text-xs text-red-400">{{ loadError }}</div>
        <div v-else-if="filteredTags.length === 0" class="text-center py-4">
          <p class="text-xs text-slate-400">暂无匹配标签</p>
          <button
            v-if="searchText.trim()"
            class="text-xs text-[#3B82F6] hover:underline mt-1"
            @click="handleCreateTag"
          >
            创建 "{{ searchText }}"
          </button>
        </div>
        <div v-else class="space-y-1">
          <button
            v-for="tag in filteredTags"
            :key="tag.id"
            :class="[
              'w-full flex items-center gap-2 px-3 py-2 rounded-btn text-sm text-left transition-smooth',
              isSelected(tag.id) ? 'bg-blue-50' : 'hover:bg-slate-50'
            ]"
            @click="toggleTag(tag.id)"
          >
            <span
              class="w-3 h-3 rounded-full shrink-0"
              :style="{ backgroundColor: tag.color || '#64748B' }"
            />
            <span class="flex-1 truncate">{{ tag.name }}</span>
            <span v-if="isSelected(tag.id)" class="text-xs text-[#3B82F6]">✓</span>
          </button>
        </div>
      </div>

      <div class="border-t border-slate-100 p-2 flex justify-end">
        <button
          class="text-xs px-3 py-1.5 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
          @click="open = false"
        >
          完成
        </button>
      </div>
    </div>
  </div>
</template>
