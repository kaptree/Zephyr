<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

const props = withDefaults(
  defineProps<{
    from?: string;
    to?: string;
    placeholder?: string;
  }>(),
  { from: '', to: '', placeholder: '选择日期范围' }
);

const emit = defineEmits<{
  'update:from': [value: string];
  'update:to': [value: string];
  change: [from: string, to: string];
}>();

const open = ref(false);
const pickerEl = ref<HTMLDivElement | null>(null);
const viewYear = ref(new Date().getFullYear());
const viewMonth = ref(new Date().getMonth()); // 0-based

const WEEK_HEADERS = ['一', '二', '三', '四', '五', '六', '日'];

function fmt(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function todayStr(): string {
  return fmt(new Date());
}

const rangeText = computed(() => {
  if (props.from && props.to) {
    return props.from === props.to ? props.from : `${props.from} ~ ${props.to}`;
  }
  if (props.from) return `${props.from} ~ 结束日期`;
  return props.placeholder;
});

const calendarDays = computed(() => {
  const days: { date: string; day: number; inMonth: boolean }[] = [];
  const first = new Date(viewYear.value, viewMonth.value, 1);
  // 周一为一周起点
  const offset = (first.getDay() + 6) % 7;
  const start = new Date(first);
  start.setDate(first.getDate() - offset);
  for (let i = 0; i < 42; i++) {
    const d = new Date(start);
    d.setDate(start.getDate() + i);
    days.push({ date: fmt(d), day: d.getDate(), inMonth: d.getMonth() === viewMonth.value });
  }
  return days;
});

const monthLabel = computed(() => `${viewYear.value}年${viewMonth.value + 1}月`);

function prevMonth() {
  viewMonth.value--;
  if (viewMonth.value < 0) {
    viewMonth.value = 11;
    viewYear.value--;
  }
}
function nextMonth() {
  viewMonth.value++;
  if (viewMonth.value > 11) {
    viewMonth.value = 0;
    viewYear.value++;
  }
}

function isSelected(d: string): boolean {
  return d === props.from || d === props.to;
}
function isInRange(d: string): boolean {
  if (!props.from || !props.to) return false;
  return d > props.from && d < props.to;
}
function isRangeEnd(d: string): boolean {
  return (d === props.from || d === props.to) && props.from !== props.to;
}
function isToday(d: string): boolean {
  return d === todayStr();
}

function selectDay(d: string) {
  if (!props.from || (props.from && props.to)) {
    // 未选起点，或已有完整范围：重新开始选范围
    emit('update:from', d);
    emit('update:to', '');
    return;
  }
  // 已选起点：确定终点（反向选择自动交换）
  let from = props.from;
  let to = d;
  if (to < from) [from, to] = [to, from];
  emit('update:from', from);
  emit('update:to', to);
  emit('change', from, to);
  open.value = false; // 选完自动收起
}

function applyPreset(days: number) {
  const to = new Date();
  const from = new Date();
  if (days > 1) from.setDate(from.getDate() - days + 1);
  const f = fmt(from);
  const t = fmt(to);
  emit('update:from', f);
  emit('update:to', t);
  emit('change', f, t);
  open.value = false;
}
function applyThisMonth() {
  const now = new Date();
  const f = fmt(new Date(now.getFullYear(), now.getMonth(), 1));
  const t = fmt(now);
  emit('update:from', f);
  emit('update:to', t);
  emit('change', f, t);
  open.value = false;
}
function clearRange() {
  emit('update:from', '');
  emit('update:to', '');
  emit('change', '', '');
}

function toggle() {
  open.value = !open.value;
  if (open.value) {
    const anchor = props.from || props.to || todayStr();
    const d = new Date(anchor);
    viewYear.value = d.getFullYear();
    viewMonth.value = d.getMonth();
  }
}

function onDocClick(e: MouseEvent) {
  if (pickerEl.value && !pickerEl.value.contains(e.target as Node)) {
    open.value = false;
  }
}

onMounted(() => document.addEventListener('click', onDocClick));
onUnmounted(() => document.removeEventListener('click', onDocClick));
</script>

<template>
  <div ref="pickerEl" class="relative">
    <button
      type="button"
      class="flex items-center gap-2 px-3.5 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:border-blue-400 dark:hover:border-blue-500 transition-smooth"
      :class="{ 'ring-2 ring-blue-500/40 border-blue-400 dark:border-blue-500': open }"
      @click.stop="toggle"
    >
      <svg class="w-4 h-4 text-slate-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
      <span class="min-w-[150px] text-left">{{ rangeText }}</span>
      <svg
        class="w-3.5 h-3.5 text-slate-400 transition-transform duration-200"
        :class="{ 'rotate-180': open }"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <transition name="picker-pop">
      <div
        v-if="open"
        class="absolute left-0 top-full mt-2 z-50 w-[300px] bg-white dark:bg-slate-800 rounded-card shadow-modal border border-slate-100 dark:border-slate-700 p-4"
        @click.stop
      >
        <!-- 快捷选择 -->
        <div class="flex flex-wrap gap-1.5 mb-3">
          <button
            type="button"
            class="px-2 py-1 text-[11px] rounded-full bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 hover:text-blue-600 dark:hover:text-blue-300 transition-smooth"
            @click="applyPreset(1)"
          >今天</button>
          <button
            type="button"
            class="px-2 py-1 text-[11px] rounded-full bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 hover:text-blue-600 dark:hover:text-blue-300 transition-smooth"
            @click="applyPreset(7)"
          >近7天</button>
          <button
            type="button"
            class="px-2 py-1 text-[11px] rounded-full bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 hover:text-blue-600 dark:hover:text-blue-300 transition-smooth"
            @click="applyPreset(30)"
          >近30天</button>
          <button
            type="button"
            class="px-2 py-1 text-[11px] rounded-full bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-300 hover:bg-blue-100 dark:hover:bg-blue-900/50 hover:text-blue-600 dark:hover:text-blue-300 transition-smooth"
            @click="applyThisMonth"
          >本月</button>
          <button
            type="button"
            class="px-2 py-1 text-[11px] rounded-full bg-red-50 dark:bg-red-900/40 text-red-500 dark:text-red-300 hover:bg-red-100 dark:hover:bg-red-900/60 transition-smooth"
            @click="clearRange"
          >清空</button>
        </div>

        <!-- 月份导航 -->
        <div class="flex items-center justify-between mb-2">
          <button
            type="button"
            class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/40 transition-smooth"
            @click="prevMonth"
          >‹</button>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ monthLabel }}</span>
          <button
            type="button"
            class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/40 transition-smooth"
            @click="nextMonth"
          >›</button>
        </div>

        <!-- 星期 -->
        <div class="grid grid-cols-7 mb-1">
          <span
            v-for="w in WEEK_HEADERS"
            :key="w"
            class="text-center text-[10px] text-slate-400 py-1"
          >{{ w }}</span>
        </div>

        <!-- 日期格 -->
        <div class="grid grid-cols-7 gap-y-0.5">
          <button
            v-for="d in calendarDays"
            :key="d.date"
            type="button"
            class="h-8 text-xs rounded-lg transition-smooth"
            :class="{
              'text-slate-300 dark:text-slate-600': !d.inMonth,
              'bg-blue-500 text-white font-semibold shadow-sm': isSelected(d),
              'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-300': isInRange(d),
              'ring-2 ring-inset ring-blue-400 dark:ring-blue-500': isRangeEnd(d) && !isSelected(d),
              'text-blue-500 font-medium': isToday(d) && !isSelected(d) && !isInRange(d),
            }"
            @click="selectDay(d.date)"
          >
            {{ d.day }}
          </button>
        </div>

        <div class="mt-3 pt-2 border-t border-slate-100 dark:border-slate-700 text-[11px] text-slate-400 dark:text-slate-500 leading-relaxed">
          先点起始日期，再点结束日期；反向选择自动交换。
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.picker-pop-enter-active,
.picker-pop-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.picker-pop-enter-from,
.picker-pop-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
</style>
