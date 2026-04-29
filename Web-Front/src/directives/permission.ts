import type { Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'

type PermissionMode = 'hide' | 'disable'

export const vPermission: Directive<HTMLElement, string[]> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string[]>) {
    const required = binding.value
    if (!required || required.length === 0) return

    const mode: PermissionMode = (binding.arg as PermissionMode) || 'hide'
    const auth = useAuthStore()
    
    const hasPermission = required.every(perm => auth.permissions.includes(perm as never))

    if (!hasPermission) {
      if (mode === 'disable') {
        el.setAttribute('disabled', 'true')
        el.classList.add('opacity-50', 'cursor-not-allowed', 'pointer-events-none')
        if (!el.getAttribute('title')) {
          el.setAttribute('title', '暂无操作权限')
        }
      } else {
        el.style.display = 'none'
        el.classList.add('permission-hidden')
      }
    }
  },

  updated(el: HTMLElement, binding: DirectiveBinding<string[]>) {
    const required = binding.value
    if (!required || required.length === 0) {
      el.style.display = ''
      el.classList.remove('permission-hidden')
      return
    }

    const mode: PermissionMode = (binding.arg as PermissionMode) || 'hide'
    const auth = useAuthStore()
    const hasPermission = required.every(perm => auth.permissions.includes(perm as never))

    if (!hasPermission) {
      if (mode === 'disable') {
        el.setAttribute('disabled', 'true')
        el.classList.add('opacity-50', 'cursor-not-allowed', 'pointer-events-none')
      } else {
        el.style.display = 'none'
        el.classList.add('permission-hidden')
      }
    } else {
      el.style.display = ''
      el.classList.remove('permission-hidden', 'opacity-50', 'cursor-not-allowed', 'pointer-events-none')
      el.removeAttribute('disabled')
    }
  },
}
