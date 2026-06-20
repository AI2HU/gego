<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import ErrorLogsTable from '@/components/logs/ErrorLogsTable.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import { useErrorLogsQuery } from '@/queries/logs'
import { useModelsQuery } from '@/queries/models'
import type { ErrorLogsQueryParams } from '@/types/logs'

const page = ref(1)
const limit = 20
const selectedLlmId = ref('')

const queryParams = computed<ErrorLogsQueryParams>(() => ({
  page: page.value,
  limit,
  llm_id: selectedLlmId.value || undefined,
}))

const logsQuery = useErrorLogsQuery(queryParams)
const modelsQuery = useModelsQuery()

const logs = computed(() => logsQuery.data.value?.data ?? [])
const pagination = computed(() => logsQuery.data.value?.pagination)
const models = computed(() => modelsQuery.data.value ?? [])

const errorMessage = computed(() => {
  const error = logsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load error logs'
})

const isInitialLoading = computed(
  () => logsQuery.isPending.value && !logsQuery.data.value,
)

const hasPreviousPage = computed(() => (pagination.value?.page ?? 1) > 1)
const hasNextPage = computed(() => {
  if (!pagination.value) return false
  return pagination.value.page < pagination.value.total_pages
})

const isFetching = computed(() => logsQuery.isFetching.value)

watch(selectedLlmId, () => {
  page.value = 1
})

function goToPreviousPage() {
  if (hasPreviousPage.value) {
    page.value -= 1
  }
}

function goToNextPage() {
  if (hasNextPage.value) {
    page.value += 1
  }
}

function refresh() {
  void logsQuery.refetch()
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <label class="block sm:max-w-xs">
        <span class="mb-1.5 block text-xs font-medium uppercase tracking-wide text-gray-500">
          Filter by model
        </span>
        <select
          v-model="selectedLlmId"
          class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200"
        >
          <option value="">All models</option>
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.name }} ({{ model.provider }})
          </option>
        </select>
      </label>

      <AppButton variant="secondary" class="shrink-0 self-start" @click="refresh">
        <span class="inline-flex items-center gap-2">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12a9 9 0 1 1-2.64-6.36" />
            <path d="M21 3v6h-6" />
          </svg>
          Refresh
        </span>
      </AppButton>
    </div>

    <LoadingState v-if="isInitialLoading" message="Loading error logs…" />

    <AppAlert v-else-if="errorMessage" title="Failed to load logs">
      {{ errorMessage }}
    </AppAlert>

    <template v-else>
      <EmptyState
        v-if="logs.length === 0"
        title="No LLM errors recorded"
        description="Failed LLM calls from scheduled runs appear here when the scheduler persists an error response."
        icon="check"
      />

      <template v-else>
        <ErrorLogsTable :logs="logs" />

        <AppCard v-if="pagination && pagination.total_pages > 1" class="px-4 py-3">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-sm text-gray-500">
              Page {{ pagination.page }} of {{ pagination.total_pages }}
              · {{ pagination.total }} error{{ pagination.total === 1 ? '' : 's' }}
            </p>

            <div class="flex items-center gap-2">
              <AppButton
                variant="secondary"
                size="sm"
                :disabled="!hasPreviousPage || isFetching"
                @click="goToPreviousPage"
              >
                Previous
              </AppButton>
              <AppButton
                variant="secondary"
                size="sm"
                :disabled="!hasNextPage || isFetching"
                @click="goToNextPage"
              >
                Next
              </AppButton>
            </div>
          </div>
        </AppCard>
      </template>
    </template>
  </div>
</template>
