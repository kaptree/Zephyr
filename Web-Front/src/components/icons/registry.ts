/**
 * 有温度的图标语言体系 —— 统一图标注册表（美化工程 · 第 7 项）
 *
 * 图标源：@heroicons/vue（Heroicons v2，圆润友好 + 专业清晰的圆角笔画，
 * 24px outline / 24px solid 双版本，npm 离线打包，内网部署零外部请求）。
 * 仅显式注册项目实际使用的图标，Vite 摇树优化后产物只包含此处引入的 path。
 *
 * ── 统一使用规范 ──────────────────────────────────────────────
 * 【尺寸体系】导航/按钮 20px · 表单/行内 16px · 标题 24px · 空状态插画 48px+
 * 【颜色体系】普通灰色（继承文本色）· 主要操作用主色 text-primary/primary-* ·
 *             状态一律用语义色（待办灰 / 进行中蓝 / 完成绿 / 盯办金 / 警告红）
 * 【可访问性】装饰性图标不传 label（组件自动 aria-hidden）；
 *             功能性图标必须传 label（生成 role="img" + aria-label）
 * 【动画】enter=初次加载 300ms 缩放淡入（配 enterDelay 依次点亮）·
 *         loading=匀速旋转 · interactive=悬停放大 + 按压弹性回弹 ·
 *         outline↔solid 切换自动 250ms 交叉淡入淡出
 * 【主题】颜色全部 currentColor 继承，明暗切换 400ms 平滑过渡；
 *         暗色模式线框图标描边自动 +5%（1.5 → 1.575）
 * ─────────────────────────────────────────────────────────────
 */
import type { FunctionalComponent, SVGAttributes } from 'vue';

import {
  ArchiveBoxIcon,
  ArrowRightStartOnRectangleIcon,
  Bars3Icon,
  BellIcon,
  BugAntIcon,
  BuildingOffice2Icon,
  ChartBarIcon,
  ChatBubbleLeftRightIcon,
  CheckCircleIcon,
  CheckIcon,
  ClipboardDocumentListIcon,
  ClockIcon,
  CogIcon,
  Cog6ToothIcon,
  ComputerDesktopIcon,
  DocumentTextIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  MagnifyingGlassIcon,
  MoonIcon,
  PencilSquareIcon,
  QueueListIcon,
  StarIcon,
  SunIcon,
  TagIcon,
  UserGroupIcon,
  XCircleIcon,
} from '@heroicons/vue/24/outline';

import {
  BellIcon as BellSolidIcon,
  CheckCircleIcon as CheckCircleSolidIcon,
  CogIcon as CogSolidIcon,
  ExclamationTriangleIcon as ExclamationTriangleSolidIcon,
  InformationCircleIcon as InformationCircleSolidIcon,
  MoonIcon as MoonSolidIcon,
  StarIcon as StarSolidIcon,
  SunIcon as SunSolidIcon,
  XCircleIcon as XCircleSolidIcon,
} from '@heroicons/vue/24/solid';

export type IconComponent = FunctionalComponent<SVGAttributes>;

export interface IconEntry {
  outline: IconComponent;
  /** solid 版本缺省时回退到 outline（Heroicons 全量双版本，此处仅语义告警类注册 solid） */
  solid?: IconComponent;
}

/** 任务状态语义（StatusIcon 使用） */
export type StatusKind = 'todo' | 'doing' | 'done' | 'urgent';

/** 空状态插画种类（EmptyIllustration 使用） */
export type EmptyKind = 'coffee' | 'cat' | 'inbox';

/** 项目图标注册表：key = 业务语义名 */
export const ICON_REGISTRY = {
  // —— 侧边导航（原 emoji 全部替换）——
  clipboard: { outline: ClipboardDocumentListIcon },
  archive: { outline: ArchiveBoxIcon },
  view: { outline: MagnifyingGlassIcon },
  chart: { outline: ChartBarIcon },
  chat: { outline: ChatBubbleLeftRightIcon },
  bug: { outline: BugAntIcon },
  users: { outline: BuildingOffice2Icon },
  user: { outline: UserGroupIcon },
  tag: { outline: TagIcon },
  template: { outline: DocumentTextIcon },
  monitor: { outline: ComputerDesktopIcon },
  settings: { outline: Cog6ToothIcon },
  list: { outline: QueueListIcon },
  // —— 通用 ——
  bars: { outline: Bars3Icon },
  logout: { outline: ArrowRightStartOnRectangleIcon },
  check: { outline: CheckIcon },
  clock: { outline: ClockIcon },
  cog: { outline: CogIcon, solid: CogSolidIcon },
  sun: { outline: SunIcon, solid: SunSolidIcon },
  moon: { outline: MoonIcon, solid: MoonSolidIcon },
  pencil: { outline: PencilSquareIcon },
  bell: { outline: BellIcon, solid: BellSolidIcon },
  // —— 状态语义（Toast / 确认框 / 提示）——
  'check-circle': { outline: CheckCircleIcon, solid: CheckCircleSolidIcon },
  'x-circle': { outline: XCircleIcon, solid: XCircleSolidIcon },
  warning: { outline: ExclamationTriangleIcon, solid: ExclamationTriangleSolidIcon },
  info: { outline: InformationCircleIcon, solid: InformationCircleSolidIcon },
  star: { outline: StarIcon, solid: StarSolidIcon },
} satisfies Record<string, IconEntry>;

export type IconName = keyof typeof ICON_REGISTRY;
