<script setup lang="ts">
import { ref, computed } from 'vue';
import { markdownToHtml } from '@/utils/markdown';

const props = withDefaults(
  defineProps<{
    modelValue: string;
    placeholder?: string;
    minHeight?: number;
  }>(),
  {
    modelValue: '',
    placeholder: '支持 Markdown 语法，如：\n# 标题\n**加粗**\n- 列表项\n1. 有序项\n> 引用\n```代码```',
    minHeight: 240,
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const textareaRef = ref<HTMLTextAreaElement | null>(null);
const activeTab = ref<'edit' | 'preview'>('edit');
const fullscreen = ref(false);

const value = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v),
});

const renderedHtml = computed(() => markdownToHtml(props.modelValue));

/** 用前后缀包裹当前选中文本（未选中则插入占位） */
function insertWrap(before: string, after: string, hint: string) {
  const el = textareaRef.value;
  if (!el) return;
  const { selectionStart: s, selectionEnd: e } = el;
  const text = el.value;
  const sel = text.slice(s, e) || hint;
  const next = text.slice(0, s) + before + sel + after + text.slice(e);
  el.value = next;
  el.setSelectionRange(s + before.length, s + before.length + sel.length);
  el.focus();
  emit('update:modelValue', next);
}

/** 给选中行（无选中则当前行）行首加前缀，如标题/列表/引用 */
function prefixLines(prefix: string) {
  const el = textareaRef.value;
  if (!el) return;
  const { selectionStart: s, selectionEnd: e } = el;
  const text = el.value;
  const lineStart = text.lastIndexOf('\n', s - 1) + 1;
  const nl = text.indexOf('\n', e);
  const lineEnd = nl === -1 ? text.length : nl;
  const block = text.slice(lineStart, lineEnd);
  const prefixed = block
    .split('\n')
    .map((l) => prefix + l)
    .join('\n');
  const next = text.slice(0, lineStart) + prefixed + text.slice(lineEnd);
  el.value = next;
  el.setSelectionRange(lineStart, lineStart + prefixed.length);
  el.focus();
  emit('update:modelValue', next);
}

function insertTable() {
  const el = textareaRef.value;
  if (!el) return;
  const { selectionStart: s } = el;
  const tpl = `| 项目 | 说明 |\n| --- | --- |\n|  |  |\n\n`;
  const next = el.value.slice(0, s) + tpl + el.value.slice(s);
  el.value = next;
  el.setSelectionRange(s, s + tpl.length);
  el.focus();
  emit('update:modelValue', next);
}

function toggleFullscreen() {
  fullscreen.value = !fullscreen.value;
}
</script>

