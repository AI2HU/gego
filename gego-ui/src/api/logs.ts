import { apiRequest } from '@/api/client'
import type { ErrorLogsQueryParams, PaginatedErrorLogsResponse } from '@/types/logs'

export function fetchErrorLogs(params: ErrorLogsQueryParams = {}): Promise<PaginatedErrorLogsResponse> {
  const searchParams = new URLSearchParams()
  searchParams.set('page', String(params.page ?? 1))
  searchParams.set('limit', String(params.limit ?? 20))

  if (params.llm_id) {
    searchParams.set('llm_id', params.llm_id)
  }

  return apiRequest<PaginatedErrorLogsResponse>(`/logs/errors?${searchParams.toString()}`)
}
