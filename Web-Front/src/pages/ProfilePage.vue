<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useDarkMode } from '@/composables/useDarkMode';
import { useToast } from '@/composables/useToast';
import { fetchNoteStats, fetchHeatmap } from '@/services/notes';
import { fetchNotes } from '@/services/notes';
import { updateMyProfile, uploadImage } from '@/services/admin';
import type { NoteFilters } from '@/types';
import type { BackgroundFill } from '@/types/user';
import AnimatedNumber from '@/components/common/AnimatedNumber.vue';

const auth = useAuthStore();
const { isDark } = useDarkMode();
const toast = useToast();

const loading = ref(true);
const loadError = ref('');

const totalNotes = ref(0);
const activeNotes = ref(0);
const archivedNotes = ref(0);
const completedNotes = ref(0);
const trendData = ref<{ date: string; count: number }[]>([]);

const savingProfile = ref(false);
const profileSaved = ref(false);
const profileError = ref('');
const editName = ref('');
const editPhone = ref('');
const editEmail = ref('');
const editRank = ref('');

// ===== 头像上传（点击头像 → 选图 → 上传 → 自助接口保存）=====
const MAX_IMAGE_SIZE = 10 * 1024 * 1024;
const avatarInput = ref<HTMLInputElement | null>(null);
const uploadingAvatar = ref(false);

function pickAvatar() {
  avatarInput.value?.click();
}

// 同步本地用户缓存（全局头像/背景即时生效依赖 auth store 响应式）
function persistUser() {
  if (auth.user) localStorage.setItem('auth_user', JSON.stringify(auth.user));
}

function handleAvatarChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = ''; // 允许重复选择同一文件
  if (!file || uploadingAvatar.value) return;
  void doUploadAvatar(file);
}

async function doUploadAvatar(file: File) {
  if (!file.type.startsWith('image/')) {
    toast.error('请选择图片文件（png/jpg/gif/webp）');
    return;
  }
  if (file.size > MAX_IMAGE_SIZE) {
    toast.error('图片大小不能超过 10MB');
    return;
  }
  uploadingAvatar.value = true;
  try {
    const up = await uploadImage(file);
    await updateMyProfile({ avatar: up.data.url });
    if (auth.user) {
      auth.user.avatar = up.data.url;
      persistUser();
    }
    toast.success('头像已更新');
  } catch (err) {
    const e = err as { friendlyMessage?: string };
    toast.error(e.friendlyMessage || '头像上传失败');
  } finally {
    uploadingAvatar.value = false;
  }
}

// ===== 平台背景图（上传 + 透明度 + 填充方式，保存后全局生效）=====
const FILL_OPTIONS: { value: BackgroundFill; label: string }[] = [
  { value: 'cover', label: '适应裁剪' },
  { value: 'contain', label: '完整显示' },
  { value: 'fill', label: '拉伸填充' },
  { value: 'tile', label: '平铺' },
];
const bgInput = ref<HTMLInputElement | null>(null);
const uploadingBg = ref(false);
const savingBg = ref(false);
const bgSaved = ref(false);
const bgError = ref('');
const bgUrl = ref(auth.user?.background || '');
const bgOpacity = ref(Math.round((typeof auth.user?.bg_opacity === 'number' ? auth.user.bg_opacity : 1) * 100));
const bgFill = ref<BackgroundFill>(auth.user?.bg_fill || 'cover');

// 填充方式 → background-size/repeat 映射（与 useUserBackground 全局层一致）
function bgSizeAndRepeat(fill: BackgroundFill) {
  return {
    backgroundSize: fill === 'tile' ? 'auto' : fill === 'fill' ? '100% 100%' : fill,
    backgroundRepeat: fill === 'tile' ? 'repeat' : 'no-repeat',
  };
}

