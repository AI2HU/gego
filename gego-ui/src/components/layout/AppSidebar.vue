<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppButton from '@/components/ui/AppButton.vue'
import NavLink from '@/components/ui/NavLink.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { useSidebar } from '@/composables/useSidebar'
import { useAuth } from '@/composables/useAuth'
import { appMeta, type NavItem } from '@/design/navigation'
import { sidebar, typography } from '@/design/classes'

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

const router = useRouter()
const { user, logout: signOut } = useAuth()
const { isVisible, isLgUp, close } = useSidebar()

const username = computed(() => user.value?.username ?? '')

const panelClass = computed(() => [
  sidebar.panel,
  isVisible.value ? sidebar.panelOpen : sidebar.panelClosed,
])

function onNavigate() {
  if (!isLgUp.value) {
    close()
  }
}

function logout() {
  signOut()
  close()
  router.push({ name: 'login' })
}

function refresh() {
  emit('refresh')
  if (!isLgUp.value) {
    close()
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-300"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-300"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <button
        v-if="isVisible && !isLgUp"
        type="button"
        :class="sidebar.backdrop"
        aria-label="Close menu"
        @click="close"
      />
    </Transition>
  </Teleport>

  <aside :class="panelClass" aria-label="Main navigation">
    <div :class="sidebar.brand">
      <div :class="sidebar.brandLogo">
        <AppIcon name="desktop" size="lg" class="text-white" />
      </div>
      <div class="min-w-0">
        <p :class="sidebar.brandTitle">{{ appMeta.title }}</p>
        <p :class="sidebar.brandSubtitle">{{ appMeta.subtitle }}</p>
      </div>
    </div>

    <nav v-if="navItems.length" :class="sidebar.nav" aria-label="Menu">
      <ul :class="sidebar.navList">
        <li v-for="item in navItems" :key="item.to">
          <NavLink
            :to="item.to"
            :icon="item.icon"
            variant="sidebar"
            exact
            @click="onNavigate"
          >
            {{ item.label }}
          </NavLink>
        </li>
      </ul>
    </nav>

    <div v-else :class="sidebar.nav" />

    <div :class="sidebar.footer">
      <StatusBadge :connected="connected" :label="connectionLabel" compact />

      <div v-if="username" class="min-w-0">
        <p :class="typography.overline">Signed in as</p>
        <p class="truncate text-sm font-medium text-gray-800">{{ username }}</p>
      </div>

      <div :class="sidebar.footerActions">
        <AppButton
          v-if="showRefresh"
          variant="secondary"
          size="sm"
          class="flex-1"
          :loading="loading"
          icon="refresh"
          @click="refresh"
        >
          Refresh
        </AppButton>
        <AppButton variant="ghost" size="sm" class="shrink-0" @click="logout">
          Sign out
        </AppButton>
      </div>
    </div>
  </aside>
</template>
