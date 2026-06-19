<script setup lang="ts">
import { computed } from 'vue'
import type { ChartData } from 'chart.js'

import BarChart from '@/components/charts/BarChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { card } from '@/design/classes'
import type { DomainMentionStats } from '@/types/stats'

const props = defineProps<{
  domains: DomainMentionStats[]
  chartData: ChartData<'bar'>
}>()

const topDomains = computed(() => props.domains.slice(0, 10))
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
        aria-label="Top cited domains across responses"
      />
    </div>

    <div v-if="topDomains.length" class="space-y-3">
      <div
        v-for="domain in topDomains"
        :key="domain.domain"
        :class="[card.inset, 'flex items-center justify-between hover:bg-slate-100 transition-colors duration-200 !p-3']"
      >
        <div class="min-w-0 pr-3">
          <div class="text-sm font-medium text-gray-800 break-all">{{ domain.domain }}</div>
          <div class="text-xs text-gray-500">
            {{ domain.unique_url_count }} unique URL{{ domain.unique_url_count !== 1 ? 's' : '' }}
          </div>
        </div>
        <div class="text-right shrink-0">
          <div class="text-lg font-semibold text-slate-700">{{ domain.citations }}</div>
          <div class="text-xs text-gray-500">citations</div>
        </div>
      </div>
    </div>

    <EmptyState
      v-else
      title="No domain data available"
      description="Run some prompts with search enabled to see domain analytics"
      icon="globe"
    />
  </AppCard>
</template>
