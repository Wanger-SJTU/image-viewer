<script setup lang="ts">
import { computed } from 'vue'
import type { Asset } from '../types/asset'
import RatingStars from './RatingStars.vue'
import ColorLabel from './ColorLabel.vue'
import { useI18n } from '../i18n'

const props = defineProps<{
  asset: Asset | null
}>()

const emit = defineEmits<{
  close: []
  prev: []
  next: []
  rate: [id: number, rating: number]
  label: [id: number, label: string]
  openInReview: []
}>()

function thumbUrl(id: number, size: string): string {
  return `/api/v1/thumbs/${id}?size=${size}`
}

const exif = computed(() => {
  const mf = props.asset?.jpg_file || props.asset?.raw_file
  return mf?.exif
})

const { t } = useI18n()

const matchLabel = computed(() => {
  if (!props.asset) return ''
  const hasRaw = !!props.asset.raw_file
  const hasJpg = !!props.asset.jpg_file
  if (hasRaw && hasJpg) return t('match.raw_jpg')
  if (hasRaw) return t('match.raw')
  if (hasJpg) return t('match.jpg')
  return t('match.unknown')
})

const matchClass = computed(() => {
  if (!props.asset) return ''
  const hasRaw = !!props.asset.raw_file
  const hasJpg = !!props.asset.jpg_file
  if (hasRaw && hasJpg) return 'paired'
  return 'orphan'
})
</script>

<template>
  <Teleport to="body">
    <div v-if="asset" class="preview-overlay" @click.self="emit('close')">
      <div class="preview-container">
        <button class="nav-btn prev" @click="emit('prev')">&lsaquo;</button>
        <button class="nav-btn next" @click="emit('next')">&rsaquo;</button>
        <button class="close-btn" @click="emit('close')">&times;</button>

        <button class="open-review-btn" @click="emit('openInReview')">{{ t('preview.open_in_review') }}</button>
        <div class="match-badge">{{ matchLabel }}</div>

        <div class="preview-image">
          <img :src="thumbUrl(asset.id, 'full')" :alt="asset.name" />
        </div>

        <div class="preview-info">
          <div class="info-left">
            <span class="filename">{{ asset.name }}</span>
            <span v-if="exif?.camera_model" class="exif-item">{{ exif.camera_model }}</span>
            <span v-if="exif?.focal_length" class="exif-item">{{ exif.focal_length }}mm</span>
            <span v-if="exif?.aperture" class="exif-item">f/{{ exif.aperture }}</span>
            <span v-if="exif?.iso" class="exif-item">ISO {{ exif.iso }}</span>
            <span v-if="exif?.shutter_speed" class="exif-item">{{ exif.shutter_speed }}</span>
          </div>
          <div class="info-right">
            <RatingStars :rating="asset!.rating" @rate="(r) => emit('rate', asset!.id, r)" />
            <ColorLabel :color-label="asset!.color_label" @select="(l) => emit('label', asset!.id, l)" />
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.preview-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.92);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-container {
  position: relative;
  max-width: 95vw;
  max-height: 95vh;
  display: flex;
  flex-direction: column;
}

.preview-image {
  display: flex;
  align-items: center;
  justify-content: center;
  max-height: 85vh;
}

.preview-image img {
  max-width: 100%;
  max-height: 85vh;
  object-fit: contain;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: #fff;
  font-size: 2.5rem;
  width: 50px;
  height: 80px;
  cursor: pointer;
  z-index: 2;
  transition: background 0.2s;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.nav-btn.prev { left: -60px; border-radius: 0 8px 8px 0; }
.nav-btn.next { right: -60px; border-radius: 8px 0 0 8px; }

.close-btn {
  position: absolute;
  top: -40px;
  right: 0;
  background: none;
  border: none;
  color: #fff;
  font-size: 2rem;
  cursor: pointer;
  z-index: 2;
}

.open-review-btn {
  position: absolute;
  top: -40px;
  right: 120px;
  background: #e94560;
  border: none;
  color: #fff;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  z-index: 2;
}

.open-review-btn:hover {
  background: #c73e54;
}

.match-badge {
  position: absolute;
  top: -40px;
  right: 40px;
  padding: 4px 10px;
  border-radius: 3px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #fff;
  z-index: 2;
}

.preview-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  margin-top: 8px;
  border-top: 1px solid #333;
}

.info-left {
  display: flex;
  gap: 16px;
  align-items: center;
  color: #ccc;
  font-size: 0.9rem;
}

.filename {
  font-weight: 600;
  color: #fff;
}

.exif-item {
  color: #999;
}

.info-right {
  display: flex;
  gap: 16px;
  align-items: center;
}
</style>
