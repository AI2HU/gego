import { queryOptions, useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { toValue } from 'vue'

import { fetchHealth, fetchLLMs, fetchStats, fetchURLStats } from '@/api/dashboard'

export const dashboardQueryKeys = {
  all: ['dashboard'] as const,
  health: ['dashboard', 'health'] as const,
  stats: ['dashboard', 'stats'] as const,
  urlStats: ['dashboard', 'url-stats'] as const,
  llms: ['dashboard', 'llms'] as const,
}

export function healthQueryOptions() {
  return queryOptions({
    queryKey: dashboardQueryKeys.health,
    queryFn: fetchHealth,
    staleTime: 30_000,
  })
}

export function statsQueryOptions() {
  return queryOptions({
    queryKey: dashboardQueryKeys.stats,
    queryFn: () => fetchStats(),
    staleTime: 60_000,
  })
}

export function urlStatsQueryOptions() {
  return queryOptions({
    queryKey: dashboardQueryKeys.urlStats,
    queryFn: () => fetchURLStats(),
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

export function useStatsQuery() {
  return useQuery(statsQueryOptions())
}

export function useURLStatsQuery() {
  return useQuery(urlStatsQueryOptions())
}

export function useLLMsQuery() {
  return useQuery(llmsQueryOptions())
}
