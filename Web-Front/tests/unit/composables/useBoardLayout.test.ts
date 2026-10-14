import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import {
  BOARD_COLS,
  BOARD_MIN_COLS,
  BOARD_MAX_COLS,
  BOARD_COLS_STORAGE_KEY,
  clampBoardCols,
  loadBoardCols,
  saveBoardCols,
  readStoredBoardCols,
  hasLegacyPixelPositions,
  stripLegacyPositions,
  assignMissingPositions,
  findFirstFreeSlot,
  relayoutAll,
  inferBoardColsFromNotes,
  fixOverflowPositions,
  resolveDuplicatePositions,
  compactBoardPositions,
} from '@/composables/useBoardLayout';
import { createMockNote } from '../../mocks/data';
import type { Note } from '@/types';

/** 构造带格子位置/置顶信息的测试卡片（pos_x=列 0..2，pos_y=行 0..N） */
function noteAt(
  id: string,
  x: number | null,
  y: number | null,
  overrides: Partial<Note> = {}
): Note {
  return createMockNote({ id, pos_x: x, pos_y: y, ...overrides });
}

describe('BOARD_COLS', () => {
  it('默认 3 列格子', () => {
    expect(BOARD_COLS).toBe(3);
  });
});

describe('hasLegacyPixelPositions', () => {
  it('空位置与合法格子索引（最大 6 列布局的 0..5）均不算遗留像素数据', () => {
    expect(hasLegacyPixelPositions([noteAt('a', null, null)])).toBe(false);
    expect(hasLegacyPixelPositions([noteAt('a', 0, 0), noteAt('b', 2, 100)])).toBe(false);
    // 4~6 列布局的合法列号（3..5）不能误判为遗留像素数据，否则每次加载整板清位重排
    expect(hasLegacyPixelPositions([noteAt('a', 3, 0), noteAt('b', 5, 12)])).toBe(false);
  });

  it('列超出 0..5 或行超过 100 判定为遗留像素数据', () => {
    expect(hasLegacyPixelPositions([noteAt('a', 6, 0)])).toBe(true);
    expect(hasLegacyPixelPositions([noteAt('a', 0, 101)])).toBe(true);
    expect(hasLegacyPixelPositions([noteAt('a', 280, 200)])).toBe(true);
  });
});

describe('stripLegacyPositions', () => {
  it('无遗留数据时原样返回（同一引用）', () => {
    const notes = [noteAt('a', 0, 0), noteAt('b', null, null)];
    expect(stripLegacyPositions(notes)).toBe(notes);
  });

  it('存在遗留像素数据时整板清空位置，等待重新分配格子', () => {
    const notes = [
      noteAt('legacy', 280, 200),
      noteAt('clean', 1, 0),
      noteAt('empty', null, null),
    ];
    const stripped = stripLegacyPositions(notes);
    expect(stripped.map((n) => [n.id, n.pos_x, n.pos_y])).toEqual([
      ['legacy', null, null],
      ['clean', null, null],
      ['empty', null, null],
    ]);
  });
});

