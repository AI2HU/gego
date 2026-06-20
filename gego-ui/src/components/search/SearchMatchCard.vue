<script setup lang="ts">
import ResponseMarkdown from '@/components/search/ResponseMarkdown.vue'
import { card } from '@/design/classes'
import type { SearchMatch } from '@/types/search'

defineProps<{
  match: SearchMatch
  caseSensitive?: boolean
}>()

function formatDate(value: string): string {
  return new Date(value).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function formatTemperature(value: number): string {
  return value === 0 ? 'N/A (legacy)' : value.toFixed(1)
}
</script>

<template>
  <article :class="[card.base, 'overflow-hidden']">
    <div :class="[card.body, 'space-y-4']">
      <dl class="grid gap-3 sm:grid-cols-2 text-sm">
        <div class="sm:col-span-2">
          <dt class="text-xs font-medium uppercase tracking-wider text-gray-400">Prompt</dt>
          <dd class="mt-1 text-gray-900">{{ match.promptName }}</dd>
          <dd v-if="match.promptTags.length > 0" class="mt-2 flex flex-wrap gap-1">
            <span
              v-for="tag in match.promptTags"
              :key="tag"
              class="inline-flex rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700"
            >
              {{ tag }}
            </span>
          </dd>
        </div>
        <div>
          <dt class="text-xs font-medium uppercase tracking-wider text-gray-400">LLM</dt>
          <dd class="mt-1 text-gray-900">
            {{ match.llmName }}
            <span class="text-gray-500">({{ match.llmProvider }})</span>
          </dd>
        </div>
        <div>
          <dt class="text-xs font-medium uppercase tracking-wider text-gray-400">Temperature</dt>
          <dd class="mt-1 text-gray-900">{{ formatTemperature(match.temperature) }}</dd>
        </div>
        <div>
          <dt class="text-xs font-medium uppercase tracking-wider text-gray-400">Date</dt>
          <dd class="mt-1 text-gray-900">{{ formatDate(match.createdAt) }}</dd>
        </div>
      </dl>

      <div>
        <p class="text-xs font-semibold uppercase tracking-wider text-emerald-700 mb-2">Response</p>
        <div :class="card.inset">
          <ResponseMarkdown
            :text="match.responseText"
            :keyword="match.keyword"
            :case-sensitive="caseSensitive"
          />
        </div>
      </div>
    </div>
  </article>
</template>
