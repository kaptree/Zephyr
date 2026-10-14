import { get, post, put, del } from './api'

/** 密码本条目（本人数据，password 为解密后的明文） */
export interface VaultEntry {
  id: string
  user_id: string
  platform: string
  website: string
  account: string
  password: string
  notes: string
  shared_at: string | null
  created_at: string
  updated_at: string
}

/** 密码广场条目（不含密码，密码需单独 reveal） */
export interface PlazaEntry {
  id: string
  platform: string
  website: string
  account: string
  notes: string
  owner_id: string
  shared_by_name: string
  shared_at: string
}

export interface VaultEntryPayload {
  platform: string
  website: string
  account: string
  password: string
  notes: string
}

/** 我的密码本列表 */
export function fetchVaultEntries() {
  return get<{ list: VaultEntry[] }>('/api/v1/vault/entries')
}

/** 新增条目 */
export function createVaultEntry(payload: VaultEntryPayload) {
  return post<VaultEntry>('/api/v1/vault/entries', payload)
}

/** 更新条目（password 留空表示保持原密码不变） */
export function updateVaultEntry(id: string, payload: VaultEntryPayload) {
  return put<VaultEntry>(`/api/v1/vault/entries/${id}`, payload)
}

/** 删除条目 */
export function deleteVaultEntry(id: string) {
  return del<null>(`/api/v1/vault/entries/${id}`)
}

/** 共享 / 取消共享到密码广场 */
export function setVaultEntryShared(id: string, shared: boolean) {
  return put<VaultEntry>(`/api/v1/vault/entries/${id}/share`, { shared })
}

/** 密码广场列表 */
export function fetchVaultPlaza() {
  return get<{ list: PlazaEntry[] }>('/api/v1/vault/plaza')
}

/** 查看广场条目密码 */
export function revealPlazaPassword(id: string) {
  return get<{ password: string }>(`/api/v1/vault/plaza/${id}/reveal`)
}
