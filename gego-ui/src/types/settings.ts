export interface SMTPSettings {
  host: string
  port: number
  username: string
  from_email: string
  from_name: string
  use_tls: boolean
  enabled: boolean
  has_password: boolean
}

export interface UpdateSMTPSettingsRequest {
  host: string
  port: number
  username: string
  password: string
  from_email: string
  from_name: string
  use_tls: boolean
  enabled: boolean
}

export interface TestSMTPRequest {
  to?: string
  host?: string
  port?: number
  username?: string
  password?: string
  from_email?: string
  from_name?: string
  use_tls?: boolean
}
