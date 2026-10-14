<script setup lang="ts">
/* 成就殿堂 —— 100 个成就按 8 大分类展示
   - 顶部总览：解锁进度（数字滚动 + 进度条）
   - 分类标签页切换 + 未解锁/已解锁筛选
   - 卡片：稀有度配色奖章 + 进度条；彩蛋成就解锁前隐藏条件（？？？） */
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAchievementStore } from '@/stores/achievements'
import AppIcon from '@/components/icons/AppIcon.vue'
import type { AchievementItem, AchievementTier } from '@/types'
import AnimatedNumber from '@/components/common/AnimatedNumber.vue'

const store = useAchievementStore()
const { categories, achievements, total, unlockedCount, unlockedPercent, loading } =
  storeToRefs(store)

onMounted(() => {
  store.fetchList()
})

// ---------------- 稀有度配色 ----------------
const TIER_STYLE: Record<
  AchievementTier,
  { ring: string; badge: string; label: string; unlockedIcon: string }
> = {
  bronze: {
    ring: 'from-amber-500 to-orange-700',
    badge: 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-300',
    label: '青铜',
    unlockedIcon: 'text-amber-500',
  },
  silver: {
    ring: 'from-slate-300 to-slate-500',
    badge: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
    label: '白银',
    unlockedIcon: 'text-slate-400',
  },
  gold: {
    ring: 'from-yellow-300 to-amber-500',
    badge: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/60 dark:text-yellow-300',
    label: '黄金',
    unlockedIcon: 'text-yellow-500',
  },
  platinum: {
    ring: 'from-cyan-300 to-violet-500',
    badge: 'bg-violet-100 text-violet-800 dark:bg-violet-900/60 dark:text-violet-300',
    label: '铂金',
    unlockedIcon: 'text-violet-400',
  },
}

// ---------------- 分类切换 ----------------
type TabKey = 'all' | string
const activeTab = ref<TabKey>('all')
const showUnlockedOnly = ref(false)

const filtered = computed(() => {
  let list = achievements.value
  if (activeTab.value !== 'all') list = list.filter((a) => a.category === activeTab.value)
  if (showUnlockedOnly.value) list = list.filter((a) => a.unlocked)
  return list
})

/** 分类分组结果（带分类元信息与解锁统计） */
const grouped = computed(() =>
  categories.value.map((c) => {
    const items = filtered.value.filter((a) => a.category === c.key)
    const all = achievements.value.filter((a) => a.category === c.key)
    return {
      ...c,
      items,
      unlocked: all.filter((a) => a.unlocked).length,
      count: all.length,
    }
  }).filter((g) => g.items.length > 0)
)

function descOf(a: AchievementItem): string {
  if (a.is_mystery && !a.unlocked) return '？？？神秘成就，完成某个隐藏条件后揭晓'
  return a.desc
}

function fmtDate(s?: string): string {
  if (!s) return ''
  return s.replace(' ', ' · ')
}
</script>

