import { queryOptions, useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import { fetchHealth, fetchBrandCitationDomains, fetchLLMs, fetchStats, fetchURLStats } from '@/api/dashboard'
import {
  brandCitationTargetKey,
  type BrandCitationTarget,
} from '@/lib/brand-citation-target'

export const dashboardQueryKeys = {
  all: ['dashboard'] as const,
  health: ['dashboard', 'health'] as const,
  stats: ['dashboard', 'stats'] as const,
  urlStats: ['dashboard', 'url-stats'] as const,
  brandCitationDomains: ['dashboard', 'brand-citation-domains'] as const,
  llms: ['dashboard', 'llms'] as const,
}

function sortedTags(tags: string[]): string[] {
  return [...tags].sort((a, b) => a.localeCompare(b))
}

export function healthQueryOptions() {
  return queryOptions({
    queryKey: dashboardQueryKeys.health,
    queryFn: fetchHealth,
    staleTime: 30_000,
  })
}

export function statsQueryOptions(tags: string[] = []) {
  const normalizedTags = sortedTags(tags)
  return queryOptions({
    queryKey: [...dashboardQueryKeys.stats, { tags: normalizedTags }] as const,
    queryFn: () => fetchStats(undefined, normalizedTags),
    staleTime: 60_000,
  })
}

export function urlStatsQueryOptions(tags: string[] = []) {
  const normalizedTags = sortedTags(tags)
  return queryOptions({
    queryKey: [...dashboardQueryKeys.urlStats, { tags: normalizedTags }] as const,
    queryFn: () => fetchURLStats(undefined, normalizedTags),
    staleTime: 60_000,
  })
}

export function llmsQueryOptions() {
  return queryOptions({
    queryKey: dashboardQueryKeys.llms,
    queryFn: fetchLLMs,
    staleTime: 5 * 60_000,
  })
}

export function useHealthQuery(enabled: MaybeRefOrGetter<boolean> = true) {
  return useQuery({
    queryKey: dashboardQueryKeys.health,
    queryFn: fetchHealth,
    staleTime: 30_000,
    enabled: () => toValue(enabled),
  })
}

export function useStatsQuery(tags: MaybeRefOrGetter<string[]> = () => []) {
  return useQuery({
    queryKey: computed(() => [...dashboardQueryKeys.stats, { tags: sortedTags(toValue(tags)) }] as const),
    queryFn: () => fetchStats(undefined, sortedTags(toValue(tags))),
    staleTime: 60_000,
  })
}

export function useURLStatsQuery(tags: MaybeRefOrGetter<string[]> = () => []) {
  return useQuery({
    queryKey: computed(() => [...dashboardQueryKeys.urlStats, { tags: sortedTags(toValue(tags)) }] as const),
    queryFn: () => fetchURLStats(undefined, sortedTags(toValue(tags))),
    staleTime: 60_000,
  })
}

export function useLLMsQuery() {
  return useQuery(llmsQueryOptions())
}

export function useBrandCitationDomainsQuery(
  target: MaybeRefOrGetter<BrandCitationTarget | null | undefined>,
  tags: MaybeRefOrGetter<string[]> = () => [],
) {
  return useQuery({
    queryKey: computed(() => {
      const value = toValue(target)
      return [
        ...dashboardQueryKeys.brandCitationDomains,
        {
          target: value ? brandCitationTargetKey(value) : '',
          tags: sortedTags(toValue(tags)),
        },
      ] as const
    }),
    queryFn: () => {
      const value = toValue(target)
      if (!value) {
        throw new Error('brand citation target is required')
      }
      return fetchBrandCitationDomains(value, undefined, sortedTags(toValue(tags)))
    },
    enabled: () => Boolean(toValue(target)),
    staleTime: 60_000,
  })
}
