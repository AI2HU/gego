<script setup lang="ts">
import AppIcon from '@/components/icons/AppIcon.vue'

withDefaults(
  defineProps<{
    title?: string
    description?: string
  }>(),
  {
    title: 'Loading',
    description: 'Fetching data from your LLM providers...',
  },
)
</script>

<template>
  <div class="flex justify-center items-center py-16 animate-fade-in">
    <div class="text-center">
      <div class="w-12 h-12 bg-slate-200 rounded-lg flex items-center justify-center mx-auto mb-4">
        <AppIcon name="spinner" size="lg" spin class="text-slate-600" />
      </div>
      <h3 class="text-lg font-medium text-gray-700 mb-1">{{ title }}</h3>
      <p class="text-gray-500">{{ description }}</p>
    </div>
  </div>
</template>
