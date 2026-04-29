<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useNoteStore } from '@/stores/notes'
import { useAuthStore } from '@/stores/auth'
import type { Note } from '@/types'
import TagSelector from '@/components/common/TagSelector.vue'
import StickyNoteCard from '@/components/note/StickyNoteCard.vue'
import UserPicker from '@/components/common/UserPicker.vue'
import { createTag } from '@/services/tags'

const noteStore = useNoteStore()
const auth = useAuthStore()
const showCreateModal = ref(false)
const showDetailPanel = ref(false)
const selectedNote = ref<Note | null>(null)

const newTitle = ref('')
const newContent = ref('')
const selectedTagIds = ref<string[]>([])
const sourceType = ref<'self' | 'assigned' | 'collaboration'>('self')
const selectedAssigneeIds = ref<string[]>([])
const creating = ref(false)
const createError = ref('')

const editingTitle = ref('')
const editingContent = ref('')
const saving = ref(false)
const completing = ref(false)

const activeTab = ref('all')

const displayedNotes = computed(() => {
  if (activeTab.value === 'red') return noteStore.activeNotes.filter(n => n.color_status === 'red')
  return noteStore.activeNotes
})

onMounted(() => {
  noteStore.fetchNotes()
})

function handleTabClick(tab: string) {
  activeTab.value = tab
  if (tab === 'all') {
    noteStore.fetchNotes({ status: undefined })
  } else if (tab === 'red') {
    noteStore.fetchNotes({ status: undefined }).then(() => {
      noteStore.activeNotes = noteStore.activeNotes.filter(n => n.color_status === 'red')
    })
  } else if (tab === 'assigned') {
    noteStore.fetchNotes({ status: undefined }).then(() => {
      noteStore.activeNotes = noteStore.activeNotes.filter(n => n.source_type === 'assigned' || n.source_type === 'collaboration')
    })
  } else {
    noteStore.fetchNotes({ status: tab })
  }
}

watch(sourceType, (val) => {
  if (val === 'self') selectedAssigneeIds.value = []
})

function openCreateModal() {
  newTitle.value = ''
  newContent.value = ''
  selectedTagIds.value = []
  selectedAssigneeIds.value = []
  sourceType.value = 'self'
  createError.value = ''
  showCreateModal.value = true
}

function openDetail(note: Note) {
  selectedNote.value = note
  editingTitle.value = note.title || ''
  editingContent.value = note.content || ''
  showDetailPanel.value = true
}

function closeDetail() {
  showDetailPanel.value = false
  selectedNote.value = null
  completing.value = false
}

async function handleCreateTag(name: string) {
  try {
    const res = await createTag({ name, color: '#3B82F6', category: '自定义', scope: 'personal' })
    const newTag = res.data as { id: string }
    selectedTagIds.value = [...selectedTagIds.value, newTag.id]
  } catch {
    createError.value = '创建标签失败'
  }
}

async function handleSubmit() {
  if (!newTitle.value.trim()) {
    createError.value = '请输入便签标题'
    return
  }
  if (sourceType.value !== 'self' && selectedAssigneeIds.value.length === 0) {
    createError.value = '请选择指派人员'
    return
  }

  creating.value = true
  createError.value = ''
  try {
    const payload: any = {
      title: newTitle.value.trim(),
      content: newContent.value,
      tags: selectedTagIds.value,
      source_type: sourceType.value,
    }

    if (sourceType.value !== 'self') {
      payload.assignees = selectedAssigneeIds.value.map(id => ({ user_id: id }))
    }
    if (sourceType.value === 'assigned' && selectedAssigneeIds.value.length > 0) {
      payload.owner_id = selectedAssigneeIds.value[0]
    }

    const created = await noteStore.createNote(payload)

    if (sourceType.value !== 'self' && created) {
      try {
        await noteStore.remindNote(created.id, `【任务指派】${auth.user?.name || '管理员'} 指派您处理：${newTitle.value.trim()}`)
      } catch { /* 提醒发送失败不阻塞流程 */ }
    }

    showCreateModal.value = false
  } catch (e: unknown) {
    const err = e as { response?: { status: number; data?: { message?: string } } }
    createError.value = err?.response?.data?.message || '创建便签失败'
  } finally {
    creating.value = false
  }
}

async function handleSaveDetail() {
  if (!selectedNote.value) return
  saving.value = true
  try {
    await noteStore.updateNoteLocally(selectedNote.value.id, {
      title: editingTitle.value.trim(),
      content: editingContent.value,
    })
    closeDetail()
  } catch {
    // 保持面板打开
  } finally {
    saving.value = false
  }
}

async function handleComplete(note: Note) {
  await noteStore.completeNote(note.id)
  if (showDetailPanel.value && selectedNote.value?.id === note.id) {
    closeDetail()
  }
}

