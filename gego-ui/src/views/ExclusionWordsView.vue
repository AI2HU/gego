<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import TagFilter from '@/components/ui/TagFilter.vue'
import {
  useCreateExclusionWordMutation,
  useDeleteExclusionWordMutation,
  useExclusionWordsQuery,
  useSuggestedBrandWordsQuery,
} from '@/queries/exclusion-words'
import { usePromptsQuery } from '@/queries/prompts'

const newWord = ref('')
const searchQuery = ref('')
const selectedTags = ref<string[]>([])
const deletingId = ref<string | null>(null)
const excludingWord = ref<string | null>(null)

const exclusionWordsQuery = useExclusionWordsQuery()
const promptsQuery = usePromptsQuery()
const suggestionsQuery = useSuggestedBrandWordsQuery(50, selectedTags)
const createMutation = useCreateExclusionWordMutation()
const deleteMutation = useDeleteExclusionWordMutation()

const exclusionWords = computed(() => exclusionWordsQuery.data.value ?? [])
const suggestions = computed(() => suggestionsQuery.data.value ?? [])

const excludedWordSet = computed(() => {
  const set = new Set<string>()
  for (const word of exclusionWords.value) {
    set.add(word.word.toLowerCase())
  }
  return set
})

const filteredWords = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) {
    return exclusionWords.value
  }
  return exclusionWords.value.filter((word) => word.word.toLowerCase().includes(query))
})

const allTags = computed(() => {
  const tags = new Set<string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    for (const tag of prompt.tags ?? []) {
      tags.add(tag)
    }
  }
  return Array.from(tags).sort((a, b) => a.localeCompare(b))
})

const hasActiveTagFilters = computed(() => selectedTags.value.length > 0)

const visibleSuggestions = computed(() =>
  suggestions.value.filter((item) => !excludedWordSet.value.has(item.word.toLowerCase())),
)

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

const errorMessage = computed(() => {
  const error = exclusionWordsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load exclusion words'
})

const isInitialLoading = computed(
  () => exclusionWordsQuery.isPending.value && !exclusionWordsQuery.data.value,
)

const excludedChipClass =
  'inline-flex items-center gap-1 rounded-full bg-slate-100 pl-3 pr-1.5 py-1.5 text-sm font-medium text-slate-700'

const suggestionChipClass =
  'inline-flex items-center gap-1.5 rounded-full border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition-colors hover:border-slate-300 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50'

async function handleAddWord() {
  const word = newWord.value.trim()
  if (!word) return

  await createMutation.mutateAsync({ word })
  newWord.value = ''
  void suggestionsQuery.refetch()
}

async function handleDelete(id: string) {
  deletingId.value = id
  try {
    await deleteMutation.mutateAsync(id)
    void suggestionsQuery.refetch()
  } finally {
    deletingId.value = null
  }
}

async function handleExcludeSuggestion(word: string) {
  excludingWord.value = word
  try {
    await createMutation.mutateAsync({ word })
    void suggestionsQuery.refetch()
  } finally {
    excludingWord.value = null
  }
}
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-xl border border-gray-200/60 bg-white/60 p-6 backdrop-blur-sm">
      <div class="mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Add exclusion word</h2>
        <p class="mt-1 text-sm text-gray-600">
          Excluded words are filtered out of brand and keyword tracking statistics.
        </p>
      </div>

      <form class="flex flex-col gap-3 sm:flex-row" @submit.prevent="handleAddWord">
        <AppInput
          v-model="newWord"
          class="flex-1"
          placeholder="e.g. However, Therefore, Additionally"
        />
        <AppButton
          type="submit"
          :disabled="!newWord.trim() || createMutation.isPending.value"
        >
          Add word
        </AppButton>
      </form>
    </section>

    <section>
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">Excluded words</h2>
          <p class="mt-1 text-sm text-gray-600">
            {{ exclusionWords.length }} word{{ exclusionWords.length === 1 ? '' : 's' }} excluded
          </p>
        </div>
        <AppInput
          v-model="searchQuery"
          class="w-full sm:max-w-xs"
          placeholder="Search excluded words..."
        />
      </div>

      <AppAlert
        v-if="errorMessage"
        title="Unable to load exclusion words"
        @retry="exclusionWordsQuery.refetch()"
      >
        {{ errorMessage }}
      </AppAlert>

      <LoadingState
        v-if="isInitialLoading"
        title="Loading exclusion words"
        description="Fetching words excluded from brand tracking..."
      />

      <template v-else>
        <EmptyState
          v-if="exclusionWords.length === 0"
          title="No exclusion words yet"
          description="Add common capitalized words that should not be tracked as brands."
          icon="tag"
        />

        <div
          v-else-if="filteredWords.length === 0"
          class="rounded-xl border border-dashed border-gray-200 px-6 py-10 text-center text-sm text-gray-500"
        >
          No excluded words match your search.
        </div>

        <div
          v-else
          class="rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm sm:p-5"
        >
          <div class="flex flex-wrap gap-2">
            <span
              v-for="word in filteredWords"
              :key="word.id"
              :class="excludedChipClass"
            >
              {{ word.word }}
              <button
                type="button"
                class="inline-flex rounded-full p-1 text-slate-400 transition-colors hover:bg-slate-200 hover:text-slate-700 disabled:opacity-50"
                :disabled="deletingId === word.id || deleteMutation.isPending.value"
                :title="`Remove ${word.word}`"
                @click="handleDelete(word.id)"
              >
                <AppIcon name="close" size="sm" />
              </button>
            </span>
          </div>
        </div>
      </template>
    </section>

    <section>
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">Detected brand words</h2>
          <p class="mt-1 text-sm text-gray-600">
            Click a word to exclude it from brand tracking.
            <span v-if="hasActiveTagFilters">Showing words from responses matching selected tags.</span>
          </p>
        </div>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="suggestionsQuery.isFetching.value"
          @click="suggestionsQuery.refetch()"
        >
          Refresh
        </AppButton>
      </div>

      <div
        v-if="allTags.length > 0"
        class="mb-4 rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm"
      >
        <TagFilter
          :tags="allTags"
          :selected-tags="selectedTags"
          :disabled="suggestionsQuery.isFetching.value"
          @toggle="toggleTag"
          @clear="clearTagFilters"
        />
      </div>

      <LoadingState
        v-if="suggestionsQuery.isPending.value && !suggestionsQuery.data.value"
        title="Loading suggestions"
        description="Scanning responses for detected brand words..."
      />

      <EmptyState
        v-else-if="visibleSuggestions.length === 0"
        title="No suggestions available"
        :description="
          hasActiveTagFilters
            ? 'No detected brand words found for the selected tags. Try different tags or clear the filter.'
            : 'Run prompts via the scheduler to collect responses, then review detected brand words here.'
        "
        icon="lightbulb"
      />

      <div
        v-else
        class="rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm sm:p-5"
      >
        <div class="flex flex-wrap gap-2">
          <button
            v-for="item in visibleSuggestions"
            :key="item.word"
            type="button"
            :class="suggestionChipClass"
            :disabled="excludingWord === item.word || createMutation.isPending.value"
            :title="`Exclude ${item.word} (${item.count} mention${item.count === 1 ? '' : 's'})`"
            @click="handleExcludeSuggestion(item.word)"
          >
            <span>{{ item.word }}</span>
            <span class="rounded-full bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-500">
              {{ item.count }}
            </span>
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
