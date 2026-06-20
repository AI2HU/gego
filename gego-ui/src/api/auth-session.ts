import type { ApiResponse, AuthUser, LoginResponse } from '@/types/auth'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export class AuthSessionError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'AuthSessionError'
    this.status = status
  }
}

async function parseResponse<T>(response: Response): Promise<T> {
  let body: ApiResponse<T> | null = null

  try {
    body = (await response.json()) as ApiResponse<T>
  } catch {
    throw new AuthSessionError(response.status, 'Invalid server response')
  }

  if (!response.ok || !body.success) {
    throw new AuthSessionError(response.status, body.error ?? 'Request failed')
  }

  return body.data as T
}

export function refreshAccessToken(): Promise<LoginResponse> {
  return fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  }).then((response) => parseResponse<LoginResponse>(response))
}

export function logoutSession(): Promise<void> {
  return fetch(`${API_BASE}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  }).then(() => undefined)
}

export type { AuthUser }
