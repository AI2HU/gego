<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppButton from '@/components/ui/AppButton.vue'
import { formatProviderName, getProviderStyle } from '@/lib/providers'
import type { ModelResponse } from '@/types/model'

const props = defineProps<{
  model: ModelResponse
  deleting?: boolean
}>()

const emit = defineEmits<{
  delete: [id: string]
}>()

const confirming = ref(false)

const style = computed(() => getProviderStyle(props.model.provider))

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
      <div class="flex items-start justify-between gap-3">
        <div class="flex items-start gap-3 min-w-0">
          <div
            class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br text-sm font-bold text-white shadow-sm"
            :class="style.gradient"
          >
            {{ style.initial }}
          </div>
          <div class="min-w-0">
            <h3 class="font-semibold text-gray-900 truncate">{{ model.name }}</h3>
            <p class="text-sm text-gray-500 truncate mt-0.5">{{ model.model }}</p>
          </div>
        </div>

        <span
          class="shrink-0 inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium"
          :class="style.badge"
        >
          {{ formatProviderName(model.provider) }}
        </span>
      </div>

      <div class="mt-4 flex items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium"
            :class="
              model.enabled
                ? 'bg-green-50 text-green-700 border border-green-200/60'
                : 'bg-gray-100 text-gray-600 border border-gray-200/60'
            "
          >
            <span
              class="h-1.5 w-1.5 rounded-full"
              :class="model.enabled ? 'bg-green-500' : 'bg-gray-400'"
            />
            {{ model.enabled ? 'Enabled' : 'Disabled' }}
          </span>
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
  </article>
</template>
