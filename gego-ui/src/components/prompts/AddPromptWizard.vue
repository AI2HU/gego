<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import { input } from '@/design/classes'
import { formatProviderName, getProviderStyle } from '@/lib/providers'
import { useModelsQuery } from '@/queries/models'
import { useCreatePromptMutation, useCreatePromptsMutation, useGeneratePromptsMutation } from '@/queries/prompts'

const emit = defineEmits<{
  close: []
  added: []
}>()

type WizardStep =
  | 'method'
  | 'custom'
  | 'generate-model'
  | 'generate-config'
  | 'generate-preview'
  | 'saving'

type AddMethod = 'generate' | 'custom'

const step = ref<WizardStep>('method')
const method = ref<AddMethod | null>(null)
const template = ref('')
const tagsInput = ref('')
const generatedTagsInput = ref('')
const selectedModelId = ref<string | null>(null)
const languageCode = ref('EN')
const userInput = ref('')
const promptCount = ref('20')
const generatedPrompts = ref<string[]>([])
const selectedPrompts = ref<Set<number>>(new Set())
const errorMessage = ref<string | null>(null)
const savedCount = ref(0)

const modelsQuery = useModelsQuery()
const createMutation = useCreatePromptMutation()
const generateMutation = useGeneratePromptsMutation()
const createPromptsMutation = useCreatePromptsMutation()

const models = computed(() => modelsQuery.data.value ?? [])

const selectedModel = computed(
  () => models.value.find((model) => model.id === selectedModelId.value) ?? null,
)

const allGeneratedSelected = computed(
  () =>
    generatedPrompts.value.length > 0 &&
    selectedPrompts.value.size === generatedPrompts.value.length,
)

const totalSteps = computed(() => (method.value === 'custom' ? 2 : 4))

const stepNumber = computed(() => {
  switch (step.value) {
    case 'method':
      return 1
    case 'custom':
    case 'generate-model':
      return 2
    case 'generate-config':
      return 3
    case 'generate-preview':
    case 'saving':
      return 4
    default:
      return 1
  }
})

const stepTitle = computed(() => {
  switch (step.value) {
    case 'method':
      return 'Choose how to add prompts'
    case 'custom':
      return 'Write a custom prompt'
    case 'generate-model':
      return 'Select a model'
    case 'generate-config':
      return 'Describe your topic'
    case 'generate-preview':
      return 'Review generated prompts'
    case 'saving':
      return 'Saving prompts'
    default:
      return 'Add prompts'
  }
})

function selectMethod(next: AddMethod) {
  method.value = next
  errorMessage.value = null

  if (next === 'custom') {
    step.value = 'custom'
    return
  }

  if (models.value.length === 0) {
    errorMessage.value = 'No models configured. Add a model first.'
  }

  step.value = 'generate-model'
}

function goBack() {
  errorMessage.value = null

  switch (step.value) {
    case 'custom':
    case 'generate-model':
      step.value = 'method'
      method.value = null
      break
    case 'generate-config':
      step.value = 'generate-model'
      break
    case 'generate-preview':
      step.value = 'generate-config'
      generatedPrompts.value = []
      selectedPrompts.value = new Set()
      break
    default:
      step.value = 'method'
  }
}

function parseTagsInput(value: string): string[] {
  if (!value.trim()) {
    return []
  }

  return value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
}

function parseTags(): string[] {
  return parseTagsInput(tagsInput.value)
}

function buildGeneratedTags(): string[] {
  const tags = ['generated', ...parseTagsInput(generatedTagsInput.value)]
  return [...new Set(tags)]
}

async function saveCustomPrompt() {
  if (!template.value.trim()) {
    errorMessage.value = 'Prompt template cannot be empty'
    return
  }

  errorMessage.value = null

  try {
    await createMutation.mutateAsync({
      template: template.value.trim(),
      tags: parseTags(),
      enabled: true,
    })
    emit('added')
    emit('close')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to create prompt'
  }
}

function selectModel(modelId: string) {
  selectedModelId.value = modelId
  step.value = 'generate-config'
  errorMessage.value = null
}