<template>
  <div
    class="md-editor rounded-btn border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 overflow-hidden transition-smooth focus-within:border-blue-400 focus-within:ring-1 focus-within:ring-blue-400"
    :class="{ 'md-fullscreen': fullscreen }"
  >
    <!-- 工具栏 -->
    <div
      class="flex flex-wrap items-center gap-0.5 px-2 py-1.5 border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/60 select-none"
    >
      <button type="button" class="md-tool-btn" v-tooltip="'一级标题'" @mousedown.prevent="prefixLines('# ')">
        <span class="text-xs font-bold">H1</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'二级标题'" @mousedown.prevent="prefixLines('## ')">
        <span class="text-xs font-bold">H2</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'三级标题'" @mousedown.prevent="prefixLines('### ')">
        <span class="text-xs font-bold">H3</span>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button type="button" class="md-tool-btn" v-tooltip="'加粗'" @mousedown.prevent="insertWrap('**', '**', '加粗文字')">
        <span class="font-bold text-sm">B</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'斜体'" @mousedown.prevent="insertWrap('*', '*', '斜体文字')">
        <span class="italic text-sm font-serif">I</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'删除线'" @mousedown.prevent="insertWrap('~~', '~~', '删除线文字')">
        <span class="line-through text-sm">S</span>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button type="button" class="md-tool-btn" v-tooltip="'无序列表'" @mousedown.prevent="prefixLines('- ')">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h.01M4 12h.01M4 18h.01M8 6h12M8 12h12M8 18h12" />
        </svg>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'有序列表'" @mousedown.prevent="prefixLines('1. ')">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h.01M4 12h.01M4 18h.01M10 7h10M10 13h10M10 19h10" />
        </svg>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'引用'" @mousedown.prevent="prefixLines('> ')">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10.5h-3a1 1 0 01-1-1v-3a1 1 0 011-1h3a1 1 0 011 1v5a4 4 0 01-4 4M19 10.5h-3a1 1 0 01-1-1v-3a1 1 0 011-1h3a1 1 0 011 1v5a4 4 0 01-4 4" />
        </svg>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button type="button" class="md-tool-btn" v-tooltip="'行内代码'" @mousedown.prevent="insertWrap('`', '`', 'code')">
        <span class="text-xs font-mono">`</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'代码块'" @mousedown.prevent="insertWrap('```\n', '\n```', '代码')">
        <span class="text-xs font-mono">&lt;/&gt;</span>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'链接'" @mousedown.prevent="insertWrap('[', '](url)', '链接文字')">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 010 5.656l-3 3a4 4 0 01-5.656-5.656l1.5-1.5M10.172 13.828a4 4 0 010-5.656l3-3a4 4 0 015.656 5.656l-1.5 1.5" />
        </svg>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'图片'" @mousedown.prevent="insertWrap('![', '](url)', '图片描述')">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </button>
      <button type="button" class="md-tool-btn" v-tooltip="'表格'" @mousedown.prevent="insertTable">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
      </button>
      <div class="flex-1" />
      <button
        type="button"
        class="md-tool-btn"
        v-tooltip="fullscreen ? '退出全屏 (Esc)' : '全屏编辑'"
        @click="toggleFullscreen"
      >
        <svg v-if="!fullscreen" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5M20 8V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5M20 16v4m0 0h-4m4 0l-5-5" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 4v4H4m0 0l5-5M16 4v4h4m0 0l-5-5M8 20v-4H4m0 0l5 5M16 20v-4h4m0 0l-5 5" />
        </svg>
      </button>
    </div>

    <!-- 编辑 / 预览切换 -->
    <div
      class="flex items-center gap-1 px-3 py-1.5 border-b border-slate-100 dark:border-slate-700 bg-white dark:bg-slate-900"
    >
      <button
        type="button"
        :class="[
          'px-2.5 py-1 rounded-md text-xs font-medium transition-smooth',
          activeTab === 'edit'
            ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300'
            : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700',
        ]"
        @click="activeTab = 'edit'"
      >
        编辑
      </button>
      <button
        type="button"
        :class="[
          'px-2.5 py-1 rounded-md text-xs font-medium transition-smooth',
          activeTab === 'preview'
            ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/40 dark:text-blue-300'
            : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700',
        ]"
        @click="activeTab = 'preview'"
      >
        预览
      </button>
      <span class="ml-auto text-[10px] text-slate-400 dark:text-slate-500"
        >支持 Markdown 语法，可全屏编写</span
      >
    </div>

    <!-- 主体：编辑源码 / 渲染预览 -->
    <div class="md-body">
      <textarea
        ref="textareaRef"
        v-model="value"
        v-show="activeTab === 'edit'"
        class="md-textarea w-full px-4 py-3 text-[13px] font-mono text-slate-700 dark:text-slate-200 leading-relaxed outline-none resize-y placeholder:text-slate-400 dark:placeholder:text-slate-500 bg-white dark:bg-slate-900"
        :style="{ minHeight: `${minHeight}px` }"
        :placeholder="placeholder"
        @keydown.esc="fullscreen = false"
      ></textarea>
      <div
        v-show="activeTab === 'preview'"
        class="md-preview px-4 py-3 overflow-y-auto prose prose-sm dark:prose-invert max-w-none text-slate-700 dark:text-slate-200"
        :style="{ minHeight: `${minHeight}px` }"
        v-html="renderedHtml"
      ></div>
    </div>
  </div>
</template>

<style scoped>
.md-tool-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 26px;
  border-radius: 6px;
  color: #475569;
  transition: all 0.15s;
}
.md-tool-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}
.dark .md-tool-btn {
  color: #94a3b8;
}
.dark .md-tool-btn:hover {
  background: #334155;
  color: #f1f5f9;
}

/* ===== 全屏模式 ===== */
.md-editor.md-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 9999;
  border-radius: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}
:global(.dark) .md-editor.md-fullscreen {
  background: #0f172a;
}
.md-editor.md-fullscreen .md-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.md-editor.md-fullscreen .md-textarea,
.md-editor.md-fullscreen .md-preview {
  flex: 1;
  height: 100%;
  min-height: 0 !important;
  resize: none;
}
</style>