describe('assignMissingPositions', () => {
  it('全部缺位置时按行优先分配，置顶任务优先占前排（置顶时间倒序）', () => {
    const notes = [
      noteAt('c', null, null),
      noteAt('a', null, null, { is_pinned: true, pinned_at: '2024-01-02T00:00:00Z' }),
      noteAt('b', null, null, { is_pinned: true, pinned_at: '2024-01-01T00:00:00Z' }),
    ];
    const updates = assignMissingPositions(notes, 3);
    // 多个置顶按置顶时间倒序：a 先于 b
    expect(updates).toEqual([
      { id: 'a', pos_x: 0, pos_y: 0 },
      { id: 'b', pos_x: 1, pos_y: 0 },
      { id: 'c', pos_x: 2, pos_y: 0 },
    ]);
  });

  it('已有格子位置的卡片不参与分配，缺位置卡片补入空闲格', () => {
    const notes = [noteAt('n1', 1, 0), noteAt('n2', null, null)];
    const updates = assignMissingPositions(notes, 3);
    // 行优先扫描：(0,0) 空闲 → 分配给 n2；n1 原格子不动
    expect(updates).toEqual([{ id: 'n2', pos_x: 0, pos_y: 0 }]);
  });

  it('全部卡片已有格子位置时返回空数组', () => {
    const notes = [noteAt('n1', 0, 0), noteAt('n2', 1, 1)];
    expect(assignMissingPositions(notes, 3)).toEqual([]);
  });

  it('同一行占满后自动换行到下一行', () => {
    const notes = [
      noteAt('n1', 0, 0),
      noteAt('n2', 1, 0),
      noteAt('n3', null, null),
    ];
    const updates = assignMissingPositions(notes, 2);
    expect(updates).toEqual([{ id: 'n3', pos_x: 0, pos_y: 1 }]);
  });

  it('跳过被同行空洞阻隔的格子：中间空洞优先补位', () => {
    const notes = [
      noteAt('n1', 0, 0),
      noteAt('n3', 2, 0),
      noteAt('n2', null, null),
    ];
    const updates = assignMissingPositions(notes, 3);
    expect(updates).toEqual([{ id: 'n2', pos_x: 1, pos_y: 0 }]);
  });
});

describe('findFirstFreeSlot', () => {
  it('无卡片时返回原点格子', () => {
    expect(findFirstFreeSlot([], 3)).toEqual({ x: 0, y: 0 });
  });

  it('跳过自身占位（skipId）', () => {
    const notes = [noteAt('self', 0, 0)];
    expect(findFirstFreeSlot(notes, 3, 'self')).toEqual({ x: 0, y: 0 });
  });

  it('首个格子被占时顺延到同一行的下一格', () => {
    const notes = [noteAt('n1', 0, 0), noteAt('n2', 1, 0)];
    expect(findFirstFreeSlot(notes, 3)).toEqual({ x: 2, y: 0 });
  });

  it('整行占满时换行到下一行首格', () => {
    const notes = [noteAt('n1', 0, 0), noteAt('n2', 1, 0), noteAt('n3', 2, 0)];
    expect(findFirstFreeSlot(notes, 3)).toEqual({ x: 0, y: 1 });
  });
});

describe('clampBoardCols', () => {
  it('范围内的值原样保留（四舍五入）', () => {
    expect(clampBoardCols(1)).toBe(1);
    expect(clampBoardCols(3)).toBe(3);
    expect(clampBoardCols(6)).toBe(6);
    expect(clampBoardCols(4.4)).toBe(4);
  });

  it('超出范围时收敛到边界', () => {
    expect(clampBoardCols(0)).toBe(BOARD_MIN_COLS);
    expect(clampBoardCols(-5)).toBe(BOARD_MIN_COLS);
    expect(clampBoardCols(7)).toBe(BOARD_MAX_COLS);
    expect(clampBoardCols(100)).toBe(BOARD_MAX_COLS);
  });

  it('非法输入回退默认 3 列', () => {
    expect(clampBoardCols(NaN)).toBe(BOARD_COLS);
    expect(clampBoardCols(Infinity)).toBe(BOARD_COLS);
    expect(clampBoardCols(-Infinity)).toBe(BOARD_COLS);
  });

  it('列数范围常量为 1 ~ 6', () => {
    expect(BOARD_MIN_COLS).toBe(1);
    expect(BOARD_MAX_COLS).toBe(6);
  });
});

