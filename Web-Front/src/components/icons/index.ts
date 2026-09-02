/**
 * 有温度的图标语言体系 —— 统一出口（美化工程 · 第 7 项）
 *
 * 所有图标组件统一管理于 src/components/icons/：
 * - AppIcon          通用图标（outline/solid 双版本、尺寸、动画、可访问性）
 * - StatusIcon       状态语义系统（待办/进行中/完成/盯办）
 * - BrandLogo        品牌 Logo（模糊到清晰浮现）
 * - EmptyIllustration 空状态呼吸插画（咖啡杯/打盹猫/收件箱）
 * - registry         图标注册表与使用规范
 */
export { default as AppIcon } from './AppIcon.vue';
export { default as StatusIcon } from './StatusIcon.vue';
export { default as BrandLogo } from './BrandLogo.vue';
export { default as EmptyIllustration } from './EmptyIllustration.vue';
export {
  ICON_REGISTRY,
  type IconName,
  type IconEntry,
  type StatusKind,
  type EmptyKind,
} from './registry';
