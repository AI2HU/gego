<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppButton from '@/components/ui/AppButton.vue'
import AppCard from '@/components/ui/AppCard.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { ApiError } from '@/api/client'
import { appMeta } from '@/design/navigation'
import { header, page, typography } from '@/design/classes'
import { useSetPasswordMutation } from '@/queries/auth'
import gegoLogo from '@/assets/gego_logo.svg'

const router = useRouter()
const route = useRoute()
const setPasswordMutation = useSetPasswordMutation()

const password = ref('')
const confirmPassword = ref('')
const error = ref<string | null>(null)

const token = computed(() => {
  const value = route.query.token
  return typeof value === 'string' ? value : ''
})

const isSubmitting = computed(() => setPasswordMutation.isPending.value)

async function submit() {
  error.value = null

  if (!token.value) {
    error.value = 'This invite link is missing a token.'
    return
  }
  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match.'
    return
  }

  try {
    await setPasswordMutation.mutateAsync({
      token: token.value,
      password: password.value,
    })
    await router.replace('/')
  } catch (err) {
    if (err instanceof ApiError) {
      error.value = err.message
      return
    }
    error.value = err instanceof Error ? err.message : 'Failed to set password.'
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
            <h2 :class="typography.overline">Set password</h2>
            <p class="mt-1 text-sm text-gray-600">
              Choose a password to finish setting up your account. This invite link expires after 1
              week.
            </p>
          </div>
        </template>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label for="password" :class="[typography.label, 'block mb-2']">Password</label>
            <AppInput
              id="password"
              v-model="password"
              type="password"
              autocomplete="new-password"
              placeholder="At least 8 characters"
              :disabled="isSubmitting"
            />
          </div>

          <div>
            <label for="confirm-password" :class="[typography.label, 'block mb-2']">
              Confirm password
            </label>
            <AppInput
              id="confirm-password"
              v-model="confirmPassword"
              type="password"
              autocomplete="new-password"
              placeholder="Repeat your password"
              :disabled="isSubmitting"
              @enter="submit"
            />
          </div>

          <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

          <AppButton type="submit" block :loading="isSubmitting">Save password</AppButton>
        </form>
      </AppCard>
    </div>
  </div>
</template>
