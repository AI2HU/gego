<script setup lang="ts">
import type { AppIconName } from '@/components/icons/AppIcon.vue'
import IconBox from '@/components/ui/IconBox.vue'
import { card } from '@/design/classes'

defineProps<{
  title: string
  subtitle?: string
  icon?: AppIconName
  iconTone?: 'default' | 'success' | 'danger'
}>()
</script>

<template>
  <div class="flex items-center space-x-3">
    <IconBox v-if="icon" :icon="icon" :tone="iconTone" />
    <div>
      <h2 :class="card.sectionTitle">{{ title }}</h2>
      <p v-if="subtitle" :class="card.sectionSubtitle">{{ subtitle }}</p>
    </div>
  </div>
</template>
