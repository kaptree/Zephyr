import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAchievementStore, CELEBRATION_DURATION } from '@/stores/achievements'
import type { NewlyUnlocked } from '@/types/achievement'

// mock 成就 API 服务
vi.mock('@/services/achievements', () => ({
  fetchAchievements: vi.fn().mockResolvedValue({
    code: 0,
    message: 'ok',
    data: {
      categories: [{ key: 'journey', name: '任务征程', icon: 'clipboard', desc: '从第一条便签到任务帝国' }],
      achievements: [
        {
          code: 'note_create_1',
          name: '第一步',
          desc: '创建第一个便签',
          category: 'journey',
          icon: 'pencil',
          tier: 'bronze',
          target: 1,
          is_mystery: false,
          unlocked: true,
          progress: 1,
        },
        {
          code: 'note_create_100',
          name: '便签大军',
          desc: '累计创建 100 个便签',
          category: 'journey',
          icon: 'star',
          tier: 'gold',
          target: 100,
          is_mystery: false,
          unlocked: false,
          progress: 42,
        },
        {
          code: 'late_complete_1',
          name: '夜猫子',
          desc: '在凌晨 0-5 点完成一个任务',
          category: 'personality',
          icon: 'moon',
          tier: 'gold',
          target: 1,
          is_mystery: true,
          unlocked: false,
          progress: 0,
        },
      ],
      total: 100,
      unlockedCount: 1,
    },
  }),
  checkAchievements: vi.fn().mockResolvedValue({
    code: 0,
    message: 'ok',
    data: { newly_unlocked: [] as NewlyUnlocked[] },
  }),
}))

import * as achievementService from '@/services/achievements'

function makeUnlock(code: string, tier: NewlyUnlocked['tier'] = 'bronze'): NewlyUnlocked {
  return {
    code,
    name: `成就 ${code}`,
    desc: '达成条件描述',
    category: 'journey',
    icon: 'star',
    tier,
    target: 1,
    is_mystery: false,
    unlocked_at: '2026-09-18 12:00:00',
  }
}

describe('achievements store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('fetchList 拉取成就列表并填充总数与解锁数', async () => {
    const store = useAchievementStore()
    await store.fetchList()
    expect(store.total).toBe(100)
    expect(store.unlockedCount).toBe(1)
    expect(store.achievements).toHaveLength(3)
    expect(store.categories).toHaveLength(1)
  })

  it('progressPercent：未解锁按进度/target 计算，已解锁恒为 100', async () => {
    const store = useAchievementStore()
    await store.fetchList()
    const [done, doing] = store.achievements
    expect(store.progressPercent(done)).toBe(100)
    expect(store.progressPercent(doing)).toBe(42)
  })

  it('checkAndNotify：新解锁成就进入弹窗队列，2s 后自动切到下一条', async () => {
    vi.mocked(achievementService.checkAchievements).mockResolvedValueOnce({
      code: 0,
      message: 'ok',
      data: { newly_unlocked: [makeUnlock('a1'), makeUnlock('a2')] },
    })
    const store = useAchievementStore()
    const newly = await store.checkAndNotify()
    expect(newly).toHaveLength(2)
    // 第一条立即展示
    expect(store.currentCelebration?.code).toBe('a1')
    // 2s 后切到第二条
    vi.advanceTimersByTime(CELEBRATION_DURATION)
    expect(store.currentCelebration?.code).toBe('a2')
    // 队列清空后关闭弹窗
    vi.advanceTimersByTime(CELEBRATION_DURATION)
    expect(store.currentCelebration).toBeNull()
  })

  it('checkAndNotify：3s 节流窗口内重复调用不再请求后端', async () => {
    const store = useAchievementStore()
    await store.checkAndNotify()
    await store.checkAndNotify()
    expect(achievementService.checkAchievements).toHaveBeenCalledTimes(1)
    // 超过节流窗口后可再次请求
    vi.advanceTimersByTime(3100)
    await store.checkAndNotify()
    expect(achievementService.checkAchievements).toHaveBeenCalledTimes(2)
  })

  it('dismissCelebration：立即关闭当前并展示下一条', async () => {
    vi.mocked(achievementService.checkAchievements).mockResolvedValueOnce({
      code: 0,
      message: 'ok',
      data: { newly_unlocked: [makeUnlock('x1'), makeUnlock('x2')] },
    })
    const store = useAchievementStore()
    await store.checkAndNotify()
    expect(store.currentCelebration?.code).toBe('x1')
    store.dismissCelebration()
    expect(store.currentCelebration?.code).toBe('x2')
  })

  it('新解锁会累加本地 unlockedCount（无需重新拉列表）', async () => {
    vi.mocked(achievementService.checkAchievements).mockResolvedValueOnce({
      code: 0,
      message: 'ok',
      data: { newly_unlocked: [makeUnlock('n1'), makeUnlock('n2'), makeUnlock('n3')] },
    })
    const store = useAchievementStore()
    await store.fetchList()
    expect(store.unlockedCount).toBe(1)
    await store.checkAndNotify()
    expect(store.unlockedCount).toBe(4)
  })

  it('reset：清空弹窗与列表状态', async () => {
    vi.mocked(achievementService.checkAchievements).mockResolvedValueOnce({
      code: 0,
      message: 'ok',
      data: { newly_unlocked: [makeUnlock('r1')] },
    })
    const store = useAchievementStore()
    await store.fetchList()
    await store.checkAndNotify()
    expect(store.currentCelebration).not.toBeNull()
    store.reset()
    expect(store.currentCelebration).toBeNull()
    expect(store.celebrationQueue).toHaveLength(0)
    expect(store.achievements).toHaveLength(0)
    expect(store.total).toBe(0)
  })
})
