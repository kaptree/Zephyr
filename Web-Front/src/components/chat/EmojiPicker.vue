<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { listEmoticons, uploadEmoticon, batchUploadEmoticons, deleteEmoticon } from '@/services/emoticons';
import type { Emoticon } from '@/services/emoticons';
import { useConfirm } from '@/composables/useConfirm';

// 全局确认对话框（轻量级通知美学）
const { confirm: appConfirm } = useConfirm();

const auth = useAuthStore();

const emit = defineEmits<{ select: [emoji: string]; sendImage: [path: string] }>();

// ===== 文本 emoji（原有） =====
// 表情包：高频趣味组合（大号网格）
const packEmojis = [
  '😀', '😂', '🤣', '😊', '😍', '😘', '😎', '🤔',
  '🙄', '😅', '🥺', '😭', '😤', '🤯', '😴', '🥳',
  '🤩', '😇', '🤠', '🤡', '👻', '💪', '🙏', '👍',
  '👏', '🎉', '🔥', '❤️', '💯', '✨', '🎂', '🍻',
  '🤝', '✌️', '🫡', '😷', '🤒', '🥵', '🥶', '🤬',
];

// emoji 分类
const groups: Record<string, string[]> = {
  smile: ['😀','😁','😂','🤣','😃','😄','😅','😆','😉','😊','😋','😎','😍','😘','🥰','😗','😙','😚','🙂','🤗','🤔','🫠','😐','😑','😶','🙄','😏','😣','😥','😮','🤐','😯','😪','😫','😴','😌','😛','😜','😝','🤤','😒','😓','😔','😕','🙃','🤑','😲','☹️','🙁','😖','😞','😟','😤','😢','😭','😦','😧','😨','😩','🤯','😬','😰','😱','🥵','🥶','😳','🤪','😵','🥴','😡','😠','🤬','😷','🤒','🤕','🤢','🤮','🤧','😇','🤠','🤡','🥳','🥺','🤥','🤫','🤭','🧐','🤨','😈','👿','👹','👺','💀','👻','👽','🤖'],
  hand: ['👍','👎','👊','✊','🤛','🤜','🤞','✌️','🤟','🤘','👌','🤏','👈','👉','👆','👇','☝️','✋','🤚','🖐','🖖','👋','🤙','💪','🦾','✍️','🙏','💅','🤳','👏','🤲','🫶'],
  animal: ['🐶','🐱','🐭','🐹','🐰','🦊','🐻','🐼','🐨','🐯','🦁','🐮','🐷','🐸','🐵','🙈','🙉','🙊','🐔','🐧','🐦','🐤','🦆','🦅','🦉','🦇','🐺','🐗','🐴','🦄','🐝','🐛','🦋','🐌','🐞','🐜','🦂','🐢','🐍','🦎','🐙','🦑','🐡','🐠','🐟','🐬','🐳','🦈','🐊','🐅','🐆','🦓','🦍','🐘','🦒','🦪','🐖','🐏','🐑','🦙','🐐','🦌','🐕','🐩','🐈','🐇','🦔','🦨','🦡','🐾'],
  food: ['🍏','🍎','🍐','🍊','🍋','🍌','🍉','🍇','🍓','🍈','🍒','🍑','🥭','🍍','🥥','🥝','🍅','🥑','🥦','🥒','🌽','🌶','🥕','🧄','🧅','🥔','🍠','🥐','🥯','🍞','🥖','🥨','🧀','🥚','🍳','🧈','🥞','🧇','🥓','🥩','🍗','🍖','🌭','🍔','🍟','🍕','🥪','🥙','🌮','🌯','🥗','🥘','🍝','🍜','🍲','🍛','🍣','🍱','🥟','🦪','🍤','🍙','🍚','🍘','🍢','🍡','🍧','🍨','🍦','🥧','🍰','🎂','🍮','🍭','🍬','🍫','🍿','🧂','🍯','🥛','🍼','☕','🍵','🧊','🍶','🍺','🍻','🥂','🍷','🥃','🍸','🍹','🧉'],
  sport: ['⚽','🏀','🏈','⚾','🥎','🏐','🏉','🎾','🏓','🏸','🥅','🏒','🏑','🥍','🏏','🥇','🥈','🥉','🏆','🏅','🎖','🎗','🎫','🎟','🎭','🎨','🎪','🎤','🎧','🎼','🎹','🥁','🎷','🎺','🎸','🎻','🎲','🎯','🎳','🎮','🎰','🧩'],
  travel: ['🚗','🚕','🚙','🚌','🚎','🏎','🚓','🚑','🚒','🚐','🚛','🚜','🏍','🛵','🚲','🛴','🚨','🚔','🚍','🚘','🚖','🚡','🚠','🚟','🚃','🚋','🚞','🚝','🚄','🚅','🚈','🚂','🚆','🚇','🚊','🚉','✈️','🛫','🛬','🛩','💺','🛰','🚀','🛸','🚁','🛶','⛵','🚤','🛥','🛳','⛴','🚢','⚓'],
  item: ['⌚','📱','📲','💻','⌨️','🖥','🖨','🖱','🖲','🕹','💽','💾','💿','📀','📼','📷','📸','📹','🎥','📽','🎞','📞','☎️','📟','📠','📺','📻','🎙','🎚','🎛','⏱','⏲','⏰','🕰','⌛','⏳','📡','🔋','🔌','💡','🔦','🕯','🗑','🛢','💸','💵','💴','💶','💷','💰','💳','💎','⚖','🔧','🔨','⚒','🛠','⛏','🔩','⚙','🧱','⛓','🔥'],
  symbol: ['❤️','🧡','💛','💚','💙','💜','🖤','🤍','🤎','💔','❣️','💕','💞','💓','💗','💖','💘','💝','💟','🉑','☢','☣','📴','📳','🈶','🈚','🈸','🈺','🈷','✴','🆚','💮','🉐','㊙','㊗','🈴','🈵','🈹','🈲','🅰','🅱','🆎','🆑','🅾','🆘','❌','⭕','💢','♨','🚫','💯','💤','♠','♥','♦','♣','🃏','🎴'],
};

