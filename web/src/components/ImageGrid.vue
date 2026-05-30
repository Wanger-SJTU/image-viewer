<script setup lang="ts">
import type { Asset } from '../types/asset'
import ImageCard from './ImageCard.vue'

defineProps<{
  assets: Asset[]
}>()

const emit = defineEmits<{
  select: [asset: Asset]
  rate: [id: number, rating: number]
  label: [id: number, label: string]
}>()
</script>

<template>
  <div class="image-grid">
    <ImageCard
      v-for="asset in assets"
      :key="asset.id"
      :asset="asset"
      @select="emit('select', $event)"
      @rate="(id, r) => emit('rate', id, r)"
      @label="(id, l) => emit('label', id, l)"
    />
  </div>
</template>

<style scoped>
.image-grid {
  column-count: 4;
  column-gap: 12px;
  padding: 12px;
}

.image-grid > * {
  break-inside: avoid;
  margin-bottom: 12px;
}

@media (max-width: 1200px) {
  .image-grid { column-count: 3; }
}

@media (max-width: 768px) {
  .image-grid { column-count: 2; }
}

@media (max-width: 480px) {
  .image-grid { column-count: 1; }
}
</style>
