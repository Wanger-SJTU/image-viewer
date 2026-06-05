<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import type { Asset } from '../types/asset'
import { useAssetStore } from '../stores/assets'
import { useKeyboardShortcut } from '../composables/useKeyboardShortcut'
import { clearAssets, getAsset } from '../api/assets'
import FilterBar from '../components/FilterBar.vue'
import ImageGrid from '../components/ImageGrid.vue'
import ImagePreview from '../components/ImagePreview.vue'
import RatingStars from '../components/RatingStars.vue'
import ColorLabel from '../components/ColorLabel.vue'
import ScanDialog from '../components/ScanDialog.vue'
import TrashPanel from '../components/TrashPanel.vue'
import { useI18n } from '../i18n'

const { t, toggleLocale, localeLabel } = useI18n()

const assetStore = useAssetStore()
const previewAsset = ref<Asset | null>(null)
const previewIndex = ref(-1)
const scanDialog = ref<InstanceType<typeof ScanDialog>>()
const viewMode = ref<'preview' | 'review'>('preview')
const reviewIndex = ref(0)
const reviewScale = ref(1)
const reviewRotation = ref(0)
const reviewFitScreen = ref(true)
const reviewPanX = ref(0)
const reviewPanY = ref(0)
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const panStartX = ref(0)
const panStartY = ref(0)
const reviewFileType = ref<'jpg' | 'raw'>('jpg')

const hasBothFiles = computed(() => {
  const a = reviewAsset.value
  return !!(a?.jpg_file && a?.raw_file)
})

function toggleFileType() {
  reviewFileType.value = reviewFileType.value === 'jpg' ? 'raw' : 'jpg'
}

function reviewZoomIn() {
  reviewFitScreen.value = false
  reviewScale.value = Math.min(reviewScale.value * 1.25, 8)
}

function reviewZoomOut() {
  reviewFitScreen.value = false
  reviewScale.value = Math.max(reviewScale.value / 1.25, 0.1)
}

function reviewRotate() {
  reviewRotation.value = (reviewRotation.value + 90) % 360
}

function resetPan() {
  reviewPanX.value = 0
  reviewPanY.value = 0
}

function reviewReset() {
  reviewScale.value = 1
  reviewRotation.value = 0
  reviewFitScreen.value = true
  resetPan()
}

function reviewToggleFit() {
  reviewFitScreen.value = !reviewFitScreen.value
  if (reviewFitScreen.value) {
    reviewScale.value = 1
    resetPan()
  }
}

function onReviewWheel(e: WheelEvent) {
  e.preventDefault()
  if (e.deltaY < 0) reviewZoomIn()
  else reviewZoomOut()
}

function onReviewMouseDown(e: MouseEvent) {
  if (reviewScale.value <= 1) return
  isDragging.value = true
  dragStartX.value = e.clientX
  dragStartY.value = e.clientY
  panStartX.value = reviewPanX.value
  panStartY.value = reviewPanY.value
}

function onReviewMouseMove(e: MouseEvent) {
  if (!isDragging.value) return
  reviewPanX.value = panStartX.value + (e.clientX - dragStartX.value)
  reviewPanY.value = panStartY.value + (e.clientY - dragStartY.value)
}

function onReviewMouseUp() {
  isDragging.value = false
}

// Reset pan when switching images and load full asset detail
watch(reviewIndex, () => {
  reviewScale.value = 1
  reviewRotation.value = 0
  reviewFitScreen.value = true
  resetPan()
  reviewFileType.value = 'jpg'
  loadReviewAsset()
})

// Load full asset when entering review mode
watch(viewMode, (mode) => {
  if (mode === 'review') {
    if (!assetStore.assets.length) {
      viewMode.value = 'preview'
      return
    }
    if (reviewIndex.value >= assetStore.assets.length) {
      reviewIndex.value = 0
    }
    resetPan()
    reviewFileType.value = 'jpg'
    loadReviewAsset()
  }
})

