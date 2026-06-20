export function decodeJwtExp(token: string): number | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return null
    }

    const payload = JSON.parse(atob(parts[1]!.replace(/-/g, '+').replace(/_/g, '/'))) as {
      exp?: number
    }

    if (typeof payload.exp !== 'number') {
      return null
    }

    return payload.exp * 1000
  } catch {
    return null
  }
}
