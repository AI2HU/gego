<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import SearchMatchCard from '@/components/search/SearchMatchCard.vue'
import BrandMapPopover from '@/components/search/BrandMapPopover.vue'
import TopBrandsQuickSearch from '@/components/search/TopBrandsQuickSearch.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import StatCard from '@/components/ui/StatCard.vue'
import TagFilter from '@/components/ui/TagFilter.vue'
import { searchQueryFromRoute } from '@/lib/search-navigation'
import {
  filterResponsesByTags,
  resolveSearchMatches,
  uniqueSearchTerms,
} from '@/lib/search-matches'
import { usePromptsQuery } from '@/queries/prompts'
import { useSearchMutation } from '@/queries/search'
import { useAuthStore } from '@/stores/auth'
import type { SearchMatch, SearchResponseItem } from '@/types/search'

const route = useRoute()
const authStore = useAuthStore()

const keyword = ref('')
const caseSensitive = ref(false)
const limitInput = ref('50')
const selectedTags = ref<string[]>([])

const searchMutation = useSearchMutation()
const promptsQuery = usePromptsQuery()

const hasSearched = ref(false)
const lastKeyword = ref('')
const lastSearchTerms = ref<string[]>([])
const lastTags = ref<string[]>([])
const lastResponses = ref<SearchResponseItem[]>([])
const allMatches = ref<SearchMatch[]>([])
const totalResponses = ref(0)
const totalMentions = ref(0)

const allTags = computed(() => {
  const tags = new Set<string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    for (const tag of prompt.tags ?? []) {
      tags.add(tag)
    }
  }
  return Array.from(tags).sort((a, b) => a.localeCompare(b))
})

const promptNameMap = computed(() => {
  const map = new Map<string, string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    map.set(prompt.id, prompt.template)
  }
  return map
})

const promptTagsMap = computed(() => {
  const map = new Map<string, string[]>()
  for (const prompt of promptsQuery.data.value ?? []) {
    map.set(prompt.id, prompt.tags ?? [])
  }
  return map
})

const displayLimit = computed(() => {
  const parsed = Number.parseInt(limitInput.value, 10)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return 50
  }
  return Math.min(parsed, 1000)
})

const visibleMatches = computed(() => allMatches.value.slice(0, displayLimit.value))

const hiddenMatchCount = computed(() =>
  Math.max(0, allMatches.value.length - displayLimit.value),
)

const lastAliasTerms = computed(() =>
  lastSearchTerms.value.filter(
    (term) => term.toLowerCase() !== lastKeyword.value.toLowerCase(),
  ),
)

const statsHintSuffix = computed(() => {
  const aliasHint =
    lastAliasTerms.value.length > 0
      ? ` · including aliases: ${lastAliasTerms.value.join(', ')}`
      : ''
  if (lastTags.value.length === 0) {
    return aliasHint
  }
  return ` · tags: ${lastTags.value.join(', ')}${aliasHint}`
})

const errorMessage = computed(() => {
  const error = searchMutation.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Search failed'
})

const canMapBrands = computed(() => authStore.hasPermission('words'))

const canSearch = computed(() => keyword.value.trim().length >= 2)

function buildMatches(responses: SearchResponseItem[], trimmed: string) {
  const prompts = promptsQuery.data.value ?? []
  const scopedResponses = filterResponsesByTags(responses, selectedTags.value, prompts)

  return {
    scopedResponses,
    matches: resolveSearchMatches(scopedResponses, trimmed, {
      caseSensitive: caseSensitive.value,
      searchTerms: lastSearchTerms.value,
      promptNames: promptNameMap.value,
      promptTags: promptTagsMap.value,
    }),
  }
}

function toggleTag(tag: string) {
  const index = selectedTags.value.indexOf(tag)
  if (index === -1) {
    selectedTags.value = [...selectedTags.value, tag]
    return
  }
  selectedTags.value = selectedTags.value.filter((value) => value !== tag)
}

function clearTagFilters() {
  selectedTags.value = []
}

async function runSearch(searchKeyword?: string) {
  const trimmed = (typeof searchKeyword === 'string' ? searchKeyword : keyword.value).trim()
  if (trimmed.length < 2) {
    return
  }

  keyword.value = trimmed
  hasSearched.value = true
  lastKeyword.value = trimmed
  lastTags.value = [...selectedTags.value]

  const fetchLimit = Math.min(displayLimit.value * 10, 1000)
  const result = await searchMutation.mutateAsync({
    keyword: trimmed,
    tags: selectedTags.value.length > 0 ? selectedTags.value : undefined,
    limit: fetchLimit,
  })

  lastSearchTerms.value = uniqueSearchTerms(trimmed, result.search_terms)
  lastResponses.value = result.responses ?? []

  const { scopedResponses, matches } = buildMatches(lastResponses.value, trimmed)
  allMatches.value = matches

  totalResponses.value = Math.max(result.total_responses ?? 0, scopedResponses.length)
  totalMentions.value = result.total_mentions ?? allMatches.value.length
}

watch(
  selectedTags,
  () => {
    if (!hasSearched.value || lastKeyword.value.trim().length < 2) {
      return
    }
    void runSearch(lastKeyword.value)
  },
  { deep: true },
)

