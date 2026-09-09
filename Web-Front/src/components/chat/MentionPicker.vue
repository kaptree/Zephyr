<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import type { GroupMemberItem } from '@/types';
import { memberDisplayName, filterMentionMembers } from '@/utils/mention';
import { matchPinyin } from '@/utils/pinyin';

/** 浮层候选项：普通成员 / 全体成员 */
export interface MentionOption {
  type: 'all' | 'member';
  member?: GroupMemberItem;
}

const props = defineProps<{
  members: GroupMemberItem[];
  /** 输入框 @ 后的过滤词（支持拼音） */
  keyword: string;
  /** 群主可 @ 全体成员 */
  showAll: boolean;
}>();

const emit = defineEmits<{ select: [option: MentionOption] }>();

const ALL_NAME = '全体成员';

const listEl = ref<HTMLElement | null>(null);
const activeIndex = ref(0);

const options = computed<MentionOption[]>(() => {
  const kw = props.keyword.trim();
  const list: MentionOption[] = [];
  // 全体成员选项置顶，被关键字过滤掉时隐藏
  if (props.showAll && (!kw || matchPinyin(kw, ALL_NAME, 'all', '所有人'))) {
    list.push({ type: 'all' });
  }
  for (const m of filterMentionMembers(props.members, kw)) {
    list.push({ type: 'member', member: m });
  }
  return list;
});

// 过滤词变化时高亮回到第一项
watch(options, () => {
  activeIndex.value = 0;
});

function optionName(opt: MentionOption): string {
  if (opt.type === 'all') return ALL_NAME;
  return opt.member ? memberDisplayName(opt.member) : '';
}

function avatarChar(opt: MentionOption): string {
  return optionName(opt).slice(0, 1) || '?';
}

// 键盘导航（由输入框的 keydown 转发调用）
function moveUp() {
  if (!options.value.length) return;
  activeIndex.value = (activeIndex.value - 1 + options.value.length) % options.value.length;
  scrollActiveIntoView();
}

function moveDown() {
  if (!options.value.length) return;
  activeIndex.value = (activeIndex.value + 1) % options.value.length;
  scrollActiveIntoView();
}

function pickActive() {
  const opt = options.value[activeIndex.value];
  if (opt) emit('select', opt);
}

function scrollActiveIntoView() {
  nextTick(() => {
    listEl.value?.querySelector('[data-active="true"]')?.scrollIntoView({ block: 'nearest' });
  });
}

defineExpose({ moveUp, moveDown, pickActive, optionCount: computed(() => options.value.length) });
</script>

<template>
  <div class="w-64 bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 shadow-modal overflow-hidden">
    <div class="px-3 pt-2 pb-1 text-[10px] text-slate-400">选择提醒的人</div>
    <div ref="listEl" class="max-h-52 overflow-y-auto scrollbar-thin pb-1">
      <div v-if="options.length === 0" class="py-5 text-center text-xs text-slate-400">无匹配成员</div>
      <button
        v-for="(opt, i) in options"
        :key="opt.type === 'all' ? '__all__' : opt.member!.user_id"
        class="w-full flex items-center gap-2 px-3 py-1.5 text-left transition-smooth"
        :class="i === activeIndex ? 'bg-blue-50 dark:bg-slate-700' : 'hover:bg-slate-50 dark:hover:bg-slate-700/60'"
        :data-active="i === activeIndex"
        @mousemove="activeIndex = i"
        @click="emit('select', opt)"
      >
        <span
          class="w-7 h-7 rounded-full shrink-0 flex items-center justify-center text-xs font-medium text-white"
          :class="opt.type === 'all' ? 'bg-blue-500' : 'bg-gradient-to-br from-emerald-400 to-teal-500'"
        >{{ opt.type === 'all' ? '@' : avatarChar(opt) }}</span>
        <span class="min-w-0">
          <span class="block text-sm text-slate-700 dark:text-slate-200 truncate">{{ optionName(opt) }}</span>
          <span v-if="opt.type === 'all'" class="block text-[10px] text-slate-400">所有成员都会收到提醒</span>
          <span v-else class="block text-[10px] text-slate-400 truncate">{{ opt.member!.user?.username }}</span>
        </span>
      </button>
    </div>
    <div class="px-3 py-1.5 border-t border-slate-100 dark:border-slate-700 text-[10px] text-slate-400 flex gap-2">
      <span>↑↓ 选择</span><span>Enter 确认</span><span>Esc 关闭</span>
    </div>
  </div>
</template>
