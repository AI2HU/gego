<script setup lang="ts">
import { ref } from 'vue'

import AppButton from '@/components/ui/AppButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import type { ScheduleResponse, ScheduleRunResponse } from '@/types/schedule'
import { formatRunStatus, getCronHint, getCronLabel } from '@/types/schedule'

withDefaults(
  defineProps<{
    schedules: ScheduleResponse[]
    latestRuns?: Record<string, ScheduleRunResponse | undefined>
    canRun?: boolean
    deletingId?: string | null
    togglingId?: string | null
    runningId?: string | null
  }>(),
  { canRun: false },
)

const emit = defineEmits<{
  delete: [id: string]
  edit: [schedule: ScheduleResponse]
  toggleEnabled: [id: string, enabled: boolean]
  run: [id: string]
  viewRun: [runId: string]
}>()

const confirmingDeleteId = ref<string | null>(null)

function formatDate(value?: string): string {
  if (!value) return '—'
  return new Date(value).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatTemperature(value: number): string {
  if (value === -1) return 'Random'
  return value.toFixed(1)
}

function runTimestamp(run?: ScheduleRunResponse): string | undefined {
  if (!run) return undefined
  return run.finished_at ?? run.started_at ?? run.created_at
}

function isActiveRun(status?: string): boolean {
  return status === 'pending' || status === 'running'
}

function requestDelete(id: string) {
  confirmingDeleteId.value = id
}

function cancelDelete() {
  confirmingDeleteId.value = null
}

function confirmDelete(id: string) {
  emit('delete', id)
  confirmingDeleteId.value = null
}
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-gray-200/60 bg-white/80 backdrop-blur-sm shadow-sm">
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="border-b border-gray-200/60 bg-slate-50/80">
          <tr class="text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
            <th class="px-4 py-3">Schedule</th>
            <th class="px-4 py-3">Frequency</th>
            <th class="px-4 py-3">Prompts</th>
            <th class="px-4 py-3">Models</th>
            <th class="px-4 py-3">Temp</th>
            <th class="px-4 py-3">Last run</th>
            <th class="px-4 py-3">Run status</th>
            <th class="px-4 py-3">Status</th>
            <th class="w-64 px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>

        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="schedule in schedules"
            :key="schedule.id"
            class="group transition-colors hover:bg-slate-50/70"
          >
            <td class="px-4 py-3">
              <p class="font-medium text-gray-900">{{ schedule.name }}</p>
              <p class="text-xs text-gray-400 mt-0.5 font-mono">{{ schedule.id }}</p>
            </td>
            <td class="px-4 py-3 text-gray-700">
              <span
                class="text-sm"
                :title="getCronHint(schedule.cron_expr)"
              >
                {{ getCronLabel(schedule.cron_expr) }}
              </span>
            </td>
            <td class="px-4 py-3 text-gray-700">{{ schedule.prompt_ids.length }}</td>
            <td class="px-4 py-3 text-gray-700">{{ schedule.llm_ids.length }}</td>
            <td class="px-4 py-3 text-gray-700">{{ formatTemperature(schedule.temperature) }}</td>
            <td class="px-4 py-3 text-gray-600 whitespace-nowrap">
              {{ formatDate(runTimestamp(latestRuns?.[schedule.id]) ?? schedule.last_run) }}
            </td>
            <td class="px-4 py-3">
              <button
                v-if="latestRuns?.[schedule.id]"
                type="button"
                class="inline-flex items-center gap-1.5 text-left hover:opacity-80"
                @click="emit('viewRun', latestRuns[schedule.id]!.id)"
              >
                <StatusBadge
                  :connected="latestRuns[schedule.id]!.status === 'completed'"
                  :label="formatRunStatus(latestRuns[schedule.id]!.status)"
                  compact
                />
                <span
                  v-if="isActiveRun(latestRuns[schedule.id]!.status)"
                  class="text-xs text-gray-400"
                >
                  {{ latestRuns[schedule.id]!.completed_jobs }}/{{ latestRuns[schedule.id]!.total_jobs }}
                </span>
              </button>
              <span v-else class="text-gray-400">—</span>
            </td>
            <td class="px-4 py-3">
              <StatusBadge
                :connected="schedule.enabled"
                :label="schedule.enabled ? 'Enabled' : 'Disabled'"
                compact
              />
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center justify-end gap-1.5">
                <AppButton
                  variant="ghost"
                  size="sm"
                  @click="emit('edit', schedule)"
                >
                  Edit
                </AppButton>
                <span
                  class="inline-flex"
                  :title="canRun ? undefined : 'No worker is started. Run: gego worker start'"
                >
                  <AppButton
                    variant="ghost"
                    size="sm"
                    :disabled="!canRun"
                    :loading="runningId === schedule.id"
                    @click="emit('run', schedule.id)"
                  >
                    Run
                  </AppButton>
                </span>
                <AppButton
                  variant="ghost"
                  size="sm"
                  :loading="togglingId === schedule.id"
                  @click="emit('toggleEnabled', schedule.id, !schedule.enabled)"
                >
                  {{ schedule.enabled ? 'Disable' : 'Enable' }}
                </AppButton>

                <template v-if="confirmingDeleteId === schedule.id">
                  <AppButton variant="danger" size="sm" @click="confirmDelete(schedule.id)">
                    Confirm
                  </AppButton>
                  <AppButton variant="ghost" size="sm" @click="cancelDelete">Cancel</AppButton>
                </template>
                <AppButton
                  v-else
                  variant="ghost"
                  size="sm"
                  :loading="deletingId === schedule.id"
                  @click="requestDelete(schedule.id)"
                >
                  Delete
                </AppButton>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
