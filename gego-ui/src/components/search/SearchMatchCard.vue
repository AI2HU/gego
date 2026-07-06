<script setup lang="ts">
import { computed } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import BrandMapSelection from '@/components/search/BrandMapSelection.vue'
import ResponseMarkdown from '@/components/search/ResponseMarkdown.vue'
import { card } from '@/design/classes'
import { linkCitationMarkers } from '@/lib/search-matches'
import type { SearchMatch } from '@/types/search'

const props = defineProps<{
  match: SearchMatch
  highlightTerms: string[]
  searchKeyword: string
  caseSensitive?: boolean
}>()

const responseText = computed(() =>
  linkCitationMarkers(props.match.responseText, props.match.searchUrls),
)

function sourceLabel(index: number, url: string, title?: string): string {
  const trimmedTitle = title?.trim()
  if (trimmedTitle) {
    return trimmedTitle
  }

  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

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
          <BrandMapSelection>
            <ResponseMarkdown
              :text="responseText"
              :highlight-terms="highlightTerms"
              :search-keyword="searchKeyword"
              :case-sensitive="caseSensitive"
            />
          </BrandMapSelection>
        </div>
      </div>

      <div v-if="match.searchUrls.length > 0">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-600 mb-2">Sources</p>
        <ul :class="[card.inset, 'space-y-2']">
          <li
            v-for="(source, index) in match.searchUrls"
            :key="`${source.url}-${source.citation_index}`"
            class="flex items-start gap-2 text-sm"
          >
            <span
              class="mt-0.5 inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded bg-slate-200 px-1 text-[10px] font-semibold text-slate-700"
            >
              {{ index + 1 }}
            </span>
            <a
              :href="source.url"
              target="_blank"
              rel="noopener noreferrer"
              class="min-w-0 text-slate-700 underline decoration-slate-300 underline-offset-2 hover:text-slate-900 hover:decoration-slate-500"
              :title="source.url"
            >
              {{ sourceLabel(index, source.url, source.title) }}
            </a>
            <AppIcon
              name="external-link"
              class="mt-0.5 h-3.5 w-3.5 shrink-0 text-slate-400"
              aria-hidden="true"
            />
          </li>
        </ul>
      </div>
    </div>
  </article>
</template>
