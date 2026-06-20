import type { Role } from '@/types/auth'

export type RoutePermission = 'dashboard' | 'models' | 'prompts' | 'scheduler'

const roleRouteAccess: Record<Role, RoutePermission[]> = {
  admin: ['dashboard', 'models', 'prompts', 'scheduler'],
  member: ['dashboard'],
}

export function canAccessRoute(role: Role, permission: RoutePermission): boolean {
  return roleRouteAccess[role]?.includes(permission) ?? false
}

export function canAccessRoutes(role: Role, permissions: RoutePermission[]): boolean {
  return permissions.every((permission) => canAccessRoute(role, permission))
}
