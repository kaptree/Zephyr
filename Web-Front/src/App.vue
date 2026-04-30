<script setup lang="ts">
import Toast from '@/components/common/Toast.vue'
import { useDarkMode } from '@/composables/useDarkMode'

useDarkMode()
</script>

<template>
  <router-view />
  <Toast />
</template>
