import { apiRequest } from '@/api/client'
import type { RunUpgradeResponse, UpgradesStatusResponse } from '@/types/upgrade'

export function fetchRequiredUpgrades(): Promise<UpgradesStatusResponse> {
  return apiRequest<UpgradesStatusResponse>('/upgrades')
}

export function runUpgrade(upgradeCode: string): Promise<RunUpgradeResponse> {
  return apiRequest<RunUpgradeResponse>('/upgrades', {
    method: 'POST',
    body: JSON.stringify({ upgrade_code: upgradeCode }),
  })
}
