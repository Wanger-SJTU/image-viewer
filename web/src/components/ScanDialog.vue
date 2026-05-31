<script setup lang="ts">
import { ref } from 'vue'
import { useScanStore } from '../stores/scan'
import { useAssetStore } from '../stores/assets'

const scanStore = useScanStore()
const assetStore = useAssetStore()

const path = ref('')
const visible = ref(false)

defineExpose({ open, close })

function open() {
  path.value = ''
  visible.value = true
}

function close() {
  visible.value = false
}

async function startScan() {
  if (!path.value.trim()) return
  // Don't close dialog — let user see progress
  await scanStore.scan(path.value.trim())
  assetStore.fetchAssets()
}

function progressPct(): number {
  const p = scanStore.progress
  if (p.found === 0) return 0
  return Math.round((p.processed / p.found) * 100)
}

const phaseLabels: Record<string, string> = {
  scanning: 'Scanning files...',
  matching: 'Matching pairs...',
  exif: 'Reading EXIF...',
  saving: 'Saving to database...',
  done: 'Done',
  error: 'Error',
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="scan-overlay" @click.self="!scanStore.scanning && close()">
      <div class="scan-dialog">
        <h3>Scan Directory</h3>
        <p class="scan-hint">Enter the absolute path to a directory containing photos.</p>
        <input
          v-model="path"
          type="text"
          class="scan-input"
          placeholder="/path/to/photos"
          :disabled="scanStore.scanning"
          @keyup.enter="startScan"
        />
        <div class="scan-actions">
          <button class="btn btn-cancel" :disabled="scanStore.scanning" @click="close">Cancel</button>
          <button class="btn btn-scan" :disabled="!path.trim() || scanStore.scanning" @click="startScan">
            {{ scanStore.scanning ? 'Scanning...' : 'Scan' }}
          </button>
        </div>

        <!-- Progress bar -->
        <div v-if="scanStore.scanning" class="progress-section">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progressPct() + '%' }" />
          </div>
          <p class="progress-text">
            {{ phaseLabels[scanStore.progress.phase] || scanStore.progress.phase }}
            <span v-if="scanStore.progress.found > 0">
              ({{ scanStore.progress.processed }} / {{ scanStore.progress.found }})
            </span>
          </p>
          <p v-if="scanStore.progress.matched > 0 || scanStore.progress.orphans > 0" class="progress-stats">
            Matched: {{ scanStore.progress.matched }} | Orphans: {{ scanStore.progress.orphans }}
          </p>
        </div>

        <p v-if="scanStore.error" class="scan-error">{{ scanStore.error }}</p>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.scan-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.scan-dialog {
  background: #16213e;
  border: 1px solid #0f3460;
  border-radius: 8px;
  padding: 24px;
  width: 420px;
  max-width: 90vw;
}

h3 {
  margin: 0 0 8px;
  font-size: 1.1rem;
}

.scan-hint {
  color: #888;
  font-size: 0.85rem;
  margin: 0 0 16px;
}

.scan-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #0f3460;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  font-size: 0.9rem;
  outline: none;
  box-sizing: border-box;
}

.scan-input:focus {
  border-color: #e94560;
}

.scan-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.btn {
  padding: 6px 18px;
  border: none;
  border-radius: 4px;
  font-size: 0.85rem;
  cursor: pointer;
}

.btn-cancel {
  background: #333;
  color: #ccc;
}

.btn-scan {
  background: #e94560;
  color: #fff;
}

.btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.scan-error {
  color: #e94560;
  font-size: 0.85rem;
  margin-top: 8px;
}

/* Progress */
.progress-section {
  margin-top: 16px;
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: #1a1a2e;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #e94560;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-text {
  color: #aaa;
  font-size: 0.8rem;
  margin: 8px 0 4px;
}

.progress-stats {
  color: #888;
  font-size: 0.75rem;
  margin: 0;
}
</style>
