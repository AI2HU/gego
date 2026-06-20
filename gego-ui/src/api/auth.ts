import { apiRequest } from '@/api/client'
import type { AuthUser, LoginRequest, LoginResponse } from '@/types/auth'

export function login(credentials: LoginRequest): Promise<LoginResponse> {
  return apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
}

export function fetchProfile(): Promise<AuthUser> {
  return apiRequest<AuthUser>('/auth/me')
}
