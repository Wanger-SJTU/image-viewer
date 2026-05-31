<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Asset } from '../types/asset'
import { useI18n } from '../i18n'

const props = defineProps<{
  assets: Asset[]
  total: number
}>()

const emit = defineEmits<{
  close: []
  restore: [id: number]
  purge: [id: number, fileType: 'both' | 'jpg' | 'raw']
}>()

const { t } = useI18n()
const purgeMenuId = ref<number | null>(null)
const purgeConfirmId = ref<number | null>(null)
const purgeConfirmType = ref<'both' | 'jpg' | 'raw'>('both')

function thumbUrl(id: number, size: string): string {
  return `/api/v1/thumbs/${id}?size=${size}`
}

function togglePurgeMenu(id: number) {
  purgeMenuId.value = purgeMenuId.value === id ? null : id
}

function confirmPurge(id: number, fileType: 'both' | 'jpg' | 'raw') {
  purgeConfirmId.value = id
  purgeConfirmType.value = fileType
  purgeMenuId.value = null
}

function execPurge() {
  if (purgeConfirmId.value !== null) {
    emit('purge', purgeConfirmId.value, purgeConfirmType.value)
    purgeConfirmId.value = null
  }
}

function cancelPurge() {
  purgeConfirmId.value = null
}

function hasFileType(asset: Asset, type: string): boolean {
  if (type === 'raw') return !!asset.raw_file
  if (type === 'jpg') return !!asset.jpg_file
  return true
}

function trashedDate(asset: Asset): string {
  if (!asset.deleted_at) return ''
  try {
    const d = new Date(asset.deleted_at)
    return d.toLocaleString()
  } catch {
    return ''
  }
}

const purgeOptions: { type: 'both' | 'jpg' | 'raw'; label: string }[] = [
  { type: 'both', label: 'trash.purge_both' },
  { type: 'jpg', label: 'trash.purge_jpg' },
  { type: 'raw', label: 'trash.purge_raw' },
]
</script>

<template>
  <div class="trash-panel">
    <div class="trash-header">
      <h2>{{ t('trash.title') }} ({{ total }})</h2>
      <button class="close-btn" @click="emit('close')">{{ t('toolbar.close_trash') }}</button>
    </div>

    <div v-if="assets.length === 0" class="trash-empty">
      {{ t('trash.empty') }}
    </div>

    <div v-else class="trash-grid">
      <div
        v-for="asset in assets"
        :key="asset.id"
        class="trash-item"
      >
        <div class="trash-thumb">
          <img
            v-if="asset.grid_thumb"
            :src="thumbUrl(asset.id, 'grid')"
            :alt="asset.name"
            loading="lazy"
          />
          <div v-else class="no-thumb">?</div>
        </div>
        <div class="trash-info">
          <span class="trash-name">{{ asset.name }}</span>
          <span class="trash-date">{{ trashedDate(asset) }}</span>
        </div>
        <div class="trash-actions">
          <button class="restore-btn" @click="emit('restore', asset.id)">
            {{ t('trash.restore') }}
          </button>
          <div class="purge-wrapper">
            <button
              v-if="purgeConfirmId !== asset.id"
              class="purge-btn"
              @click="togglePurgeMenu(asset.id)"
            >
              {{ t('trash.purge') }}
            </button>
            <div v-else class="purge-confirm">
              <span>{{ t('trash.purge_confirm') }}</span>
              <button class="confirm-yes" @click="execPurge">{{ t('trash.purge_ok') }}</button>
              <button class="confirm-no" @click="cancelPurge">Cancel</button>
            </div>
            <div v-if="purgeMenuId === asset.id" class="purge-menu">
              <button
                v-for="opt in purgeOptions"
                v-show="hasFileType(asset, opt.type)"
                :key="opt.type"
                @click="confirmPurge(asset.id, opt.type)"
              >
                {{ t(opt.label) }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.trash-panel {
  background: #1a1a2e;
  border-radius: 8px;
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.trash-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.trash-header h2 {
  color: #e0e0e0;
  font-size: 1.2rem;
  margin: 0;
}

.close-btn {
  background: #333;
  color: #ccc;
  border: none;
  padding: 6px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
}

.close-btn:hover {
  background: #555;
}

.trash-empty {
  color: #888;
  text-align: center;
  padding: 60px 0;
  font-size: 1rem;
}

.trash-grid {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.trash-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 8px;
  background: #16213e;
  border-radius: 6px;
  transition: background 0.15s;
}

.trash-item:hover {
  background: #1c2a4a;
}

.trash-thumb {
  width: 60px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 3px;
  overflow: hidden;
  background: #0a0a1a;
  display: flex;
  align-items: center;
  justify-content: center;
}

.trash-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-thumb {
  color: #555;
  font-size: 1.2rem;
}

.trash-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.trash-name {
  color: #ddd;
  font-size: 0.9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trash-date {
  color: #777;
  font-size: 0.75rem;
}

.trash-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.restore-btn {
  background: #2d5a27;
  color: #e0e0e0;
  border: none;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
}

.restore-btn:hover {
  background: #3a7a33;
}

.purge-wrapper {
  position: relative;
}

.purge-btn {
  background: #5a2727;
  color: #e0e0e0;
  border: none;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
}

.purge-btn:hover {
  background: #7a3333;
}

.purge-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: #2a2a3e;
  border: 1px solid #444;
  border-radius: 4px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  min-width: 100px;
}

.purge-menu button {
  background: none;
  border: none;
  color: #ddd;
  padding: 6px 12px;
  text-align: left;
  cursor: pointer;
  font-size: 0.8rem;
}

.purge-menu button:hover {
  background: #3a3a4e;
}

.purge-confirm {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  color: #ccc;
}

.confirm-yes {
  background: #8b0000;
  color: #fff;
  border: none;
  padding: 3px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 0.75rem;
}

.confirm-no {
  background: #333;
  color: #ccc;
  border: none;
  padding: 3px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 0.75rem;
}
</style>
