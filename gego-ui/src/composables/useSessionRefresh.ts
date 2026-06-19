import { onMounted, onUnmounted, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { useAuthStore } from '@/stores/auth'

export function useSessionRefresh(): void {
  const authStore = useAuthStore()
  const { isAuthenticated } = storeToRefs(authStore)

  function onVisibilityChange(): void {
    if (document.visibilityState === 'visible' && isAuthenticated.value) {
      authStore.refreshIfNeeded()
    }
  }

  watch(
    isAuthenticated,
    (authed) => {
      if (authed) {
        authStore.scheduleRefresh()
        authStore.refreshIfNeeded()
      } else {
        authStore.clearRefreshTimer()
      }
    },
    { immediate: true },
  )

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibilityChange)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    authStore.clearRefreshTimer()
  })
}
