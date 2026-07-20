import { apiRequest } from '@/api/client'
import type { AuthUser, CreateUserRequest, UpdateUserRequest } from '@/types/auth'

export function fetchUsers(): Promise<AuthUser[]> {
  return apiRequest<AuthUser[]>('/users')
}

export function createUser(payload: CreateUserRequest): Promise<AuthUser> {
  return apiRequest<AuthUser>('/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateUser(id: string, payload: UpdateUserRequest): Promise<AuthUser> {
  return apiRequest<AuthUser>(`/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteUser(id: string): Promise<void> {
  return apiRequest<void>(`/users/${id}`, { method: 'DELETE' })
}
