import { computed, type Ref } from 'vue'
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

import {
  createSchedule,
  deleteSchedule,
  fetchSchedules,
  runSchedule,
  updateSchedule,
} from '@/api/schedules'
import { fetchScheduleRun, fetchScheduleRuns } from '@/api/scheduleRuns'
import { fetchSchedulerStatus, reloadScheduler, startScheduler, stopScheduler } from '@/api/scheduler'
import { dashboardQueryKeys } from '@/queries/dashboard'
import type { CreateScheduleRequest, UpdateScheduleRequest } from '@/types/schedule'

export const schedulesQueryKeys = {
  all: ['schedules'] as const,
  list: ['schedules', 'list'] as const,
  status: ['scheduler', 'status'] as const,
  runsList: (limit: number) => ['schedule-runs', 'list', limit] as const,
  run: (id: string) => ['schedule-runs', id] as const,
}

export function schedulesListQueryOptions() {
  return queryOptions({
    queryKey: schedulesQueryKeys.list,
    queryFn: fetchSchedules,
    staleTime: 30_000,
  })
}

export function schedulerStatusQueryOptions() {
  return queryOptions({
    queryKey: schedulesQueryKeys.status,
    queryFn: fetchSchedulerStatus,
    refetchInterval: 10_000,
    staleTime: 5_000,
    retry: false,
  })
}

export function scheduleRunsListQueryOptions(limit = 100) {
  return queryOptions({
    queryKey: schedulesQueryKeys.runsList(limit),
    queryFn: () => fetchScheduleRuns(undefined, undefined, limit),
    refetchInterval: 10_000,
    staleTime: 5_000,
    retry: false,
  })
}

export function useSchedulesQuery() {
  return useQuery(schedulesListQueryOptions())
}

export function useSchedulerStatusQuery() {
  return useQuery(schedulerStatusQueryOptions())
}

export function useScheduleRunsListQuery(limit = 100) {
  return useQuery(scheduleRunsListQueryOptions(limit))
}

export function useCreateScheduleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateScheduleRequest) => createSchedule(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useUpdateScheduleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateScheduleRequest }) =>
      updateSchedule(id, payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
    },
  })
}

export function useDeleteScheduleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteSchedule(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useRunScheduleMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => runSchedule(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.all })
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
      void queryClient.invalidateQueries({ queryKey: ['schedule-runs', 'list'] })
      void queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.stats })
    },
  })
}

export function useScheduleRunQuery(runId: Ref<string | null>) {
  return useQuery({
    queryKey: computed(() => schedulesQueryKeys.run(runId.value ?? '')),
    queryFn: () => fetchScheduleRun(runId.value!),
    enabled: computed(() => !!runId.value),
    refetchInterval: (query) => {
      const status = query.state.data?.run.status
      if (status === 'pending' || status === 'running') return 2000
      return false
    },
  })
}

export function useStartSchedulerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: startScheduler,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
    },
  })
}

export function useStopSchedulerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: stopScheduler,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
    },
  })
}

export function useReloadSchedulerMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: reloadScheduler,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.status })
      void queryClient.invalidateQueries({ queryKey: schedulesQueryKeys.all })
    },
  })
}
