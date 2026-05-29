<script setup lang="ts">
const props = defineProps<{
  colorLabel: string
}>()

const emit = defineEmits<{
  select: [label: string]
}>()

const labels = [
  { value: '', label: 'None', color: '#666' },
  { value: 'red', label: 'Red', color: '#e94560' },
  { value: 'orange', label: 'Orange', color: '#f4a460' },
  { value: 'yellow', label: 'Yellow', color: '#f4c430' },
  { value: 'green', label: 'Green', color: '#4caf50' },
  { value: 'blue', label: 'Blue', color: '#42a5f5' },
  { value: 'purple', label: 'Purple', color: '#ab47bc' },
]
</script>

<template>
  <div class="color-labels">
    <button
      v-for="item in labels"
      :key="item.value"
      class="label-dot"
      :class="{ active: colorLabel === item.value }"
      :style="{ backgroundColor: item.color }"
      :title="item.label"
      @click="emit('select', item.value)"
    />
  </div>
</template>

<style scoped>
.color-labels {
  display: inline-flex;
  gap: 4px;
}

.label-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  transition: border-color 0.1s;
}

.label-dot.active {
  border-color: #fff;
  box-shadow: 0 0 4px rgba(255, 255, 255, 0.5);
}

.label-dot:hover {
  border-color: #aaa;
}
</style>
