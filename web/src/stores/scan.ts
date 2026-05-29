import { defineStore } from 'pinia'
import { ref } from 'vue'
import { startScan } from '../api/scan'

export const useScanStore = defineStore('scan', () => {
  const scanning = ref(false)
  const progress = ref({
    phase: '',
    found: 0,
    processed: 0,
    matched: 0,
    orphans: 0,
  })
  const error = ref('')

  async function scan(path: string) {
    scanning.value = true
    error.value = ''
    try {
      await startScan(path)
    } catch (e) {
      error.value = (e as Error).message
    } finally {
      scanning.value = false
    }
  }

  return { scanning, progress, error, scan }
})
