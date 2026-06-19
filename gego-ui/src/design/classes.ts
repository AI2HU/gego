export const page = {
  root: 'min-h-screen bg-gradient-to-br from-gray-50 via-slate-50 to-gray-100',
  shell: 'flex min-h-screen',
  content: 'flex flex-1 flex-col min-w-0 w-full lg:ml-64',
  main: 'flex-1 px-4 sm:px-6 lg:px-8 py-4 md:py-8',
} as const

export const header = {
  bar: 'bg-white/60 backdrop-blur-md shadow-sm border-b border-gray-200/50 sticky top-0 z-30',
  inner: 'px-4 sm:px-6 lg:px-8',
  row: 'flex justify-between items-center py-3 md:py-4',
  logoBox: 'w-8 h-8 md:w-10 md:h-10 bg-gradient-to-br from-slate-400 to-slate-500 rounded-lg flex items-center justify-center shadow-sm',
  title: 'text-lg md:text-xl font-semibold text-gray-800',
  subtitle: 'hidden sm:block text-sm text-gray-500',
} as const

export const sidebar = {
  panel:
    'fixed top-0 left-0 z-50 flex h-full w-64 flex-col border-r border-gray-200/50 bg-white/95 backdrop-blur-md shadow-lg transition-transform duration-300 ease-in-out lg:translate-x-0',
  panelOpen: 'translate-x-0',
  panelClosed: '-translate-x-full',
  backdrop: 'fixed inset-0 z-40 bg-gray-900/40 backdrop-blur-sm lg:hidden',
  brand: 'flex items-center gap-3 border-b border-gray-200/50 px-5 py-5',
  brandLogo:
    'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-slate-400 to-slate-500 shadow-sm',
  brandTitle: 'text-lg font-semibold text-gray-800 leading-tight',
  brandSubtitle: 'text-xs text-gray-500 leading-snug',
  nav: 'flex-1 overflow-y-auto custom-scrollbar px-3 py-4',
  navList: 'flex flex-col gap-1',
  sectionLabel:
    'px-3 pt-4 pb-2 text-[11px] font-semibold uppercase tracking-wider text-gray-400 first:pt-0',
  footer: 'border-t border-gray-200/50 px-4 py-4 space-y-3',
  footerActions: 'flex items-center justify-between gap-2',
} as const

export const nav = {
  link: 'px-4 py-2 rounded-lg transition-colors duration-200 font-medium text-sm',
  active: 'bg-slate-600 text-white hover:bg-slate-700',
  inactive: 'text-gray-700 hover:text-gray-900 hover:bg-gray-100',
  mobilePanel: 'md:hidden border-t border-gray-200/50 pt-4 pb-4 space-y-3',
  sidebarLink:
    'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors duration-200',
  sidebarActive: 'bg-slate-600 text-white shadow-sm',
  sidebarInactive: 'text-gray-700 hover:bg-gray-100 hover:text-gray-900',
} as const

export const card = {
  base: 'bg-white/80 backdrop-blur-sm rounded-lg shadow-sm border border-gray-200/50',
  interactive: 'bg-white/80 backdrop-blur-sm rounded-lg shadow-sm hover:shadow-md transition-all duration-200 border border-gray-200/50',
  header: 'px-6 py-4 border-b border-gray-200/50',
  body: 'p-6',
  inset: 'bg-slate-50 rounded-lg p-4 border border-gray-200/50',
  insetItem: 'p-3 bg-white rounded border border-gray-200/50',
  sectionTitle: 'text-lg font-semibold text-gray-800',
  sectionSubtitle: 'text-sm text-gray-600',
} as const

export const button = {
  base: 'inline-flex items-center justify-center space-x-2 rounded-lg font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed',
  sm: 'px-3 py-1.5 text-xs',
  md: 'px-4 py-2 text-sm',
  lg: 'px-6 py-2 md:py-3 text-sm md:text-base',
  primary: 'bg-slate-600 text-white hover:bg-slate-700 shadow-sm hover:shadow-md',
  secondary: 'bg-gray-100 text-gray-700 hover:bg-gray-200',
  danger: 'bg-red-600 text-white hover:bg-red-700',
  ghost: 'text-gray-700 hover:bg-gray-100',
  chip: 'px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-gray-700 hover:text-gray-900 border border-transparent hover:border-slate-300 shadow-sm hover:shadow',
} as const

export const input = {
  base: 'w-full px-4 py-2 md:py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-slate-500 focus:border-transparent bg-white text-gray-900 placeholder-gray-500 shadow-sm text-sm md:text-base',
} as const

export const status = {
  connected: 'flex items-center px-3 py-1.5 bg-green-50 rounded-lg border border-green-200/50',
  connectedDot: 'w-2 h-2 bg-green-400 rounded-full mr-2',
  connectedText: 'text-sm font-medium text-green-700',
  disconnected: 'flex items-center px-3 py-1.5 bg-red-50 rounded-lg border border-red-200/50',
  disconnectedDot: 'w-2 h-2 bg-red-400 rounded-full mr-2',
  disconnectedText: 'text-sm font-medium text-red-700',
} as const

export const iconBox = {
  sm: 'w-8 h-8 bg-slate-100 rounded-lg flex items-center justify-center',
  md: 'w-10 h-10 bg-slate-200 rounded-lg flex items-center justify-center',
  lg: 'w-12 h-12 bg-slate-100 rounded-lg flex items-center justify-center',
} as const

export const typography = {
  label: 'text-sm font-medium text-gray-600',
  value: 'text-2xl font-semibold text-gray-800',
  hint: 'text-xs text-gray-500',
  overline: 'text-xs font-semibold text-gray-600 uppercase tracking-wide',
} as const
