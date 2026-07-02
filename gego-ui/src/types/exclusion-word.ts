export interface ExclusionWord {
  id: string
  word: string
  created_at: string
  updated_at: string
}

export interface CreateExclusionWordRequest {
  word: string
}

export interface SuggestedBrandWord {
  word: string
  count: number
}
