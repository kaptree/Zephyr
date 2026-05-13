# 任务02：核心基础设施

## 任务目标

搭建项目核心基础设施：TypeScript 类型系统、Axios 实例配置、Vue Router 路由骨架、Pinia Store 骨架。

## 依赖关系

- 依赖任务01（项目初始化）完成

## 技术要求

- 遵循文档第7.2节 API契约规范
- 遵循文档第4节路由设计
- 遵循文档第6节状态管理设计

## 具体步骤

### 2.1 TypeScript 类型定义系统 (`src/types/`)

#### `src/types/note.ts` — 任务类型

```typescript
interface Note {
  id: string;
  title: string;
  content: string;
  status: 'active' | 'completed' | 'archived';
  source_type: 'self' | 'assigned' | 'collaboration';
  priority: 'normal' | 'urgent';
  owner_id: string;
  creator_id: string;
  tags: Tag[];
  assignees: UserBrief[];
  group_id?: string;
  template_type?: string;
  due_time?: string;
  completed_at?: string;
  archived_at?: string;
  created_at: string;
  updated_at: string;
  allowed_actions: string[]; // 服务端返回的权限位
}

interface CreateNotePayload {
  title: string;
  content: string;
  tags: string[];
  source_type: 'self' | 'assigned' | 'collaboration';
  due_time?: string;
  owner_id?: string;
  template_type?: string;
  group_id?: string;
  assignees?: string[];
}
```

#### `src/types/user.ts` — 用户类型

```typescript
interface User {
  id: string;
  name: string;
  avatar: string;
  email: string;
  phone: string;
  dept_id: string;
  dept_name: string;
  role: 'super_admin' | 'dept_admin' | 'group_leader' | 'user' | 'screen_role';
  permissions: string[];
}

interface UserBrief {
  id: string;
  name: string;
  avatar: string;
  dept_name: string;
  role: string;
}

interface Department {
  id: string;
  name: string;
  parent_id: string | null;
  children?: Department[];
  member_count: number;
}
```

#### `src/types/tag.ts`

```typescript
interface Tag {
  id: string;
  name: string;
  color: string;
  scope: 'personal' | 'system';
  category: string;
  usage_count: number;
}
```

#### `src/types/api.ts` — API响应通用类型

```typescript
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface PaginatedData<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
}

interface TreeNode {
  id: string;
  label: string;
  children?: TreeNode[];
}
```

### 2.2 Axios 实例配置 (`src/services/api.ts`)

- BaseURL 从环境变量 `VITE_API_BASE_URL` 读取，默认 `/api`
- 请求拦截器：自动注入 Token（从 Pinia AuthStore 读取）
- 请求拦截器：自动注入 `X-Permission` 头
- 响应拦截器：统一错误处理（401跳转登录、403提示、500 Toast）
- 响应拦截器：数据解包（自动返回 `response.data`）
- 超时设置：30秒
- 导出类型安全的请求方法：`api.get<T>()`, `api.post<T>()`, `api.put<T>()`, `api.delete<T>()`

### 2.3 Vue Router 配置 (`src/router/index.ts`)

- 创建全部10个路由（详见文档第4.1节路由表）
- 路由懒加载（动态 import）
- 路由元信息：`meta: { title, requiresAuth, permissions }`
- 全局前置守卫：`beforeEach` 检查认证状态
- 全局后置守卫：`afterEach` 更新页面标题
- 滚动行为控制

### 2.4 Pinia Store 骨架 (`src/stores/`)

#### `useAuthStore`

- State：`user`, `token`, `permissions`
- Getters：`isLoggedIn`, `isAdmin`, `isDeptAdmin`, `canCreateForOthers`
- Actions：`login()`, `logout()`, `fetchUserInfo()`, `updatePermissions()`
- 持久化：token 存储至 localStorage

#### `useNoteStore`

- State：`activeNotes`, `archivedNotes`, `currentNote`, `loading`, `filters`
- Actions 骨架（暂不实现业务逻辑）

#### `useCollaborationStore`

- State：`roomId`, `participants`, `canvasData`, `syncStatus`, `socket`
- Actions 骨架（暂不实现业务逻辑）

### 2.5 环境变量配置 (`.env` / `.env.development`)

```
VITE_API_BASE_URL=http://localhost:8080/api
VITE_WS_URL=http://localhost:8080
VITE_APP_TITLE=轻燕工作台
```

## 验收标准

1. 所有 TypeScript 类型定义完整且无编译错误
2. Axios 实例可正常发起 HTTP 请求（Mock测试）
3. 路由10个路径全部可访问（页面组件为占位）
4. Pinia Store 可在组件中使用
5. 环境变量正确加载

## 预计工时：4小时

## 交付物

- `src/types/` 下全部类型定义文件
- `src/services/api.ts` Axios 实例
- `src/router/index.ts` 路由配置
- `src/stores/` 下三个 Store 文件
- `.env` / `.env.development` 环境变量文件
