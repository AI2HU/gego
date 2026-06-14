import { useAuthStore } from '@/stores/auth'
import type { ApiResponse } from '@/types/auth'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

function authHeaders(): HeadersInit {
  const token = useAuthStore().accessToken
  if (!token) {
    return {}
  }

  return { Authorization: `Bearer ${token}` }
}

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')

  for (const [key, value] of Object.entries(authHeaders())) {
    headers.set(key, value)
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  let body: ApiResponse<T> | null = null
  try {
    body = (await response.json()) as ApiResponse<T>
  } catch {
    throw new ApiError(response.status, 'Invalid server response')
  }

  if (!response.ok || !body.success) {
    throw new ApiError(response.status, body.error ?? 'Request failed')
  }

  return body.data as T
}
