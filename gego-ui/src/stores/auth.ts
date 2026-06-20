import { useLocalStorage } from '@vueuse/core'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { logoutSession, refreshAccessToken } from '@/api/auth-session'
import { canAccessRoute, canAccessRoutes, type RoutePermission } from '@/auth/permissions'
import { decodeJwtExp } from '@/lib/jwt'
import { queryClient } from '@/lib/query-client'
import { authQueryKeys, profileQueryOptions } from '@/queries/auth'
import router from '@/router'
import type { AuthUser } from '@/types/auth'

const REFRESH_BUFFER_MS = 60_000

export const useAuthStore = defineStore('auth', () => {
  const accessToken = useLocalStorage<string | null>('gego_access_token', null)
  const expiresAt = useLocalStorage<number | null>('gego_token_expires_at', null)
  const user = ref<AuthUser | null>(null)
  const initialized = ref(false)

  let refreshTimer: ReturnType<typeof setTimeout> | null = null
  let refreshPromise: Promise<void> | null = null

  const isAuthenticated = computed(() => user.value !== null)

  function resolveExpiresAt(token: string, expiresIn?: number): number {
    if (expiresIn) {
      return Date.now() + expiresIn * 1000
    }

    const jwtExp = decodeJwtExp(token)
    if (jwtExp) {
      return jwtExp
    }

    return Date.now() + 15 * 60 * 1000
  }

  function isTokenExpiringSoon(): boolean {
    if (!accessToken.value) {
      return false
    }

    const expiry = expiresAt.value ?? decodeJwtExp(accessToken.value)
    if (!expiry) {
      return true
    }

    return Date.now() >= expiry - REFRESH_BUFFER_MS
  }

  function clearRefreshTimer(): void {
    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
  }

  function scheduleRefresh(): void {
    clearRefreshTimer()

    if (!accessToken.value) {
      return
    }

    const expiry = expiresAt.value ?? decodeJwtExp(accessToken.value)
    if (!expiry) {
      return
    }

    const delay = Math.max(expiry - Date.now() - REFRESH_BUFFER_MS, 0)
    refreshTimer = setTimeout(() => {
      refreshSession().catch(() => {
        handleSessionExpired()
      })
    }, delay)
  }

  async function refreshSession(): Promise<void> {
    if (refreshPromise) {
      return refreshPromise
    }

    refreshPromise = (async () => {
      const response = await refreshAccessToken()
      setSession(response.access_token, response.user, response.expires_in)
    })().finally(() => {
      refreshPromise = null
    })

    return refreshPromise
  }

  function refreshIfNeeded(): Promise<void> | undefined {
    if (!accessToken.value || !isTokenExpiringSoon()) {
      return
    }

    return refreshSession().catch(() => {
      handleSessionExpired()
    })
  }

  async function ensureSession(): Promise<void> {
    if (initialized.value) {
      return
    }

    if (!accessToken.value) {
      initialized.value = true
      return
    }

    try {
      if (isTokenExpiringSoon()) {
        await refreshSession()
      }

      user.value = await queryClient.fetchQuery(profileQueryOptions())
      scheduleRefresh()
    } catch {
      try {
        await refreshSession()
        user.value = await queryClient.fetchQuery(profileQueryOptions())
        scheduleRefresh()
      } catch {
        clearSession()
      }
    } finally {
      initialized.value = true
    }
  }

  function setUser(nextUser: AuthUser): void {
    user.value = nextUser
    queryClient.setQueryData(authQueryKeys.profile, nextUser)
  }

  function setSession(token: string, nextUser: AuthUser, expiresIn?: number): void {
    accessToken.value = token
    expiresAt.value = resolveExpiresAt(token, expiresIn)
    setUser(nextUser)
    scheduleRefresh()
  }

  function clearSession(): void {
    clearRefreshTimer()
    accessToken.value = null
    expiresAt.value = null
    user.value = null
    queryClient.removeQueries({ queryKey: authQueryKeys.all })
  }

  function handleSessionExpired(): void {
    const redirect = router.currentRoute.value.fullPath
    clearSession()

    if (router.currentRoute.value.name !== 'login') {
      void router.push({
        name: 'login',
        query: redirect !== '/login' ? { redirect } : undefined,
      })
    }
  }

  async function logout(): Promise<void> {
    clearRefreshTimer()

    try {
      await logoutSession()
    } finally {
      clearSession()
    }
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
    expiresAt,
    user,
    initialized,
    isAuthenticated,
    ensureSession,
    setSession,
    setUser,
    refreshSession,
    refreshIfNeeded,
    scheduleRefresh,
    clearRefreshTimer,
    handleSessionExpired,
    logout,
    hasPermission,
    hasPermissions,
  }
})
