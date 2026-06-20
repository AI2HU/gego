<script setup lang="ts">
import type { ChartData } from 'chart.js'

import DoughnutChart from '@/components/charts/DoughnutChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import IconBox from '@/components/ui/IconBox.vue'
import { card } from '@/design/classes'
import type { ProviderDistribution } from '@/types/model'

defineProps<{
  distribution: ProviderDistribution[]
  chartData: ChartData<'doughnut'>
}>()
</script>

<template>
  <AppCard>
    <template #header>
      <CardHeader
        title="Provider Distribution"
        subtitle="Response distribution across LLM providers"
        icon="server"
      />
    </template>

    <div v-if="distribution.length" class="mb-6 h-64">
      <DoughnutChart
        :data="chartData"
        aria-label="Response distribution by LLM provider"
      />
    </div>

    <div v-if="distribution.length" class="space-y-4">
      <div
        v-for="item in distribution"
        :key="item.provider"
        :class="[card.inset, 'hover:bg-slate-100 transition-colors duration-200']"
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-3">
            <IconBox icon="server" size="md" class="!bg-slate-200" />
            <div>
              <h4 class="font-semibold text-gray-800 capitalize">{{ item.provider }}</h4>
              <p class="text-xs text-gray-500">
                {{ item.modelCount }} model{{ item.modelCount !== 1 ? 's' : '' }}
              </p>
            </div>
          </div>
          <div class="text-right">
            <div class="text-2xl font-semibold text-slate-700">
              {{ item.totalResponses.toLocaleString() }}
            </div>
            <div class="text-xs text-gray-500">responses</div>
          </div>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-3 mb-2">
          <div
            class="bg-slate-500 h-3 rounded-full transition-all duration-500"
            :style="{ width: `${item.percentage}%` }"
          />
        </div>
        <div class="flex justify-between items-center text-xs text-gray-600">
          <span>{{ item.percentage.toFixed(1) }}% of total</span>
          <span>Avg {{ Math.round(item.avgTokens) }} tokens</span>
        </div>
      </div>
    </div>

    <EmptyState
      v-else
      title="No provider data available"
      description="Run some prompts to see provider distribution"
      icon="server"
    />
  </AppCard>
</template>
