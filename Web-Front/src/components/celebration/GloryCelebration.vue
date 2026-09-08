<script setup lang="ts">
/* 归档「荣耀时刻」庆祝层（美化工程 · 任务闭环仪式感）
   挂载于 App.vue 全局单例，Teleport 到 body（z-[90]，高于反馈弹窗 z-[80]）。
   演出编排（与 useGloryCelebration 的 JS 时间线对齐）：
     绽放 0–1.2s   纸屑爆发（JS 触发）+ 勋章弹跳落位 + 对勾描画 600ms + 金色光晕
     升华 1.2–2.5s 「🎉 归档成功！」弹性浮现 + 数字滚动 1s + 背景气泡升腾
     余韵 2.5–3.5s 勋章缩小淡出 + 「查看成果」浮现 + 右上角通知呼应滑入
   动画期间遮罩拦截全部交互（防误触），仅「查看成果」可点；Esc 可提前收场。 */
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue';
import { useGloryCelebration } from '@/composables/useGloryCelebration';
import AnimatedNumber from '@/components/common/AnimatedNumber.vue';

const { visible, celebration, countStarted, viewResult, dismiss } = useGloryCelebration();

const viewBtn = ref<HTMLButtonElement | null>(null);

// 背景气泡：每次庆祝随机生成位置/尺寸/延迟/速率（纯 CSS float 动画，向上飘移淡出）
interface Bubble {
  left: number;
  size: number;
  delay: number;
  dur: number;
  o: number;
}
const bubbles = ref<Bubble[]>([]);
watch(visible, (v) => {
  if (!v) return;
  bubbles.value = Array.from({ length: 14 }, () => ({
    left: Math.random() * 96,
    size: 6 + Math.random() * 14,
    delay: Math.random() * 2.4,
    dur: 3.2 + Math.random() * 2.4,
    o: 0.25 + Math.random() * 0.35,
  }));
  // 焦点移入「对话框」（唯一可交互元素），保证键盘导航连贯
  nextTick(() => viewBtn.value?.focus({ preventScroll: true }));
});

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && visible.value) dismiss();
}
onMounted(() => window.addEventListener('keydown', onKeydown));
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown));
</script>

