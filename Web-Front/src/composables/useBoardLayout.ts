import type { Note, NotePositionInput } from '@/types';

/**
 * 工作台格子画布布局：每行 N 列（用户可调，范围见 BOARD_MIN/MAX_COLS），
 * 每个便签占一个格子，卡片宽度随容器自适应。
 * pos_x / pos_y 的语义为「格子索引」：pos_x = 列号（0..cols-1），pos_y = 行号（0..N），
 * 渲染时换算为 CSS Grid 的 grid-column / grid-row。
 */
/** 默认列数 */
export const BOARD_COLS = 3;
/** 列数可调范围（含边界） */
export const BOARD_MIN_COLS = 1;
export const BOARD_MAX_COLS = 6;
/** 列数偏好的 localStorage 键 */
export const BOARD_COLS_STORAGE_KEY = 'workbench_board_cols';
/** 格子间距（px），用于画布 grid gap 样式 */
export const BOARD_GAP = 20;

/** 把任意输入收敛到合法列数范围（非法值回退默认 3 列） */
export function clampBoardCols(n: number): number {
  if (!Number.isFinite(n)) return BOARD_COLS;
  return Math.min(BOARD_MAX_COLS, Math.max(BOARD_MIN_COLS, Math.round(n)));
}

/**
 * 从 localStorage 读取列数偏好；无记录 / 环境不支持 / 值非法时返回 null。
 * 与 loadBoardCols 的区别：能区分「用户显式保存过 3 列」与「从未保存过（默认 3 列）」。
 */
export function readStoredBoardCols(): number | null {
  try {
    const raw = localStorage.getItem(BOARD_COLS_STORAGE_KEY);
    if (raw === null) return null;
    const n = Number(raw);
    return Number.isFinite(n) ? clampBoardCols(n) : null;
  } catch {
    return null;
  }
}

/** 从 localStorage 读取列数偏好（无记录 / 环境不支持时返回默认值） */
export function loadBoardCols(): number {
  return readStoredBoardCols() ?? BOARD_COLS;
}

/** 持久化列数偏好（环境不支持时静默跳过） */
export function saveBoardCols(cols: number): void {
  try {
    localStorage.setItem(BOARD_COLS_STORAGE_KEY, String(clampBoardCols(cols)));
  } catch {
    /* 忽略隐私模式等存储异常 */
  }
}

/**
 * 旧版自由画布像素坐标判定阈值。合法格子索引最大为 BOARD_MAX_COLS - 1（6 列布局的列号 0..5），
 * 旧版像素坐标（列宽约 360px 起）远大于该值。注意不能用默认列数 BOARD_COLS - 1 作阈值，
 * 否则 4~6 列布局的合法位置（pos_x 3..5）会被误判为遗留数据，导致每次加载整板清位重排。
 */
const LEGACY_MAX_COL = BOARD_MAX_COLS - 1;
const LEGACY_MAX_ROW = 100;
/** 空闲格扫描防御上限（行数） */
const MAX_SCAN_ROWS = 1000;

/**
 * 检测旧版自由画布遗留的像素坐标（pos 为像素值时数值远大于格子索引）。
 * 命中后整板清空位置，按置顶 + 列表顺序重新分配格子索引完成迁移。
 */
export function hasLegacyPixelPositions(notes: Note[]): boolean {
  return notes.some(
    (n) =>
      (n.pos_x !== null && n.pos_x > LEGACY_MAX_COL) ||
      (n.pos_y !== null && n.pos_y > LEGACY_MAX_ROW)
  );
}

/** 旧像素数据清洗：存在遗留坐标时整板视为无位置，待重新分配格子索引 */
export function stripLegacyPositions(notes: Note[]): Note[] {
  if (!hasLegacyPixelPositions(notes)) return notes;
  return notes.map((n) =>
    n.pos_x === null && n.pos_y === null ? n : { ...n, pos_x: null, pos_y: null }
  );
}

