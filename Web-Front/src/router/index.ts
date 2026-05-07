import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/LoginPage.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/',
    redirect: '/workbench',
  },
  {
    path: '/workbench',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { title: '工作台', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Workbench',
        component: () => import('@/pages/WorkbenchPage.vue'),
        meta: { title: '工作台' },
      },
      {
        path: 'archive',
        name: 'Archive',
        component: () => import('@/pages/ArchivePage.vue'),
        meta: { title: '归档查询' },
      },
      {
        path: 'collaboration/:id',
        name: 'Collaboration',
        component: () => import('@/pages/CollaborationPage.vue'),
        meta: { title: '协同编辑室', permissions: [] },
      },
      {
        path: 'groups/:id',
        name: 'WorkGroupDetail',
        component: () => import('@/pages/GroupDetailPage.vue'),
        meta: { title: '专项行动详情' },
      },
      {
        path: 'groups/:id/dashboard',
        name: 'WorkGroupDashboard',
        component: () => import('@/pages/GroupDashboardPage.vue'),
        meta: { title: '数据大屏' },
      },
    ],
  },
  {
    path: '/analytics',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Analytics',
        component: () => import('@/pages/AnalyticsPage.vue'),
        meta: { title: '工作成效分析' },
      },
    ],
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'departments',
        name: 'Departments',
        component: () => import('@/pages/admin/DepartmentsPage.vue'),
        meta: { title: '部门管理', permissions: ['manage_departments'] },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/pages/admin/UsersPage.vue'),
        meta: { title: '人员管理', permissions: ['manage_users'] },
      },
      {
        path: 'tags',
        name: 'Tags',
        component: () => import('@/pages/admin/TagsPage.vue'),
        meta: { title: '标签管理', permissions: ['manage_tags'] },
      },
      {
        path: 'templates',
        name: 'Templates',
        component: () => import('@/pages/admin/TemplatesPage.vue'),
        meta: { title: '模板管理', permissions: ['manage_templates'] },
      },
      {
        path: 'system',
        name: 'SystemSettings',
        component: () => import('@/pages/admin/SystemSettingsPage.vue'),
        meta: { title: '系统管理', permissions: ['manage_system'] },
      },
      {
        path: 'operation-log',
        name: 'OperationLog',
        component: () => import('@/pages/admin/OperationLogPage.vue'),
        meta: { title: '操作日志', permissions: ['manage_system'] },
      },
    ],
  },
  {
    path: '/screen/:id',
    name: 'Screen',
    component: () => import('@/pages/ScreenPage.vue'),
    meta: { title: '数据大屏', permissions: ['access_screen'] },
  },
  {
    path: '/profile',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Profile',
        component: () => import('@/pages/ProfilePage.vue'),
        meta: { title: '个人中心' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/pages/ForbiddenPage.vue'),
    meta: { title: '无权访问' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/pages/NotFoundPage.vue'),
    meta: { title: '页面不存在' },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || '轻燕工作台'} - 轻燕`;

  if (to.meta.public) {
    const token = localStorage.getItem('auth_token');
    if (token && to.path === '/login') {
      next('/workbench');
      return;
    }
    next();
    return;
  }

  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      next(`/login?redirect=${encodeURIComponent(to.fullPath)}`);
      return;
    }

    const userStr = localStorage.getItem('auth_user');
    if (userStr && to.meta.permissions && (to.meta.permissions as string[]).length > 0) {
      try {
        const user = JSON.parse(userStr);
        if (user.role === 'super_admin') {
          next();
          return;
        }
        const userPerms: string[] = user.permissions || [];
        const requiredPerms = to.meta.permissions as string[];
        const hasPermission = requiredPerms.every((p: string) => userPerms.includes(p));
        if (!hasPermission) {
          next('/403');
          return;
        }
      } catch {
        // ignore parse errors
      }
    }
  }

  next();
});

router.afterEach(() => {
  // 滚动到顶部
  window.scrollTo(0, 0);
});

export default router;
