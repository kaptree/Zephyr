<script setup lang="ts">
/* 积分排行榜 —— 日/夜双主题领奖台：
   日间「晨曦晴空」唯美风（浅蓝天空 + 暖阳光晕 + 樱粉云霞 + 白色玻璃卡），
   夜间「深空科技」风（极光 + 星点 + 光柱 + 深色玻璃卡）。
   第一名居顶居中（旋转光环+光柱+浮动皇冠），第二名居左，第三名居右；
   其余人员依次排列。展示头像/姓名/部门/职位/积分与段位。 */
import { computed, onMounted, ref } from 'vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import AnimatedNumber from '@/components/common/AnimatedNumber.vue'
import { fetchLeaderboard, fetchMyPoints, fetchPointRules } from '@/services/points'
import type { LeaderboardEntry, LeaderboardPeriod, MePoints, PointRule } from '@/services/points'

const loading = ref(true)
const errorMsg = ref('')
const period = ref<LeaderboardPeriod>('total')
const list = ref<LeaderboardEntry[]>([])
const me = ref<MePoints | null>(null)
const rules = ref<PointRule[]>([])
const showRules = ref(false)

async function loadBoard() {
  loading.value = true
  errorMsg.value = ''
  try {
    const lb = await fetchLeaderboard(period.value)
    list.value = lb.data?.list ?? []
  } catch (e: unknown) {
    errorMsg.value = (e as { friendlyMessage?: string })?.friendlyMessage || '加载排行榜失败'
  } finally {
    loading.value = false
  }
}

function switchPeriod(p: LeaderboardPeriod) {
  if (period.value === p) return
  period.value = p
  loadBoard()
}

onMounted(async () => {
  loadBoard()
  try {
    me.value = (await fetchMyPoints()).data
  } catch {
    /* 概览失败不阻塞排行榜 */
  }
  try {
    rules.value = (await fetchPointRules()).data?.list ?? []
  } catch {
    /* 规则失败不阻塞 */
  }
})

/** 领奖台数据：第一名 / 第二名 / 第三名 / 其余名单 */
const podium = computed(() => ({
  first: list.value[0] ?? null,
  second: list.value[1] ?? null,
  third: list.value[2] ?? null,
  rest: list.value.slice(3),
}))

const PERIOD_TABS: { key: LeaderboardPeriod; label: string }[] = [
  { key: 'total', label: '总积分' },
  { key: 'week', label: '本周' },
  { key: 'month', label: '本月' },
]

// ---------------- 段位配色（日/夜双主题） ----------------
const TIER_STYLE: Record<string, { label: string; badge: string; bar: string }> = {
  bronze: {
    label: '青铜',
    badge:
      'bg-amber-500/10 text-amber-700 ring-1 ring-amber-500/25 dark:bg-amber-400/10 dark:text-amber-300 dark:ring-amber-400/30',
    bar: 'from-amber-500 to-orange-500 dark:from-amber-500 dark:to-orange-400',
  },
  silver: {
    label: '白银',
    badge:
      'bg-slate-500/10 text-slate-600 ring-1 ring-slate-500/25 dark:bg-slate-300/10 dark:text-slate-200 dark:ring-slate-300/30',
    bar: 'from-slate-400 to-slate-300 dark:from-slate-400 dark:to-slate-200',
  },
  gold: {
    label: '黄金',
    badge:
      'bg-yellow-500/10 text-yellow-700 ring-1 ring-yellow-500/25 dark:bg-yellow-300/10 dark:text-yellow-300 dark:ring-yellow-300/30',
    bar: 'from-amber-400 to-yellow-400 dark:from-yellow-400 dark:to-amber-300',
  },
  platinum: {
    label: '铂金',
    badge:
      'bg-cyan-500/10 text-cyan-700 ring-1 ring-cyan-500/25 dark:bg-cyan-300/10 dark:text-cyan-300 dark:ring-cyan-300/30',
    bar: 'from-sky-500 to-cyan-400 dark:from-cyan-400 dark:to-sky-300',
  },
}

