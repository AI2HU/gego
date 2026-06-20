<script setup lang="ts">
import { computed } from 'vue'

import type { AppIconName } from '@/components/icons/AppIcon.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import { iconBox } from '@/design/classes'

const props = withDefaults(
  defineProps<{
    icon: AppIconName
    size?: 'sm' | 'md' | 'lg'
    tone?: 'default' | 'success' | 'danger'
  }>(),
  {
    size: 'sm',
    tone: 'default',
  },
)

const boxClass = computed(() => {
  const base = iconBox[props.size]
  if (props.tone === 'success') {
    return `${base} bg-green-100 text-green-600`
  }
  if (props.tone === 'danger') {
    return `${base} bg-red-100 text-red-600`
  }
  return `${base} text-slate-600`
})
</script>

<template>
  <div :class="boxClass">
    <AppIcon :name="icon" :size="size === 'lg' ? 'lg' : 'sm'" />
  </div>
</template>
