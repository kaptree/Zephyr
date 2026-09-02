<script setup lang="ts">
/**
 * StatusIcon —— 状态语义图标系统（有温度的图标语言体系）
 *
 * todo   待办   空心时钟缓慢脉冲（2.4s）        灰色
 * doing  进行中 实心齿轮匀速旋转（2.6s 一圈）   蓝色
 * done   已完成 勾号绘制动画（400ms 描画）      绿色
 * urgent 盯办   金色星标闪烁（1.8s + 光晕）     金色
 *
 * tone：semantic = 按状态语义着色（默认）；inherit = 继承父级颜色
 * （用于深色徽章内部，如便签「已签收」徽章中的白色勾画）。
 */
import { computed } from 'vue';
import AppIcon from './AppIcon.vue';
import type { StatusKind } from './registry';

const props = withDefaults(
  defineProps<{
    status: StatusKind;
    size?: number;
    tone?: 'semantic' | 'inherit';
    label?: string;
  }>(),
  { size: 16, tone: 'semantic' }
);

const SEMANTIC: Record<StatusKind, { cls: string; text: string }> = {
  todo: { cls: 'animate-icon-clock-pulse text-slate-400 dark:text-slate-500', text: '待办' },
  doing: { cls: 'animate-icon-spin-slow text-blue-500 dark:text-blue-400', text: '进行中' },
  done: { cls: 'text-green-500 dark:text-green-400', text: '已完成' },
  urgent: { cls: 'animate-icon-star-twinkle text-amber-400 dark:text-amber-300', text: '盯办/高优' },
};

const toneCls = computed(() => (props.tone === 'semantic' ? SEMANTIC[props.status].cls : ''));
const aria = computed(() => props.label ?? SEMANTIC[props.status].text);
const hostStyle = computed(() => ({ width: `${props.size}px`, height: `${props.size}px` }));
</script>

<template>
  <!-- 已完成：内联勾画 SVG，400ms stroke-dashoffset 从左向右绘制 -->
  <span
    v-if="status === 'done'"
    class="icon-host"
    :class="toneCls"
    :style="hostStyle"
    role="img"
    :aria-label="aria"
  >
    <svg viewBox="0 0 24 24" fill="none" class="w-full h-full icon-check-draw">
      <path
        d="M4.5 12.75l6 6 9-13.5"
        stroke="currentColor"
        stroke-width="2.4"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-dasharray="32"
      />
    </svg>
  </span>

  <!-- 进行中：实心齿轮匀速旋转 -->
  <AppIcon
    v-else-if="status === 'doing'"
    name="cog"
    variant="solid"
    :size="size"
    :label="aria"
    :class="toneCls"
  />

  <!-- 待办：空心时钟缓慢脉冲 -->
  <AppIcon
    v-else-if="status === 'todo'"
    name="clock"
    variant="outline"
    :size="size"
    :label="aria"
    :class="toneCls"
  />

  <!-- 盯办/高优：金色星标闪烁 -->
  <AppIcon
    v-else
    name="star"
    variant="solid"
    :size="size"
    :label="aria"
    :class="toneCls"
  />
</template>
