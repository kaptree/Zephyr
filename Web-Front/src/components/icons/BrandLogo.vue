<script setup lang="ts">
/**
 * BrandLogo —— 品牌标志（有温度的图标语言体系）
 * 圆角方块 + 燕子剪影（渐变主色），页面加载时以 600ms 模糊到清晰浮现。
 * useId 保证多实例渐变 id 不冲突。
 */
import { useId } from 'vue';

withDefaults(
  defineProps<{
    size?: number;
    label?: string;
  }>(),
  { size: 30, label: '轻燕工作台' }
);

const gid = `zy-logo-g-${useId().replace(/[^a-zA-Z0-9-]/g, '')}`;
</script>

<template>
  <span
    class="icon-host animate-icon-logo-reveal"
    :style="{ width: `${size}px`, height: `${size}px` }"
    role="img"
    :aria-label="label"
  >
    <svg viewBox="0 0 32 32" fill="none" class="w-full h-full">
      <defs>
        <linearGradient :id="gid" x1="5" y1="3" x2="27" y2="29" gradientUnits="userSpaceOnUse">
          <stop stop-color="#60A5FA" />
          <stop offset="1" stop-color="#2563EB" />
        </linearGradient>
      </defs>
      <rect x="1" y="1" width="30" height="30" rx="9" :fill="`url(#${gid})`" />
      <!-- 燕子剪影：展翅向右，尾羽分叉 -->
      <path
        d="M6.8 20.4c3.4-.7 5.9-2.8 7.1-6.1 1 2.3 2.7 3.8 5.5 4.2-1.3 3.1-3.9 5.2-7.3 5.9.8-1.2 1.2-2.4 1.3-3.7-2.3.4-4.5.2-6.6-.3z"
        fill="#fff"
        fill-opacity="0.96"
      />
      <circle cx="21.6" cy="10.4" r="1.7" fill="#fff" fill-opacity="0.85" />
    </svg>
  </span>
</template>
