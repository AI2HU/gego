import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import { createUser, deleteUser, fetchUsers, updateUser } from '@/api/users'
import type { CreateUserRequest, UpdateUserRequest } from '@/types/auth'

export const usersQueryKeys = {
  all: ['users'] as const,
  list: ['users', 'list'] as const,
}

export function usersListQueryOptions() {
  return queryOptions({
    queryKey: usersQueryKeys.list,
    queryFn: fetchUsers,
    staleTime: 30_000,
  })
}

export function useUsersQuery() {
  return useQuery(usersListQueryOptions())
}

export function useCreateUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateUserRequest) => createUser(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: usersQueryKeys.all })
    },
  })
}

export function useUpdateUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateUserRequest }) =>
      updateUser(id, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: usersQueryKeys.all })
    },
  })
}

export function useDeleteUserMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: usersQueryKeys.all })
    },
  })
}
