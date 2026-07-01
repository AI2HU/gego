<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { fetchRequiredUpgrades, runUpgrade } from '@/api/upgrades'
import AppButton from '@/components/ui/AppButton.vue'
import { UPGRADE_SQLITE_TO_POSTGRES } from '@/types/upgrade'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const loading = ref(false)
const running = ref(false)
const showConfirm = ref(false)
const requiredCodes = ref<string[]>([])
const error = ref<string | null>(null)
const successMessage = ref<string | null>(null)
const restartRequired = ref(false)

const isAdmin = computed(() => authStore.user?.role === 'admin')
const needsSQLiteUpgrade = computed(() => requiredCodes.value.includes(UPGRADE_SQLITE_TO_POSTGRES))
const showBanner = computed(() => isAdmin.value && (needsSQLiteUpgrade.value || restartRequired.value))

async function loadUpgrades() {
  if (!isAdmin.value) {
    return
  }
  loading.value = true
  error.value = null
  try {
    const status = await fetchRequiredUpgrades()
    requiredCodes.value = status.required_upgrade_codes ?? []
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to check upgrades'
  } finally {
    loading.value = false
  }
}

async function confirmUpgrade() {
  running.value = true
  error.value = null
  try {
    const result = await runUpgrade(UPGRADE_SQLITE_TO_POSTGRES)
    successMessage.value = result.message
    restartRequired.value = result.restart_required
    requiredCodes.value = requiredCodes.value.filter((code) => code !== UPGRADE_SQLITE_TO_POSTGRES)
    showConfirm.value = false
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Upgrade failed'
  } finally {
    running.value = false
  }
}

onMounted(() => {
  void loadUpgrades()
})
</script>

<template>
  <div v-if="showBanner" class="border-b border-amber-200 bg-amber-50 px-6 py-4">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p class="font-semibold text-amber-900">
          <template v-if="restartRequired">Database upgrade completed</template>
          <template v-else>Database upgrade required</template>
        </p>
        <p class="mt-1 text-sm text-amber-800">
          <template v-if="restartRequired">
            {{ successMessage ?? 'Restart the API and worker containers to finish switching to PostgreSQL.' }}
          </template>
          <template v-else>
            Migrate from legacy SQLite to PostgreSQL. The app will need a restart after the upgrade.
          </template>
        </p>
        <p v-if="error" class="mt-2 text-sm text-red-700">{{ error }}</p>
      </div>

      <div v-if="!restartRequired && needsSQLiteUpgrade" class="flex shrink-0 gap-2">
        <AppButton variant="secondary" :loading="loading" @click="loadUpgrades">Refresh</AppButton>
        <AppButton variant="primary" :loading="running" @click="showConfirm = true">
          Upgrade to PostgreSQL
        </AppButton>
      </div>
    </div>

    <div
      v-if="showConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      @click.self="showConfirm = false"
    >
      <div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <h3 class="text-lg font-semibold text-gray-900">Confirm database upgrade</h3>
        <p class="mt-3 text-sm text-gray-600">
          This will copy LLMs, schedules, users, and sessions from SQLite into PostgreSQL, update the
          configuration, and back up the old SQLite file. Expect brief downtime and restart the API and
          worker afterward.
        </p>
        <p v-if="error" class="mt-3 text-sm text-red-600">{{ error }}</p>
        <div class="mt-6 flex justify-end gap-2">
          <AppButton variant="secondary" :disabled="running" @click="showConfirm = false">Cancel</AppButton>
          <AppButton variant="primary" :loading="running" @click="confirmUpgrade">Run upgrade</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
