<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import {
  useSMTPSettingsQuery,
  useTestSMTPSettingsMutation,
  useUpdateSMTPSettingsMutation,
} from '@/queries/settings'

const settingsQuery = useSMTPSettingsQuery()
const updateMutation = useUpdateSMTPSettingsMutation()
const testMutation = useTestSMTPSettingsMutation()

const host = ref('')
const port = ref('587')
const username = ref('')
const password = ref('')
const fromEmail = ref('')
const fromName = ref('')
const useTls = ref(true)
const enabled = ref(false)
const testTo = ref('')
const formError = ref<string | null>(null)
const formSuccess = ref<string | null>(null)
const testError = ref<string | null>(null)
const testSuccess = ref<string | null>(null)
const hasPassword = ref(false)

const errorMessage = computed(() => {
  const error = settingsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load SMTP settings'
})

const isInitialLoading = computed(
  () => settingsQuery.isPending.value && !settingsQuery.data.value,
)

watch(
  () => settingsQuery.data.value,
  (settings) => {
    if (!settings) return
    host.value = settings.host
    port.value = String(settings.port || 587)
    username.value = settings.username
    password.value = ''
    fromEmail.value = settings.from_email
    fromName.value = settings.from_name
    useTls.value = settings.use_tls
    enabled.value = settings.enabled
    hasPassword.value = settings.has_password
    if (!testTo.value && settings.from_email) {
      testTo.value = settings.from_email
    }
  },
  { immediate: true },
)

function buildPayload() {
  const parsedPort = Number.parseInt(port.value, 10)
  return {
    host: host.value.trim(),
    port: Number.isFinite(parsedPort) ? parsedPort : 0,
    username: username.value.trim(),
    password: password.value,
    from_email: fromEmail.value.trim(),
    from_name: fromName.value.trim(),
    use_tls: useTls.value,
    enabled: enabled.value,
  }
}

async function handleSave() {
  formError.value = null
  formSuccess.value = null
  testError.value = null
  testSuccess.value = null

  try {
    const result = await updateMutation.mutateAsync(buildPayload())
    hasPassword.value = result.has_password
    password.value = ''
    formSuccess.value = 'SMTP settings saved.'
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Failed to save SMTP settings'
  }
}

async function handleTest() {
  formError.value = null
  formSuccess.value = null
  testError.value = null
  testSuccess.value = null

  const payload = buildPayload()
  try {
    await testMutation.mutateAsync({
      to: testTo.value.trim() || undefined,
      host: payload.host,
      port: payload.port,
      username: payload.username,
      password: payload.password || undefined,
      from_email: payload.from_email,
      from_name: payload.from_name,
      use_tls: payload.use_tls,
    })
    testSuccess.value = 'Test email sent successfully.'
  } catch (error) {
    testError.value = error instanceof Error ? error.message : 'SMTP test failed'
  }
}
</script>

<template>
  <div class="space-y-8">
    <AppAlert
      v-if="errorMessage"
      title="Unable to load SMTP settings"
      @retry="settingsQuery.refetch()"
    >
      {{ errorMessage }}
    </AppAlert>

    <LoadingState v-if="!errorMessage && isInitialLoading" label="Loading configuration…" />

    <section
      v-else-if="!errorMessage"
      class="rounded-xl border border-gray-200/60 bg-white/60 p-6 backdrop-blur-sm"
    >
      <div class="mb-6">
        <h2 class="text-lg font-semibold text-gray-900">SMTP</h2>
        <p class="mt-1 text-sm text-gray-600">
          Configure the outbound mail server used to send emails from Gego.
        </p>
      </div>

      <form class="space-y-5" @submit.prevent="handleSave">
        <label class="flex items-center gap-3 text-sm text-gray-800">
          <input
            v-model="enabled"
            type="checkbox"
            class="h-4 w-4 rounded border-gray-300 text-slate-600 focus:ring-slate-500"
          />
          Enable SMTP sending
        </label>

        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2 sm:col-span-2">
            <label class="text-sm font-medium text-gray-700">Host</label>
            <AppInput v-model="host" placeholder="smtp.example.com" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">Port</label>
            <AppInput v-model="port" type="number" placeholder="587" />
          </div>

          <div class="space-y-2">
            <label class="flex items-center gap-3 pt-7 text-sm text-gray-800">
              <input
                v-model="useTls"
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-slate-600 focus:ring-slate-500"
              />
              Use TLS / STARTTLS
            </label>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">Username</label>
            <AppInput v-model="username" placeholder="Optional" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">Password</label>
            <AppInput
              v-model="password"
              type="password"
              :placeholder="hasPassword ? '•••••••• (leave blank to keep)' : 'Optional'"
            />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">From email</label>
            <AppInput v-model="fromEmail" type="email" placeholder="noreply@example.com" />
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">From name</label>
            <AppInput v-model="fromName" placeholder="Gego" />
          </div>
        </div>

        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <p v-if="formSuccess" class="text-sm text-green-700">{{ formSuccess }}</p>

        <div class="flex flex-wrap gap-3">
          <AppButton type="submit" :loading="updateMutation.isPending.value">
            Save settings
          </AppButton>
        </div>
      </form>

      <div class="mt-8 border-t border-gray-200/70 pt-6">
        <h3 class="text-base font-semibold text-gray-900">Send test email</h3>
        <p class="mt-1 text-sm text-gray-600">
          Uses the values above (including an unsaved password). Leave blank to send to the from
          address.
        </p>

        <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-end">
          <div class="min-w-0 flex-1 space-y-2">
            <label class="text-sm font-medium text-gray-700">Recipient</label>
            <AppInput v-model="testTo" type="email" placeholder="you@example.com" />
          </div>
          <AppButton
            variant="secondary"
            :loading="testMutation.isPending.value"
            @click="handleTest"
          >
            Send test
          </AppButton>
        </div>

        <p v-if="testError" class="mt-3 text-sm text-red-600">{{ testError }}</p>
        <p v-if="testSuccess" class="mt-3 text-sm text-green-700">{{ testSuccess }}</p>
      </div>
    </section>
  </div>
</template>
