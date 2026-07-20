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
  useUpdateUserMutation,
  useUsersQuery,
} from '@/queries/users'
import type { Role } from '@/types/auth'

const { user: currentUser } = useAuth()

const username = ref('')
const password = ref('')
const role = ref<Role>('member')
const formError = ref<string | null>(null)
const actionError = ref<string | null>(null)
const deletingId = ref<string | null>(null)
const updatingId = ref<string | null>(null)
const resetPasswordId = ref<string | null>(null)
const newPassword = ref('')

const usersQuery = useUsersQuery()
const createMutation = useCreateUserMutation()
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

async function handleCreate() {
  formError.value = null
  const name = username.value.trim()
  if (!name) {
    formError.value = 'Username is required'
    return
  }
  if (password.value.length < 8) {
    formError.value = 'Password must be at least 8 characters'
    return
  }

  try {
    await createMutation.mutateAsync({
      username: name,
      password: password.value,
      role: role.value,
    })
    username.value = ''
    password.value = ''
    role.value = 'member'
  } catch (error) {
    formError.value = error instanceof Error ? error.message : 'Failed to create user'
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

async function handleResetPassword(id: string) {
  if (newPassword.value.length < 8) return
  actionError.value = null
  updatingId.value = id
  try {
    await updateMutation.mutateAsync({ id, payload: { password: newPassword.value } })
    resetPasswordId.value = null
    newPassword.value = ''
  } catch (error) {
    actionError.value = error instanceof Error ? error.message : 'Failed to reset password'
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
          Create admin or member accounts. Admins manage configuration; members can view the
          dashboard and search.
        </p>
      </div>

      <form class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4" @submit.prevent="handleCreate">
        <AppInput v-model="username" placeholder="Username" />
        <AppInput v-model="password" type="password" placeholder="Password (min 8)" />
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
              <th class="px-4 py-3 font-medium">Username</th>
              <th class="px-4 py-3 font-medium">Role</th>
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
              <td class="px-4 py-3 whitespace-nowrap text-gray-600">
                {{ formatDate(account.created_at) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-col items-end gap-2">
                  <div class="flex flex-wrap justify-end gap-2">
                    <AppButton
                      variant="ghost"
                      size="sm"
                      :disabled="updatingId === account.id"
                      @click="
                        resetPasswordId =
                          resetPasswordId === account.id ? null : account.id;
                        newPassword = ''
                      "
                    >
                      Reset password
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

                  <form
                    v-if="resetPasswordId === account.id"
                    class="flex w-full max-w-sm flex-col gap-2 sm:flex-row"
                    @submit.prevent="handleResetPassword(account.id)"
                  >
                    <AppInput
                      v-model="newPassword"
                      type="password"
                      class="flex-1"
                      placeholder="New password (min 8)"
                    />
                    <AppButton
                      type="submit"
                      size="sm"
                      :disabled="
                        newPassword.length < 8 || updatingId === account.id
                      "
                    >
                      Save
                    </AppButton>
                  </form>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
