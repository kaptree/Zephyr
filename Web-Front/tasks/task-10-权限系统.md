# 任务10：权限系统

## 任务目标

实现完整的权限控制系统：v-permission 自定义指令、路由守卫、按钮级权限控制、数据级权限隔离。

## 依赖关系

- 依赖任务03（认证系统）完成

## 技术要求

- 遵循文档第8节权限矩阵
- 4个角色：super_admin、dept_admin、group_leader、user、screen_role
- 三层面控制：路由级、按钮级、数据级

## 具体步骤

### 10.1 权限数据模型扩展

在 `src/types/user.ts` 中定义：

```typescript
type Role = 'super_admin' | 'dept_admin' | 'group_leader' | 'user' | 'screen_role';

type Permission =
  | 'create_note_self' // 给自己创建任务
  | 'create_note_assigned' // 给他人创建任务（can_assign）
  | 'edit_others_note' // 编辑他人任务
  | 'delete_note' // 删除任务
  | 'remind' // 盯办操作
  | 'view_all_archive' // 查看全部归档
  | 'view_dept_archive' // 查看本部门归档
  | 'view_group_archive' // 查看本组归档
  | 'manage_departments' // 管理部门
  | 'manage_users' // 管理用户
  | 'manage_tags' // 管理标签
  | 'manage_templates' // 管理模板
  | 'access_screen' // 进入大屏端
  | 'send_command'; // 下发指令（协同中）
```

这些权限位由服务端在登录时返回，存储于 AuthStore。

### 10.2 路由权限守卫完善

在 `src/router/guards.ts` 中：

```typescript
// 路由元信息类型扩展
declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    requiresAuth?: boolean;
    permissions?: Permission[]; // 所需权限位列表
    roles?: Role[]; // 所需角色列表
  }
}

// 权限校验逻辑
function checkRoutePermission(
  route: RouteLocationNormalized,
  permissions: Permission[],
  role: Role
): boolean {
  // 1. 检查角色要求
  if (route.meta.roles && !route.meta.roles.includes(role)) {
    return false;
  }
  // 2. 检查权限位要求
  if (route.meta.permissions) {
    return route.meta.permissions.every((p) => permissions.includes(p));
  }
  return true;
}
```

#### 各路由权限配置

| 路由                           | permissions              | roles                  |
| ------------------------------ | ------------------------ | ---------------------- |
| `/admin/departments`           | `['manage_departments']` | -                      |
| `/admin/users`                 | `['manage_users']`       | -                      |
| `/admin/tags`                  | `['manage_tags']`        | -                      |
| `/admin/templates`             | `['manage_templates']`   | -                      |
| `/screen/:id`                  | `['access_screen']`      | -                      |
| `/workbench/collaboration/:id` | -                        | 动态检查是否为协同成员 |

### 10.3 v-permission 自定义指令

在 `src/directives/permission.ts` 中：

```typescript
// 用法：
// <button v-permission="['remind']">盯办</button>
// <button v-permission:disable="['create_note_assigned']">指派</button>  // 无权限时禁用（灰色）
// <button v-permission:hide="['manage_templates']">管理</button>  // 无权限时隐藏（默认）

const permissionDirective: Directive = {
  mounted(el, binding) {
    const { value, arg } = binding;
    const permissions = useAuthStore().permissions;
    const hasPermission = value.every((p: Permission) => permissions.includes(p));

    if (!hasPermission) {
      if (arg === 'disable') {
        el.setAttribute('disabled', 'true');
        el.classList.add('opacity-50', 'cursor-not-allowed');
        // 可选：添加 Tooltip 提示
      } else {
        // 默认 hide
        el.style.display = 'none';
      }
    }
  },
};
```

### 10.4 数据级权限（allowed_actions）

- 服务端在每条任务数据中返回 `allowed_actions` 字段
- 前端根据此字段动态渲染操作按钮
- StickyNoteCard 组件接收 `permissions` prop 或直接读取 note.allowed_actions
- 例如：某任务 `allowed_actions: ['edit', 'view']` → 仅显示编辑和查看按钮，不显示删除/盯办

```typescript
// 在 StickyNoteCard 中：
const canEdit = computed(() => props.note.allowed_actions?.includes('edit'));
const canRemind = computed(() => props.note.allowed_actions?.includes('remind'));
const canDelete = computed(() => props.note.allowed_actions?.includes('delete'));
const canComplete = computed(() => props.note.allowed_actions?.includes('complete'));
```

### 10.5 侧边栏菜单权限过滤

Sidebar 组件中根据 AuthStore.permissions 过滤菜单项：

- 无权限的菜单项不显示
- 或显示但灰色+tooltip"无权限"

### 10.6 权限动态更新

- WebSocket 推送权限变更消息
- AuthStore.updatePermissions() 实时更新
- 前端无需刷新，即时生效
- 权限降级时自动跳转到工作台

## 验收标准

1. v-permission 指令正确控制按钮显隐/禁用
2. 路由守卫正确拦截无权限访问
3. allowed_actions 控制操作按钮正确
4. 侧边栏菜单正确按权限过滤
5. 权限变更后即时生效

## 预计工时：3小时

## 交付物

- `src/directives/permission.ts`
- `src/router/guards.ts`（完善版）
- 完善 `src/types/user.ts` 权限类型
- 完善 Sidebar 组件权限控制
