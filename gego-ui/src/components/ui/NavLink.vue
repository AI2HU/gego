<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { nav } from '@/design/classes'

const props = defineProps<{
  to: string
  exact?: boolean
}>()

const route = useRoute()

const isActive = computed(() => {
  if (props.exact) {
    return route.path === props.to
  }
  return route.path === props.to || route.path.startsWith(`${props.to}/`)
})
</script>

<template>
  <RouterLink :to="to" :class="[nav.link, isActive ? nav.active : nav.inactive]">
    <slot />
  </RouterLink>
</template>
