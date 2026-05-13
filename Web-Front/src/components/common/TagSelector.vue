<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import type { Tag } from '@/types';
import { fetchTags, createTag } from '@/services/tags';

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    max?: number;
    scope?: 'personal' | 'system' | 'all';
  }>(),
  {
    max: 10,
    scope: 'all',
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string[]];
}>();

const open = ref(false);
const searchText = ref('');
const allTags = ref<Tag[]>([]);
const loading = ref(false);
const loadError = ref('');

const triggerRef = ref<HTMLElement | null>(null);
const panelRef = ref<HTMLElement | null>(null);
const panelStyle = ref({ top: '0px', left: '0px', minWidth: '288px' });

const selectedTags = computed(() => allTags.value.filter((t) => props.modelValue.includes(t.id)));

const tagDisplayLabel = (tag: Tag) => (tag.sub_tag ? `${tag.name} › ${tag.sub_tag}` : tag.name);

const recentTags = computed(() => {
  const recent = JSON.parse(localStorage.getItem('recent_tags') || '[]') as string[];
  return recent
    .map((id: string) => allTags.value.find((t) => t.id === id))
    .filter(Boolean) as Tag[];
});

const tagGroups = computed(() => {
  const groups: Record<string, Tag[]> = {};
  for (const t of allTags.value) {
    if (!groups[t.name]) groups[t.name] = [];
    groups[t.name].push(t);
  }
  return Object.entries(groups).map(([name, tags]) => ({
    name,
    tags,
    hasSubTags: tags.some((t) => !!t.sub_tag),
  }));
});

const filteredTagGroups = computed(() => {
  const q = searchText.value.toLowerCase().trim();
  if (!q) return tagGroups.value;
  return tagGroups.value
    .map((g) => ({
      ...g,
      tags: g.tags.filter(
        (t) => t.name.toLowerCase().includes(q) || (t.sub_tag || '').toLowerCase().includes(q)
      ),
    }))
    .filter((g) => g.tags.length > 0);
});

onMounted(async () => {
  loading.value = true;
  try {
    const res = await fetchTags(props.scope);
    allTags.value = res.data as unknown as Tag[];
  } catch {
    loadError.value = '加载标签失败';
  } finally {
    loading.value = false;
  }
});

function recalcPosition() {
  if (!triggerRef.value) return;
  const rect = triggerRef.value.getBoundingClientRect();
  panelStyle.value = {
    top: rect.bottom + 4 + 'px',
    left: rect.left + 'px',
    minWidth: Math.max(288, rect.width) + 'px',
  };
}

function toggleOpen() {
  open.value = !open.value;
  if (open.value) {
    searchText.value = '';
    nextTick(() => {
      recalcPosition();
      const input = panelRef.value?.querySelector('input');
      input?.focus();
    });
  }
}

function handleClickOutside(e: MouseEvent) {
  if (
    panelRef.value &&
    !panelRef.value.contains(e.target as Node) &&
    triggerRef.value &&
    !triggerRef.value.contains(e.target as Node)
  ) {
    open.value = false;
  }
}

let scrollHandler: (() => void) | null = null;

watch(open, (val) => {
  if (val) {
    scrollHandler = () => {
      if (open.value) recalcPosition();
    };
    window.addEventListener('scroll', scrollHandler, true);
    document.addEventListener('click', handleClickOutside, true);
  } else {
    if (scrollHandler) {
      window.removeEventListener('scroll', scrollHandler, true);
      scrollHandler = null;
    }
    document.removeEventListener('click', handleClickOutside, true);
  }
});

onUnmounted(() => {
  if (scrollHandler) {
    window.removeEventListener('scroll', scrollHandler, true);
  }
  document.removeEventListener('click', handleClickOutside, true);
});

function toggleTag(tagId: string) {
  const current = [...props.modelValue];
  const idx = current.indexOf(tagId);
  if (idx >= 0) {
    current.splice(idx, 1);
  } else if (current.length < props.max) {
    current.push(tagId);
  }
  emit('update:modelValue', current);
}

function removeTag(tagId: string) {
  const current = props.modelValue.filter((id) => id !== tagId);
  emit('update:modelValue', current);
}

function isSelected(tagId: string): boolean {
  return props.modelValue.includes(tagId);
}

async function handleCreateTag(subTag?: string) {
  const name = searchText.value.trim();
  if (!name) return;

  try {
    const payload: {
      name: string;
      sub_tag?: string;
      color: string;
      category: string;
      scope: string;
    } = {
      name,
      sub_tag: subTag || '',
      color: '#3B82F6',
      category: '自定义',
      scope: 'personal',
    };
    const res = await createTag(payload);
    const newTag = res.data as unknown as Tag;
    allTags.value.push(newTag);
    toggleTag(newTag.id);
    searchText.value = '';
  } catch {
    loadError.value = '创建标签失败';
  }
}
</script>

