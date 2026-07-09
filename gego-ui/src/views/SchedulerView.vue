<script setup lang="ts">
import { computed, ref } from 'vue'

import RunDetailDrawer from '@/components/scheduler/RunDetailDrawer.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import AddScheduleWizard from '@/components/scheduler/AddScheduleWizard.vue'
import SchedulesTable from '@/components/scheduler/SchedulesTable.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { getCronHint, getCronLabel } from '@/types/schedule'
import type { ScheduleRunResponse, ScheduleResponse } from '@/types/schedule'
import {
  useDeleteScheduleMutation,
  useReloadSchedulerMutation,
  useRunScheduleMutation,
  useScheduleRunQuery,
  useScheduleRunsListQuery,
  useSchedulerStatusQuery,
  useSchedulesQuery,
  useStartSchedulerMutation,
  useStopSchedulerMutation,
  useUpdateScheduleMutation,
} from '@/queries/schedules'

const showWizard = ref(false)
const editingSchedule = ref<ScheduleResponse | null>(null)
const deletingId = ref<string | null>(null)
const togglingId = ref<string | null>(null)
const runningId = ref<string | null>(null)
const activeRunId = ref<string | null>(null)
const actionError = ref<string | null>(null)

const schedulesQuery = useSchedulesQuery()
const statusQuery = useSchedulerStatusQuery()
const runsListQuery = useScheduleRunsListQuery()
const startMutation = useStartSchedulerMutation()
const stopMutation = useStopSchedulerMutation()
const reloadMutation = useReloadSchedulerMutation()
const deleteMutation = useDeleteScheduleMutation()
const updateMutation = useUpdateScheduleMutation()
const runMutation = useRunScheduleMutation()
const activeRunQuery = useScheduleRunQuery(activeRunId)

const schedules = computed(() => schedulesQuery.data.value ?? [])
const enabledSchedules = computed(() => schedules.value.filter((schedule) => schedule.enabled))

const latestRunByScheduleId = computed(() => {
  const map: Record<string, ScheduleRunResponse> = {}
  for (const run of runsListQuery.data.value?.data ?? []) {
    const existing = map[run.schedule_id]
    if (!existing || new Date(run.created_at) > new Date(existing.created_at)) {
      map[run.schedule_id] = run
    }
  }
  return map
})

const schedulerRunning = computed(() => statusQuery.data.value?.running ?? false)
const enabledScheduleCount = computed(() => enabledSchedules.value.length)
const pendingJobs = computed(() => statusQuery.data.value?.pending_jobs ?? 0)
const activeWorkers = computed(() => statusQuery.data.value?.active_workers ?? 0)
const isLeader = computed(() => statusQuery.data.value?.is_leader ?? false)

const errorMessage = computed(() => {
  const error = schedulesQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load scheduler data'
})

const statusError = computed(() => {
  const error = statusQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load scheduler status'
})

const isInitialLoading = computed(
  () => schedulesQuery.isPending.value && schedulesQuery.data.value === undefined,
)

const isSchedulerBusy = computed(
  () =>
    startMutation.isPending.value ||
    stopMutation.isPending.value ||
    reloadMutation.isPending.value,
)

const canRunSchedules = computed(
  () => !statusQuery.isError.value && !statusQuery.isPending.value && activeWorkers.value > 0,
)

function openWizard() {
  editingSchedule.value = null
  showWizard.value = true
}

function closeWizard() {
  showWizard.value = false
  editingSchedule.value = null
}

function openEditWizard(schedule: ScheduleResponse) {
  editingSchedule.value = schedule
  showWizard.value = true
}

function onScheduleAdded() {
  void schedulesQuery.refetch()
  void statusQuery.refetch()
}

function onScheduleUpdated() {
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
    const result = await runMutation.mutateAsync(id)
    activeRunId.value = result.run_id
    void statusQuery.refetch()
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to run schedule'
  } finally {
    runningId.value = null
  }
}

function closeRunDetail() {
  activeRunId.value = null
}

function handleViewRun(runId: string) {
  activeRunId.value = runId
}
</script>

<template>
  <div>
    <div class="flex justify-end mb-6">
      <AppButton class="shrink-0 self-start sm:self-auto" @click="openWizard">
        <span class="inline-flex items-center gap-2">
          <AppIcon name="plus" size="sm" />
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
                Start or stop the cron scheduler. Requires etcd and at least one worker process.
              </p>
            </div>
            <StatusBadge
              :connected="schedulerRunning"
              :label="statusQuery.isPending.value ? 'Loading…' : (schedulerRunning ? 'Running' : 'Stopped')"
            />
          </div>
        </template>

        <div class="space-y-4">
          <div
            v-if="statusError"
            class="rounded-lg border border-amber-200/80 bg-amber-50/80 px-4 py-3 text-sm text-amber-900"
          >
            Scheduler status unavailable: {{ statusError }}.
            Start etcd with <code class="font-mono text-xs">make etcd-dev</code>.
          </div>

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
            <span class="hidden sm:inline text-gray-300">|</span>
            <span>
              <span class="font-semibold text-gray-900">{{ pendingJobs }}</span>
              pending job{{ pendingJobs === 1 ? '' : 's' }}
            </span>
            <span class="hidden sm:inline text-gray-300">|</span>
            <span>
              <span class="font-semibold text-gray-900">{{ activeWorkers }}</span>
              active worker{{ activeWorkers === 1 ? '' : 's' }}
            </span>
            <template v-if="schedulerRunning">
              <span class="hidden sm:inline text-gray-300">|</span>
              <span>
                leader: <span class="font-semibold text-gray-900">{{ isLeader ? 'this node' : 'other' }}</span>
              </span>
            </template>
          </div>

          <div v-if="!statusQuery.isPending.value && activeWorkers === 0" class="rounded-lg border border-amber-200/80 bg-amber-50/80 px-4 py-3 text-sm text-amber-900">
            <template v-if="pendingJobs > 0">
              Jobs are queued but no workers are connected.
            </template>
            <template v-else>
              No workers are connected.
            </template>
            Start a worker with <code class="font-mono text-xs">gego worker start</code> before running schedules.
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
        :latest-runs="latestRunByScheduleId"
        :can-run="canRunSchedules"
        :deleting-id="deletingId"
        :toggling-id="togglingId"
        :running-id="runningId"
        @delete="handleDelete"
        @edit="openEditWizard"
        @toggle-enabled="handleToggleEnabled"
        @run="handleRun"
        @view-run="handleViewRun"
      />
    </template>

    <AddScheduleWizard
      v-if="showWizard"
      :schedule="editingSchedule ?? undefined"
      @close="closeWizard"
      @added="onScheduleAdded"
      @updated="onScheduleUpdated"
    />
    <RunDetailDrawer
      v-if="activeRunId && activeRunQuery.data.value"
      :run="activeRunQuery.data.value.run"
      :jobs="activeRunQuery.data.value.jobs"
      @close="closeRunDetail"
      @retried="() => { activeRunQuery.refetch(); runsListQuery.refetch() }"
    />
  </div>
</template>
