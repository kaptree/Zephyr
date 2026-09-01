/* 全局按钮点击涟漪（美化工程 · 微交互动画语言）
   事件委托：主要按钮（btn-primary / btn-secondary 及内联蓝色主按钮）点击时，
   从点击位置向外扩散 400ms 圆形涟漪，颜色为按钮主色的半透明版本，营造"触达"感 */

const RIPPLE_SELECTOR =
  'button.btn-primary, a.btn-primary, button.btn-secondary, a.btn-secondary, button.bg-blue-500, button.bg-blue-600, button[class*="bg-[#3B82F6]"]'

let listening = false

export function setupRipple(): void {
  if (listening || typeof document === 'undefined') return
  listening = true

  document.addEventListener('click', (e: MouseEvent) => {
    const target = e.target as Element | null
    const host = target?.closest?.(RIPPLE_SELECTOR) as HTMLElement | null
    if (!host || host.classList.contains('btn-disabled') || host.hasAttribute('disabled')) return

    const rect = host.getBoundingClientRect()
    const size = Math.max(rect.width, rect.height) * 2
    const ink = document.createElement('span')
    ink.className = 'ripple-ink'
    ink.style.width = ink.style.height = `${size}px`
    ink.style.left = `${e.clientX - rect.left - size / 2}px`
    ink.style.top = `${e.clientY - rect.top - size / 2}px`

    host.appendChild(ink)
    window.setTimeout(() => ink.remove(), 450)
  })
}