/** 收集已占用的格子（"列,行" 键集合），skipId 自身不参与占用判断 */
function collectOccupied(notes: Note[], skipId?: string): Set<string> {
  const set = new Set<string>();
  for (const n of notes) {
    if (n.id === skipId || n.pos_x === null || n.pos_y === null) continue;
    set.add(`${n.pos_x},${n.pos_y}`);
  }
  return set;
}

/** 置顶任务优先（多个置顶按置顶时间倒序），其余保持传入顺序 */
function sortForPlacement(notes: Note[]): Note[] {
  return [...notes].sort((a, b) => {
    if (a.is_pinned !== b.is_pinned) return a.is_pinned ? -1 : 1;
    if (a.is_pinned && b.is_pinned) return (b.pinned_at || '').localeCompare(a.pinned_at || '');
    return 0;
  });
}

/** 行优先扫描第一个空闲格，超过扫描上限返回兜底位 */
function nextFreeSlot(occupied: Set<string>, cols: number): { x: number; y: number } {
  for (let row = 0; row <= MAX_SCAN_ROWS; row++) {
    for (let col = 0; col < cols; col++) {
      if (!occupied.has(`${col},${row}`)) return { x: col, y: row };
    }
  }
  return { x: 0, y: MAX_SCAN_ROWS + 1 };
}

/**
 * 为缺少格子记录的便签按行优先分配空闲格：置顶任务优先占前排。
 * 只返回需要回存的便签（已有格子记录的不动）。
 */
export function assignMissingPositions(notes: Note[], cols: number): NotePositionInput[] {
  const occupied = collectOccupied(notes);
  const missing = sortForPlacement(notes.filter((n) => n.pos_x === null || n.pos_y === null));
  return missing.map((note) => {
    const slot = nextFreeSlot(occupied, cols);
    occupied.add(`${slot.x},${slot.y}`);
    return { id: note.id, pos_x: slot.x, pos_y: slot.y };
  });
}

/**
 * 找到当前视图的第一个空闲格（行优先扫描，skipId 自身不参与占用判断）。
 * 用于置顶联动：置顶便签自动移到最前面的空闲位置。
 */
export function findFirstFreeSlot(
  notes: Note[],
  cols: number,
  skipId?: string
): { x: number; y: number } {
  return nextFreeSlot(collectOccupied(notes, skipId), cols);
}

/**
 * 列数变更后整板重排：格网形状随列数改变，全部便签清位后
 * 按行优先重新铺满（置顶任务优先占前排），返回全量需要回存的便签。
 */
export function relayoutAll(notes: Note[], cols: number): NotePositionInput[] {
  const cleared = notes.map((n) => ({ ...n, pos_x: null, pos_y: null }));
  return assignMissingPositions(cleared, cols);
}

/**
 * 从便签位置推断曾用的布局列数（仅用于列数偏好丢失的场景）：
 * pos_x 最大值 + 1 即曾用列数的下界。只有超出默认列数才可推断
 * （更窄布局与「3 列中恰好没占满」无法区分，保持默认），非法超大值收敛到列数上限。
 */
export function inferBoardColsFromNotes(notes: Note[]): number | null {
  let maxCol = -1;
  for (const n of notes) {
    if (n.pos_x !== null && n.pos_x > maxCol) maxCol = n.pos_x;
  }
  if (maxCol + 1 <= BOARD_COLS) return null;
  return clampBoardCols(maxCol + 1);
}

/**
 * 超界自愈：存在 pos_x >= cols 的便签时（如列数偏好丢失回落后，
 * 后端位置仍按更宽布局存储），整板按当前列数重排补齐空洞，
 * 返回全量需要回存的位置；无超界时返回空数组（不改动现有布局）。
 */
export function fixOverflowPositions(notes: Note[], cols: number): NotePositionInput[] {
  const hasOverflow = notes.some((n) => n.pos_x !== null && n.pos_x >= cols);
  if (!hasOverflow) return [];
  return relayoutAll(notes, cols);
}
