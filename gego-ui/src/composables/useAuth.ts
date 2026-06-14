import { storeToRefs } from 'pinia'
import { computed } from 'vue'

import { useLoginMutation, useProfileQuery } from '@/queries/auth'
import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const authStore = useAuthStore()
  const { user, isAuthenticated, initialized } = storeToRefs(authStore)
  const loginMutation = useLoginMutation()
  const profileQuery = useProfileQuery()

  const isLoggingIn = computed(() => loginMutation.isPending.value)

  return {
    user,
    isAuthenticated,
    initialized,
    isLoggingIn,
    profileQuery,
    login: loginMutation.mutateAsync,
    logout: authStore.logout,
    hasPermission: authStore.hasPermission,
    hasPermissions: authStore.hasPermissions,
  }
}
