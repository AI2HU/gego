import type { NavigationGuard } from 'vue-router'

import { useUpgradesStore } from '@/stores/upgrades'

export const upgradeGuard: NavigationGuard = async (to) => {
  const upgradesStore = useUpgradesStore()

  if (!upgradesStore.loaded) {
    await upgradesStore.refresh()
  }

  if (upgradesStore.hasMajorUpgrade && to.name !== 'upgrade') {
    return { name: 'upgrade' }
  }

  if (!upgradesStore.hasMajorUpgrade && to.name === 'upgrade') {
    return { name: 'login' }
  }

  return true
}
