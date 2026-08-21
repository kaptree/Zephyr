<script setup lang="ts">
import Toast from '@/components/common/Toast.vue'
import FloatingPopups from '@/components/common/FloatingPopups.vue'
import { useDarkMode } from '@/composables/useDarkMode'

useDarkMode()
</script>

<template>
  <router-view />
  <Toast />
  <FloatingPopups />
</template>
