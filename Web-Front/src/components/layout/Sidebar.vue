<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { AppIcon, BrandLogo } from '@/components/icons';
import type { IconName } from '@/components/icons';

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const collapsed = ref(localStorage.getItem('sidebar_collapsed') === 'true');

watch(collapsed, (val) => {
  localStorage.setItem('sidebar_collapsed', String(val));
});

interface MenuItem {
  icon: IconName;
  label: string;
  path: string;
  permission?: string;
  adminOnly?: boolean;
}

const menuItems: MenuItem[] = [
  { icon: 'clipboard', label: '工作台', path: '/workbench' },
  { icon: 'archive', label: '归档查询', path: '/workbench/archive' },
  {
    icon: 'view',
    label: '用户工作台',
    path: '/workbench/inspect',
    permission: 'inspect_user_workbench',
  },
  { icon: 'chart', label: '工作成效分析', path: '/analytics' },
  { icon: 'chat', label: '聊天', path: '/chat' },
  { icon: 'bug', label: 'Bug 反馈', path: '/issues' },
  {
    icon: 'users',
    label: '部门管理',
    path: '/admin/departments',
    permission: 'manage_departments',
  },
  { icon: 'user', label: '人员管理', path: '/admin/users', permission: 'manage_users' },
  { icon: 'tag', label: '标签管理', path: '/admin/tags', permission: 'manage_tags' },
  { icon: 'template', label: '模板管理', path: '/admin/templates' },
  { icon: 'monitor', label: '数据大屏', path: '/screen/default', permission: 'access_screen' },
  {
    icon: 'settings',
    label: '系统管理',
    path: '/admin/system',
    permission: 'manage_system',
    adminOnly: true,
  },
  {
    icon: 'list',
    label: '操作日志',
    path: '/admin/operation-log',
    permission: 'manage_system',
    adminOnly: true,
  },
];

const bottomItems: MenuItem[] = [{ icon: 'settings', label: '个人中心', path: '/profile' }];

const visibleMenuItems = computed(() =>
  menuItems.filter((item) => {
    if (item.adminOnly && !auth.isAdmin) return false;
    if (item.adminOnly && auth.isAdmin) return true;
    if (!item.permission) return true;
    return auth.permissions.includes(item.permission as never);
  })
);

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/');
}

function navigate(path: string) {
  router.push(path);
}
</script>

<template>
  <aside
    :class="[
      'bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-700 flex flex-col transition-colors duration-300 overflow-hidden',
      collapsed ? 'w-16' : 'w-60',
    ]"
  >
    <!-- 顶部 Logo -->
    <div
      class="h-14 flex items-center border-b border-slate-200 dark:border-slate-700 px-4 shrink-0 transition-colors duration-300"
    >
      <button
        class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth shrink-0"
        @click="collapsed = !collapsed"
      >
        <AppIcon name="bars" :size="20" label="折叠/展开侧边栏" interactive />
      </button>
      <BrandLogo :size="28" class="ml-2.5 shrink-0" />
      <transition name="fade">
        <span
          v-if="!collapsed"
          class="ml-2 text-sm font-semibold text-slate-900 dark:text-slate-100 whitespace-nowrap truncate"
          >轻燕</span
        >
      </transition>
    </div>

    <!-- 菜单项 -->
    <nav class="flex-1 overflow-y-auto scrollbar-thin py-2">
      <ul class="space-y-0.5 px-2">
        <li v-for="(item, idx) in visibleMenuItems" :key="item.path">
          <button
            :class="[
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-btn text-sm transition-smooth',
              isActive(item.path)
                ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 font-medium border-l-[3px] border-blue-500 dark:border-blue-400'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 border-l-[3px] border-transparent',
            ]"
            v-tooltip="collapsed ? item.label : ''"
            @click="navigate(item.path)"
          >
            <!-- 图标：初次加载 300ms 缩放淡入，60ms 间隔依次点亮 -->
            <AppIcon :name="item.icon" :size="20" enter :enter-delay="100 + idx * 60" />
            <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
          </button>
        </li>
      </ul>
    </nav>

    <!-- 底部区域 -->
    <div
      class="border-t border-slate-200 dark:border-slate-700 p-2 shrink-0 transition-colors duration-300"
    >
      <ul class="space-y-0.5">
        <li v-for="(item, idx) in bottomItems" :key="item.path">
          <button
            :class="[
              'w-full flex items-center gap-3 px-3 py-2.5 rounded-btn text-sm transition-smooth',
              isActive(item.path)
                ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 font-medium border-l-[3px] border-blue-500 dark:border-blue-400'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 border-l-[3px] border-transparent',
            ]"
            v-tooltip="collapsed ? item.label : ''"
            @click="navigate(item.path)"
          >
            <AppIcon :name="item.icon" :size="20" enter :enter-delay="1000 + idx * 60" />
            <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
          </button>
        </li>
      </ul>

      <!-- 用户信息 -->
      <div v-if="auth.user" class="flex items-center gap-3 px-3 py-2.5 mt-1">
        <div
          class="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-medium shrink-0"
        >
          {{ auth.user.name.charAt(0) }}
        </div>
        <div v-if="!collapsed" class="flex-1 min-w-0">
          <div class="text-xs font-medium text-slate-900 dark:text-slate-100 truncate">
            {{ auth.user.name }}
          </div>
          <div class="text-[10px] text-slate-400 dark:text-slate-500 truncate">
            {{ auth.user.dept_name }}
          </div>
        </div>
        <button
          v-if="!collapsed"
          class="shrink-0 p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-smooth"
          v-tooltip="'退出登录'"
          @click="
            auth.logout();
            router.push('/login');
          "
        >
          <AppIcon name="logout" :size="16" label="退出登录" interactive />
        </button>
      </div>
    </div>
  </aside>
</template>
