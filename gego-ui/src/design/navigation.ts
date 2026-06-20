import type { AppIconName } from '@/components/icons/AppIcon.vue'
import type { RoutePermission } from '@/auth/permissions'

export interface NavItem {
  to: string
  label: string
  icon?: AppIconName
  permission?: RoutePermission
}

export interface NavSection {
  label?: string
  items: NavItem[]
  adminOnly?: boolean
}

export const navSections: NavSection[] = [
  {
    items: [{ to: '/', label: 'Dashboard', icon: 'chart-bar', permission: 'dashboard' }],
  },
  {
    label: 'Administration',
    adminOnly: true,
    items: [
      { to: '/admin/models', label: 'Models', icon: 'server', permission: 'models' },
      { to: '/admin/prompts', label: 'Prompts', icon: 'comment', permission: 'prompts' },
      { to: '/admin/scheduler', label: 'Scheduler', icon: 'clock', permission: 'scheduler' },
    ],
  },
]

export const appMeta = {
  title: 'Gego',
  subtitle: 'Generative Engine Optimization Platform',
} as const