const tabs = ref<{ key: string; label: string }[]>([
  { key: 'pack', label: '表情包' },
  { key: 'smile', label: '笑脸' },
  { key: 'hand', label: '手势' },
  { key: 'animal', label: '动物' },
  { key: 'food', label: '食物' },
  { key: 'sport', label: '活动' },
  { key: 'travel', label: '旅行' },
  { key: 'item', label: '物品' },
  { key: 'symbol', label: '符号' },
]);

// ===== 图片表情（需求 #22） =====
const imageMap = ref<Record<string, Emoticon[]>>({});
const imageCategories = ref<string[]>([]);
const loadingImages = ref(false);
const uploading = ref(false);
const batchUploading = ref(false);
const deleteId = ref('');

const isAdmin = computed(() => auth.isAdmin);
const IMG_PREFIX = 'img:';

function isImageTab(key: string) {
  return key.startsWith(IMG_PREFIX);
}
function imageCatOf(key: string) {
  return key.slice(IMG_PREFIX.length);
}

// 加载全部分类（懒加载：切换 tab 时按需拉取）
async function loadImageTabs() {
  try {
    const res = await listEmoticons();
    imageCategories.value = res.data?.categories || [];
    // 同步 tabs：图片分类插入到「表情包」之后
    const imgTabs = imageCategories.value.map((c) => ({ key: `${IMG_PREFIX}${c}`, label: c }));
    tabs.value = [
      { key: 'pack', label: '表情包' },
      ...imgTabs,
      { key: 'smile', label: '笑脸' },
      { key: 'hand', label: '手势' },
      { key: 'animal', label: '动物' },
      { key: 'food', label: '食物' },
      { key: 'sport', label: '活动' },
      { key: 'travel', label: '旅行' },
      { key: 'item', label: '物品' },
      { key: 'symbol', label: '符号' },
    ];
  } catch {
    /* ignore */
  }
}

