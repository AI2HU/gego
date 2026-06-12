<script setup lang="ts">
import type { AppIconName } from '@/components/icons/AppIcon.vue'
import IconBox from '@/components/ui/IconBox.vue'

withDefaults(
  defineProps<{
    title: string
    description?: string
    icon?: AppIconName
  }>(),
  {
    icon: 'file',
  },
)
</script>

<template>
  <div class="text-center text-gray-500 py-8">
    <div class="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center mx-auto mb-3">
      <IconBox :icon="icon" size="lg" class="!bg-transparent text-gray-400" />
    </div>
    <p class="font-medium">{{ title }}</p>
    <p v-if="description" class="text-sm mt-1">{{ description }}</p>
  </div>
</template>
