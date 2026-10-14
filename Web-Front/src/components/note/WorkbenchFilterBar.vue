<script setup lang="ts">
import { computed } from 'vue';
import DateRangePicker from '@/components/common/DateRangePicker.vue';

/**
 * 工作台任务筛选行：创建时间范围 + 标题/内容模糊查询。
 * 纯受控 UI 组件，不含请求逻辑：
 * - 日期范围选完 / 快捷预设 / 清空 → emit('date-change')，由父组件立即查询
 * - 关键词回车或点击「搜索」→ emit('search')
 * - 点击「清空」→ emit('reset')（父组件负责清空受控值并重新查询）
 */
const props = defineProps<{
  keyword: string;
  dateFrom: string;
  dateTo: string;
}>();

const emit = defineEmits<{
  (e: 'update:keyword', v: string): void;
  (e: 'update:dateFrom', v: string): void;
  (e: 'update:dateTo', v: string): void;
  (e: 'search'): void;
  (e: 'date-change'): void;
  (e: 'reset'): void;
}>();

const keywordModel = computed({
  get: () => props.keyword,
  set: (v: string) => emit('update:keyword', v),
});
const fromModel = computed({
  get: () => props.dateFrom,
  set: (v: string) => emit('update:dateFrom', v),
});
const toModel = computed({
  get: () => props.dateTo,
  set: (v: string) => emit('update:dateTo', v),
});

/** 是否已有生效的筛选条件（用于「筛选生效中」提示） */
const hasFilter = computed(() => !!(keywordModel.value.trim() || fromModel.value || toModel.value));
</script>

<template>
  <div class="glass-card p-4 mb-5 flex flex-wrap items-center gap-3" data-testid="workbench-filter-bar">
    <div class="card-top-line" aria-hidden="true" />
    <DateRangePicker
      v-model:from="fromModel"
      v-model:to="toModel"
      placeholder="按创建时间筛选"
      @change="emit('date-change')"
    />
    <input
      v-model="keywordModel"
      class="input-field !w-48"
      placeholder="🔍 搜索标题 / 内容关键词"
      @keyup.enter="emit('search')"
    />
    <button class="btn-primary text-sm !py-2" data-testid="filter-search-btn" @click="emit('search')">
      搜索
    </button>
    <button
      class="px-4 py-2 text-sm text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 transition-smooth"
      data-testid="filter-reset-btn"
      @click="emit('reset')"
    >
      清空
    </button>
    <span
      v-if="hasFilter"
      class="ml-auto text-xs text-slate-400 dark:text-slate-500"
      data-testid="filter-active-hint"
      >筛选生效中</span
    >
  </div>
</template>

<style scoped>
/* glass-card 全局样式带 overflow:hidden（用于圆角装饰裁剪），会把 DateRangePicker
   的 absolute 下拉日历面板整块裁掉（面板只露出顶部快捷行，月历不可见不可点击）。
   scoped 选择器特异性高于全局 .glass-card，确定性地放开裁剪。
   同时 z-index:1（配合 glass-card 自带 position:relative）提升层级：任务卡片同为
   glass-card（backdrop-filter 形成堆叠上下文）且 DOM 靠后，会把面板盖住。 */
.glass-card {
  overflow: visible;
  z-index: 1;
}
/* 放开裁剪后，顶部高光线两端会在圆角处外露，内收一个圆角半径保持原视觉 */
.glass-card > .card-top-line {
  left: 1.5rem;
  right: 1.5rem;
}
</style>
