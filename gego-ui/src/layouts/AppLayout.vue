<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
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
    <div :class="page.shell">
      <AppSidebar
        :nav-items="mainNavItems"
        :connected="connected"
        :connection-label="connectionLabel"
        :loading="loading"
        :show-refresh="showRefresh"
        @refresh="$emit('refresh')"
      />

      <div :class="page.content">
        <AppHeader
          :loading="loading"
          :show-refresh="showRefresh"
          @refresh="$emit('refresh')"
        />

        <main :class="page.main">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>
