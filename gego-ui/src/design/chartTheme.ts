import type { ChartOptions } from 'chart.js'

export const chartColors = {
  primary: 'rgba(15, 23, 42, 0.7)',
  blue: 'rgba(37, 99, 235, 0.7)',
  doughnut: ['#0f172a', '#1f2937', '#4b5563', '#6b7280', '#9ca3af', '#d1d5db'],
  line: ['#0f172a', '#2563eb', '#059669', '#d97706', '#7c3aed'],
  tick: '#4b5563',
  grid: '#e5e7eb',
} as const

export const barChartOptions: ChartOptions<'bar'> = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index',
    intersect: false,
  },
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      backgroundColor: 'rgba(15, 23, 42, 0.92)',
      titleColor: '#f9fafb',
      bodyColor: '#f9fafb',
      padding: 10,
      cornerRadius: 8,
      displayColors: false,
      callbacks: {
        title(items) {
          return items[0]?.label ?? ''
        },
        label(context) {
          const value = context.parsed.y
          if (value == null) {
            return ''
          }
          return value.toLocaleString()
        },
      },
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

export const lineChartOptions: ChartOptions<'line'> = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index',
    intersect: false,
  },
  plugins: {
    legend: {
      position: 'bottom',
      labels: {
        color: chartColors.tick,
        boxWidth: 12,
        usePointStyle: true,
      },
    },
  },
  scales: {
    x: {
      grid: {
        display: false,
      },
      ticks: {
        color: chartColors.tick,
        maxRotation: 0,
        autoSkip: true,
        maxTicksLimit: 8,
      },
    },
    y: {
      beginAtZero: true,
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
