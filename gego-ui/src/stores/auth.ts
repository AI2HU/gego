import { useLocalStorage } from '@vueuse/core'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { canAccessRoute, canAccessRoutes, type RoutePermission } from '@/auth/permissions'
import { queryClient } from '@/lib/query-client'
import { authQueryKeys, profileQueryOptions } from '@/queries/auth'
import type { AuthUser } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = useLocalStorage<string | null>('gego_access_token', null)
  const user = ref<AuthUser | null>(null)
  const initialized = ref(false)

  const isAuthenticated = computed(() => user.value !== null)

  async function ensureSession(): Promise<void> {
    if (initialized.value) {
      return
    }

    if (!accessToken.value) {
      initialized.value = true
      return
    }

    try {
      user.value = await queryClient.fetchQuery(profileQueryOptions())
    } catch {
      clearSession()
    } finally {
      initialized.value = true
    }
  }

  function setUser(nextUser: AuthUser): void {
    user.value = nextUser
    queryClient.setQueryData(authQueryKeys.profile, nextUser)
  }

  function setSession(token: string, nextUser: AuthUser): void {
    accessToken.value = token
    setUser(nextUser)
  }

  function clearSession(): void {
    accessToken.value = null
    user.value = null
    queryClient.removeQueries({ queryKey: authQueryKeys.all })
  }

  function logout(): void {
    clearSession()
  }

  function hasPermission(permission: RoutePermission): boolean {
    if (!user.value) {
      return false
    }

    return canAccessRoute(user.value.role, permission)
  }

  function hasPermissions(permissions: RoutePermission[]): boolean {
    if (!user.value) {
      return false
    }

    return canAccessRoutes(user.value.role, permissions)
  }

  return {
    accessToken,
    user,
    initialized,
    isAuthenticated,
    ensureSession,
    setSession,
    setUser,
    logout,
    hasPermission,
    hasPermissions,
  }
})
