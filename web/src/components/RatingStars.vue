<script setup lang="ts">
const props = defineProps<{
  rating: number
}>()

const emit = defineEmits<{
  rate: [rating: number]
}>()

const stars = [1, 2, 3, 4, 5]
</script>

<template>
  <div class="rating-stars">
    <button
      v-for="star in stars"
      :key="star"
      class="star"
      :class="{ active: star <= rating }"
      @click="emit('rate', star)"
    >
      &#9733;
    </button>
  </div>
</template>

<style scoped>
.rating-stars {
  display: inline-flex;
  gap: 1px;
}

.star {
  background: none;
  border: none;
  color: #444;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0 1px;
  transition: color 0.1s;
}

.star.active {
  color: #f4c430;
}

.star:hover,
.star:hover ~ .star {
  color: #f4c430;
}
</style>
