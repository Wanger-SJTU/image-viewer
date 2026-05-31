import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Asset } from '../types/asset'
import type { AssetFilter } from '../types/filter'
import { listAssets, getAsset, rateAsset, labelAsset, trashAsset, restoreAsset, purgeAsset, listTrashedAssets } from '../api/assets'

export const useAssetStore = defineStore('assets', () => {
  const assets = ref<Asset[]>([])
  const currentAsset = ref<Asset | null>(null)
  const filter = ref<AssetFilter>({})
  const loading = ref(false)
  const total = ref(0)

  // Trash state
  const trashedAssets = ref<Asset[]>([])
  const trashedTotal = ref(0)
  const showTrash = ref(false)

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
    const tIdx = trashedAssets.value.findIndex((a) => a.id === id)
    if (tIdx >= 0) {
      trashedAssets.value[tIdx].rating = rating
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
    const tIdx = trashedAssets.value.findIndex((a) => a.id === id)
    if (tIdx >= 0) {
      trashedAssets.value[tIdx].color_label = label as Asset['color_label']
    }
  }

  async function removeAsset(id: number) {
    await trashAsset(id)
    assets.value = assets.value.filter((a) => a.id !== id)
    total.value--
  }

  async function fetchTrashedAssets() {
    loading.value = true
    try {
      const result = await listTrashedAssets(1, 10000)
      trashedAssets.value = result.data
      trashedTotal.value = result.meta.total
    } finally {
      loading.value = false
    }
  }

  async function restoreTrashedAsset(id: number) {
    await restoreAsset(id)
    trashedAssets.value = trashedAssets.value.filter((a) => a.id !== id)
    trashedTotal.value--
    // Refresh main asset list to include restored asset
    await fetchAssets()
  }

  async function purgeTrashedAsset(id: number, fileType: 'both' | 'jpg' | 'raw') {
    await purgeAsset(id, fileType)
    trashedAssets.value = trashedAssets.value.filter((a) => a.id !== id)
    trashedTotal.value--
  }

  function toggleTrash() {
    showTrash.value = !showTrash.value
    if (showTrash.value) {
      fetchTrashedAssets()
    }
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
    trashedAssets,
    trashedTotal,
    showTrash,
    fetchAssets,
    fetchAsset,
    setRating,
    setLabel,
    removeAsset,
    fetchTrashedAssets,
    restoreTrashedAsset,
    purgeTrashedAsset,
    toggleTrash,
    updateFilter,
  }
})
