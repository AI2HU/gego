<script setup lang="ts">
import { computed } from 'vue'
import type { ChartData } from 'chart.js'

import LineChart from '@/components/charts/LineChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import type { BrandTrendSeries } from '@/types/stats'

const props = defineProps<{
  brandTrends: BrandTrendSeries[]
  chartData: ChartData<'line'>
  hasKeywords: boolean
}>()

const hasTrends = computed(() =>
  props.brandTrends.some((series) => series.points.some((point) => point.count > 0)),
)

const emptyDescription = computed(() => {
  if (!props.hasKeywords) {
    return 'Run prompts to start tracking brand mentions over time.'
  }
  if (props.brandTrends.length === 0) {
    return 'Trend data is not available from the API yet. Restart the server to pick up the latest backend changes.'
  }
  return 'No brand mentions were recorded in the last 30 days.'
})
</script>

<template>
  <AppCard>
    <template #header>
      <CardHeader
        title="Brand Mentions Over Time"
        subtitle="Daily mentions for your top brands over the last 30 days"
        icon="chart-line"
      />
    </template>

    <div v-if="hasTrends" class="h-72">
      <LineChart
        :data="chartData"
        aria-label="Daily brand mention trends over the last 30 days"
      />
    </div>

    <EmptyState
      v-else
      title="No brand trend data yet"
      :description="emptyDescription"
      icon="chart-line"
    />
  </AppCard>
</template>
