import { apiRequest } from '@/api/client'
import type {
  CreateModelRequest,
  ListProviderModelsRequest,
  ModelInfo,
  ModelResponse,
  ProviderApiKey,
  ProviderInfo,
  TestModelAccessResponse,
  UpdateModelRequest,
} from '@/types/model'

export function fetchModels(): Promise<ModelResponse[]> {
  return apiRequest<ModelResponse[]>('/models')
}

export function fetchProviders(): Promise<ProviderInfo[]> {
  return apiRequest<ProviderInfo[]>('/providers')
}

export function fetchProviderApiKeys(provider: string): Promise<ProviderApiKey[]> {
  return apiRequest<ProviderApiKey[]>(`/providers/${provider}/api-keys`)
}

export function fetchProviderModels(
  provider: string,
  payload: ListProviderModelsRequest,
): Promise<ModelInfo[]> {
  return apiRequest<ModelInfo[]>(`/providers/${provider}/models`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function createModel(payload: CreateModelRequest): Promise<ModelResponse> {
  return apiRequest<ModelResponse>('/models', {
    method: 'POST',
    body: JSON.stringify({ enabled: true, ...payload }),
  })
}

export function deleteModel(id: string): Promise<void> {
  return apiRequest<void>(`/models/${id}`, { method: 'DELETE' })
}

export function updateModel(id: string, payload: UpdateModelRequest): Promise<ModelResponse> {
  return apiRequest<ModelResponse>(`/models/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function testModelAccess(id: string): Promise<TestModelAccessResponse> {
  return apiRequest<TestModelAccessResponse>(`/models/${id}/test`, { method: 'POST' })
}
