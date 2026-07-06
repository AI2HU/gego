import { splitHighlightedTextMulti, type HighlightSegment } from '@/lib/search-matches'

export type { HighlightSegment as TextSegment }

type HighlightContext = {
  terms: string[]
  searchKeyword: string
  caseSensitive: boolean
}

export type InlinePart =
  | { type: 'text'; segments: TextSegment[] }
  | { type: 'bold'; parts: InlinePart[] }
  | { type: 'italic'; parts: InlinePart[] }
  | { type: 'link'; url: string; parts: InlinePart[] }

export type MarkdownBlock =
  | { type: 'heading'; level: number; parts: InlinePart[] }
  | { type: 'list-item'; parts: InlinePart[] }
  | { type: 'paragraph'; parts: InlinePart[] }
  | { type: 'spacer' }

const INLINE_TOKEN_REGEX =
  /(\[([^\]]+)\]\(([^)]+)\)|\*\*(.+?)\*\*|__(.+?)__|\*([^*\n]+?)\*|_([^_\n]+?)_)/g

function isSafeUrl(url: string): boolean {
  try {
    const parsed = new URL(url)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function normalizeHighlightContext(
  highlightTerms: string | string[],
  caseSensitive: boolean,
  searchKeyword?: string,
): HighlightContext {
  const terms = Array.isArray(highlightTerms)
    ? highlightTerms
    : highlightTerms
      ? [highlightTerms]
      : []

  return {
    terms,
    searchKeyword: searchKeyword?.trim() ?? '',
    caseSensitive,
  }
}

function toTextPart(text: string, context: HighlightContext): InlinePart {
  const segments =
    context.terms.length > 0
      ? splitHighlightedTextMulti(
          text,
          context.terms,
          context.caseSensitive,
          context.searchKeyword,
        )
      : [{ text }]

  return { type: 'text', segments }
}

function appendTextPart(parts: InlinePart[], text: string, context: HighlightContext): void {
  if (!text) {
    return
  }
  parts.push(toTextPart(text, context))
}

export function parseInline(
  text: string,
  highlightTerms: string | string[],
  caseSensitive = false,
  searchKeyword = '',
): InlinePart[] {
  const context = normalizeHighlightContext(highlightTerms, caseSensitive, searchKeyword)
  const parts: InlinePart[] = []
  let lastIndex = 0
  let foundToken = false

  for (const match of text.matchAll(INLINE_TOKEN_REGEX)) {
    foundToken = true
    const start = match.index ?? 0

    if (start > lastIndex) {
      appendTextPart(parts, text.slice(lastIndex, start), context)
    }

    if (match[2] !== undefined && match[3] !== undefined) {
      const label = match[2]
      const url = match[3].trim()
      if (isSafeUrl(url)) {
        parts.push({
          type: 'link',
          url,
          parts: parseInline(label, context.terms, context.caseSensitive, context.searchKeyword),
        })
      } else {
        appendTextPart(parts, match[0], context)
      }
    } else if (match[4] !== undefined || match[5] !== undefined) {
      const inner = match[4] ?? match[5] ?? ''
      parts.push({
        type: 'bold',
        parts: parseInline(inner, context.terms, context.caseSensitive, context.searchKeyword),
      })
    } else if (match[6] !== undefined || match[7] !== undefined) {
      const inner = match[6] ?? match[7] ?? ''
      parts.push({
        type: 'italic',
        parts: parseInline(inner, context.terms, context.caseSensitive, context.searchKeyword),
      })
    }

    lastIndex = start + match[0].length
  }

  if (!foundToken) {
    appendTextPart(parts, text, context)
    return parts
  }

  if (lastIndex < text.length) {
    appendTextPart(parts, text.slice(lastIndex), context)
  }

  return parts
}

export function parseMarkdownBlocks(
  text: string,
  highlightTerms: string | string[],
  caseSensitive = false,
  searchKeyword = '',
): MarkdownBlock[] {
  const lines = text.split('\n')
  const blocks: MarkdownBlock[] = []

  for (const line of lines) {
    if (line.trim() === '') {
      blocks.push({ type: 'spacer' })
      continue
    }

    const headingMatch = line.match(/^(#{1,6})\s+(.+)$/)
    if (headingMatch?.[1] && headingMatch[2]) {
      blocks.push({
        type: 'heading',
        level: headingMatch[1].length,
        parts: parseInline(headingMatch[2], highlightTerms, caseSensitive, searchKeyword),
      })
      continue
    }

    const listMatch = line.match(/^(?:[-*]|\d+\.)\s+(.+)$/)
    if (listMatch?.[1]) {
      blocks.push({
        type: 'list-item',
        parts: parseInline(listMatch[1], highlightTerms, caseSensitive, searchKeyword),
      })
      continue
    }

    blocks.push({
      type: 'paragraph',
      parts: parseInline(line, highlightTerms, caseSensitive, searchKeyword),
    })
  }

  return blocks
}
