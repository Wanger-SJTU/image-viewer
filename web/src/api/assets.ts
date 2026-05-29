import client, { unwrap, unwrapWithMeta } from './client'
import type { Asset } from '../types/asset'
import type { AssetFilter } from '../types/filter'
import type { RateRequest, LabelRequest } from '../types/api'

export async function listAssets(
  filter: AssetFilter,
  page = 1,
  limit = 50
): Promise<{ data: Asset[]; meta: { total: number; page: number; limit: number } }> {
  const params: Record<string, string | number> = { page, limit }
  if (filter.rating) params.rating = filter.rating
  if (filter.color_label) params.color_label = filter.color_label
  if (filter.camera_model) params.camera_model = filter.camera_model
  if (filter.match_status) params.match_status = filter.match_status
  if (filter.search) params.search = filter.search
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
