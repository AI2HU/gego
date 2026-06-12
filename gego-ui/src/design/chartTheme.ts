import type { ChartOptions } from 'chart.js'

export const chartColors = {
  primary: 'rgba(15, 23, 42, 0.7)',
  blue: 'rgba(37, 99, 235, 0.7)',
  doughnut: ['#0f172a', '#1f2937', '#4b5563', '#6b7280', '#9ca3af', '#d1d5db'],
  tick: '#4b5563',
  grid: '#e5e7eb',
} as const

export const barChartOptions: ChartOptions<'bar'> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
  },
  scales: {
    x: {
      grid: {
        display: false,
      },
      ticks: {
        color: chartColors.tick,
        maxRotation: 45,
        minRotation: 45,
      },
    },
    y: {
      grid: {
        color: chartColors.grid,
      },
      ticks: {
        color: chartColors.tick,
        precision: 0,
      },
    },
  },
}

export const doughnutChartOptions: ChartOptions<'doughnut'> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom',
      labels: {
        color: chartColors.tick,
      },
    },
  },
}
