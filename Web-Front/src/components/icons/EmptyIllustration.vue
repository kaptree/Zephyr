<script setup lang="ts">
/**
 * EmptyIllustration —— 空状态品牌插画（有温度的图标语言体系）
 * 缓慢呼吸（3.2s）整体动画；咖啡杯热气错频飘动、打盹猫 zzz 轻柔浮现。
 * 颜色继承父级文本色（细线），primary 蓝作点缀，明暗主题自适应。
 */
import type { EmptyKind } from './registry';

withDefaults(
  defineProps<{
    kind?: EmptyKind;
    size?: number;
    label?: string;
  }>(),
  { kind: 'coffee', size: 72, label: '暂无内容' }
);
</script>

<template>
  <span
    class="icon-host icon-breathe-slow"
    :style="{ width: `${size}px`, height: `${size}px` }"
    role="img"
    :aria-label="label"
  >
    <!-- ☕ 咖啡杯：热气三条错频飘起 -->
    <svg v-if="kind === 'coffee'" viewBox="0 0 64 64" fill="none" class="w-full h-full">
      <g class="icon-steam" stroke="currentColor" stroke-width="2.6" stroke-linecap="round">
        <path d="M24 15c-2.4-2.6 1.6-4.6 0-8.4" opacity="0.5" />
        <path d="M32 13c-2.4-2.6 1.6-4.6 0-8.4" opacity="0.7" />
        <path d="M40 15c-2.4-2.6 1.6-4.6 0-8.4" opacity="0.5" />
      </g>
      <!-- 杯身 -->
      <rect x="13" y="24" width="30" height="24" rx="6.5" stroke="currentColor" stroke-width="2.6" />
      <!-- 把手 -->
      <path
        d="M43 30h3.5a6.5 6.5 0 010 13H43"
        stroke="currentColor"
        stroke-width="2.6"
        stroke-linecap="round"
      />
      <!-- 杯中咖啡液面（品牌蓝点缀） -->
      <path d="M19 32h13" stroke="#3B82F6" stroke-width="2.6" stroke-linecap="round" opacity="0.75" />
      <!-- 碟子 -->
      <path d="M9 54h40" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" opacity="0.45" />
    </svg>

    <!-- 😴 打盹猫：蜷缩 + zzz 浮现 -->
    <svg v-else-if="kind === 'cat'" viewBox="0 0 64 64" fill="none" class="w-full h-full">
      <!-- 身体 -->
      <circle cx="30" cy="42" r="14.5" stroke="currentColor" stroke-width="2.6" />
      <!-- 耳朵 -->
      <path d="M20.5 31.5l1.8-7 6.2 3.4" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M39.5 31.5l-1.8-7-6.2 3.4" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" />
      <!-- 闭眼（安睡） -->
      <path d="M23 42q2.2 2.4 4.4 0" stroke="#3B82F6" stroke-width="2.2" stroke-linecap="round" />
      <path d="M33 42q2.2 2.4 4.4 0" stroke="#3B82F6" stroke-width="2.2" stroke-linecap="round" />
      <!-- 鼻嘴 -->
      <path d="M29.5 47.5h1l-.5.9-.5-.9h1" fill="#3B82F6" opacity="0.8" />
      <!-- 尾巴 -->
      <path d="M44.5 44c5.5 1.5 8.5-1.5 7-7.5" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" />
      <!-- zzz -->
      <g class="icon-zzz" stroke="#3B82F6" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M50 12h5.5l-5.5 5.5H55" />
        <path d="M56.5 24h4l-4 4h4" opacity="0.6" />
      </g>
    </svg>

    <!-- 📥 收件箱（保留原 heroicons 线条，呼吸呈现） -->
    <svg v-else viewBox="0 0 24 24" fill="none" class="w-full h-full">
      <path
        d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
  </span>
</template>
