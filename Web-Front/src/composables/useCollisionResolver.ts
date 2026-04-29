export interface ConflictResolution {
  content: string
  showWarning: boolean
  message: string
  conflict: boolean
}

export function useCollisionResolver() {
  function detectConflict(localVersion: number, remoteVersion: number): boolean {
    return localVersion !== remoteVersion
  }

  function resolveConflict(
    local: string,
    remote: string,
    remoteUser: string
  ): ConflictResolution {
    return {
      content: remote,
      showWarning: true,
      message: `${remoteUser} 刚刚修改了此处，内容已更新`,
      conflict: true,
    }
  }

  return {
    detectConflict,
    resolveConflict,
  }
}
