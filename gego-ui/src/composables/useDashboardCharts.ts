import type { ChartData } from 'chart.js'
import type { Ref } from 'vue'
import { computed } from 'vue'
import { format } from 'date-fns'

import { chartColors } from '@/design/chartTheme'
import type { ModelResponse, ProviderDistribution } from '@/types/model'
import type { StatsResponse, URLStatsResponse } from '@/types/stats'

function buildProviderDistribution(
  stats: StatsResponse | undefined,
  llms: ModelResponse[] | undefined,
): ProviderDistribution[] {
  if (!stats?.llm_stats?.length) {
    return []
  }

  const providerMap = new Map<
    string,
    { provider: string; totalResponses: number; totalTokens: number; models: Set<string> }
  >()
  const totalResponses = stats.llm_stats.reduce((sum, stat) => sum + stat.total_responses, 0)

  for (const stat of stats.llm_stats) {
    const llm = llms?.find((item) => item.id === stat.llm_id)
    const provider = llm?.provider ?? 'unknown'

    if (!providerMap.has(provider)) {
      providerMap.set(provider, {
        provider,
        totalResponses: 0,
        totalTokens: 0,
        models: new Set(),
      })
    }

    const providerStat = providerMap.get(provider)!
    providerStat.totalResponses += stat.total_responses
    providerStat.totalTokens += stat.avg_tokens * stat.total_responses
    if (llm?.model) {
      providerStat.models.add(llm.model)
    }
  }

  return Array.from(providerMap.values())
    .map((item) => ({
      provider: item.provider,
      totalResponses: item.totalResponses,
      avgTokens: item.totalResponses > 0 ? item.totalTokens / item.totalResponses : 0,
      modelCount: item.models.size,
      percentage: totalResponses > 0 ? (item.totalResponses / totalResponses) * 100 : 0,
    }))
    .sort((a, b) => b.totalResponses - a.totalResponses)
}

export function useDashboardCharts(
  stats: Ref<StatsResponse | undefined>,
  urlStats: Ref<URLStatsResponse | undefined>,
  llms: Ref<ModelResponse[] | undefined>,
) {
  const providerDistribution = computed(() => buildProviderDistribution(stats.value, llms.value))

  const topKeywordsChartData = computed<ChartData<'bar'>>(() => {
    if (!stats.value?.top_keywords?.length) {
      return { labels: [], datasets: [] }
    }

    const items = stats.value.top_keywords.slice(0, 10)

    return {
      labels: items.map((item) => item.keyword),
      datasets: [
        {
          data: items.map((item) => item.count),
          backgroundColor: chartColors.primary,
          hoverBackgroundColor: 'rgba(15, 23, 42, 0.9)',
          borderRadius: 6,
        },
      ],
    }
  })

  const providerChartData = computed<ChartData<'doughnut'>>(() => {
    const distribution = providerDistribution.value
    if (!distribution.length) {
      return { labels: [], datasets: [] }
    }

    const labels = distribution.map((item) => item.provider)
    const data = distribution.map((item) => item.totalResponses)

    return {
      labels,
      datasets: [
        {
          data,
          backgroundColor: labels.map((_, index) => chartColors.doughnut[index % chartColors.doughnut.length]),
        },
      ],
    }
  })

  const domainChartData = computed<ChartData<'bar'>>(() => {
    if (!urlStats.value?.top_domains?.length) {
      return { labels: [], datasets: [] }
    }

    const items = urlStats.value.top_domains.slice(0, 10)

    return {
      labels: items.map((item) => item.domain),
      datasets: [
        {
          data: items.map((item) => item.citations),
          backgroundColor: chartColors.blue,
          hoverBackgroundColor: 'rgba(37, 99, 235, 0.9)',
          borderRadius: 6,
        },
      ],
    }
  })

  const brandTrendsChartData = computed<ChartData<'line'>>(() => {
    const series = stats.value?.brand_trends ?? []
    if (!series.length || !series[0]?.points.length) {
      return { labels: [], datasets: [] }
    }

    const labels = series[0].points.map((point) =>
      format(new Date(point.timestamp), 'MMM d'),
    )

    return {
      labels,
      datasets: series.map((item, index) => ({
        label: item.keyword,
        data: item.points.map((point) => point.count),
        borderColor: chartColors.line[index % chartColors.line.length],
        backgroundColor: 'transparent',
        tension: 0.3,
        pointRadius: 2,
        pointHoverRadius: 4,
      })),
    }
  })

  return {
    providerDistribution,
    topKeywordsChartData,
    providerChartData,
    domainChartData,
    brandTrendsChartData,
  }
}
