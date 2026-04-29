<script setup lang="ts">
withDefaults(defineProps<{
  lines?: number
  width?: string
  height?: string
}>(), {
  lines: 3,
  width: '100%',
  height: '44px',
})
</script>

<template>
  <div :style="{ width }">
    <div
      v-for="i in lines"
      :key="i"
      class="skeleton rounded-md mb-3"
      :style="{
        height,
        width: i === lines ? '60%' : '100%',
      }"
    />
  </div>
</template>
