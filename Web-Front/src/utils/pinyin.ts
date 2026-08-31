import { match } from 'pinyin-pro';

/**
 * 需求36：通用拼音搜索匹配
 * 支持全拼（hanyu）、拼音首字母（hy）、原文包含、中英混合文本，
 * 全程无视大小写。
 *
 * @param keyword 用户输入的搜索关键词
 * @param texts 待匹配的候选文本（如姓名、部门名、标题等）
 * @returns 任一文本命中即返回 true；keyword 为空视为全部命中
 */
export function matchPinyin(keyword: string, ...texts: string[]): boolean {
  const kw = keyword.trim();
  if (!kw) return true;
  for (const t of texts) {
    if (!t) continue;
    // 原文直接命中（含英文/编号等非中文场景）
    if (t.toLowerCase().includes(kw.toLowerCase())) return true;
    // 拼音匹配：match() 内部对小写化处理，支持全拼 / 首字母 / 混合输入
    if (match(t, kw)) return true;
  }
  return false;
}
