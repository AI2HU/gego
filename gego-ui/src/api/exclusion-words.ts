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

export function fetchSuggestedBrandWords(limit = 50): Promise<SuggestedBrandWord[]> {
  return apiRequest<SuggestedBrandWord[]>(`/exclusion-words/suggestions?limit=${limit}`)
}
