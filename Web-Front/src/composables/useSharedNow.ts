import { ref, onMounted, onUnmounted } from 'vue';

/**
 * 全站共享的「当前时间」ticker。
 *
 * 背景：任务卡片需要每 30s 刷新一次倒计时文本。若每张卡片各自 setInterval，
 * 工作台 20+ 张卡片就会产生 20+ 个定时器，带来持续的调度开销。
 * 这里改为模块级单例：所有订阅者共享同一个 now ref 与唯一 interval，
 * 首个订阅者启动、最后一个取消时清除，语义不变而开销恒定。
 */
const INTERVAL_MS = 30000;

const now = ref(Date.now());
let timer: ReturnType<typeof setInterval> | null = null;
let subscribers = 0;

function startTicker() {
  if (timer !== null) return;
  now.value = Date.now();
  timer = setInterval(() => {
    now.value = Date.now();
  }, INTERVAL_MS);
}

function stopTicker() {
  if (timer === null) return;
  clearInterval(timer);
  timer = null;
}

/** 订阅共享 ticker：返回共享的当前时间 ref；组件卸载时自动退订 */
export function useSharedNow() {
  onMounted(() => {
    subscribers += 1;
    startTicker();
  });
  onUnmounted(() => {
    subscribers = Math.max(0, subscribers - 1);
    if (subscribers === 0) stopTicker();
  });
  return now;
}
