import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import { fetchProfile, login } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import type { LoginRequest } from '@/types/auth'

export const authQueryKeys = {
  all: ['auth'] as const,
  profile: ['auth', 'profile'] as const,
}

export function profileQueryOptions() {
  return queryOptions({
    queryKey: authQueryKeys.profile,
    queryFn: fetchProfile,
    enabled: () => Boolean(useAuthStore().accessToken),
    staleTime: 5 * 60 * 1000,
  })
}

export function useProfileQuery() {
  return useQuery(profileQueryOptions())
}

export function useLoginMutation() {
  const queryClient = useQueryClient()
  const authStore = useAuthStore()

  return useMutation({
    mutationFn: (credentials: LoginRequest) => login(credentials),
    onSuccess: (response) => {
      authStore.setSession(response.access_token, response.user)
      queryClient.setQueryData(authQueryKeys.profile, response.user)
    },
  })
}
