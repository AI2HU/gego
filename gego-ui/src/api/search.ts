import { apiRequest } from '@/api/client'
import type { SearchRequest, SearchResponse } from '@/types/search'

export function searchResponses(payload: SearchRequest): Promise<SearchResponse> {
  return apiRequest<SearchResponse>('/search', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
