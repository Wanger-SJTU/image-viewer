import client, { unwrap } from './client'
import type { ScanRequest } from '../types/api'

export interface ScanProgress {
  phase: string
  found: number
  processed: number
  matched: number
  orphans: number
  error: string
}

export async function startScan(path: string): Promise<{ path: string; status: string }> {
  const body: ScanRequest = { path }
  const resp = await client.post('/scan', body)
  return unwrap(resp)
}

export async function getScanStatus(): Promise<ScanProgress> {
  const resp = await client.get('/scan/status')
  return unwrap<ScanProgress>(resp)
}
