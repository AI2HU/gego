import type { PromptResponse } from '@/types/prompt'
import type { SearchMatch, SearchResponseItem, SearchURL } from '@/types/search'

const CITATION_MARKER_REGEX = /\[(\d+)\]/g

function sortSearchUrls(urls: SearchURL[]): SearchURL[] {
  return [...urls].sort((a, b) => a.citation_index - b.citation_index)
}

export function linkCitationMarkers(text: string, sources: SearchURL[]): string {
  if (sources.length === 0) {
    return text
  }

  const byIndex = new Map<number, SearchURL>()
  for (const source of sources) {
    byIndex.set(source.citation_index, source)
  }

  return text.replace(CITATION_MARKER_REGEX, (match, rawIndex: string) => {
    const index = Number.parseInt(rawIndex, 10)
    if (!Number.isFinite(index) || index < 1) {
      return match
    }

    const source = byIndex.get(index - 1) ?? sources[index - 1]
    if (!source?.url) {
      return match
    }

    const label = source.title?.trim() || `[${index}]`
    return `[${label}](${source.url})`
  })
}

function tagsMatch(selectedTag: string, promptTag: string): boolean {
  return selectedTag.toLowerCase() === promptTag.toLowerCase()
}

export function filterResponsesByTags(
  responses: SearchResponseItem[],
  tags: string[],
  prompts: PromptResponse[],
): SearchResponseItem[] {
  if (tags.length === 0) {
    return responses
  }

  if (prompts.length === 0) {
    return responses
  }

  const allowedPromptIds = new Set(
    prompts
      .filter((prompt) =>
        (prompt.tags ?? []).some((promptTag) =>
          tags.some((selectedTag) => tagsMatch(selectedTag, promptTag)),
        ),
      )
      .map((prompt) => prompt.id),
  )

  if (allowedPromptIds.size === 0) {
    return []
  }

  return responses.filter((response) => allowedPromptIds.has(response.prompt_id))
}

