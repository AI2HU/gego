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

type ApiRequestOptions = RequestInit & {
  skipAuthRetry?: boolean
}

function authHeaders(): HeadersInit {
  const token = useAuthStore().accessToken
  if (!token) {
    return {}
  }

  return { Authorization: `Bearer ${token}` }
}

function shouldRetryAuth(path: string): boolean {
  return !path.startsWith('/auth/login')
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const { skipAuthRetry, ...fetchOptions } = options

  const headers = new Headers(fetchOptions.headers)
  headers.set('Content-Type', 'application/json')

  for (const [key, value] of Object.entries(authHeaders())) {
    headers.set(key, value)
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...fetchOptions,
    headers,
    credentials: 'include',
  })

  if (response.status === 401 && !skipAuthRetry && shouldRetryAuth(path)) {
    const authStore = useAuthStore()

    try {
      await authStore.refreshSession()
    } catch {
      authStore.handleSessionExpired()
      throw new ApiError(401, 'Session expired')
    }

    return apiRequest<T>(path, { ...options, skipAuthRetry: true })
  }

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
