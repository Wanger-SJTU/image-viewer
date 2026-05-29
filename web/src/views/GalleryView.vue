<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Asset } from '../types/asset'
import { useAssetStore } from '../stores/assets'
import { useKeyboardShortcut } from '../composables/useKeyboardShortcut'
import FilterBar from '../components/FilterBar.vue'
import ImageGrid from '../components/ImageGrid.vue'
import ImagePreview from '../components/ImagePreview.vue'
import ScanDialog from '../components/ScanDialog.vue'

const assetStore = useAssetStore()
const previewAsset = ref<Asset | null>(null)
const previewIndex = ref(-1)
const scanDialog = ref<InstanceType<typeof ScanDialog>>()

onMounted(() => {
  assetStore.fetchAssets()
})

function openPreview(asset: Asset) {
  previewIndex.value = assetStore.assets.findIndex((a) => a.id === asset.id)
  previewAsset.value = asset
}

function closePreview() {
  previewAsset.value = null
  previewIndex.value = -1
}

function goPrev() {
  if (previewIndex.value > 0) {
    previewIndex.value--
    previewAsset.value = assetStore.assets[previewIndex.value]
  }
}

function goNext() {
  if (previewIndex.value < assetStore.assets.length - 1) {
    previewIndex.value++
    previewAsset.value = assetStore.assets[previewIndex.value]
  }
}

useKeyboardShortcut(() => ({
  '1': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 1),
  '2': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 2),
  '3': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 3),
  '4': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 4),
  '5': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 5),
  '0': () => previewAsset.value && assetStore.setRating(previewAsset.value.id, 0),
  ArrowLeft: goPrev,
  ArrowRight: goNext,
  Escape: closePreview,
  Delete: () => {
    if (previewAsset.value && confirm('Delete this asset?')) {
      assetStore.removeAsset(previewAsset.value.id)
      closePreview()
    }
  },
}))

function openScan() {
  scanDialog.value?.open()
}
</script>

<template>
  <div class="gallery">
    <div class="gallery-toolbar">
      <div class="toolbar-left">
        <button class="scan-btn" @click="openScan">Scan</button>
        <span class="count-info" v-if="assetStore.total">
          {{ assetStore.total }} assets
        </span>
        <span class="count-info" v-else>
          No assets. Click Scan to import photos.
        </span>
      </div>
      <div class="toolbar-right">
        <span v-if="assetStore.totalPages > 1" class="page-info">
          Page {{ assetStore.page }} / {{ assetStore.totalPages }}
        </span>
        <div class="page-btns">
          <button
            class="page-btn"
            :disabled="assetStore.page <= 1"
            @click="assetStore.setPage(assetStore.page - 1)"
          >
            Prev
          </button>
          <button
            class="page-btn"
            :disabled="assetStore.page >= assetStore.totalPages"
            @click="assetStore.setPage(assetStore.page + 1)"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <FilterBar
      :filter="assetStore.filter"
      @update="assetStore.updateFilter($event)"
      @clear="assetStore.updateFilter({})"
    />

    <div class="gallery-content">
      <ImageGrid
        :assets="assetStore.assets"
        @select="openPreview"
        @rate="(id, r) => assetStore.setRating(id, r)"
        @label="(id, l) => assetStore.setLabel(id, l)"
      />
    </div>

    <ImagePreview
      :asset="previewAsset"
      @close="closePreview"
      @prev="goPrev"
      @next="goNext"
      @rate="(id, r) => assetStore.setRating(id, r)"
      @label="(id, l) => assetStore.setLabel(id, l)"
    />

    <ScanDialog ref="scanDialog" />
  </div>
</template>

<style scoped>
.gallery {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.gallery-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.scan-btn {
  padding: 4px 16px;
  background: #e94560;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}

.scan-btn:hover {
  background: #c73e54;
}

.count-info {
  font-size: 0.85rem;
  color: #999;
}

.page-info {
  font-size: 0.8rem;
  color: #888;
}

.page-btns {
  display: flex;
  gap: 4px;
}

.page-btn {
  padding: 3px 10px;
  border: 1px solid #0f3460;
  border-radius: 3px;
  background: #1a1a2e;
  color: #ccc;
  font-size: 0.8rem;
  cursor: pointer;
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: default;
}

.page-btn:hover:not(:disabled) {
  border-color: #e94560;
}

.gallery-content {
  flex: 1;
  overflow-y: auto;
}
</style>
