<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useNoteStore } from '@/stores/notes'
import type { Note } from '@/types'

const noteStore = useNoteStore()
const showCreateModal = ref(false)
const showDetailPanel = ref(false)
const selectedNote = ref<Note | null>(null)

onMounted(() => {
  noteStore.fetchNotes()
})

function openCreateModal() {
  showCreateModal.value = true
}

function openDetail(note: Note) {
  selectedNote.value = note
  showDetailPanel.value = true
}

function closeDetail() {
  showDetailPanel.value = false
  selectedNote.value = null
}

function getNoteClass(note: Note) {
  if (note.priority === 'urgent') return 'card-note-red'
  if (note.status === 'completed') return 'card-note-green'
  return 'card-note-yellow'
}

function displayTags(note: Note) {
  const max = 2
  const visible = note.tags.slice(0, max)
  const remaining = note.tags.length - max
  return { visible, remaining }
}
</script>

<template>
  <div class="relative min-h-full">
    <!-- 状态筛选栏 -->
    <div class="flex items-center gap-3 mb-6">
      <button
        v-for="tab in [
          { label: '全部', value: undefined },
          { label: '待办', value: 'active' },
          { label: '盯办', value: 'urgent' },
          { label: '已完成', value: 'completed' }
        ]"
        :key="tab.label"
        :class="[
          'px-4 py-1.5 rounded-btn text-sm font-medium transition-smooth',
          noteStore.filters.status === tab.value
            ? 'bg-[#3B82F6] text-white'
            : 'bg-white text-slate-600 hover:bg-slate-50 border border-slate-200'
        ]"
        @click="noteStore.fetchNotes({ status: tab.value as NoteStatus | undefined })"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="noteStore.loading && noteStore.activeNotes.length === 0" class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <div v-for="n in 6" :key="n" class="skeleton h-44 rounded-card" />
    </div>

    <!-- 空态 -->
    <div v-else-if="!noteStore.loading && noteStore.activeNotes.length === 0" class="flex flex-col items-center justify-center py-24">
      <div class="w-24 h-24 bg-slate-100 rounded-3xl flex items-center justify-center mb-6">
        <svg class="w-12 h-12 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      </div>
      <p class="text-slate-400 text-sm">点击右下角 '+' 新建便签</p>
    </div>

    <!-- 便签墙 -->
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <div
        v-for="note in noteStore.activeNotes"
        :key="note.id"
        :class="[
          'p-5 rounded-card shadow-note transition-smooth cursor-pointer relative',
          getNoteClass(note)
        ]"
        style="animation: spring-enter 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards"
        @click="openDetail(note)"
        @mouseenter="$event.currentTarget.style.transform = 'translateY(-2px)'; $event.currentTarget.style.boxShadow = '0 8px 32px -8px rgba(0,0,0,0.12)'"
        @mouseleave="$event.currentTarget.style.transform = ''; $event.currentTarget.style.boxShadow = ''"
      >
        <!-- 盯办徽章 -->
        <span v-if="note.priority === 'urgent'" class="badge-corner bg-red-500 text-white">
          盯办
        </span>
        <!-- 已完成角标 -->
        <span v-if="note.status === 'completed'" class="badge-corner bg-green-500 text-white">
          已完成
        </span>

        <h3 class="text-base font-semibold text-slate-900 mb-2 line-clamp-1">{{ note.title || '无标题' }}</h3>
        <p class="text-sm text-slate-500 line-clamp-3 note-content-mask">{{ note.content || '暂无内容' }}</p>

        <!-- 标签区 -->
        <div v-if="note.tags.length" class="flex items-center gap-1.5 mt-3 flex-wrap">
          <span
            v-for="tag in displayTags(note).visible"
            :key="tag.id"
            class="tag-capsule text-white text-[11px]"
            :style="{ backgroundColor: tag.color || '#64748B' }"
          >
            {{ tag.name }}
          </span>
          <span v-if="displayTags(note).remaining > 0" class="text-xs text-slate-400">
            +{{ displayTags(note).remaining }}
          </span>
        </div>

        <!-- 底部信息 -->
        <div class="flex items-center justify-between mt-4 pt-3 border-t border-slate-200/50">
          <span class="text-xs text-slate-400">{{ note.created_at?.slice(0, 10) }}</span>
          <span v-if="note.due_time" class="text-xs text-slate-400">截止 {{ note.due_time.slice(0, 10) }}</span>
        </div>
      </div>
    </div>

    <!-- 悬浮新建按钮 -->
    <button
      class="fixed right-8 bottom-8 w-14 h-14 rounded-full bg-[#3B82F6] text-white shadow-btn-float hover:bg-blue-600 active:scale-95 transition-smooth flex items-center justify-center z-30"
      @click="openCreateModal"
    >
      <svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
      </svg>
    </button>

    <!-- 新建便签模态框 -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="overlay-backdrop" @click="showCreateModal = false" />
        <div class="relative bg-white rounded-card shadow-modal w-full max-w-lg mx-4 animate-fade-in">
          <div class="p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-semibold text-slate-900">新建便签</h2>
              <button class="p-1 rounded-lg hover:bg-slate-100 transition-smooth" @click="showCreateModal = false">
                <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <form class="space-y-4" @submit.prevent>
              <input class="input-field" placeholder="便签标题" autofocus />
              <textarea class="input-field h-32 resize-none" placeholder="便签内容..." />

              <div class="flex items-center gap-2">
                <span class="tag-capsule bg-blue-100 text-blue-700">
                  标签1
                  <button class="ml-1 hover:text-blue-900">&times;</button>
                </span>
                <button class="tag-capsule border border-dashed border-slate-300 text-slate-400 hover:border-slate-400 transition-smooth">
                  + 添加标签
                </button>
              </div>

              <div class="flex gap-4 text-sm">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="type" checked class="w-4 h-4 text-[#3B82F6]" />
                  <span class="text-slate-700">仅自己</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="type" class="w-4 h-4 text-[#3B82F6]" />
                  <span class="text-slate-700">指派他人</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" name="type" class="w-4 h-4 text-[#3B82F6]" />
                  <span class="text-slate-700">开启协同</span>
                </label>
              </div>

              <div class="flex justify-end gap-3 pt-4 border-t border-slate-100">
                <button class="px-5 py-2.5 text-sm text-slate-600 bg-slate-100 rounded-btn hover:bg-slate-200 transition-smooth" @click="showCreateModal = false">
                  取消
                </button>
                <button class="px-5 py-2.5 text-sm text-white bg-[#3B82F6] rounded-btn hover:bg-blue-600 transition-smooth">
                  创建便签
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 详情侧滑面板 -->
    <Teleport to="body">
      <div v-if="showDetailPanel && selectedNote">
        <div class="overlay-backdrop" @click="closeDetail" />
        <div class="slide-panel">
          <div class="p-6 h-full flex flex-col">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-semibold text-slate-900">便签详情</h2>
              <button class="p-1 rounded-lg hover:bg-slate-100 transition-smooth" @click="closeDetail">
                <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div class="flex-1 overflow-auto space-y-5">
              <input
                class="input-field text-lg font-semibold"
                :value="selectedNote.title"
                placeholder="标题"
              />
              <div class="min-h-[200px] p-3 border border-slate-200 rounded-btn text-sm text-slate-700">
                {{ selectedNote.content || '暂无内容' }}
              </div>

              <div v-if="selectedNote.tags.length" class="flex flex-wrap gap-2">
                <span
                  v-for="tag in selectedNote.tags"
                  :key="tag.id"
                  class="tag-capsule text-white"
                  :style="{ backgroundColor: tag.color || '#64748B' }"
                >
                  {{ tag.name }}
                </span>
              </div>

              <div class="text-xs text-slate-400 space-y-1">
                <p>创建人：ID-{{ selectedNote.creator_id }}</p>
                <p>创建时间：{{ selectedNote.created_at }}</p>
                <p v-if="selectedNote.due_time">截止时间：{{ selectedNote.due_time }}</p>
              </div>
            </div>

            <div class="flex gap-3 pt-4 border-t border-slate-100 mt-4">
              <button class="flex-1 py-2.5 btn-primary text-sm">保存</button>
              <button class="flex-1 py-2.5 btn-secondary text-sm" @click="closeDetail">
                关闭
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
