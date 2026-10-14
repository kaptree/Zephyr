<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useAchievementStore } from '@/stores/achievements';
import { useDarkMode } from '@/composables/useDarkMode';
import { useToast } from '@/composables/useToast';
import { fetchNoteStats, fetchHeatmap } from '@/services/notes';
import { fetchNotes } from '@/services/notes';
import { updateMyProfile, uploadImage } from '@/services/admin';
import type { NoteFilters } from '@/types';
import type { BackgroundFill } from '@/types/user';
import AnimatedNumber from '@/components/common/AnimatedNumber.vue';
import FloatingField from '@/components/common/FloatingField.vue';
import SubmitButton from '@/components/common/SubmitButton.vue';
import FormToggle from '@/components/common/FormToggle.vue';
import { AppIcon } from '@/components/icons';
import { useGloryCelebration } from '@/composables/useGloryCelebration';

const auth = useAuthStore();
const { isDark } = useDarkMode();
const toast = useToast();

// 庆祝动画偏好：任务归档「荣耀时刻」开关（localStorage 持久化，尊重用户选择）
const { celebrationEnabled, setCelebrationEnabled } = useGloryCelebration();

const loading = ref(true);
const loadError = ref('');

const totalNotes = ref(0);
const activeNotes = ref(0);
const archivedNotes = ref(0);
const completedNotes = ref(0);
const trendData = ref<{ date: string; count: number }[]>([]);

const savingProfile = ref(false);
// 沉浸式书写：保存按钮状态机（idle/loading/success/fail）
type SubmitState = 'idle' | 'loading' | 'success' | 'fail';
const profileSubmitState = ref<SubmitState>('idle');
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
    // 成就检查：头像装扮里程碑
    useAchievementStore().checkAndNotify();
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
const bgInitial = ref(auth.user?.background || '');
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
    bgInitial.value = bgUrl.value;
    bgSaved.value = true;
    toast.success('背景设置已保存');
    // 成就检查：背景装扮里程碑
    useAchievementStore().checkAndNotify();
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
  if (!auth.user || savingProfile.value) return;
  savingProfile.value = true;
  profileSubmitState.value = 'loading';
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
    // 保存成功：按钮绿色弹动（500ms）
    profileSubmitState.value = 'success';
    toast.success('✓ 个人信息已保存');
    setTimeout(() => {
      profileSubmitState.value = 'idle';
    }, 1200);
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } };
    profileError.value = err?.response?.data?.message || '保存失败';
    // 保存失败：按钮红色抖动 2 次，随后恢复可提交状态
    profileSubmitState.value = 'fail';
    setTimeout(() => {
      profileSubmitState.value = 'idle';
    }, 600);
  } finally {
    savingProfile.value = false;
  }
}
</script>

