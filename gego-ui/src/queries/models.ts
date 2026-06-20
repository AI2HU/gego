import { computed } from 'vue'
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import {
  createModel,
  deleteModel,
  fetchModels,
  fetchProviderApiKeys,
  fetchProviderModels,
  fetchProviders,
} from '@/api/models'
import { dashboardQueryKeys } from '@/queries/dashboard'
import type { CreateModelRequest, ListProviderModelsRequest } from '@/types/model'

export const modelsQueryKeys = {
  all: ['models'] as const,
  list: ['models', 'list'] as const,
  providers: ['models', 'providers'] as const,
  providerKeys: (provider: string) => ['models', 'provider-keys', provider] as const,
  providerModels: (provider: string, payload: ListProviderModelsRequest) =>
    ['models', 'provider-models', provider, payload] as const,
}

export function modelsListQueryOptions() {
  return queryOptions({
    queryKey: modelsQueryKeys.list,
    queryFn: fetchModels,
    staleTime: 30_000,
  })
}

export function providersQueryOptions() {
  return queryOptions({
    queryKey: modelsQueryKeys.providers,
    queryFn: fetchProviders,
    staleTime: 5 * 60_000,
  })
}

export function useModelsQuery() {
  return useQuery(modelsListQueryOptions())
}

export function useProvidersQuery() {
  return useQuery(providersQueryOptions())
}

export function useProviderApiKeysQuery(provider: () => string | null) {
  const providerRef = computed(provider)

  return useQuery({
    queryKey: computed(() => modelsQueryKeys.providerKeys(providerRef.value ?? '')),
    queryFn: () => fetchProviderApiKeys(providerRef.value!),
    enabled: computed(() => providerRef.value !== null),
    staleTime: 30_000,
  })
}

export function useCreateModelMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateModelRequest) => createModel(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: modelsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.llms })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useDeleteModelMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteModel(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: modelsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.llms })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export async function discoverProviderModels(
  provider: string,
  payload: ListProviderModelsRequest,
): Promise<Awaited<ReturnType<typeof fetchProviderModels>>> {
  return fetchProviderModels(provider, payload)
}
