export interface ScheduleResponse {
  id: string
  name: string
  prompt_ids: string[]
  llm_ids: string[]
  cron_expr: string
  temperature: number
  enabled: boolean
  last_run?: string
  next_run?: string
  created_at: string
  updated_at: string
}

export interface CreateScheduleRequest {
  name: string
  prompt_ids: string[]
  llm_ids: string[]
  cron_expr: string
  temperature?: number
  enabled?: boolean
}

export interface UpdateScheduleRequest {
  name?: string
  prompt_ids?: string[]
  llm_ids?: string[]
  cron_expr?: string
  temperature?: number
  enabled?: boolean
}

export interface PaginatedSchedulesResponse {
  data: ScheduleResponse[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export interface SchedulerStatusResponse {
  running: boolean
  enabled_schedules: number
}

export type CronPreset = 'daily' | 'weekly' | 'monthly' | 'custom'

export const CRON_PRESETS: Record<Exclude<CronPreset, 'custom'>, string> = {
  daily: '0 9 * * *',
  weekly: '0 9 * * MON',
  monthly: '0 9 1 * *',
}
