import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as achievementService from '@/services/achievements'
import type {
  AchievementItem,
  AchievementCategory,
  NewlyUnlocked,
} from '@/types/achievement'

/** 成就弹窗展示时长（用户要求：2 秒恭喜弹窗） */
export const CELEBRATION_DURATION = 2000
/** 成就检查节流间隔（避免连续操作触发过多请求） */
const CHECK_THROTTLE = 3000

export const useAchievementStore = defineStore('achievements', () => {
  // ---------------- 成就列表 ----------------
  const categories = ref<AchievementCategory[]>([])
  const achievements = ref<AchievementItem[]>([])
  const total = ref(0)
  const unlockedCount = ref(0)
  const loading = ref(false)

  // ---------------- 恭喜弹窗（队列 + 当前展示） ----------------
  const celebrationQueue = ref<NewlyUnlocked[]>([])
  const currentCelebration = ref<NewlyUnlocked | null>(null)
  let pumping = false
  let celebrationTimer: number | null = null

  // 进度总数（总成就 vs 已解锁）
  const unlockedPercent = computed(() =>
    total.value > 0 ? Math.round((unlockedCount.value / total.value) * 100) : 0
  )

  /** 拉取成就列表（成就殿堂页 + 侧边栏进度） */
  async function fetchList() {
    if (loading.value) return
    loading.value = true
    try {
      const res = await achievementService.fetchAchievements()
      const d = res.data
      if (d) {
        categories.value = d.categories || []
        achievements.value = d.achievements || []
        total.value = d.total || 0
        unlockedCount.value = d.unlockedCount || 0
      }
    } finally {
      loading.value = false
    }
  }

  /** 单个成就进度百分比（0-100，target 为 0 时按解锁状态） */
  function progressPercent(item: AchievementItem): number {
    if (item.unlocked) return 100
    if (item.target <= 0) return 0
    return Math.min(100, Math.round((item.progress / item.target) * 100))
  }

  // ---------------- 弹窗队列泵 ----------------

  function showNext() {
    if (pumping) return
    pumping = true
    const step = () => {
      const next = celebrationQueue.value.shift()
      if (!next) {
        currentCelebration.value = null
        pumping = false
        return
      }
      currentCelebration.value = next
      celebrationTimer = window.setTimeout(() => {
        celebrationTimer = null
        step()
      }, CELEBRATION_DURATION)
    }
    step()
  }

  /** 立即关闭当前弹窗并展示下一个 */
  function dismissCelebration() {
    if (celebrationTimer !== null) {
      window.clearTimeout(celebrationTimer)
      celebrationTimer = null
    }
    currentCelebration.value = null
    pumping = false
    showNext()
  }

  /** 一次成功检查后入队新解锁并开始弹窗 */
  function enqueueCelebrations(list: NewlyUnlocked[]) {
    if (!list.length) return
    celebrationQueue.value.push(...list)
    // 已解锁数量即时 +list.length（无需重新拉列表）
    unlockedCount.value += list.length
    showNext()
  }

  // ---------------- 成就检查 ----------------

  let lastCheckAt = 0
  let checking = false

  /**
   * 触发后端成就检查；新解锁的成就进入 2s 恭喜弹窗队列。
   * 关键动作（创建/完成任务、发消息、写报告、登录等）后调用。
   */
  async function checkAndNotify(force = false): Promise<NewlyUnlocked[]> {
    const now = Date.now()
    if (!force && now - lastCheckAt < CHECK_THROTTLE) return []
    if (checking) return []
    checking = true
    lastCheckAt = now
    try {
      const res = await achievementService.checkAchievements()
      const newly = res.data?.newly_unlocked || []
      enqueueCelebrations(newly)
      return newly
    } catch {
      return []
    } finally {
      checking = false
    }
  }

  /** 清空弹窗状态（登出时调用） */
  function reset() {
    if (celebrationTimer !== null) {
      window.clearTimeout(celebrationTimer)
      celebrationTimer = null
    }
    celebrationQueue.value = []
    currentCelebration.value = null
    pumping = false
    achievements.value = []
    categories.value = []
    total.value = 0
    unlockedCount.value = 0
    lastCheckAt = 0
  }

  return {
    // 列表
    categories,
    achievements,
    total,
    unlockedCount,
    unlockedPercent,
    loading,
    // 弹窗
    celebrationQueue,
    currentCelebration,
    // 方法
    fetchList,
    progressPercent,
    checkAndNotify,
    dismissCelebration,
    reset,
  }
})
