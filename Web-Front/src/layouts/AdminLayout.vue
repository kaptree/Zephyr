<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from '@/components/layout/Sidebar.vue'
import TopBar from '@/components/layout/TopBar.vue'
import { useUserBackground } from '@/composables/useUserBackground'

const route = useRoute()

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
      <main class="flex-1 overflow-auto p-6 scrollbar-thin transition-colors duration-300">
        <router-view v-slot="{ Component, route: viewRoute }">
          <transition :name="transitionName" mode="out-in">
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
/* ===== 空间感知路由过渡 ===== */

/* 下钻（列表 → 详情）：详情滑入放大，列表缩小左移退出 */
.slide-forward-enter-active {
  transition:
    opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-forward-leave-active {
  transition:
    opacity 0.18s ease-in,
    transform 0.18s ease-in;
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
    opacity 0.3s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-back-leave-active {
  transition:
    opacity 0.2s ease-in,
    transform 0.2s ease-in;
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
    opacity 0.3s ease-out,
    transform 0.3s ease-out;
}
.fade-up-leave-active {
  transition:
    opacity 0.18s ease-in,
    transform 0.18s ease-in;
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
