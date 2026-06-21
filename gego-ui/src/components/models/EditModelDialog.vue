<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import { useUpdateModelMutation } from '@/queries/models'
import type { ModelResponse } from '@/types/model'

const props = defineProps<{
  model: ModelResponse
}>()

const emit = defineEmits<{
  close: []
  updated: []
}>()

const apiKey = ref('')
const baseUrl = ref(props.model.base_url ?? 'http://localhost:11434')
const errorMessage = ref<string | null>(null)

const updateMutation = useUpdateModelMutation()

const isOllama = computed(() => props.model.provider === 'ollama')
const requiresApiKey = computed(() => props.model.provider !== 'ollama')

async function save() {
  errorMessage.value = null

  const payload = isOllama.value
    ? { base_url: baseUrl.value.trim() }
    : { api_key: apiKey.value.trim() }

  if (isOllama.value && !payload.base_url) {
    errorMessage.value = 'Base URL is required'
    return
  }

  if (requiresApiKey.value && !payload.api_key) {
    errorMessage.value = 'API key is required'
    return
  }

  try {
    await updateMutation.mutateAsync({ id: props.model.id, payload })
    emit('updated')
    emit('close')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to update model'
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        class="absolute inset-0 bg-gray-900/50 backdrop-blur-sm"
        aria-label="Close"
        @click="emit('close')"
      />

      <div
        class="relative w-full max-w-md overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-2xl"
      >
        <div class="border-b border-gray-200/60 px-6 py-5 bg-gradient-to-r from-slate-50 to-white">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 min-w-0">
              <ProviderLogo :provider="model.provider" />
              <div class="min-w-0">
                <h2 class="text-lg font-semibold text-gray-900 truncate">{{ model.name }}</h2>
                <p class="text-sm text-gray-500 truncate">{{ model.model }}</p>
              </div>
            </div>
            <button
              type="button"
              class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 transition-colors"
              @click="emit('close')"
            >
              <AppIcon name="close" size="md" />
            </button>
          </div>
        </div>

        <div class="px-6 py-5 space-y-4">
          <AppAlert v-if="errorMessage" title="Error">
            {{ errorMessage }}
          </AppAlert>

          <template v-if="isOllama">
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">Ollama Base URL</label>
              <AppInput v-model="baseUrl" placeholder="http://localhost:11434" />
            </div>
          </template>

          <template v-else>
            <div v-if="model.api_key" class="text-sm text-gray-500">
              Current key:
              <span class="font-mono text-gray-700">{{ model.api_key }}</span>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">New API Key</label>
              <AppInput v-model="apiKey" type="password" placeholder="Enter new API key" />
            </div>
          </template>
        </div>

        <div class="border-t border-gray-200/60 px-6 py-4 bg-gray-50/80 flex items-center justify-end gap-2">
          <AppButton variant="secondary" :disabled="updateMutation.isPending.value" @click="emit('close')">
            Cancel
          </AppButton>
          <AppButton :loading="updateMutation.isPending.value" @click="save">
            Save
          </AppButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>
