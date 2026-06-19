import { apiRequest } from '@/api/client'
import type { HealthResponse } from '@/types/health'
import type { ModelResponse } from '@/types/model'
import type { StatsResponse, URLStatsResponse } from '@/types/stats'

const STATS_LIMIT = 20

export function fetchHealth(): Promise<HealthResponse> {
  return apiRequest<HealthResponse>('/health')
}

export function fetchStats(keywordLimit = STATS_LIMIT): Promise<StatsResponse> {
  return apiRequest<StatsResponse>(`/stats?keyword_limit=${keywordLimit}`)
}

export function fetchURLStats(limit = STATS_LIMIT): Promise<URLStatsResponse> {
  return apiRequest<URLStatsResponse>(`/stats/urls?limit=${limit}`)
}

export function fetchLLMs(): Promise<ModelResponse[]> {
  return apiRequest<ModelResponse[]>('/models')
}
