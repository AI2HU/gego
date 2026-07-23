export interface BrandAlias {
  id: string
  brand_id: string
  alias: string
  case_sensitive: boolean
  created_at: string
  updated_at: string
}

export interface Brand {
  id: string
  name: string
  is_target: boolean
  aliases: BrandAlias[]
  created_at: string
  updated_at: string
}

export interface CreateBrandRequest {
  name: string
  aliases?: Array<{
    alias: string
    case_sensitive?: boolean
  }>
}

export interface UpdateBrandRequest {
  name: string
  is_target?: boolean
}

export interface CreateBrandAliasRequest {
  alias: string
  case_sensitive?: boolean
}

export interface UpdateBrandAliasRequest {
  alias: string
  case_sensitive?: boolean
}

export interface MapBrandRequest {
  alias: string
  name: string
  case_sensitive?: boolean
}

export interface SuggestedBrandWord {
  word: string
  count: number
}