<template>
  <div class="page-shell min-h-full">
    <!-- ===== 背景装饰层（日间晨曦 / 夜间深空） ===== -->
    <div class="decor-layer" aria-hidden="true">
      <div class="aurora aurora-1" />
      <div class="aurora aurora-2" />
      <div class="aurora aurora-3" />
      <div class="grid-overlay" />
      <div class="stars" />
      <div class="horizon-glow" />
      <div class="sun-glow" />
    </div>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 py-6 space-y-6">
      <!-- ===== 顶部总览横幅 ===== -->
      <div
        class="rise relative overflow-hidden rounded-2xl bg-gradient-to-br from-blue-600 via-indigo-600 to-violet-600 dark:from-blue-800 dark:via-indigo-800 dark:to-violet-800 p-6 sm:p-8 text-white shadow-xl"
      >
        <!-- 装饰光斑 -->
        <div class="absolute -top-16 -right-16 w-56 h-56 rounded-full bg-white/10 blur-2xl" />
        <div class="absolute -bottom-20 -left-10 w-48 h-48 rounded-full bg-white/10 blur-2xl" />

        <div class="relative flex flex-col sm:flex-row sm:items-center gap-6">
          <!-- 大奖章 -->
          <div
            class="flex w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-white/15 backdrop-blur items-center justify-center shrink-0 shadow-inner"
          >
            <AppIcon name="crown" variant="solid" :size="40" />
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold tracking-tight">成就殿堂</h1>
            <p class="mt-1 text-sm text-blue-100/90">
              已点亮
              <span class="text-2xl font-extrabold mx-0.5">
                <AnimatedNumber :value="unlockedCount" :duration="800" />
              </span>
              / {{ total }} 个成就，继续加油！
            </p>

            <!-- 总进度条 -->
            <div class="mt-4 flex items-center gap-3">
              <div class="flex-1 h-2.5 rounded-full bg-white/20 overflow-hidden">
                <div
                  class="h-full rounded-full bg-gradient-to-r from-yellow-300 to-amber-400 transition-all duration-700"
                  :style="{ width: unlockedPercent + '%' }"
                />
              </div>
              <span class="text-sm font-semibold tabular-nums">{{ unlockedPercent }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== 筛选栏 ===== -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- 全部 -->
        <button
          :class="[
            'px-3.5 py-1.5 rounded-full text-sm font-medium transition-smooth',
            activeTab === 'all'
              ? 'bg-blue-600 text-white shadow-md shadow-blue-500/25'
              : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700 hover:border-blue-300',
          ]"
          @click="activeTab = 'all'"
        >
          全部 {{ total }}
        </button>
        <!-- 分类 -->
        <button
          v-for="c in categories"
          :key="c.key"
          :class="[
            'px-3.5 py-1.5 rounded-full text-sm font-medium transition-smooth flex items-center gap-1.5',
            activeTab === c.key
              ? 'bg-blue-600 text-white shadow-md shadow-blue-500/25'
              : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700 hover:border-blue-300',
          ]"
          @click="activeTab = activeTab === c.key ? 'all' : c.key"
        >
          <AppIcon :name="(c.icon as never)" :size="14" />
          {{ c.name }}
        </button>

        <div class="flex-1" />

        <!-- 已解锁筛选 -->
        <label
          class="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-300 cursor-pointer select-none"
        >
          <input v-model="showUnlockedOnly" type="checkbox" class="accent-blue-600 w-4 h-4" />
          只看已解锁
        </label>
      </div>

      <!-- ===== 加载骨架 ===== -->
      <div v-if="loading && achievements.length === 0" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="i in 6"
          :key="i"
          class="h-40 rounded-2xl bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 animate-pulse"
        />
      </div>

      <!-- ===== 分类分组 ===== -->
      <section v-for="g in grouped" :key="g.key" class="space-y-3">
        <div class="flex items-center gap-3">
          <span
            class="flex w-9 h-9 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 items-center justify-center text-blue-600 dark:text-blue-400 shadow-sm"
          >
            <AppIcon :name="(g.icon as never)" :size="18" />
          </span>
          <div class="min-w-0">
            <h2 class="text-base font-bold text-slate-800 dark:text-slate-100">{{ g.name }}</h2>
            <p class="text-xs text-slate-400">{{ g.desc }}</p>
          </div>
          <div class="flex-1" />
          <span class="text-xs font-medium text-slate-400 tabular-nums shrink-0">
            {{ g.unlocked }}/{{ g.count }}
          </span>
        </div>

        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <!-- 成就卡片 -->
          <div
            v-for="a in g.items"
            :key="a.code"
            :class="[
              'group relative rounded-2xl border p-4 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg',
              a.unlocked
                ? 'bg-white dark:bg-slate-800 border-slate-100 dark:border-slate-700 shadow-sm'
                : 'bg-slate-100/60 dark:bg-slate-900/60 border-dashed border-slate-200 dark:border-slate-800',
            ]"
          >
            <!-- 已解锁角标 -->
            <span
              v-if="a.unlocked"
              class="absolute top-3 right-3 flex items-center justify-center w-6 h-6 rounded-full bg-emerald-100 dark:bg-emerald-900/60"
            >
              <AppIcon name="check" :size="14" class="text-emerald-600 dark:text-emerald-400" />
            </span>
            <!-- 彩蛋标识 -->
            <span
              v-else-if="a.is_mystery"
              class="absolute top-3 right-3 text-sm"
              title="神秘成就"
              >✨</span
            >

            <div class="flex items-start gap-3">
              <!-- 奖章 -->
              <div
                :class="[
                  'flex w-12 h-12 rounded-xl items-center justify-center shrink-0 shadow-sm transition-smooth',
                  a.unlocked
                    ? `bg-gradient-to-br text-white ${TIER_STYLE[a.tier].ring} group-hover:scale-110`
                    : 'bg-slate-200 dark:bg-slate-700/60 text-slate-400 dark:text-slate-500',
                ]"
              >
                <AppIcon :name="(a.icon as never)" :variant="a.unlocked ? 'solid' : 'outline'" :size="24" />
              </div>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <h3
                    :class="[
                      'text-sm font-bold truncate',
                      a.unlocked
                        ? 'text-slate-800 dark:text-slate-100'
                        : 'text-slate-500 dark:text-slate-400',
                    ]"
                  >
                    {{ a.unlocked || !a.is_mystery ? a.name : '？？？' }}
                  </h3>
                  <span
                    :class="[
                      'text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0',
                      TIER_STYLE[a.tier].badge,
                    ]"
                  >
                    {{ TIER_STYLE[a.tier].label }}
                  </span>
                </div>

                <p class="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400 line-clamp-2">
                  {{ descOf(a) }}
                </p>

                <!-- 进度 -->
                <div class="mt-2.5">
                  <div class="h-1.5 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
                    <div
                      :class="[
                        'h-full rounded-full transition-all duration-500',
                        a.unlocked
                          ? 'bg-gradient-to-r from-emerald-400 to-emerald-500'
                          : 'bg-gradient-to-r from-blue-400 to-blue-500',
                      ]"
                      :style="{ width: store.progressPercent(a) + '%' }"
                    />
                  </div>
                  <div class="mt-1.5 flex items-center justify-between text-[11px]">
                    <span v-if="a.unlocked" class="text-emerald-600 dark:text-emerald-400 font-medium">
                      {{ fmtDate(a.unlocked_at) }} 达成
                    </span>
                    <span v-else class="text-slate-400 tabular-nums">
                      {{ a.is_mystery && !a.unlocked ? '进度保密' : `${a.progress} / ${a.target}` }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 空态（筛选无结果） -->
      <div
        v-if="!loading && achievements.length > 0 && grouped.length === 0"
        class="text-center py-16 text-slate-400"
      >
        <p class="text-4xl mb-3">🔍</p>
        <p class="text-sm">没有符合筛选条件的成就</p>
      </div>
    </div>
  </div>
</template>
