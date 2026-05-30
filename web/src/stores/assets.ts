import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Asset } from '../types/asset'
import type { AssetFilter } from '../types/filter'
import { listAssets, getAsset, rateAsset, labelAsset, deleteAsset } from '../api/assets'

export const useAssetStore = defineStore('assets', () => {
  const assets = ref<Asset[]>([])
  const currentAsset = ref<Asset | null>(null)
  const filter = ref<AssetFilter>({})
  const loading = ref(false)
  const total = ref(0)

  async function fetchAssets() {
    loading.value = true
    try {
      const result = await listAssets(filter.value, 1, 10000)
      assets.value = result.data
      total.value = result.meta.total
    } finally {
      loading.value = false
    }
  }

  async function fetchAsset(id: number) {
    currentAsset.value = await getAsset(id)
  }

  async function setRating(id: number, rating: number) {
    await rateAsset(id, rating)
    if (currentAsset.value?.id === id) {
      currentAsset.value.rating = rating
    }
    const idx = assets.value.findIndex((a) => a.id === id)
    if (idx >= 0) {
      assets.value[idx].rating = rating
    }
  }

  async function setLabel(id: number, label: string) {
    await labelAsset(id, label)
    if (currentAsset.value?.id === id) {
      currentAsset.value.color_label = label as Asset['color_label']
    }
    const idx = assets.value.findIndex((a) => a.id === id)
    if (idx >= 0) {
      assets.value[idx].color_label = label as Asset['color_label']
    }
  }

  async function removeAsset(id: number) {
    await deleteAsset(id)
    assets.value = assets.value.filter((a) => a.id !== id)
    total.value--
  }

  function updateFilter(newFilter: Partial<AssetFilter>) {
    filter.value = { ...filter.value, ...newFilter }
    fetchAssets()
  }

  return {
    assets,
    currentAsset,
    filter,
    loading,
    total,
    fetchAssets,
    fetchAsset,
    setRating,
    setLabel,
    removeAsset,
    updateFilter,
  }
})
