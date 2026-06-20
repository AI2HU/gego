<script setup lang="ts">
import { computed, ref } from 'vue'

import ProviderDistributionSection from '@/components/dashboard/ProviderDistributionSection.vue'
import BrandTrendsSection from '@/components/dashboard/BrandTrendsSection.vue'
import TopDomainsSection from '@/components/dashboard/TopDomainsSection.vue'
import TopKeywordsSection from '@/components/dashboard/TopKeywordsSection.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import StatCard from '@/components/ui/StatCard.vue'
import TagFilter from '@/components/ui/TagFilter.vue'
import { useDashboardCharts } from '@/composables/useDashboardCharts'
import { useDashboardStatus } from '@/composables/useDashboardStatus'
import { useLLMsQuery, useStatsQuery, useURLStatsQuery } from '@/queries/dashboard'
import { usePromptsQuery } from '@/queries/prompts'

const selectedTags = ref<string[]>([])

const { refresh, loading: isRefreshing } = useDashboardStatus()

const statsQuery = useStatsQuery(selectedTags)
const urlStatsQuery = useURLStatsQuery(selectedTags)
const llmsQuery = useLLMsQuery()
const promptsQuery = usePromptsQuery()

const stats = computed(() => statsQuery.data.value)
const urlStats = computed(() => urlStatsQuery.data.value)
const llms = computed(() => llmsQuery.data.value)

const allTags = computed(() => {
  const tags = new Set<string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    for (const tag of prompt.tags ?? []) {
      tags.add(tag)
    }
  }
  return Array.from(tags).sort((a, b) => a.localeCompare(b))
})

const hasActiveTagFilters = computed(() => selectedTags.value.length > 0)

const filterHint = computed(() =>
  hasActiveTagFilters.value ? 'Matching selected tags' : 'Across all LLMs',
)

const {
  providerDistribution,
  topKeywordsChartData,
  providerChartData,
  domainChartData,
  brandTrendsChartData,
} = useDashboardCharts(stats, urlStats, llms)

const isInitialLoading = computed(
  () =>
    (statsQuery.isPending.value && !statsQuery.data.value) ||
    (urlStatsQuery.isPending.value && !urlStatsQuery.data.value),
)

const errorMessage = computed(() => {
  const error = statsQuery.error.value ?? urlStatsQuery.error.value ?? llmsQuery.error.value
  if (!error) {
    return null
  }
  return error instanceof Error ? error.message : 'Failed to load dashboard data'
})

const topKeywords = computed(() => stats.value?.top_keywords ?? [])
const brandTrends = computed(() => stats.value?.brand_trends ?? [])
const topDomains = computed(() => urlStats.value?.top_domains ?? [])

function toggleTag(tag: string) {
  const index = selectedTags.value.indexOf(tag)
  if (index === -1) {
    selectedTags.value = [...selectedTags.value, tag]
    return
  }
  selectedTags.value = selectedTags.value.filter((value) => value !== tag)
}

function clearTagFilters() {
  selectedTags.value = []
}
</script>

<template>
  <div>
    <AppAlert
      v-if="errorMessage"
      title="Unable to Connect to Gego API"
      :loading="isRefreshing"
      @retry="refresh"
    >
      {{ errorMessage }}
    </AppAlert>

    <LoadingState
      v-if="isInitialLoading"
      title="Loading Analytics"
      description="Fetching data from your LLM providers..."
    />

    <template v-else-if="stats">
      <div
        v-if="allTags.length > 0"
        class="mb-6 rounded-xl border border-gray-200/60 bg-white/60 backdrop-blur-sm p-4"
      >
        <TagFilter
          :tags="allTags"
          :selected-tags="selectedTags"
          @toggle="toggleTag"
          @clear="clearTagFilters"
        />
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 mb-6 md:mb-8">
        <StatCard
          label="Total Responses"
          :value="(stats.total_responses ?? 0).toLocaleString()"
          :hint="filterHint"
          icon="file"
        />
        <StatCard
          label="Active Prompts"
          :value="stats.total_prompts ?? 0"
          :hint="hasActiveTagFilters ? 'Matching selected tags' : 'Currently running'"
          icon="lightbulb"
        />
        <StatCard
          label="LLM Providers"
          :value="stats.total_llms ?? 0"
          :hint="hasActiveTagFilters ? 'Used in filtered responses' : 'Connected models'"
          icon="server"
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6 mb-6 md:mb-8">
        <TopKeywordsSection
          :keywords="topKeywords"
          :chart-data="topKeywordsChartData"
        />
        <TopDomainsSection
          :domains="topDomains"
          :chart-data="domainChartData"
        />
      </div>

      <div class="mb-6 md:mb-8">
        <BrandTrendsSection
          :brand-trends="brandTrends"
          :chart-data="brandTrendsChartData"
          :has-keywords="topKeywords.length > 0"
        />
      </div>

      <ProviderDistributionSection
        :distribution="providerDistribution"
        :chart-data="providerChartData"
      />
    </template>
  </div>
</template>
