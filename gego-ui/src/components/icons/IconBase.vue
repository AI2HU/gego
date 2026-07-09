<script setup lang="ts">
import { computed, useId } from 'vue'

const props = withDefaults(
  defineProps<{
    iconName?: string
    width?: number | string
    height?: number | string
    viewBox?: string
    size?: 'sm' | 'md' | 'lg'
    spin?: boolean
    decorative?: boolean
    filled?: boolean
  }>(),
  {
    iconName: 'icon',
    viewBox: '0 0 24 24',
    size: 'md',
    spin: false,
    decorative: true,
    filled: false,
  },
)

const titleId = useId()

const sizeClass = computed(() => {
  if (props.width != null || props.height != null) {
    return ''
  }

  switch (props.size) {
    case 'sm':
      return 'h-4 w-4'
    case 'lg':
      return 'h-6 w-6'
    default:
      return 'h-5 w-5'
  }
})
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :class="[sizeClass, spin ? 'animate-spin' : '']"
    :width="width"
    :height="height"
    :viewBox="viewBox"
    :fill="filled ? 'currentColor' : 'none'"
    :stroke="filled ? 'none' : 'currentColor'"
    stroke-width="1.75"
    stroke-linecap="round"
    stroke-linejoin="round"
    :role="decorative ? 'presentation' : 'img'"
    :aria-hidden="decorative ? 'true' : undefined"
    :aria-labelledby="decorative ? undefined : titleId"
  >
    <title v-if="!decorative" :id="titleId">{{ iconName }} icon</title>
    <slot />
  </svg>
</template>
