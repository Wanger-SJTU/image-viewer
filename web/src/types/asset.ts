// Mirrors shared/types/asset.go — source of truth for frontend

export type MediaType = 'raw' | 'jpg'
export type MatchStatus = 'paired' | 'orphan'
export type ColorLabel = '' | 'red' | 'orange' | 'yellow' | 'green' | 'blue' | 'purple'

export interface ExifMeta {
  camera_model: string
  lens_model?: string
  focal_length?: number
  aperture?: number
  shutter_speed?: string
  iso?: number
  captured_at?: string
  width: number
  height: number
  orientation: number
}

export interface MediaFile {
  id: number
  asset_id: number
  file_path: string
  file_name: string
  file_size: number
  media_type: MediaType
  exif?: ExifMeta
  created_at: string
}

export interface Asset {
  id: number
  name: string
  dir_path: string
  raw_file?: MediaFile
  jpg_file?: MediaFile
  match_status: MatchStatus
  rating: number
  color_label: ColorLabel
  ai_status: string
  captured_at?: string
  deleted_at?: string
  grid_thumb: string
  full_thumb: string
  created_at: string
  updated_at: string
}

export interface PurgeRequest {
  file_type: 'both' | 'jpg' | 'raw'
}