<template>
  <div
    class="relative min-h-full rounded-3xl overflow-hidden bg-gradient-to-b from-[#e8f3fe] via-[#f2f8ff] to-[#eef3fc] text-slate-700 shadow-xl shadow-sky-900/5 ring-1 ring-sky-900/5 dark:from-[#070b18] dark:via-[#070b18] dark:to-[#0b1226] dark:text-slate-200 dark:shadow-2xl dark:ring-white/5"
  >
    <!-- ===== 背景装饰层（日间晨曦 / 夜间深空） ===== -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      <div class="aurora aurora-1" />
      <div class="aurora aurora-2" />
      <div class="aurora aurora-3" />
      <div class="grid-overlay" />
      <div class="stars" />
      <div class="horizon-glow" />
      <!-- 日间太阳光晕 -->
      <div class="sun-glow" />
    </div>

    <div class="relative max-w-6xl mx-auto px-4 sm:px-6 py-8 space-y-6">
      <!-- ===== 顶部标题行 ===== -->
      <header class="rise flex items-center gap-4">
        <div class="relative shrink-0">
          <div
            class="absolute -inset-1.5 rounded-2xl bg-sky-400/30 blur-lg dark:bg-cyan-400/25"
            aria-hidden="true"
          />
          <div
            class="relative flex w-12 h-12 sm:w-14 sm:h-14 rounded-2xl items-center justify-center bg-gradient-to-br from-sky-400 to-indigo-500 shadow-lg shadow-sky-500/30"
          >
            <AppIcon name="user" :size="26" class="text-white" />
          </div>
        </div>
        <div class="min-w-0">
          <h1
            class="text-2xl sm:text-3xl font-black tracking-tight bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 bg-clip-text text-transparent dark:from-amber-200 dark:via-yellow-300 dark:to-orange-400"
          >
            个人中心
          </h1>
          <p
            class="mt-1 text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1.5 flex-wrap"
          >
            <span class="live-dot" aria-hidden="true" />
            数据实时同步
            <span class="text-slate-300 dark:text-slate-600">·</span>
            {{ auth.user?.name || '未登录' }}
            <span class="text-slate-300 dark:text-slate-600">·</span>
            {{ auth.user?.dept_name || '—' }}
          </p>
        </div>
      </header>

      <div v-if="loading" class="space-y-4">
        <div class="skeleton h-40 rounded-card" />
        <div class="skeleton h-64 rounded-card" />
        <div class="skeleton h-64 rounded-card" />
      </div>

      <div
        v-else-if="loadError"
        class="py-16 text-center rounded-3xl glass-card text-sm text-rose-500 dark:text-rose-300"
      >
        {{ loadError }}
        <button class="block mx-auto mt-2 text-sky-600 dark:text-cyan-300 hover:underline" @click="loadData">
          重试
        </button>
      </div>

      <template v-else>
        <!-- ===== 个人名片 + 数据总览 ===== -->
        <section class="glass-card rise" style="animation-delay: 60ms">
          <div class="card-top-line" />
          <div
            class="absolute -top-24 -right-24 w-64 h-64 rounded-full bg-sky-300/25 blur-3xl dark:bg-indigo-500/20"
            aria-hidden="true"
          />

          <div class="relative p-5 sm:p-6 flex flex-col lg:flex-row lg:items-center gap-6">
            <!-- 名片区：光环头像 + 姓名/部门/角色 -->
            <div class="flex items-center gap-4 shrink-0">
              <div class="halo">
                <div class="avatar-inner">
                  <button
                    type="button"
                    class="relative group block w-20 h-20 rounded-full overflow-hidden focus:outline-none"
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
                      class="w-full h-full bg-gradient-to-br from-sky-400 to-indigo-500 flex items-center justify-center text-2xl font-black text-white"
                    >
                      {{ auth.user?.name?.charAt(0) || '用' }}
                    </div>
                    <div
                      class="absolute inset-0 bg-black/45 opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center justify-center text-white text-[10px] font-medium"
                    >
                      {{ uploadingAvatar ? '上传中…' : '更换头像' }}
                    </div>
                  </button>
                </div>
              </div>
              <input
                ref="avatarInput"
                type="file"
                accept="image/png,image/jpeg,image/gif,image/webp"
                class="hidden"
                @change="handleAvatarChange"
              />
              <div class="min-w-0">
                <h3 class="text-xl font-black text-slate-800 dark:text-white truncate">
                  {{ auth.user?.name || '未登录' }}
                </h3>
                <p class="mt-1 text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                  <AppIcon name="users" :size="13" class="text-sky-500 dark:text-cyan-300 shrink-0" />
                  {{ auth.user?.dept_name || '—' }}
                </p>
                <span
                  class="mt-1.5 inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-gradient-to-r from-sky-500/15 to-indigo-500/15 text-sky-600 ring-1 ring-sky-500/25 dark:from-cyan-400/10 dark:to-indigo-400/10 dark:text-cyan-300 dark:ring-cyan-400/25"
                >
                  {{ roleLabel }}
                </span>
              </div>
            </div>

            <!-- 统计区：4 个玻璃数据块 -->
            <div class="flex-1 grid grid-cols-2 sm:grid-cols-4 gap-3">
              <div class="glass-cell stat-cell">
                <div class="text-[11px] text-slate-500 dark:text-slate-400">活跃任务</div>
                <div class="text-2xl font-extrabold text-sky-600 dark:text-cyan-300 tabular-nums">
                  <AnimatedNumber :value="activeNotes" :format="fmtNum" />
                </div>
              </div>
              <div class="glass-cell stat-cell">
                <div class="text-[11px] text-slate-500 dark:text-slate-400">已完成</div>
                <div class="text-2xl font-extrabold text-emerald-600 dark:text-emerald-300 tabular-nums">
                  <AnimatedNumber :value="completedNotes" :format="fmtNum" />
                </div>
              </div>
              <div class="glass-cell stat-cell">
                <div class="text-[11px] text-slate-500 dark:text-slate-400">已归档</div>
                <div class="text-2xl font-extrabold text-rose-500 dark:text-rose-300 tabular-nums">
                  <AnimatedNumber :value="archivedNotes" :format="fmtNum" />
                </div>
              </div>
              <div class="glass-cell stat-cell">
                <div class="text-[11px] text-slate-500 dark:text-slate-400">任务总数</div>
                <div class="text-2xl font-extrabold text-slate-700 dark:text-slate-100 tabular-nums">
                  <AnimatedNumber :value="totalNotes" :format="fmtNum" />
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 活动热力图 (全宽) -->
        <section class="glass-card rise" style="animation-delay: 120ms">
          <div class="card-top-line line-cyan" />
          <div class="relative p-5 sm:p-6">
            <div class="flex items-center justify-between mb-4">
              <div>
                <h4 class="text-sm font-bold text-slate-800 dark:text-white flex items-center gap-2">
                  <AppIcon name="chart" :size="16" class="text-sky-500 dark:text-cyan-300" />
                  归档活动热力图
                </h4>
                <span v-if="hoveredCell" class="text-xs text-slate-500 dark:text-slate-400 ml-6">
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
        </section>

        <!-- 个人信息编辑 + 平台背景设置（并排） -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
          <!-- 个人信息编辑表单 -->
          <section class="glass-card rise" style="animation-delay: 180ms">
            <div class="card-top-line" />
            <div class="relative p-5 sm:p-6">
              <h4 class="text-sm font-bold text-slate-800 dark:text-white mb-4 flex items-center gap-2">
                <AppIcon name="pencil" :size="16" class="text-sky-500 dark:text-cyan-300" />
                个人信息
              </h4>

              <form @submit.prevent="handleSaveProfile" class="space-y-3">
                <FloatingField v-model="editName" label="姓名" />
                <div>
                  <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">部门</span>
                  <input
                    :value="auth.user?.dept_name || ''"
                    class="input-field !py-1.5 !text-sm bg-white/60 dark:bg-white/5 text-slate-400 dark:text-slate-500"
                    disabled
                  />
                </div>
                <FloatingField v-model="editRank" label="警衔/职级" placeholder="如：二级警督" />
                <FloatingField v-model="editPhone" label="手机号" type="tel" inputmode="tel" />
                <FloatingField v-model="editEmail" label="邮箱" type="email" inputmode="email" />
                <div>
                  <span class="text-xs text-slate-400 dark:text-slate-500 mb-1 block">角色</span>
                  <input
                    :value="roleLabel"
                    class="input-field !py-1.5 !text-sm bg-white/60 dark:bg-white/5 text-slate-400 dark:text-slate-500"
                    disabled
                  />
                </div>

                <transition name="error-slide">
                  <p v-if="profileError" class="text-xs text-red-500">{{ profileError }}</p>
                </transition>

                <SubmitButton
                  type="submit"
                  class="w-full"
                  :state="profileSubmitState"
                  label="保存修改"
                  loading-label="保存中..."
                  success-label="✓ 已保存"
                  fail-label="✗ 保存失败"
                />
              </form>
            </div>
          </section>

          <!-- 平台背景设置：上传背景图 + 透明度 + 填充方式，保存后全局生效 -->
          <section class="glass-card rise" style="animation-delay: 240ms">
            <div class="card-top-line" />
            <div class="relative p-5 sm:p-6">
              <h4 class="text-sm font-bold text-slate-800 dark:text-white mb-4 flex items-center gap-2">
                <AppIcon name="view" :size="16" class="text-sky-500 dark:text-cyan-300" />
                平台背景
              </h4>

              <!-- 预览区：实时反映透明度 / 填充方式调整 -->
              <div
                class="relative h-36 rounded-xl overflow-hidden border border-sky-900/10 dark:border-white/10 bg-white/50 dark:bg-white/5 mb-4"
              >
                <div v-if="bgPreviewStyle" class="absolute inset-0" :style="bgPreviewStyle" />
                <div
                  v-else
                  class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-xs text-slate-400 dark:text-slate-500"
                >
                  <AppIcon name="monitor" :size="28" class="text-slate-300 dark:text-slate-600" />
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
                :disabled="savingBg || (!bgUrl && !bgInitial)"
                @click="handleSaveBackground"
              >
                {{ savingBg ? '保存中...' : '保存背景设置' }}
              </button>
            </div>
          </section>
        </div>

        <!-- 庆祝动画偏好：任务归档「荣耀时刻」开关（localStorage 持久化） -->
        <section class="glass-card rise" style="animation-delay: 300ms">
          <div class="card-top-line line-cyan" />
          <div class="relative p-5 sm:p-6">
            <h4 class="text-sm font-bold text-slate-800 dark:text-white mb-4 flex items-center gap-2">
              <AppIcon name="star" :size="16" class="text-amber-500 dark:text-amber-300" />
              庆祝动画
            </h4>
            <div class="flex items-center justify-between gap-4">
              <div class="min-w-0">
                <p class="text-sm text-slate-700 dark:text-slate-200">任务归档荣耀时刻</p>
                <p class="mt-0.5 text-xs text-slate-400 dark:text-slate-500">
                  任务归档成功后播放五彩纸屑与勋章庆祝动画；关闭后以普通提示代替
                </p>
              </div>
              <FormToggle
                :model-value="celebrationEnabled"
                @update:model-value="setCelebrationEnabled"
              />
            </div>
          </div>
        </section>
      </template>
    </div>
  </div>
