<script setup lang="ts">
import { computed } from 'vue'

import BarChart from '@/components/charts/BarChart.vue'
import AppCard from '@/components/ui/AppCard.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import type { ChartData } from 'chart.js'
import { card } from '@/design/classes'
import type { KeywordCount } from '@/types/stats'

const props = defineProps<{
  keywords: KeywordCount[]
  chartData: ChartData<'bar'>
}>()

const topKeywords = computed(() => props.keywords.slice(0, 10))
const maxCount = computed(() => props.keywords[0]?.count ?? 1)
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
        aria-label="Top keyword mentions across all LLM responses"
      />
    </div>

    <div v-if="topKeywords.length" class="space-y-3">
      <div
        v-for="(keyword, index) in topKeywords"
        :key="keyword.keyword"
        :class="[card.inset, 'flex items-center hover:bg-slate-100 transition-colors duration-200 !p-3']"
      >
        <div class="w-8 h-8 bg-slate-200 rounded-lg flex items-center justify-center text-sm font-semibold text-slate-700 mr-3 shrink-0">
          {{ index + 1 }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex justify-between items-center mb-1 gap-2">
            <span class="font-medium text-gray-800 truncate">{{ keyword.keyword }}</span>
            <span class="text-sm text-gray-600 shrink-0">{{ keyword.count }} mentions</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-2">
            <div
              class="bg-slate-500 h-2 rounded-full transition-all duration-500"
              :style="{ width: `${(keyword.count / maxCount) * 100}%` }"
            />
          </div>
        </div>
      </div>
    </div>

    <EmptyState
      v-else
      title="No keyword data available"
      description="Run some prompts to see keyword analytics"
      icon="file"
    />
  </AppCard>
</template>
