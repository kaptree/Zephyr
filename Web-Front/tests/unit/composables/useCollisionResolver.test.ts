import { describe, it, expect } from 'vitest'
import { useCollisionResolver } from '@/composables/useCollisionResolver'

describe('useCollisionResolver', () => {
  it('detectConflict 在同一版本应返回 false', () => {
    const { detectConflict } = useCollisionResolver()
    expect(detectConflict(5, 5)).toBe(false)
  })

  it('detectConflict 在不同版本应返回 true', () => {
    const { detectConflict } = useCollisionResolver()
    expect(detectConflict(5, 6)).toBe(true)
  })

  it('resolveConflict 应以远程版本覆盖本地', () => {
    const { resolveConflict } = useCollisionResolver()
    const result = resolveConflict('本地内容', '远程内容', '张三')
    expect(result.content).toBe('远程内容')
    expect(result.showWarning).toBe(true)
    expect(result.conflict).toBe(true)
    expect(result.message).toContain('张三')
  })
})
