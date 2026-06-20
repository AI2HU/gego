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

const topKeywords = computed(() => props.keywords.slice(0, 10))

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
        title="Top Keywords"
        subtitle="Most mentioned keywords across all LLM responses"
        icon="tag"
      />
    </template>

    <div v-if="topKeywords.length" class="mb-6 h-64">
      <BarChart
        :data="chartData"
        :options="chartOptions"
        aria-label="Top keyword mentions across all LLM responses"
      />
    </div>

    <div v-if="topKeywords.length" class="flex flex-wrap gap-1.5">
      <RouterLink
        v-for="keyword in topKeywords"
        :key="keyword.keyword"
        :to="searchRouteFor(keyword.keyword)"
        :class="chipClass"
        :title="`${keyword.count.toLocaleString()} mentions · Search`"
      >
        <span>{{ keyword.keyword }}</span>
        <span class="text-[10px] text-gray-400">{{ keyword.count }}</span>
      </RouterLink>
    </div>

    <EmptyState
      v-else
      title="No keyword data available"
      description="Run some prompts to see keyword analytics"
      icon="file"
    />
  </AppCard>
</template>
