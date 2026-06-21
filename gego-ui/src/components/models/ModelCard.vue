<script setup lang="ts">
import { computed, ref } from 'vue'

import EditModelDialog from '@/components/models/EditModelDialog.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import AppButton from '@/components/ui/AppButton.vue'
import { useProvidersQuery, useTestModelAccessMutation } from '@/queries/models'
import type { ModelResponse } from '@/types/model'

const props = defineProps<{
  model: ModelResponse
  deleting?: boolean
}>()

const emit = defineEmits<{
  delete: [id: string]
  updated: []
}>()

const confirming = ref(false)
const showEditDialog = ref(false)
const testStatus = ref<'idle' | 'success' | 'error'>('idle')
const testMessage = ref<string | null>(null)

const testMutation = useTestModelAccessMutation()
const providersQuery = useProvidersQuery()

const consoleUrl = computed(
  () => providersQuery.data.value?.find((provider) => provider.id === props.model.provider)?.console_url,
)

function requestDelete() {
  confirming.value = true
}

function cancelDelete() {
  confirming.value = false
}

function confirmDelete() {
  emit('delete', props.model.id)
  confirming.value = false
}

function openEditDialog() {
  showEditDialog.value = true
}

function closeEditDialog() {
  showEditDialog.value = false
}

function onUpdated() {
  testStatus.value = 'idle'
  testMessage.value = null
  emit('updated')
}

async function testAccess() {
  testStatus.value = 'idle'
  testMessage.value = null

  try {
    const result = await testMutation.mutateAsync(props.model.id)
    if (result.success) {
      testStatus.value = 'success'
      testMessage.value = result.message
    } else {
      testStatus.value = 'error'
      testMessage.value = result.message
    }
  } catch (error) {
    testStatus.value = 'error'
    testMessage.value = error instanceof Error ? error.message : 'Connection test failed'
  }
}
</script>

<template>
  <article
    class="group rounded-xl border border-gray-200/60 bg-white/80 backdrop-blur-sm shadow-sm transition-all duration-200 hover:shadow-md hover:border-gray-300/80"
  >
    <div class="p-5">
      <div class="flex items-start gap-3 min-w-0">
        <ProviderLogo :provider="model.provider" size="lg" rounded="xl" />
        <div class="min-w-0 flex-1">
          <h3 class="font-semibold text-gray-900 truncate">{{ model.name }}</h3>
          <p class="text-sm text-gray-500 truncate mt-0.5">{{ model.model }}</p>
          <p
            v-if="model.api_key"
            class="text-xs text-gray-400 font-mono mt-2 truncate"
            :title="model.api_key"
          >
            Key {{ model.api_key }}
          </p>
          <p
            v-else-if="model.base_url"
            class="text-xs text-gray-400 font-mono mt-2 truncate"
            :title="model.base_url"
          >
            {{ model.base_url }}
          </p>
          <a
            v-if="consoleUrl"
            :href="consoleUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1 text-xs text-slate-600 hover:text-slate-800 mt-2"
          >
            Provider console
            <AppIcon name="external-link" size="sm" />
          </a>
        </div>
      </div>

      <p
        v-if="testMessage"
        class="mt-3 text-xs rounded-lg px-2.5 py-1.5"
        :class="
          testStatus === 'success'
            ? 'bg-green-50 text-green-700 border border-green-200/60'
            : 'bg-red-50 text-red-700 border border-red-200/60'
        "
      >
        {{ testMessage }}
      </p>

      <div class="mt-4 flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <AppButton
            variant="secondary"
            size="sm"
            :loading="testMutation.isPending.value"
            @click="testAccess"
          >
            <AppIcon v-if="testStatus === 'success' && !testMutation.isPending.value" name="check" size="sm" />
            Test access
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="openEditDialog">
            <AppIcon name="settings" size="sm" />
            {{ model.provider === 'ollama' ? 'Edit URL' : 'Update key' }}
          </AppButton>
        </div>

        <div v-if="!confirming">
          <AppButton
            variant="ghost"
            size="sm"
            :loading="deleting"
            class="!text-red-600 hover:!bg-red-50 opacity-0 group-hover:opacity-100 transition-opacity"
            @click="requestDelete"
          >
            <AppIcon name="trash" size="sm" />
            Remove
          </AppButton>
        </div>

        <div v-else class="flex items-center gap-2">
          <span class="text-xs text-red-600 font-medium">Remove?</span>
          <AppButton variant="danger" size="sm" :loading="deleting" @click="confirmDelete">
            Yes
          </AppButton>
          <AppButton variant="ghost" size="sm" :disabled="deleting" @click="cancelDelete">
            No
          </AppButton>
        </div>
      </div>
    </div>

    <EditModelDialog
      v-if="showEditDialog"
      :model="model"
      @close="closeEditDialog"
      @updated="onUpdated"
    />
  </article>
</template>
