<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import AppInput from '@/components/ui/AppInput.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import {
  discoverProviderModels,
  useCreateModelMutation,
  useProviderApiKeysQuery,
  useProvidersQuery,
} from '@/queries/models'
import type { ModelInfo, ProviderInfo } from '@/types/model'

const emit = defineEmits<{
  close: []
  added: []
}>()

type WizardStep = 'provider' | 'credentials' | 'models' | 'adding'

const step = ref<WizardStep>('provider')
const selectedProvider = ref<ProviderInfo | null>(null)
const apiKeyMode = ref<'existing' | 'new'>('new')
const selectedKeyIndex = ref<number | null>(null)
const newApiKey = ref('')
const baseUrl = ref('http://localhost:11434')
const availableModels = ref<ModelInfo[]>([])
const selectedModelIds = ref<Set<string>>(new Set())
const loadingModels = ref(false)
const errorMessage = ref<string | null>(null)
const addedCount = ref(0)

const providersQuery = useProvidersQuery()
const createMutation = useCreateModelMutation()

const providerId = computed(() => selectedProvider.value?.id ?? null)
const apiKeysQuery = useProviderApiKeysQuery(() => providerId.value)

const providers = computed(() => providersQuery.data.value ?? [])
const existingKeys = computed(() => apiKeysQuery.data.value ?? [])

const selectedModels = computed(() =>
  availableModels.value.filter((model) => selectedModelIds.value.has(model.id)),
)

const allSelected = computed(
  () =>
    availableModels.value.length > 0 &&
    selectedModelIds.value.size === availableModels.value.length,
)

const stepNumber = computed(() => {
  switch (step.value) {
    case 'provider':
      return 1
    case 'credentials':
      return 2
    case 'models':
    case 'adding':
      return 3
    default:
      return 1
  }
})

watch(existingKeys, (keys) => {
  if (keys.length > 0) {
    apiKeyMode.value = 'existing'
    selectedKeyIndex.value = keys[0]?.index ?? 0
  } else {
    apiKeyMode.value = 'new'
    selectedKeyIndex.value = null
  }
})

function selectProvider(provider: ProviderInfo) {
  selectedProvider.value = provider
  errorMessage.value = null
  step.value = 'credentials'
}

function goBack() {
  errorMessage.value = null
  if (step.value === 'credentials') {
    step.value = 'provider'
    selectedProvider.value = null
  } else if (step.value === 'models') {
    step.value = 'credentials'
    availableModels.value = []
    selectedModelIds.value = new Set()
  }
}

async function fetchModels() {
  if (!selectedProvider.value) return

  errorMessage.value = null
  loadingModels.value = true

  try {
    const payload =
      selectedProvider.value.requires_api_key
        ? apiKeyMode.value === 'existing' && selectedKeyIndex.value !== null
          ? { existing_key_index: selectedKeyIndex.value }
          : { api_key: newApiKey.value }
        : { base_url: baseUrl.value || 'http://localhost:11434' }

    if (selectedProvider.value.requires_api_key && apiKeyMode.value === 'new' && !newApiKey.value) {
      errorMessage.value = 'API key is required'
      return
    }

    const models = await discoverProviderModels(selectedProvider.value.id, payload)
    availableModels.value = models
    selectedModelIds.value = new Set()
    step.value = 'models'

    if (models.length === 0) {
      errorMessage.value = 'No models found for this provider'
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to fetch models'
  } finally {
    loadingModels.value = false
  }
}

function toggleModel(modelId: string) {
  const next = new Set(selectedModelIds.value)
  if (next.has(modelId)) {
    next.delete(modelId)
  } else {
    next.add(modelId)
  }
  selectedModelIds.value = next
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedModelIds.value = new Set()
  } else {
    selectedModelIds.value = new Set(availableModels.value.map((model) => model.id))
  }
}

async function addSelectedModels() {
  if (!selectedProvider.value || selectedModels.value.length === 0) return

  errorMessage.value = null
  step.value = 'adding'
  addedCount.value = 0

  const credentials =
    selectedProvider.value.requires_api_key
      ? apiKeyMode.value === 'existing' && selectedKeyIndex.value !== null
        ? { existing_key_index: selectedKeyIndex.value }
        : { api_key: newApiKey.value }
      : { base_url: baseUrl.value || 'http://localhost:11434' }

  for (const model of selectedModels.value) {
    try {
      await createMutation.mutateAsync({
        name: model.name,
        provider: selectedProvider.value.id,
        model: model.id,
        enabled: true,
        ...credentials,
      })
      addedCount.value++
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : 'Failed to add model'
      break
    }
  }

  if (addedCount.value > 0) {
    emit('added')
  }

  if (addedCount.value === selectedModels.value.length) {
    emit('close')
  } else {
    step.value = 'models'
  }
}

