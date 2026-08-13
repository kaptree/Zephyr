<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue: string;
    placeholder?: string;
    minHeight?: number;
  }>(),
  {
    modelValue: '',
    placeholder: '请输入内容...',
    minHeight: 180,
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const editorRef = ref<HTMLDivElement | null>(null);
const focused = ref(false);
let lastHtml = '';

function syncFromEditor() {
  if (!editorRef.value) return;
  const html = editorRef.value.innerHTML;
  if (html !== lastHtml) {
    lastHtml = html;
    emit('update:modelValue', html);
  }
  updatePlaceholder();
}

function updatePlaceholder() {
  const el = editorRef.value;
  if (!el) return;
  const isEmpty = !el.innerHTML || el.innerHTML === '<br>';
  el.classList.toggle('is-empty', isEmpty);
}

function execCommand(cmd: string, value?: string) {
  editorRef.value?.focus();
  document.execCommand(cmd, false, value);
  syncFromEditor();
}

function toggleBlock(tag: string) {
  document.execCommand('formatBlock', false, tag);
  syncFromEditor();
}

function insertUnorderedList() {
  document.execCommand('insertUnorderedList');
  syncFromEditor();
}

function insertOrderedList() {
  document.execCommand('insertOrderedList');
  syncFromEditor();
}

function insertQuote() {
  document.execCommand('formatBlock', false, 'blockquote');
  syncFromEditor();
}

function insertCode() {
  document.execCommand('formatBlock', false, 'pre');
  syncFromEditor();
}

function clearFormat() {
  editorRef.value?.focus();
  document.execCommand('removeFormat');
  document.execCommand('formatBlock', false, 'div');
  syncFromEditor();
}

function undo() {
  document.execCommand('undo');
  syncFromEditor();
}

function redo() {
  document.execCommand('redo');
  syncFromEditor();
}

watch(
  () => props.modelValue,
  (val) => {
    if (editorRef.value && val !== lastHtml && document.activeElement !== editorRef.value) {
      editorRef.value.innerHTML = val || '';
      lastHtml = val || '';
      updatePlaceholder();
    }
  }
);

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = props.modelValue || '';
    lastHtml = props.modelValue || '';
    updatePlaceholder();
  }
});

function handleInput() {
  syncFromEditor();
}

function handleBlur() {
  focused.value = false;
  syncFromEditor();
}
</script>

<template>
  <div
    class="rich-editor rounded-btn border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-900 overflow-hidden transition-smooth focus-within:border-blue-400 focus-within:ring-1 focus-within:ring-blue-400"
  >
    <!-- 工具栏 -->
    <div
      class="flex flex-wrap items-center gap-0.5 px-2 py-1.5 border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/60 select-none"
    >
      <button
        type="button"
        class="rich-tool-btn"
        title="撤销"
        @mousedown.prevent="undo"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M3 10h10a5 5 0 015 5v2M3 10l4-4m-4 4l4 4"
          />
        </svg>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="重做"
        @mousedown.prevent="redo"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 10h-10a5 5 0 00-5 5v2m15-7l-4-4m4 4l-4 4"
          />
        </svg>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button type="button" class="rich-tool-btn" title="标题" @mousedown.prevent="toggleBlock('h3')">
        <span class="text-xs font-bold">H</span>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="正文"
        @mousedown.prevent="toggleBlock('div')"
      >
        <span class="text-xs">¶</span>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button
        type="button"
        class="rich-tool-btn"
        title="加粗"
        @mousedown.prevent="execCommand('bold')"
      >
        <span class="font-bold text-sm">B</span>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="斜体"
        @mousedown.prevent="execCommand('italic')"
      >
        <span class="italic text-sm font-serif">I</span>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="下划线"
        @mousedown.prevent="execCommand('underline')"
      >
        <span class="underline text-sm">U</span>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="删除线"
        @mousedown.prevent="execCommand('strikeThrough')"
      >
        <span class="line-through text-sm">S</span>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button
        type="button"
        class="rich-tool-btn"
        title="无序列表"
        @mousedown.prevent="insertUnorderedList"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h.01M4 12h.01M4 18h.01M8 6h12M8 12h12M8 18h12"
          />
        </svg>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="有序列表"
        @mousedown.prevent="insertOrderedList"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h.01M4 12h.01M4 18h.01M10 7h10M10 13h10M10 19h10"
          />
        </svg>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="引用"
        @mousedown.prevent="insertQuote"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M8 10.5h-3a1 1 0 01-1-1v-3a1 1 0 011-1h3a1 1 0 011 1v5a4 4 0 01-4 4M19 10.5h-3a1 1 0 01-1-1v-3a1 1 0 011-1h3a1 1 0 011 1v5a4 4 0 01-4 4"
          />
        </svg>
      </button>
      <button
        type="button"
        class="rich-tool-btn"
        title="代码块"
        @mousedown.prevent="insertCode"
      >
        <span class="text-xs font-mono">&lt;/&gt;</span>
      </button>
      <span class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-1" />
      <button
        type="button"
        class="rich-tool-btn"
        title="清除格式"
        @mousedown.prevent="clearFormat"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>

    <!-- 编辑区 -->
    <div
      ref="editorRef"
      class="px-4 py-3 text-sm text-slate-700 dark:text-slate-200 leading-relaxed outline-none overflow-y-auto rich-content"
      :style="{ minHeight: `${minHeight}px`, maxHeight: '50vh' }"
      contenteditable="true"
      :data-placeholder="placeholder"
      @input="handleInput"
      @blur="handleBlur"
      @focus="focused = true"
    ></div>
  </div>
</template>

<style scoped>
.rich-tool-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 26px;
  border-radius: 6px;
  color: #475569;
  transition: all 0.15s;
}
.rich-tool-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}
.dark .rich-tool-btn {
  color: #94a3b8;
}
.dark .rich-tool-btn:hover {
  background: #334155;
  color: #f1f5f9;
}

.rich-content:empty::before {
  content: attr(data-placeholder);
  color: #94a3b8;
  pointer-events: none;
}
.rich-content.is-empty:not(:focus)::before {
  content: attr(data-placeholder);
  color: #94a3b8;
  pointer-events: none;
}
</style>
