<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Asset } from '../types/asset'
import { useAssetStore } from '../stores/assets'
import { useKeyboardShortcut } from '../composables/useKeyboardShortcut'
import { clearAssets, getAsset } from '../api/assets'
import FilterBar from '../components/FilterBar.vue'
import ImageGrid from '../components/ImageGrid.vue'
import ImagePreview from '../components/ImagePreview.vue'
import ScanDialog from '../components/ScanDialog.vue'
import { useI18n } from '../i18n'

const { t, toggleLocale, localeLabel } = useI18n()

const assetStore = useAssetStore()
const previewAsset = ref<Asset | null>(null)
const previewIndex = ref(-1)
const scanDialog = ref<InstanceType<typeof ScanDialog>>()
const viewMode = ref<'preview' | 'review'>('preview')
const reviewIndex = ref(0)

onMounted(() => {
  assetStore.fetchAssets()
})

async function openPreview(asset: Asset) {
  previewIndex.value = assetStore.assets.findIndex((a) => a.id === asset.id)
  try {
    previewAsset.value = await getAsset(asset.id)
  } catch {
    previewAsset.value = asset
  }
}

function closePreview() {
  previewAsset.value = null
  previewIndex.value = -1
}

async function goPrev() {
  if (previewIndex.value > 0) {
    previewIndex.value--
    try {
      previewAsset.value = await getAsset(assetStore.assets[previewIndex.value].id)
    } catch {
      previewAsset.value = assetStore.assets[previewIndex.value]
    }
  }
}

async function goNext() {
  if (previewIndex.value < assetStore.assets.length - 1) {
    previewIndex.value++
    try {
      previewAsset.value = await getAsset(assetStore.assets[previewIndex.value].id)
    } catch {
      previewAsset.value = assetStore.assets[previewIndex.value]
    }
  }
}

function openReview(asset: Asset) {
  reviewIndex.value = assetStore.assets.findIndex((a) => a.id === asset.id)
  viewMode.value = 'review'
}

function reviewPrev() {
  if (reviewIndex.value > 0) reviewIndex.value--
}

function reviewNext() {
  if (reviewIndex.value < assetStore.assets.length - 1) reviewIndex.value++
}

useKeyboardShortcut(() => ({
  '1': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 1)
  },
  '2': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 2)
  },
  '3': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 3)
  },
  '4': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 4)
  },
  '5': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 5)
  },
  '0': () => {
    const a = previewAsset.value || assetStore.assets[reviewIndex.value]
    if (a) assetStore.setRating(a.id, 0)
  },
  ArrowLeft: () => {
    if (viewMode.value === 'review') reviewPrev()
    else goPrev()
  },
  ArrowRight: () => {
    if (viewMode.value === 'review') reviewNext()
    else goNext()
  },
  Escape: closePreview,
  Delete: () => {
    if (previewAsset.value && confirm(t('preview.delete_confirm'))) {
      assetStore.removeAsset(previewAsset.value.id)
      closePreview()
    }
  },
}))

async function clearAll() {
  if (!confirm(t('toolbar.clear_all_confirm'))) return
  try {
    const count = await clearAssets()
    alert(`Deleted ${count} assets`)
    assetStore.fetchAssets()
  } catch (e: any) {
    alert('Clear failed: ' + e.message)
  }
}

function openScan() {
  scanDialog.value?.open()
}
</script>

