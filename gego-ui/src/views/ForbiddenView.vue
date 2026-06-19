<script setup lang="ts">
import { useRouter } from 'vue-router'

import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import { useAuth } from '@/composables/useAuth'
import { card } from '@/design/classes'

const router = useRouter()
const { logout: signOut } = useAuth()

function goBack() {
  router.push({ name: 'dashboard' })
}

function logout() {
  void signOut().finally(() => {
    router.push({ name: 'login' })
  })
}
</script>

<template>
  <div class="max-w-lg mx-auto">
    <AppCard>
      <div class="text-center space-y-4">
        <h2 :class="card.sectionTitle">Access denied</h2>
        <p :class="card.sectionSubtitle">
          You do not have permission to view this page.
        </p>
        <div class="flex items-center justify-center gap-3 pt-2">
          <AppButton variant="secondary" @click="goBack">Go to dashboard</AppButton>
          <AppButton variant="ghost" @click="logout">Sign out</AppButton>
        </div>
      </div>
    </AppCard>
  </div>
</template>
