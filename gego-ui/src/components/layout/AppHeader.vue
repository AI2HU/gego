<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppIcon from '@/components/icons/AppIcon.vue'
import { useSidebar } from '@/composables/useSidebar'
import { pageMeta } from '@/design/navigation'
import { header } from '@/design/classes'

const route = useRoute()
const { toggle } = useSidebar()

const page = computed(() => {
  const name = typeof route.name === 'string' ? route.name : ''
  return pageMeta[name] ?? { title: 'Gego', description: '' }
})
</script>

<template>
  <header :class="header.bar">
    <div :class="header.inner">
      <div :class="header.row">
        <div class="flex items-center gap-3 min-w-0">
          <button
            type="button"
            class="lg:hidden flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-700 hover:bg-gray-100 transition-colors duration-200"
            aria-label="Open menu"
            @click="toggle"
          >
            <AppIcon name="menu" size="lg" />
          </button>

          <div class="min-w-0">
            <h1 :class="header.title">{{ page.title }}</h1>
            <p v-if="page.description" :class="header.subtitle">{{ page.description }}</p>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
