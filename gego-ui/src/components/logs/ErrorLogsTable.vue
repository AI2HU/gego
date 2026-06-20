<script setup lang="ts">
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import { formatProviderName } from '@/lib/providers'
import type { ErrorLogEntry } from '@/types/logs'

defineProps<{
  logs: ErrorLogEntry[]
}>()

function formatDate(value: string): string {
  return new Date(value).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return `${text.slice(0, maxLength)}…`
}
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-gray-200/60 bg-white/80 backdrop-blur-sm shadow-sm">
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="border-b border-gray-200/60 bg-slate-50/80">
          <tr class="text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
            <th class="px-4 py-3">Time</th>
            <th class="px-4 py-3">Model</th>
            <th class="px-4 py-3">Prompt</th>
            <th class="px-4 py-3">Error</th>
          </tr>
        </thead>

        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="log in logs"
            :key="log.id"
            class="group align-top transition-colors hover:bg-slate-50/70"
          >
            <td class="whitespace-nowrap px-4 py-3 text-xs text-gray-500">
              {{ formatDate(log.created_at) }}
            </td>

            <td class="px-4 py-3">
              <div class="flex items-center gap-2.5">
                <ProviderLogo :provider="log.llm_provider" size="sm" />
                <div class="min-w-0">
                  <p class="font-medium text-gray-900">{{ log.llm_name }}</p>
                  <p class="text-xs text-gray-500">
                    {{ formatProviderName(log.llm_provider) }} · {{ log.llm_model }}
                  </p>
                </div>
              </div>
            </td>

            <td class="max-w-xs px-4 py-3">
              <p class="text-gray-800" :title="log.prompt_text">
                {{ truncate(log.prompt_text, 120) }}
              </p>
              <p v-if="log.schedule_id" class="mt-1 text-xs text-gray-400 font-mono">
                schedule: {{ log.schedule_id.slice(0, 8) }}…
              </p>
            </td>

            <td class="max-w-lg px-4 py-3">
              <p class="rounded-lg bg-red-50 px-3 py-2 font-mono text-xs leading-relaxed text-red-800 break-words">
                {{ log.error }}
              </p>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
