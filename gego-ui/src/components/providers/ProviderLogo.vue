<script setup lang="ts">
import { computed } from 'vue'

import { formatProviderName, getProviderLogo } from '@/lib/providers'

const props = withDefaults(
  defineProps<{
    provider: string
    size?: 'sm' | 'md' | 'lg'
    rounded?: 'lg' | 'xl'
  }>(),
  {
    size: 'md',
    rounded: 'lg',
  },
)

const boxClass = computed(() => {
  const size =
    props.size === 'sm'
      ? 'h-8 w-8 p-1'
      : props.size === 'lg'
        ? 'h-11 w-11 p-1.5'
        : 'h-10 w-10 p-1.5'

  const radius = props.rounded === 'xl' ? 'rounded-xl' : 'rounded-lg'

  return `${size} ${radius}`
})
</script>

<template>
  <div
    :class="[
      boxClass,
      'flex shrink-0 items-center justify-center border border-gray-200/70 bg-white shadow-sm',
    ]"
  >
    <img
      :src="getProviderLogo(provider)"
      :alt="formatProviderName(provider)"
      class="h-full w-full object-contain"
    />
  </div>
</template>
