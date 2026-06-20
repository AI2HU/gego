import { apiRequest } from '@/api/client'
import type {
  BulkCreatePromptsRequest,
  BulkCreatePromptsResponse,
  CreatePromptRequest,
  GeneratePromptsRequest,
  GeneratePromptsResponse,
  PaginatedPromptsResponse,
  PromptResponse,
  UpdatePromptRequest,
} from '@/types/prompt'

function fetchPromptsPage(page: number, limit: number): Promise<PaginatedPromptsResponse> {
  return apiRequest<PaginatedPromptsResponse>(`/prompts?page=${page}&limit=${limit}`)
}

export async function fetchPrompts(): Promise<PromptResponse[]> {
  const all: PromptResponse[] = []
  let page = 1
  let totalPages = 1

  while (page <= totalPages) {
    const result = await fetchPromptsPage(page, 100)
    all.push(...result.data)
    totalPages = result.pagination.total_pages
    page++
  }

  return all
}

export function createPrompt(payload: CreatePromptRequest): Promise<PromptResponse> {
  return apiRequest<PromptResponse>('/prompts', {
    method: 'POST',
    body: JSON.stringify({ enabled: true, ...payload }),
  })
}

export function createPrompts(payload: BulkCreatePromptsRequest): Promise<BulkCreatePromptsResponse> {
  return apiRequest<BulkCreatePromptsResponse>('/prompts', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updatePrompt(id: string, payload: UpdatePromptRequest): Promise<PromptResponse> {
  return apiRequest<PromptResponse>(`/prompts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function generatePrompts(payload: GeneratePromptsRequest): Promise<GeneratePromptsResponse> {
  return apiRequest<GeneratePromptsResponse>('/prompts/generate', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function deletePrompt(id: string): Promise<void> {
  return apiRequest<void>(`/prompts/${id}`, { method: 'DELETE' })
}