async function refreshImages(category: string) {
  if (!category) return;
  loadingImages.value = true;
  try {
    const res = await listEmoticons(category);
    imageMap.value[category] = res.data?.list || [];
  } catch {
    /* ignore */
  } finally {
    loadingImages.value = false;
  }
}

async function onTabChange(key: string) {
  if (isImageTab(key)) {
    const cat = imageCatOf(key);
    if (!imageMap.value[cat]) await refreshImages(cat);
  }
}

function currentImages(): Emoticon[] {
  if (!isImageTab(activeTab.value)) return [];
  return imageMap.value[imageCatOf(activeTab.value)] || [];
}

function pick(e: string) {
  emit('select', e);
}

function sendImage(path: string) {
  emit('sendImage', path);
}

// ===== 上传 / 批量上传 / 删除 =====
async function onUploadFile(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file) return;
  uploading.value = true;
  try {
    await uploadEmoticon(file);
    alert('表情上传成功');
    await refreshImages('我的表情');
    await loadImageTabs();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    alert(e2?.response?.data?.message || '表情上传失败');
  } finally {
    uploading.value = false;
  }
}

async function onBatchUpload(e: Event) {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files || []);
  input.value = '';
  if (!files.length) return;
  batchUploading.value = true;
  try {
    const res = await batchUploadEmoticons(files);
    alert(`批量上传完成：成功 ${res.data?.success ?? 0} 个，跳过 ${res.data?.skipped ?? 0} 个`);
    await loadImageTabs();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    alert(e2?.response?.data?.message || '批量上传失败');
  } finally {
    batchUploading.value = false;
  }
}

async function onDelete(emo: Emoticon) {
  if (!(await appConfirm({ message: `确定删除表情「${emo.name}」吗？`, danger: true, confirmText: '删除' }))) return;
  deleteId.value = emo.id;
  try {
    await deleteEmoticon(emo.id);
    const cat = imageCatOf(activeTab.value);
    imageMap.value[cat] = (imageMap.value[cat] || []).filter((x) => x.id !== emo.id);
    if (cat === '我的表情') await loadImageTabs();
  } catch (err) {
    const e2 = err as { response?: { data?: { message?: string } } };
    alert(e2?.response?.data?.message || '删除失败');
  } finally {
    deleteId.value = '';
  }
}

function currentEmojis(): string[] {
  return activeTab.value === 'pack' ? packEmojis : groups[activeTab.value] || [];
}

const activeTab = ref('pack');

// ===== 需求32：鼠标悬停表情图片时放大预览，便于筛选查看以免发错 =====
const hoveredEmo = ref<Emoticon | null>(null);
const previewPos = ref<{ x: number; y: number } | null>(null);

function onPreviewEnter(e: MouseEvent, img: Emoticon) {
  hoveredEmo.value = img;
  previewPos.value = { x: e.clientX, y: e.clientY };
}
function onPreviewLeave() {
  hoveredEmo.value = null;
  previewPos.value = null;
}

// 预览浮层跟随鼠标，靠近视口边缘时自动反向偏移避免溢出
const previewStyle = computed(() => {
  const pos = previewPos.value;
  if (!pos) return {};
  const w = 152; // 预览卡宽（含 padding）
  const h = 172; // 预览卡高（含名称行）
  const gap = 14;
  const left = pos.x + gap + w > window.innerWidth ? pos.x - w - gap : pos.x + gap;
  const top = pos.y + gap + h > window.innerHeight ? pos.y - h - gap : pos.y + gap;
  return { left: `${Math.max(4, left)}px`, top: `${Math.max(4, top)}px` };
});

onMounted(() => {
  loadImageTabs();
});
</script>

