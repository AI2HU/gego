import { queryOptions, useQuery } from '@tanstack/vue-query'
import { computed, type Ref } from 'vue'

import { fetchErrorLogs } from '@/api/logs'
import type { ErrorLogsQueryParams } from '@/types/logs'

export const logsQueryKeys = {
  all: ['logs'] as const,
  errors: (params: ErrorLogsQueryParams) => ['logs', 'errors', params] as const,
}

export function errorLogsQueryOptions(params: ErrorLogsQueryParams) {
  return queryOptions({
    queryKey: logsQueryKeys.errors(params),
    queryFn: () => fetchErrorLogs(params),
    staleTime: 15_000,
  })
}

export function useErrorLogsQuery(params: Ref<ErrorLogsQueryParams>) {
  return useQuery({
    queryKey: computed(() => logsQueryKeys.errors(params.value)),
    queryFn: () => fetchErrorLogs(params.value),
    staleTime: 15_000,
  })
}
