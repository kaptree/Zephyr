// 成就系统类型定义（与后端 internal/achievements 对齐）

/** 成就稀有度 */
export type AchievementTier = 'bronze' | 'silver' | 'gold' | 'platinum'

/** 成就定义 */
export interface AchievementDef {
  code: string
  name: string
  desc: string
  /** 分类 key：journey/efficiency/color/social/command/hunter/explorer/personality */
  category: string
  /** 前端图标名（registry 图标注册表） */
  icon: string
  tier: AchievementTier
  /** 达成阈值 */
  target: number
  /** 彩蛋成就：解锁前隐藏达成条件 */
  is_mystery: boolean
}

/** 成就条目（定义 + 解锁状态 + 进度） */
export interface AchievementItem extends AchievementDef {
  unlocked: boolean
  /** 解锁时间 yyyy-MM-dd HH:mm:ss */
  unlocked_at?: string
  /** 当前进度值（彩蛋未解锁时为 0） */
  progress: number
}

/** 成就分类元信息 */
export interface AchievementCategory {
  key: string
  name: string
  icon: string
  desc: string
}

/** GET /achievements 响应 */
export interface AchievementListResp {
  categories: AchievementCategory[]
  achievements: AchievementItem[]
  total: number
  unlockedCount: number
}

/** POST /achievements/check 返回的新解锁条目（弹窗数据源） */
export interface NewlyUnlocked {
  code: string
  name: string
  desc: string
  category: string
  icon: string
  tier: AchievementTier
  target: number
  is_mystery: boolean
  /** 解锁时间 yyyy-MM-dd HH:mm:ss */
  unlocked_at: string
}