async function runGenerate() {
  if (!selectedModel.value) {
    errorMessage.value = 'Select a model first'
    return
  }

  const count = Number.parseInt(promptCount.value, 10)
  if (!userInput.value.trim()) {
    errorMessage.value = 'Description is required'
    return
  }
  if (!Number.isFinite(count) || count < 1) {
    errorMessage.value = 'Enter a valid prompt count'
    return
  }
  if (!languageCode.value.trim()) {
    errorMessage.value = 'Language code is required'
    return
  }

  errorMessage.value = null

  try {
    const result = await generateMutation.mutateAsync({
      llm_id: selectedModel.value.id,
      language_code: languageCode.value.trim().toUpperCase(),
      user_input: userInput.value.trim(),
      prompt_count: count,
    })

    generatedPrompts.value = result.prompts
    selectedPrompts.value = new Set(result.prompts.map((_, index) => index))
    step.value = 'generate-preview'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to generate prompts'
  }
}

function toggleGeneratedPrompt(index: number) {
  const next = new Set(selectedPrompts.value)
  if (next.has(index)) {
    next.delete(index)
  } else {
    next.add(index)
  }
  selectedPrompts.value = next
}

function toggleSelectAllGenerated() {
  if (allGeneratedSelected.value) {
    selectedPrompts.value = new Set()
  } else {
    selectedPrompts.value = new Set(generatedPrompts.value.map((_, index) => index))
  }
}

