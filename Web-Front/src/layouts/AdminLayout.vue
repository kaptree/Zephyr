<script setup lang="ts">
import Sidebar from '@/components/layout/Sidebar.vue'
import TopBar from '@/components/layout/TopBar.vue'
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-[#F8FAFC]">
    <Sidebar />
    <div class="flex-1 flex flex-col overflow-hidden">
      <TopBar />
      <main class="flex-1 overflow-auto p-6 scrollbar-thin">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