<template>
  <div class="relative">
    <div class="flex flex-wrap gap-1.5 mb-1.5">
      <span
        v-for="tag in selectedTags"
        :key="tag.id"
        class="tag-capsule text-white"
        :style="{ backgroundColor: tag.color || '#64748B' }"
      >
        {{ tagDisplayLabel(tag) }}
        <button type="button" class="ml-1 hover:opacity-70" @click="removeTag(tag.id)">
          &times;
        </button>
      </span>
      <button
        type="button"
        ref="triggerRef"
        class="tag-capsule border border-dashed border-slate-300 text-slate-400 hover:border-slate-400 transition-smooth text-xs"
        @click.stop="toggleOpen"
      >
        + 添加标签
      </button>
    </div>

    <Teleport to="body">
      <div
        v-if="open"
        ref="panelRef"
        class="fixed bg-white rounded-card shadow-modal border border-slate-100 z-[60] overflow-hidden"
        :style="{ top: panelStyle.top, left: panelStyle.left, minWidth: panelStyle.minWidth }"
        @click.stop
      >
        <div class="p-3 border-b border-slate-100">
          <input
            v-model="searchText"
            class="input-field !text-xs"
            placeholder="搜索标签...（Enter 创建一级标签）"
            @keyup.enter.prevent="handleCreateTag()"
          />
        </div>

        <div v-if="recentTags.length && !searchText" class="px-3 pt-2">
          <span class="text-[10px] text-slate-400 uppercase">最近使用</span>
          <div class="flex flex-wrap gap-1 mt-1">
            <span
              v-for="tag in recentTags"
              :key="tag.id"
              :class="[
                'tag-capsule cursor-pointer text-xs transition-smooth',
                isSelected(tag.id) ? 'ring-2 ring-offset-1 ring-blue-400' : '',
              ]"
              :style="{
                backgroundColor: isSelected(tag.id) ? tag.color || '#64748B' : '#F1F5F9',
                color: isSelected(tag.id) ? '#fff' : '#475569',
              }"
              @click.stop="toggleTag(tag.id)"
            >
              {{ tagDisplayLabel(tag) }}
            </span>
          </div>
        </div>

        <div class="max-h-56 overflow-y-auto scrollbar-thin px-3 py-2">
          <div v-if="loading" class="text-center py-4 text-xs text-slate-400">加载中...</div>
          <div v-else-if="loadError" class="text-center py-4 text-xs text-red-400">
            {{ loadError }}
          </div>
          <div v-else-if="filteredTagGroups.length === 0" class="text-center py-4">
            <p class="text-xs text-slate-400">暂无匹配标签</p>
            <button
              type="button"
              v-if="searchText.trim()"
              class="text-xs text-[#3B82F6] hover:underline mt-1"
              @click.stop="handleCreateTag()"
            >
              创建一级标签 "{{ searchText }}"
            </button>
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="group in filteredTagGroups"
              :key="group.name"
              class="border-b border-slate-50 last:border-0 pb-1"
            >
              <div class="text-[10px] text-slate-400 uppercase font-medium px-1 pt-0.5">
                {{ group.name }}
              </div>
              <div class="ml-2 space-y-0.5">
                <button
                  v-for="tag in group.tags"
                  :key="tag.id"
                  :class="[
                    'w-full flex items-center gap-2 px-3 py-1.5 rounded-btn text-sm text-left transition-smooth',
                    isSelected(tag.id) ? 'bg-blue-50' : 'hover:bg-slate-50',
                  ]"
                  @click.stop="toggleTag(tag.id)"
                >
                  <span
                    class="w-3 h-3 rounded-full shrink-0"
                    :style="{ backgroundColor: tag.color || '#64748B' }"
                  />
                  <span class="flex-1 truncate">
                    <template v-if="tag.sub_tag">{{ tag.sub_tag }}</template>
                    <template v-else>
                      <span class="text-slate-400 text-xs">（通用）</span>
                    </template>
                  </span>
                  <span v-if="isSelected(tag.id)" class="text-xs text-[#3B82F6]">✓</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t border-slate-100 p-2 flex justify-end">
          <button
            type="button"
            class="text-xs px-3 py-1.5 bg-slate-100 text-slate-600 rounded-btn hover:bg-slate-200 transition-smooth"
            @click.stop="open = false"
          >
            完成
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>
