<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { AssetFilter } from '../types/filter'
import type { FilterOptions } from '../types/api'
import { getFilterOptions } from '../api/assets'

const props = defineProps<{
  filter: AssetFilter
}>()

const emit = defineEmits<{
  update: [filter: Partial<AssetFilter>]
  clear: []
}>()

const options = ref<FilterOptions>({
  camera_models: [],
  focal_lengths: [],
  apertures: [],
  isos: [],
  color_labels: [],
  file_types: ['jpg', 'raw', 'both'],
})

onMounted(async () => {
  try {
    options.value = await getFilterOptions()
  } catch {
    // use defaults
  }
})

function setRating(r: number) {
  emit('update', { rating: props.filter.rating === r ? undefined : r })
}

function updateFilter(patch: Partial<AssetFilter>) {
  emit('update', patch)
}

function clearAll() {
  emit('clear')
}
</script>

<template>
  <div class="filter-sidebar">
    <div class="filter-title">Filters</div>

    <!-- Search -->
    <input
      type="text"
      class="filter-search"
      placeholder="Search filename..."
      :value="filter.search || ''"
      @input="updateFilter({ search: ($event.target as HTMLInputElement).value || undefined })"
    />

    <!-- Rating -->
    <div class="filter-section">
      <div class="filter-label">Rating</div>
      <div class="rating-stars">
        <button
          v-for="r in [1, 2, 3, 4, 5]"
          :key="r"
          class="filter-star"
          :class="{ active: filter.rating && r <= filter.rating }"
          @click="setRating(r)"
        >&#9733;</button>
        <button
          v-if="filter.rating"
          class="filter-star clear-star"
          @click="setRating(filter.rating!)"
        >&#10005;</button>
      </div>
    </div>

    <!-- File type -->
    <div class="filter-section">
      <label class="filter-label">File Type</label>
      <select
        class="filter-select"
        :value="filter.file_type || ''"
        @change="updateFilter({ file_type: ($event.target as HTMLSelectElement).value || undefined })"
      >
        <option value="">All</option>
        <option v-for="t in options.file_types" :key="t" :value="t">{{ t.toUpperCase() }}</option>
      </select>
    </div>

    <!-- Camera -->
    <div class="filter-section">
      <label class="filter-label">Camera</label>
      <select
        class="filter-select"
        :value="filter.camera_model || ''"
        @change="updateFilter({ camera_model: ($event.target as HTMLSelectElement).value || undefined })"
      >
        <option value="">All</option>
        <option v-for="m in options.camera_models" :key="m" :value="m">{{ m }}</option>
      </select>
    </div>

    <!-- Date range -->
    <div class="filter-section">
      <label class="filter-label">Date</label>
      <div class="date-row">
        <span class="date-hint">From</span>
        <input
          type="date"
          class="filter-select"
          :value="filter.captured_after || ''"
          :max="filter.captured_before || undefined"
          @input="updateFilter({ captured_after: ($event.target as HTMLInputElement).value || undefined })"
        />
      </div>
      <div class="date-row">
        <span class="date-hint">To</span>
        <input
          type="date"
          class="filter-select"
          :value="filter.captured_before || ''"
          :min="filter.captured_after || undefined"
          @input="updateFilter({ captured_before: ($event.target as HTMLInputElement).value || undefined })"
        />
      </div>
    </div>

    <!-- Focal length -->
    <div class="filter-section">
      <label class="filter-label">Focal Length (mm)</label>
      <div class="filter-range">
        <select
          class="filter-select flex-1"
          :value="filter.focal_length_min || ''"
          @change="updateFilter({ focal_length_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.focal_lengths" :key="v" :value="v">{{ v }}</option>
        </select>
        <span class="range-sep">-</span>
        <select
          class="filter-select flex-1"
          :value="filter.focal_length_max || ''"
          @change="updateFilter({ focal_length_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.focal_lengths" :key="v" :value="v">{{ v }}</option>
        </select>
      </div>
    </div>

    <!-- Aperture -->
    <div class="filter-section">
      <label class="filter-label">Aperture</label>
      <div class="filter-range">
        <select
          class="filter-select flex-1"
          :value="filter.aperture_min || ''"
          @change="updateFilter({ aperture_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.apertures" :key="v" :value="v">f/{{ v }}</option>
        </select>
        <span class="range-sep">-</span>
        <select
          class="filter-select flex-1"
          :value="filter.aperture_max || ''"
          @change="updateFilter({ aperture_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.apertures" :key="v" :value="v">f/{{ v }}</option>
        </select>
      </div>
    </div>

    <!-- ISO -->
    <div class="filter-section">
      <label class="filter-label">ISO</label>
      <div class="filter-range">
        <select
          class="filter-select flex-1"
          :value="filter.iso_min || ''"
          @change="updateFilter({ iso_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.isos" :key="v" :value="v">{{ v }}</option>
        </select>
        <span class="range-sep">-</span>
        <select
          class="filter-select flex-1"
          :value="filter.iso_max || ''"
          @change="updateFilter({ iso_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.isos" :key="v" :value="v">{{ v }}</option>
        </select>
      </div>
    </div>

    <!-- Color label -->
    <div class="filter-section">
      <label class="filter-label">Color Label</label>
      <select
        class="filter-select"
        :value="filter.color_label || ''"
        @change="updateFilter({ color_label: ($event.target as HTMLSelectElement).value || undefined })"
      >
        <option value="">All</option>
        <option v-for="l in options.color_labels" :key="l" :value="l">{{ l }}</option>
      </select>
    </div>

    <button class="clear-filter-btn" @click="clearAll">Clear All Filters</button>
  </div>
</template>

<script lang="ts">
function toNum(v: string): number | undefined {
  if (v === '') return undefined
  return parseFloat(v)
}
</script>

<style scoped>
.filter-sidebar {
  width: 220px;
  flex-shrink: 0;
  padding: 12px;
  background: #16213e;
  border-right: 1px solid #0f3460;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: #eee;
  margin-bottom: 4px;
}

.filter-search {
  padding: 5px 8px;
  border: 1px solid #0f3460;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.8rem;
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.filter-search:focus {
  border-color: #e94560;
}

.filter-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-label {
  font-size: 0.72rem;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.date-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.date-hint {
  font-size: 0.7rem;
  color: #666;
  width: 28px;
  flex-shrink: 0;
}

.filter-select {
  padding: 4px 6px;
  border: 1px solid #0f3460;
  border-radius: 3px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.78rem;
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.filter-select:focus {
  border-color: #e94560;
}

.filter-range {
  display: flex;
  align-items: center;
  gap: 4px;
}

.flex-1 {
  flex: 1;
  min-width: 0;
}

.range-sep {
  color: #555;
  font-size: 0.7rem;
  flex-shrink: 0;
}

.rating-stars {
  display: flex;
  gap: 1px;
  align-items: center;
}

.filter-star {
  background: none;
  border: none;
  color: #444;
  font-size: 1rem;
  cursor: pointer;
  padding: 0 1px;
  line-height: 1;
}

.filter-star.active {
  color: #f4c430;
}

.filter-star.clear-star {
  color: #666;
  font-size: 0.75rem;
  margin-left: 2px;
}

.clear-filter-btn {
  margin-top: 8px;
  padding: 5px 0;
  border: 1px solid #444;
  border-radius: 4px;
  background: transparent;
  color: #999;
  font-size: 0.78rem;
  cursor: pointer;
  width: 100%;
}

.clear-filter-btn:hover {
  border-color: #e94560;
  color: #e94560;
}
</style>
