import { apiRequest } from '@/api/client'
import type { HealthResponse } from '@/types/health'
import type { ModelResponse } from '@/types/model'
import type { BrandCitationDomainsResponse, StatsResponse, URLStatsResponse } from '@/types/stats'
import type { BrandCitationTarget } from '@/lib/brand-citation-target'

const STATS_LIMIT = 20

function appendTags(params: URLSearchParams, tags: string[]): void {
  for (const tag of tags) {
    params.append('tags', tag)
  }
}

export function fetchHealth(): Promise<HealthResponse> {
  return apiRequest<HealthResponse>('/health')
}

export function fetchStats(keywordLimit = STATS_LIMIT, tags: string[] = []): Promise<StatsResponse> {
  const params = new URLSearchParams({ keyword_limit: String(keywordLimit) })
  appendTags(params, tags)
  return apiRequest<StatsResponse>(`/stats?${params.toString()}`)
}

export function fetchURLStats(limit = STATS_LIMIT, tags: string[] = []): Promise<URLStatsResponse> {
  const params = new URLSearchParams({ limit: String(limit) })
  appendTags(params, tags)
  return apiRequest<URLStatsResponse>(`/stats/urls?${params.toString()}`)
}

export function fetchBrandCitationDomains(
  target: BrandCitationTarget,
  limit = STATS_LIMIT,
  tags: string[] = [],
): Promise<BrandCitationDomainsResponse> {
  const params = new URLSearchParams({ limit: String(limit) })
  if (target.kind === 'brand') {
    params.set('brand_id', target.brandId)
  } else {
    params.set('keyword', target.keyword)
  }
  appendTags(params, tags)
  return apiRequest<BrandCitationDomainsResponse>(`/stats/brand-citation-domains?${params.toString()}`)
}

export function fetchLLMs(): Promise<ModelResponse[]> {
  return apiRequest<ModelResponse[]>('/models')
}
