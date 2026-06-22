import { apiRequest } from '@/api/client'
import type {
  ScheduleRunDetailResponse,
  ScheduleRunListResponse,
} from '@/types/schedule'

export function fetchScheduleRuns(scheduleId?: string, cursor?: string, limit = 20): Promise<ScheduleRunListResponse> {
  const params = new URLSearchParams({ limit: String(limit) })
  if (scheduleId) params.set('schedule_id', scheduleId)
  if (cursor) params.set('cursor', cursor)
  return apiRequest<ScheduleRunListResponse>(`/schedule-runs?${params}`)
}

export function fetchScheduleRun(id: string): Promise<ScheduleRunDetailResponse> {
  return apiRequest<ScheduleRunDetailResponse>(`/schedule-runs/${id}`)
}

export function cancelScheduleRun(id: string): Promise<void> {
  return apiRequest<void>(`/schedule-runs/${id}/cancel`, { method: 'POST' })
}

export function retryScheduleJob(runId: string, jobId: string): Promise<void> {
  return apiRequest<void>(`/schedule-runs/${runId}/jobs/${jobId}/retry`, { method: 'POST' })
}
