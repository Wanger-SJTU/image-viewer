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

const showAdvanced = ref(false)
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
  <div class="filter-bar">
    <div class="filter-row">
      <input
        type="text"
        class="search-input"
        placeholder="Search by filename..."
        :value="filter.search || ''"
        @input="updateFilter({ search: ($event.target as HTMLInputElement).value || undefined })"
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

      <button class="clear-btn" @click="clearAll">Clear</button>
    </div>

    <div v-if="showAdvanced" class="filter-advanced">
      <!-- File type -->
      <label class="filter-label">
        Type
        <select
          class="filter-select"
          :value="filter.file_type || ''"
          @change="updateFilter({ file_type: ($event.target as HTMLSelectElement).value || undefined })"
        >
          <option value="">All</option>
          <option v-for="t in options.file_types" :key="t" :value="t">{{ t.toUpperCase() }}</option>
        </select>
      </label>

      <!-- Camera model -->
      <label class="filter-label">
        Camera
        <select
          class="filter-select"
          :value="filter.camera_model || ''"
          @change="updateFilter({ camera_model: ($event.target as HTMLSelectElement).value || undefined })"
        >
          <option value="">All</option>
          <option v-for="m in options.camera_models" :key="m" :value="m">{{ m }}</option>
        </select>
      </label>

      <!-- Date range -->
      <label class="filter-label">
        Date
        <input
          type="date"
          class="filter-input filter-input--date"
          :value="filter.captured_after || ''"
          @input="updateFilter({ captured_after: ($event.target as HTMLInputElement).value || undefined })"
        />
        <span class="filter-sep">-</span>
        <input
          type="date"
          class="filter-input filter-input--date"
          :value="filter.captured_before || ''"
          @input="updateFilter({ captured_before: ($event.target as HTMLInputElement).value || undefined })"
        />
      </label>

      <!-- Focal length range -->
      <label class="filter-label">
        FL (mm)
        <select
          class="filter-select"
          :value="filter.focal_length_min || ''"
          @change="updateFilter({ focal_length_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.focal_lengths" :key="v" :value="v">{{ v }}</option>
        </select>
        <span class="filter-sep">-</span>
        <select
          class="filter-select"
          :value="filter.focal_length_max || ''"
          @change="updateFilter({ focal_length_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.focal_lengths" :key="v" :value="v">{{ v }}</option>
        </select>
      </label>

      <!-- Aperture range -->
      <label class="filter-label">
        Aperture
        <select
          class="filter-select"
          :value="filter.aperture_min || ''"
          @change="updateFilter({ aperture_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.apertures" :key="v" :value="v">f/{{ v }}</option>
        </select>
        <span class="filter-sep">-</span>
        <select
          class="filter-select"
          :value="filter.aperture_max || ''"
          @change="updateFilter({ aperture_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.apertures" :key="v" :value="v">f/{{ v }}</option>
        </select>
      </label>

      <!-- ISO range -->
      <label class="filter-label">
        ISO
        <select
          class="filter-select"
          :value="filter.iso_min || ''"
          @change="updateFilter({ iso_min: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">min</option>
          <option v-for="v in options.isos" :key="v" :value="v">{{ v }}</option>
        </select>
        <span class="filter-sep">-</span>
        <select
          class="filter-select"
          :value="filter.iso_max || ''"
          @change="updateFilter({ iso_max: toNum(($event.target as HTMLSelectElement).value) })"
        >
          <option value="">max</option>
          <option v-for="v in options.isos" :key="v" :value="v">{{ v }}</option>
        </select>
      </label>

      <!-- Color label -->
      <label class="filter-label">
        Label
        <select
          class="filter-select"
          :value="filter.color_label || ''"
          @change="updateFilter({ color_label: ($event.target as HTMLSelectElement).value || undefined })"
        >
          <option value="">All</option>
          <option v-for="l in options.color_labels" :key="l" :value="l">{{ l }}</option>
        </select>
      </label>
    </div>
  </div>
</template>

<script lang="ts">
function toNum(v: string): number | undefined {
  if (v === '') return undefined
  return parseFloat(v)
}
</script>

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
  max-width: 240px;
  padding: 5px 10px;
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
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #999;
  font-size: 0.78rem;
  white-space: nowrap;
}

.filter-select,
.filter-input {
  padding: 3px 6px;
  border: 1px solid #0f3460;
  border-radius: 3px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.8rem;
  outline: none;
}

.filter-select:focus,
.filter-input:focus {
  border-color: #e94560;
}

.filter-input--date {
  width: 120px;
  min-width: 0;
}

.filter-sep {
  color: #555;
  font-size: 0.75rem;
}
</style>
