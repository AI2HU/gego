<script setup lang="ts">
import { Line } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'

import '@/lib/charts'
import { lineChartOptions } from '@/design/chartTheme'

withDefaults(
  defineProps<{
    data: ChartData<'line'>
    options?: ChartOptions<'line'>
    ariaLabel?: string
  }>(),
  {
    options: () => lineChartOptions,
  },
)
</script>

<template>
  <Line :data="data" :options="options" :aria-label="ariaLabel" />
</template>
