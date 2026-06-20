export type Role = 'admin' | 'member'

export interface AuthUser {
  id: string
  username: string
  role: Role
  created_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}
