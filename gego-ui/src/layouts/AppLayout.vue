<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import { mainNavItems } from '@/design/navigation'
import { page } from '@/design/classes'

withDefaults(
  defineProps<{
    connected?: boolean
    connectionLabel?: string
    loading?: boolean
    showRefresh?: boolean
  }>(),
  {
    connected: true,
    connectionLabel: undefined,
    loading: false,
    showRefresh: false,
  },
)

defineEmits<{
  refresh: []
}>()
</script>

<template>
  <div :class="page.root">
    <AppHeader
      :nav-items="mainNavItems"
      :connected="connected"
      :connection-label="connectionLabel"
      :loading="loading"
      :show-refresh="showRefresh"
      @refresh="$emit('refresh')"
    />

    <main :class="page.main">
      <RouterView />
    </main>
  </div>
</template>
