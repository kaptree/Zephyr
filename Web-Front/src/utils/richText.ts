/**
 * 富文本内容渲染工具
 * - 新内容为富文本 HTML（含标签）时原样渲染
 * - 旧内容为纯文本时转义并转换换行为 <br>，保证多行显示
 */
export function renderNoteContent(content?: string | null): string {
  const text = content || ''
  if (!text) return ''
  // 含 HTML 标签视为富文本
  if (/<\/?[a-z][\s\S]*>/i.test(text)) return text
  const escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
  return escaped.replace(/\n/g, '<br>')
}
