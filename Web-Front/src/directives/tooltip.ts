import type { Directive, DirectiveBinding } from 'vue'

/**
 * v-tooltip 全局工具提示指令（美化工程 · 轻量级通知美学）
 *
 * 用法：
 *   v-tooltip="'新建任务'"              — 提示出现在元素上方（空间不足自动放到下方）
 *   v-tooltip="{ content: '文案', below: true }" — 强制显示在元素下方
 *
 * 交互：悬停时「上浮 + 淡入」200ms 出现，带小三角指示箭头，12px 字体；
 *       离开 / 按下 / 页面滚动时立即隐藏。
 */

interface TooltipState {
  text: string
  below: boolean
}

let tipEl: HTMLDivElement | null = null

function getTip(): HTMLDivElement {
  if (!tipEl) {
    tipEl = document.createElement('div')
    tipEl.className = 'v-tooltip'
    document.body.appendChild(tipEl)
  }
  return tipEl
}

function hideTip() {
  tipEl?.classList.remove('show')
}

function showTip(el: HTMLElement, state: TooltipState) {
  if (!state.text) return
  const tip = getTip()
  tip.textContent = state.text
  tip.classList.remove('show', 'v-tooltip-below')
  tip.style.visibility = 'hidden'
  const tipRect = tip.getBoundingClientRect()
  const rect = el.getBoundingClientRect()
  const gap = 8
  // 默认在上方；顶部空间不足则放下方，且尊重 below 显式指定
  let below = state.below
  if (!below && rect.top - tipRect.height - gap < 8) below = true
  if (!state.below && rect.bottom + tipRect.height + gap > window.innerHeight - 8) below = false
  tip.classList.toggle('v-tooltip-below', below)
  const left = Math.min(
    Math.max(8, rect.left + rect.width / 2 - tipRect.width / 2),
    window.innerWidth - tipRect.width - 8,
  )
  const top = below ? rect.bottom + gap : rect.top - tipRect.height - gap
  tip.style.left = `${Math.round(left)}px`
  tip.style.top = `${Math.round(top)}px`
  tip.style.visibility = ''
  requestAnimationFrame(() => tip.classList.add('show'))
}

function readValue(binding: DirectiveBinding, state: TooltipState) {
  if (typeof binding.value === 'string' || typeof binding.value === 'number') {
    state.text = String(binding.value)
    state.below = false
  } else if (binding.value) {
    state.text = String(binding.value.content ?? '')
    state.below = !!binding.value.below
  } else {
    state.text = ''
  }
}

interface Handlers {
  enter: () => void
  leave: () => void
  down: () => void
  scroll: () => void
}

const states = new WeakMap<HTMLElement, TooltipState>()
const handlersMap = new WeakMap<HTMLElement, Handlers>()

export const vTooltip: Directive<HTMLElement> = {
  mounted(el, binding) {
    const state: TooltipState = { text: '', below: false }
    readValue(binding, state)
    states.set(el, state)
    const handlers: Handlers = {
      enter: () => {
        const s = states.get(el)
        if (s) showTip(el, s)
      },
      leave: hideTip,
      down: hideTip,
      scroll: hideTip,
    }
    el.addEventListener('mouseenter', handlers.enter)
    el.addEventListener('mouseleave', handlers.leave)
    el.addEventListener('mousedown', handlers.down)
    window.addEventListener('scroll', handlers.scroll, true)
    handlersMap.set(el, handlers)
  },
  updated(el, binding) {
    const state = states.get(el)
    if (state) readValue(binding, state)
  },
  unmounted(el) {
    const handlers = handlersMap.get(el)
    if (handlers) {
      el.removeEventListener('mouseenter', handlers.enter)
      el.removeEventListener('mouseleave', handlers.leave)
      el.removeEventListener('mousedown', handlers.down)
      window.removeEventListener('scroll', handlers.scroll, true)
      handlersMap.delete(el)
    }
    states.delete(el)
    hideTip()
  },
}