const bgPreviewStyle = computed(() => {
  if (!bgUrl.value) return null;
  return {
    backgroundImage: `url(${bgUrl.value})`,
    backgroundPosition: 'center',
    opacity: String(bgOpacity.value / 100),
    ...bgSizeAndRepeat(bgFill.value),
  };
});

function handleBgChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file || uploadingBg.value) return;
  void doUploadBackground(file);
}

async function doUploadBackground(file: File) {
  if (!file.type.startsWith('image/')) {
    toast.error('请选择图片文件（png/jpg/gif/webp）');
    return;
  }
  if (file.size > MAX_IMAGE_SIZE) {
    toast.error('图片大小不能超过 10MB');
    return;
  }
  uploadingBg.value = true;
  try {
    const up = await uploadImage(file);
    bgUrl.value = up.data.url; // 仅本地预览，需点击保存生效
    // 新图沿用当前透明度与填充方式，用户可继续微调
    toast.success('背景图已上传，调整后点击保存');
  } catch (err) {
    const e = err as { friendlyMessage?: string };
    toast.error(e.friendlyMessage || '背景图上传失败');
  } finally {
    uploadingBg.value = false;
  }
}

function removeBackground() {
  bgUrl.value = ''; // 本地清除，保存后生效
  bgError.value = '';
}

async function handleSaveBackground() {
  savingBg.value = true;
  bgError.value = '';
  bgSaved.value = false;
  try {
    const opacity = bgOpacity.value / 100;
    await updateMyProfile({
      background: bgUrl.value,
      bg_opacity: opacity,
      bg_fill: bgFill.value,
    });
    if (auth.user) {
      auth.user.background = bgUrl.value;
      auth.user.bg_opacity = opacity;
      auth.user.bg_fill = bgFill.value;
      persistUser();
    }
    bgSaved.value = true;
    toast.success('背景设置已保存');
    setTimeout(() => {
      bgSaved.value = false;
    }, 2000);
  } catch (err) {
    const e = err as { friendlyMessage?: string; response?: { data?: { message?: string } } };
    bgError.value = e.response?.data?.message || e.friendlyMessage || '保存失败';
  } finally {
    savingBg.value = false;
  }
}

const thisYear = new Date().getFullYear();
const yearOptions = computed(() => {
  const years: number[] = [];
  for (let y = thisYear; y >= thisYear - 3; y--) years.push(y);
  return years;
});
const selectedYear = ref(thisYear);
const hoveredCell = ref<HeatCell | null>(null);

// ===== 活动热力图（GitHub 风格）=====
// 固定 7 行网格（周一 → 周日），每列一周，首尾不足周用透明格占位，保证列整齐
interface HeatCell {
  date: string;
  count: number;
  level: number;
  inYear: boolean; // 是否属于所选年份（首尾跨年占位格为 false）
}

