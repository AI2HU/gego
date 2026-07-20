<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { ApiError } from '@/api/client'
import { useAuth } from '@/composables/useAuth'
import { appMeta } from '@/design/navigation'
import { header, page, typography } from '@/design/classes'
import gegoLogo from '@/assets/gego_logo.svg'

const router = useRouter()
const route = useRoute()
const { login, isLoggingIn } = useAuth()

const username = ref('')
const password = ref('')
const error = ref<string | null>(null)

const redirectPath = computed(() => {
  const redirect = route.query.redirect
  return typeof redirect === 'string' && redirect.startsWith('/') ? redirect : '/'
})

async function submit() {
  error.value = null

  if (!username.value.trim() || !password.value) {
    error.value = 'Username and password are required.'
    return
  }

  try {
    await login({
      username: username.value.trim(),
      password: password.value,
    })
    await router.replace(redirectPath.value)
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      error.value = err.message || 'Invalid username or password.'
      return
    }

    error.value = err instanceof Error ? err.message : 'Login failed. Please try again.'
  }
}
</script>

<template>
  <div :class="page.root">
    <div class="min-h-screen flex flex-col items-center justify-center px-4 py-12">
      <div class="mb-8 flex items-center space-x-3">
        <div :class="header.logoBox">
          <img :src="gegoLogo" alt="Gego" class="h-full w-full" />
        </div>
        <div>
          <h1 :class="header.title">{{ appMeta.title }}</h1>
          <p class="text-sm text-gray-500">{{ appMeta.subtitle }}</p>
        </div>
      </div>

      <AppCard class="w-full max-w-md">
        <template #header>
          <div>
            <h2 :class="typography.overline">Sign in</h2>
            <p class="mt-1 text-sm text-gray-600">Use your Gego account to access the dashboard.</p>
          </div>
        </template>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label for="username" :class="[typography.label, 'block mb-2']">Username</label>
            <AppInput
              id="username"
              v-model="username"
              autocomplete="username"
              placeholder="Enter your username"
              :disabled="isLoggingIn"
            />
          </div>

          <div>
            <label for="password" :class="[typography.label, 'block mb-2']">Password</label>
            <AppInput
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              placeholder="Enter your password"
              :disabled="isLoggingIn"
              @enter="submit"
            />
          </div>

          <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

          <AppButton type="submit" block :loading="isLoggingIn">Sign in</AppButton>
        </form>
      </AppCard>
    </div>
  </div>
</template>
