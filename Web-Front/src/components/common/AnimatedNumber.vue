<script setup lang="ts">
/* 数字滚动动画（美化工程 · 微交互动画语言）
   统计卡片数字变化时以计数器方式逐位滚动更新，600ms ease-out 缓动。
   用法：<AnimatedNumber :value="stats.total" /> 或带格式化：
        <AnimatedNumber :value="rate" :format="(n) => n.toFixed(1) + '%'" /> */
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'

const props = withDefaults(defineProps<{
  value: number
  duration?: number
  format?: (n: number) => string
}>(), {
  duration: 600,
  format: (n: number) => String(n),
})

const display = ref(0)
let raf = 0

function stop() {
  if (raf) {
    cancelAnimationFrame(raf)
    raf = 0
  }
}

function animateTo(to: number, from: number) {
  stop()
  if (!isFinite(from) || !isFinite(to)) {
    display.value = isFinite(to) ? to : 0
    return
  }
  const t0 = performance.now()
  const duration = props.duration
  const step = (t: number) => {
    const p = duration > 0 ? Math.min(1, (t - t0) / duration) : 1
    const eased = 1 - Math.pow(1 - p, 3) // ease-out cubic
    display.value = from + (to - from) * eased
    raf = p < 1 ? requestAnimationFrame(step) : 0
  }
  raf = requestAnimationFrame(step)
}

onMounted(() => animateTo(Number(props.value) || 0, 0))

watch(() => props.value, (nv, ov) => {
  animateTo(Number(nv) || 0, Number(ov) || 0)
})

onBeforeUnmount(stop)
</script>

<template>
  <span class="tabular-nums">{{ props.format(display) }}</span>
</template>
