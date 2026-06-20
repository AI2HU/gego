export interface PromptResponse {
  id: string
  template: string
  tags?: string[]
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreatePromptRequest {
  template: string
  tags?: string[]
  enabled?: boolean
}

export interface UpdatePromptRequest {
  template?: string
  tags?: string[]
  enabled?: boolean
}

export interface PaginatedPromptsResponse {
  data: PromptResponse[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export interface GeneratePromptsRequest {
  llm_id: string
  language_code: string
  user_input: string
  prompt_count?: number
}

export interface GeneratePromptsResponse {
  prompts: string[]
}

export interface BulkCreatePromptsRequest {
  prompts: { template: string }[]
  tags?: string[]
}

export interface BulkCreatePromptsResponse {
  prompts: PromptResponse[]
  saved_count: number
}
