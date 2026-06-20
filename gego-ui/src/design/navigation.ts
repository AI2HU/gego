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
    items: [
      { to: '/', label: 'Dashboard', icon: 'chart-bar', permission: 'dashboard' },
      { to: '/search', label: 'Search', icon: 'search', permission: 'search' },
    ],
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

export const pageMeta: Record<string, { title: string; description: string }> = {
  dashboard: {
    title: 'Dashboard',
    description:
      'Track keyword mentions, provider distribution, and top cited domains across your responses.',
  },
  search: {
    title: 'Search',
    description:
      'Search for keywords across all LLM responses and review each matching response in full.',
  },
  models: {
    title: 'Models',
    description:
      'Manage LLM providers and models used for prompt execution. Add models from OpenAI, Anthropic, Google, Perplexity, or Ollama.',
  },
  prompts: {
    title: 'Prompts',
    description:
      'Scan, filter, and organize prompt templates. Search by content or tag and edit tags inline.',
  },
  scheduler: {
    title: 'Scheduler',
    description:
      'Manage execution schedules and control the background scheduler that runs prompts on a cron.',
  },
  forbidden: {
    title: 'Access denied',
    description: 'You do not have permission to view this page.',
  },
}

export const appMeta = {
  title: 'Gego',
  subtitle: 'See what AI says about your brand',
} as const