describe('loadBoardCols / saveBoardCols', () => {
  beforeEach(() => {
    localStorage.removeItem(BOARD_COLS_STORAGE_KEY);
  });

  afterEach(() => {
    localStorage.removeItem(BOARD_COLS_STORAGE_KEY);
  });

  it('无存储记录时返回默认 3 列', () => {
    expect(loadBoardCols()).toBe(BOARD_COLS);
  });

  it('saveBoardCols 持久化偏好，loadBoardCols 读取还原', () => {
    saveBoardCols(5);
    expect(localStorage.getItem(BOARD_COLS_STORAGE_KEY)).toBe('5');
    expect(loadBoardCols()).toBe(5);
  });

  it('存储值越界或非法时收敛到合法范围', () => {
    localStorage.setItem(BOARD_COLS_STORAGE_KEY, '99');
    expect(loadBoardCols()).toBe(BOARD_MAX_COLS);
    localStorage.setItem(BOARD_COLS_STORAGE_KEY, 'abc');
    expect(loadBoardCols()).toBe(BOARD_COLS);
  });
});

describe('relayoutAll', () => {
  it('整板清位后按行优先重新铺满（行优先、紧凑无空洞）', () => {
    // 4 列布局下的存量位置：a(0,0) b(1,0) c(3,0) d(3,1)
    const notes = [
      noteAt('a', 0, 0),
      noteAt('b', 1, 0),
      noteAt('c', 3, 0),
      noteAt('d', 3, 1),
    ];
    // 收缩到 3 列：忽略旧位置，按传入顺序行优先铺满
    expect(relayoutAll(notes, 3)).toEqual([
      { id: 'a', pos_x: 0, pos_y: 0 },
      { id: 'b', pos_x: 1, pos_y: 0 },
      { id: 'c', pos_x: 2, pos_y: 0 },
      { id: 'd', pos_x: 0, pos_y: 1 },
    ]);
  });

  it('扩张列数时同样重排，卡片铺满新网格右侧', () => {
    // 1 列存量：a(0,0) b(0,1) c(0,2)
    const notes = [noteAt('a', 0, 0), noteAt('b', 0, 1), noteAt('c', 0, 2)];
    // 扩张到 3 列：三卡横向铺满首行
    expect(relayoutAll(notes, 3)).toEqual([
      { id: 'a', pos_x: 0, pos_y: 0 },
      { id: 'b', pos_x: 1, pos_y: 0 },
      { id: 'c', pos_x: 2, pos_y: 0 },
    ]);
  });

  it('重排时置顶任务优先占前排（置顶时间倒序）', () => {
    const notes = [
      noteAt('plain', 0, 0),
      noteAt('pinned', 0, 1, { is_pinned: true, pinned_at: '2024-01-01T00:00:00Z' }),
    ];
    expect(relayoutAll(notes, 3)).toEqual([
      { id: 'pinned', pos_x: 0, pos_y: 0 },
      { id: 'plain', pos_x: 1, pos_y: 0 },
    ]);
  });
});

describe('readStoredBoardCols', () => {
  beforeEach(() => {
    localStorage.removeItem(BOARD_COLS_STORAGE_KEY);
  });

  afterEach(() => {
    localStorage.removeItem(BOARD_COLS_STORAGE_KEY);
  });

  it('无存储记录时返回 null（区别于「显式保存过默认 3 列」）', () => {
    expect(readStoredBoardCols()).toBeNull();
    saveBoardCols(BOARD_COLS);
    expect(readStoredBoardCols()).toBe(BOARD_COLS);
  });

  it('已保存偏好时返回收敛后的列数', () => {
    saveBoardCols(4);
    expect(readStoredBoardCols()).toBe(4);
    localStorage.setItem(BOARD_COLS_STORAGE_KEY, '99');
    expect(readStoredBoardCols()).toBe(BOARD_MAX_COLS);
  });

  it('存储值非法（非数字）时返回 null', () => {
    localStorage.setItem(BOARD_COLS_STORAGE_KEY, 'abc');
    expect(readStoredBoardCols()).toBeNull();
  });
});

