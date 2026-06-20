<script setup lang="ts">
import { ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import ProviderLogo from '@/components/providers/ProviderLogo.vue'
import AppButton from '@/components/ui/AppButton.vue'
import type { ModelResponse } from '@/types/model'

const props = defineProps<{
  model: ModelResponse
  deleting?: boolean
}>()

const emit = defineEmits<{
  delete: [id: string]
}>()

const confirming = ref(false)

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
</script>

<template>
  <article
    class="group rounded-xl border border-gray-200/60 bg-white/80 backdrop-blur-sm shadow-sm transition-all duration-200 hover:shadow-md hover:border-gray-300/80"
  >
    <div class="p-5">
      <div class="flex items-start gap-3 min-w-0">
        <ProviderLogo :provider="model.provider" size="lg" rounded="xl" />
        <div class="min-w-0">
          <h3 class="font-semibold text-gray-900 truncate">{{ model.name }}</h3>
          <p class="text-sm text-gray-500 truncate mt-0.5">{{ model.model }}</p>
        </div>
      </div>

      <div class="mt-4 flex items-center justify-end gap-3">
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
  </article>
</template>