function escapeRegex(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

export function uniqueSearchTerms(...groups: (string | string[] | undefined)[]): string[] {
  const seen = new Set<string>()
  const unique: string[] = []

  for (const group of groups) {
    const items = typeof group === 'string' ? [group] : (group ?? [])
    for (const term of items) {
      const trimmed = term.trim()
      if (!trimmed) continue
      const key = trimmed.toLowerCase()
      if (seen.has(key)) continue
      seen.add(key)
      unique.push(trimmed)
    }
  }

  return unique
}

function normalizeSearchTerms(keyword: string, searchTerms?: string[]): string[] {
  return uniqueSearchTerms(searchTerms?.length ? searchTerms : keyword)
}

function findSearchMatches(
  responses: SearchResponseItem[],
  keyword: string,
  options: {
    caseSensitive?: boolean
    searchTerms?: string[]
    promptNames?: Map<string, string>
    promptTags?: Map<string, string[]>
  } = {},
): SearchMatch[] {
  const {
    caseSensitive = false,
    searchTerms,
    promptNames = new Map(),
    promptTags = new Map(),
  } = options

  const terms = normalizeSearchTerms(keyword, searchTerms)
  const flags = caseSensitive ? 'g' : 'gi'
  const matches: SearchMatch[] = []

  for (const response of responses) {
    const text = response.response_text
    for (const term of terms) {
      const regex = new RegExp(escapeRegex(term), flags)
      for (const match of text.matchAll(regex)) {
        matches.push({
          responseId: response.id,
          promptId: response.prompt_id,
          promptName: promptNames.get(response.prompt_id) ?? 'Unknown Prompt',
          promptTags: promptTags.get(response.prompt_id) ?? [],
          responseText: text,
          searchUrls: sortSearchUrls(response.search_urls ?? []),
          llmName: response.llm_name,
          llmProvider: response.llm_provider,
          temperature: response.temperature ?? 0,
          keyword: term,
          createdAt: response.created_at,
        })
      }
    }
  }

  return matches
}

function buildFallbackSearchMatches(
  responses: SearchResponseItem[],
  keyword: string,
  options: {
    searchTerms?: string[]
    promptNames?: Map<string, string>
    promptTags?: Map<string, string[]>
  } = {},
): SearchMatch[] {
  const {
    searchTerms,
    promptNames = new Map(),
    promptTags = new Map(),
  } = options

  const terms = normalizeSearchTerms(keyword, searchTerms)
  const primaryTerm = terms[0] ?? keyword

  return responses.map((response) => ({
    responseId: response.id,
    promptId: response.prompt_id,
    promptName: promptNames.get(response.prompt_id) ?? 'Unknown Prompt',
    promptTags: promptTags.get(response.prompt_id) ?? [],
    responseText: response.response_text,
    searchUrls: sortSearchUrls(response.search_urls ?? []),
    llmName: response.llm_name,
    llmProvider: response.llm_provider,
    temperature: response.temperature ?? 0,
    keyword: primaryTerm,
    createdAt: response.created_at,
  }))
}

export function resolveSearchMatches(
  responses: SearchResponseItem[],
  keyword: string,
  options: {
    caseSensitive?: boolean
    searchTerms?: string[]
    promptNames?: Map<string, string>
    promptTags?: Map<string, string[]>
  } = {},
): SearchMatch[] {
  const matches = findSearchMatches(responses, keyword, options)
  if (matches.length > 0 || responses.length === 0) {
    return matches
  }

  return buildFallbackSearchMatches(responses, keyword, options)
}

type TextRange = { start: number; end: number }

function mergeRanges(ranges: TextRange[]): TextRange[] {
  if (ranges.length === 0) {
    return []
  }

  const sorted = [...ranges].sort((a, b) => a.start - b.start)
  const merged: TextRange[] = [sorted[0]!]

  for (let i = 1; i < sorted.length; i++) {
    const current = sorted[i]!
    const last = merged[merged.length - 1]!
    if (current.start <= last.end) {
      last.end = Math.max(last.end, current.end)
      continue
    }
    merged.push(current)
  }

  return merged
}

function findHighlightRanges(
  text: string,
  keywords: string[],
  caseSensitive = false,
): TextRange[] {
  const ranges: TextRange[] = []
  const flags = caseSensitive ? 'g' : 'gi'

  for (const keyword of keywords) {
    const regex = new RegExp(escapeRegex(keyword), flags)
    for (const match of text.matchAll(regex)) {
      const start = match.index ?? 0
      ranges.push({ start, end: start + match[0].length })
    }
  }

  return mergeRanges(ranges)
}

export type HighlightKind = 'keyword' | 'alias'

export type HighlightSegment = {
  text: string
  highlight?: HighlightKind
}

function isPrimaryTerm(term: string, primaryKeyword: string, caseSensitive: boolean): boolean {
  if (!primaryKeyword) {
    return true
  }
  return caseSensitive
    ? term === primaryKeyword
    : term.toLowerCase() === primaryKeyword.toLowerCase()
}

function rangesToSegments(
  text: string,
  keywordRanges: TextRange[],
  aliasRanges: TextRange[],
): HighlightSegment[] {
  if (text.length === 0) {
    return [{ text: '' }]
  }

  const kinds: Array<HighlightKind | null> = Array.from({ length: text.length }, () => null)

  for (const range of keywordRanges) {
    for (let index = range.start; index < range.end; index++) {
      kinds[index] = 'keyword'
    }
  }

  for (const range of aliasRanges) {
    for (let index = range.start; index < range.end; index++) {
      if (kinds[index] === null) {
        kinds[index] = 'alias'
      }
    }
  }

  const segments: HighlightSegment[] = []
  let start = 0
  let currentKind = kinds[0]

  for (let index = 1; index <= text.length; index++) {
    const nextKind = index < text.length ? kinds[index] : null
    if (nextKind === currentKind) {
      continue
    }

    segments.push({
      text: text.slice(start, index),
      highlight: currentKind ?? undefined,
    })
    start = index
    currentKind = nextKind
  }

  return segments.length > 0 ? segments : [{ text }]
}

export function splitHighlightedTextMulti(
  text: string,
  keywords: string[],
  caseSensitive = false,
  primaryKeyword = '',
): HighlightSegment[] {
  const terms = normalizeSearchTerms(primaryKeyword, keywords)
  if (terms.length === 0) {
    return [{ text }]
  }

  const primary = primaryKeyword.trim()
  const keywordTerms = terms.filter((term) => isPrimaryTerm(term, primary, caseSensitive))
  const aliasTerms = primary
    ? terms.filter((term) => !isPrimaryTerm(term, primary, caseSensitive))
    : []

  const keywordRanges = findHighlightRanges(
    text,
    keywordTerms.length > 0 ? keywordTerms : terms,
    caseSensitive,
  )
  const aliasRanges = findHighlightRanges(text, aliasTerms, caseSensitive)

  return rangesToSegments(text, keywordRanges, aliasRanges)
}
