import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchRequiredUpgrades, runUpgrade } from '@/api/upgrades'
import type { UpgradeItem } from '@/types/upgrade'

export const useUpgradesStore = defineStore('upgrades', () => {
  const upgrades = ref<UpgradeItem[]>([])
  const loading = ref(false)
  const running = ref(false)
  const error = ref<string | null>(null)
  const lastRunMessage = ref<string | null>(null)
  const restartRequired = ref(false)
  const loaded = ref(false)

  const majorUpgrades = computed(() => upgrades.value.filter((item) => item.severity === 'major'))
  const minorUpgrades = computed(() => upgrades.value.filter((item) => item.severity === 'minor'))
  const hasMajorUpgrade = computed(() => majorUpgrades.value.length > 0)
  const hasMinorUpgrade = computed(() => minorUpgrades.value.length > 0)

  async function refresh() {
    loading.value = true
    error.value = null
    try {
      upgrades.value = await fetchRequiredUpgrades()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to check upgrades'
    } finally {
      loading.value = false
      loaded.value = true
    }
  }

  async function execute(code: string) {
    running.value = true
    error.value = null
    try {
      const result = await runUpgrade(code)
      lastRunMessage.value = result.message
      restartRequired.value = result.restart_required
      await refresh()
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Upgrade failed'
      throw err
    } finally {
      running.value = false
    }
  }

  return {
    upgrades,
    loading,
    running,
    error,
    lastRunMessage,
    restartRequired,
    loaded,
    majorUpgrades,
    minorUpgrades,
    hasMajorUpgrade,
    hasMinorUpgrade,
    refresh,
    execute,
  }
})
