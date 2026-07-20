export type Role = 'admin' | 'member'

export interface AuthUser {
  id: string
  username: string
  role: Role
  password_pending?: boolean
  created_at: string
}

export interface CreateUserRequest {
  email: string
  role: Role
}

export interface CreateUserResponse {
  user: AuthUser
  invite_url: string
  email_sent: boolean
}

export interface InviteUserResponse {
  user: AuthUser
  invite_url: string
  email_sent: boolean
}

export interface UpdateUserRequest {
  role?: Role
  password?: string
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

export interface SetPasswordRequest {
  token: string
  password: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}
