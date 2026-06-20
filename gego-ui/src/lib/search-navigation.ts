import type { LocationQuery } from 'vue-router'

export function searchRouteFor(term: string) {
  return {
    name: 'search' as const,
    query: { q: term },
  }
}

export function searchQueryFromRoute(query: LocationQuery): string | null {
  const value = query.q
  if (typeof value !== 'string') {
    return null
  }
  const trimmed = value.trim()
  return trimmed.length >= 2 ? trimmed : null
}
