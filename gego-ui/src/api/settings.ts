import { apiRequest } from '@/api/client'
import type { SMTPSettings, TestSMTPRequest, UpdateSMTPSettingsRequest } from '@/types/settings'

export function fetchSMTPSettings(): Promise<SMTPSettings> {
  return apiRequest<SMTPSettings>('/settings/smtp')
}

export function updateSMTPSettings(payload: UpdateSMTPSettingsRequest): Promise<SMTPSettings> {
  return apiRequest<SMTPSettings>('/settings/smtp', {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function testSMTPSettings(payload: TestSMTPRequest): Promise<{ ok: boolean }> {
  return apiRequest<{ ok: boolean }>('/settings/smtp/test', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
