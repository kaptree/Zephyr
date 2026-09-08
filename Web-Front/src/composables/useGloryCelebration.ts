/* 归档「荣耀时刻」庆祝（美化工程 · 任务闭环仪式感）
   模块级单例（同 useToast 模式），全局共享一份庆祝状态。
   三阶段时序（总计 3.5s）：
     绽放 0–1.2s   纸屑爆发 + 勋章弹跳落位 + 对勾描画
     升华 1.2–2.5s 主标题弹性浮现 + 数字滚动 + 背景气泡
     余韵 2.5–3.5s 纸屑/勋章消散 + 「查看成果」按钮浮现 + 通知呼应
   视觉演出由 GloryCelebration.vue 的 CSS 时间线承担（animation-delay 编排），
   本组合式仅负责：状态机、纸屑触发、用户偏好开关、焦点管理与自动收场。 */
import { ref } from 'vue'
import confetti from 'canvas-confetti'

/** 关闭庆祝动画的用户偏好（localStorage 持久化，尊重用户选择） */
const STORAGE_KEY = 'zephyr-celebration-disabled'

/** 庆祝数据 */
export interface GloryCelebration {
  taskName: string
  /** 累计完成任务数（归档后的新值），缺省则不展示数字滚动 */
  countTo?: number
  /** 点击「查看成果」回调（如跳转任务详情/统计面板） */
  onView?: () => void
}

/** 庆祝层可见性与演出数据（全局单例） */
const visible = ref(false)
const celebration = ref<GloryCelebration | null>(null)
/** 升华阶段标记：1.2s 后挂载数字滚动组件，让计数与标题同步登场 */
const countStarted = ref(false)

/** 用户偏好：是否开启庆祝动画（默认开启） */
const celebrationEnabled = ref(localStorage.getItem(STORAGE_KEY) !== '1')

function setCelebrationEnabled(enabled: boolean) {
  celebrationEnabled.value = enabled
  if (enabled) localStorage.removeItem(STORAGE_KEY)
  else localStorage.setItem(STORAGE_KEY, '1')
}

/** 纸屑颜色：取自品牌主色及亮色变体（蓝/紫/青/粉/金/绿），与主题色协调 */
const CONFETTI_LIGHT = ['#3B82F6', '#8B5CF6', '#22D3EE', '#EC4899', '#F59E0B', '#4ADE80', '#FDE047']
/** 暗色模式：纸屑颜色加深一档，在深色背景上保持饱和度协调 */
const CONFETTI_DARK = ['#2563EB', '#7C3AED', '#0891B2', '#DB2777', '#D97706', '#16A34A', '#FBBF24']

let resolveRun: ((played: boolean) => void) | null = null
let countTimer = 0
let closeTimer = 0
let wingTimer = 0
let previousFocus: HTMLElement | null = null

/** 五彩纸屑：中央主爆发（150 片）+ 左右两翼补发，带重力与旋转（canvas-confetti 自带） */
function fireConfetti() {
  const isDark = document.documentElement.classList.contains('dark')
  const colors = isDark ? CONFETTI_DARK : CONFETTI_LIGHT
  const zIndex = 2147483000 // 盖过庆祝层 z-[90] 与反馈弹窗 z-[80]
  confetti({
    particleCount: 150,
    spread: 70,
    origin: { y: 0.6 },
    colors,
    startVelocity: 38,
    gravity: 0.9,
    scalar: 0.9,
    ticks: 240,
    zIndex,
    disableForReducedMotion: true,
  })
  // 0.25s 后左右两翼斜向补发，营造「绽放」的空间层次
  wingTimer = window.setTimeout(() => {
    confetti({ particleCount: 60, angle: 60, spread: 55, origin: { x: 0, y: 0.7 }, colors, ticks: 240, zIndex, disableForReducedMotion: true })
    confetti({ particleCount: 60, angle: 120, spread: 55, origin: { x: 1, y: 0.7 }, colors, ticks: 240, zIndex, disableForReducedMotion: true })
  }, 250)
}

function clearTimers() {
  window.clearTimeout(countTimer)
  window.clearTimeout(closeTimer)
  window.clearTimeout(wingTimer)
}

/**
 * 启动庆祝序列。返回 Promise<boolean>：
 * - true ：完整播放了庆祝动画（3.5s 后自动收场，或用户点击「查看成果」提前收场）
 * - false：用户已关闭庆祝开关 / 系统偏好减少动效，调用方应走普通成功反馈路径
 */
function celebrate(opts: GloryCelebration): Promise<boolean> {
  // 用户偏好或 prefers-reduced-motion：整段动画跳过（canvas-confetti 也自带降级）
  if (
    !celebrationEnabled.value ||
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  ) {
    return Promise.resolve(false)
  }
  // 防重入：上一次尚未收场时先静默结束
  if (visible.value) finish(false)
  celebration.value = opts
  countStarted.value = false
  visible.value = true
  // 焦点管理：记住触发源，收场后归还，保证键盘导航连贯
  previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  // 绽放（0s）：纸屑爆发
  fireConfetti()
  // 升华（1.2s）：数字滚动与副标题同步登场
  countTimer = window.setTimeout(() => {
    countStarted.value = true
  }, 1200)
  return new Promise((resolve) => {
    resolveRun = resolve
    // 余韵结束（3.5s）：整层淡出，反馈弹窗由调用方在 await 后关闭
    closeTimer = window.setTimeout(() => finish(false), 3500)
  })
}

/** 收场：隐藏庆祝层、清理纸屑与计时器、归还焦点并 resolve */
function finish(withView: boolean) {
  const opts = celebration.value
  clearTimers()
  confetti.reset()
  visible.value = false
  celebration.value = null
  previousFocus?.focus?.()
  previousFocus = null
  const resolve = resolveRun
  resolveRun = null
  resolve?.(true)
  if (withView) opts?.onView?.()
}

export function useGloryCelebration() {
  return {
    visible,
    celebration,
    countStarted,
    celebrationEnabled,
    setCelebrationEnabled,
    celebrate,
    /** 点击「查看成果」：先收场，再由 onView 跳转 */
    viewResult: () => finish(true),
    /** 提前关闭（如 Esc） */
    dismiss: () => finish(false),
  }
}
