# 任务01：项目初始化

## 任务目标

使用 Vite 搭建 Vue 3 + TypeScript 前端项目脚手架，集成 Tailwind CSS、DaisyUI 及所有必要依赖。

## 依赖关系

- 无前置依赖，为所有后续任务的基础

## 技术要求

1. **构建工具**：Vite 5.x
2. **框架**：Vue 3.4+ (Composition API + `<script setup>`)
3. **语言**：TypeScript 5.x（严格模式）
4. **样式**：Tailwind CSS 3.x + DaisyUI 4.x
5. **字体**：引入 Inter 和 Noto Sans SC 字体
6. **代码规范**：ESLint + Prettier

## 具体步骤

### 1.1 创建 Vite 项目

```bash
npm create vite@latest . -- --template vue-ts
```

### 1.2 安装核心依赖

```bash
npm install vue-router@4 pinia axios socket.io-client
```

### 1.3 安装样式依赖

```bash
npm install -D tailwindcss @tailwindcss/typography daisyui postcss autoprefixer
```

### 1.4 安装开发工具

```bash
npm install -D @types/node eslint prettier eslint-plugin-vue @vue/eslint-config-typescript
```

### 1.5 配置 Tailwind CSS

- 创建 `tailwind.config.js`，配置 DaisyUI 插件
- 扩展颜色系统（slate色阶、任务语义色、交互蓝）
- 配置字体族（Inter + Noto Sans SC）
- 配置自定义阴影（任务卡片、弹窗、脉冲）

### 1.6 配置 DaisyUI

- 主题设置为 "light" 模式
- 禁用不需要的组件以减少体积

### 1.7 项目目录结构

```
src/
├── assets/          # 静态资源（字体、图标、图片）
├── components/      # 可复用组件
│   ├── common/      # 通用组件（按钮、输入框、模态框等）
│   ├── note/        # 任务相关组件
│   └── layout/      # 布局组件
├── composables/     # 组合式函数
├── directives/      # 自定义指令（v-permission等）
├── layouts/         # 布局模板
├── pages/           # 页面组件
├── router/          # 路由配置
├── services/        # API服务层
├── stores/          # Pinia状态管理
├── types/           # TypeScript类型定义
├── utils/           # 工具函数
├── App.vue
└── main.ts
```

### 1.8 配置全局 CSS

- 引入 Tailwind 指令（@tailwind base/components/utilities）
- 设置全局字体、基础样式
- 定义 CSS 变量（颜色、圆角、阴影）
- 定义关键帧动画（盯办脉冲、归档动画、弹簧动画）

## 验收标准

1. `npm run dev` 能成功启动开发服务器
2. Tailwind CSS 样式正常生效
3. DaisyUI 组件可正常使用
4. TypeScript 编译无错误
5. 自定义颜色（任务黄/绿/红、交互蓝）在 Tailwind 中可用
6. 全局字体正常加载并渲染

## 预计工时：2小时

## 交付物

- 完整的 Vite + Vue 3 + TS 项目结构
- Tailwind/DaisyUI 配置文件
- 全局 CSS 样式文件
- 项目根目录 `package.json` 及所有依赖
