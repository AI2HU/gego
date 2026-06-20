<script setup lang="ts">
import { Doughnut } from 'vue-chartjs'
import type { ChartData, ChartOptions } from 'chart.js'

import '@/lib/charts'
import { doughnutChartOptions } from '@/design/chartTheme'

withDefaults(
  defineProps<{
    data: ChartData<'doughnut'>
    options?: ChartOptions<'doughnut'>
    ariaLabel?: string
  }>(),
  {
    options: () => doughnutChartOptions,
  },
)
</script>

<template>
  <Doughnut :data="data" :options="options" :aria-label="ariaLabel" />
</template>
