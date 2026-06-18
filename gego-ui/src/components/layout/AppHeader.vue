<script setup lang="ts">
import AppIcon from '@/components/icons/AppIcon.vue'
import AppButton from '@/components/ui/AppButton.vue'
import { useSidebar } from '@/composables/useSidebar'
import { appMeta } from '@/design/navigation'
import { header } from '@/design/classes'

withDefaults(
  defineProps<{
    loading?: boolean
    showRefresh?: boolean
  }>(),
  {
    loading: false,
    showRefresh: true,
  },
)

const emit = defineEmits<{
  refresh: []
}>()

const { toggle } = useSidebar()
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

          <div class="lg:hidden flex items-center gap-2 min-w-0">
            <div :class="header.logoBox">
              <AppIcon name="desktop" size="md" class="text-white" />
            </div>
            <div class="min-w-0">
              <h1 :class="header.title">{{ appMeta.title }}</h1>
            </div>
          </div>

          <div class="hidden lg:block min-w-0">
            <h1 :class="header.title">{{ appMeta.title }}</h1>
            <p :class="header.subtitle">{{ appMeta.subtitle }}</p>
          </div>
        </div>

        <div class="flex items-center gap-2 sm:gap-4 shrink-0">
          <AppButton
            v-if="showRefresh"
            class="hidden sm:inline-flex"
            :loading="loading"
            icon="refresh"
            size="sm"
            @click="emit('refresh')"
          >
            <span class="hidden md:inline">Refresh</span>
          </AppButton>
        </div>
      </div>
    </div>
  </header>
</template>
