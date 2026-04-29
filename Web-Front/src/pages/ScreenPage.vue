<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const currentTime = ref(new Date())
let timer: ReturnType<typeof setInterval>

onMounted(() => {
  timer = setInterval(() => {
    currentTime.value = new Date()
  }, 1000)
})

onUnmounted(() => {
  clearInterval(timer)
})

function formatTime(date: Date): string {
  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const s = date.getSeconds().toString().padStart(2, '0')
  return `${h}:${m}:${s}`
}

function formatDate(date: Date): string {
  const y = date.getFullYear()
  const m = (date.getMonth() + 1).toString().padStart(2, '0')
  const d = date.getDate().toString().padStart(2, '0')
  return `${y}年${m}月${d}日`
}
</script>

<template>
  <div class="fixed inset-0 bg-[#0F172A] flex flex-col overflow-hidden">
    <!-- 顶部信息条 -->
    <div class="h-16 border-b border-[#1E293B] flex items-center justify-between px-8 shrink-0">
      <div class="flex items-center gap-6">
        <div class="text-2xl font-bold text-slate-200 tracking-wide" style="font-size: clamp(16px, 1.4vw, 24px)">
          {{ formatTime(currentTime) }}
        </div>
        <span class="text-slate-500" style="font-size: clamp(12px, 1vw, 16px)">{{ formatDate(currentTime) }}</span>
      </div>
      <div class="flex items-center gap-4">
        <h1 class="text-slate-200 font-semibold tracking-wide" style="font-size: clamp(16px, 1.4vw, 24px)">
          资警数智·轻燕 · 应急指挥大屏
        </h1>
        <div class="flex items-center gap-2">
          <span class="status-dot status-dot-online" />
          <span class="text-slate-400" style="font-size: clamp(10px, 0.9vw, 14px)">已连接</span>
        </div>
      </div>
    </div>

    <!-- 主内容区：4栏分栏 -->
    <div class="flex-1 p-6 grid grid-cols-4 gap-5 overflow-hidden">
      <div
        v-for="col in 4"
        :key="col"
        class="bg-[#1E293B] border border-[#334155] rounded-card flex flex-col overflow-hidden"
      >
        <div class="flex items-center gap-2 px-4 py-3 bg-[#0F172A] border-b border-[#334155]">
          <div class="w-7 h-7 rounded-full bg-slate-600 flex items-center justify-center text-xs text-slate-300">
            {{ col }}
          </div>
          <span class="text-sm text-slate-300 font-medium">栏位 {{ col }}</span>
          <span class="status-dot status-dot-online ml-auto" />
        </div>

        <div class="flex-1 p-5 overflow-auto space-y-4">
          <div
            v-if="col === 1"
            class="p-4 rounded-lg border"
            style="background-color: rgba(66, 32, 6, 0.4); border-color: rgba(217, 119, 6, 0.5)"
          >
            <h4 class="text-sm font-semibold mb-2" style="color: #FBBF24">待办提醒</h4>
            <p class="text-xs text-slate-400">便签内容展示区域，等待数据同步...</p>
          </div>
          <div
            v-else
            class="p-4 rounded-lg border border-[#334155] bg-[#0F172A]/30"
          >
            <p class="text-xs text-slate-500">等待数据推送中...</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部消息滚动条 -->
    <div class="h-10 bg-[#0F172A] border-t border-[#1E293B] flex items-center px-8 shrink-0 overflow-hidden">
      <span class="text-slate-400 mr-3 text-xs shrink-0" style="font-size: clamp(10px, 0.9vw, 14px)">📢</span>
      <span class="text-slate-500 text-xs whitespace-nowrap" style="font-size: clamp(10px, 0.9vw, 14px)">
        系统就绪，等待指挥指令下发...
      </span>
    </div>
  </div>
</template>
