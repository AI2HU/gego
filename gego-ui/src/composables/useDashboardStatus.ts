import { useIsFetching, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { dashboardQueryKeys, useHealthQuery } from '@/queries/dashboard'

export function useDashboardStatus() {
  const route = useRoute()
  const queryClient = useQueryClient()
  const isDashboard = computed(() => route.name === 'dashboard')

  const healthQuery = useHealthQuery(isDashboard)

  const fetchCount = useIsFetching({
    queryKey: dashboardQueryKeys.all,
  })

  const connected = computed(() => {
    if (!isDashboard.value) {
      return true
    }
    return healthQuery.isSuccess.value
  })

  const connectionLabel = computed(() => {
    if (!isDashboard.value) {
      return undefined
    }
    if (healthQuery.isPending.value) {
      return 'Connecting...'
    }
    return connected.value ? 'Connected' : 'Disconnected'
  })

  const loading = computed(() => isDashboard.value && fetchCount.value > 0)

  function refresh() {
    if (!isDashboard.value) {
      return
    }
    queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.all })
  }

  return {
    isDashboard,
    connected,
    connectionLabel,
    loading,
    refresh,
  }
}
