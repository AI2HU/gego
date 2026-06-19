<script setup lang="ts">
import { Bar } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'

import '@/lib/charts'
import { barChartOptions } from '@/design/chartTheme'

withDefaults(
  defineProps<{
    data: ChartData<'bar'>
    options?: ChartOptions<'bar'>
    ariaLabel?: string
  }>(),
  {
    options: () => barChartOptions,
  },
)
</script>

<template>
  <Bar :data="data" :options="options" :aria-label="ariaLabel" />
</template>
