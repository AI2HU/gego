<script setup lang="ts">
import { computed, ref } from 'vue'

import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import AppButton from '@/components/ui/AppButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import {
  cancelScheduleRun,
  retryScheduleJob,
} from '@/api/scheduleRuns'
import { formatProviderName } from '@/lib/providers'
import { useModelsQuery } from '@/queries/models'
import { usePromptsQuery } from '@/queries/prompts'
import type { ScheduleJobResponse, ScheduleRunResponse } from '@/types/schedule'
import { formatRunStatus } from '@/types/schedule'

const props = defineProps<{
  run: ScheduleRunResponse
  jobs: ScheduleJobResponse[]
}>()

const emit = defineEmits<{
  close: []
  retried: []
}>()

const retryingId = ref<string | null>(null)
const cancelling = ref(false)

const promptsQuery = usePromptsQuery()
const modelsQuery = useModelsQuery()

const promptLabels = computed(() => {
  const map = new Map<string, string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    const label = prompt.template.trim()
    map.set(prompt.id, label.length > 80 ? `${label.slice(0, 80)}…` : label)
  }
  return map
})

const modelLabels = computed(() => {
  const map = new Map<string, string>()
  for (const model of modelsQuery.data.value ?? []) {
    map.set(model.id, model.name || model.model)
  }
  return map
})

const jobsByProvider = computed(() => {
  const groups = new Map<string, ScheduleJobResponse[]>()
  for (const job of props.jobs) {
    const provider = job.provider || 'unknown'
    const list = groups.get(provider) ?? []
    list.push(job)
    groups.set(provider, list)
  }
  return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b))
})

function formatDate(value?: string): string {
  if (!value) return '—'
  return new Date(value).toLocaleString()
}

function isRunning(status: string): boolean {
  return status === 'pending' || status === 'running'
}

function promptLabelsForJob(job: ScheduleJobResponse): string {
  const labels = job.prompt_ids.map((id) => promptLabels.value.get(id) ?? id)
  if (labels.length === 0) {
    return 'No prompts'
  }
  if (labels.length === 1) {
    return labels[0]!
  }
  return `${labels.length} prompts`
}

function promptDetailsForJob(job: ScheduleJobResponse): string[] {
  return job.prompt_ids.map((id) => promptLabels.value.get(id) ?? id)
}

function modelLabel(job: ScheduleJobResponse): string {
  return modelLabels.value.get(job.llm_id) ?? job.llm_id
}

async function handleCancel() {
  cancelling.value = true
  try {
    await cancelScheduleRun(props.run.id)
    emit('retried')
  } finally {
    cancelling.value = false
  }
}

async function handleRetry(jobId: string) {
  retryingId.value = jobId
  try {
    await retryScheduleJob(props.run.id, jobId)
    emit('retried')
  } finally {
    retryingId.value = null
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex justify-end">
    <div class="absolute inset-0 bg-black/30" @click="emit('close')" />
    <div class="relative w-full max-w-lg bg-white shadow-xl h-full overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-5 py-4 flex items-center justify-between">
        <div>
          <h3 class="text-base font-semibold text-gray-900">Schedule run</h3>
          <p class="text-xs text-gray-400 font-mono mt-0.5">{{ run.id }}</p>
        </div>
        <AppButton variant="ghost" size="sm" @click="emit('close')">Close</AppButton>
      </div>

      <div class="p-5 space-y-4">
        <div class="flex items-center gap-2 flex-wrap">
          <StatusBadge :connected="run.status === 'completed'" :label="formatRunStatus(run.status)" compact />
          <span class="text-sm text-gray-500">{{ run.completed_jobs }}/{{ run.total_jobs }} jobs</span>
          <span v-if="run.failed_jobs > 0" class="text-sm text-red-600">{{ run.failed_jobs }} failed</span>
          <AppButton
            v-if="isRunning(run.status)"
            variant="danger"
            size="sm"
            :loading="cancelling"
            @click="handleCancel"
          >
            Cancel run
          </AppButton>
        </div>

        <dl class="grid grid-cols-2 gap-3 text-sm">
          <div>
            <dt class="text-gray-500">Trigger</dt>
            <dd class="font-medium text-gray-900">{{ run.trigger }}</dd>
          </div>
          <div>
            <dt class="text-gray-500">Started</dt>
            <dd class="text-gray-900">{{ formatDate(run.started_at) }}</dd>
          </div>
          <div>
            <dt class="text-gray-500">Finished</dt>
            <dd class="text-gray-900">{{ formatDate(run.finished_at) }}</dd>
          </div>
        </dl>

        <div>
          <h4 class="text-sm font-semibold text-gray-900 mb-3">Jobs by provider</h4>
          <div class="space-y-4">
            <section
              v-for="[provider, providerJobs] in jobsByProvider"
              :key="provider"
              class="rounded-lg border border-gray-200 overflow-hidden"
            >
              <div class="flex items-center gap-2 px-3 py-2 bg-slate-50 border-b border-gray-200">
                <ProviderLogo :provider="provider" class="h-5 w-5" />
                <span class="text-sm font-medium text-gray-900">{{ formatProviderName(provider) }}</span>
                <span class="text-xs text-gray-500">{{ providerJobs.length }} model{{ providerJobs.length === 1 ? '' : 's' }}</span>
              </div>
              <ul class="divide-y divide-gray-100">
                <li
                  v-for="job in providerJobs"
                  :key="job.id"
                  class="px-3 py-2 text-sm"
                >
                  <div class="flex items-center justify-between gap-2">
                    <StatusBadge
                      :connected="job.status === 'completed'"
                      :label="formatRunStatus(job.status)"
                      compact
                    />
                    <AppButton
                      v-if="job.status === 'dead'"
                      variant="ghost"
                      size="sm"
                      :loading="retryingId === job.id"
                      @click="handleRetry(job.id)"
                    >
                      Retry
                    </AppButton>
                  </div>
                  <p class="text-sm text-gray-800 mt-1">{{ promptLabelsForJob(job) }}</p>
                  <ul
                    v-if="job.prompt_ids.length > 1"
                    class="mt-1 space-y-0.5 text-xs text-gray-600 list-disc list-inside"
                  >
                    <li v-for="(label, index) in promptDetailsForJob(job)" :key="`${job.id}-${index}`">
                      {{ label }}
                    </li>
                  </ul>
                  <p class="text-xs text-gray-500 mt-0.5">{{ modelLabel(job) }}</p>
                  <p v-if="job.error" class="text-xs text-red-600 mt-1">{{ job.error }}</p>
                </li>
              </ul>
            </section>
          </div>
        </div>

        <p v-if="isRunning(run.status)" class="text-xs text-gray-500">Auto-refreshing while run is active…</p>
      </div>
    </div>
  </div>
</template>
