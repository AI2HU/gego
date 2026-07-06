export interface SearchRequest {
  keyword: string
  tags?: string[]
  start_time?: string
  end_time?: string
  limit?: number
}

export interface SearchURL {
  search_query?: string
  url: string
  title?: string
  citation_index: number
}

export interface SearchResponseItem {
  id: string
  prompt_id: string
  prompt_text: string
  llm_id: string
  llm_name: string
  llm_provider: string
  llm_model: string
  response_text: string
  search_urls?: SearchURL[]
  temperature?: number
  created_at: string
}

export interface SearchResponse {
  keyword: string
  search_terms?: string[]
  total_responses: number
  total_mentions: number
  unique_prompts: number
  unique_llms: number
  by_prompt: Record<string, number>
  by_llm: Record<string, number>
  by_provider: Record<string, number>
  first_seen: string
  last_seen: string
  responses?: SearchResponseItem[]
}

export interface SearchMatch {
  responseId: string
  promptId: string
  promptName: string
  promptTags: string[]
  responseText: string
  searchUrls: SearchURL[]
  llmName: string
  llmProvider: string
  temperature: number
  keyword: string
  createdAt: string
}
