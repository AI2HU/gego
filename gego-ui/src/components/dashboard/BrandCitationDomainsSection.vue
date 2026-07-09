<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { ChartData, ChartOptions } from 'chart.js'

import BarChart from '@/components/charts/BarChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import BrandCitationTargetCombobox from '@/components/dashboard/BrandCitationTargetCombobox.vue'
import { chartColors, horizontalBarChartOptions } from '@/design/chartTheme'
import { searchRouteFor } from '@/lib/search-navigation'
import type { BrandCitationTarget } from '@/lib/brand-citation-target'
import type { BrandCitationDomainsResponse } from '@/types/stats'

const props = defineProps<{
  targets: BrandCitationTarget[]
  selectedTarget: BrandCitationTarget | null
  data: BrandCitationDomainsResponse | undefined
  isLoading: boolean
}>()

const emit = defineEmits<{
  'update:selectedTarget': [value: BrandCitationTarget | null]
}>()

const router = useRouter()

const domains = computed(() => props.data?.domains ?? [])
const brandName = computed(() => props.data?.brand_name ?? '')
const totalHits = computed(() => props.data?.total_hits ?? 0)

const chartItems = computed(() => domains.value)

const chartData = computed<ChartData<'bar'>>(() => {
  if (!chartItems.value.length) {
    return { labels: [], datasets: [] }
  }

  return {
    labels: chartItems.value.map((item) => item.domain),
    datasets: [
      {
        data: chartItems.value.map((item) => item.citations),
        backgroundColor: chartColors.emerald,
        hoverBackgroundColor: 'rgba(5, 150, 105, 0.9)',
        borderRadius: 6,
      },
    ],
  }
})

function goToSearch(term: string) {
  if (term.trim().length < 2) {
    return
  }
  void router.push(searchRouteFor(term.trim()))
}

const chartOptions = computed<ChartOptions<'bar'>>(() => ({
  ...horizontalBarChartOptions,
  scales: {
    ...horizontalBarChartOptions.scales,
    y: {
      ...horizontalBarChartOptions.scales?.y,
      reverse: true,
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

const chartHeight = computed(() => `${Math.max(chartItems.value.length * 36, 220)}px`)

const chipClass =
  'inline-flex items-center gap-1 rounded-md border border-gray-200/60 bg-emerald-50/80 px-2 py-0.5 text-xs text-gray-600 cursor-pointer transition-colors hover:border-emerald-200/80 hover:bg-emerald-100/80 hover:text-gray-800'
</script>

<template>
  <AppCard>
    <template #header>
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <CardHeader
          title="Brand Citation Sources"
          subtitle="Domains cited closest to brand mentions in responses"
          icon="external-link"
        />

        <div class="w-full sm:w-72">
          <BrandCitationTargetCombobox
            :model-value="selectedTarget"
            :targets="targets"
            label="Filter by brand"
            placeholder="Search brands or detected words..."
            @update:model-value="emit('update:selectedTarget', $event)"
          />
        </div>
      </div>
    </template>

    <LoadingState
      v-if="isLoading && selectedTarget"
      title="Loading citation sources"
      description="Analyzing brand mentions and nearby citations..."
    />

    <template v-else-if="selectedTarget && domains.length">
      <p class="mb-4 text-sm text-gray-500">
        <span class="font-medium text-gray-700">{{ brandName }}</span>
        was linked to
        <span class="font-medium text-gray-700">{{ totalHits.toLocaleString() }}</span>
        nearby citation{{ totalHits === 1 ? '' : 's' }}
        across
        <span class="font-medium text-gray-700">{{ domains.length.toLocaleString() }}</span>
        domain{{ domains.length === 1 ? '' : 's' }}.
      </p>

      <div class="mb-6" :style="{ height: chartHeight }">
        <BarChart
          :data="chartData"
          :options="chartOptions"
          aria-label="Domains most often cited near the selected brand"
        />
      </div>

      <div class="flex flex-wrap gap-1.5">
        <RouterLink
          v-for="domain in domains"
          :key="domain.domain"
          :to="searchRouteFor(domain.domain)"
          :class="chipClass"
          :title="`${domain.citations.toLocaleString()} nearby citation${domain.citations !== 1 ? 's' : ''}, ${domain.unique_url_count} unique URL${domain.unique_url_count !== 1 ? 's' : ''} · Search`"
        >
          <span>{{ domain.domain }}</span>
          <span class="text-[10px] text-gray-400">{{ domain.citations }}</span>
        </RouterLink>
      </div>
    </template>

    <EmptyState
      v-else-if="!selectedTarget"
      title="Select a brand or detected word"
      description="Choose a configured brand or a detected word to see which domains are cited closest to its mentions"
      icon="external-link"
    />

    <EmptyState
      v-else
      title="No nearby citation data"
      description="Run prompts with search enabled and inline citations to see which domains cite this brand"
      icon="external-link"
    />
  </AppCard>
</template>
