<script setup lang="ts">
/* 成就解锁 2s 恭喜弹窗（全局单例，挂载于 App.vue）
   成就解锁后从顶部弹性滑入，展示 2 秒后自动切换下一条（队列泵在 store 中）。
   稀有度不同配色不同：青铜/白银/黄金/铂金。 */
import { computed, watch } from 'vue'
import confetti from 'canvas-confetti'
import { storeToRefs } from 'pinia'
import { useAchievementStore } from '@/stores/achievements'
import AppIcon from '@/components/icons/AppIcon.vue'
import type { AchievementTier } from '@/types/achievement'

const store = useAchievementStore()
const { currentCelebration } = storeToRefs(store)

/** 稀有度 → 配色 / 中文标签 */
const TIER_STYLE: Record<AchievementTier, { ring: string; badge: string; label: string }> = {
  bronze: {
    ring: 'from-amber-500 to-orange-700',
    badge: 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-300',
    label: '青铜',
  },
  silver: {
    ring: 'from-slate-300 to-slate-500',
    badge: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
    label: '白银',
  },
  gold: {
    ring: 'from-yellow-300 to-amber-500',
    badge: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/60 dark:text-yellow-300',
    label: '黄金',
  },
  platinum: {
    ring: 'from-cyan-300 to-violet-500',
    badge: 'bg-violet-100 text-violet-800 dark:bg-violet-900/60 dark:text-violet-300',
    label: '铂金',
  },
}

const tierStyle = computed(() =>
  TIER_STYLE[currentCelebration.value?.tier || 'bronze']
)

// 每次新成就弹出：顶部小纸屑喷发（铂金/黄金更盛大），克制不遮挡操作
watch(currentCelebration, (n) => {
  if (!n) return
  const colors = n.tier === 'platinum' ? ['#22d3ee', '#a78bfa', '#f0abfc', '#fbbf24'] : n.tier === 'gold' ? ['#fbbf24', '#f59e0b', '#fde68a'] : ['#94a3b8', '#cbd5e1', '#f59e0b']
  confetti({
    particleCount: n.tier === 'bronze' ? 40 : 70,
    angle: 90,
    spread: n.tier === 'platinum' ? 100 : 70,
    startVelocity: 38,
    origin: { x: 0.5, y: -0.05 },
    colors,
    ticks: 120,
    zIndex: 95,
    disableForReducedMotion: true,
  })
})
</script>

<template>
  <Teleport to="body">
    <Transition name="achv-pop">
      <div
        v-if="currentCelebration"
        :key="currentCelebration.code"
        class="fixed top-5 left-1/2 -translate-x-1/2 z-[92] pointer-events-none"
        role="status"
        aria-live="polite"
      >
        <div
          class="flex items-center gap-4 rounded-2xl bg-white dark:bg-slate-800 shadow-2xl border border-slate-100 dark:border-slate-700 px-5 py-4 min-w-[340px] max-w-[420px]"
        >
          <!-- 奖章：稀有度渐变底 + 图标 -->
          <div
            class="achv-medal flex w-14 h-14 rounded-2xl bg-gradient-to-br items-center justify-center shrink-0 text-white shadow-lg"
            :class="tierStyle.ring"
          >
            <AppIcon :name="(currentCelebration.icon as never)" variant="solid" :size="28" />
          </div>

          <!-- 文案区 -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold tracking-widest text-blue-600 dark:text-blue-400">
                🏆 成就解锁
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded-full font-medium"
                :class="tierStyle.badge"
              >
                {{ tierStyle.label }}
              </span>
            </div>
            <p class="mt-0.5 text-base font-bold text-slate-800 dark:text-slate-100 truncate">
              {{ currentCelebration.name }}
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400 truncate">
              {{ currentCelebration.desc }}
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 顶部弹性滑入 / 滑出 */
.achv-pop-enter-active {
  transition:
    transform 0.45s cubic-bezier(0.34, 1.56, 0.64, 1),
    opacity 0.3s ease;
}
.achv-pop-leave-active {
  transition:
    transform 0.3s ease-in,
    opacity 0.3s ease-in;
}
.achv-pop-enter-from {
  transform: translate(-50%, -110%) scale(0.9);
  opacity: 0;
}
.achv-pop-leave-to {
  transform: translate(-50%, -130%) scale(0.95);
  opacity: 0;
}
/* 奖章落位微弹跳 */
.achv-medal {
  animation: achv-medal-bounce 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes achv-medal-bounce {
  0% {
    transform: scale(0.3) rotate(-12deg);
  }
  60% {
    transform: scale(1.12) rotate(4deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .achv-medal {
    animation: none;
  }
}
</style>
