import type { AppIconName } from '@/components/icons/AppIcon.vue'

export interface NavItem {
  to: string
  label: string
  icon?: AppIconName
}

export const mainNavItems: NavItem[] = [
  { to: '/', label: 'Dashboard', icon: 'chart-bar' },
]

export const appMeta = {
  title: 'Gego',
  subtitle: 'Generative Engine Optimization Platform',
} as const
