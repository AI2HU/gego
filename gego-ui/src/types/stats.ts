export interface KeywordCount {
  keyword: string
  count: number
}

export interface TimeSeriesPoint {
  timestamp: string
  count: number
}

export interface BrandTrendSeries {
  keyword: string
  points: TimeSeriesPoint[]
}

export interface PromptStats {
  prompt_id: string
  total_responses: number
  unique_llms: number
  llm_counts: Record<string, number>
  avg_tokens: number
  updated_at: string
}

export interface LlmUsageStats {
  llm_id: string
  total_responses: number
  unique_prompts: number
  prompt_counts: Record<string, number>
  avg_tokens: number
  updated_at: string
}

export interface StatsResponse {
  total_responses: number
  total_prompts: number
  total_llms: number
  total_schedules: number
  top_keywords: KeywordCount[]
  brand_trends: BrandTrendSeries[]
  prompt_stats: PromptStats[]
  llm_stats: LlmUsageStats[]
  last_updated: string
}

export interface DomainMentionStats {
  domain: string
  citations: number
  unique_url_count: number
}

export interface URLStatsResponse {
  top_urls: unknown[]
  top_domains: DomainMentionStats[]
}

export interface BrandCitationDomainStats {
  domain: string
  citations: number
  unique_url_count: number
}

export interface BrandCitationDomainsResponse {
  brand_id: string
  brand_name: string
  total_hits: number
  domains: BrandCitationDomainStats[]
}
