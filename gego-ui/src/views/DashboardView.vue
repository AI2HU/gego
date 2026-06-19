<script setup lang="ts">
import { computed } from 'vue'

import ProviderDistributionSection from '@/components/dashboard/ProviderDistributionSection.vue'
import TopDomainsSection from '@/components/dashboard/TopDomainsSection.vue'
import TopKeywordsSection from '@/components/dashboard/TopKeywordsSection.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import StatCard from '@/components/ui/StatCard.vue'
import { useDashboardCharts } from '@/composables/useDashboardCharts'
import { useDashboardStatus } from '@/composables/useDashboardStatus'
import { useLLMsQuery, useStatsQuery, useURLStatsQuery } from '@/queries/dashboard'

const { refresh, loading: isRefreshing } = useDashboardStatus()

const statsQuery = useStatsQuery()
const urlStatsQuery = useURLStatsQuery()
const llmsQuery = useLLMsQuery()

const stats = computed(() => statsQuery.data.value)
const urlStats = computed(() => urlStatsQuery.data.value)
const llms = computed(() => llmsQuery.data.value)

const {
  providerDistribution,
  topKeywordsChartData,
  providerChartData,
  domainChartData,
} = useDashboardCharts(stats, urlStats, llms)

const isInitialLoading = computed(
  () => statsQuery.isPending.value && !statsQuery.data.value,
)

const errorMessage = computed(() => {
  const error = statsQuery.error.value ?? urlStatsQuery.error.value ?? llmsQuery.error.value
  if (!error) {
    return null
  }
  return error instanceof Error ? error.message : 'Failed to load dashboard data'
})

const topKeywords = computed(() => stats.value?.top_keywords ?? [])
const topDomains = computed(() => urlStats.value?.top_domains ?? [])
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
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 mb-6 md:mb-8">
        <StatCard
          label="Total Responses"
          :value="(stats.total_responses ?? 0).toLocaleString()"
          hint="Across all LLMs"
          icon="file"
        />
        <StatCard
          label="Active Prompts"
          :value="stats.total_prompts ?? 0"
          hint="Currently running"
          icon="lightbulb"
        />
        <StatCard
          label="LLM Providers"
          :value="stats.total_llms ?? 0"
          hint="Connected models"
          icon="server"
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6 mb-6 md:mb-8">
        <TopKeywordsSection
          :keywords="topKeywords"
          :chart-data="topKeywordsChartData"
        />
        <ProviderDistributionSection
          :distribution="providerDistribution"
          :chart-data="providerChartData"
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
        <TopDomainsSection
          :domains="topDomains"
          :chart-data="domainChartData"
        />
      </div>
    </template>
  </div>
</template>
