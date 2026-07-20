import { apiRequest } from '@/api/client'
import type { AuthUser, LoginRequest, LoginResponse, SetPasswordRequest } from '@/types/auth'

export function login(credentials: LoginRequest): Promise<LoginResponse> {
  return apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
}

export function setPassword(payload: SetPasswordRequest): Promise<LoginResponse> {
  return apiRequest<LoginResponse>('/auth/set-password', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function fetchProfile(): Promise<AuthUser> {
  return apiRequest<AuthUser>('/auth/me')
}
