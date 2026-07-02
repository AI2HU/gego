<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'

import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import { getUpgradeDoc } from '@/config/upgrade-docs'
import { appMeta } from '@/design/navigation'
import { header, page, typography } from '@/design/classes'
import { useUpgradesStore } from '@/stores/upgrades'

const router = useRouter()
const upgradesStore = useUpgradesStore()

const majorUpgrades = computed(() => upgradesStore.majorUpgrades)

async function runUpgrade(code: string) {
  await upgradesStore.execute(code)
  await router.replace({ name: 'login', query: { upgraded: '1' } })
}

async function refresh() {
  await upgradesStore.refresh()
  if (!upgradesStore.hasMajorUpgrade) {
    await router.replace({ name: 'login', query: { upgraded: '1' } })
  }
}

watch(
  () => upgradesStore.hasMajorUpgrade,
  (required) => {
    if (upgradesStore.loaded && !required) {
      void router.replace({ name: 'login', query: { upgraded: '1' } })
    }
  },
  { immediate: true },
)
</script>

<template>
  <div :class="page.root">
    <div class="flex min-h-screen items-center justify-center px-4 py-12">
      <div class="w-full max-w-2xl space-y-6">
        <div class="text-center">
          <h1 :class="[typography.h1, header.title]">{{ appMeta.name }}</h1>
          <p :class="[typography.body, 'mt-2 text-gray-600']">Upgrade required before you can use Gego.</p>
        </div>

        <AppCard v-for="item in majorUpgrades" :key="item.code" class="p-6">
          <h2 class="text-xl font-semibold text-gray-900">{{ getUpgradeDoc(item.code).title }}</h2>
          <p class="mt-2 text-sm text-gray-600">{{ getUpgradeDoc(item.code).summary }}</p>

          <ol class="mt-4 list-decimal space-y-2 pl-5 text-sm text-gray-700">
            <li v-for="(step, index) in getUpgradeDoc(item.code).steps" :key="index">{{ step }}</li>
          </ol>

          <div v-if="upgradesStore.restartRequired" class="mt-6 rounded-lg border border-green-200 bg-green-50 p-4">
            <p class="text-sm font-medium text-green-900">Upgrade completed</p>
            <p class="mt-1 text-sm text-green-800">
              {{ upgradesStore.lastRunMessage ?? 'Restart the API and worker containers to finish.' }}
            </p>
          </div>

          <p v-if="upgradesStore.error" class="mt-4 text-sm text-red-600">{{ upgradesStore.error }}</p>

          <div class="mt-6 flex flex-wrap gap-2">
            <AppButton variant="secondary" :loading="upgradesStore.loading" @click="refresh">Refresh status</AppButton>
            <AppButton
              v-if="!upgradesStore.restartRequired"
              variant="primary"
              :loading="upgradesStore.running"
              @click="runUpgrade(item.code)"
            >
              Run upgrade
            </AppButton>
          </div>
        </AppCard>
      </div>
    </div>
  </div>
</template>
