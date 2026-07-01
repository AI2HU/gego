export const UPGRADE_SQLITE_TO_POSTGRES = 'sqlite_to_postgres'

export interface UpgradesStatusResponse {
  required_upgrade_codes: string[]
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
