<script setup lang="ts">
import { computed, onMounted } from 'vue'

import AppButton from '@/components/ui/AppButton.vue'
import { getUpgradeDoc } from '@/config/upgrade-docs'
import { useUpgradesStore } from '@/stores/upgrades'

const upgradesStore = useUpgradesStore()

const minorUpgrades = computed(() => upgradesStore.minorUpgrades)
const showBanner = computed(
  () => upgradesStore.hasMinorUpgrade || (upgradesStore.restartRequired && !upgradesStore.hasMajorUpgrade),
)

onMounted(() => {
  if (!upgradesStore.loaded) {
    void upgradesStore.refresh()
  }
})

async function confirmUpgrade(code: string) {
  await upgradesStore.execute(code)
}
</script>

<template>
  <div v-if="showBanner" class="border-b border-amber-200 bg-amber-50 px-6 py-4">
    <div v-for="item in minorUpgrades" :key="item.code" class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p class="font-semibold text-amber-900">{{ getUpgradeDoc(item.code).title }}</p>
        <p class="mt-1 text-sm text-amber-800">{{ getUpgradeDoc(item.code).summary }}</p>
        <p v-if="upgradesStore.error" class="mt-2 text-sm text-red-700">{{ upgradesStore.error }}</p>
      </div>
      <AppButton variant="primary" :loading="upgradesStore.running" @click="confirmUpgrade(item.code)">
        Run upgrade
      </AppButton>
    </div>

    <div v-if="upgradesStore.restartRequired && minorUpgrades.length === 0" class="text-sm text-amber-800">
      <p class="font-semibold text-amber-900">Restart required</p>
      <p class="mt-1">
        {{ upgradesStore.lastRunMessage ?? 'Restart the API and worker containers to apply the upgrade.' }}
      </p>
    </div>
  </div>
</template>
