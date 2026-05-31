import client, { unwrap, unwrapWithMeta } from './client'
import type { Asset, PurgeRequest } from '../types/asset'
import type { AssetFilter } from '../types/filter'
import type { RateRequest, LabelRequest, FilterOptions } from '../types/api'

export async function listAssets(
  filter: AssetFilter,
  page = 1,
  limit = 50
): Promise<{ data: Asset[]; meta: { total: number; page: number; limit: number } }> {
  const params: Record<string, string | number | boolean> = { page, limit }
  if (filter.rating) params.rating = filter.rating
  if (filter.color_label) params.color_label = filter.color_label
  if (filter.camera_model) params.camera_model = filter.camera_model
  if (filter.match_status) params.match_status = filter.match_status
  if (filter.file_type) params.file_type = filter.file_type
  if (filter.search) params.search = filter.search
  if (filter.focal_length_min) params.focal_length_min = filter.focal_length_min
  if (filter.focal_length_max) params.focal_length_max = filter.focal_length_max
  if (filter.aperture_min) params.aperture_min = filter.aperture_min
  if (filter.aperture_max) params.aperture_max = filter.aperture_max
  if (filter.iso_min) params.iso_min = filter.iso_min
  if (filter.iso_max) params.iso_max = filter.iso_max
  if (filter.captured_after) {
    params.captured_after = filter.captured_after.includes('T')
      ? filter.captured_after
      : filter.captured_after + 'T00:00:00Z'
  }
  if (filter.captured_before) {
    params.captured_before = filter.captured_before.includes('T')
      ? filter.captured_before
      : filter.captured_before + 'T00:00:00Z'
  }
  if (filter.trashed !== undefined) params.trashed = filter.trashed
  const resp = await client.get('/assets', { params })
  return unwrapWithMeta<Asset[]>(resp)
}

export async function getAsset(id: number): Promise<Asset> {
  const resp = await client.get(`/assets/${id}`)
  return unwrap<Asset>(resp)
}

export async function rateAsset(id: number, rating: number): Promise<void> {
  const body: RateRequest = { rating }
  await client.post(`/assets/${id}/rate`, body)
}

export async function labelAsset(id: number, colorLabel: string): Promise<void> {
  const body: LabelRequest = { color_label: colorLabel }
  await client.post(`/assets/${id}/label`, body)
}

export async function deleteAsset(id: number): Promise<void> {
  await client.delete(`/assets/${id}`)
}

export async function getFilterOptions(): Promise<FilterOptions> {
  const resp = await client.get('/filters')
  return unwrap<FilterOptions>(resp)
}

export async function clearAssets(): Promise<number> {
  const resp = await client.delete('/assets')
  return unwrap<{ deleted: number }>(resp).deleted
}

export async function trashAsset(id: number): Promise<void> {
  await client.post(`/assets/${id}/trash`)
}

export async function restoreAsset(id: number): Promise<void> {
  await client.post(`/assets/${id}/restore`)
}

export async function purgeAsset(id: number, fileType: 'both' | 'jpg' | 'raw'): Promise<void> {
  const body: PurgeRequest = { file_type: fileType }
  await client.post(`/assets/${id}/purge`, body)
}

export async function listTrashedAssets(page = 1, limit = 50): Promise<{ data: Asset[]; meta: { total: number; page: number; limit: number } }> {
  const params: Record<string, string | number | boolean> = { page, limit, trashed: true }
  const resp = await client.get('/assets', { params })
  return unwrapWithMeta<Asset[]>(resp)
}
