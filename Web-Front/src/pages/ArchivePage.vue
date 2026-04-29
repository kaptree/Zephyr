<script setup lang="ts">
import { ref } from 'vue'

const viewMode = ref<'timeline' | 'card'>('card')
const dateFrom = ref('')
const dateTo = ref('')
const keyword = ref('')
</script>

<template>
  <div>
    <!-- 筛选栏 -->
    <div class="bg-white rounded-card p-4 mb-6 flex flex-wrap items-center gap-3 border border-slate-100">
      <div class="flex items-center gap-2">
        <input v-model="dateFrom" type="date" class="input-field !w-auto" />
        <span class="text-slate-400 text-sm">至</span>
        <input v-model="dateTo" type="date" class="input-field !w-auto" />
      </div>
      <button class="px-4 py-2 bg-slate-100 rounded-btn text-sm text-slate-600 hover:bg-slate-200 transition-smooth">
        标签 ▼
      </button>
      <button class="px-4 py-2 bg-slate-100 rounded-btn text-sm text-slate-600 hover:bg-slate-200 transition-smooth">
        人员 ▼
      </button>
      <button class="px-4 py-2 bg-slate-100 rounded-btn text-sm text-slate-600 hover:bg-slate-200 transition-smooth">
        部门 ▼
      </button>
      <input v-model="keyword" class="input-field !w-40" placeholder="关键词搜索" />
      <button class="btn-primary text-sm !py-2">搜索</button>
      <button class="px-4 py-2 text-sm text-slate-500 hover:text-slate-700 transition-smooth">清空</button>
    </div>

    <!-- 视图切换 -->
    <div class="flex items-center justify-between mb-6">
      <span class="text-xs text-slate-400">共 0 条归档记录</span>
      <div class="flex bg-slate-100 rounded-btn p-0.5">
        <button
          :class="[
            'px-4 py-1.5 rounded-md text-sm font-medium transition-smooth',
            viewMode === 'timeline' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'
          ]"
          @click="viewMode = 'timeline'"
        >时间轴</button>
        <button
          :class="[
            'px-4 py-1.5 rounded-md text-sm font-medium transition-smooth',
            viewMode === 'card' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'
          ]"
          @click="viewMode = 'card'"
        >卡片</button>
      </div>
    </div>

    <!-- 时间轴视图 -->
    <div v-if="viewMode === 'timeline'" class="relative pl-8">
      <div class="absolute left-3 top-0 bottom-0 w-0.5 bg-slate-200" />

      <div class="mb-8">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-2.5 h-2.5 rounded-full bg-slate-300 -ml-[32px] ring-4 ring-white" />
          <span class="text-sm font-semibold text-slate-700">2024年12月</span>
        </div>

        <div class="space-y-3">
          <div class="flex items-start gap-4">
            <span class="text-xs text-slate-400 w-10 shrink-0 pt-1">12-15</span>
            <div class="flex-1 bg-white rounded-card border border-slate-100 p-4 relative hover:shadow-note transition-smooth cursor-pointer">
              <h4 class="text-sm font-medium text-slate-900 mb-1 truncate">便签标题示例</h4>
              <p class="text-xs text-slate-400 line-clamp-2">便签内容摘要...</p>
              <span class="watermark-archived">已归档</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 卡片视图 -->
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
      <div class="p-5 rounded-card bg-slate-50 border border-slate-100 relative">
        <h3 class="text-base font-semibold text-slate-400 mb-2">已归档便签示例</h3>
        <p class="text-sm text-slate-300 line-clamp-3">暂无更多归档便签...</p>
        <span class="watermark-archived">已归档</span>
        <div class="flex gap-2 mt-4 pt-3 border-t border-slate-100">
          <button class="btn-secondary text-xs !py-1.5 !px-3">查看详情</button>
          <button class="btn-secondary text-xs !py-1.5 !px-3">导出Word</button>
          <button class="btn-primary text-xs !py-1.5 !px-3">恢复</button>
        </div>
      </div>
    </div>

    <!-- 空态 -->
    <div class="text-center py-16">
      <p class="text-slate-400 text-sm">暂无归档便签</p>
    </div>
  </div>
</template>