async function saveSelectedPrompts() {
  if (selectedPrompts.value.size === 0) {
    errorMessage.value = 'Select at least one prompt to save'
    return
  }

  errorMessage.value = null
  step.value = 'saving'

  const templates = [...selectedPrompts.value]
    .map((index) => generatedPrompts.value[index]?.trim())
    .filter((template): template is string => Boolean(template))

  try {
    const result = await createPromptsMutation.mutateAsync({
      prompts: templates.map((template) => ({ template })),
      tags: parseTagsInput(generatedTagsInput.value),
    })
    savedCount.value = result.saved_count
    emit('added')
    emit('close')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to save prompts'
    step.value = 'generate-preview'
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
                Step {{ stepNumber }} of {{ totalSteps }}
              </p>
              <h2 class="text-xl font-semibold text-gray-900 mt-1">Add Prompts</h2>
              <p class="text-sm text-gray-500 mt-0.5">{{ stepTitle }}</p>
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
              v-for="n in totalSteps"
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

          <div v-if="step === 'method'" class="space-y-4">
            <button
              type="button"
              class="group relative w-full text-left rounded-2xl border-2 border-indigo-200 bg-gradient-to-br from-indigo-50 via-white to-violet-50 p-5 transition-all duration-200 hover:border-indigo-300 hover:shadow-md hover:-translate-y-0.5"
              @click="selectMethod('generate')"
            >
              <span
                class="absolute top-4 right-4 inline-flex items-center rounded-full bg-indigo-600 px-2.5 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-white"
              >
                Fastest
              </span>

              <div class="flex items-start gap-4 pr-24">
                <div
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 text-white shadow-sm"
                >
                  <AppIcon name="lightbulb" size="md" />
                </div>
                <div class="min-w-0">
                  <p class="text-base font-semibold text-gray-900">
                    Generate using LLM
                  </p>
                  <p class="mt-1 text-sm text-gray-600">
                    Easiest way to start LLM visibility stats — describe a topic, get prompts in
                    seconds.
                  </p>
                </div>
              </div>
            </button>

            <button
              type="button"
              class="group w-full text-left rounded-xl border border-gray-200/80 p-4 transition-all duration-200 hover:border-slate-300 hover:bg-slate-50/80"
              @click="selectMethod('custom')"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500 to-purple-600 text-white shadow-sm"
                >
                  <AppIcon name="comment" size="sm" />
                </div>
                <div>
                  <p class="font-medium text-gray-900">Add a custom prompt</p>
                  <p class="text-xs text-gray-500 mt-1">Write one prompt yourself.</p>
                </div>
              </div>
            </button>
          </div>

          <div v-else-if="step === 'custom'" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Prompt template</label>
              <textarea
                v-model="template"
                rows="6"
                placeholder="What are the top streaming services for watching movies?"
                :class="input.base"
                class="resize-y min-h-[140px]"
              />
              <p class="mt-2 text-xs text-gray-500">Sent to LLMs for keyword tracking.</p>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Tags
                <span class="font-normal text-gray-400">(optional, comma-separated)</span>
              </label>
              <AppInput
                v-model="tagsInput"
                placeholder="streaming, entertainment, movies"
              />
            </div>
          </div>

          <LoadingState
            v-else-if="step === 'generate-model' && modelsQuery.isPending.value"
            title="Loading models"
            description="Fetching configured LLM models..."
          />

          <div v-else-if="step === 'generate-model'" class="space-y-4">
            <p v-if="models.length === 0" class="text-sm text-gray-600">
              Add a model first.
            </p>
            <p v-else class="text-sm text-gray-600">
              Which model should draft your prompts?
            </p>

            <div v-if="models.length > 0" class="grid grid-cols-1 gap-3">
              <button
                v-for="model in models"
                :key="model.id"
                type="button"
                class="group text-left rounded-xl border border-gray-200/80 p-4 transition-all duration-200 hover:border-slate-300 hover:shadow-md bg-white"
                @click="selectModel(model.id)"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br text-sm font-bold text-white shadow-sm"
                    :class="getProviderStyle(model.provider).gradient"
                  >
                    {{ getProviderStyle(model.provider).initial }}
                  </div>
                  <div class="min-w-0">
                    <p class="font-medium text-gray-900 truncate">{{ model.name }}</p>
                    <p class="text-xs text-gray-500 truncate">
                      {{ formatProviderName(model.provider) }} · {{ model.model }}
                    </p>
                  </div>
                </div>
              </button>
            </div>
          </div>

          <div v-else-if="step === 'generate-config'" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Language</label>
              <AppInput v-model="languageCode" placeholder="EN" />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Topic</label>
              <textarea
                v-model="userInput"
                rows="4"
                placeholder="Streaming services in France"
                :class="input.base"
                class="resize-y min-h-[100px]"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">Count</label>
              <AppInput v-model="promptCount" type="number" placeholder="20" />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Default tags
                <span class="font-normal text-gray-400">(optional, comma-separated)</span>
              </label>
              <AppInput
                v-model="generatedTagsInput"
                placeholder="streaming, france, competitors"
              />
              <p class="mt-2 text-xs text-gray-500">
                Always includes
                <span class="font-medium text-slate-600">generated</span>.
              </p>
              <div v-if="buildGeneratedTags().length > 0" class="mt-2 flex flex-wrap gap-1.5">
                <span
                  v-for="tag in buildGeneratedTags()"
                  :key="tag"
                  class="inline-flex rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>

          <LoadingState
            v-else-if="step === 'saving'"
            title="Saving prompts"
            :description="`Saved ${savedCount} of ${selectedPrompts.size} prompts...`"
          />

          <div v-else-if="step === 'generate-preview'" class="space-y-4">
            <p class="text-sm text-gray-600">
              {{ generatedPrompts.length }} generated · pick what to keep
            </p>

            <div class="flex flex-wrap items-center gap-2 text-sm text-gray-600">
              <span class="flex flex-wrap gap-1">
                <span
                  v-for="tag in buildGeneratedTags()"
                  :key="tag"
                  class="inline-flex rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700"
                >
                  {{ tag }}
                </span>
              </span>
            </div>

            <div class="flex items-center justify-end">
              <AppButton variant="ghost" size="sm" @click="toggleSelectAllGenerated">
                {{ allGeneratedSelected ? 'Deselect all' : 'Select all' }}
              </AppButton>
            </div>

            <div class="space-y-2 max-h-80 overflow-y-auto">
              <label
                v-for="(promptText, index) in generatedPrompts"
                :key="index"
                class="flex items-start gap-3 rounded-lg border border-gray-200/80 p-3 cursor-pointer hover:bg-slate-50"
              >
                <input
                  type="checkbox"
                  class="mt-1"
                  :checked="selectedPrompts.has(index)"
                  @change="toggleGeneratedPrompt(index)"
                />
                <span class="text-sm text-gray-800">{{ promptText }}</span>
              </label>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-200/60 px-6 py-4 bg-slate-50/50 flex items-center justify-between gap-3">
          <AppButton
            v-if="step !== 'method' && step !== 'saving'"
            variant="ghost"
            @click="goBack"
          >
            Back
          </AppButton>
          <div v-else />

          <div class="flex items-center gap-2">
            <AppButton variant="secondary" @click="closeWizard">Cancel</AppButton>

            <AppButton
              v-if="step === 'custom'"
              :loading="createMutation.isPending.value"
              @click="saveCustomPrompt"
            >
              Save prompt
            </AppButton>

            <AppButton
              v-else-if="step === 'generate-config'"
              :loading="generateMutation.isPending.value"
              @click="runGenerate"
            >
              Generate prompts
            </AppButton>

            <AppButton
              v-else-if="step === 'generate-preview'"
              :disabled="selectedPrompts.size === 0"
              @click="saveSelectedPrompts"
            >
              Save selected ({{ selectedPrompts.size }})
            </AppButton>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
