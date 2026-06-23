<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import { formatProviderName } from '@/lib/providers'
import { useModelsQuery } from '@/queries/models'
import { usePromptsQuery } from '@/queries/prompts'
import { useCreateScheduleMutation } from '@/queries/schedules'
import type { CronPreset } from '@/types/schedule'
import { CRON_PRESETS, CRON_PRESET_HINTS, CRON_PRESET_LABELS } from '@/types/schedule'

const frequencyOptions: { id: CronPreset; label: string; hint: string }[] = [
  { id: 'daily', label: CRON_PRESET_LABELS.daily, hint: CRON_PRESET_HINTS.daily },
  { id: 'weekly', label: CRON_PRESET_LABELS.weekly, hint: CRON_PRESET_HINTS.weekly },
  { id: 'monthly', label: CRON_PRESET_LABELS.monthly, hint: CRON_PRESET_HINTS.monthly },
  { id: 'custom', label: 'Custom', hint: 'Enter your own cron expression' },
]

const emit = defineEmits<{
  close: []
  added: []
}>()

type WizardStep = 'details' | 'prompts' | 'models' | 'saving'

const step = ref<WizardStep>('details')
const name = ref('')
const cronPreset = ref<CronPreset>('daily')
const customCron = ref('')
const temperatureMode = ref<'default' | 'random' | 'custom'>('default')
const customTemperature = ref('0.7')
const selectedPromptIds = ref<Set<string>>(new Set())
const selectedModelIds = ref<Set<string>>(new Set())
const activeTagFilter = ref<string | null>(null)
const searchQuery = ref('')
const errorMessage = ref<string | null>(null)

const promptsQuery = usePromptsQuery()
const modelsQuery = useModelsQuery()
const createMutation = useCreateScheduleMutation()

const prompts = computed(() => promptsQuery.data.value ?? [])
const models = computed(() => modelsQuery.data.value ?? [])

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
  let result = prompts.value

  if (activeTagFilter.value) {
    result = result.filter((prompt) => (prompt.tags ?? []).includes(activeTagFilter.value!))
  }

  const query = searchQuery.value.trim().toLowerCase()
  if (!query) {
    return result
  }

  return result.filter((prompt) => {
    if (prompt.template.toLowerCase().includes(query)) {
      return true
    }

    return (prompt.tags ?? []).some((tag) => tag.toLowerCase().includes(query))
  })
})

const hasActiveFilters = computed(
  () => activeTagFilter.value !== null || searchQuery.value.trim().length > 0,
)

const allFilteredPromptsSelected = computed(
  () =>
    filteredPrompts.value.length > 0 &&
    filteredPrompts.value.every((prompt) => selectedPromptIds.value.has(prompt.id)),
)

const modelsByProvider = computed(() => {
  const groups = new Map<string, typeof models.value>()
  for (const model of models.value) {
    const list = groups.get(model.provider) ?? []
    list.push(model)
    groups.set(model.provider, list)
  }
  return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b))
})

const allProvidersSelected = computed(
  () =>
    modelsByProvider.value.length > 0 &&
    modelsByProvider.value.every(([, providerModels]) =>
      providerModels.some((model) => selectedModelIds.value.has(model.id)),
    ),
)

const stepNumber = computed(() => {
  switch (step.value) {
    case 'details':
      return 1
    case 'prompts':
      return 2
    case 'models':
    case 'saving':
      return 3
    default:
      return 1
  }
})

const isLoading = computed(
  () =>
    (step.value === 'prompts' && promptsQuery.isPending.value && !promptsQuery.data.value) ||
    (step.value === 'models' && modelsQuery.isPending.value && !modelsQuery.data.value),
)

function closeWizard() {
  emit('close')
}

function goBack() {
  errorMessage.value = null
  if (step.value === 'prompts') {
    step.value = 'details'
  } else if (step.value === 'models') {
    step.value = 'prompts'
  }
}

function resolveCronExpr(): string {
  if (cronPreset.value === 'custom') {
    return customCron.value.trim()
  }
  return CRON_PRESETS[cronPreset.value]
}

function resolveTemperature(): number {
  if (temperatureMode.value === 'random') {
    return -1
  }
  if (temperatureMode.value === 'custom') {
    return Number.parseFloat(customTemperature.value)
  }
  return 0.7
}

function validateDetails(): boolean {
  if (!name.value.trim()) {
    errorMessage.value = 'Schedule name is required'
    return false
  }

  const cronExpr = resolveCronExpr()
  if (!cronExpr) {
    errorMessage.value = 'Cron expression is required'
    return false
  }

  if (temperatureMode.value === 'custom') {
    const temp = Number.parseFloat(customTemperature.value)
    if (Number.isNaN(temp) || temp < 0 || temp > 1) {
      errorMessage.value = 'Temperature must be between 0.0 and 1.0'
      return false
    }
  }

  errorMessage.value = null
  return true
}

function continueFromDetails() {
  if (!validateDetails()) return
  step.value = 'prompts'
}

function continueFromPrompts() {
  if (selectedPromptIds.value.size === 0) {
    errorMessage.value = 'Select at least one prompt'
    return
  }
  errorMessage.value = null
  step.value = 'models'
}

