export interface ErrorLogEntry {
  id: string
  prompt_id: string
  prompt_text: string
  llm_id: string
  llm_name: string
  llm_provider: string
  llm_model: string
  error: string
  schedule_id?: string
  temperature?: number
  created_at: string
}

export interface PaginatedErrorLogsResponse {
  data: ErrorLogEntry[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export interface ErrorLogsQueryParams {
  page?: number
  limit?: number
  llm_id?: string
}
