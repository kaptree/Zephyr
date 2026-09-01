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
