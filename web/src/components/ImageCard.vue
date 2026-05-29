<script setup lang="ts">
import type { Asset } from '../types/asset'
import RatingStars from './RatingStars.vue'
import ColorLabel from './ColorLabel.vue'

const props = defineProps<{
  asset: Asset
}>()

const emit = defineEmits<{
  select: [asset: Asset]
  rate: [id: number, rating: number]
  label: [id: number, label: string]
}>()

function thumbUrl(id: number, size: string): string {
  return `/api/v1/thumbs/${id}?size=${size}`
}
</script>

<template>
  <div class="image-card" @click="emit('select', asset)">
    <div class="card-image">
      <img
        v-if="asset.grid_thumb"
        :src="thumbUrl(asset.id, 'grid')"
        :alt="asset.name"
        loading="lazy"
      />
      <div v-else class="no-thumb">?</div>
      <div class="card-overlay">
        <span class="match-badge" :class="asset.match_status">
          {{ asset.match_status === 'paired' ? 'RAW+JPG' : 'ORPHAN' }}
        </span>
      </div>
    </div>
    <div class="card-info">
      <div class="card-name">{{ asset.name }}</div>
      <RatingStars :rating="asset.rating" @rate="(r) => emit('rate', asset.id, r)" />
      <ColorLabel :color-label="asset.color_label" @select="(l) => emit('label', asset.id, l)" />
    </div>
  </div>
</template>

<style scoped>
.image-card {
  border-radius: 6px;
  overflow: hidden;
  background: #16213e;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}

.image-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.card-image {
  position: relative;
  aspect-ratio: 3/2;
  background: #0f3460;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-thumb {
  font-size: 2rem;
  color: #555;
}

.card-overlay {
  position: absolute;
  top: 0;
  right: 0;
  padding: 4px;
}

.match-badge {
  font-size: 0.65rem;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
}

.match-badge.paired {
  background: #2e7d32;
  color: #fff;
}

.match-badge.orphan {
  background: #e65100;
  color: #fff;
}

.card-info {
  padding: 6px 8px 8px;
}

.card-name {
  font-size: 0.8rem;
  color: #ccc;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