// 本地时区日期字符串（与后端 DATE(completed_at) 对齐，避免 toISOString 的 UTC 偏移）
function localDateStr(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

const heatMap = computed<HeatCell[][]>(() => {
  const year = selectedYear.value;
  const byDate = new Map<string, number>();
  // 兼容后端可能返回的 ISO 格式（如 "2026-05-07T00:00:00Z"），统一取前 10 位
  for (const t of trendData.value) byDate.set(String(t.date).slice(0, 10), t.count);

  // 起点：1月1日所在周的周一（前面可能跨到上一年 12 月）
  const jan1 = new Date(year, 0, 1);
  const mondayOffset = (jan1.getDay() + 6) % 7; // 周一=0
  const start = new Date(year, 0, 1 - mondayOffset);

  // 终点：12月31日所在周的周日（后面可能跨到次年 1 月）
  const dec31 = new Date(year, 11, 31);
  const sundayOffset = (7 - dec31.getDay()) % 7; // 周日=0
  const end = new Date(year, 11, 31 + sundayOffset);

  const cols: HeatCell[][] = [];
  let col: HeatCell[] = [];
  for (const d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    const dateStr = localDateStr(d);
    const count = byDate.get(dateStr) || 0;
    let level = 0;
    if (count > 0) level = 1;
    if (count >= 2) level = 2;
    if (count >= 5) level = 3;
    if (count >= 10) level = 4;
    col.push({ date: dateStr, count, level, inYear: d.getFullYear() === year });
    if (d.getDay() === 0) {
      cols.push(col);
      col = [];
    }
  }
  if (col.length) cols.push(col);
  return cols;
});

const monthLabels = computed(() => {
  const labels: { col: number; label: string }[] = [];
  const cols = heatMap.value;
  let prevMonth = -1;
  cols.forEach((col, colIdx) => {
    const first = col.find((c) => c.inYear) || col[0];
    const [y, m] = (first?.date || '').split('-').map(Number);
    if (y && m !== prevMonth) {
      labels.push({ col: colIdx, label: `${m}月` });
      prevMonth = m;
    }
  });
  return labels;
});

// 色阶：浅色模式与暗色模式分开配色，保证暗色下不刺眼
const lightColors = ['#EFF4F9', '#BFDBFE', '#93C5FD', '#60A5FA', '#2563EB'];
const darkColors = ['#1E293B', '#1E3A8A', '#1D4ED8', '#3B82F6', '#93C5FD'];

function getCellColor(level: number, inYear = true): string {
  if (!inYear) return 'transparent';
  const palette = isDark.value ? darkColors : lightColors;
  return palette[level] || palette[0];
}

function getCellTitle(cell: HeatCell): string {
  if (!cell.inYear) return '';
  if (cell.count === 0) return `${cell.date} · 无归档活动`;
  return `${cell.date} · 归档 ${cell.count} 条任务`;
}

onMounted(loadData);

async function loadData() {
  loading.value = true;
  loadError.value = '';
  try {
    const [statsRes, heatmapRes] = await Promise.all([
      fetchNoteStats({ days: 30 }),
      fetchHeatmap(selectedYear.value),
    ]);
    totalNotes.value = statsRes.data.total_notes || 0;
    activeNotes.value = statsRes.data.active_notes || 0;
    trendData.value = heatmapRes.data.daily || [];
    archivedNotes.value = heatmapRes.data.total_archived || 0;

    try {
      const completedRes = await fetchNotes({
        status: 'completed' as any,
        page: 1,
        page_size: 1,
      } as NoteFilters);
      completedNotes.value = (completedRes.data as unknown as { total: number }).total || 0;
    } catch {
      completedNotes.value = 0;
    }

    if (auth.user) {
      editName.value = auth.user.name || '';
      editPhone.value = auth.user.phone || '';
      editEmail.value = auth.user.email || '';
      editRank.value = auth.user.rank || '';
    }
  } catch {
    loadError.value = '加载数据失败';
  } finally {
    loading.value = false;
  }
}

function fmtNum(n: number): string {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
  return String(n);
}

const roleLabel = computed(() => {
  const map: Record<string, string> = {
    super_admin: '系统管理员',
    company_leader: '公司领导',
    dept_admin: '部门管理员',
    group_leader: '组长',
    user: '普通员工',
    screen_role: '大屏角色',
  };
  return map[auth.user?.role || ''] || '—';
});

async function handleSaveProfile() {
  if (!auth.user) return;
  savingProfile.value = true;
  profileSaved.value = false;
  profileError.value = '';
  try {
    // 自助接口（仅本人字段），普通用户也可保存
    await updateMyProfile({
      name: editName.value.trim(),
      phone: editPhone.value.trim() || '',
      email: editEmail.value.trim() || '',
      rank: editRank.value.trim() || '',
    });
    auth.user.name = editName.value.trim();
    auth.user.phone = editPhone.value.trim();
    auth.user.email = editEmail.value.trim();
    auth.user.rank = editRank.value.trim();
    localStorage.setItem('auth_user', JSON.stringify(auth.user));
    profileSaved.value = true;
    setTimeout(() => {
      profileSaved.value = false;
    }, 2000);
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } };
    profileError.value = err?.response?.data?.message || '保存失败';
  } finally {
    savingProfile.value = false;
  }
}
</script>