<template>
  <div class="w-[340px] bg-white dark:bg-slate-800 rounded-card border border-slate-200 dark:border-slate-700 shadow-modal overflow-hidden">
    <!-- 分类 tab -->
    <div class="flex items-center gap-0.5 px-2 pt-2 overflow-x-auto scrollbar-none">
      <button
        v-for="t in tabs"
        :key="t.key"
        class="shrink-0 px-2.5 py-1 text-xs rounded-md transition-smooth"
        :class="activeTab === t.key
          ? 'bg-blue-500 text-white font-medium'
          : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700'"
        @click="activeTab = t.key; onTabChange(t.key)"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- 图片表情网格 -->
    <div v-if="isImageTab(activeTab)" class="h-52 overflow-y-auto scrollbar-thin p-2">
      <div v-if="loadingImages" class="py-8 text-center text-xs text-slate-400">加载中...</div>
      <div v-else-if="currentImages().length === 0" class="py-8 text-center text-xs text-slate-400">暂无表情，点击下方「上传表情」添加</div>
      <div v-else class="grid grid-cols-5 gap-1.5">
        <div
          v-for="img in currentImages()"
          :key="img.id"
          class="relative group"
        >
          <button
            class="w-full aspect-square rounded-lg overflow-hidden bg-slate-100 dark:bg-slate-700 hover:ring-2 hover:ring-blue-400 transition-smooth"
            v-tooltip="img.name"
            @mouseenter="onPreviewEnter($event, img)"
            @mouseleave="onPreviewLeave"
            @click="sendImage(img.path)"
          >
            <img :src="img.path" :alt="img.name" class="w-full h-full object-cover" loading="lazy" />
          </button>
          <!-- 删除（仅自己上传的表情） -->
          <button
            v-if="!img.is_system"
            class="absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full bg-red-500 text-white text-[10px] leading-4 text-center opacity-0 group-hover:opacity-100 transition-smooth"
            :disabled="deleteId === img.id"
            v-tooltip="'删除表情'"
            @click.stop="onDelete(img)"
          >×</button>
        </div>
      </div>
    </div>

    <!-- emoji 网格 -->
    <div v-else class="h-52 overflow-y-auto scrollbar-thin p-2 grid grid-cols-8 gap-0.5">
      <button
        v-for="e in currentEmojis()"
        :key="e"
        class="w-8 h-8 rounded-lg flex items-center justify-center text-xl hover:bg-slate-100 dark:hover:bg-slate-700 transition-smooth"
        v-tooltip="e"
        @click="pick(e)"
      >
        {{ e }}
      </button>
    </div>

    <!-- 底部工具条：上传 / 批量上传 -->
    <div class="flex items-center gap-2 px-2 py-1.5 border-t border-slate-100 dark:border-slate-700">
      <label
        class="shrink-0 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-xs text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-md cursor-pointer transition-smooth"
      >
        <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" class="hidden" @change="onUploadFile" />
        <span>{{ uploading ? '上传中...' : '➕ 上传表情' }}</span>
      </label>
      <label
        v-if="isAdmin"
        class="shrink-0 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-xs text-blue-600 dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 hover:bg-blue-100 dark:hover:bg-blue-900/50 rounded-md cursor-pointer transition-smooth"
        v-tooltip="'支持多选文件或选择整个文件夹批量导入'"
      >
        <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" multiple webkitdirectory class="hidden" @change="onBatchUpload" />
        <span>{{ batchUploading ? '批量中...' : '🗂 批量上传' }}</span>
      </label>
    </div>

    <!-- 需求32：悬停放大预览浮层（Teleport 到 body，避免被面板裁剪） -->
    <Teleport to="body">
      <div
        v-if="hoveredEmo && previewPos"
        class="fixed z-[200] pointer-events-none"
        :style="previewStyle"
      >
        <div class="bg-white dark:bg-slate-800 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-700 p-2 flex flex-col items-center">
          <img
            :src="hoveredEmo.path"
            :alt="hoveredEmo.name"
            class="w-28 h-28 object-contain rounded-lg bg-slate-50 dark:bg-slate-900"
          />
          <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-1 max-w-[128px] truncate">
            {{ hoveredEmo.name }}
          </p>
        </div>
      </div>
    </Teleport>
  </div>
</template>
