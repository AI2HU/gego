import type { MaybeRefOrGetter, Ref } from 'vue'
import { computed, ref, watch } from 'vue'

import {
  buildBrandCitationTargets,
  isBrandCitationTargetInList,
  type BrandCitationTarget,
} from '@/lib/brand-citation-target'
import { useBrandCitationDomainsQuery } from '@/queries/dashboard'
import { useBrandsQuery } from '@/queries/brands'
import type { StatsResponse } from '@/types/stats'

export function useBrandCitationTargets(
  stats: Ref<StatsResponse | undefined>,
  tags: MaybeRefOrGetter<string[]> = () => [],
) {
  const brandsQuery = useBrandsQuery()
  const selectedTarget = ref<BrandCitationTarget | null>(null)

  const targets = computed(() =>
    buildBrandCitationTargets(brandsQuery.data.value ?? [], stats.value?.top_keywords ?? []),
  )

  const domainsQuery = useBrandCitationDomainsQuery(selectedTarget, tags)

  watch(
    targets,
    (items) => {
      if (items.length === 0) {
        selectedTarget.value = null
        return
      }
      if (isBrandCitationTargetInList(selectedTarget.value, items)) {
        return
      }
      selectedTarget.value = items[0] ?? null
    },
    { immediate: true },
  )

  return {
    targets,
    selectedTarget,
    domainsQuery,
  }
}
