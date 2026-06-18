<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon, { type AppIconName } from '@/components/icons/AppIcon.vue'
import { nav } from '@/design/classes'

const props = withDefaults(
  defineProps<{
    to: string
    exact?: boolean
    icon?: AppIconName
    variant?: 'default' | 'sidebar'
  }>(),
  {
    exact: false,
    variant: 'default',
  },
)

const route = useRoute()

const isActive = computed(() => {
  if (props.exact) {
    return route.path === props.to
  }
  return route.path === props.to || route.path.startsWith(`${props.to}/`)
})

const linkClass = computed(() => {
  if (props.variant === 'sidebar') {
    return [nav.sidebarLink, isActive.value ? nav.sidebarActive : nav.sidebarInactive]
  }
  return [nav.link, isActive.value ? nav.active : nav.inactive]
})
</script>

<template>
  <RouterLink :to="to" :class="linkClass">
    <AppIcon v-if="icon" :name="icon" size="sm" class="shrink-0" />
    <span><slot /></span>
  </RouterLink>
</template>
