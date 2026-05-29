import axios, { type AxiosError } from 'axios'
import type { APIResponse } from '../types/api'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

// Response interceptor: unwrap API envelope
client.interceptors.response.use(
  (response) => {
    const body = response.data as APIResponse
    if (body.success === false) {
      return Promise.reject(new Error(body.error || 'Unknown error'))
    }
    return response
  },
  (error: AxiosError<APIResponse>) => {
    const message = error.response?.data?.error || error.message
    return Promise.reject(new Error(message))
  }
)

// Helper to unwrap data from API response
export function unwrap<T>(response: { data: APIResponse<T> }): T {
  return response.data.data
}

export function unwrapWithMeta<T>(response: { data: APIResponse<T> }): {
  data: T
  meta: { total: number; page: number; limit: number }
} {
  return {
    data: response.data.data,
    meta: response.data.meta || { total: 0, page: 1, limit: 50 },
  }
}

export default client
