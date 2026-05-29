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
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 12px;
  padding: 12px;
}
</style>
