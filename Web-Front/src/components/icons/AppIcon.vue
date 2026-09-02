<script setup lang="ts">
/**
 * AppIcon —— 统一图标入口组件（有温度的图标语言体系）
 *
 * - variant：outline（线框，默认）/ solid（填色），切换时 250ms 交叉淡入淡出
 * - size：16（表单）/ 20（导航按钮）/ 24（标题）等任意 px
 * - label：功能性图标必传（role=img + aria-label）；缺省视为装饰性（aria-hidden）
 * - loading：匀速旋转加载态
 * - interactive：悬停轻微放大变色、点击按压下沉 + 弹性回弹
 * - enter + enterDelay：页面初次加载 300ms 缩放淡入，延迟依次点亮
 */
import { computed, type CSSProperties } from 'vue';
import { ICON_REGISTRY, type IconName } from './registry';

const props = withDefaults(
  defineProps<{
    name: IconName;
    variant?: 'outline' | 'solid';
    size?: number;
    label?: string;
    loading?: boolean;
    interactive?: boolean;
    enter?: boolean;
    enterDelay?: number;
  }>(),
  { variant: 'outline', size: 20 }
);

const comp = computed(() => {
  const entry = ICON_REGISTRY[props.name];
  if (!entry) throw new Error(`[AppIcon] 未注册图标: ${props.name}`);
  return props.variant === 'solid' ? (entry.solid ?? entry.outline) : entry.outline;
});

const hostStyle = computed(() => {
  const style: Record<string, string | undefined> = {
    width: `${props.size}px`,
    height: `${props.size}px`,
  };
  if (props.enter && props.enterDelay) style['--icon-stagger'] = `${props.enterDelay}ms`;
  return style as CSSProperties;
});
</script>

<template>
  <span
    class="icon-host"
    :class="[
      variant === 'outline' && 'icon-outline',
      loading && 'animate-spin',
      interactive && 'icon-interactive',
      enter && 'icon-enter',
    ]"
    :style="hostStyle"
    :role="label ? 'img' : undefined"
    :aria-label="label || undefined"
    :aria-hidden="label ? undefined : true"
  >
    <!-- outline ↔ solid 切换：250ms 交叉淡入淡出（icon-fade 全局过渡） -->
    <Transition name="icon-fade">
      <component :is="comp" :key="`${props.name}-${props.variant}`" class="w-full h-full" />
    </Transition>
  </span>
</template>
