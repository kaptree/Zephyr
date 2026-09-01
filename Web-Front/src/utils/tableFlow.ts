/**
 * 表格"呼吸式数据流"滚动视差（美化工程 · 表格模块化）。
 *
 * 用法：滚动容器挂 `.table-flow` 类并监听 `@scroll="onTableFlowScroll"`，
 * 滚动位移写入 CSS 变量 --flow-scroll，驱动 tbody 以极小系数滞后滚动，
 * 营造内容行"轻微错位"的沉浸感；在滚动两端自动归零，避免露底空隙。
 */
export function onTableFlowScroll(e: Event): void {
  const el = e.target as HTMLElement
  const maxScroll = el.scrollHeight - el.clientHeight
  if (maxScroll <= 0) return
  const lag = Math.max(0, Math.min(el.scrollTop, maxScroll - el.scrollTop))
  el.style.setProperty('--flow-scroll', String(lag))
}
