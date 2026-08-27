<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useCollaborationStore } from '@/stores/collaboration'

const route = useRoute()
const collabStore = useCollaborationStore()
const roomId = route.params.id as string
const isProjection = ref(false)

// 需求29：指令输入与实时列表
const commandText = ref('')
const sendingCommand = ref(false)
const commandError = ref('')
const commandsEl = ref<HTMLElement | null>(null)

onMounted(() => {
  collabStore.joinRoom(roomId)
  collabStore.fetchCommands()
})

onUnmounted(() => {
  collabStore.leaveRoom()
})

// 新指令到达/发送后自动滚动到底部（动态刷新）
watch(
  () => collabStore.commands.length,
  async () => {
    await nextTick()
    if (commandsEl.value) {
      commandsEl.value.scrollTop = commandsEl.value.scrollHeight
    }
  }
)

async function handleSendCommand() {
  const text = commandText.value.trim()
  if (!text) {
    commandError.value = '请输入指令内容'
    return
  }
  sendingCommand.value = true
  commandError.value = ''
  try {
    await collabStore.sendCommand(text)
    commandText.value = ''
  } catch {
    commandError.value = '指令发送失败，请重试'
  } finally {
    sendingCommand.value = false
  }
}

function formatTime(ts: string): string {
  if (!ts) return ''
  const d = new Date(ts)
  return `${d.toLocaleDateString('zh-CN')} ${d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
}

function getStatusClass(status: string) {
  return status === 'connected' ? 'status-dot-online' : 'status-dot-offline'
}
</script>

<template>
  <div :class="['flex flex-col h-full', isProjection && 'fixed inset-0 bg-[#0F172A] z-[100]']">
    <!-- 顶部栏（非投屏模式显示） -->
    <div v-if="!isProjection" class="flex items-center justify-between mb-4 bg-white rounded-card p-4 border border-slate-100">
      <div class="flex items-center gap-4">
        <router-link to="/workbench" class="text-sm text-slate-400 hover:text-slate-600 transition-smooth flex items-center gap-1">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          返回
        </router-link>
        <h2 class="text-lg font-semibold text-slate-900">协同编辑室</h2>
        <span :class="['status-dot', getStatusClass(collabStore.syncStatus)]" />
        <span class="text-xs text-slate-400">{{ collabStore.syncStatus === 'connected' ? '已连接' : '已断开' }}</span>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-xs text-slate-400 mr-1">分栏：</span>
        <button
          v-for="n in [2, 4, 6, 8]"
          :key="n"
          :class="[
            'px-2 py-1 text-xs rounded-md transition-smooth',
            collabStore.columns === n ? 'bg-[#3B82F6] text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
          ]"
          @click="collabStore.setColumns(n)"
        >{{ n }}</button>
        <button
          class="ml-4 px-3 py-1.5 text-xs bg-slate-800 text-white rounded-btn hover:bg-slate-900 transition-smooth"
          @click="isProjection = true"
        >投屏</button>
      </div>
    </div>

    <!-- 协同画布 -->
    <div
      :class="[
        'flex-1 grid gap-4',
        isProjection ? 'p-8' : ''
      ]"
      :style="{ gridTemplateColumns: `repeat(${collabStore.columns}, 1fr)` }"
    >
      <div
        v-for="col in collabStore.columns"
        :key="col"
        :class="[
          'rounded-card flex flex-col overflow-hidden',
          isProjection ? 'bg-[#1E293B] border border-[#334155]' : 'bg-white border border-slate-100'
        ]"
      >
        <!-- 人员信息条 -->
        <div
          :class="[
            'flex items-center gap-2 px-4 py-2.5',
            isProjection ? 'bg-[#0F172A] border-b border-[#334155]' : 'bg-slate-50 border-b border-slate-100'
          ]"
        >
          <div class="w-6 h-6 rounded-full bg-slate-300 flex items-center justify-center text-[10px] font-medium text-slate-500">
            {{ col }}
          </div>
          <span :class="isProjection ? 'text-sm text-slate-300' : 'text-sm font-medium text-slate-700'">
            {{ collabStore.participants[col - 1]?.user_name || `栏位 ${col}` }}
          </span>
          <span v-if="collabStore.participants[col - 1]?.role === 'leader'" class="text-[10px] px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded">组长</span>
          <span :class="['status-dot', getStatusClass(collabStore.participants[col - 1]?.is_online ? 'connected' : 'disconnected')]" />
        </div>

        <!-- 编辑区 -->
        <div
          :class="[
            'flex-1 p-4 text-sm overflow-auto',
            isProjection ? 'text-slate-200 placeholder-slate-500' : 'text-slate-700 placeholder-slate-400'
          ]"
          contenteditable="false"
        >
          <p class="text-slate-400 italic text-xs">— 等待连接中 —</p>
        </div>
      </div>
    </div>

    <!-- 指令记录（需求29：实时动态刷新） -->
    <div
      v-if="!isProjection"
      class="mt-3 bg-white rounded-card border border-slate-100 flex flex-col max-h-56 overflow-hidden"
    >
      <div class="flex items-center justify-between px-4 py-2 bg-slate-50 border-b border-slate-100">
        <span class="text-xs font-medium text-slate-600">📋 指令记录</span>
        <span class="text-[10px] text-slate-400">{{ collabStore.commands.length }} 条</span>
      </div>
      <div ref="commandsEl" class="flex-1 overflow-auto px-4 py-2 space-y-2">
        <div v-if="collabStore.commands.length === 0" class="text-xs text-slate-300 py-3 text-center">
          暂无指令，等待领导下发...
        </div>
        <div
          v-for="cmd in collabStore.commands"
          :key="cmd.id"
          class="rounded-lg bg-slate-50 border-l-2 border-blue-500 px-3 py-2"
        >
          <div class="flex items-center justify-between gap-2 mb-0.5">
            <span class="text-xs font-medium text-slate-700">🎯 {{ cmd.user_name }}</span>
            <span class="text-[10px] text-slate-400">{{ formatTime(cmd.created_at) }}</span>
          </div>
          <p class="text-xs text-slate-600 whitespace-pre-wrap break-words">{{ cmd.content }}</p>
        </div>
      </div>
    </div>

    <!-- 底部状态条 -->
    <div
      v-if="!isProjection"
      :class="[
        'mt-3 bg-white rounded-card p-3 border border-slate-100 flex items-center justify-between',
      ]"
    >
      <span class="text-xs text-slate-400">{{ collabStore.typingStatusText || '—' }}</span>
      <div class="flex items-center gap-3">
        <input
          v-model="commandText"
          class="input-field !w-80"
          placeholder="领导下发指令..."
          maxlength="2000"
          @keyup.enter="handleSendCommand"
        />
        <button
          class="btn-primary text-xs !py-1.5 disabled:opacity-50"
          :disabled="sendingCommand"
          @click="handleSendCommand"
        >
          {{ sendingCommand ? '发送中...' : '下发指令' }}
        </button>
      </div>
    </div>
    <p v-if="commandError && !isProjection" class="text-xs text-red-500 mt-1 ml-auto">
      {{ commandError }}
    </p>

    <!-- 退出投屏按钮 -->
    <div
      v-if="isProjection"
      class="fixed bottom-4 right-4 z-[101]"
    >
      <button
        class="px-4 py-2 text-xs bg-slate-700 text-slate-300 rounded-btn hover:bg-slate-600 transition-smooth"
        @click="isProjection = false"
      >
        退出投屏 (ESC)
      </button>
    </div>
  </div>
</template>
