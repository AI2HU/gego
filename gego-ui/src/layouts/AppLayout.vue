<script setup lang="ts">
import { RouterView } from 'vue-router'

import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { navSections } from '@/design/navigation'
import { page } from '@/design/classes'
</script>

<template>
  <div :class="page.root">
    <div :class="page.shell">
      <AppSidebar :nav-sections="navSections" />

      <div :class="page.content">
        <AppHeader />

        <main :class="page.main">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>
