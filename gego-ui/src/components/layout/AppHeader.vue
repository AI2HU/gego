<script setup lang="ts">
import { ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppButton from '@/components/ui/AppButton.vue'
import NavLink from '@/components/ui/NavLink.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { appMeta, type NavItem } from '@/design/navigation'
import { header, nav } from '@/design/classes'

withDefaults(
  defineProps<{
    navItems?: NavItem[]
    connected?: boolean
    connectionLabel?: string
    loading?: boolean
    showRefresh?: boolean
  }>(),
  {
    navItems: () => [],
    connected: true,
    connectionLabel: undefined,
    loading: false,
    showRefresh: true,
  },
)

const emit = defineEmits<{
  refresh: []
}>()

const mobileMenuOpen = ref(false)

const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}
</script>

<template>
  <header :class="header.bar">
    <div :class="header.inner">
      <div :class="header.row">
        <div class="flex items-center space-x-2 md:space-x-4">
          <div :class="header.logoBox">
            <AppIcon name="desktop" size="lg" class="text-white" />
          </div>
          <div>
            <h1 :class="header.title">{{ appMeta.title }}</h1>
            <p :class="header.subtitle">{{ appMeta.subtitle }}</p>
          </div>
        </div>

        <div class="flex items-center space-x-2 md:space-x-4">
          <nav v-if="navItems.length" class="hidden md:flex items-center space-x-2 mr-4">
            <NavLink v-for="item in navItems" :key="item.to" :to="item.to" exact>
              {{ item.label }}
            </NavLink>
          </nav>

          <StatusBadge
            class="hidden md:flex"
            :connected="connected"
            :label="connectionLabel"
          />

          <AppButton
            v-if="showRefresh"
            class="hidden md:inline-flex"
            :loading="loading"
            icon="refresh"
            @click="emit('refresh')"
          >
            Refresh
          </AppButton>

          <button
            type="button"
            class="md:hidden p-2 rounded-lg text-gray-700 hover:bg-gray-100 transition-colors duration-200"
            aria-label="Toggle menu"
            @click="mobileMenuOpen = !mobileMenuOpen"
          >
            <AppIcon :name="mobileMenuOpen ? 'close' : 'menu'" size="lg" />
          </button>
        </div>
      </div>

      <div v-if="mobileMenuOpen && navItems.length" :class="nav.mobilePanel">
        <nav class="flex flex-col space-y-2">
          <NavLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            exact
            @click="closeMobileMenu"
          >
            {{ item.label }}
          </NavLink>
        </nav>

        <div class="flex items-center justify-between pt-2 border-t border-gray-200/50">
          <StatusBadge :connected="connected" :label="connectionLabel" compact />
          <AppButton
            v-if="showRefresh"
            size="sm"
            :loading="loading"
            icon="refresh"
            @click="
              () => {
                emit('refresh')
                closeMobileMenu()
              }
            "
          >
            Refresh
          </AppButton>
        </div>
      </div>
    </div>
  </header>
</template>