function closeWizard() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        class="absolute inset-0 bg-gray-900/50 backdrop-blur-sm"
        aria-label="Close"
        @click="closeWizard"
      />

      <div
        class="relative w-full max-w-2xl max-h-[90vh] overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-2xl flex flex-col"
      >
        <div class="border-b border-gray-200/60 px-6 py-5 bg-gradient-to-r from-slate-50 to-white">
          <div class="flex items-center justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
                Step {{ stepNumber }} of 3
              </p>
              <h2 class="text-xl font-semibold text-gray-900 mt-1">Add Models</h2>
              <p class="text-sm text-gray-500 mt-0.5">
                {{
                  step === 'provider'
                    ? 'Choose an LLM provider'
                    : step === 'credentials'
                      ? 'Configure provider credentials'
                      : 'Select models to add'
                }}
              </p>
            </div>
            <button
              type="button"
              class="flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 transition-colors"
              @click="closeWizard"
            >
              <AppIcon name="close" size="md" />
            </button>
          </div>

          <div class="mt-4 flex gap-2">
            <div
              v-for="n in 3"
              :key="n"
              class="h-1 flex-1 rounded-full transition-colors duration-300"
              :class="n <= stepNumber ? 'bg-slate-600' : 'bg-gray-200'"
            />
          </div>
        </div>

        <div class="flex-1 overflow-y-auto px-6 py-5">
          <AppAlert v-if="errorMessage" title="Error">
            {{ errorMessage }}
          </AppAlert>

          <LoadingState
            v-if="providersQuery.isPending.value && step === 'provider'"
            title="Loading providers"
            description="Fetching available LLM providers..."
          />

          <div v-else-if="step === 'provider'" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <button
              v-for="provider in providers"
              :key="provider.id"
              type="button"
              class="group text-left rounded-xl border border-gray-200/80 p-4 transition-all duration-200 hover:border-slate-300 hover:shadow-md hover:-translate-y-0.5 bg-white"
              @click="selectProvider(provider)"
            >
              <div class="flex items-center gap-3">
                <ProviderLogo :provider="provider.id" />
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 group-hover:text-slate-700">
                    {{ provider.display_name }}
                  </p>
                  <p class="text-xs text-gray-500 mt-0.5 truncate">{{ provider.id }}</p>
                </div>
              </div>
            </button>
          </div>

          <div v-else-if="step === 'credentials'" class="space-y-5">
            <AppCard :padding="true">
              <div class="flex items-center gap-3 mb-4">
                <ProviderLogo :provider="selectedProvider!.id" />
                <div>
                  <p class="font-medium text-gray-900">{{ selectedProvider!.display_name }}</p>
                  <a
                    v-if="selectedProvider!.console_url"
                    :href="selectedProvider!.console_url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-xs text-slate-600 hover:text-slate-800 inline-flex items-center gap-1 mt-0.5"
                  >
                    Get API key
                    <AppIcon name="external-link" size="sm" />
                  </a>
                </div>
              </div>

              <template v-if="selectedProvider!.requires_api_key">
                <div v-if="existingKeys.length > 0" class="space-y-3">
                  <p class="text-sm font-medium text-gray-700">API Key</p>
                  <div class="space-y-2">
                    <label
                      v-for="key in existingKeys"
                      :key="key.index"
                      class="flex items-center gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
                      :class="
                        apiKeyMode === 'existing' && selectedKeyIndex === key.index
                          ? 'border-slate-500 bg-slate-50'
                          : 'border-gray-200 hover:border-gray-300'
                      "
                    >
                      <input
                        type="radio"
                        :name="'api-key-' + selectedProvider!.id"
                        :checked="apiKeyMode === 'existing' && selectedKeyIndex === key.index"
                        class="text-slate-600"
                        @change="
                          () => {
                            apiKeyMode = 'existing'
                            selectedKeyIndex = key.index
                          }
                        "
                      />
                      <span class="font-mono text-sm text-gray-700">{{ key.masked }}</span>
                      <span class="text-xs text-gray-400 ml-auto">Existing</span>
                    </label>

                    <label
                      class="flex items-center gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
                      :class="
                        apiKeyMode === 'new'
                          ? 'border-slate-500 bg-slate-50'
                          : 'border-gray-200 hover:border-gray-300'
                      "
                    >
                      <input
                        type="radio"
                        :name="'api-key-' + selectedProvider!.id"
                        :checked="apiKeyMode === 'new'"
                        class="text-slate-600"
                        @change="apiKeyMode = 'new'"
                      />
                      <span class="text-sm text-gray-700">Use a new API key</span>
                    </label>
                  </div>

                  <div v-if="apiKeyMode === 'new'" class="mt-3">
                    <AppInput
                      v-model="newApiKey"
                      type="password"
                      placeholder="Enter your API key"
                    />
                  </div>
                </div>

                <div v-else class="space-y-2">
                  <label class="text-sm font-medium text-gray-700">API Key</label>
                  <AppInput v-model="newApiKey" type="password" placeholder="Enter your API key" />
                </div>
              </template>

              <template v-else-if="selectedProvider!.requires_base_url">
                <div class="space-y-2">
                  <label class="text-sm font-medium text-gray-700">Ollama Base URL</label>
                  <AppInput
                    v-model="baseUrl"
                    placeholder="http://localhost:11434"
                  />
                  <p class="text-xs text-gray-500">
                    Make sure Ollama is running and accessible at this URL.
                  </p>
                </div>
              </template>
            </AppCard>
          </div>

          <div v-else-if="step === 'models' || step === 'adding'" class="space-y-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-gray-600">
                <span class="font-medium text-gray-900">{{ availableModels.length }}</span>
                models available
              </p>
              <AppButton variant="ghost" size="sm" @click="toggleSelectAll">
                {{ allSelected ? 'Deselect all' : 'Select all' }}
              </AppButton>
            </div>

            <div class="space-y-2 max-h-80 overflow-y-auto pr-1">
              <label
                v-for="model in availableModels"
                :key="model.id"
                class="flex items-start gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
                :class="
                  selectedModelIds.has(model.id)
                    ? 'border-slate-500 bg-slate-50'
                    : 'border-gray-200 hover:border-gray-300'
                "
              >
                <input
                  type="checkbox"
                  :checked="selectedModelIds.has(model.id)"
                  class="mt-0.5 rounded text-slate-600"
                  :disabled="step === 'adding'"
                  @change="toggleModel(model.id)"
                />
                <div class="min-w-0 flex-1">
                  <p class="font-medium text-gray-900 text-sm">{{ model.name }}</p>
                  <p v-if="model.description" class="text-xs text-gray-500 mt-0.5 line-clamp-2">
                    {{ model.description }}
                  </p>
                  <div class="flex items-center gap-2 mt-1.5">
                    <span
                      v-if="model.used_in_chat"
                      class="inline-flex items-center rounded-full bg-green-50 px-2 py-0.5 text-[10px] font-medium text-green-700 border border-green-200/60"
                    >
                      Chat
                    </span>
                    <span class="text-[10px] text-gray-400 font-mono truncate">{{ model.id }}</span>
                  </div>
                </div>
              </label>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-200/60 px-6 py-4 bg-gray-50/80 flex items-center justify-between gap-3">
          <AppButton
            v-if="step !== 'provider'"
            variant="ghost"
            :disabled="step === 'adding'"
            @click="goBack"
          >
            Back
          </AppButton>
          <div v-else />

          <div class="flex items-center gap-2">
            <AppButton variant="secondary" :disabled="step === 'adding'" @click="closeWizard">
              Cancel
            </AppButton>

            <AppButton
              v-if="step === 'credentials'"
              :loading="loadingModels"
              @click="fetchModels"
            >
              Fetch models
            </AppButton>

            <AppButton
              v-else-if="step === 'models'"
              :disabled="selectedModels.length === 0"
              @click="addSelectedModels"
            >
              Add {{ selectedModels.length || '' }} model{{ selectedModels.length === 1 ? '' : 's' }}
            </AppButton>

            <AppButton v-else-if="step === 'adding'" loading disabled>
              Adding {{ addedCount }}/{{ selectedModels.length }}...
            </AppButton>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
