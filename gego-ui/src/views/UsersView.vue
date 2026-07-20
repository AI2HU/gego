<script setup lang="ts">
import { computed, ref } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import { useAuth } from '@/composables/useAuth'
import {
  useCreateUserMutation,
  useDeleteUserMutation,
  useInviteUserMutation,
  useUpdateUserMutation,
  useUsersQuery,
} from '@/queries/users'
import type { Role } from '@/types/auth'

const { user: currentUser } = useAuth()

const email = ref('')
const role = ref<Role>('member')
const formError = ref<string | null>(null)
const formSuccess = ref<string | null>(null)
const inviteURL = ref<string | null>(null)
const actionError = ref<string | null>(null)
const deletingId = ref<string | null>(null)
const updatingId = ref<string | null>(null)
const invitingId = ref<string | null>(null)
const copied = ref(false)

const usersQuery = useUsersQuery()
const createMutation = useCreateUserMutation()
const inviteMutation = useInviteUserMutation()
const updateMutation = useUpdateUserMutation()
const deleteMutation = useDeleteUserMutation()

const users = computed(() => usersQuery.data.value ?? [])

const errorMessage = computed(() => {
  const error = usersQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load users'
})

const isInitialLoading = computed(
  () => usersQuery.isPending.value && !usersQuery.data.value,
)

const selectClass =
  'w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-200'

function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
}

async function handleCreate() {
  formError.value = null
  formSuccess.value = null
  inviteURL.value = null
  copied.value = false

  const value = email.value.trim()
  if (!value) {
    formError.value = 'Email is required'
    return
  }
  if (!isValidEmail(value)) {
    formError.value = 'Enter a valid email address'
    return
  }

  try {
    const result = await createMutation.mutateAsync({
      email: value,
      role: role.value,
    })
    email.value = ''
    role.value = 'member'
    inviteURL.value = result.invite_url
    formSuccess.value = result.email_sent
      ? 'Account created. An invite email was sent.'
      : 'Account created. Copy the invite link below (SMTP is not configured).'
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Failed to create user'
  }
}

async function handleInvite(id: string) {
  actionError.value = null
  invitingId.value = id
  inviteURL.value = null
  formSuccess.value = null
  copied.value = false
  try {
    const result = await inviteMutation.mutateAsync(id)
    inviteURL.value = result.invite_url
    formSuccess.value = result.email_sent
      ? 'Invite email sent.'
      : 'Invite link generated. Copy it below (SMTP is not configured).'
  } catch (error) {
    actionError.value =
      error instanceof Error ? error.message : 'Failed to generate invite link'
  } finally {
    invitingId.value = null
  }
}

async function handleRoleChange(id: string, nextRole: Role) {
  actionError.value = null
  updatingId.value = id
  try {
    await updateMutation.mutateAsync({ id, payload: { role: nextRole } })
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to update role'
    void usersQuery.refetch()
  } finally {
    updatingId.value = null
  }
}

async function handleDelete(id: string) {
  actionError.value = null
  deletingId.value = id
  try {
    await deleteMutation.mutateAsync(id)
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to delete user'
  } finally {
    deletingId.value = null
  }
}

async function copyInviteLink() {
  if (!inviteURL.value) return
  try {
    await navigator.clipboard.writeText(inviteURL.value)
    copied.value = true
  } catch {
    copied.value = false
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString()
}
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-xl border border-gray-200/60 bg-white/60 p-6 backdrop-blur-sm">
      <div class="mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Add account</h2>
        <p class="mt-1 text-sm text-gray-600">
          Create admin or member accounts by email. The user sets their password through a 1-week
          invite link (emailed when SMTP is configured).
        </p>
      </div>

      <form class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3" @submit.prevent="handleCreate">
        <AppInput v-model="email" type="email" placeholder="Email" />
        <select v-model="role" :class="selectClass">
          <option value="member">Member</option>
          <option value="admin">Admin</option>
        </select>
        <AppButton type="submit" :disabled="createMutation.isPending.value">
          <span class="inline-flex items-center gap-2">
            <AppIcon name="plus" size="sm" />
            Create account
          </span>
        </AppButton>
      </form>

      <p v-if="formError" class="mt-3 text-sm text-red-600">{{ formError }}</p>
      <p v-if="formSuccess" class="mt-3 text-sm text-green-700">{{ formSuccess }}</p>

      <div
        v-if="inviteURL"
        class="mt-4 rounded-lg border border-slate-200 bg-slate-50/80 p-4"
      >
        <p class="text-sm font-medium text-gray-800">Invite link (valid 1 week)</p>
        <p class="mt-2 break-all font-mono text-xs text-gray-600">{{ inviteURL }}</p>
        <AppButton class="mt-3" variant="secondary" size="sm" @click="copyInviteLink">
          {{ copied ? 'Copied' : 'Copy link' }}
        </AppButton>
      </div>
    </section>

    <AppAlert
      v-if="errorMessage"
      title="Unable to load users"
      @retry="usersQuery.refetch()"
    >
      {{ errorMessage }}
    </AppAlert>

    <p v-if="actionError" class="text-sm text-red-600">{{ actionError }}</p>

    <LoadingState v-if="!errorMessage && isInitialLoading" label="Loading users…" />

    <EmptyState
      v-else-if="!errorMessage && users.length === 0"
      title="No accounts yet"
      description="Create the first admin or member account to get started."
    />

    <section v-else-if="!errorMessage" class="overflow-hidden rounded-xl border border-gray-200/60 bg-white/60">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
          <thead class="bg-slate-50/80 text-xs uppercase tracking-wide text-gray-500">
            <tr>
              <th class="px-4 py-3 font-medium">Email</th>
              <th class="px-4 py-3 font-medium">Role</th>
              <th class="px-4 py-3 font-medium">Status</th>
              <th class="px-4 py-3 font-medium">Created</th>
              <th class="px-4 py-3 font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-for="account in users" :key="account.id" class="align-top">
              <td class="px-4 py-3">
                <div class="font-medium text-gray-900">{{ account.username }}</div>
                <div
                  v-if="account.id === currentUser?.id"
                  class="mt-0.5 text-xs text-gray-500"
                >
                  You
                </div>
              </td>
              <td class="px-4 py-3">
                <select
                  :value="account.role"
                  :class="[selectClass, 'max-w-[9rem]']"
                  :disabled="
                    updatingId === account.id ||
                    account.id === currentUser?.id
                  "
                  @change="
                    handleRoleChange(
                      account.id,
                      ($event.target as HTMLSelectElement).value as Role,
                    )
                  "
                >
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
              </td>
              <td class="px-4 py-3 text-gray-600">
                {{ account.password_pending ? 'Pending invite' : 'Active' }}
              </td>
              <td class="px-4 py-3 whitespace-nowrap text-gray-600">
                {{ formatDate(account.created_at) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap justify-end gap-2">
                  <AppButton
                    variant="ghost"
                    size="sm"
                    :disabled="invitingId === account.id"
                    :loading="invitingId === account.id"
                    @click="handleInvite(account.id)"
                  >
                    {{ account.password_pending ? 'Resend invite' : 'Reset via invite' }}
                  </AppButton>
                  <AppButton
                    variant="ghost"
                    size="sm"
                    :disabled="
                      account.id === currentUser?.id ||
                      deletingId === account.id
                    "
                    @click="handleDelete(account.id)"
                  >
                    <span class="inline-flex items-center gap-1.5 text-red-600">
                      <AppIcon name="trash" size="sm" />
                      Delete
                    </span>
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
