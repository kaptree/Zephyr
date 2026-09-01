<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import Toast from '@/components/common/Toast.vue'
import FloatingPopups from '@/components/common/FloatingPopups.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useDarkMode } from '@/composables/useDarkMode'
import { useOffline } from '@/composables/useOffline'

useDarkMode()
// 离线检测（轻量级通知美学）：
// - 后端不可达（轮询 /api/v1/ping 校验）→ 顶部「离线缓存提醒」横幅
// - 网络级断开（navigator.onLine）→ 顶部「离线模式」横幅
const { isOnline, backendOnline } = useOffline()

// 「优雅退场」：页面关闭（标签页/窗口）前，内容从四周向中心收拢并淡出（500ms），
// 模拟“文件归档”的隐喻。受浏览器机制限制为尽力而为：卸载流程允许时动画可见。
function handlePageCollapse() {
  const app = document.getElementById('app')
  if (app) app.classList.add('page-fold-out')
}

onMounted(() => {
  window.addEventListener('beforeunload', handlePageCollapse)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') handlePageCollapse()
  })
})
onUnmounted(() => {
  window.removeEventListener('beforeunload', handlePageCollapse)
})
</script>

<template>
  <router-view />
  <!-- 离线双横幅：从顶部滑入 400ms，白色圆点微弱脉冲提醒 -->
  <Transition name="offline-banner">
    <div
      v-if="!isOnline"
      class="fixed top-0 left-0 right-0 z-[200] flex items-center justify-center gap-2 py-2 px-4 bg-amber-500/90 dark:bg-amber-600/90 backdrop-blur-sm text-white text-sm font-medium shadow-md"
    >
      <span class="w-2 h-2 rounded-full bg-white animate-pulse shrink-0" />
      离线模式：网络连接已断开，部分功能可能不可用，恢复后自动同步
    </div>
    <div
      v-else-if="!backendOnline"
      class="fixed top-0 left-0 right-0 z-[200] flex items-center justify-center gap-2 py-2 px-4 bg-orange-500/90 dark:bg-orange-600/90 backdrop-blur-sm text-white text-sm font-medium shadow-md"
    >
      <span class="w-2 h-2 rounded-full bg-white animate-pulse shrink-0" />
      离线缓存提醒：无法连接服务器，当前展示本地缓存数据，恢复后自动同步
    </div>
  </Transition>
  <Toast />
  <FloatingPopups />
  <ConfirmDialog />
</template>