async function handleRemind(note: Note) {
  await noteStore.remindNote(note.id, '请尽快处理')
  if (showDetailPanel.value && selectedNote.value?.id === note.id) {
    selectedNote.value = { ...selectedNote.value, color_status: 'red' as const }
  }
}
</script>

<template>
  <div class="relative min-h-full">
    <!-- 状态筛选栏 -->
    <div class="flex items-center gap-3 mb-6">
      <button
        v-for="tab in [
          { label: '全部', value: 'all' },
          { label: '待办', value: 'active' },
          { label: '盯办提醒', value: 'red' },
          { label: '已完成', value: 'completed' }
        ]"
        :key="tab.value"
        :class="[
          'px-4 py-1.5 rounded-btn text-sm font-medium transition-smooth',
          activeTab === tab.value
            ? 'bg-[#3B82F6] text-white'
            : 'bg-white text-slate-600 hover:bg-slate-50 border border-slate-200'
        ]"
        @click="handleTabClick(tab.value)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 错误提示 -->
    <div v-if="noteStore.error" class="mb-6 px-4 py-3 bg-red-50 border border-red-200 rounded-card text-sm text-red-600 flex items-center justify-between">
      <span>{{ noteStore.error }}</span>
      <button class="text-xs text-red-500 underline hover:text-red-700 ml-4" @click="noteStore.fetchNotes()">重试</button>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="noteStore.loading && noteStore.activeNotes.length === 0" class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <div v-for="n in 6" :key="n" class="skeleton h-44 rounded-card" />
    </div>

    <!-- 空态 -->
    <div v-else-if="!noteStore.loading && displayedNotes.length === 0 && !noteStore.error" class="flex flex-col items-center justify-center py-24">
      <div class="w-24 h-24 bg-slate-100 rounded-3xl flex items-center justify-center mb-6">
        <svg class="w-12 h-12 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      </div>
      <p class="text-slate-400 text-sm">{{ activeTab === 'completed' ? '暂无已完成便签' : '暂无活跃便签' }}</p>
      <p class="text-slate-300 text-xs mt-1">点击右下角 '+' 新建便签</p>
    </div>

    <!-- 便签墙 -->
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <StickyNoteCard
        v-for="note in displayedNotes"
        :key="note.id"
        :note="note"
        mode="web"
        :archived="false"
        class="animate-spring-enter"
        @click="openDetail(note)"
        @complete="handleComplete"
        @remind="handleRemind"
      />
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

    <!-- ====== 新建便签模态框 ====== -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-start justify-center pt-[10vh]">
        <div class="overlay-backdrop" @click="showCreateModal = false" />
        <div class="relative z-50 bg-white rounded-card shadow-modal w-full max-w-xl mx-4 animate-fade-in">
          <div class="p-6">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-semibold text-slate-900">新建便签</h2>
              <button class="p-1 rounded-lg hover:bg-slate-100 transition-smooth" @click="showCreateModal = false">
                <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <form class="space-y-4" @submit.prevent="handleSubmit">
              <input v-model="newTitle" class="input-field" placeholder="便签标题" autofocus />

              <textarea v-model="newContent" class="input-field h-32 resize-none" placeholder="便签内容..." />

              <!-- 标签 -->
              <div>
                <span class="text-xs text-slate-500 mb-1.5 block">标签</span>
                <TagSelector v-model="selectedTagIds" :max="5" @create-tag="handleCreateTag" />
              </div>

              <!-- 类型选择 -->
              <div>
                <span class="text-xs text-slate-500 mb-2 block">便签类型</span>
                <div class="flex gap-3">
                  <label
                    :class="[
                      'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-btn border-2 cursor-pointer transition-smooth',
                      sourceType === 'self' ? 'border-[#3B82F6] bg-blue-50 text-blue-700' : 'border-slate-200 text-slate-500 hover:border-slate-300'
                    ]"
                  >
                    <input v-model="sourceType" type="radio" value="self" class="sr-only" />
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
                    <span class="text-sm font-medium">仅自己</span>
                  </label>
                  <label
                    :class="[
                      'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-btn border-2 cursor-pointer transition-smooth',
                      sourceType === 'assigned' ? 'border-[#3B82F6] bg-blue-50 text-blue-700' : 'border-slate-200 text-slate-500 hover:border-slate-300'
                    ]"
                  >
                    <input v-model="sourceType" type="radio" value="assigned" class="sr-only" />
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                    <span class="text-sm font-medium">指派他人</span>
                  </label>
                </div>
              </div>

              <!-- 指派人员选择器 -->
              <div v-if="sourceType !== 'self'">
                <span class="text-xs text-slate-500 mb-1.5 block">
                  {{ sourceType === 'assigned' ? '选择负责人' : '选择协作人员' }}
                </span>
                <UserPicker v-model="selectedAssigneeIds" :multiple="sourceType === 'collaboration'" :max="sourceType === 'assigned' ? 1 : 20" />
              </div>

              <p v-if="createError" class="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-btn">{{ createError }}</p>

              <div class="flex justify-end gap-3 pt-4 border-t border-slate-100">
                <button type="button" class="px-5 py-2.5 text-sm text-slate-600 bg-slate-100 rounded-btn hover:bg-slate-200 transition-smooth" @click="showCreateModal = false" :disabled="creating">
                  取消
                </button>
                <button type="submit" class="px-5 py-2.5 text-sm text-white bg-[#3B82F6] rounded-btn hover:bg-blue-600 transition-smooth disabled:opacity-50" :disabled="creating">
                  {{ creating ? '创建中...' : '创建便签' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ====== 详情侧滑面板 ====== -->
    <Teleport to="body">
      <div v-if="showDetailPanel && selectedNote">
        <div class="overlay-backdrop" @click="closeDetail" />
        <div class="slide-panel">
          <div class="p-6 h-full flex flex-col">
            <!-- 面板标题 -->
            <div class="flex items-center justify-between mb-6">
              <div class="flex items-center gap-2">
                <h2 class="text-lg font-semibold text-slate-900">便签详情</h2>
                <span
                  v-if="selectedNote.color_status === 'red'"
                  class="text-xs px-2 py-0.5 bg-red-100 text-red-700 rounded-tag"
                >盯办中</span>
              </div>
              <button class="p-1 rounded-lg hover:bg-slate-100 transition-smooth" @click="closeDetail">
                <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div class="flex-1 overflow-auto space-y-5">
              <!-- 标题 -->
              <div>
                <span class="text-xs text-slate-400 mb-1 block">标题</span>
                <input v-model="editingTitle" class="input-field text-base font-semibold" placeholder="便签标题" />
              </div>

              <!-- 内容 -->
              <div>
                <span class="text-xs text-slate-400 mb-1 block">内容</span>
                <textarea v-model="editingContent" class="input-field min-h-[180px] resize-y text-sm" placeholder="便签内容..." />
              </div>

              <!-- 标签 -->
              <div>
                <span class="text-xs text-slate-400 mb-1 block">标签</span>
                <div v-if="selectedNote.tags.length" class="flex flex-wrap gap-2">
                  <span v-for="tag in selectedNote.tags" :key="tag.id" class="tag-capsule text-white" :style="{ backgroundColor: tag.color || '#64748B' }">{{ tag.name }}</span>
                </div>
                <span v-else class="text-xs text-slate-300">无标签</span>
              </div>

              <!-- 来源与指派信息 -->
              <div class="bg-slate-50 rounded-card p-4 space-y-2">
                <div class="flex justify-between text-xs">
                  <span class="text-slate-400">来源类型</span>
                  <span class="text-slate-700">{{ selectedNote.source_type === 'self' ? '自己创建' : selectedNote.source_type === 'assigned' ? '上级指派' : '协同任务' }}</span>
                </div>
                <div class="flex justify-between text-xs" v-if="selectedNote.assignees?.length">
                  <span class="text-slate-400">负责人</span>
                  <span class="text-slate-700">{{ selectedNote.assignees.map(a => a.name).join('、') }}</span>
                </div>
                <div class="flex justify-between text-xs">
                  <span class="text-slate-400">创建时间</span>
                  <span class="text-slate-700">{{ selectedNote.created_at?.slice(0, 16).replace('T', ' ') }}</span>
                </div>
                <div class="flex justify-between text-xs" v-if="selectedNote.due_time">
                  <span class="text-slate-400">截止时间</span>
                  <span class="text-slate-700 text-red-500">{{ selectedNote.due_time.slice(0, 16).replace('T', ' ') }}</span>
                </div>
              </div>
            </div>

            <!-- 底部操作栏 -->
            <div class="pt-4 border-t border-slate-100 mt-4 space-y-3">
              <div class="flex gap-2">
                <button class="flex-1 py-2.5 btn-primary text-sm disabled:opacity-50" :disabled="saving" @click="handleSaveDetail">
                  {{ saving ? '保存中...' : '保存' }}
                </button>
                <button class="flex-1 py-2.5 text-sm bg-green-500 text-white rounded-btn hover:bg-green-600 transition-smooth disabled:opacity-50" :disabled="completing" @click="completing = true; handleComplete(selectedNote!)">
                  {{ completing ? '归档中...' : '完成并归档' }}
                </button>
                <button v-if="selectedNote.color_status !== 'red'" class="flex-1 py-2.5 text-sm bg-red-50 text-red-600 rounded-btn hover:bg-red-100 transition-smooth" @click="handleRemind(selectedNote!)">
                  盯办
                </button>
              </div>
              <button class="w-full py-2 btn-secondary text-sm" @click="closeDetail">关闭</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
