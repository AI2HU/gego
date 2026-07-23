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
import type { KeywordCount } from '@/types/stats'

const props = defineProps<{
  keywords: KeywordCount[]
  chartData: ChartData<'bar'>
}>()

const router = useRouter()

const topBrands = computed(() => props.keywords)

function goToSearch(term: string) {
  if (term.trim().length < 2) {
    return
  }
  void router.push(searchRouteFor(term.trim()))
}

const chartOptions = computed<ChartOptions<'bar'>>(() => ({
  ...barChartOptions,
  scales: {
    ...barChartOptions.scales,
    x: {
      ...barChartOptions.scales?.x,
      stacked: true,
    },
    y: {
      ...barChartOptions.scales?.y,
      stacked: true,
    },
  },
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

function chipClass(isTarget?: boolean) {
  return [
    'inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs cursor-pointer transition-colors',
    isTarget
      ? 'border border-dashed border-slate-400/80 bg-slate-50/60 text-slate-700 hover:border-slate-500 hover:bg-slate-100'
      : 'border border-gray-200/60 bg-slate-50/80 text-gray-600 hover:border-gray-300/60 hover:bg-slate-100 hover:text-gray-800',
  ]
}
</script>

<template>
  <AppCard>
    <template #header>
      <CardHeader
        title="Top Brands"
        subtitle="Most mentioned brands across all LLM responses"
        icon="tag"
      />
    </template>

    <div v-if="topBrands.length" class="mb-6 h-64">
      <BarChart
        :data="chartData"
        :options="chartOptions"
        aria-label="Top brand mentions across all LLM responses"
      />
    </div>

    <div v-if="topBrands.length" class="flex flex-wrap gap-1.5">
      <RouterLink
        v-for="keyword in topBrands"
        :key="keyword.keyword"
        :to="searchRouteFor(keyword.keyword)"
        :class="chipClass(keyword.is_target)"
        :title="`${keyword.count.toLocaleString()} mentions · Search${keyword.is_target ? ' · Target brand' : ''}`"
      >
        <span>{{ keyword.keyword }}</span>
        <span class="text-[10px] text-gray-400">{{ keyword.count }}</span>
      </RouterLink>
    </div>

    <EmptyState
      v-else
      title="No brand data available"
      description="Run some prompts to see brand analytics"
      icon="file"
    />
  </AppCard>
</template>
