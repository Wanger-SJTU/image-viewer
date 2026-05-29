import client, { unwrap } from './client'
import type { ScanRequest } from '../types/api'

export async function startScan(path: string): Promise<{ path: string; status: string }> {
  const body: ScanRequest = { path }
  const resp = await client.post('/scan', body)
  return unwrap(resp)
}

export function subscribeScanProgress(
  _onProgress: (progress: { phase: string; found: number; processed: number; matched: number; orphans: number }) => void,
  _onDone: () => void,
  _onError: (err: string) => void
): () => void {
  // SSE-based scan progress — to be implemented with EventSource
  // For now, return a no-op unsubscribe
  return () => {}
}
