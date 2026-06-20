<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { ChartData, ChartOptions } from 'chart.js'

import BarChart from '@/components/charts/BarChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { barChartOptions } from '@/design/chartTheme'
import { searchRouteFor } from '@/lib/search-navigation'
import type { DomainMentionStats } from '@/types/stats'

const props = defineProps<{
  domains: DomainMentionStats[]
  chartData: ChartData<'bar'>
}>()

const router = useRouter()

const topDomains = computed(() => props.domains.slice(0, 10))

function goToSearch(term: string) {
  if (term.trim().length < 2) {
    return
  }
  void router.push(searchRouteFor(term.trim()))
}

const chartOptions = computed<ChartOptions<'bar'>>(() => ({
  ...barChartOptions,
  onClick(_event, elements, chart) {
    if (elements.length === 0) {
      return
    }
    const index = elements[0]?.index
    if (index == null) {
      return
    }
    const label = chart.data.labels?.[index]
    if (typeof label === 'string') {
      goToSearch(label)
    }
  },
}))

const chipClass =
  'inline-flex items-center gap-1 rounded-md border border-gray-200/60 bg-slate-50/80 px-2 py-0.5 text-xs text-gray-600 cursor-pointer transition-colors hover:border-gray-300/60 hover:bg-slate-100 hover:text-gray-800'
</script>

<template>
  <AppCard>
    <template #header>
      <CardHeader
        title="Top Domains"
        subtitle="Domains most often cited in responses"
        icon="globe"
      />
    </template>

    <div v-if="topDomains.length" class="mb-6 h-64">
      <BarChart
        :data="chartData"
        :options="chartOptions"
        aria-label="Top cited domains across responses"
      />
    </div>

    <div v-if="topDomains.length" class="flex flex-wrap gap-1.5">
      <RouterLink
        v-for="domain in topDomains"
        :key="domain.domain"
        :to="searchRouteFor(domain.domain)"
        :class="chipClass"
        :title="`${domain.citations.toLocaleString()} citations, ${domain.unique_url_count} unique URL${domain.unique_url_count !== 1 ? 's' : ''} · Search`"
      >
        <span>{{ domain.domain }}</span>
        <span class="text-[10px] text-gray-400">{{ domain.citations }}</span>
      </RouterLink>
    </div>

    <EmptyState
      v-else
      title="No domain data available"
      description="Run some prompts with search enabled to see domain analytics"
      icon="globe"
    />
  </AppCard>
</template>
