<script setup lang="ts">
import AppButton from '@/components/ui/AppButton.vue'
import IconBox from '@/components/ui/IconBox.vue'

defineProps<{
  title: string
  loading?: boolean
}>()

defineEmits<{
  retry: []
}>()
</script>

<template>
  <div class="mb-8 bg-red-50 border border-red-200 rounded-lg p-6 animate-fade-in">
    <div class="flex items-start">
      <div class="flex-shrink-0">
        <IconBox icon="error" size="lg" tone="danger" />
      </div>
      <div class="ml-4 flex-1">
        <h3 class="text-lg font-semibold text-red-800 mb-2">{{ title }}</h3>
        <div class="text-sm text-red-700 space-y-2">
          <p class="font-medium">
            <slot />
          </p>
          <div v-if="$slots.actions" class="mt-4 flex items-center space-x-3">
            <slot name="actions" />
          </div>
          <div v-else class="mt-4 flex items-center space-x-3">
            <AppButton variant="danger" :loading="loading" icon="refresh" @click="$emit('retry')">
              Retry Connection
            </AppButton>
            <a
              href="https://github.com/AI2HU/gego"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex"
            >
              <AppButton variant="secondary" icon="github">View Documentation</AppButton>
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
