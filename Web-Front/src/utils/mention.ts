/**
 * 群聊 @ 提及工具（参考微信交互）
 * - 输入框：解析光标前的活动 @ 记号，驱动成员选择浮层
 * - 输入框镜像层：把「@名字」分段，配合 CSS 做蓝底高亮
 * - 发送：从消息内容反推被 @ 的成员名（删除 @文本 自动失效）
 * - 气泡：把「@名字」包上 mention-tag 高亮标记
 */
import { matchPinyin } from './pinyin';
import { renderNoteContent } from './richText';
import type { GroupMemberItem } from '@/types';

/** @ 提及的名字长度上限 */
const MAX_MENTION_NAME_LEN = 32;
/** @ 过滤词长度上限（超长视为普通文本，不再弹出选择器） */
const MAX_MENTION_KEYWORD_LEN = 20;

/** 群成员显示名：昵称 > 用户名 > 用户 ID 前 6 位 */
export function memberDisplayName(m: GroupMemberItem): string {
  return m.user?.name || m.user?.username || m.user_id.slice(0, 6);
}

/** 选中成员后插入输入框的文本：@名字 + 空格 */
export function buildMentionInsert(name: string): string {
  return `@${name} `;
}

/**
 * 解析输入框中光标前的活动 @ 记号（微信交互：输入 @ 弹出成员选择浮层）：
 * - @ 须位于文本开头或空白字符之后（避免命中邮箱 a@b.com）
 * - @ 与光标之间不能出现空白或另一个 @
 * - 过滤词过长视为普通文本
 *
 * @returns @ 的下标与过滤词；无活动记号返回 null
 */
export function findActiveMentionToken(
  content: string,
  caret: number
): { at: number; keyword: string } | null {
  if (caret <= 0) return null;
  const before = content.slice(0, caret);
  const at = before.lastIndexOf('@');
  if (at < 0) return null;
  // @ 必须位于文本开头或空白之后
  if (at > 0 && !/\s/.test(before[at - 1])) return null;
  const keyword = before.slice(at + 1);
  // @ 与光标之间出现空白/另一个 @，或过滤词过长：不视为提及
  if (/[\s@]/.test(keyword)) return null;
  if (keyword.length > MAX_MENTION_KEYWORD_LEN) return null;
  return { at, keyword };
}

/** 文本分段：mentioned 段为有效的「@名字」 */
export interface MentionSegment {
  text: string;
  mentioned: boolean;
}

/**
 * 将消息文本按「@名字」分段（名字须在候选名中才视为有效提及）：
 * - @ 须位于文本开头或空白后，名字不含空白
 * - 用于输入框镜像高亮层与消息气泡高亮渲染
 */
export function splitMentionSegments(content: string, candidateNames: string[]): MentionSegment[] {
  if (!content) return [];
  const nameSet = new Set(candidateNames);
  if (nameSet.size === 0) return [{ text: content, mentioned: false }];
  const segments: MentionSegment[] = [];
  const re = new RegExp(`(^|\\s)@([^\\s@]{1,${MAX_MENTION_NAME_LEN}})`, 'g');
  let last = 0;
  let m: RegExpExecArray | null;
  while ((m = re.exec(content))) {
    const name = m[2];
    if (!nameSet.has(name)) continue;
    const start = m.index + m[1].length;
    if (start > last) segments.push({ text: content.slice(last, start), mentioned: false });
    segments.push({ text: `@${name}`, mentioned: true });
    last = start + name.length + 1;
  }
  if (last < content.length) segments.push({ text: content.slice(last), mentioned: false });
  return segments;
}

/** 从消息文本中提取有效的 @ 名字列表（与候选名取交集，去重保序） */
export function extractMentionedNames(content: string, candidateNames: string[]): string[] {
  const names: string[] = [];
  const seen = new Set<string>();
  for (const s of splitMentionSegments(content, candidateNames)) {
    if (!s.mentioned) continue;
    const name = s.text.slice(1);
    if (!seen.has(name)) {
      seen.add(name);
      names.push(name);
    }
  }
  return names;
}

/**
 * 消息气泡 HTML 渲染：与 renderNoteContent 相同的转义 + 换行处理，
 * 并把「@名字」包上 mention-tag 高亮标记（由 CSS 上色）
 */
export function renderMessageContent(content: string, candidateNames: string[]): string {
  if (!content) return '';
  // 含 HTML 标签视为富文本，原样渲染（与 renderNoteContent 保持一致）
  if (/<\/?[a-z][\s\S]*>/i.test(content)) return content;
  return splitMentionSegments(content, candidateNames)
    .map((seg) => {
      const html = renderNoteContent(seg.text);
      return seg.mentioned ? `<span class="mention-tag">${html}</span>` : html;
    })
    .join('');
}

/** 群成员按关键字过滤：昵称 / 用户名 / 部门名 / 用户 ID 兜底，支持拼音全拼与首字母 */
export function filterMentionMembers(members: GroupMemberItem[], keyword: string): GroupMemberItem[] {
  const kw = keyword.trim();
  if (!kw) return members;
  return members.filter((m) =>
    matchPinyin(
      kw,
      m.user?.name || '',
      m.user?.username || '',
      m.user?.department?.name || '',
      m.user_id
    )
  );
}
