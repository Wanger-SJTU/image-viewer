import { ref, computed } from 'vue'

const zh: Record<string, string> = {
  'filter.title': '筛选',
  'filter.search': '搜索文件名...',
  'filter.rating': '评分',
  'filter.file_type': '文件类型',
  'filter.camera': '相机',
  'filter.date': '日期',
  'filter.focal_length': '焦段 (mm)',
  'filter.aperture': '光圈',
  'filter.iso': 'ISO',
  'filter.color_label': '颜色标签',
  'filter.clear': '清除筛选',
  'filter.all': '全部',
  'filter.min': '最小',
  'filter.max': '最大',
  'filter.cal_hint': '点击日期设置范围',
  'toolbar.scan': '扫描',
  'toolbar.clear_all': '清空',
  'toolbar.refresh': '刷新',
  'toolbar.assets': '{n} 张照片',
  'toolbar.no_assets': '暂无照片，点击扫描导入',
  'match.raw_jpg': 'RAW + JPG',
  'match.raw': 'RAW',
  'match.jpg': 'JPG',
  'match.unknown': '未知',
  'preview.close': '关闭',
  'preview.open_in_review': '审片模式打开',
  'preview.delete_confirm': '确定移入回收站吗？',
  'toolbar.clear_all_confirm': '确定删除所有照片和缩略图吗？此操作不可撤销。',
  'toolbar.preview_mode': '预览模式',
  'toolbar.review_mode': '审片模式',
  'trash.title': '回收站',
  'trash.empty': '回收站为空',
  'trash.restore': '恢复',
  'trash.purge': '彻底删除',
  'trash.purge_both': '全部',
  'trash.purge_jpg': '仅 JPG',
  'trash.purge_raw': '仅 RAW',
  'trash.purge_confirm': '确定彻底删除吗？此操作不可撤销。',
  'trash.restore_ok': '已恢复',
  'trash.purge_ok': '已彻底删除',
  'toolbar.trash': '回收站 ({n})',
  'toolbar.close_trash': '关闭回收站',
  'lang': 'English',
}

const en: Record<string, string> = {
  'filter.title': 'Filters',
  'filter.search': 'Search filename...',
  'filter.rating': 'Rating',
  'filter.file_type': 'File Type',
  'filter.camera': 'Camera',
  'filter.date': 'Date',
  'filter.focal_length': 'Focal Length (mm)',
  'filter.aperture': 'Aperture',
  'filter.iso': 'ISO',
  'filter.color_label': 'Color Label',
  'filter.clear': 'Clear All Filters',
  'filter.all': 'All',
  'filter.min': 'min',
  'filter.max': 'max',
  'filter.cal_hint': 'Click a date to set the range',
  'toolbar.scan': 'Scan',
  'toolbar.clear_all': 'Clear All',
  'toolbar.refresh': 'Refresh',
  'toolbar.assets': '{n} assets',
  'toolbar.no_assets': 'No assets. Click Scan to import photos.',
  'match.raw_jpg': 'RAW + JPG',
  'match.raw': 'RAW',
  'match.jpg': 'JPG',
  'match.unknown': 'Unknown',
  'preview.close': 'Close',
  'preview.open_in_review': 'Open in Review',
  'preview.delete_confirm': 'Move to trash?',
  'toolbar.clear_all_confirm': 'Delete all assets and thumbnails? This cannot be undone.',
  'toolbar.preview_mode': 'Preview',
  'toolbar.review_mode': 'Review',
  'trash.title': 'Trash',
  'trash.empty': 'Trash is empty',
  'trash.restore': 'Restore',
  'trash.purge': 'Delete Permanently',
  'trash.purge_both': 'Both',
  'trash.purge_jpg': 'JPG Only',
  'trash.purge_raw': 'RAW Only',
  'trash.purge_confirm': 'Delete permanently? This cannot be undone.',
  'trash.restore_ok': 'Restored',
  'trash.purge_ok': 'Deleted permanently',
  'toolbar.trash': 'Trash ({n})',
  'toolbar.close_trash': 'Close Trash',
  'lang': '中文',
}

type Locale = 'zh' | 'en'

const locale = ref<Locale>((localStorage.getItem('locale') as Locale) || 'zh')

const messages = { zh, en }

export function useI18n() {
  function t(key: string, params?: Record<string, string | number>): string {
    const msg = messages[locale.value][key]
    if (msg === undefined) return key
    if (!params) return msg
    return msg.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? `{${k}}`))
  }

  function toggleLocale() {
    locale.value = locale.value === 'zh' ? 'en' : 'zh'
    localStorage.setItem('locale', locale.value)
  }

  const localeLabel = computed(() => locale.value === 'zh' ? 'English' : '中文')

  return { t, locale, localeLabel, toggleLocale }
}
