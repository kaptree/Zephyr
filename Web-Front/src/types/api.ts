export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PaginatedData<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface TreeNode {
  id: string
  label: string
  children?: TreeNode[]
}