function tierBadge(tier?: string) {
  return TIER_STYLE[tier ?? 'bronze'] ?? TIER_STYLE.bronze
}

const RULE_CATEGORY_LABEL: Record<string, string> = {
  task: '任务',
  collab: '协作',
  report: '报告',
  achievement: '成就',
}

function rulePointsLabel(r: PointRule): string {
  if (r.code === 'achievement_unlock') return '按档位 +5~20'
  return `+${r.points}`
}

// ---------------- 趋势柱状图 ----------------
const trendMax = computed(() => Math.max(1, ...(me.value?.trend ?? []).map((t) => t.points)))

/** 积分相对第一名的进度条宽度（0 分显示为 0） */
function barWidth(points: number): string {
  const max = Math.max(list.value[0]?.points ?? 0, 1)
  const pct = Math.round((points / max) * 100)
  return `${points > 0 ? Math.max(4, pct) : 0}%`
}

/** 2026-09-20 → 9/20 */
function dayLabel(date: string): string {
  const parts = date.split('-')
  return parts.length === 3 ? `${Number(parts[1])}/${Number(parts[2])}` : date
}

/** 入场动画错峰延迟 */
function delayFor(i: number): string {
  return `${Math.min(i, 14) * 45}ms`
}

function initial(name?: string): string {
  return name?.charAt(0) || '?'
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
      <header class="flex flex-wrap items-center justify-between gap-4 rise">
        <div class="flex items-center gap-4">
          <div class="relative">
            <div
              class="absolute -inset-1.5 rounded-2xl bg-amber-400/30 blur-lg dark:bg-amber-400/25"
              aria-hidden="true"
            />
            <div
              class="relative flex w-12 h-12 sm:w-14 sm:h-14 rounded-2xl items-center justify-center bg-gradient-to-br from-amber-300 to-orange-500 shadow-lg shadow-orange-500/30"
            >
              <AppIcon name="crown" variant="solid" :size="28" class="text-slate-950" />
            </div>
          </div>
          <div>
            <h1
              class="text-2xl sm:text-3xl font-black tracking-tight bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600 bg-clip-text text-transparent dark:from-amber-200 dark:via-yellow-300 dark:to-orange-400"
            >
              积分排行榜
            </h1>
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1.5 flex-wrap">
              <span class="live-dot" aria-hidden="true" />
              实时统计
              <span class="text-slate-300 dark:text-slate-600">·</span>
              {{ list.length }} 人参与
              <template v-if="me">
                <span class="text-slate-300 dark:text-slate-600">·</span>
                我的排名
                <span class="font-bold text-amber-600 dark:text-amber-300">#{{ me.rank }}</span>
              </template>
            </p>
          </div>
        </div>

        <!-- 周期切换（发光选中态） -->
        <div
          class="inline-flex rounded-full border border-sky-900/10 bg-white/70 backdrop-blur p-1 shadow-sm dark:border-white/10 dark:bg-white/[0.05] dark:shadow-none"
        >
          <button
            v-for="p in PERIOD_TABS"
            :key="p.key"
            :class="[
              'px-4 py-1.5 rounded-full text-sm font-semibold transition-smooth',
              period === p.key
                ? 'bg-gradient-to-r from-amber-400 to-orange-500 text-slate-950 shadow-lg shadow-amber-500/30'
                : 'text-slate-500 hover:text-slate-800 dark:text-slate-300 dark:hover:text-white',
            ]"
            @click="switchPeriod(p.key)"
          >
            {{ p.label }}
          </button>
        </div>
      </header>

      <!-- ===== 我的积分战绩卡 ===== -->
      <section
        v-if="me"
        class="glass-card rise"
        style="animation-delay: 60ms"
      >
        <div class="card-top-line" />
        <div
          class="absolute -top-24 -right-24 w-64 h-64 rounded-full bg-sky-300/25 blur-3xl dark:bg-indigo-500/20"
          aria-hidden="true"
        />

        <div class="relative p-5 sm:p-6">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="flex items-center gap-3">
              <div
                class="flex w-11 h-11 rounded-xl bg-white/70 ring-1 ring-sky-900/10 items-center justify-center dark:bg-white/5 dark:ring-white/10"
              >
                <AppIcon name="user" :size="22" class="text-sky-500 dark:text-cyan-300" />
              </div>
              <div>
                <div class="text-sm font-bold text-slate-800 dark:text-white">我的积分战绩</div>
                <div class="mt-0.5 text-[11px] text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                  排名
                  <span class="text-base font-extrabold text-amber-600 dark:text-amber-300 tabular-nums">
                    #<AnimatedNumber :value="me.rank" :duration="800" />
                  </span>
                  / {{ me.total_users }}
                  <span :class="['px-1.5 py-0.5 rounded-full text-[10px] font-medium', tierBadge(me.tier).badge]">
                    {{ tierBadge(me.tier).label }}段位
                  </span>
                </div>
              </div>
            </div>

            <button
              class="text-xs px-3.5 py-1.5 rounded-full border border-sky-900/10 text-slate-500 hover:text-slate-800 hover:border-amber-400/60 hover:bg-amber-50 dark:border-white/10 dark:text-slate-300 dark:hover:text-white dark:hover:border-amber-300/50 dark:hover:bg-white/5 transition-smooth"
              @click="showRules = !showRules"
            >
              {{ showRules ? '收起规则' : '积分规则' }}
            </button>
          </div>

          <!-- 数据块 -->
          <div class="mt-4 grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div class="glass-cell">
              <div class="text-[11px] text-slate-500 dark:text-slate-400">累计积分</div>
              <div class="text-xl font-extrabold text-amber-600 dark:text-amber-300 tabular-nums">
                <AnimatedNumber :value="me.total_points" :duration="900" />
              </div>
            </div>
            <div class="glass-cell">
              <div class="text-[11px] text-slate-500 dark:text-slate-400">今日已得</div>
              <div class="text-xl font-extrabold text-slate-800 dark:text-white tabular-nums">
                <AnimatedNumber :value="me.today_points" :duration="700" />
              </div>
            </div>
            <div class="glass-cell">
              <div class="text-[11px] text-slate-500 dark:text-slate-400">本周</div>
              <div class="text-xl font-extrabold text-slate-800 dark:text-white tabular-nums">
                <AnimatedNumber :value="me.week_points" :duration="700" />
              </div>
            </div>
            <div class="glass-cell">
              <div class="text-[11px] text-slate-500 dark:text-slate-400">本月</div>
              <div class="text-xl font-extrabold text-slate-800 dark:text-white tabular-nums">
                <AnimatedNumber :value="me.month_points" :duration="700" />
              </div>
            </div>
          </div>

          <!-- 近 14 天趋势（发光柱） -->
          <div class="mt-5">
            <div class="flex items-center justify-between mb-2">
              <span class="text-[11px] text-slate-500 dark:text-slate-400">近 14 天积分趋势</span>
              <span class="text-[10px] text-slate-400 dark:text-slate-500 font-mono">PEAK {{ trendMax }}</span>
            </div>
            <div class="flex items-end gap-1 sm:gap-1.5 h-20">
              <div
                v-for="t in me.trend"
                :key="t.date"
                class="group/bar flex-1 h-full flex flex-col justify-end"
                :title="`${t.date}：${t.points} 分`"
              >
                <div
                  :class="[
                    'w-full rounded-t transition-all duration-500',
                    t.points > 0
                      ? 'bg-gradient-to-t from-sky-400/50 via-sky-400/80 to-sky-300 group-hover/bar:from-blue-500 group-hover/bar:to-sky-200 dark:from-sky-500/50 dark:via-cyan-400/70 dark:to-cyan-200/90 dark:group-hover/bar:from-indigo-400 dark:group-hover/bar:to-white'
                      : 'bg-sky-900/10 dark:bg-white/10',
                  ]"
                  :style="{ height: Math.max(4, (t.points / trendMax) * 100) + '%' }"
                />
              </div>
            </div>
            <div class="mt-1.5 flex justify-between text-[9px] text-slate-400 dark:text-slate-500 font-mono">
              <span>{{ dayLabel(me.trend[0]?.date ?? '') }}</span>
              <span>今天</span>
            </div>
          </div>
        </div>
      </section>

      <!-- ===== 积分规则面板 ===== -->
      <div v-if="showRules" class="glass-card p-5 rise">
        <div class="card-top-line line-cyan" />
        <h3 class="text-sm font-semibold text-slate-800 dark:text-white mb-3">积分规则</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
          <div v-for="r in rules" :key="r.code" class="glass-cell flex items-center justify-between gap-2">
            <div class="min-w-0">
              <span class="text-sm text-slate-700 dark:text-slate-200">{{ r.name }}</span>
              <span
                class="ml-1.5 text-[10px] px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-600 ring-1 ring-sky-500/20 dark:bg-cyan-400/10 dark:text-cyan-300 dark:ring-cyan-400/20"
              >
                {{ RULE_CATEGORY_LABEL[r.category] || r.category }}
              </span>
            </div>
            <div class="shrink-0 text-right">
              <span class="text-sm font-bold text-amber-600 dark:text-amber-300">{{ rulePointsLabel(r) }}</span>
              <span v-if="r.daily_cap > 0" class="ml-1 text-[10px] text-slate-400 dark:text-slate-500">日限{{ r.daily_cap }}</span>
            </div>
          </div>
        </div>
        <p class="mt-3 text-xs text-slate-400 dark:text-slate-500">
          段位：青铜 &lt;200 · 白银 ≥200 · 黄金 ≥500 · 铂金 ≥1000（按累计积分） · 积分数据每分钟更新
        </p>
      </div>

      <!-- ===== 加载 / 错误 / 空状态 ===== -->
      <div v-if="loading" class="py-20 text-center">
        <div class="spinner mx-auto mb-4" aria-hidden="true" />
        <p class="text-sm text-slate-500 dark:text-slate-400">正在加载排行榜…</p>
      </div>
      <div
        v-else-if="errorMsg"
        class="py-16 text-center rounded-3xl border border-rose-300 bg-rose-50/80 text-rose-600 text-sm dark:border-rose-400/20 dark:bg-rose-500/5 dark:text-rose-300"
      >
        {{ errorMsg }}
      </div>
      <div
        v-else-if="list.length === 0"
        class="py-16 text-center rounded-3xl border border-sky-900/10 bg-white/60 dark:border-white/10 dark:bg-white/[0.03]"
      >
        <div
          class="mx-auto mb-4 flex w-16 h-16 rounded-2xl bg-white/70 ring-1 ring-sky-900/10 items-center justify-center dark:bg-white/5 dark:ring-white/10"
        >
          <AppIcon name="crown" :size="30" class="text-slate-400 dark:text-slate-500" />
        </div>
        <p class="text-sm text-slate-500 dark:text-slate-400">暂无积分数据，完成任务即可上榜</p>
      </div>

      <template v-else>
        <!-- ===== 领奖台：第一名居顶居中，第二名居左，第三名居右 ===== -->
        <div class="relative rise" style="animation-delay: 120ms">
          <!-- 第一名背后光柱 -->
          <div
            class="pointer-events-none absolute left-1/2 -translate-x-1/2 bottom-20 sm:bottom-28 w-40 sm:w-56 h-72 bg-gradient-to-t from-amber-300/40 via-amber-200/10 to-transparent blur-2xl dark:from-amber-400/20 dark:via-amber-300/5"
            aria-hidden="true"
          />

          <div class="relative flex items-end justify-center gap-4 sm:gap-8 pt-8 sm:pt-10">
            <!-- 第二名（左） -->
            <div v-if="podium.second" class="flex flex-col items-center w-1/4 max-w-[190px]">
              <div class="halo halo-2 mb-3">
                <div class="avatar-inner">
                  <img
                    v-if="podium.second.avatar_url"
                    :src="podium.second.avatar_url"
                    :alt="podium.second.name"
                    class="w-14 h-14 sm:w-16 sm:h-16 rounded-full object-cover"
                  />
                  <div
                    v-else
                    class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-gradient-to-br from-slate-400 to-slate-600 flex items-center justify-center text-xl font-bold text-white"
                  >
                    {{ initial(podium.second.name) }}
                  </div>
                </div>
              </div>
              <div class="text-sm font-semibold text-slate-800 dark:text-white truncate max-w-full">
                {{ podium.second.name }}
              </div>
              <div class="text-[11px] text-slate-500 dark:text-slate-400 truncate max-w-full">
                {{ podium.second.department || podium.second.position || '—' }}
              </div>
              <div class="mt-0.5 text-lg font-extrabold text-slate-700 dark:text-slate-200 tabular-nums">
                {{ podium.second.points }}
              </div>
              <div
                class="relative w-full h-20 sm:h-24 rounded-t-2xl border border-sky-900/10 border-b-0 bg-gradient-to-b from-white/80 to-white/30 backdrop-blur-sm overflow-hidden dark:border-white/10 dark:from-white/10 dark:to-white/[0.02]"
              >
                <div
                  class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-slate-400/50 to-transparent dark:via-slate-200/60"
                />
                <div
                  class="absolute inset-x-0 top-3 text-center text-3xl font-black text-slate-900/15 dark:text-slate-300/25 font-mono"
                >
                  2
                </div>
              </div>
            </div>

            <!-- 第一名（居中，最高） -->
            <div v-if="podium.first" class="flex flex-col items-center w-1/3 max-w-[230px] -mt-2 sm:-mt-4">
              <div class="float-crown mb-1">
                <AppIcon
                  name="crown"
                  variant="solid"
                  :size="34"
                  class="text-amber-400 drop-shadow-[0_0_12px_rgba(251,191,36,0.9)] dark:text-amber-300 dark:drop-shadow-[0_0_12px_rgba(251,191,36,0.8)]"
                />
              </div>
              <div class="halo mb-3">
                <div class="avatar-inner">
                  <img
                    v-if="podium.first.avatar_url"
                    :src="podium.first.avatar_url"
                    :alt="podium.first.name"
                    class="w-20 h-20 sm:w-24 sm:h-24 rounded-full object-cover"
                  />
                  <div
                    v-else
                    class="w-20 h-20 sm:w-24 sm:h-24 rounded-full bg-gradient-to-br from-amber-300 to-orange-600 flex items-center justify-center text-3xl font-bold text-white"
                  >
                    {{ initial(podium.first.name) }}
                  </div>
                </div>
              </div>
              <div class="text-base sm:text-lg font-bold text-slate-800 dark:text-white truncate max-w-full">
                {{ podium.first.name }}
              </div>
              <div class="text-[11px] text-slate-500 dark:text-slate-400 truncate max-w-full">
                {{ podium.first.department || podium.first.position || '—' }}
              </div>
              <div
                class="mt-1 px-3 py-1 rounded-full text-sm font-extrabold tabular-nums bg-gradient-to-r from-amber-400 to-orange-500 text-white shadow-lg shadow-amber-500/40 dark:from-amber-300 dark:to-orange-400 dark:text-slate-950 dark:shadow-amber-500/30"
              >
                {{ podium.first.points }} 分
              </div>
              <div
                class="relative w-full h-28 sm:h-36 rounded-t-2xl border border-amber-400/30 border-b-0 bg-gradient-to-b from-amber-200/50 to-white/25 backdrop-blur-sm overflow-hidden dark:border-amber-300/20 dark:from-amber-300/15 dark:to-white/[0.02]"
              >
                <!-- 向上光柱 -->
                <div
                  class="absolute bottom-full left-1/2 -translate-x-1/2 w-24 sm:w-32 h-40 rounded-full bg-gradient-to-t from-amber-300/45 to-transparent blur-xl dark:from-amber-300/30"
                  aria-hidden="true"
                />
                <div
                  class="absolute inset-x-0 top-0 h-0.5 bg-gradient-to-r from-transparent via-amber-400 to-transparent dark:via-amber-300"
                />
                <div class="podium-shine" aria-hidden="true" />
                <div
                  class="absolute inset-x-0 top-4 text-center text-5xl font-black bg-gradient-to-b from-amber-500 to-orange-600 bg-clip-text text-transparent dark:from-amber-200 dark:to-amber-500"
                >
                  1
                </div>
              </div>
            </div>

            <!-- 第三名（右） -->
            <div v-if="podium.third" class="flex flex-col items-center w-1/4 max-w-[190px]">
              <div class="halo halo-3 mb-3">
                <div class="avatar-inner">
                  <img
                    v-if="podium.third.avatar_url"
                    :src="podium.third.avatar_url"
                    :alt="podium.third.name"
                    class="w-14 h-14 sm:w-16 sm:h-16 rounded-full object-cover"
                  />
                  <div
                    v-else
                    class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-gradient-to-br from-orange-400 to-orange-700 flex items-center justify-center text-xl font-bold text-white"
                  >
                    {{ initial(podium.third.name) }}
                  </div>
                </div>
              </div>
              <div class="text-sm font-semibold text-slate-800 dark:text-white truncate max-w-full">
                {{ podium.third.name }}
              </div>
              <div class="text-[11px] text-slate-500 dark:text-slate-400 truncate max-w-full">
                {{ podium.third.department || podium.third.position || '—' }}
              </div>
              <div class="mt-0.5 text-lg font-extrabold text-orange-500 dark:text-orange-300 tabular-nums">
                {{ podium.third.points }}
              </div>
              <div
                class="relative w-full h-14 sm:h-16 rounded-t-2xl border border-orange-400/30 border-b-0 bg-gradient-to-b from-orange-200/40 to-white/25 backdrop-blur-sm overflow-hidden dark:border-orange-300/20 dark:from-orange-300/10 dark:to-white/[0.02]"
              >
                <div
                  class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-orange-400/60 to-transparent dark:via-orange-300/60"
                />
                <div
                  class="absolute inset-x-0 top-2 text-center text-xl font-black text-orange-900/20 dark:text-orange-300/25 font-mono"
                >
                  3
                </div>
              </div>
            </div>
          </div>

          <!-- 发光底座 -->
          <div class="relative max-w-2xl mx-auto">
            <div
              class="h-px bg-gradient-to-r from-transparent via-amber-400/80 to-transparent dark:via-amber-300/70"
            />
            <div
              class="h-6 bg-gradient-to-b from-amber-300/25 to-transparent blur-sm dark:from-amber-300/15"
              aria-hidden="true"
            />
          </div>
        </div>

        <!-- ===== 其余人员依次排列 ===== -->
        <div
          v-if="podium.rest.length"
          class="glass-card rise"
          style="animation-delay: 200ms"
        >
          <div class="divide-y divide-slate-900/5 dark:divide-white/5">
            <div
              v-for="(u, i) in podium.rest"
              :key="u.user_id"
              class="group flex items-center gap-3 sm:gap-4 px-4 sm:px-6 py-3.5 hover:bg-sky-50/80 dark:hover:bg-white/[0.04] transition-colors rise"
              :style="{ animationDelay: delayFor(i) }"
            >
              <span
                class="w-7 text-center font-mono text-sm font-bold text-slate-400 group-hover:text-amber-500 dark:text-slate-500 dark:group-hover:text-amber-300 transition-colors tabular-nums"
              >
                {{ u.rank }}
              </span>

              <div class="relative shrink-0">
                <img
                  v-if="u.avatar_url"
                  :src="u.avatar_url"
                  :alt="u.name"
                  class="w-9 h-9 rounded-full object-cover ring-1 ring-sky-900/10 dark:ring-white/15"
                />
                <div
                  v-else
                  class="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-500 to-cyan-500 flex items-center justify-center text-sm font-bold text-white ring-1 ring-sky-900/10 dark:ring-white/15"
                >
                  {{ initial(u.name) }}
                </div>
                <span
                  v-if="u.points > 0"
                  class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full bg-emerald-500 ring-2 ring-white dark:bg-emerald-400 dark:ring-[#0a0f22]"
                  title="已获得积分"
                />
              </div>

              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="text-sm font-semibold text-slate-800 dark:text-white truncate">{{ u.name }}</span>
                  <span :class="['text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0', tierBadge(u.tier).badge]">
                    {{ tierBadge(u.tier).label }}
                  </span>
                </div>
                <div class="mt-1 flex items-center gap-2.5">
                  <div class="h-1 flex-1 max-w-[220px] rounded-full bg-sky-900/10 dark:bg-white/5 overflow-hidden">
                    <div
                      :class="['h-full rounded-full bg-gradient-to-r transition-all duration-700', tierBadge(u.tier).bar]"
                      :style="{ width: barWidth(u.points) }"
                    />
                  </div>
                  <span class="hidden sm:inline text-[11px] text-slate-400 dark:text-slate-500 truncate">
                    {{ [u.department, u.position].filter(Boolean).join(' · ') || '—' }}
                  </span>
                </div>
              </div>

              <div class="text-right shrink-0">
                <div class="text-base font-extrabold text-slate-800 dark:text-white tabular-nums">{{ u.points }}</div>
                <div v-if="u.points > 0" class="text-[10px] text-slate-400 dark:text-slate-500">距上名 {{ u.delta_to_prev }} 分</div>
                <div v-else class="text-[10px] text-slate-300 dark:text-slate-600">暂未得分</div>
              </div>
            </div>
          </div>
        </div>
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
  background: linear-gradient(to right, transparent, rgba(251, 191, 36, 0.7), transparent);
}
.card-top-line.line-cyan {
  background: linear-gradient(to right, transparent, rgba(14, 165, 233, 0.5), transparent);
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

/* ===== 领奖台旋转光环 ===== */
.halo {
  position: relative;
  border-radius: 9999px;
  padding: 3px;
  background: conic-gradient(from 0deg, #fbbf24, #f59e0b, #fef3c7, #ffffff, #fde68a, #fbbf24);
  animation: halo-spin 7s linear infinite;
  box-shadow:
    0 0 20px rgba(251, 191, 36, 0.5),
    0 0 48px rgba(251, 146, 60, 0.2);
}
.halo-2 {
  background: conic-gradient(from 0deg, #cbd5e1, #94a3b8, #f1f5f9, #e2e8f0, #cbd5e1);
  box-shadow: 0 0 16px rgba(148, 163, 184, 0.45);
  animation-duration: 9s;
}
.halo-3 {
  background: conic-gradient(from 0deg, #fb923c, #ea580c, #fed7aa, #fdba74, #fb923c);
  box-shadow: 0 0 16px rgba(251, 146, 60, 0.45);
  animation-duration: 11s;
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

/* ===== 皇冠浮动 ===== */
.float-crown {
  animation: floaty 3s ease-in-out infinite;
}
@keyframes floaty {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

/* ===== 奖台扫光 ===== */
.podium-shine {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -60%;
  width: 40%;
  background: linear-gradient(105deg, transparent, rgba(255, 255, 255, 0.35), transparent);
  transform: skewX(-15deg);
  animation: shine 5s ease-in-out infinite;
}
.dark .podium-shine {
  background: linear-gradient(105deg, transparent, rgba(255, 255, 255, 0.08), transparent);
}
@keyframes shine {
  0% {
    left: -60%;
  }
  60%,
  100% {
    left: 130%;
  }
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

/* ===== 加载 spinner ===== */
.spinner {
  width: 36px;
  height: 36px;
  border-radius: 9999px;
  border: 3px solid rgba(100, 116, 139, 0.2);
  border-top-color: #0ea5e9;
  animation: halo-spin 0.9s linear infinite;
}

/* ===== 降低动态效果偏好 ===== */
@media (prefers-reduced-motion: reduce) {
  .aurora,
  .halo,
  .halo-2,
  .halo-3,
  .float-crown,
  .podium-shine,
  .live-dot,
  .spinner {
    animation: none !important;
  }
  .rise {
    animation: none !important;
    opacity: 1 !important;
  }
}
</style>
