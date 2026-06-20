import type { NavigationGuard } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

export const permissionGuard: NavigationGuard = (to) => {
  const permissions = to.meta.permissions
  if (!permissions?.length) {
    return true
  }

  const authStore = useAuthStore()
  if (!authStore.hasPermissions(permissions)) {
    return { name: 'forbidden' }
  }

  return true
}
