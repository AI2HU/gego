import type { NavigationGuard } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

export const authGuard: NavigationGuard = async (to) => {
  const authStore = useAuthStore()

  await authStore.ensureSession()

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return { name: 'dashboard' }
  }

  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)
  if (requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  return true
}
