export interface ModelResponse {
  id: string
  name: string
  provider: string
  model: string
  api_key?: string
  base_url?: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface ProviderInfo {
  id: string
  display_name: string
  console_url?: string
  requires_api_key: boolean
  requires_base_url: boolean
}

export interface ProviderApiKey {
  index: number
  masked: string
}

export interface ModelInfo {
  id: string
  name: string
  description?: string
  used_in_chat?: boolean
}

export interface CreateModelRequest {
  name: string
  provider: string
  model: string
  api_key?: string
  existing_key_index?: number
  base_url?: string
  enabled?: boolean
}

export interface ListProviderModelsRequest {
  api_key?: string
  existing_key_index?: number
  base_url?: string
}

export interface ProviderDistribution {
  provider: string
  totalResponses: number
  avgTokens: number
  modelCount: number
  percentage: number
}
