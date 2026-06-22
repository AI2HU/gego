<script setup lang="ts">
import { computed, ref } from 'vue'

import AddScheduleWizard from '@/components/scheduler/AddScheduleWizard.vue'
import SchedulesTable from '@/components/scheduler/SchedulesTable.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { getCronHint, getCronLabel } from '@/types/schedule'
import {
  useDeleteScheduleMutation,
  useReloadSchedulerMutation,
  useRunScheduleMutation,
  useSchedulerStatusQuery,
  useSchedulesQuery,
  useStartSchedulerMutation,
  useStopSchedulerMutation,
  useUpdateScheduleMutation,
} from '@/queries/schedules'

const showWizard = ref(false)
const deletingId = ref<string | null>(null)
const togglingId = ref<string | null>(null)
const runningId = ref<string | null>(null)
const actionError = ref<string | null>(null)

const schedulesQuery = useSchedulesQuery()
const statusQuery = useSchedulerStatusQuery()
const startMutation = useStartSchedulerMutation()
const stopMutation = useStopSchedulerMutation()
const reloadMutation = useReloadSchedulerMutation()
const deleteMutation = useDeleteScheduleMutation()
const updateMutation = useUpdateScheduleMutation()
const runMutation = useRunScheduleMutation()

const schedules = computed(() => schedulesQuery.data.value ?? [])
const enabledSchedules = computed(() => schedules.value.filter((schedule) => schedule.enabled))

const schedulerRunning = computed(() => statusQuery.data.value?.running ?? false)
const enabledScheduleCount = computed(() => enabledSchedules.value.length)

const errorMessage = computed(() => {
  const error = schedulesQuery.error.value ?? statusQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load scheduler data'
})

const isInitialLoading = computed(
  () =>
    (schedulesQuery.isPending.value && !schedulesQuery.data.value) ||
    (statusQuery.isPending.value && !statusQuery.data.value),
)

const isSchedulerBusy = computed(
  () =>
    startMutation.isPending.value ||
    stopMutation.isPending.value ||
    reloadMutation.isPending.value,
)

function openWizard() {
  showWizard.value = true
}

function closeWizard() {
  showWizard.value = false
}

function onScheduleAdded() {
  void schedulesQuery.refetch()
  void statusQuery.refetch()
}

async function handleStart() {
  actionError.value = null
  try {
    await startMutation.mutateAsync()
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to start scheduler'
  }
}

async function handleStop() {
  actionError.value = null
  try {
    await stopMutation.mutateAsync()
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to stop scheduler'
  }
}

async function handleReload() {
  actionError.value = null
  try {
    await reloadMutation.mutateAsync()
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to reload scheduler'
  }
}

async function handleDelete(id: string) {
  deletingId.value = id
  actionError.value = null
  try {
    await deleteMutation.mutateAsync(id)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to delete schedule'
  } finally {
    deletingId.value = null
  }
}

async function handleToggleEnabled(id: string, enabled: boolean) {
  togglingId.value = id
  actionError.value = null
  try {
    await updateMutation.mutateAsync({ id, payload: { enabled } })
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to update schedule'
  } finally {
    togglingId.value = null
  }
}

async function handleRun(id: string) {
  runningId.value = id
  actionError.value = null
  try {
    await runMutation.mutateAsync(id)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to run schedule'
  } finally {
    runningId.value = null
  }
}
</script>

<template>
  <div>
    <div class="flex justify-end mb-6">
      <AppButton class="shrink-0 self-start sm:self-auto" @click="openWizard">
        <span class="inline-flex items-center gap-2">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Add schedule
        </span>
      </AppButton>
    </div>

    <AppAlert
      v-if="errorMessage"
      title="Unable to load scheduler"
      @retry="() => { schedulesQuery.refetch(); statusQuery.refetch() }"
    >
      {{ errorMessage }}
    </AppAlert>

    <AppAlert v-else-if="actionError" title="Action failed">
      {{ actionError }}
    </AppAlert>

    <LoadingState
      v-if="isInitialLoading"
      title="Loading scheduler"
      description="Fetching schedules and scheduler status..."
    />

    <template v-else>
      <AppCard class="mb-6">
        <template #header>
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div>
              <h2 class="text-base font-semibold text-gray-900">Scheduler control</h2>
              <p class="text-sm text-gray-500 mt-1">
                Start or stop the cron scheduler. Reload after changing schedules.
              </p>
            </div>
            <StatusBadge
              :connected="schedulerRunning"
              :label="schedulerRunning ? 'Running' : 'Stopped'"
            />
          </div>
        </template>

        <div class="space-y-4">
          <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600">
            <span>
              <span class="font-semibold text-gray-900">{{ enabledScheduleCount }}</span>
              enabled schedule{{ enabledScheduleCount === 1 ? '' : 's' }}
            </span>
            <span class="hidden sm:inline text-gray-300">|</span>
            <span>
              <span class="font-semibold text-gray-900">{{ schedules.length }}</span>
              total schedule{{ schedules.length === 1 ? '' : 's' }}
            </span>
          </div>

          <div v-if="enabledScheduleCount === 0" class="rounded-lg border border-amber-200/80 bg-amber-50/80 px-4 py-3 text-sm text-amber-900">
            No enabled schedules found. Add a schedule or enable an existing one before starting the scheduler.
          </div>

          <div v-else-if="schedulerRunning && enabledSchedules.length > 0" class="rounded-lg border border-gray-200/80 bg-slate-50/80 px-4 py-3">
            <p class="text-sm font-medium text-gray-900 mb-2">Active schedules</p>
            <ul class="space-y-1.5 text-sm text-gray-700">
              <li v-for="schedule in enabledSchedules" :key="schedule.id">
                {{ schedule.name }}
                <span
                  class="text-gray-400 text-xs ml-1"
                  :title="getCronHint(schedule.cron_expr)"
                >
                  {{ getCronLabel(schedule.cron_expr) }}
                </span>
              </li>
            </ul>
          </div>

          <div class="flex flex-wrap gap-2">
            <AppButton
              :disabled="schedulerRunning || enabledScheduleCount === 0"
              :loading="startMutation.isPending.value"
              @click="handleStart"
            >
              Start scheduler
            </AppButton>
            <AppButton
              variant="secondary"
              :disabled="!schedulerRunning"
              :loading="stopMutation.isPending.value"
              @click="handleStop"
            >
              Stop scheduler
            </AppButton>
            <AppButton
              variant="ghost"
              :disabled="!schedulerRunning || isSchedulerBusy"
              :loading="reloadMutation.isPending.value"
              @click="handleReload"
            >
              Reload schedules
            </AppButton>
          </div>
        </div>
      </AppCard>

      <div v-if="schedules.length === 0" class="text-center">
        <EmptyState
          title="No schedules configured"
          description="Create a schedule to run prompts and models automatically on a cron expression."
          icon="clock"
        />
        <AppButton class="mt-4" @click="openWizard">Add your first schedule</AppButton>
      </div>

      <SchedulesTable
        v-else
        :schedules="schedules"
        :deleting-id="deletingId"
        :toggling-id="togglingId"
        :running-id="runningId"
        @delete="handleDelete"
        @toggle-enabled="handleToggleEnabled"
        @run="handleRun"
      />
    </template>

    <AddScheduleWizard
      v-if="showWizard"
      @close="closeWizard"
      @added="onScheduleAdded"
    />
  </div>
</template>