// Reset review index when assets change (e.g., filter applied)
watch(() => assetStore.assets, () => {
  if (viewMode.value === 'review') {
    if (!assetStore.assets.length) {
      viewMode.value = 'preview'
      return
    }
    if (reviewIndex.value >= assetStore.assets.length) {
      reviewIndex.value = 0
      resetPan()
      reviewFileType.value = 'jpg'
      loadReviewAsset()
    }
  }
})

const reviewAssetBasic = computed(() => assetStore.assets[reviewIndex.value] || null)

// Full asset detail (with EXIF) loaded on demand for review mode
const reviewAssetFull = ref<Asset | null>(null)

async function loadReviewAsset() {
  const basic = reviewAssetBasic.value
  if (!basic) {
    reviewAssetFull.value = null
    return
  }
  try {
    reviewAssetFull.value = await getAsset(basic.id)
  } catch {
    reviewAssetFull.value = basic
  }
}

// Keep the full asset in sync with what's stored (rating/label changes)
function syncReviewAsset(id: number, rating?: number, label?: string) {
  if (reviewAssetFull.value?.id === id) {
    if (rating !== undefined) reviewAssetFull.value.rating = rating
    if (label !== undefined) reviewAssetFull.value.color_label = label as Asset['color_label']
  }
}

const reviewAsset = computed(() => reviewAssetFull.value || reviewAssetBasic.value)

const reviewExif = computed(() => {
  const mf = reviewAsset.value?.jpg_file || reviewAsset.value?.raw_file
  return mf?.exif || null
})

function imgStyle(): Record<string, string> {
  const totalAngle = reviewRotation.value
  const t: string[] = []
  // CSS applies right-to-left: rotate, then scale, then translate
  if (reviewPanX.value !== 0 || reviewPanY.value !== 0) {
    t.push(`translate(${reviewPanX.value}px, ${reviewPanY.value}px)`)
  }
  if (!reviewFitScreen.value) t.push(`scale(${reviewScale.value})`)
  if (totalAngle !== 0) t.push(`rotate(${totalAngle}deg)`)
  const cursor = isDragging.value ? 'grabbing' : reviewScale.value > 1 ? 'grab' : 'default'
  return {
    transform: t.length > 0 ? t.join(' ') : 'none',
    transformOrigin: 'center center',
    transition: isDragging.value ? 'none' : 'transform 0.15s',
    cursor,
  }
}

function reviewThumbUrl(): string {
  if (!reviewAsset.value) return ''
  const fileParam = hasBothFiles.value ? `&file=${reviewFileType.value}` : ''
  return `/api/v1/thumbs/${reviewAsset.value.id}?size=full${fileParam}`
}

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

function jumpToReview() {
  const asset = previewAsset.value
  if (!asset) return
  closePreview()
  openReview(asset)
}

function reviewPrev() {
  if (reviewIndex.value > 0) reviewIndex.value--
}

function reviewNext() {
  if (reviewIndex.value < assetStore.assets.length - 1) reviewIndex.value++
}

