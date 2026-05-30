<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  dateAfter?: string
  dateBefore?: string
}>()

const emit = defineEmits<{
  update: [range: { captured_after?: string; captured_before?: string }]
}>()

const currentMonth = ref(new Date())
const pickingAfter = ref(true) // true=picking from, false=picking to

const monthLabel = computed(() => {
  const d = currentMonth.value
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'long' })
})

const weeks = computed(() => {
  const year = currentMonth.value.getFullYear()
  const month = currentMonth.value.getMonth()
  const first = new Date(year, month, 1)
  const last = new Date(year, month + 1, 0)
  const startOffset = first.getDay()

  const days: (number | null)[] = []
  for (let i = 0; i < startOffset; i++) days.push(null)
  for (let d = 1; d <= last.getDate(); d++) days.push(d)

  const weeks: (number | null)[][] = []
  for (let i = 0; i < days.length; i += 7) {
    weeks.push(days.slice(i, i + 7))
  }
  return weeks
})

const afterDate = computed(() => {
  if (!props.dateAfter) return null
  return new Date(props.dateAfter + 'T00:00:00')
})
const beforeDate = computed(() => {
  if (!props.dateBefore) return null
  return new Date(props.dateBefore + 'T00:00:00')
})

function fmt(d: Date): string {
  return d.toISOString().slice(0, 10)
}

function isInRange(day: number): boolean {
  const d = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), day)
  if (afterDate.value && beforeDate.value) {
    return d >= afterDate.value && d <= beforeDate.value
  }
  return false
}

function isAfter(day: number): boolean {
  if (!afterDate.value) return false
  const d = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), day)
  return fmt(d) === fmt(afterDate.value)
}

function isBefore(day: number): boolean {
  if (!beforeDate.value) return false
  const d = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), day)
  return fmt(d) === fmt(beforeDate.value)
}

function pickDay(day: number) {
  const d = new Date(currentMonth.value.getFullYear(), currentMonth.value.getMonth(), day)
  const s = fmt(d)

  if (pickingAfter.value) {
    emit('update', { captured_after: s, captured_before: props.dateBefore })
    pickingAfter.value = false
  } else {
    if (afterDate.value && d < afterDate.value) {
      // clicked before current "from", reset and start new range
      emit('update', { captured_after: s, captured_before: undefined })
      pickingAfter.value = false
    } else {
      emit('update', { captured_after: props.dateAfter, captured_before: s })
      pickingAfter.value = true
    }
  }
}

function clearDates() {
  emit('update', { captured_after: undefined, captured_before: undefined })
  pickingAfter.value = true
}

function prevMonth() {
  const m = currentMonth.value
  currentMonth.value = new Date(m.getFullYear(), m.getMonth() - 1, 1)
}

function nextMonth() {
  const m = currentMonth.value
  currentMonth.value = new Date(m.getFullYear(), m.getMonth() + 1, 1)
}
</script>

<template>
  <div class="calendar">
    <div class="cal-header">
      <button class="cal-nav" @click="prevMonth">&lsaquo;</button>
      <span class="cal-month">{{ monthLabel }}</span>
      <button class="cal-nav" @click="nextMonth">&rsaquo;</button>
    </div>

    <div class="cal-weekdays">
      <span v-for="w in ['Su','Mo','Tu','We','Th','Fr','Sa']" :key="w" class="cal-wd">{{ w }}</span>
    </div>

    <div v-for="(week, wi) in weeks" :key="wi" class="cal-week">
      <button
        v-for="(day, di) in week"
        :key="di"
        class="cal-day"
        :class="{
          empty: day === null,
          'in-range': day !== null && isInRange(day!),
          after: day !== null && isAfter(day!),
          before: day !== null && isBefore(day!),
        }"
        :disabled="day === null"
        @click="day !== null && pickDay(day)"
      >
        {{ day }}
      </button>
    </div>

    <div class="cal-info" v-if="dateAfter || dateBefore">
      <span v-if="dateAfter" class="cal-picked">{{ dateAfter }}</span>
      <span v-if="dateBefore" class="cal-sep"> &ndash; </span>
      <span v-if="dateBefore" class="cal-picked">{{ dateBefore }}</span>
      <button class="cal-clear" @click="clearDates">&times;</button>
    </div>
    <div class="cal-info cal-hint" v-else>
      Click a date to set the range
    </div>
  </div>
</template>

<style scoped>
.calendar {
  user-select: none;
}

.cal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.cal-month {
  font-size: 0.78rem;
  color: #ccc;
  font-weight: 500;
}

.cal-nav {
  background: none;
  border: none;
  color: #888;
  font-size: 1rem;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

.cal-nav:hover {
  color: #e94560;
}

.cal-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 1px;
  margin-bottom: 2px;
}

.cal-wd {
  font-size: 0.6rem;
  color: #666;
  text-align: center;
  padding: 2px 0;
}

.cal-week {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 1px;
}

.cal-day {
  aspect-ratio: 1;
  border: none;
  background: transparent;
  color: #aaa;
  font-size: 0.7rem;
  cursor: pointer;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.cal-day.empty {
  cursor: default;
}

.cal-day:hover:not(.empty) {
  background: #e94560;
  color: #fff;
}

.cal-day.in-range {
  background: #3a1a2e;
  color: #e94560;
  border-radius: 0;
}

.cal-day.after {
  background: #e94560;
  color: #fff;
  border-radius: 3px;
}

.cal-day.before {
  background: #e94560;
  color: #fff;
  border-radius: 3px;
}

.cal-info {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.cal-picked {
  font-size: 0.7rem;
  color: #e94560;
  background: #3a1a2e;
  padding: 2px 5px;
  border-radius: 2px;
}

.cal-sep {
  color: #666;
  font-size: 0.7rem;
}

.cal-clear {
  background: none;
  border: none;
  color: #666;
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0;
  margin-left: auto;
}

.cal-clear:hover {
  color: #e94560;
}

.cal-hint {
  color: #555;
  font-size: 0.65rem;
}
</style>
