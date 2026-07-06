import { apiRequest } from '@/api/client'
import type {
  Brand,
  CreateBrandAliasRequest,
  CreateBrandRequest,
  MapBrandRequest,
  SuggestedBrandWord,
  UpdateBrandAliasRequest,
  UpdateBrandRequest,
} from '@/types/brand'

export function fetchBrands(): Promise<Brand[]> {
  return apiRequest<Brand[]>('/brands')
}

export function createBrand(payload: CreateBrandRequest): Promise<Brand> {
  return apiRequest<Brand>('/brands', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateBrand(id: string, payload: UpdateBrandRequest): Promise<Brand> {
  return apiRequest<Brand>(`/brands/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteBrand(id: string): Promise<void> {
  return apiRequest<void>(`/brands/${id}`, { method: 'DELETE' })
}

export function createBrandAlias(brandId: string, payload: CreateBrandAliasRequest): Promise<Brand['aliases'][number]> {
  return apiRequest<Brand['aliases'][number]>(`/brands/${brandId}/aliases`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function updateBrandAlias(
  brandId: string,
  aliasId: string,
  payload: UpdateBrandAliasRequest,
): Promise<Brand['aliases'][number]> {
  return apiRequest<Brand['aliases'][number]>(`/brands/${brandId}/aliases/${aliasId}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function deleteBrandAlias(brandId: string, aliasId: string): Promise<void> {
  return apiRequest<void>(`/brands/${brandId}/aliases/${aliasId}`, { method: 'DELETE' })
}

export function mapBrandFromDetection(payload: MapBrandRequest): Promise<Brand> {
  return apiRequest<Brand>('/brands/map', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function fetchSuggestedBrandWords(limit = 50, tags: string[] = []): Promise<SuggestedBrandWord[]> {
  const params = new URLSearchParams({ limit: String(limit) })
  for (const tag of tags) {
    params.append('tags', tag)
  }
  return apiRequest<SuggestedBrandWord[]>(`/brands/suggestions?${params.toString()}`)
}