</template>

<style scoped>
/* ===== 玻璃卡（日间白色玻璃 / 夜间深色玻璃） ===== */
.glass-card {
  position: relative;
  overflow: hidden;
  border-radius: 1.5rem;
  background: rgba(255, 255, 255, 0.66);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(125, 165, 220, 0.25);
  box-shadow:
    0 8px 30px rgba(56, 108, 180, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.8);
}
:global(.dark) .glass-card,
.dark .glass-card {
  background: rgba(255, 255, 255, 0.045);
  border-color: rgba(255, 255, 255, 0.1);
  box-shadow: none;
}

/* 卡片顶部高光线 */
.card-top-line {
  position: absolute;
  inset-inline: 0;
  top: 0;
  height: 1px;
  background: linear-gradient(to right, transparent, rgba(14, 165, 233, 0.55), transparent);
}
.card-top-line.line-cyan {
  background: linear-gradient(to right, transparent, rgba(251, 191, 36, 0.6), transparent);
}

/* ===== 玻璃数据块 ===== */
.glass-cell {
  border-radius: 0.75rem;
  border: 1px solid rgba(125, 165, 220, 0.22);
  background: rgba(255, 255, 255, 0.55);
  padding: 0.625rem 0.75rem;
}
:global(.dark) .glass-cell,
.dark .glass-cell {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.04);
}

