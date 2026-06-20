<script setup lang="ts">
import { computed } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import { button } from '@/design/classes'

type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost' | 'chip'
type ButtonSize = 'sm' | 'md' | 'lg'

const props = withDefaults(
  defineProps<{
    variant?: ButtonVariant
    size?: ButtonSize
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
    loading?: boolean
    icon?: 'refresh' | 'search' | 'github' | null
    block?: boolean
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    disabled: false,
    loading: false,
    icon: null,
    block: false,
  },
)

const classes = computed(() => {
  if (props.variant === 'chip') {
    return [button.base, button.chip, props.block ? 'w-full' : ''].filter(Boolean).join(' ')
  }

  return [button.base, button[props.size], button[props.variant], props.block ? 'w-full' : '']
    .filter(Boolean)
    .join(' ')
})
</script>

<template>
  <button :type="type" :disabled="disabled || loading" :class="classes">
    <AppIcon v-if="loading" name="spinner" size="sm" spin />
    <AppIcon v-else-if="icon" :name="icon" size="sm" />
    <slot />
  </button>
</template>
