# 任务03：认证系统

## 任务目标
实现完整的认证系统：登录页面、AuthStore 业务逻辑、路由守卫、Token 管理。

## 依赖关系
- 依赖任务02（核心基础设施）完成

## 技术要求
- 遵循文档第4.1节路由设计（`/login`）
- 遵循文档第3节设计规范（极简白底，公安蓝点缀）
- 遵循文档第6.2节 AuthStore 设计
- 遵循文档第8节权限矩阵

## 具体步骤

### 3.1 登录页面 (`src/pages/LoginPage.vue`)

#### 视觉设计
- 全屏纯白底（`#FFFFFF`）
- 左侧：系统Logo + "资警数智·轻燕" 标题（slate-900, 24px, font-weight 600）
- 右侧：登录表单区域
- 交互蓝（`#3B82F6`）用于按钮和链接
- 公安徽章装饰元素

#### 功能实现
- **账密登录模式**：用户名 + 密码输入框，登录按钮
- **扫码登录模式**（预留UI，显示"暂未开放"）
- 表单验证：用户名/密码非空校验
- 加载状态：登录按钮 loading 动画
- 错误展示：登录失败时输入框下方红色提示文字
- 记住密码功能（7天有效期 localStorage）

#### 登录流程
1. 用户填写账密 → 点击"登录"
2. 调用 `POST /api/v1/auth/login` → 获取 token
3. 存储 token 至 localStorage
4. 调用 `GET /api/v1/auth/me` → 获取用户信息和权限
5. 更新 AuthStore 状态
6. 路由跳转至 `/workbench`

### 3.2 AuthStore 完善 (`src/stores/auth.ts`)

```typescript
// Actions 实现：
login(credentials: {username: string, password: string}): Promise<void>
logout(): void
fetchUserInfo(): Promise<void>
refreshToken(): Promise<void>
updatePermissions(permissions: string[]): void  // WebSocket推送时调用

// Getters 实现：
isLoggedIn: computed(() => !!token)
isAdmin: computed(() => role === 'super_admin')
isDeptAdmin: computed(() => role === 'dept_admin' || role === 'super_admin')
canCreateForOthers: computed(() => permissions.includes('can_assign'))
```

### 3.3 路由守卫 (`src/router/guards.ts`)

- **认证守卫**：未登录访问需认证页面 → 重定向到 `/login?redirect=原路径`
- **权限守卫**：已登录但无权限访问 → 跳转 403 页面
- **登录页守卫**：已登录访问 `/login` → 重定向到 `/workbench`
- Token 过期检测：401 响应时自动清除状态并跳转登录页

### 3.4 Token 管理

- Token 存储：localStorage（key: `auth_token`）
- Token 刷新：Axios 响应拦截器检测 401 → 尝试 refresh → 失败则登出
- 登出时清除所有本地数据

## 验收标准
1. 登录页面视觉符合设计规范（纯白底、公安蓝点缀）
2. 账密登录流程完整可用
3. 表单验证正确拦截无效输入
4. 登录成功后自动跳转工作台
5. 路由守卫正确拦截未登录/无权限访问
6. Token 过期自动跳转登录页
7. 登出功能清除所有状态

## 预计工时：3小时

## 交付物
- `src/pages/LoginPage.vue`
- `src/stores/auth.ts`（完善版）
- `src/router/guards.ts`
- `src/services/auth.ts`（认证相关API）
