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
  await scanStore.scan(path.value.trim())
  close()
  assetStore.fetchAssets()
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="scan-overlay" @click.self="close">
      <div class="scan-dialog">
        <h3>Scan Directory</h3>
        <p class="scan-hint">Enter the absolute path to a directory containing photos.</p>
        <input
          v-model="path"
          type="text"
          class="scan-input"
          placeholder="/path/to/photos"
          @keyup.enter="startScan"
        />
        <div class="scan-actions">
          <button class="btn btn-cancel" @click="close">Cancel</button>
          <button class="btn btn-scan" :disabled="!path.trim() || scanStore.scanning" @click="startScan">
            {{ scanStore.scanning ? 'Scanning...' : 'Scan' }}
          </button>
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
</style>
