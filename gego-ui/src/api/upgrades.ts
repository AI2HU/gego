import { apiRequest } from '@/api/client'
import type { RunUpgradeResponse, UpgradeItem } from '@/types/upgrade'

export function fetchRequiredUpgrades(): Promise<UpgradeItem[]> {
  return apiRequest<UpgradeItem[]>('/upgrades', { skipAuthRetry: true })
}

export function runUpgrade(upgradeCode: string): Promise<RunUpgradeResponse> {
  return apiRequest<RunUpgradeResponse>('/upgrades', {
    method: 'POST',
    body: JSON.stringify({ upgrade_code: upgradeCode }),
    skipAuthRetry: true,
  })
}
