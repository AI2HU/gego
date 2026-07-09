<script setup lang="ts">
import { computed, ref } from 'vue'

import AddModelWizard from '@/components/models/AddModelWizard.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import ModelCard from '@/components/models/ModelCard.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import { formatProviderName } from '@/lib/providers'
import { useDeleteModelMutation, useModelsQuery } from '@/queries/models'

const showWizard = ref(false)
const deletingId = ref<string | null>(null)

const modelsQuery = useModelsQuery()
const deleteMutation = useDeleteModelMutation()

const models = computed(() => modelsQuery.data.value ?? [])

const groupedModels = computed(() => {
  const groups = new Map<string, typeof models.value>()
  for (const model of models.value) {
    const list = groups.get(model.provider) ?? []
    list.push(model)
    groups.set(model.provider, list)
  }
  return Array.from(groups.entries()).sort(([a], [b]) => a.localeCompare(b))
})

const errorMessage = computed(() => {
  const error = modelsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load models'
})

const isInitialLoading = computed(
  () => modelsQuery.isPending.value && !modelsQuery.data.value,
)

async function handleDelete(id: string) {
  deletingId.value = id
  try {
    await deleteMutation.mutateAsync(id)
  } finally {
    deletingId.value = null
  }
}

function openWizard() {
  showWizard.value = true
}

function closeWizard() {
  showWizard.value = false
}

function onModelsAdded() {
  void modelsQuery.refetch()
}
</script>

<template>
  <div>
    <div class="flex justify-end mb-6">
      <AppButton class="shrink-0 self-start sm:self-auto" @click="openWizard">
        <span class="inline-flex items-center gap-2">
          <AppIcon name="plus" size="sm" />
          Add models
        </span>
      </AppButton>
    </div>

    <AppAlert
      v-if="errorMessage"
      title="Unable to load models"
      @retry="modelsQuery.refetch()"
    >
      {{ errorMessage }}
    </AppAlert>

    <LoadingState
      v-if="isInitialLoading"
      title="Loading models"
      description="Fetching configured LLM models..."
    />

    <template v-else>
      <div
        v-if="models.length > 0"
        class="mb-6 flex flex-wrap items-center gap-3 rounded-xl border border-gray-200/60 bg-white/60 backdrop-blur-sm px-4 py-3"
      >
        <div class="flex items-center gap-2 text-sm text-gray-600">
          <span class="font-semibold text-gray-900">{{ models.length }}</span>
          model{{ models.length === 1 ? '' : 's' }} configured
        </div>
        <span class="hidden sm:inline text-gray-300">|</span>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="[provider, items] in groupedModels"
            :key="provider"
            class="inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-700"
          >
            {{ formatProviderName(provider) }}
            <span class="text-slate-400">{{ items.length }}</span>
          </span>
        </div>
      </div>

      <div v-if="models.length === 0" class="text-center">
        <EmptyState
          title="No models configured"
          description="Add your first LLM model to start running prompts and collecting analytics."
          icon="server"
        />
        <AppButton class="mt-4" @click="openWizard">Add your first model</AppButton>
      </div>

      <div v-else class="space-y-8">
        <section v-for="[provider, providerModels] in groupedModels" :key="provider">
          <div class="flex items-center gap-2 mb-4">
            <h2 class="text-sm font-semibold uppercase tracking-wider text-gray-500">
              {{ formatProviderName(provider) }}
            </h2>
            <span class="text-xs text-gray-400">{{ providerModels.length }}</span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            <ModelCard
              v-for="model in providerModels"
              :key="model.id"
              :model="model"
              :deleting="deletingId === model.id"
              @delete="handleDelete"
            />
          </div>
        </section>
      </div>
    </template>

    <AddModelWizard
      v-if="showWizard"
      @close="closeWizard"
      @added="onModelsAdded"
    />
  </div>
</template>
