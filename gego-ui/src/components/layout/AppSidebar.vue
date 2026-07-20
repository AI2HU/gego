<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import AppButton from '@/components/ui/AppButton.vue'
import NavLink from '@/components/ui/NavLink.vue'
import { useAuth } from '@/composables/useAuth'
import { useSidebar } from '@/composables/useSidebar'
import { canAccessRoute } from '@/auth/permissions'
import { appMeta, type NavSection } from '@/design/navigation'
import { sidebar, typography } from '@/design/classes'
import gegoLogo from '@/assets/gego_logo.svg'

const props = withDefaults(
  defineProps<{
    navSections?: NavSection[]
  }>(),
  {
    navSections: () => [],
  },
)

const router = useRouter()
const { user, logout: signOut } = useAuth()
const { isVisible, isLgUp, close } = useSidebar()

const username = computed(() => user.value?.username ?? '')
const userRole = computed(() => user.value?.role ?? 'member')

const visibleSections = computed(() =>
  props.navSections
    .filter((section) => !section.adminOnly || userRole.value === 'admin')
    .map((section) => ({
      ...section,
      items: section.items.filter(
        (item) => !item.permission || canAccessRoute(userRole.value, item.permission),
      ),
    }))
    .filter((section) => section.items.length > 0),
)

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
  void signOut().finally(() => {
    close()
    router.push({ name: 'login' })
  })
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
        <img :src="gegoLogo" alt="Gego" class="h-full w-full" />
      </div>
      <div class="min-w-0">
        <p :class="sidebar.brandTitle">{{ appMeta.title }}</p>
        <p :class="sidebar.brandSubtitle">{{ appMeta.subtitle }}</p>
      </div>
    </div>

    <nav v-if="visibleSections.length" :class="sidebar.nav" aria-label="Menu">
      <div v-for="(section, sectionIndex) in visibleSections" :key="sectionIndex">
        <p v-if="section.label" :class="sidebar.sectionLabel">{{ section.label }}</p>
        <ul :class="sidebar.navList">
          <li v-for="item in section.items" :key="item.to">
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
      </div>
    </nav>

    <div v-else :class="sidebar.nav" />

    <div :class="sidebar.footer">
      <div v-if="username" class="min-w-0">
        <p :class="typography.overline">Signed in as</p>
        <p class="truncate text-sm font-medium text-gray-800">{{ username }}</p>
      </div>

      <div :class="sidebar.footerActions">
        <AppButton variant="ghost" size="sm" class="w-full" @click="logout">
          Sign out
        </AppButton>
      </div>
    </div>
  </aside>
</template>
