import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import {
  createBrand,
  createBrandAlias,
  deleteBrand,
  deleteBrandAlias,
  fetchBrands,
  fetchSuggestedBrandWords,
  mapBrandFromDetection,
  updateBrand,
  updateBrandAlias,
} from '@/api/brands'
import { dashboardQueryKeys } from '@/queries/dashboard'
import { exclusionWordsQueryKeys } from '@/queries/exclusion-words'
import type {
  CreateBrandAliasRequest,
  CreateBrandRequest,
  MapBrandRequest,
  UpdateBrandAliasRequest,
  UpdateBrandRequest,
} from '@/types/brand'

function sortedTags(tags: string[]): string[] {
  return [...tags].sort((a, b) => a.localeCompare(b))
}

export const brandsQueryKeys = {
  all: ['brands'] as const,
  list: ['brands', 'list'] as const,
  suggestions: (limit: number, tags: string[] = []) =>
    ['brands', 'suggestions', limit, { tags: sortedTags(tags) }] as const,
}

export function brandsListQueryOptions() {
  return queryOptions({
    queryKey: brandsQueryKeys.list,
    queryFn: fetchBrands,
    staleTime: 30_000,
  })
}

export function useBrandsQuery() {
  return useQuery(brandsListQueryOptions())
}

export function useSuggestedBrandWordsQuery(
  limit: MaybeRefOrGetter<number> = 50,
  tags: MaybeRefOrGetter<string[]> = () => [],
) {
  return useQuery({
    queryKey: computed(() =>
      brandsQueryKeys.suggestions(toValue(limit), sortedTags(toValue(tags))),
    ),
    queryFn: () => fetchSuggestedBrandWords(toValue(limit), sortedTags(toValue(tags))),
    staleTime: 30_000,
  })
}

function invalidateBrandRelated(queryClient: ReturnType<typeof useQueryClient>) {
  void queryClient.invalidateQueries({ queryKey: brandsQueryKeys.all })
  void queryClient.invalidateQueries({ queryKey: exclusionWordsQueryKeys.all })
  void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
}

export function useCreateBrandMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateBrandRequest) => createBrand(payload),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useUpdateBrandMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateBrandRequest }) =>
      updateBrand(id, payload),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useDeleteBrandMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteBrand(id),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useCreateBrandAliasMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ brandId, payload }: { brandId: string; payload: CreateBrandAliasRequest }) =>
      createBrandAlias(brandId, payload),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useUpdateBrandAliasMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      brandId,
      aliasId,
      payload,
    }: {
      brandId: string
      aliasId: string
      payload: UpdateBrandAliasRequest
    }) => updateBrandAlias(brandId, aliasId, payload),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useDeleteBrandAliasMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ brandId, aliasId }: { brandId: string; aliasId: string }) =>
      deleteBrandAlias(brandId, aliasId),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}

export function useMapBrandMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: MapBrandRequest) => mapBrandFromDetection(payload),
    onSuccess: () => invalidateBrandRelated(queryClient),
  })
}
