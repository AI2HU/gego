import { apiRequest } from '@/api/client'
import type {
  CreateScheduleRequest,
  PaginatedSchedulesResponse,
  ScheduleResponse,
  UpdateScheduleRequest,
} from '@/types/schedule'

function fetchSchedulesPage(page: number, limit: number): Promise<PaginatedSchedulesResponse> {
  return apiRequest<PaginatedSchedulesResponse>(`/schedules?page=${page}&limit=${limit}`)
}

export async function fetchSchedules(): Promise<ScheduleResponse[]> {
  const all: ScheduleResponse[] = []
  let page = 1
  let totalPages = 1

  while (page <= totalPages) {
    const result = await fetchSchedulesPage(page, 100)
    all.push(...result.data)
    totalPages = result.pagination.total_pages
    page++
  }

  return all
}

export function createSchedule(payload: CreateScheduleRequest): Promise<ScheduleResponse> {
  return apiRequest<ScheduleResponse>('/schedules', {
    method: 'POST',
    body: JSON.stringify({ enabled: true, temperature: 0.7, ...payload }),
  })
}

export function updateSchedule(id: string, payload: UpdateScheduleRequest): Promise<ScheduleResponse> {
  return apiRequest<ScheduleResponse>(`/schedules/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteSchedule(id: string): Promise<void> {
  return apiRequest<void>(`/schedules/${id}`, { method: 'DELETE' })
}

export function runSchedule(id: string): Promise<void> {
  return apiRequest<void>(`/schedules/${id}/run`, { method: 'POST' })
}
