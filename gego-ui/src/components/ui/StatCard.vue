<script setup lang="ts">
import type { AppIconName } from '@/components/icons/AppIcon.vue'
import IconBox from '@/components/ui/IconBox.vue'
import { card, typography } from '@/design/classes'

defineProps<{
  label: string
  value: string | number
  hint?: string
  icon: AppIconName
}>()
</script>

<template>
  <div :class="[card.interactive, 'p-6']">
    <div class="flex items-center justify-between gap-4">
      <div class="min-w-0">
        <p :class="typography.label">{{ label }}</p>
        <p :class="[typography.value, 'mt-1']">{{ value }}</p>
        <p v-if="hint" :class="[typography.hint, 'mt-1']">{{ hint }}</p>
      </div>
      <IconBox :icon="icon" size="lg" />
    </div>
  </div>
</template>
