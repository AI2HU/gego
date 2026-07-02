import { UPGRADE_SQLITE_TO_POSTGRES, type UpgradeDoc } from '@/types/upgrade'

export const UPGRADE_DOCS: Record<string, UpgradeDoc> = {
  [UPGRADE_SQLITE_TO_POSTGRES]: {
    title: 'Migrate to PostgreSQL',
    summary:
      'This release replaces SQLite with PostgreSQL for LLMs, schedules, users, and sessions. Add a PostgreSQL database before running the upgrade.',
    steps: [
      'Provision a PostgreSQL instance reachable from the API and worker containers.',
      'Set GEGO_POSTGRES_URI on both the API and worker (for example postgres://user:pass@postgres:5432/gego?sslmode=disable).',
      'Ensure GEGO_MONGODB_URI still points to your existing MongoDB instance.',
      'Restart the API container so it can reach PostgreSQL.',
      'Run the upgrade below to copy SQLite data into PostgreSQL and update configuration.',
      'Restart the API and worker after the upgrade completes.',
    ],
  },
}

export function getUpgradeDoc(code: string): UpgradeDoc {
  return (
    UPGRADE_DOCS[code] ?? {
      title: code,
      summary: 'An upgrade is required before Gego can continue.',
      steps: ['Contact your administrator to complete this upgrade.'],
    }
  )
}
