import { marked } from 'marked'
import TurndownService from 'turndown'

marked.setOptions({
  gfm: true,
  breaks: true, // 单个换行渲染为 <br>，兼容旧纯文本换行习惯
})

const turndown = new TurndownService({
  headingStyle: 'atx',
  codeBlockStyle: 'fenced',
  bulletListMarker: '-',
  emDelimiter: '*',
})

/** Markdown → HTML。任务/模板内容保存时转为 HTML，与既有富文本展示（renderNoteContent）兼容 */
export function markdownToHtml(md: string): string {
  return marked.parse(md || '') as string
}

/**
 * HTML → Markdown。打开旧富文本内容时转回 Markdown 源码编辑；
 * 纯文本原样返回，避免 turndown 转义 markdown 符号（如 #、-）
 */
export function htmlToMarkdown(html: string): string {
  const text = html || ''
  if (!text) return ''
  // 与 renderNoteContent 一致：不含 HTML 标签视为纯文本
  if (!/<\/?[a-z][\s\S]*>/i.test(text)) return text
  return turndown.turndown(text)
}
