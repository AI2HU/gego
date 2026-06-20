import { apiRequest } from '@/api/client'
import type { SchedulerStatusResponse } from '@/types/schedule'

export function fetchSchedulerStatus(): Promise<SchedulerStatusResponse> {
  return apiRequest<SchedulerStatusResponse>('/scheduler/status')
}

export function startScheduler(): Promise<SchedulerStatusResponse> {
  return apiRequest<SchedulerStatusResponse>('/scheduler/start', { method: 'POST' })
}

export function stopScheduler(): Promise<SchedulerStatusResponse> {
  return apiRequest<SchedulerStatusResponse>('/scheduler/stop', { method: 'POST' })
}

export function reloadScheduler(): Promise<SchedulerStatusResponse> {
  return apiRequest<SchedulerStatusResponse>('/scheduler/reload', { method: 'POST' })
}
