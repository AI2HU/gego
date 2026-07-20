import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import { fetchSMTPSettings, testSMTPSettings, updateSMTPSettings } from '@/api/settings'
import type { TestSMTPRequest, UpdateSMTPSettingsRequest } from '@/types/settings'

export const settingsQueryKeys = {
  all: ['settings'] as const,
  smtp: ['settings', 'smtp'] as const,
}

export function smtpSettingsQueryOptions() {
  return queryOptions({
    queryKey: settingsQueryKeys.smtp,
    queryFn: fetchSMTPSettings,
    staleTime: 30_000,
  })
}

export function useSMTPSettingsQuery() {
  return useQuery(smtpSettingsQueryOptions())
}

export function useUpdateSMTPSettingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: UpdateSMTPSettingsRequest) => updateSMTPSettings(payload),
    onSuccess: (data) => {
      queryClient.setQueryData(settingsQueryKeys.smtp, data)
    },
  })
}

export function useTestSMTPSettingsMutation() {
  return useMutation({
    mutationFn: (payload: TestSMTPRequest) => testSMTPSettings(payload),
  })
}
