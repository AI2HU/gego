<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import SearchMatchCard from '@/components/search/SearchMatchCard.vue'
import TopBrandsQuickSearch from '@/components/search/TopBrandsQuickSearch.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import StatCard from '@/components/ui/StatCard.vue'
import TagFilter from '@/components/ui/TagFilter.vue'
import { searchQueryFromRoute } from '@/lib/search-navigation'
import { filterResponsesByTags, findSearchMatches } from '@/lib/search-matches'
import { usePromptsQuery } from '@/queries/prompts'
import { useSearchMutation } from '@/queries/search'
import type { SearchMatch } from '@/types/search'

const route = useRoute()

const keyword = ref('')
const caseSensitive = ref(false)
const limitInput = ref('50')
const selectedTags = ref<string[]>([])

const searchMutation = useSearchMutation()
const promptsQuery = usePromptsQuery()

const hasSearched = ref(false)
const lastKeyword = ref('')
const lastTags = ref<string[]>([])
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

const statsHintSuffix = computed(() => {
  if (lastTags.value.length === 0) {
    return ''
  }
  return ` · tags: ${lastTags.value.join(', ')}`
})

const errorMessage = computed(() => {
  const error = searchMutation.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Search failed'
})

const canSearch = computed(() => keyword.value.trim().length >= 2)

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

  const prompts = promptsQuery.data.value ?? []
  const scopedResponses = filterResponsesByTags(result.responses ?? [], selectedTags.value, prompts)

  allMatches.value = findSearchMatches(scopedResponses, trimmed, {
    caseSensitive: caseSensitive.value,
    promptNames: promptNameMap.value,
    promptTags: promptTagsMap.value,
  })

  totalResponses.value = result.total_responses ?? scopedResponses.length
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
  void runSearch(lastKeyword.value)
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
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end">
        <div class="flex-1">
          <label class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-400">
            Keyword
          </label>
          <AppInput
            v-model="keyword"
            placeholder="Enter at least 2 characters..."
            :disabled="searchMutation.isPending.value"
            @enter="runSearch"
          />
          <TopBrandsQuickSearch
            :disabled="searchMutation.isPending.value"
            :active-keyword="lastKeyword"
            :tags="selectedTags"
            @select="runSearch"
          />
        </div>

        <div class="w-full lg:w-32">
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

        <AppButton
          icon="search"
          class="shrink-0"
          :disabled="!canSearch || searchMutation.isPending.value"
          @click="runSearch"
        >
          {{ searchMutation.isPending.value ? 'Searching...' : 'Search' }}
        </AppButton>
      </div>

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
        <SearchMatchCard
          v-for="(match, index) in visibleMatches"
          :key="`${match.responseId}-${index}`"
          :match="match"
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
  </div>
</template>
