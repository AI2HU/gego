export interface NavItem {
  to: string
  label: string
}

export const mainNavItems: NavItem[] = [
  { to: '/', label: 'Dashboard' },
]

export const appMeta = {
  title: 'Gego',
  subtitle: 'Generative Engine Optimization Platform',
} as const
