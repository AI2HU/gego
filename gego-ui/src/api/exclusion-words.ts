import { apiRequest } from '@/api/client'
import type {
  CreateExclusionWordRequest,
  ExclusionWord,
  SuggestedBrandWord,
} from '@/types/exclusion-word'

export function fetchExclusionWords(): Promise<ExclusionWord[]> {
  return apiRequest<ExclusionWord[]>('/exclusion-words')
}

export function createExclusionWord(payload: CreateExclusionWordRequest): Promise<ExclusionWord> {
  return apiRequest<ExclusionWord>('/exclusion-words', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function deleteExclusionWord(id: string): Promise<void> {
  return apiRequest<void>(`/exclusion-words/${id}`, { method: 'DELETE' })
}

function appendTags(params: URLSearchParams, tags: string[]): void {
  for (const tag of tags) {
    params.append('tags', tag)
  }
}

export function fetchSuggestedBrandWords(
  limit = 50,
  tags: string[] = [],
): Promise<SuggestedBrandWord[]> {
  const params = new URLSearchParams({ limit: String(limit) })
  appendTags(params, tags)
  return apiRequest<SuggestedBrandWord[]>(`/exclusion-words/suggestions?${params.toString()}`)
}
