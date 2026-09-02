/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'note-yellow': {
          bg: '#FEF3C7',
          border: '#D97706',
        },
        'note-green': {
          bg: '#DCFCE7',
          border: '#16A34A',
        },
        'note-red': {
          bg: '#FEE2E2',
          border: '#DC2626',
        },
        'interactive': '#3B82F6',
        'surface': {
          DEFAULT: '#FFFFFF',
          alt: '#F8FAFC',
        },
      },
      fontFamily: {
        sans: ['"Inter"', '"PingFang SC"', '"Microsoft YaHei"', '"Noto Sans CJK SC"', '-apple-system', 'sans-serif'],
      },
      borderRadius: {
        'card': '16px',
        'btn': '10px',
        'tag': '6px',
      },
      boxShadow: {
        'note': '0 4px 24px -4px rgba(0,0,0,0.08)',
        'note-hover': '0 8px 32px -8px rgba(0,0,0,0.12)',
        'modal': '0 24px 48px -12px rgba(0,0,0,0.18)',
        'btn-float': '0 4px 12px rgba(59,130,246,0.4)',
        'note-red-pulse': '0 0 0 4px rgba(220,38,38,0.2)',
      },
      keyframes: {
        'pulse-alert': {
          '0%, 100%': { transform: 'scale(1)', boxShadow: '0 0 0 4px rgba(220,38,38,0.2)' },
          '50%': { transform: 'scale(1.02)', boxShadow: '0 0 0 8px rgba(220,38,38,0.1)' },
        },
        'spring-enter': {
          '0%': { transform: 'scale(0.8)', opacity: '0' },
          '60%': { transform: 'scale(1.03)', opacity: '1' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        'archiving': {
          '0%': { opacity: '1', transform: 'scale(1) translateY(0)' },
          '100%': { opacity: '0', transform: 'scale(0.95) translateY(20px)' },
        },
        'skeleton-pulse': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.4' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-out-right': {
          '0%': { transform: 'translateX(0)' },
          '100%': { transform: 'translateX(100%)' },
        },
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        // ===== 表格"呼吸式数据流"（美化工程 · 表格模块化） =====
        // 行入场：左侧滑动 + 淡入，背景从 primary/20 渐变为透明（300ms）
        'table-row-enter': {
          '0%': { opacity: '0', transform: 'translateX(-24px)', backgroundColor: 'rgba(59, 130, 246, 0.2)' },
          '100%': { opacity: '1', transform: 'translateX(0)', backgroundColor: 'transparent' },
        },
        // 状态突变：脉冲光晕闪烁 600ms（黄/绿/红 对应 待办/完成/盯办 语义）
        'status-pulse-yellow': {
          '0%': { boxShadow: '0 0 0 0 rgba(217, 119, 6, 0.45)' },
          '70%': { boxShadow: '0 0 0 8px rgba(217, 119, 6, 0)' },
          '100%': { boxShadow: '0 0 0 0 rgba(217, 119, 6, 0)' },
        },
        'status-pulse-green': {
          '0%': { boxShadow: '0 0 0 0 rgba(22, 163, 74, 0.45)' },
          '70%': { boxShadow: '0 0 0 8px rgba(22, 163, 74, 0)' },
          '100%': { boxShadow: '0 0 0 0 rgba(22, 163, 74, 0)' },
        },
        'status-pulse-red': {
          '0%': { boxShadow: '0 0 0 0 rgba(220, 38, 38, 0.45)' },
          '70%': { boxShadow: '0 0 0 8px rgba(220, 38, 38, 0)' },
          '100%': { boxShadow: '0 0 0 0 rgba(220, 38, 38, 0)' },
        },
        // 空状态插画：2s 周期缓慢呼吸脉动
        breathe: {
          '0%, 100%': { transform: 'scale(1)', opacity: '0.6' },
          '50%': { transform: 'scale(1.08)', opacity: '1' },
        },
        // ===== 微交互语言（美化工程 · 第 2 项） =====
        // 状态标签切换：300ms 翻转过渡（:key 变化时重放，正面旧状态翻入新状态）
        'badge-flip': {
          '0%': { transform: 'perspective(600px) rotateX(-80deg)', opacity: '0' },
          '100%': { transform: 'perspective(600px) rotateX(0deg)', opacity: '1' },
        },
        // ===== 沉浸式书写（美化工程 · 第 6 项） =====
        // 校验通过：对勾图标"弹出 + 旋转"出现（300ms）
        'check-pop': {
          '0%': { transform: 'scale(0) rotate(-90deg)', opacity: '0' },
          '60%': { transform: 'scale(1.2) rotate(8deg)', opacity: '1' },
          '100%': { transform: 'scale(1) rotate(0deg)', opacity: '1' },
        },
        // 校验失败：输入框"水平抖动"（400ms，振幅 5px）
        'shake-x': {
          '0%, 100%': { transform: 'translateX(0)' },
          '15%': { transform: 'translateX(-5px)' },
          '30%': { transform: 'translateX(5px)' },
          '45%': { transform: 'translateX(-5px)' },
          '60%': { transform: 'translateX(5px)' },
          '75%': { transform: 'translateX(-3px)' },
          '90%': { transform: 'translateX(3px)' },
        },
        // 提交按钮空闲："呼吸光泽"持续微动，吸引点击
        'btn-shimmer': {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        // 提交成功：绿色"缩小 → 正常"弹动反馈（500ms）
        'btn-success-pop': {
          '0%': { transform: 'scale(0.85)' },
          '55%': { transform: 'scale(1.06)' },
          '80%': { transform: 'scale(0.98)' },
          '100%': { transform: 'scale(1)' },
        },
        // 提交失败：红色"抖动 2 次"（500ms 两轮），随后恢复可提交
        'btn-fail-shake': {
          '0%, 100%': { transform: 'translateX(0)' },
          '12%': { transform: 'translateX(-6px)' },
          '24%': { transform: 'translateX(6px)' },
          '36%': { transform: 'translateX(-6px)' },
          '48%': { transform: 'translateX(6px)' },
          '64%': { transform: 'translateX(-4px)' },
          '80%': { transform: 'translateX(4px)' },
          '92%': { transform: 'translateX(-1px)' },
        },
        // 日期选中：脉冲光晕闪烁一次（500ms）
        'date-pulse': {
          '0%': { boxShadow: '0 0 0 0 rgba(59, 130, 246, 0.5)' },
          '70%': { boxShadow: '0 0 0 8px rgba(59, 130, 246, 0)' },
          '100%': { boxShadow: '0 0 0 0 rgba(59, 130, 246, 0)' },
        },
        // Wizard 当前步骤指示器：脉动光晕（2s 循环）
        'step-dot-pulse': {
          '0%, 100%': { boxShadow: '0 0 0 0 rgba(59, 130, 246, 0.4)' },
          '50%': { boxShadow: '0 0 0 6px rgba(59, 130, 246, 0)' },
        },
        // ===== 通知中心「生命力衰减」交互体系（美化工程 · 通知生命周期） =====
        // 呼吸脉冲圆点：2s 周期无限脉动（opacity 0.4↔1.0，scale 0.9↔1.0），颜色随优先级
        'notif-dot-pulse': {
          '0%, 100%': { opacity: '0.4', transform: 'scale(0.9)' },
          '50%': { opacity: '1', transform: 'scale(1)' },
        },
        // 时间戳流光：高光从左向右扫过（2.4s 周期，配合 background-clip: text）
        'notif-shimmer': {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        // 铃铛晃动：振幅 8°（600ms），新消息到达时提示
        'bell-shake': {
          '0%, 100%': { transform: 'rotate(0deg)' },
          '15%': { transform: 'rotate(8deg)' },
          '30%': { transform: 'rotate(-8deg)' },
          '45%': { transform: 'rotate(6deg)' },
          '60%': { transform: 'rotate(-6deg)' },
          '75%': { transform: 'rotate(3deg)' },
          '90%': { transform: 'rotate(-3deg)' },
        },
        // 角标弹跳放大：未读 +1 时弹性反馈（spring 缓动）
        'badge-pop': {
          '0%': { transform: 'scale(1)' },
          '40%': { transform: 'scale(1.35)' },
          '70%': { transform: 'scale(0.92)' },
          '100%': { transform: 'scale(1)' },
        },
        // 脉冲圆点谢幕：放大 → 消散（scale 1.0→1.5→0，潮汐批量动画）
        'notif-dot-burst': {
          '0%': { opacity: '1', transform: 'scale(1)' },
          '40%': { opacity: '1', transform: 'scale(1.5)' },
          '100%': { opacity: '0', transform: 'scale(0)' },
        },
        // 单条已读：脉冲圆点缩小淡出（400ms）
        'notif-dot-out': {
          '0%': { opacity: '1', transform: 'scale(1)' },
          '100%': { opacity: '0', transform: 'scale(0.3)' },
        },
        // 「已读」对勾：spring 弹出 + 轻微旋转 + 上浮 2px 回弹（400ms）
        'notif-check-pop': {
          '0%': { opacity: '0', transform: 'translateY(4px) scale(0.4) rotate(-30deg)' },
          '55%': { opacity: '1', transform: 'translateY(-2px) scale(1.15) rotate(6deg)' },
          '80%': { transform: 'translateY(0) scale(0.96) rotate(-2deg)' },
          '100%': { opacity: '1', transform: 'translateY(0) scale(1) rotate(0deg)' },
        },
        // ===== 有温度的图标语言体系 =====
        // 初次加载：缩放淡入依次点亮（300ms，配 --icon-stagger 延迟）
        'icon-enter': {
          '0%': { opacity: '0', transform: 'scale(0.55)' },
          '65%': { opacity: '1', transform: 'scale(1.08)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        // 待办：空心圆/时钟缓慢脉冲
        'icon-clock-pulse': {
          '0%, 100%': { opacity: '0.72', transform: 'scale(1)' },
          '50%': { opacity: '1', transform: 'scale(1.1)' },
        },
        // 进行中：齿轮匀速旋转
        'icon-spin-slow': {
          '0%': { transform: 'rotate(0deg)' },
          '100%': { transform: 'rotate(360deg)' },
        },
        // 盯办/高优：星标金色闪烁（光晕呼吸）
        'icon-star-twinkle': {
          '0%, 100%': { opacity: '0.55', filter: 'drop-shadow(0 0 0 rgba(245, 158, 11, 0))' },
          '50%': { opacity: '1', filter: 'drop-shadow(0 0 5px rgba(245, 158, 11, 0.75))' },
        },
        // 警告三角：快速闪烁 3 次后停止（不 infinite）
        'icon-flash-warn': {
          '0%, 100%': { opacity: '1' },
          '25%': { opacity: '0.15' },
          '50%': { opacity: '1' },
          '75%': { opacity: '0.35' },
        },
        // 已完成：绘制勾号（stroke-dashoffset 32 → 0，400ms）
        'icon-check-draw': {
          '0%': { strokeDashoffset: '32' },
          '100%': { strokeDashoffset: '0' },
        },
        // 品牌 Logo：模糊到清晰浮现（600ms）
        'icon-logo-reveal': {
          '0%': { opacity: '0', filter: 'blur(8px)', transform: 'scale(0.9)' },
          '100%': { opacity: '1', filter: 'blur(0)', transform: 'scale(1)' },
        },
      },
      animation: {
        'pulse-alert': 'pulse-alert 2s ease-in-out infinite',
        'spring-enter': 'spring-enter 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards',
        'archiving': 'archiving 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards',
        'skeleton': 'skeleton-pulse 1.5s ease-in-out infinite',
        'slide-in-right': 'slide-in-right 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards',
        'slide-out-right': 'slide-out-right 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards',
        'fade-in': 'fade-in 0.2s ease-out forwards',
        // ===== 表格"呼吸式数据流" =====
        'table-row-enter': 'table-row-enter 0.3s cubic-bezier(0.4, 0, 0.2, 1) both',
        'status-pulse-yellow': 'status-pulse-yellow 0.6s ease-out',
        'status-pulse-green': 'status-pulse-green 0.6s ease-out',
        'status-pulse-red': 'status-pulse-red 0.6s ease-out',
        'breathe': 'breathe 2s ease-in-out infinite',
        // ===== 微交互语言 =====
        'badge-flip': 'badge-flip 0.3s cubic-bezier(0.4, 0, 0.2, 1) both',
        // ===== 沉浸式书写 =====
        'check-pop': 'check-pop 0.3s cubic-bezier(0.4, 0, 0.2, 1) both',
        'shake-x': 'shake-x 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
        'btn-shimmer': 'btn-shimmer 2.5s linear infinite',
        'btn-success-pop': 'btn-success-pop 0.5s cubic-bezier(0.4, 0, 0.2, 1)',
        'btn-fail-shake': 'btn-fail-shake 0.5s cubic-bezier(0.4, 0, 0.2, 1)',
        'date-pulse': 'date-pulse 0.5s ease-out',
        'step-dot-pulse': 'step-dot-pulse 2s ease-in-out infinite',
        // ===== 通知中心「生命力衰减」 =====
        'notif-dot-pulse': 'notif-dot-pulse 2s ease-in-out infinite',
        'notif-shimmer': 'notif-shimmer 2.4s linear infinite',
        'bell-shake': 'bell-shake 0.6s cubic-bezier(0.4, 0, 0.2, 1)',
        'badge-pop': 'badge-pop 0.45s cubic-bezier(0.34, 1.56, 0.64, 1)',
        'notif-dot-burst': 'notif-dot-burst 0.45s ease-out forwards',
        'notif-dot-out': 'notif-dot-out 0.4s ease-in forwards',
        'notif-check-pop': 'notif-check-pop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both',
        // ===== 有温度的图标语言体系 =====
        'icon-enter': 'icon-enter 0.3s cubic-bezier(0.34, 1.56, 0.64, 1) both',
        'icon-clock-pulse': 'icon-clock-pulse 2.4s ease-in-out infinite',
        'icon-spin-slow': 'icon-spin-slow 2.6s linear infinite',
        'icon-star-twinkle': 'icon-star-twinkle 1.8s ease-in-out infinite',
        'icon-flash-warn': 'icon-flash-warn 0.9s ease-in-out 3',
        'icon-check-draw': 'icon-check-draw 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards',
        'icon-logo-reveal': 'icon-logo-reveal 0.6s cubic-bezier(0.4, 0, 0.2, 1) both',
      },
      transitionTimingFunction: {
        'smooth': 'cubic-bezier(0.4, 0, 0.2, 1)',
      },
    },
  },
  plugins: [
    require('daisyui'),
    require('@tailwindcss/typography'),
  ],
  daisyui: {
    themes: [
      {
        light: {
          "primary": "#3B82F6",
          "secondary": "#64748B",
          "accent": "#8B5CF6",
          "neutral": "#0F172A",
          "base-100": "#FFFFFF",
          "base-200": "#F8FAFC",
          "base-300": "#E2E8F0",
          "info": "#3B82F6",
          "success": "#16A34A",
          "warning": "#D97706",
          "error": "#DC2626",
        },
        dark: {
          "primary": "#60A5FA",
          "secondary": "#94A3B8",
          "accent": "#A78BFA",
          "neutral": "#F1F5F9",
          "base-100": "#1E293B",
          "base-200": "#0F172A",
          "base-300": "#334155",
          "info": "#60A5FA",
          "success": "#4ADE80",
          "warning": "#FBBF24",
          "error": "#F87171",
        },
      },
    ],
    darkTheme: 'dark',
  },
}
