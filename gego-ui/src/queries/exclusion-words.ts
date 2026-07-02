import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import {
  createExclusionWord,
  deleteExclusionWord,
  fetchExclusionWords,
  fetchSuggestedBrandWords,
} from '@/api/exclusion-words'
import { dashboardQueryKeys } from '@/queries/dashboard'
import type { CreateExclusionWordRequest } from '@/types/exclusion-word'

export const exclusionWordsQueryKeys = {
  all: ['exclusion-words'] as const,
  list: ['exclusion-words', 'list'] as const,
  suggestions: (limit: number) => ['exclusion-words', 'suggestions', limit] as const,
}

export function exclusionWordsListQueryOptions() {
  return queryOptions({
    queryKey: exclusionWordsQueryKeys.list,
    queryFn: fetchExclusionWords,
    staleTime: 30_000,
  })
}

export function suggestedBrandWordsQueryOptions(limit = 50) {
  return queryOptions({
    queryKey: exclusionWordsQueryKeys.suggestions(limit),
    queryFn: () => fetchSuggestedBrandWords(limit),
    staleTime: 30_000,
  })
}

export function useExclusionWordsQuery() {
  return useQuery(exclusionWordsListQueryOptions())
}

export function useSuggestedBrandWordsQuery(limit = 50) {
  return useQuery(suggestedBrandWordsQueryOptions(limit))
}

export function useCreateExclusionWordMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateExclusionWordRequest) => createExclusionWord(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: exclusionWordsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useDeleteExclusionWordMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteExclusionWord(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: exclusionWordsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}
