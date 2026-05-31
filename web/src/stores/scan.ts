import { defineStore } from 'pinia'
import { ref } from 'vue'
import { startScan, getScanStatus, type ScanProgress } from '../api/scan'

export const useScanStore = defineStore('scan', () => {
  const scanning = ref(false)
  const progress = ref<ScanProgress>({
    phase: '',
    found: 0,
    processed: 0,
    matched: 0,
    orphans: 0,
    error: '',
  })
  const error = ref('')

  let pollTimer: ReturnType<typeof setInterval> | null = null

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  async function scan(path: string) {
    scanning.value = true
    error.value = ''
    progress.value = { phase: 'scanning', found: 0, processed: 0, matched: 0, orphans: 0, error: '' }
    try {
      await startScan(path)
      // Poll for progress
      pollTimer = setInterval(async () => {
        try {
          const p = await getScanStatus()
          progress.value = p
          if (p.phase === 'done' || p.phase === 'error') {
            stopPolling()
            scanning.value = false
            if (p.phase === 'error') {
              error.value = p.error || 'Unknown scan error'
            }
          }
        } catch {
          // ignore poll errors
        }
      }, 500)
    } catch (e) {
      error.value = (e as Error).message
      scanning.value = false
      stopPolling()
    }
  }

  return { scanning, progress, error, scan }
})
