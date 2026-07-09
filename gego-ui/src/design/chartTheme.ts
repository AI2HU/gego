import type { ChartOptions } from 'chart.js'

export const chartColors = {
  primary: 'rgba(15, 23, 42, 0.7)',
  blue: 'rgba(37, 99, 235, 0.7)',
  emerald: 'rgba(5, 150, 105, 0.75)',
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

export const horizontalBarChartOptions: ChartOptions<'bar'> = {
  ...barChartOptions,
  indexAxis: 'y',
  interaction: {
    mode: 'index',
    axis: 'y',
    intersect: false,
  },
  plugins: {
    ...barChartOptions.plugins,
    tooltip: {
      ...barChartOptions.plugins?.tooltip,
      position: 'nearest',
      callbacks: {
        title(items) {
          return items[0]?.label ?? ''
        },
        label(context) {
          const value = context.parsed.x
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
      beginAtZero: true,
      grid: {
        color: chartColors.grid,
      },
      ticks: {
        color: chartColors.tick,
        precision: 0,
      },
    },
    y: {
      grid: {
        display: false,
      },
      ticks: {
        color: chartColors.tick,
        autoSkip: false,
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
