<script setup lang="ts">
import { card } from '@/design/classes'

withDefaults(
  defineProps<{
    interactive?: boolean
    padding?: boolean
  }>(),
  {
    interactive: false,
    padding: true,
  },
)
</script>

<template>
  <section :class="interactive ? card.interactive : card.base">
    <header v-if="$slots.header" :class="card.header">
      <slot name="header" />
    </header>
    <div :class="padding ? card.body : ''">
      <slot />
    </div>
  </section>
</template>