<template>
  <div class="w-full">
    <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">个人中心</h2>

    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-24 rounded-card" />
      <div class="skeleton h-16 rounded-card" />
      <div class="skeleton h-64 rounded-card" />
    </div>

    <div v-else-if="loadError" class="text-center py-16 text-sm text-red-400">
      {{ loadError }}
      <button class="block mx-auto mt-2 text-blue-500 hover:underline" @click="loadData">
        重试
      </button>
    </div>

    <template v-else>
      <!-- 用户信息 + 统计卡片合并行 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-6">
        <!-- 用户信息卡 -->
        <div
          class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300 flex items-center gap-4"
        >
          <!-- 头像：点击更换，悬停显示遮罩提示 -->
          <button
            type="button"
            class="relative group w-14 h-14 rounded-full shrink-0 overflow-hidden focus:outline-none"
            title="点击更换头像"
            :disabled="uploadingAvatar"
            @click="pickAvatar"
          >
            <img
              v-if="auth.user?.avatar"
              :src="auth.user.avatar"
              alt="头像"
              class="w-full h-full object-cover"
            />
            <div
              v-else
              class="w-full h-full bg-blue-500 flex items-center justify-center text-lg font-semibold text-white"
            >
              {{ auth.user?.name?.charAt(0) || '用' }}
            </div>
            <div
              class="absolute inset-0 bg-black/45 opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center justify-center text-white text-[10px] font-medium"
            >
              {{ uploadingAvatar ? '上传中…' : '更换头像' }}
            </div>
          </button>
          <input
            ref="avatarInput"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            class="hidden"
            @change="handleAvatarChange"
          />
          <div class="min-w-0">
            <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 truncate">
              {{ auth.user?.name || '未登录' }}
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {{ auth.user?.dept_name || '' }}
            </p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">{{ roleLabel }}</p>
          </div>
        </div>

        <!-- 统计卡片 -->
        <div class="col-span-2 grid grid-cols-4 gap-3">
          <div
            class="rounded-card p-4 text-center bg-amber-50 dark:bg-amber-900/20 transition-colors duration-300"
          >
            <div class="text-2xl font-bold text-amber-700 dark:text-amber-400 tabular-nums">
              <AnimatedNumber :value="activeNotes" :format="fmtNum" />
            </div>
            <div class="text-xs text-amber-600 dark:text-amber-500 mt-1">活跃任务</div>
          </div>
          <div
            class="rounded-card p-4 text-center bg-green-50 dark:bg-green-900/20 transition-colors duration-300"
          >
            <div class="text-2xl font-bold text-green-700 dark:text-green-400 tabular-nums">
              <AnimatedNumber :value="completedNotes" :format="fmtNum" />
            </div>
            <div class="text-xs text-green-600 dark:text-green-500 mt-1">已完成</div>
          </div>
          <div
            class="rounded-card p-4 text-center bg-red-50 dark:bg-red-900/20 transition-colors duration-300"
          >
            <div class="text-2xl font-bold text-red-700 dark:text-red-400 tabular-nums">
              <AnimatedNumber :value="archivedNotes" :format="fmtNum" />
            </div>
            <div class="text-xs text-red-600 dark:text-red-500 mt-1">已归档</div>
          </div>
          <div
            class="rounded-card p-4 text-center bg-slate-50 dark:bg-slate-800 transition-colors duration-300"
          >
            <div class="text-2xl font-bold text-slate-700 dark:text-slate-300 tabular-nums">
              <AnimatedNumber :value="totalNotes" :format="fmtNum" />
            </div>
            <div class="text-xs text-slate-500 dark:text-slate-400 mt-1">任务总数</div>
          </div>
        </div>
      </div>

      <!-- 活动热力图 (全宽) -->
      <div class="mb-5">
        <div
          class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300"
        >
          <div class="flex items-center justify-between mb-4">
            <div>
              <h4 class="text-sm font-semibold text-slate-900 dark:text-slate-100">
                归档活动热力图
              </h4>
              <span v-if="hoveredCell" class="text-xs text-slate-500 dark:text-slate-400 ml-2">
                {{ hoveredCell.date }} · {{ hoveredCell.count }}条归档
              </span>
            </div>
            <div class="flex items-center gap-3">
              <select
                v-model="selectedYear"
                class="input-field !w-auto !py-1 !text-xs"
                @change="loadData"
              >
                <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}年</option>
              </select>
              <div class="hidden sm:flex items-center gap-1">
                <span class="text-[9px] text-slate-400 dark:text-slate-500 mr-1">少</span>
                <div
                  v-for="lv in 5"
                  :key="lv"
                  class="w-3 h-3 rounded-[2px]"
                  :style="{ backgroundColor: getCellColor(lv - 1) }"
                />
                <span class="text-[9px] text-slate-400 dark:text-slate-500 ml-1">多</span>
              </div>
            </div>
          </div>

          <div class="flex gap-1">
            <!-- 星期标签（固定 7 行：周一 → 周日） -->
            <div class="flex flex-col gap-[2px] pt-3 mr-1">
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3">一</span>
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3" />
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3">三</span>
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3" />
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3">五</span>
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3" />
              <span class="text-[9px] text-slate-400 dark:text-slate-600 h-3 leading-3">日</span>
            </div>

            <div class="flex-1 overflow-x-auto scrollbar-thin">
              <!-- 月份标签 -->
              <div class="flex gap-[2px] mb-1 h-3">
                <template v-for="(ml, idx) in monthLabels" :key="idx">
                  <div
                    class="shrink-0"
                    :style="{
                      width: `${(idx === 0 ? ml.col : ml.col - monthLabels[idx - 1].col) * 14}px`,
                    }"
                  />
                  <span
                    class="text-[9px] text-slate-400 dark:text-slate-600 whitespace-nowrap shrink-0"
                    >{{ ml.label }}</span
                  >
                </template>
              </div>

              <!-- 色块网格：固定每列 7 格，跨年占位透明 -->
              <div class="flex gap-[2px]">
                <div v-for="(col, ci) in heatMap" :key="ci" class="flex flex-col gap-[2px]">
                  <div
                    v-for="(cell, ri) in col"
                    :key="ri"
                    class="w-3 h-3 rounded-[2px] transition-colors duration-150"
                    :class="cell.inYear ? 'cursor-pointer hover:ring-1 hover:ring-slate-400 dark:hover:ring-slate-500' : 'cursor-default'"
                    :style="{ backgroundColor: getCellColor(cell.level, cell.inYear) }"
                    :title="getCellTitle(cell)"
                    @mouseenter="hoveredCell = cell"
                    @mouseleave="hoveredCell = null"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 个人信息编辑表单 -->
      <div class="mb-6">
        <div
          class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300 max-w-lg"
        >
          <h4 class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">个人信息</h4>

          <form @submit.prevent="handleSaveProfile" class="space-y-3">
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">姓名</span>
              <input v-model="editName" class="input-field !py-1.5 !text-sm" placeholder="姓名" />
            </div>
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">部门</span>
              <input
                :value="auth.user?.dept_name || ''"
                class="input-field !py-1.5 !text-sm bg-slate-50 dark:bg-slate-900 text-slate-400 dark:text-slate-500"
                disabled
              />
            </div>
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">警衔/职级</span>
              <input
                v-model="editRank"
                class="input-field !py-1.5 !text-sm"
                placeholder="如：二级警督"
              />
            </div>
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">手机号</span>
              <input
                v-model="editPhone"
                class="input-field !py-1.5 !text-sm"
                placeholder="手机号"
              />
            </div>
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">邮箱</span>
              <input v-model="editEmail" class="input-field !py-1.5 !text-sm" placeholder="邮箱" />
            </div>
            <div>
              <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">角色</span>
              <input
                :value="roleLabel"
                class="input-field !py-1.5 !text-sm bg-slate-50 dark:bg-slate-900 text-slate-400 dark:text-slate-500"
                disabled
              />
            </div>

            <p v-if="profileError" class="text-xs text-red-500">{{ profileError }}</p>
            <p v-if="profileSaved" class="text-xs text-green-500">✓ 已保存</p>

            <button
              type="submit"
              class="w-full btn-primary text-sm !py-2 disabled:opacity-50"
              :disabled="savingProfile"
            >
              {{ savingProfile ? '保存中...' : '保存修改' }}
            </button>
          </form>
        </div>
      </div>

      <!-- 平台背景设置：上传背景图 + 透明度 + 填充方式，保存后全局生效 -->
      <div class="mb-6">
        <div
          class="bg-white dark:bg-slate-800 rounded-card border border-slate-100 dark:border-slate-700 p-6 transition-colors duration-300 max-w-lg"
        >
          <h4 class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-4">平台背景</h4>

          <!-- 预览区：实时反映透明度 / 填充方式调整 -->
          <div
            class="relative h-36 rounded-xl overflow-hidden border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-900 mb-4"
          >
            <div v-if="bgPreviewStyle" class="absolute inset-0" :style="bgPreviewStyle" />
            <div
              v-else
              class="absolute inset-0 flex items-center justify-center text-xs text-slate-400 dark:text-slate-500"
            >
              暂未设置背景图
            </div>
          </div>

          <input
            ref="bgInput"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            class="hidden"
            @change="handleBgChange"
          />
          <div class="flex items-center gap-2 mb-4">
            <button
              type="button"
              class="btn-primary text-xs !py-1.5 disabled:opacity-50"
              :disabled="uploadingBg"
              @click="bgInput?.click()"
            >
              {{ uploadingBg ? '上传中…' : bgUrl ? '更换背景图' : '上传背景图' }}
            </button>
            <button
              v-if="bgUrl"
              type="button"
              class="btn-secondary text-xs !py-1.5"
              @click="removeBackground"
            >
              移除
            </button>
          </div>

          <!-- 透明度滑块 -->
          <div class="mb-4" :class="!bgUrl && 'opacity-40 pointer-events-none'">
            <div class="flex justify-between text-xs text-slate-400 dark:text-slate-500 mb-1">
              <span>透明度</span>
              <span class="tabular-nums">{{ bgOpacity }}%</span>
            </div>
            <input
              v-model.number="bgOpacity"
              type="range"
              min="0"
              max="100"
              step="5"
              class="w-full accent-blue-500 cursor-pointer"
            />
          </div>

          <!-- 填充方式 -->
          <div class="mb-4" :class="!bgUrl && 'opacity-40 pointer-events-none'">
            <span class="text-xs text-slate-400 dark:text-slate-500 mb-1.5 block">填充方式</span>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="opt in FILL_OPTIONS"
                :key="opt.value"
                type="button"
                class="px-2 py-1.5 text-xs rounded-lg border transition-colors duration-150"
                :class="
                  bgFill === opt.value
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 font-medium'
                    : 'border-slate-200 dark:border-slate-700 text-slate-500 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600'
                "
                @click="bgFill = opt.value"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>

          <p v-if="bgError" class="text-xs text-red-500 mb-2">{{ bgError }}</p>
          <p v-if="bgSaved" class="text-xs text-green-500 mb-2">✓ 已保存，背景全局生效</p>

          <button
            type="button"
            class="w-full btn-primary text-sm !py-2 disabled:opacity-50"
            :disabled="savingBg || !bgUrl"
            @click="handleSaveBackground"
          >
            {{ savingBg ? '保存中...' : '保存背景设置' }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
