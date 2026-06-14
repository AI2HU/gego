/// <reference types="vite/client" />

import type { RoutePermission } from '@/auth/permissions'

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    guestOnly?: boolean
    permissions?: RoutePermission[]
  }
}

export {}
