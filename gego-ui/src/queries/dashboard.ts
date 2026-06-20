import { queryOptions, useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import { fetchHealth, fetchLLMs, fetchStats, fetchURLStats } from '@/api/dashboard'

export const dashboardQueryKeys = {
  all: ['dashboard'] as const,
  health: ['dashboard', 'health'] as const,
  stats: ['dashboard', 'stats'] as const,
  urlStats: ['dashboard', 'url-stats'] as const,
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