describe('inferBoardColsFromNotes', () => {
  it('无便签或位置全空时返回 null', () => {
    expect(inferBoardColsFromNotes([])).toBeNull();
    expect(inferBoardColsFromNotes([noteAt('a', null, null)])).toBeNull();
  });

  it('位置均在默认 3 列内时返回 null（无法推断更宽布局，保持默认）', () => {
    expect(inferBoardColsFromNotes([noteAt('a', 0, 0), noteAt('b', 2, 5)])).toBeNull();
  });

  it('存在超出默认列数的位置时按最大列号 + 1 推断曾用列数', () => {
    expect(inferBoardColsFromNotes([noteAt('a', 0, 0), noteAt('b', 3, 1)])).toBe(4);
    expect(inferBoardColsFromNotes([noteAt('a', 5, 2)])).toBe(BOARD_MAX_COLS);
  });

  it('非法超大的列号收敛到列数上限', () => {
    expect(inferBoardColsFromNotes([noteAt('a', 9, 0)])).toBe(BOARD_MAX_COLS);
  });
});

describe('fixOverflowPositions', () => {
  it('无超界便签时返回空数组（不改动现有布局）', () => {
    expect(fixOverflowPositions([noteAt('a', 0, 0), noteAt('b', 2, 1)], 3)).toEqual([]);
    // 位置在 4 列网格内不算超界
    expect(fixOverflowPositions([noteAt('a', 3, 0)], 4)).toEqual([]);
  });

  it('存在超界便签时整板按当前列数重排：结果无超界、无重叠、数量一致', () => {
    // 4 列布局存量：a(0,0) b(3,0) c(3,1) d(0,1)，回落 3 列后 b/c 超界
    const notes = [
      noteAt('a', 0, 0),
      noteAt('b', 3, 0),
      noteAt('c', 3, 1),
      noteAt('d', 0, 1),
    ];
    const updates = fixOverflowPositions(notes, 3);
    expect(updates).toHaveLength(4);
    const cells = new Set(updates.map((u) => `${u.pos_x},${u.pos_y}`));
    expect(cells.size).toBe(4);
    for (const u of updates) expect(u.pos_x).toBeLessThan(3);
  });

  it('超界自愈重排时置顶任务优先占前排', () => {
    const notes = [
      noteAt('plain', 3, 0),
      noteAt('pinned', 3, 1, { is_pinned: true, pinned_at: '2024-01-01T00:00:00Z' }),
    ];
    expect(fixOverflowPositions(notes, 3)).toEqual([
      { id: 'pinned', pos_x: 0, pos_y: 0 },
      { id: 'plain', pos_x: 1, pos_y: 0 },
    ]);
  });
});

