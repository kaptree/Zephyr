<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useConfirm } from '@/composables/useConfirm'
import { AppIcon } from '@/components/icons'

const { state, settle } = useConfirm()

function onKeydown(e: KeyboardEvent) {
  if (!state.value.visible) return
  if (e.key === 'Escape') settle(false)
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <Transition name="shrink-out">
      <div
        v-if="state.visible"
        class="fixed inset-0 z-[130] flex items-center justify-center bg-black/40 px-4"
        @click.self="settle(false)"
      >
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-6 max-w-sm w-full">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center text-lg shrink-0"
              :class="
                state.options.danger
                  ? 'bg-red-100 dark:bg-red-900/50 text-red-500'
                  : 'bg-blue-100 dark:bg-blue-900/50 text-blue-500'
              "
            >
              <AppIcon
                :name="state.options.danger ? 'warning' : 'info'"
                variant="solid"
                :size="20"
              />
            </div>
            <h4 class="text-base font-semibold text-slate-800 dark:text-slate-100">
              {{ state.options.title || (state.options.danger ? '确认操作' : '提示') }}
            </h4>
          </div>
          <p class="mt-3 text-sm text-slate-500 dark:text-slate-400 whitespace-pre-line leading-relaxed">
            {{ state.options.message }}
          </p>
          <div class="mt-5 flex justify-end gap-2">
            <!-- 错位出现：取消先现（0ms），确认延迟 200ms 依次浮现 -->
            <button
              class="btn-secondary !px-4 !py-2 text-sm confirm-btn-stagger"
              @click="settle(false)"
            >
              {{ state.options.cancelText || '取消' }}
            </button>
            <button
              class="confirm-btn-stagger !px-4 !py-2 text-sm"
              :class="state.options.danger ? 'btn-danger' : 'btn-primary'"
              style="animation-delay: 0.2s"
              @click="settle(true)"
            >
              {{ state.options.confirmText || (state.options.danger ? '确认删除' : '确认') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
