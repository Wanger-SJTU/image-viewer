import { ref, computed, onMounted, onUnmounted, type Ref } from 'vue'

export function useVirtualScroll(
  containerRef: Ref<HTMLElement | null>,
  itemHeight: number,
  totalCount: Ref<number>,
  loadPage: (page: number) => void
) {
  const viewportHeight = ref(0)
  const scrollTop = ref(0)

  const visibleCount = computed(() => Math.ceil(viewportHeight.value / itemHeight) + 4) // +buffer
  const startIndex = computed(() => Math.max(0, Math.floor(scrollTop.value / itemHeight) - 2))
  const totalPages = computed(() => Math.ceil(totalCount.value / 50))
  const currentPage = computed(() => Math.floor(startIndex.value / 50) + 1)

  let lastPage = 0
  function onScroll(e: Event) {
    const el = e.target as HTMLElement
    scrollTop.value = el.scrollTop
    viewportHeight.value = el.clientHeight

    const page = Math.floor((el.scrollTop + el.clientHeight) / (itemHeight * 50)) + 1
    if (page !== lastPage && page <= totalPages.value) {
      lastPage = page
      loadPage(page)
    }
  }

  onMounted(() => {
    if (containerRef.value) {
      containerRef.value.addEventListener('scroll', onScroll, { passive: true })
      viewportHeight.value = containerRef.value.clientHeight
    }
  })

  onUnmounted(() => {
    containerRef.value?.removeEventListener('scroll', onScroll)
  })

  return { visibleCount, startIndex, currentPage }
}
