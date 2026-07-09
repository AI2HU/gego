import type { Brand } from '@/types/brand'
import type { KeywordCount } from '@/types/stats'

export type BrandCitationTarget =
  | { kind: 'brand'; brandId: string; label: string; aliases: string[] }
  | { kind: 'keyword'; keyword: string; label: string; count: number }

export function brandCitationTargetKey(target: BrandCitationTarget): string {
  return target.kind === 'brand' ? `brand:${target.brandId}` : `keyword:${target.keyword}`
}

function isKeywordCoveredByBrand(keyword: string, brands: Brand[]): boolean {
  const normalized = keyword.toLowerCase()
  for (const brand of brands) {
    if (brand.name.toLowerCase() === normalized) {
      return true
    }
    if (brand.aliases.some((alias) => alias.alias.toLowerCase() === normalized)) {
      return true
    }
  }
  return false
}

export function buildBrandCitationTargets(
  brands: Brand[],
  keywords: KeywordCount[],
): BrandCitationTarget[] {
  const targets: BrandCitationTarget[] = brands.map((brand) => ({
    kind: 'brand',
    brandId: brand.id,
    label: brand.name,
    aliases: brand.aliases.map((alias) => alias.alias),
  }))

  for (const item of keywords) {
    if (isKeywordCoveredByBrand(item.keyword, brands)) {
      continue
    }
    targets.push({
      kind: 'keyword',
      keyword: item.keyword,
      label: item.keyword,
      count: item.count,
    })
  }

  return targets
}

export function filterBrandCitationTargets(
  targets: BrandCitationTarget[],
  query: string,
  limit = 12,
): BrandCitationTarget[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) {
    return targets.slice(0, limit)
  }

  return targets
    .filter((target) => {
      if (target.label.toLowerCase().includes(normalized)) {
        return true
      }
      if (target.kind === 'brand') {
        return target.aliases.some((alias) => alias.toLowerCase().includes(normalized))
      }
      return target.keyword.toLowerCase().includes(normalized)
    })
    .slice(0, limit)
}

export function isBrandCitationTargetInList(
  target: BrandCitationTarget | null,
  targets: BrandCitationTarget[],
): boolean {
  if (!target) {
    return false
  }
  const key = brandCitationTargetKey(target)
  return targets.some((item) => brandCitationTargetKey(item) === key)
}