<template>
  <div class="gallery">
    <FilterBar
      :filter="assetStore.filter"
      @update="assetStore.updateFilter($event)"
      @clear="assetStore.updateFilter({})"
    />

    <div class="gallery-main">
      <div class="gallery-toolbar">
        <div class="toolbar-left">
          <button class="scan-btn" @click="openScan">{{ t('toolbar.scan') }}</button>
          <button class="clear-btn" @click="clearAll">{{ t('toolbar.clear_all') }}</button>
          <span class="count-info" v-if="assetStore.total">
            {{ t('toolbar.assets', { n: assetStore.total }) }}
          </span>
          <span class="count-info" v-else>
            {{ t('toolbar.no_assets') }}
          </span>
        </div>
        <div class="toolbar-right">
          <button class="mode-btn" :class="{ active: viewMode === 'preview' }" @click="viewMode = 'preview'">
            {{ t('toolbar.preview_mode') }}
          </button>
          <button class="mode-btn" :class="{ active: viewMode === 'review' }" @click="viewMode = 'review'">
            {{ t('toolbar.review_mode') }}
          </button>
          <button class="refresh-btn" @click="assetStore.fetchAssets()">{{ t('toolbar.refresh') }}</button>
          <button class="lang-btn" @click="toggleLocale()">{{ localeLabel }}</button>
        </div>
      </div>

      <div class="gallery-content">
        <ImageGrid
          v-if="viewMode === 'preview'"
          :assets="assetStore.assets"
          @select="openPreview"
          @rate="(id, r) => assetStore.setRating(id, r)"
          @label="(id, l) => assetStore.setLabel(id, l)"
        />
        <div v-else-if="viewMode === 'review' && assetStore.assets.length" class="review-mode">
          <div class="review-nav review-prev" @click="reviewPrev">
            <span>&lsaquo;</span>
          </div>
          <div class="review-main">
            <img
              :key="assetStore.assets[reviewIndex]?.id"
              :src="'/api/v1/thumbs/' + assetStore.assets[reviewIndex]?.id + '?size=full'"
              :alt="assetStore.assets[reviewIndex]?.name"
            />
            <div class="review-info-bar">
              <span class="review-filename">{{ assetStore.assets[reviewIndex]?.name }}</span>
              <span class="review-index">{{ reviewIndex + 1 }} / {{ assetStore.assets.length }}</span>
            </div>
          </div>
          <div class="review-nav review-next" @click="reviewNext">
            <span>&rsaquo;</span>
          </div>
        </div>
        <div v-else class="review-empty">
          <p>{{ t('toolbar.no_assets') }}</p>
        </div>
      </div>
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
  flex-direction: row;
  height: 100%;
  overflow: hidden;
}

.gallery-main {
  flex: 1;
  display: flex;
  flex-direction: column;
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

.clear-btn {
  padding: 4px 16px;
  background: #555;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}

.clear-btn:hover {
  background: #777;
}

.refresh-btn {
  padding: 4px 16px;
  background: #1a1a2e;
  color: #ccc;
  border: 1px solid #0f3460;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}

.refresh-btn:hover {
  border-color: #e94560;
}

.count-info {
  font-size: 0.85rem;
  color: #999;
}

.gallery-content {
  flex: 1;
  overflow-y: auto;
}

/* Mode toggle buttons */
.mode-btn {
  padding: 4px 12px;
  background: #1a1a2e;
  color: #888;
  border: 1px solid #0f3460;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
}

.mode-btn.active {
  color: #e94560;
  border-color: #e94560;
}

.mode-btn:hover:not(.active) {
  color: #ccc;
}

/* Language toggle */
.lang-btn {
  padding: 4px 12px;
  background: #1a1a2e;
  color: #ccc;
  border: 1px solid #0f3460;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
}

.lang-btn:hover {
  border-color: #e94560;
}

/* Review mode */
.review-mode {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 0;
}

.review-nav {
  flex-shrink: 0;
  width: 60px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #555;
  font-size: 3rem;
  transition: color 0.2s, background 0.2s;
  user-select: none;
}

.review-nav:hover {
  color: #e94560;
  background: rgba(233, 69, 96, 0.05);
}

.review-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-height: 100%;
  padding: 16px;
}

.review-main img {
  max-width: 100%;
  max-height: calc(100vh - 120px);
  object-fit: contain;
  border-radius: 4px;
}

.review-info-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  max-width: 800px;
  margin-top: 12px;
  padding: 8px 16px;
  background: #16213e;
  border-radius: 4px;
}

.review-filename {
  color: #ccc;
  font-size: 0.9rem;
}

.review-index {
  color: #888;
  font-size: 0.85rem;
}

.review-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #666;
  font-size: 1rem;
}
</style>
