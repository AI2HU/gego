<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { useDashboardStatus } from '@/composables/useDashboardStatus'
import { navSections } from '@/design/navigation'
import { page } from '@/design/classes'

const { connected, connectionLabel } = useDashboardStatus()
</script>

<template>
  <div :class="page.root">
    <div :class="page.shell">
      <AppSidebar
        :nav-sections="navSections"
        :connected="connected"
        :connection-label="connectionLabel"
      />

      <div :class="page.content">
        <AppHeader />

        <main :class="page.main">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>
