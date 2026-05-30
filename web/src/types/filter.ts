export interface AssetFilter {
  rating?: number
  color_label?: string
  camera_model?: string
  match_status?: string
  file_type?: string
  focal_length_min?: number
  focal_length_max?: number
  aperture_min?: number
  aperture_max?: number
  iso_min?: number
  iso_max?: number
  captured_after?: string
  captured_before?: string
  search?: string
}
