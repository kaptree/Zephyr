<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const stats = ref([
  { label: '活跃便签', value: 12, color: 'bg-amber-50 text-amber-700' },
  { label: '已完成', value: 45, color: 'bg-green-50 text-green-700' },
  { label: '盯办中', value: 3, color: 'bg-red-50 text-red-700' },
  { label: '已归档', value: 89, color: 'bg-slate-50 text-slate-700' },
])

const weeklyData = [2, 5, 3, 4, 7, 1, 8]
const weeklyLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const maxWeekly = Math.max(...weeklyData)

const preferences = ref({
  desktopSyncNotify: true,
  completeSound: false,
  urgentAlert: true,
  defaultSort: 'created_at',
  gridWidth: 'auto',
})
</script>

<template>
  <div class="max-w-3xl">
    <h2 class="text-lg font-semibold text-slate-900 mb-6">个人中心</h2>

    <!-- 用户信息卡片 -->
    <div class="bg-white rounded-card border border-slate-100 p-6 mb-6">
      <div class="flex items-center gap-4">
        <div class="w-16 h-16 rounded-full bg-blue-500 flex items-center justify-center text-xl font-semibold text-white shrink-0">
          {{ auth.user?.name?.charAt(0) || '用' }}
        </div>
        <div>
          <h3 class="text-lg font-semibold text-slate-900">{{ auth.user?.name || '未登录' }}</h3>
          <p class="text-sm text-slate-500">{{ auth.user?.dept_name || '' }}</p>
          <p class="text-xs text-slate-400 mt-1">角色：{{ auth.user?.role || '—' }}</p>
        </div>
      </div>
    </div>

    <!-- 便签统计 -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div v-for="s in stats" :key="s.label" :class="['rounded-card p-5 text-center', s.color]">
        <div class="text-2xl font-bold">{{ s.value }}</div>
        <div class="text-xs mt-1 opacity-70">{{ s.label }}</div>
      </div>
    </div>

    <!-- 本周趋势 -->
    <div class="bg-white rounded-card border border-slate-100 p-6 mb-6">
      <h4 class="text-sm font-semibold text-slate-900 mb-4">本周趋势</h4>
      <div class="flex items-end justify-between gap-3" style="height: 120px">
        <div v-for="(val, i) in weeklyData" :key="i" class="flex flex-col items-center gap-2 flex-1 h-full justify-end">
          <span class="text-xs text-slate-500">{{ val }}</span>
          <div
            class="w-full rounded-t-md transition-smooth"
            :style="{
              height: `${(val / maxWeekly) * 80}px`,
              backgroundColor: val === maxWeekly ? '#3B82F6' : '#E2E8F0',
            }"
          />
          <span class="text-[10px] text-slate-400">{{ weeklyLabels[i] }}</span>
        </div>
      </div>
    </div>

    <!-- 偏好设置 -->
    <div class="bg-white rounded-card border border-slate-100 p-6 mb-6">
      <h4 class="text-sm font-semibold text-slate-900 mb-4">偏好设置</h4>
      <div class="space-y-4">
        <label v-for="(val, key) in preferences" :key="key" class="flex items-center justify-between cursor-pointer">
          <span class="text-sm text-slate-600">
            {{ key === 'desktopSyncNotify' ? '桌面端便签同步提醒' : key === 'completeSound' ? '完成时播放音效' : key === 'urgentAlert' ? '盯办强提醒' : '默认便签排序' }}
          </span>
          <template v-if="typeof val === 'boolean'">
            <input
              v-model="(preferences as Record<string, boolean>)[key]"
              type="checkbox"
              class="toggle toggle-sm toggle-primary"
            />
          </template>
          <select v-else class="input-field !w-auto text-xs" v-model="(preferences as Record<string, string>)[key]">
            <option value="created_at">按创建时间</option>
            <option value="updated_at">按更新时间</option>
            <option value="priority">按优先级</option>
          </select>
        </label>
      </div>
    </div>

    <!-- 退出登录 -->
    <div class="flex justify-between">
      <button class="btn-primary text-sm">保存设置</button>
      <button
        class="px-5 py-2.5 text-sm text-red-600 bg-red-50 rounded-btn hover:bg-red-100 transition-smooth"
        @click="auth.logout(); $router.push('/login')"
      >
        退出登录
      </button>
    </div>
  </div>
</template>
