<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from '@/components/layout/Sidebar.vue'
import TopBar from '@/components/layout/TopBar.vue'
import { useUserBackground } from '@/composables/useUserBackground'
import { useAchievementStore } from '@/stores/achievements'

const route = useRoute()
const achievementStore = useAchievementStore()

// <main> 引用：路由过渡期间打 data-route-transitioning 标记，
// decor.css 据此临时关闭玻璃卡 backdrop-filter（切页掉帧主因），动画结束恢复
const mainEl = ref<HTMLElement | null>(null)
function setTransitioning(v: boolean) {
  if (mainEl.value) mainEl.value.dataset.routeTransitioning = String(v)
}

// 进入主布局：拉取成就解锁进度（侧边栏角标）+ 触发一轮成就检查（登录成就等，新解锁弹 2s 恭喜弹窗）
onMounted(() => {
  achievementStore.fetchList()
  achievementStore.checkAndNotify()
})

// KeepAlive 缓存的高频页面（组件内需 defineOptions 声明同名 name）
const keepAlivePages = ['WorkbenchPage', 'IssuesPage']

// 过渡方向由 router.afterEach 依据路由层级写入 meta.transitionName
const transitionName = computed(() => (route.meta.transitionName as string) || 'fade-up')

// 平台背景图（个人中心设置）：fixed 层铺在内容区下方，卡片保持不透明保证可读性
const { bgStyle } = useUserBackground()
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-slate-50 dark:bg-slate-950 transition-colors duration-300">
    <!-- 平台背景图层：位于根背景之上、内容之下，透明度由用户调节 -->
    <div
      v-if="bgStyle"
      class="fixed inset-0 z-0 pointer-events-none transition-opacity duration-300"
      :style="bgStyle"
      aria-hidden="true"
    />
    <Sidebar class="relative z-10" />
    <div class="relative z-10 flex-1 flex flex-col overflow-hidden">
      <TopBar />
      <main ref="mainEl" class="flex-1 overflow-auto p-6 scrollbar-thin transition-colors duration-300">
        <router-view v-slot="{ Component, route: viewRoute }">
          <transition
            :name="transitionName"
            mode="out-in"
            @before-leave="setTransitioning(true)"
            @after-enter="setTransitioning(false)"
            @enter-cancelled="setTransitioning(false)"
            @leave-cancelled="setTransitioning(false)"
          >
            <keep-alive :include="keepAlivePages">
              <component :is="Component" :key="viewRoute.path" />
            </keep-alive>
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* ===== 空间感知路由过渡 =====
   性能要点：active 阶段统一 will-change 提前提升合成层；时长收紧，
   out-in 串行模式下总等待 = leave + enter，原 0.6s+ 收敛至约 0.4s */

/* 下钻（列表 → 详情）：详情滑入放大，列表缩小左移退出 */
.slide-forward-enter-active {
  transition:
    opacity 0.28s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
  will-change: opacity, transform;
}
.slide-forward-leave-active {
  transition:
    opacity 0.12s ease-in,
    transform 0.12s ease-in;
  will-change: opacity, transform;
}
.slide-forward-enter-from {
  opacity: 0;
  transform: translateX(48px) scale(0.98);
}
.slide-forward-leave-to {
  opacity: 0;
  transform: translateX(-28px) scale(0.96);
}

/* 返回（详情 → 列表）：详情右滑缩小退出，列表从右滑入放大 */
.slide-back-enter-active {
  transition:
    opacity 0.24s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.24s cubic-bezier(0.16, 1, 0.3, 1);
  will-change: opacity, transform;
}
.slide-back-leave-active {
  transition:
    opacity 0.12s ease-in,
    transform 0.12s ease-in;
  will-change: opacity, transform;
}
.slide-back-enter-from {
  opacity: 0;
  transform: translateX(48px) scale(0.96);
}
.slide-back-leave-to {
  opacity: 0;
  transform: translateX(28px) scale(0.98);
}

/* 同级切换（侧边导航 / Tab）：淡入淡出 + 轻微上移 */
.fade-up-enter-active {
  transition:
    opacity 0.22s ease-out,
    transform 0.22s ease-out;
  will-change: opacity, transform;
}
.fade-up-leave-active {
  transition:
    opacity 0.12s ease-in,
    transform 0.12s ease-in;
  will-change: opacity, transform;
}
.fade-up-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.fade-up-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
