import { onMounted, onUnmounted } from 'vue'

interface ShortcutMap {
  [key: string]: () => void
}

export function useKeyboardShortcut(shortcuts: () => ShortcutMap) {
  function handler(e: KeyboardEvent) {
    // Ignore when typing in inputs
    const target = e.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT') {
      return
    }

    const key = e.key
    const map = shortcuts()
    if (map[key]) {
      e.preventDefault()
      map[key]()
    }
  }

  onMounted(() => document.addEventListener('keydown', handler))
  onUnmounted(() => document.removeEventListener('keydown', handler))
}
