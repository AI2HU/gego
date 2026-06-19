<script setup lang="ts">
import { computed, ref } from 'vue'

import AddPromptWizard from '@/components/prompts/AddPromptWizard.vue'
import PromptsTable from '@/components/prompts/PromptsTable.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import {
  useDeletePromptMutation,
  usePromptsQuery,
  useUpdatePromptMutation,
} from '@/queries/prompts'

const showWizard = ref(false)
const searchQuery = ref('')
const deletingId = ref<string | null>(null)
const savingTagsId = ref<string | null>(null)

const promptsQuery = usePromptsQuery()
const deleteMutation = useDeletePromptMutation()
const updateMutation = useUpdatePromptMutation()

const prompts = computed(() => promptsQuery.data.value ?? [])

const allTags = computed(() => {
  const tags = new Set<string>()
  for (const prompt of prompts.value) {
    for (const tag of prompt.tags ?? []) {
      tags.add(tag)
    }
  }
  return Array.from(tags).sort((a, b) => a.localeCompare(b))
})

const filteredPrompts = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) {
    return prompts.value
  }

  return prompts.value.filter((prompt) => {
    if (prompt.template.toLowerCase().includes(query)) {
      return true
    }

    return (prompt.tags ?? []).some((tag) => tag.toLowerCase().includes(query))
  })
})

const hasActiveFilters = computed(() => searchQuery.value.trim().length > 0)

const errorMessage = computed(() => {
  const error = promptsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load prompts'
})

const isInitialLoading = computed(
  () => promptsQuery.isPending.value && !promptsQuery.data.value,
)

function applyTagFilter(tag: string) {
  searchQuery.value = tag
}

function clearFilters() {
  searchQuery.value = ''
}

async function handleDelete(id: string) {
  deletingId.value = id
  try {
    await deleteMutation.mutateAsync(id)
  } finally {
    deletingId.value = null
  }
}

async function handleUpdateTags(id: string, tags: string[]) {
  savingTagsId.value = id
  try {
    await updateMutation.mutateAsync({ id, payload: { tags } })
  } finally {
    savingTagsId.value = null
  }
}

function openWizard() {
  showWizard.value = true
}

function closeWizard() {
  showWizard.value = false
}

function onPromptsAdded() {
  void promptsQuery.refetch()
}
</script>

<template>
  <div>
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6 md:mb-8">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Administration</p>
        <h1 class="text-2xl md:text-3xl font-semibold text-gray-900 mt-1">Prompts</h1>
        <p class="text-sm text-gray-500 mt-1 max-w-xl">
          Scan, filter, and organize prompt templates. Search by content or tag and edit tags inline.
        </p>
      </div>

      <AppButton class="shrink-0 self-start sm:self-auto" @click="openWizard">
        <span class="inline-flex items-center gap-2">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Add prompts
        </span>
      </AppButton>
    </div>

    <AppAlert
      v-if="errorMessage"
      title="Unable to load prompts"
      @retry="promptsQuery.refetch()"
    >
      {{ errorMessage }}
    </AppAlert>

    <LoadingState
      v-if="isInitialLoading"
      title="Loading prompts"
      description="Fetching configured prompt templates..."
    />

    <template v-else>
      <div v-if="prompts.length === 0" class="text-center">
        <EmptyState
          title="No prompts configured"
          description="Add your first prompt template to start running LLM queries and collecting keyword analytics."
          icon="comment"
        />
        <AppButton class="mt-4" @click="openWizard">Add your first prompt</AppButton>
      </div>

      <template v-else>
        <div
          class="mb-4 space-y-4 rounded-xl border border-gray-200/60 bg-white/60 backdrop-blur-sm p-4"
        >
          <AppInput
            v-model="searchQuery"
            placeholder="Search prompt text or tags..."
          />

          <div v-if="allTags.length > 0" class="flex flex-wrap items-center gap-2">
            <span class="text-xs font-medium uppercase tracking-wider text-gray-400">
              Tags
            </span>
            <button
              v-for="tag in allTags"
              :key="tag"
              type="button"
              class="inline-flex items-center rounded-full border border-slate-200 bg-white px-2.5 py-1 text-xs font-medium text-slate-700 transition-colors hover:border-slate-300 hover:bg-slate-50"
              @click="applyTagFilter(tag)"
            >
              {{ tag }}
            </button>
          </div>

          <div class="flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600">
            <p>
              Showing
              <span class="font-semibold text-gray-900">{{ filteredPrompts.length }}</span>
              of
              <span class="font-semibold text-gray-900">{{ prompts.length }}</span>
              prompts
            </p>

            <AppButton
              v-if="hasActiveFilters"
              variant="ghost"
              size="sm"
              @click="clearFilters"
            >
              Clear filters
            </AppButton>
          </div>
        </div>

        <EmptyState
          v-if="filteredPrompts.length === 0"
          title="No prompts match your filters"
          description="Try a different search term."
          icon="search"
        />

        <PromptsTable
          v-else
          :prompts="filteredPrompts"
          :deleting-id="deletingId"
          :saving-tags-id="savingTagsId"
          @delete="handleDelete"
          @update-tags="handleUpdateTags"
        />
      </template>
    </template>

    <AddPromptWizard
      v-if="showWizard"
      @close="closeWizard"
      @added="onPromptsAdded"
    />
  </div>
</template>