/* 统计块悬浮微动效 */
.stat-cell {
  transition:
    transform 0.25s ease,
    box-shadow 0.25s ease;
}
.stat-cell:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(56, 108, 180, 0.12);
}
.dark .stat-cell:hover {
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.35);
}

/* ===== 极光光晕（日间晨曦云霞 / 夜间深空极光） ===== */
/* 渐变本身已渐隐至透明，无需 blur 滤镜；漂移动画已取消，装饰为静态图层，降低合成开销 */
.aurora {
  position: absolute;
  border-radius: 9999px;
}
.aurora-1 {
  width: 520px;
  height: 520px;
  top: -160px;
  left: -120px;
  background: radial-gradient(circle, rgba(125, 211, 252, 0.55), transparent 65%);
}
.aurora-2 {
  width: 460px;
  height: 460px;
  top: 8%;
  right: -140px;
  background: radial-gradient(circle, rgba(253, 186, 116, 0.4), transparent 65%);
}
.aurora-3 {
  width: 420px;
  height: 420px;
  bottom: -140px;
  left: 30%;
  background: radial-gradient(circle, rgba(249, 168, 212, 0.4), transparent 65%);
}
.dark .aurora-1 {
  background: radial-gradient(circle, rgba(99, 102, 241, 0.5), transparent 65%);
}
.dark .aurora-2 {
  background: radial-gradient(circle, rgba(34, 211, 238, 0.35), transparent 65%);
}
.dark .aurora-3 {
  background: radial-gradient(circle, rgba(244, 63, 94, 0.22), transparent 65%);
}

/* ===== 科技网格（顶部渐隐） ===== */
.grid-overlay {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(59, 130, 246, 0.09) 1px, transparent 1px),
    linear-gradient(90deg, rgba(59, 130, 246, 0.09) 1px, transparent 1px);
  background-size: 44px 44px;
  -webkit-mask-image: radial-gradient(ellipse 80% 60% at 50% 0%, black 25%, transparent 75%);
  mask-image: radial-gradient(ellipse 80% 60% at 50% 0%, black 25%, transparent 75%);
}
.dark .grid-overlay {
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.07) 1px, transparent 1px);
}

/* ===== 日间：太阳光晕 ===== */
.sun-glow {
  position: absolute;
  top: -140px;
  right: 6%;
  width: 320px;
  height: 320px;
  background: radial-gradient(circle, rgba(254, 240, 138, 0.75), rgba(253, 186, 116, 0.2) 45%, transparent 70%);
}
.dark .sun-glow {
  display: none;
}

