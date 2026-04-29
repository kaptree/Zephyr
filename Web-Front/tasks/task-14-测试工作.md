# 任务14：测试工作

## 任务目标
为整个前端应用建立完整的测试体系，包括单元测试、集成测试和端到端测试，确保所有功能模块质量。

## 依赖关系
- 贯穿所有开发任务的后续验证工作

## 测试框架
- **单元测试**：Vitest + @vue/test-utils
- **组件测试**：Vitest + @vue/test-utils + jsdom
- **E2E测试**：Playwright (可选) 或 Vitest Browser Mode
- **覆盖率**：c8 / istanbul

## 具体步骤

### 14.1 测试环境搭建

```bash
npm install -D vitest @vue/test-utils jsdom @vitejs/plugin-vue
npm install -D @vitest/coverage-c8 happy-dom
```

配置 `vitest.config.ts`：
```typescript
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
    coverage: {
      provider: 'c8',
      reporter: ['text', 'json', 'html'],
    },
  },
})
```

### 14.2 测试目录结构

```
tests/
├── unit/                    # 单元测试
│   ├── stores/              # Pinia Store 测试
│   │   ├── auth.test.ts
│   │   ├── notes.test.ts
│   │   └── collaboration.test.ts
│   ├── services/            # API 服务测试
│   │   ├── api.test.ts
│   │   └── notes.test.ts
│   ├── composables/         # 组合式函数测试
│   │   ├── useOffline.test.ts
│   │   └── useVirtualScroll.test.ts
│   ├── directives/          # 自定义指令测试
│   │   └── permission.test.ts
│   └── utils/               # 工具函数测试
│       └── errorHandler.test.ts
├── components/              # 组件测试
│   ├── StickyNoteCard.test.ts
│   ├── TagSelector.test.ts
│   ├── UserPicker.test.ts
│   ├── Sidebar.test.ts
│   └── ...
├── integration/             # 集成测试
│   ├── login-flow.test.ts
│   ├── note-lifecycle.test.ts
│   ├── collaboration.test.ts
│   └── archive-search.test.ts
├── e2e/                     # 端到端测试
│   ├── workbench.spec.ts
│   └── admin.spec.ts
├── setup.ts                 # 测试全局设置
└── mocks/                   # Mock 数据
    ├── handlers.ts          # MSW handlers
    ├── data.ts              # Mock 数据工厂
    └── server.ts            # MSW server 配置
```

### 14.3 单元测试用例

#### 14.3.1 AuthStore 测试 (`tests/unit/stores/auth.test.ts`)

```typescript
describe('useAuthStore', () => {
  it('初始状态应为未登录', () => {})
  it('login 应正确更新 token 和 user', () => {})
  it('login 失败应保持未登录状态', () => {})
  it('logout 应清除所有状态和 localStorage', () => {})
  it('isAdmin getter 应正确判断角色', () => {})
  it('isDeptAdmin getter 应正确判断角色', () => {})
  it('canCreateForOthers 应正确判断权限', () => {})
  it('updatePermissions 应正确更新权限数组', () => {})
})
```

#### 14.3.2 NoteStore 测试 (`tests/unit/stores/notes.test.ts`)

```typescript
describe('useNoteStore', () => {
  describe('fetchNotes', () => {
    it('应正确获取活跃便签列表', () => {})
    it('应支持按状态筛选', () => {})
    it('应支持分页', () => {})
    it('加载失败应设置错误状态', () => {})
  })
  describe('createNote', () => {
    it('创建成功后应插入列表顶部', () => {})
    it('创建失败应回滚本地状态', () => {})
  })
  describe('updateNote', () => {
    it('应乐观更新本地状态', () => {})
    it('接口失败应回滚本地状态', () => {})
  })
  describe('completeNote', () => {
    it('完成后应从活跃列表移除', () => {})
  })
  describe('remindNote', () => {
    it('盯办后便签状态应变红', () => {})
  })
})
```

#### 14.3.3 API 服务测试 (`tests/unit/services/api.test.ts`)

```typescript
describe('Axios API', () => {
  it('请求拦截器应自动注入 Token', () => {})
  it('响应拦截器应处理 401 登出', () => {})
  it('响应拦截器应处理 403 提示', () => {})
  it('响应拦截器应处理网络超时', () => {})
})
```

#### 14.3.4 Composable 测试

```typescript
describe('useOffline', () => {
  it('离线时应将操作加入队列', () => {})
  it('恢复在线时应批量同步挂起操作', () => {})
  it('同步失败的操作应保留在队列中', () => {})
})

describe('useVirtualScroll', () => {
  it('应正确计算可见区域项', () => {})
  it('滚动时应更新可见项', () => {})
})
```

### 14.4 组件测试