describe('resolveDuplicatePositions', () => {
  it('无重复坐标时返回空数组（不改动现有布局）', () => {
    expect(resolveDuplicatePositions([noteAt('a', 0, 0), noteAt('b', 2, 1)], 3)).toEqual([]);
  });

  it('两张便签占用同一格子时，把后一张移到空闲格', () => {
    const notes = [noteAt('a', 2, 2), noteAt('b', 2, 2)];
    // 保留 a(2,2)，b 被移到行优先扫描到的最近空闲格 (0,0)
    expect(resolveDuplicatePositions(notes, 3)).toEqual([{ id: 'b', pos_x: 0, pos_y: 0 }]);
  });

  it('重复格在行中间时，优先补入靠前的空闲格（非末尾）', () => {
    // a(2,2) 与 b(2,2) 重复；occupied 含 (2,2)/(0,0)，行优先扫描先到 (1,0)
    const notes = [noteAt('a', 2, 2), noteAt('occupy', 0, 0), noteAt('b', 2, 2)];
    const updates = resolveDuplicatePositions(notes, 3);
    expect(updates).toEqual([{ id: 'b', pos_x: 1, pos_y: 0 }]);
  });

  it('保留置顶任务占位，重复的普通任务被移走', () => {
    const notes = [
      noteAt('plain', 1, 1),
      noteAt('pinned', 1, 1, { is_pinned: true, pinned_at: '2024-01-01T00:00:00Z' }),
    ];
    // 置顶优先保留 (1,1)，普通任务移到空闲格 (0,0)
    expect(resolveDuplicatePositions(notes, 3)).toEqual([{ id: 'plain', pos_x: 0, pos_y: 0 }]);
  });

  it('多个重复位置全部去重，结果无重叠且数量一致', () => {
    const notes = [
      noteAt('a', 0, 0),
      noteAt('b', 0, 0),
      noteAt('c', 1, 0),
      noteAt('d', 1, 0),
    ];
    const updates = resolveDuplicatePositions(notes, 3);
    expect(updates).toHaveLength(2);
    // a 保留 (0,0)，c 保留 (1,0)；b、d 分别移到空闲格
    const byId = new Map(updates.map((u) => [u.id, `${u.pos_x},${u.pos_y}`]));
    const cells = new Set(['0,0', '1,0', byId.get('b')!, byId.get('d')!]);
    expect(cells.size).toBe(4);
    // b、d 不再落在原重复格
    expect(byId.get('b')).not.toBe('0,0');
    expect(byId.get('d')).not.toBe('1,0');
  });
});

describe('compactBoardPositions', () => {
  /** 以 note.pos_x/pos_y 作为当前渲染格子的取位函数 */
  const posOf = (n: Note) => ({ x: n.pos_x ?? 0, y: n.pos_y ?? 0 });

  it('归档空洞之后的便签各自前移一格，洞前卡片不动', () => {
    // a(0,0) b(1,0) [空洞 2,0] c(0,1) d(1,1)，b 归档后 c/d 前移
    const notes = [
      noteAt('a', 0, 0),
      noteAt('c', 0, 1),
      noteAt('d', 1, 1),
    ];
    expect(compactBoardPositions(notes, 3, posOf)).toEqual([
      { id: 'c', pos_x: 1, pos_y: 0 },
      { id: 'd', pos_x: 2, pos_y: 0 },
    ]);
  });

  it('位置已紧凑无空洞时返回空数组（无需补位）', () => {
    const notes = [noteAt('a', 0, 0), noteAt('b', 1, 0), noteAt('c', 2, 0)];
    expect(compactBoardPositions(notes, 3, posOf)).toEqual([]);
  });

  it('同一行多个空洞一次补齐，跨行卡片前移到行尾空格', () => {
    // 3 列：a(0,0) [空] [空] c(2,1)，b/c 归档后 c 移到 (1,0)
    const notes = [noteAt('a', 0, 0), noteAt('c', 2, 1)];
    expect(compactBoardPositions(notes, 3, posOf)).toEqual([{ id: 'c', pos_x: 1, pos_y: 0 }]);
  });

  it('保留用户拖拽交换过的相对顺序（按当前位置排序，而非列表顺序）', () => {
    // 用户曾把 a 拖到 c 前面：a(0,0) c(1,0)；列表顺序是 c a（后端返回序），
    // 按当前位置排序应为 a c → 两者均无需移动（若按列表序会错误互换）
    const notes = [noteAt('c', 1, 0), noteAt('a', 0, 0)];
    expect(compactBoardPositions(notes, 3, posOf)).toEqual([]);
  });

  it('补位结果行优先连续且无重叠', () => {
    // 3 列存量：a(0,0) b(2,0) c(1,1) d(2,1)，归档 b 后 c/d 前移
    const notes = [noteAt('a', 0, 0), noteAt('c', 1, 1), noteAt('d', 2, 1)];
    const updates = compactBoardPositions(notes, 3, posOf);
    expect(updates).toEqual([
      { id: 'c', pos_x: 1, pos_y: 0 },
      { id: 'd', pos_x: 2, pos_y: 0 },
    ]);
  });
});