/* ===== 星点（日间金色光尘 / 夜间银白星点） ===== */
.stars {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(1.5px 1.5px at 20% 30%, rgba(253, 224, 71, 0.75), transparent 55%),
    radial-gradient(1px 1px at 70% 15%, rgba(251, 191, 36, 0.5), transparent 55%),
    radial-gradient(1.5px 1.5px at 85% 45%, rgba(56, 189, 248, 0.5), transparent 55%),
    radial-gradient(1px 1px at 40% 70%, rgba(249, 168, 212, 0.5), transparent 55%),
    radial-gradient(1px 1px at 10% 80%, rgba(253, 224, 71, 0.45), transparent 55%),
    radial-gradient(1.2px 1.2px at 60% 85%, rgba(56, 189, 248, 0.45), transparent 55%),
    radial-gradient(1px 1px at 33% 12%, rgba(251, 146, 60, 0.5), transparent 55%),
    radial-gradient(1px 1px at 92% 68%, rgba(249, 168, 212, 0.45), transparent 55%);
}
.dark .stars {
  background-image:
    radial-gradient(1.5px 1.5px at 20% 30%, rgba(255, 255, 255, 0.8), transparent 55%),
    radial-gradient(1px 1px at 70% 15%, rgba(255, 255, 255, 0.55), transparent 55%),
    radial-gradient(1.5px 1.5px at 85% 45%, rgba(255, 255, 255, 0.5), transparent 55%),
    radial-gradient(1px 1px at 40% 70%, rgba(255, 255, 255, 0.45), transparent 55%),
    radial-gradient(1px 1px at 10% 80%, rgba(255, 255, 255, 0.4), transparent 55%),
    radial-gradient(1.2px 1.2px at 60% 85%, rgba(255, 255, 255, 0.5), transparent 55%),
    radial-gradient(1px 1px at 33% 12%, rgba(255, 255, 255, 0.6), transparent 55%),
    radial-gradient(1px 1px at 92% 68%, rgba(255, 255, 255, 0.45), transparent 55%);
}

/* ===== 底部地平线光（日间暖阳 / 夜间蓝紫） ===== */
.horizon-glow {
  position: absolute;
  bottom: -120px;
  left: 50%;
  transform: translateX(-50%);
  width: 70%;
  height: 240px;
  background: radial-gradient(ellipse, rgba(253, 186, 116, 0.3), rgba(125, 211, 252, 0.18) 45%, transparent 70%);
}
.dark .horizon-glow {
  background: radial-gradient(ellipse, rgba(56, 189, 248, 0.22), rgba(99, 102, 241, 0.12) 45%, transparent 70%);
}

/* ===== 头像旋转光环（个人名片） ===== */
.halo {
  position: relative;
  border-radius: 9999px;
  padding: 3px;
  background: conic-gradient(from 0deg, #38bdf8, #818cf8, #e0f2fe, #ffffff, #a5b4fc, #38bdf8);
  animation: halo-spin 8s linear infinite;
  box-shadow:
    0 0 20px rgba(56, 189, 248, 0.45),
    0 0 48px rgba(99, 102, 241, 0.2);
}
.dark .halo {
  background: conic-gradient(from 0deg, #fbbf24, #f59e0b, #fef3c7, #ffffff, #fde68a, #fbbf24);
  box-shadow:
    0 0 20px rgba(251, 191, 36, 0.5),
    0 0 48px rgba(251, 146, 60, 0.2);
}
@keyframes halo-spin {
  to {
    transform: rotate(360deg);
  }
}
.avatar-inner {
  border-radius: 9999px;
  padding: 2px;
  background: #ffffff;
}
.dark .avatar-inner {
  background: #0a0f22;
}

/* ===== LIVE 呼吸点 ===== */
/* box-shadow 动画逐帧重绘，改为伪元素 transform/opacity 动画（合成器处理，不触发重绘） */
.live-dot {
  position: relative;
  width: 7px;
  height: 7px;
  border-radius: 9999px;
  background: #10b981;
}
.live-dot::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  border: 1.5px solid rgba(16, 185, 129, 0.6);
  animation: live-ring 2s ease-out infinite;
}
@keyframes live-ring {
  to {
    transform: scale(2.8);
    opacity: 0;
  }
}
@media (prefers-reduced-motion: reduce) {
  .live-dot::after {
    animation: none;
  }
}

/* ===== 入场动画 ===== */
.rise {
  opacity: 0;
  animation: rise-in 0.7s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}
@keyframes rise-in {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

/* ===== 降低动态效果偏好 ===== */
@media (prefers-reduced-motion: reduce) {
  .aurora,
  .halo,
  .live-dot {
    animation: none !important;
  }
  .rise {
    animation: none !important;
    opacity: 1 !important;
  }
}
</style>
