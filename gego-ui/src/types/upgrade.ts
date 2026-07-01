export const UPGRADE_SQLITE_TO_POSTGRES = 'sqlite_to_postgres'

export type UpgradeSeverity = 'major' | 'minor'

export interface UpgradeItem {
  code: string
  severity: UpgradeSeverity
}

export interface RunUpgradeRequest {
  upgrade_code: string
}

export interface RunUpgradeResponse {
  upgrade_code: string
  status: string
  message: string
  restart_required: boolean
}

export interface UpgradeDoc {
  title: string
  summary: string
  steps: string[]
}
