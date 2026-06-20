import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import {
  createPrompt,
  createPrompts,
  deletePrompt,
  fetchPrompts,
  generatePrompts,
  updatePrompt,
} from '@/api/prompts'
import { dashboardQueryKeys } from '@/queries/dashboard'
import type {
  BulkCreatePromptsRequest,
  CreatePromptRequest,
  GeneratePromptsRequest,
  UpdatePromptRequest,
} from '@/types/prompt'

export const promptsQueryKeys = {
  all: ['prompts'] as const,
  list: ['prompts', 'list'] as const,
}

export function promptsListQueryOptions() {
  return queryOptions({
    queryKey: promptsQueryKeys.list,
    queryFn: fetchPrompts,
    staleTime: 30_000,
  })
}

export function usePromptsQuery() {
  return useQuery(promptsListQueryOptions())
}

export function useCreatePromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreatePromptRequest) => createPrompt(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: promptsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useGeneratePromptsMutation() {
  return useMutation({
    mutationFn: (payload: GeneratePromptsRequest) => generatePrompts(payload),
  })
}

export function useCreatePromptsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: BulkCreatePromptsRequest) => createPrompts(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: promptsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useUpdatePromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdatePromptRequest }) =>
      updatePrompt(id, payload),
    onSuccess: (updatedPrompt) => {
      queryClient.setQueryData(
        promptsQueryKeys.list,
        (current: Awaited<ReturnType<typeof fetchPrompts>> | undefined) => {
          if (!current) {
            return current
          }

          return current.map((prompt) =>
            prompt.id === updatedPrompt.id ? updatedPrompt : prompt,
          )
        },
      )
      void queryClient.invalidateQueries({ queryKey: promptsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useDeletePromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deletePrompt(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: promptsQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}
