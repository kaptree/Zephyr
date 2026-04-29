import { ref, computed } from 'vue'

interface VirtualScrollOptions {
  itemHeight: number
  overscan?: number
}

export function useVirtualScroll<T>(items: Ref<T[]>, options: VirtualScrollOptions) {
  const containerRef = ref<HTMLElement | null>(null)
  const scrollTop = ref(0)
  const containerHeight = ref(0)

  const visibleCount = computed(() => Math.ceil(containerHeight.value / options.itemHeight) + (options.overscan || 5) * 2)
  const startIndex = computed(() => Math.max(0, Math.floor(scrollTop.value / options.itemHeight) - (options.overscan || 5)))
  const endIndex = computed(() => Math.min(items.value.length, startIndex.value + visibleCount.value))

  const visibleItems = computed(() => items.value.slice(startIndex.value, endIndex.value))
  const totalHeight = computed(() => items.value.length * options.itemHeight)
  const offsetY = computed(() => startIndex.value * options.itemHeight)

  function handleScroll() {
    if (containerRef.value) {
      scrollTop.value = containerRef.value.scrollTop
      containerHeight.value = containerRef.value.clientHeight
    }
  }

  function initObserver() {
    if (containerRef.value) {
      containerHeight.value = containerRef.value.clientHeight
    }
  }

  return {
    containerRef,
    visibleItems,
    totalHeight,
    offsetY,
    handleScroll,
    initObserver,
  }
}

// 对于类型辅助
import { type Ref } from 'vue'
