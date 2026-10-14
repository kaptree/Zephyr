<script setup lang="ts">
</script>

<template>
  <div class="page-shell page-shell--screen min-h-screen flex items-center justify-center">
    <!-- ===== 背景装饰层（日间晨曦 / 夜间深空） ===== -->
    <div class="decor-layer" aria-hidden="true">
      <div class="aurora aurora-1" />
      <div class="aurora aurora-2" />
      <div class="aurora aurora-3" />
      <div class="grid-overlay" />
      <div class="stars" />
      <div class="horizon-glow" />
      <div class="sun-glow" />
    </div>

    <div class="rise text-center">
      <h1 class="text-6xl font-bold page-title mb-4">403</h1>
      <p class="text-slate-500 text-lg">暂无访问权限</p>
      <router-link to="/workbench" class="btn-primary inline-block mt-6 no-underline">返回工作台</router-link>
    </div>
  </div>
</template>