function togglePrompt(id: string) {
  const next = new Set(selectedPromptIds.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  selectedPromptIds.value = next
}

function toggleAllPrompts() {
  const next = new Set(selectedPromptIds.value)

  if (allFilteredPromptsSelected.value) {
    for (const prompt of filteredPrompts.value) {
      next.delete(prompt.id)
    }
  } else {
    for (const prompt of filteredPrompts.value) {
      next.add(prompt.id)
    }
  }

  selectedPromptIds.value = next
}

function applyTagFilter(tag: string) {
  activeTagFilter.value = activeTagFilter.value === tag ? null : tag
}

function clearFilters() {
  activeTagFilter.value = null
  searchQuery.value = ''
}

function toggleModel(id: string) {
  const model = models.value.find((entry) => entry.id === id)
  if (!model) return

  const next = new Set(selectedModelIds.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    for (const other of models.value) {
      if (other.provider === model.provider) {
        next.delete(other.id)
      }
    }
    next.add(id)
  }
  selectedModelIds.value = next
}

function toggleAllModels() {
  if (allProvidersSelected.value) {
    selectedModelIds.value = new Set()
    return
  }

  const next = new Set<string>()
  for (const [, providerModels] of modelsByProvider.value) {
    if (providerModels[0]) {
      next.add(providerModels[0].id)
    }
  }
  selectedModelIds.value = next
}

async function saveSchedule() {
  if (selectedModelIds.value.size === 0) {
    errorMessage.value = 'Select at least one model'
    return
  }

  errorMessage.value = null
  step.value = 'saving'

  try {
    await createMutation.mutateAsync({
      name: name.value.trim(),
      prompt_ids: Array.from(selectedPromptIds.value),
      llm_ids: Array.from(selectedModelIds.value),
      cron_expr: resolveCronExpr(),
      temperature: resolveTemperature(),
      enabled: true,
    })
    emit('added')
    closeWizard()
  } catch (error) {
    step.value = 'models'
    errorMessage.value = error instanceof Error ? error.message : 'Failed to create schedule'
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
        aria-label="Close dialog"
        @click="closeWizard"
      />

      <div
        class="relative w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl border border-gray-200/80 bg-white shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="add-schedule-title"
      >
        <div class="border-b border-gray-200/60 px-6 py-5 bg-gradient-to-r from-slate-50 to-white">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
                Step {{ stepNumber }} of 3
              </p>
              <h2 id="add-schedule-title" class="text-xl font-semibold text-gray-900 mt-1">
                Add schedule
              </h2>
            </div>
            <button
              type="button"
              class="rounded-lg p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
              aria-label="Close"
              @click="closeWizard"
            >
              <AppIcon name="close" size="sm" />
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto px-6 py-5">
          <AppAlert v-if="errorMessage" title="Unable to continue">
            {{ errorMessage }}
          </AppAlert>

          <div v-if="step === 'details'" class="space-y-5">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Name</label>
              <AppInput v-model="name" placeholder="Daily brand monitoring" />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Frequency</label>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <label
                  v-for="option in frequencyOptions"
                  :key="option.id"
                  class="flex items-start gap-3 rounded-xl border p-3 cursor-pointer transition-colors"
                  :class="cronPreset === option.id ? 'border-slate-400 bg-slate-50' : 'border-gray-200/80'"
                >
                  <input v-model="cronPreset" type="radio" :value="option.id" class="mt-1" />
                  <span>
                    <span class="block text-sm font-medium text-gray-900">{{ option.label }}</span>
                    <span class="block text-xs text-gray-500 mt-0.5">{{ option.hint }}</span>
                  </span>
                </label>
              </div>
              <AppInput
                v-if="cronPreset === 'custom'"
                v-model="customCron"
                class="mt-3"
                placeholder="*/15 * * * *"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Temperature</label>
              <div class="space-y-2">
                <label class="flex items-center gap-2 text-sm text-gray-700">
                  <input v-model="temperatureMode" type="radio" value="default" />
                  Default (0.7)
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-700">
                  <input v-model="temperatureMode" type="radio" value="random" />
                  Random (0.0–1.0 per run)
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-700">
                  <input v-model="temperatureMode" type="radio" value="custom" />
                  Custom
                </label>
              </div>
              <AppInput
                v-if="temperatureMode === 'custom'"
                v-model="customTemperature"
                class="mt-3"
                placeholder="0.7"
              />
            </div>
          </div>

          <LoadingState
            v-else-if="step === 'prompts' && isLoading"
            title="Loading prompts"
            description="Fetching available prompt templates..."
          />

          <div v-else-if="step === 'prompts'" class="space-y-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-gray-600">
                Select prompts to include in this schedule.
              </p>
              <AppButton
                variant="ghost"
                size="sm"
                :disabled="filteredPrompts.length === 0"
                @click="toggleAllPrompts"
              >
                {{ allFilteredPromptsSelected ? 'Deselect all' : 'Select all' }}
              </AppButton>
            </div>

            <div v-if="prompts.length === 0" class="rounded-xl border border-dashed border-gray-200 p-6 text-center text-sm text-gray-500">
              No prompts available. Add prompts first.
            </div>

            <template v-else>
              <div class="space-y-3 rounded-xl border border-gray-200/60 bg-slate-50/60 p-3">
                <AppInput
                  v-model="searchQuery"
                  placeholder="Search prompt text or tags..."
                />

                <div v-if="allTags.length > 0" class="flex flex-wrap items-center gap-2">
                  <span class="text-xs font-medium uppercase tracking-wider text-gray-400">
                    Filter by tag
                  </span>
                  <button
                    v-for="tag in allTags"
                    :key="tag"
                    type="button"
                    class="inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium transition-colors"
                    :class="
                      activeTagFilter === tag
                        ? 'border-slate-400 bg-slate-200 text-slate-900'
                        : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50'
                    "
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
                    <span v-if="selectedPromptIds.size > 0">
                      ·
                      <span class="font-semibold text-gray-900">{{ selectedPromptIds.size }}</span>
                      selected
                    </span>
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

              <div
                v-if="filteredPrompts.length === 0"
                class="rounded-xl border border-dashed border-gray-200 p-6 text-center text-sm text-gray-500"
              >
                No prompts match your filters. Try a different search or tag.
              </div>

              <div v-else class="space-y-2 max-h-80 overflow-y-auto">
                <label
                  v-for="prompt in filteredPrompts"
                  :key="prompt.id"
                  class="flex items-start gap-3 rounded-lg border border-gray-200/80 p-3 cursor-pointer hover:bg-slate-50"
                >
                  <input
                    type="checkbox"
                    class="mt-1"
                    :checked="selectedPromptIds.has(prompt.id)"
                    @change="togglePrompt(prompt.id)"
                  />
                  <span class="min-w-0">
                    <span class="block text-sm text-gray-800">{{ prompt.template }}</span>
                    <span v-if="prompt.tags?.length" class="mt-1.5 flex flex-wrap gap-1">
                      <span
                        v-for="tag in prompt.tags"
                        :key="tag"
                        class="inline-flex rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-600"
                      >
                        {{ tag }}
                      </span>
                    </span>
                  </span>
                </label>
              </div>
            </template>
          </div>

          <LoadingState
            v-else-if="step === 'models' && isLoading"
            title="Loading models"
            description="Fetching configured LLM models..."
          />

          <div v-else-if="step === 'models'" class="space-y-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-gray-600">
                Select one model per provider. Each prompt runs once per provider per schedule run.
              </p>
              <AppButton variant="ghost" size="sm" @click="toggleAllModels">
                {{ allProvidersSelected ? 'Deselect all' : 'Select all providers' }}
              </AppButton>
            </div>

            <div v-if="models.length === 0" class="rounded-xl border border-dashed border-gray-200 p-6 text-center text-sm text-gray-500">
              No models available. Add models first.
            </div>

            <div v-else class="space-y-4 max-h-80 overflow-y-auto">
              <section
                v-for="[provider, providerModels] in modelsByProvider"
                :key="provider"
                class="rounded-lg border border-gray-200/80 overflow-hidden"
              >
                <div class="flex items-center gap-2 px-3 py-2 bg-slate-50 border-b border-gray-200/80">
                  <ProviderLogo :provider="provider" class="h-5 w-5" />
                  <span class="text-sm font-medium text-gray-900">{{ formatProviderName(provider) }}</span>
                </div>
                <div class="divide-y divide-gray-100">
                  <label
                    v-for="model in providerModels"
                    :key="model.id"
                    class="flex items-start gap-3 p-3 cursor-pointer hover:bg-slate-50"
                  >
                    <input
                      type="radio"
                      class="mt-1"
                      :name="`provider-${provider}`"
                      :checked="selectedModelIds.has(model.id)"
                      @change="toggleModel(model.id)"
                    />
                    <span>
                      <span class="block text-sm font-medium text-gray-900">{{ model.name }}</span>
                      <span class="block text-xs text-gray-500 mt-0.5">{{ model.model }}</span>
                    </span>
                  </label>
                </div>
              </section>
            </div>
          </div>

          <LoadingState
            v-else-if="step === 'saving'"
            title="Creating schedule"
            description="Saving your new schedule..."
          />
        </div>

        <div class="border-t border-gray-200/60 px-6 py-4 bg-slate-50/50 flex items-center justify-between gap-3">
          <AppButton
            v-if="step !== 'details' && step !== 'saving'"
            variant="ghost"
            @click="goBack"
          >
            Back
          </AppButton>
          <div v-else />

          <div class="flex items-center gap-2">
            <AppButton variant="secondary" @click="closeWizard">Cancel</AppButton>

            <AppButton v-if="step === 'details'" @click="continueFromDetails">
              Continue
            </AppButton>

            <AppButton
              v-else-if="step === 'prompts'"
              :disabled="prompts.length === 0"
              @click="continueFromPrompts"
            >
              Continue
            </AppButton>

            <AppButton
              v-else-if="step === 'models'"
              :disabled="models.length === 0"
              :loading="createMutation.isPending.value"
              @click="saveSchedule"
            >
              Create schedule
            </AppButton>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
