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

export const CRON_PRESET_LABELS: Record<Exclude<CronPreset, 'custom'>, string> = {
  daily: 'Every day',
  weekly: 'Every week',
  monthly: 'Every month',
}

export const CRON_PRESET_HINTS: Record<Exclude<CronPreset, 'custom'>, string> = {
  daily: 'Once a day at 9:00 AM',
  weekly: 'Every Monday at 9:00 AM',
  monthly: 'On the 1st of each month at 9:00 AM',
}

const CRON_EXPR_TO_PRESET = Object.fromEntries(
  Object.entries(CRON_PRESETS).map(([preset, expr]) => [expr, preset]),
) as Record<string, Exclude<CronPreset, 'custom'>>

export function getCronLabel(cronExpr: string): string {
  const preset = CRON_EXPR_TO_PRESET[cronExpr.trim()]
  return preset ? CRON_PRESET_LABELS[preset] : 'Custom'
}

export function getCronHint(cronExpr: string): string | undefined {
  const preset = CRON_EXPR_TO_PRESET[cronExpr.trim()]
  return preset ? CRON_PRESET_HINTS[preset] : cronExpr.trim() || undefined
}
