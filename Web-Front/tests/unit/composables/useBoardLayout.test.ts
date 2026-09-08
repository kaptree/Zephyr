import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import {
  BOARD_COLS,
  BOARD_MIN_COLS,
  BOARD_MAX_COLS,
  BOARD_COLS_STORAGE_KEY,
  clampBoardCols,
  loadBoardCols,
  saveBoardCols,
  hasLegacyPixelPositions,
  stripLegacyPositions,
  assignMissingPositions,
  findFirstFreeSlot,
  relayoutAll,
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
  it('空位置与合法格子索引均不算遗留像素数据', () => {
    expect(hasLegacyPixelPositions([noteAt('a', null, null)])).toBe(false);
    expect(hasLegacyPixelPositions([noteAt('a', 0, 0), noteAt('b', 2, 100)])).toBe(false);
  });

  it('列超出 0..2 或行超过 100 判定为遗留像素数据', () => {
    expect(hasLegacyPixelPositions([noteAt('a', 3, 0)])).toBe(true);
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
