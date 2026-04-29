<script setup lang="ts">
const templates = [
  { id: 't-1', name: '工作日报模板', type: 'work_report', desc: '每日工作情况汇报', fields: 5 },
  { id: 't-2', name: '情报研判模板', type: 'intel_report', desc: '情报信息研判报告', fields: 8 },
  { id: 't-3', name: '会议纪要模板', type: 'meeting', desc: '会议内容记录与纪要', fields: 6 },
  { id: 't-4', name: '案件协查模板', type: 'investigation', desc: '跨部门案件协查通报', fields: 7 },
]
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">模板库管理</h2>
      <span class="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded-btn">仅限超管</span>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div
        v-for="t in templates"
        :key="t.id"
        class="bg-white rounded-card border border-slate-100 p-5 hover:shadow-note transition-smooth cursor-pointer"
      >
        <div class="flex items-start justify-between mb-2">
          <h4 class="text-sm font-semibold text-slate-900">{{ t.name }}</h4>
          <span class="text-xs px-2 py-0.5 bg-slate-100 text-slate-500 rounded-tag">{{ t.type }}</span>
        </div>
        <p class="text-xs text-slate-400 mb-3">{{ t.desc }}</p>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-400">{{ t.fields }} 个字段</span>
          <div class="flex gap-2">
            <button class="text-xs px-2 py-1 bg-slate-100 text-slate-600 rounded hover:bg-slate-200 transition-smooth">编辑</button>
            <button class="text-xs px-2 py-1 bg-blue-50 text-blue-600 rounded hover:bg-blue-100 transition-smooth">使用</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
