import type { PromptResponse } from '@/types/prompt'
import type { SearchMatch, SearchResponseItem } from '@/types/search'

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

export function findSearchMatches(
  responses: SearchResponseItem[],
  keyword: string,
  options: {
    caseSensitive?: boolean
    promptNames?: Map<string, string>
    promptTags?: Map<string, string[]>
  } = {},
): SearchMatch[] {
  const { caseSensitive = false, promptNames = new Map(), promptTags = new Map() } = options
  const flags = caseSensitive ? 'g' : 'gi'
  const regex = new RegExp(escapeRegex(keyword), flags)
  const matches: SearchMatch[] = []

  for (const response of responses) {
    const text = response.response_text
    for (const match of text.matchAll(regex)) {
      matches.push({
        responseId: response.id,
        promptId: response.prompt_id,
        promptName: promptNames.get(response.prompt_id) ?? 'Unknown Prompt',
        promptTags: promptTags.get(response.prompt_id) ?? [],
        responseText: text,
        llmName: response.llm_name,
        llmProvider: response.llm_provider,
        temperature: response.temperature ?? 0,
        keyword,
        createdAt: response.created_at,
      })
    }
  }

  return matches
}

export function splitHighlightedText(
  text: string,
  keyword: string,
  caseSensitive = false,
): Array<{ text: string; highlight: boolean }> {
  if (!keyword) {
    return [{ text, highlight: false }]
  }

  const flags = caseSensitive ? 'g' : 'gi'
  const regex = new RegExp(escapeRegex(keyword), flags)
  const parts: Array<{ text: string; highlight: boolean }> = []
  let lastIndex = 0

  for (const match of text.matchAll(regex)) {
    const start = match.index ?? 0
    if (start > lastIndex) {
      parts.push({ text: text.slice(lastIndex, start), highlight: false })
    }
    parts.push({ text: match[0], highlight: true })
    lastIndex = start + match[0].length
  }

  if (lastIndex < text.length) {
    parts.push({ text: text.slice(lastIndex), highlight: false })
  }

  return parts.length > 0 ? parts : [{ text, highlight: false }]
}
