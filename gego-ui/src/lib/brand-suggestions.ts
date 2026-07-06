import type { Brand, SuggestedBrandWord } from '@/types/brand'

export function filterBrandSuggestions(
  brands: Brand[],
  query: string,
  limit = 8,
): Brand[] {
  const normalized = query.trim().toLowerCase()

  if (!normalized) {
    return brands.slice(0, limit)
  }

  return brands
    .filter((brand) => {
      if (brand.name.toLowerCase().includes(normalized)) {
        return true
      }
      return brand.aliases.some((alias) => alias.alias.toLowerCase().includes(normalized))
    })
    .slice(0, limit)
}

export function filterDetectedBrandWordSuggestions(
  words: SuggestedBrandWord[],
  query: string,
  limit = 8,
): SuggestedBrandWord[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) {
    return words.slice(0, limit)
  }

  return words
    .filter((item) => item.word.toLowerCase().includes(normalized))
    .slice(0, limit)
}