#### 14.4.1 StickyNoteCard 测试 (`tests/components/StickyNoteCard.test.ts`)

```typescript
describe('StickyNoteCard', () => {
  it('待办便签应显示黄色样式', () => {})
  it('已完成便签应显示绿色样式和角标', () => {})
  it('盯办便签应显示红色样式和徽章', () => {})
  it('单击应触发 click 事件', () => {})
  it('右键应弹出上下文菜单', () => {})
  it('内容超出3行应显示渐变遮罩', () => {})
  it('标签超过2个应显示 +N', () => {})
  it('无编辑权限应隐藏编辑按钮', () => {})
})
```

#### 14.4.2 TagSelector 测试

```typescript
describe('TagSelector', () => {
  it('应正确渲染标签列表', () => {})
  it('多选功能正常', () => {})
  it('搜索过滤功能正常', () => {})
  it('应支持新建标签', () => {})
  it('最近使用标签应置顶', () => {})
})
```

#### 14.4.3 UserPicker 测试

```typescript
describe('UserPicker', () => {
  it('应正确渲染部门树', () => {})
  it('展开部门应显示人员列表', () => {})
  it('搜索功能应正确过滤', () => {})
  it('已选人员应在胶囊中展示', () => {})
  it('组长应显示特殊角标', () => {})
})
```

#### 14.4.4 Sidebar 测试

```typescript
describe('Sidebar', () => {
  it('应正确渲染所有菜单项', () => {})
  it('折叠/展开功能正常', () => {})
  it('折叠状态应持久化到 localStorage', () => {})
  it('无权限菜单项应隐藏', () => {})
  it('当前路由对应的菜单项应高亮', () => {})
})
```

### 14.5 集成测试

#### 14.5.1 登录流程测试 (`tests/integration/login-flow.test.ts`)

```typescript
describe('登录流程集成测试', () => {
  it('正确账密登录 → 跳转工作台', () => {})
  it('错误账密登录 → 显示错误提示', () => {})
  it('未登录访问工作台 → 重定向登录页', () => {})
  it('登录过期 → 自动登出', () => {})
  it('无权限访问 admin → 403', () => {})
})
```

#### 14.5.2 便签生命周期测试 (`tests/integration/note-lifecycle.test.ts`)

```typescript
describe('便签完整生命周期', () => {
  it('创建 → 编辑 → 完成归档 完整流程', () => {})
  it('创建 → 指派他人 → 被指派人完成', () => {})
  it('创建 → 盯办提醒 → 变红 → 完成归档', () => {})
})
```

#### 14.5.3 归档查询测试

```typescript
describe('归档查询集成测试', () => {
  it('筛选 → 查询 → 视图切换 → 导出 完整流程', () => {})
  it('恢复归档 → 出现在工作台', () => {})
})
```

### 14.6 E2E 测试

使用 Playwright（如已安装）或 Vitest 浏览器模式：

```typescript
// tests/e2e/workbench.spec.ts
describe('工作台 E2E', () => {
  it('页面加载应显示便签墙', () => {})
  it('新建便签完整流程', () => {})
  it('点击便签打开详情面板', () => {})
  it('拖拽排序便签', () => {})
  it('完成便签触发归档动画', () => {})
})
```

### 14.7 Mock 基础设施

使用 MSW (Mock Service Worker) 来模拟 API：

```typescript
// tests/mocks/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  // 便签列表
  http.get('/api/v1/notes', () => {
    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: { data: mockNotes, total: mockNotes.length, page: 1, page_size: 20 }
    })
  }),
  
  // 创建便签
  http.post('/api/v1/notes', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({
      code: 0,
      data: { ...body, id: 'new-id' }
    })
  }),
  
  // ... 更多 handlers
]
```

### 14.8 测试报告

在 `package.json` 中添加测试命令：
```json
{
  "scripts": {
    "test": "vitest",
    "test:unit": "vitest --dir tests/unit",
    "test:components": "vitest --dir tests/components",
    "test:integration": "vitest --dir tests/integration",
    "test:coverage": "vitest --coverage",
    "test:ui": "vitest --ui"
  }
}
```

生成 HTML 测试报告到 `tests/reports/` 目录。

## 验收标准
1. 单元测试覆盖所有 Store 和 Composable
2. 组件测试覆盖5个核心组件
3. 集成测试覆盖登录流程和便签生命周期
4. 测试用例总数 ≥ 50 个
5. 所有测试用例通过
6. 代码行覆盖率 ≥ 70%
7. 测试报告可正常生成

## 预计工时：8小时

## 交付物
- `tests/` 目录完整测试结构
- 所有单元测试文件
- 所有组件测试文件
- 集成测试文件
- Mock 基础设施
- 测试配置文件
- 测试报告