<template>
  <Teleport to="body">
    <Transition name="glory-fade">
      <div
        v-if="visible && celebration"
        class="fixed inset-0 z-[90] overflow-hidden select-none"
        role="dialog"
        aria-modal="true"
        aria-label="任务归档庆祝"
      >
        <!-- 遮罩：拦截全部点击（防干扰），明暗自适应柔和压暗 -->
        <div class="absolute inset-0 bg-slate-900/30 dark:bg-black/55 backdrop-blur-[2px]" />

        <!-- 背景气泡：以不同速率向上飘移并淡出，不干扰前景 -->
        <span
          v-for="(b, i) in bubbles"
          :key="i"
          class="glory-bubble absolute -bottom-8 rounded-full bg-white/50 dark:bg-white/10"
          :style="{
            left: b.left + '%',
            width: b.size + 'px',
            height: b.size + 'px',
            '--glory-delay': b.delay + 's',
            '--glory-dur': b.dur + 's',
            '--glory-bubble-o': b.o,
          }"
        />

        <!-- 右上角通知呼应：「任务已归档」滑入，与庆祝动画形成呼应 -->
        <div class="glory-notify-in absolute top-20 right-5 pointer-events-none">
          <div
            class="flex items-center gap-3 rounded-xl bg-white dark:bg-slate-800 shadow-xl border border-slate-100 dark:border-slate-700 px-4 py-3"
          >
            <span
              class="flex w-8 h-8 rounded-full bg-emerald-100 dark:bg-emerald-900/60 items-center justify-center shrink-0"
            >
              <svg class="w-4 h-4 text-emerald-600 dark:text-emerald-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path d="M5 12.5 L10 17 L19 8" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </span>
            <div class="min-w-0">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-100">任务已归档</p>
              <p class="text-xs text-slate-400 truncate max-w-[220px]">{{ celebration.taskName }}</p>
            </div>
          </div>
        </div>

        <!-- 中央演出 -->
        <div class="relative h-full flex flex-col items-center justify-center gap-5 pointer-events-none px-6">
          <!-- 勋章：外层负责余韵期缩小淡出，内层负责弹跳落位 + 微脉动 -->
          <div class="glory-medal-out">
            <div class="relative w-28 h-32">
              <!-- 金色光晕：对勾描画完成后向外扩散一圈 -->
              <span
                class="glory-ripple absolute left-1/2 top-[82px] w-[96px] h-[96px] rounded-full border-2 border-amber-400/70 dark:border-amber-300/50"
              />
              <div class="glory-medal-drop">
                <svg
                  class="glory-medal-pulse w-28 h-32 drop-shadow-[0_18px_32px_rgba(245,158,11,0.35)]"
                  viewBox="0 0 120 140"
                  fill="none"
                  aria-hidden="true"
                >
                  <defs>
                    <!-- 描金光泽渐变（内联，无外部依赖） -->
                    <linearGradient id="glory-gold" x1="20" y1="48" x2="100" y2="128" gradientUnits="userSpaceOnUse">
                      <stop stop-color="#FDE68A" />
                      <stop offset="0.55" stop-color="#F59E0B" />
                      <stop offset="1" stop-color="#B45309" />
                    </linearGradient>
                    <linearGradient id="glory-ribbon" x1="28" y1="6" x2="92" y2="48" gradientUnits="userSpaceOnUse">
                      <stop stop-color="#60A5FA" />
                      <stop offset="1" stop-color="#2563EB" />
                    </linearGradient>
                    <radialGradient id="glory-inner" cx="0.35" cy="0.3" r="0.9">
                      <stop stop-color="#FFFBEB" />
                      <stop offset="0.6" stop-color="#FCD34D" />
                      <stop offset="1" stop-color="#F59E0B" />
                    </radialGradient>
                  </defs>
                  <!-- 绶带（品牌蓝） -->
                  <path d="M45 6 L60 38 L75 6 L92 6 L70 48 L50 48 L28 6 Z" fill="url(#glory-ribbon)" />
                  <!-- 勋章主体：描金 -->
                  <circle cx="60" cy="88" r="40" fill="url(#glory-gold)" stroke="#B45309" stroke-width="2" />
                  <circle cx="60" cy="88" r="31" fill="url(#glory-inner)" stroke="#FCD34D" stroke-width="1.5" />
                  <!-- 对勾：stroke-dasharray 从 0→100% 绘制（600ms） -->
                  <path
                    class="glory-check-path"
                    d="M44 89 L55 100 L78 76"
                    stroke="#92400E"
                    stroke-width="6"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    pathLength="100"
                  />
                  <!-- 光泽高光 -->
                  <ellipse cx="47" cy="72" rx="15" ry="7.5" fill="white" opacity="0.4" transform="rotate(-24 47 72)" />
                </svg>
              </div>
            </div>
          </div>

          <!-- 主标题：缩放 + 淡入 + 弹性回弹 -->
          <h1 class="glory-title-in text-3xl md:text-4xl font-bold text-emerald-600 dark:text-emerald-400 tracking-tight">
            🎉 归档成功！
          </h1>

          <!-- 副标题：已完成 × 任务名（数字滚动 1s ease-out，与升华阶段同步登场） -->
          <p
            class="glory-fade-up text-sm md:text-base text-slate-600 dark:text-slate-300 text-center"
            style="--glory-delay: 1.35s"
          >
            <template v-if="celebration.countTo != null">
              已完成
              <span class="text-2xl font-bold text-blue-600 dark:text-blue-400 align-middle mx-1">
                <AnimatedNumber v-if="countStarted" :value="celebration.countTo" :duration="1000" />
              </span>
              个任务 ·
            </template>
            <span class="font-medium text-slate-800 dark:text-slate-100">「{{ celebration.taskName }}」</span>
          </p>

          <!-- 「查看成果」：余韵阶段浮现，庆祝层唯一可交互元素 -->
          <button
            ref="viewBtn"
            class="glory-fade-up pointer-events-auto mt-1 px-6 py-2.5 rounded-xl bg-blue-500 dark:bg-blue-600 text-white text-sm font-medium shadow-lg shadow-blue-500/30 hover:bg-blue-600 dark:hover:bg-blue-500 active:scale-[0.98] transition-smooth focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-300 dark:focus-visible:ring-blue-700"
            style="--glory-delay: 2.5s"
            @click="viewResult"
          >
            查看成果 →
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
