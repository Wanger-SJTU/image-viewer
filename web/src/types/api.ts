export interface PaginationMeta {
  total: number
  page: number
  limit: number
}

export interface APIResponse<T = unknown> {
  success: boolean
  data: T
  error?: string
  meta?: PaginationMeta
}

export interface ScanRequest {
  path: string
}

export interface RateRequest {
  rating: number
}

export interface LabelRequest {
  color_label: string
}