watch(caseSensitive, () => {
  if (!hasSearched.value || lastKeyword.value.trim().length < 2) {
    return
  }
  const { matches } = buildMatches(lastResponses.value, lastKeyword.value)
  allMatches.value = matches
})

function applyRouteSearch() {
  const queryTerm = searchQueryFromRoute(route.query)
  if (!queryTerm) {
    return
  }
  if (queryTerm === lastKeyword.value && hasSearched.value) {
    return
  }
  void runSearch(queryTerm)
}

onMounted(applyRouteSearch)

watch(() => route.query.q, applyRouteSearch)
</script>

<template>
  <div>
    <div
      class="mb-6 space-y-4 rounded-xl border border-gray-200/60 bg-white/60 backdrop-blur-sm p-4"
    >
      <div class="flex flex-col gap-4 sm:flex-row sm:items-end">
        <div class="min-w-0 flex-1">
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-400">
            Keyword
          </label>
          <AppInput
            v-model="keyword"
            placeholder="Enter at least 2 characters..."
            :disabled="searchMutation.isPending.value"
            @enter="runSearch"
          />
        </div>

        <div class="w-full sm:w-32">
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-400">
            Limit
          </label>
          <AppInput
            v-model="limitInput"
            type="number"
            placeholder="50"
            :disabled="searchMutation.isPending.value"
          />
        </div>

        <div class="w-full shrink-0 sm:w-auto">
          <span
            class="mb-1.5 block text-xs font-medium uppercase tracking-wider opacity-0 select-none"
            aria-hidden="true"
          >
            Search
          </span>
          <AppButton
            icon="search"
            class="w-full"
            :disabled="!canSearch || searchMutation.isPending.value"
            @click="runSearch"
          >
            {{ searchMutation.isPending.value ? 'Searching...' : 'Search' }}
          </AppButton>
        </div>
      </div>

      <TopBrandsQuickSearch
        :disabled="searchMutation.isPending.value"
        :active-keyword="lastKeyword"
        :tags="selectedTags"
        @select="runSearch"
      />

      <label class="inline-flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
        <input
          v-model="caseSensitive"
          type="checkbox"
          class="rounded border-gray-300 text-slate-600 focus:ring-slate-500"
          :disabled="searchMutation.isPending.value"
        />
        Case-sensitive search
      </label>

      <TagFilter
        :tags="allTags"
        :selected-tags="selectedTags"
        :disabled="searchMutation.isPending.value"
        @toggle="toggleTag"
        @clear="clearTagFilters"
      />
    </div>

    <AppAlert
      v-if="errorMessage"
      title="Search failed"
      @retry="runSearch"
    >
      {{ errorMessage }}
    </AppAlert>

    <template v-if="hasSearched && !searchMutation.isPending.value && !errorMessage">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 mb-6 md:mb-8">
        <StatCard
          label="Responses"
          :value="totalResponses.toLocaleString()"
          :hint="'Containing \'' + lastKeyword + '\'' + statsHintSuffix"
          icon="file"
        />
        <StatCard
          label="Matches"
          :value="totalMentions.toLocaleString()"
          :hint="'For \'' + lastKeyword + '\'' + statsHintSuffix"
          icon="search"
        />
        <StatCard
          label="Showing"
          :value="visibleMatches.length.toLocaleString()"
          :hint="hiddenMatchCount > 0 ? `${hiddenMatchCount} more hidden` : 'All fetched matches visible'"
          icon="chart-bar"
        />
      </div>

      <EmptyState
        v-if="allMatches.length === 0"
        title="No matches found"
        :description="
          lastTags.length > 0
            ? 'No responses match your keyword and selected tags.'
            : 'No responses contain \'' + lastKeyword + '\'.'
        "
        icon="search"
      />

      <div v-else class="space-y-6">
        <div
          v-if="lastAliasTerms.length > 0"
          class="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-lg border border-gray-200/60 bg-white/60 px-4 py-3 text-xs text-gray-600"
        >
          <span class="font-medium text-gray-700">Highlights:</span>
          <span class="inline-flex items-center gap-1.5">
            <mark class="rounded bg-amber-100 px-1 font-medium text-amber-900">term</mark>
            search — {{ lastKeyword }}
          </span>
          <span class="inline-flex items-center gap-1.5">
            <mark class="rounded bg-sky-100 px-1 font-medium text-sky-900">alias</mark>
            brand {{ lastAliasTerms.length === 1 ? 'alias' : 'aliases' }} — {{ lastAliasTerms.join(', ') }}
          </span>
        </div>

        <p
          v-if="canMapBrands"
          class="text-sm text-gray-600"
        >
          Select text in a response to map it to a canonical brand name.
        </p>

        <SearchMatchCard
          v-for="(match, index) in visibleMatches"
          :key="`${match.responseId}-${match.keyword}-${index}`"
          :match="match"
          :highlight-terms="lastSearchTerms"
          :search-keyword="lastKeyword"
          :case-sensitive="caseSensitive"
        />

        <p
          v-if="hiddenMatchCount > 0"
          class="text-center text-sm text-gray-500"
        >
          ... and {{ hiddenMatchCount.toLocaleString() }} more matches (increase limit to see more)
        </p>
      </div>
    </template>

    <BrandMapPopover />
  </div>
</template>