useKeyboardShortcut(() => ({
  '1': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 1); syncReviewAsset(a.id, 1) }
  },
  '2': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 2); syncReviewAsset(a.id, 2) }
  },
  '3': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 3); syncReviewAsset(a.id, 3) }
  },
  '4': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 4); syncReviewAsset(a.id, 4) }
  },
  '5': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 5); syncReviewAsset(a.id, 5) }
  },
  '0': () => {
    const a = previewAsset.value || reviewAssetFull.value
    if (a) { assetStore.setRating(a.id, 0); syncReviewAsset(a.id, 0) }
  },
  x: () => {
    if (viewMode.value === 'review') {
      const a = reviewAssetFull.value
      if (a) {
        const newLabel = a.color_label === 'red' ? '' : 'red'
        assetStore.setLabel(a.id, newLabel)
        syncReviewAsset(a.id, undefined, newLabel)
      }
    }
  },
  f: () => {
    if (viewMode.value === 'review') reviewToggleFit()
  },
  r: () => {
    if (viewMode.value === 'review') reviewRotate()
  },
  '+': () => {
    if (viewMode.value === 'review') reviewZoomIn()
  },
  '-': () => {
    if (viewMode.value === 'review') reviewZoomOut()
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
    const a = previewAsset.value || reviewAssetFull.value
    if (a && confirm(t('preview.delete_confirm'))) {
      if (previewAsset.value) {
        assetStore.removeAsset(previewAsset.value.id)
        closePreview()
      } else if (viewMode.value === 'review') {
        assetStore.removeAsset(a.id)
        // Move to next or exit review if last
        if (assetStore.assets.length > 1) {
          reviewNext()
        } else {
          viewMode.value = 'preview'
        }
      }
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

async function handleRestore(id: number) {
  await assetStore.restoreTrashedAsset(id)
}

async function handlePurge(id: number, fileType: 'both' | 'jpg' | 'raw') {
  await assetStore.purgeTrashedAsset(id, fileType)
}

function toggleTrash() {
  assetStore.toggleTrash()
  // Close preview when entering trash
  if (assetStore.showTrash) {
    closePreview()
  }
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
          <button class="trash-btn" @click="toggleTrash">
            {{ assetStore.showTrash ? t('toolbar.close_trash') : t('toolbar.trash', { n: assetStore.trashedTotal }) }}
          </button>
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
          <button
            class="mode-btn"
            :class="{ active: viewMode === 'review' }"
            :disabled="!assetStore.assets.length"
            @click="reviewIndex = 0; viewMode = 'review'"
          >
            {{ t('toolbar.review_mode') }}
          </button>
          <button class="refresh-btn" @click="assetStore.fetchAssets()">{{ t('toolbar.refresh') }}</button>
          <button class="lang-btn" @click="toggleLocale()">{{ localeLabel }}</button>
        </div>
      </div>

      <div class="gallery-content">
        <TrashPanel
          v-if="assetStore.showTrash"
          :assets="assetStore.trashedAssets"
          :total="assetStore.trashedTotal"
          @close="toggleTrash"
          @restore="handleRestore"
          @purge="handlePurge"
        />
        <ImageGrid
          v-else-if="viewMode === 'preview'"
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
            <div class="review-top-bar">
              <span class="review-filename">{{ reviewAsset?.name }}</span>
              <span class="review-index">{{ reviewIndex + 1 }} / {{ assetStore.assets.length }}</span>
              <div class="review-controls">
                <button title="Zoom out" @click="reviewZoomOut">&minus;</button>
                <span class="zoom-pct">{{ Math.round(reviewScale * 100) }}%</span>
                <button title="Zoom in" @click="reviewZoomIn">+</button>
                <button title="Rotate" @click="reviewRotate">&#8635;</button>
                <button :title="reviewFitScreen ? '1:1' : 'Fit'" @click="reviewToggleFit">
                  {{ reviewFitScreen ? '1:1' : 'Fit' }}
                </button>
                <button title="Reset" @click="reviewReset">Reset</button>
                <button
                  v-if="hasBothFiles"
                  :title="reviewFileType === 'jpg' ? 'Show RAW' : 'Show JPG'"
                  class="file-type-btn"
                  @click="toggleFileType"
                >
                  {{ reviewFileType === 'jpg' ? 'RAW' : 'JPG' }}
                </button>
              </div>
            </div>
            <div
              class="review-image-wrap"
              @wheel="onReviewWheel"
              @mousedown="onReviewMouseDown"
              @mousemove="onReviewMouseMove"
              @mouseup="onReviewMouseUp"
              @mouseleave="onReviewMouseUp"
            >
              <img
                :key="reviewAsset?.id"
                :src="reviewThumbUrl()"
                :alt="reviewAsset?.name"
                :style="imgStyle()"
                class="review-img"
                draggable="false"
              />
            </div>
            <div class="review-info-bar">
              <div class="review-info-left">
                <span v-if="reviewExif?.camera_model" class="exif-tag">{{ reviewExif.camera_model }}</span>
                <span v-if="reviewExif?.focal_length" class="exif-tag">{{ reviewExif.focal_length }}mm</span>
                <span v-if="reviewExif?.aperture" class="exif-tag">f/{{ reviewExif.aperture }}</span>
                <span v-if="reviewExif?.iso" class="exif-tag">ISO {{ reviewExif.iso }}</span>
                <span v-if="reviewExif?.shutter_speed" class="exif-tag">{{ reviewExif.shutter_speed }}</span>
                <span v-if="reviewExif?.captured_at" class="exif-tag">{{ reviewExif.captured_at }}</span>
              </div>
              <div class="review-info-right">
                <RatingStars v-if="reviewAsset" :rating="reviewAsset.rating" @rate="(r: number) => { assetStore.setRating(reviewAsset!.id, r); syncReviewAsset(reviewAsset!.id, r) }" />
                <ColorLabel v-if="reviewAsset" :color-label="reviewAsset.color_label" @select="(l: string) => { assetStore.setLabel(reviewAsset!.id, l); syncReviewAsset(reviewAsset!.id, undefined, l) }" />
              </div>
            </div>
            <div class="review-key-hints">
              <span>1-5 Rate</span>
              <span>0 Zero</span>
              <span>X Reject</span>
              <span>R Rotate</span>
              <span>+/- Zoom</span>
              <span>F Fit</span>
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
      @openInReview="jumpToReview()"
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

.trash-btn {
  padding: 4px 16px;
  background: #1a1a2e;
  color: #ccc;
  border: 1px solid #0f3460;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}

.trash-btn:hover {
  border-color: #e94560;
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

.mode-btn:disabled {
  opacity: 0.35;
  cursor: default;
}

.mode-btn:hover:not(.active):not(:disabled) {
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
  align-items: stretch;
  height: 100%;
  gap: 0;
}

.review-nav {
  flex-shrink: 0;
  width: 60px;
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
  overflow: hidden;
}

/* Top bar with filename + controls */
.review-top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 16px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
  flex-shrink: 0;
}

.review-top-bar .review-filename {
  color: #ccc;
  font-size: 0.85rem;
  font-weight: 600;
}

.review-top-bar .review-index {
  color: #666;
  font-size: 0.8rem;
}

.review-controls {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}

.review-controls button {
  padding: 2px 8px;
  background: #1a1a2e;
  color: #999;
  border: 1px solid #0f3460;
  border-radius: 3px;
  font-size: 0.75rem;
  cursor: pointer;
  min-width: 28px;
}

.review-controls button:hover {
  border-color: #e94560;
  color: #ccc;
}

.zoom-pct {
  color: #888;
  font-size: 0.7rem;
  min-width: 36px;
  text-align: center;
}

.file-type-btn {
  color: #e94560 !important;
  border-color: #e94560 !important;
}

/* Image area */
.review-image-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 8px;
  min-height: 0;
}

.review-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

/* Info bar: EXIF + rating */
.review-info-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #16213e;
  border-top: 1px solid #0f3460;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
}

.review-info-left {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.exif-tag {
  color: #999;
  font-size: 0.78rem;
  padding: 2px 6px;
  background: #1a1a2e;
  border-radius: 3px;
  white-space: nowrap;
}

.review-info-right {
  display: flex;
  gap: 16px;
  align-items: center;
}

/* Keyboard hints */
.review-key-hints {
  display: flex;
  gap: 12px;
  padding: 4px 16px;
  background: #0d1b2a;
  border-top: 1px solid #0f3460;
  flex-shrink: 0;
  justify-content: center;
}

.review-key-hints span {
  color: #444;
  font-size: 0.65rem;
  text-transform: uppercase;
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
