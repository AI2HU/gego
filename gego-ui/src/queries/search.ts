import { useMutation } from '@tanstack/vue-query'

import { searchResponses } from '@/api/search'
import type { SearchRequest, SearchResponse } from '@/types/search'

export const searchQueryKeys = {
  all: ['search'] as const,
}

export function useSearchMutation() {
  return useMutation({
    mutationKey: searchQueryKeys.all,
    mutationFn: (payload: SearchRequest) => searchResponses(payload),
  })
}

export type { SearchRequest, SearchResponse }
