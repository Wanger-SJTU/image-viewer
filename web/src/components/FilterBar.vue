<script setup lang="ts">
import { ref } from 'vue'
import type { AssetFilter } from '../types/filter'

const props = defineProps<{
  filter: AssetFilter
}>()

const emit = defineEmits<{
  update: [filter: Partial<AssetFilter>]
  clear: []
}>()

const showAdvanced = ref(false)

function setRating(r: number) {
  emit('update', { rating: props.filter.rating === r ? undefined : r })
}

function setLabel(l: string) {
  emit('update', { color_label: props.filter.color_label === l ? undefined : l })
}

function setSearch(e: Event) {
  const target = e.target as HTMLInputElement
  emit('update', { search: target.value || undefined })
}
</script>

<template>
  <div class="filter-bar">
    <div class="filter-row">
      <input
        type="text"
        class="search-input"
        placeholder="Search by filename..."
        :value="filter.search || ''"
        @input="setSearch"
      />

      <div class="rating-filter">
        <button
          v-for="r in [1, 2, 3, 4, 5]"
          :key="r"
          class="filter-star"
          :class="{ active: filter.rating && r <= filter.rating }"
          @click="setRating(r)"
        >
          &#9733;
        </button>
      </div>

      <button class="toggle-btn" @click="showAdvanced = !showAdvanced">
        {{ showAdvanced ? 'Hide' : 'Filters' }}
      </button>

      <button class="clear-btn" @click="emit('clear')">Clear</button>
    </div>

    <div v-if="showAdvanced" class="filter-advanced">
      <select
        class="filter-select"
        :value="filter.match_status || ''"
        @change="emit('update', { match_status: ($event.target as HTMLSelectElement).value || undefined })"
      >
        <option value="">All types</option>
        <option value="paired">Paired (RAW+JPG)</option>
        <option value="orphan">Orphan</option>
      </select>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  padding: 8px 12px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-input {
  flex: 1;
  max-width: 300px;
  padding: 6px 12px;
  border: 1px solid #0f3460;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.85rem;
  outline: none;
}

.search-input:focus {
  border-color: #e94560;
}

.rating-filter {
  display: flex;
  gap: 2px;
}

.filter-star {
  background: none;
  border: none;
  color: #444;
  font-size: 1rem;
  cursor: pointer;
  padding: 0 1px;
}

.filter-star.active {
  color: #f4c430;
}

.toggle-btn,
.clear-btn {
  padding: 4px 12px;
  border: 1px solid #0f3460;
  border-radius: 4px;
  background: #1a1a2e;
  color: #ccc;
  font-size: 0.8rem;
  cursor: pointer;
}

.toggle-btn:hover,
.clear-btn:hover {
  border-color: #e94560;
}

.filter-advanced {
  margin-top: 8px;
  display: flex;
  gap: 8px;
}

.filter-select {
  padding: 4px 8px;
  border: 1px solid #0f3460;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.85rem;
}
</style>
