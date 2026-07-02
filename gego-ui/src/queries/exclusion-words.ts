import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import {
  createExclusionWord,
  deleteExclusionWord,
  fetchExclusionWords,
  fetchSuggestedBrandWords,
} from '@/api/exclusion-words'
import { dashboardQueryKeys } from '@/queries/dashboard'
import type { CreateExclusionWordRequest } from '@/types/exclusion-word'

function sortedTags(tags: string[]): string[] {
  return [...tags].sort((a, b) => a.localeCompare(b))
}

export const exclusionWordsQueryKeys = {
  all: ['exclusion-words'] as const,
  list: ['exclusion-words', 'list'] as const,
  suggestions: (limit: number, tags: string[] = []) =>
    ['exclusion-words', 'suggestions', limit, { tags: sortedTags(tags) }] as const,
}

export function exclusionWordsListQueryOptions() {
  return queryOptions({
    queryKey: exclusionWordsQueryKeys.list,
    queryFn: fetchExclusionWords,
    staleTime: 30_000,
  })
}

export function suggestedBrandWordsQueryOptions(limit = 50, tags: string[] = []) {
  const normalizedTags = sortedTags(tags)
  return queryOptions({
    queryKey: exclusionWordsQueryKeys.suggestions(limit, normalizedTags),
    queryFn: () => fetchSuggestedBrandWords(limit, normalizedTags),
    staleTime: 30_000,
  })
}

export function useExclusionWordsQuery() {
  return useQuery(exclusionWordsListQueryOptions())
}

export function useSuggestedBrandWordsQuery(
  limit: MaybeRefOrGetter<number> = 50,
  tags: MaybeRefOrGetter<string[]> = () => [],
) {
  return useQuery({
    queryKey: computed(() =>
      exclusionWordsQueryKeys.suggestions(toValue(limit), sortedTags(toValue(tags))),
    ),
    queryFn: () => fetchSuggestedBrandWords(toValue(limit), sortedTags(toValue(tags))),
    staleTime: 30_000,
  })
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
