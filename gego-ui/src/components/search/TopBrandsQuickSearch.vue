<script setup lang="ts">
import { computed } from 'vue'

import { useStatsQuery } from '@/queries/dashboard'

const props = withDefaults(
  defineProps<{
    disabled?: boolean
    activeKeyword?: string
    tags?: string[]
    limit?: number
  }>(),
  {
    disabled: false,
    activeKeyword: '',
    tags: () => [],
    limit: 10,
  },
)

const emit = defineEmits<{
  select: [brand: string]
}>()

const statsQuery = useStatsQuery(() => props.tags)

const topBrands = computed(
  () => statsQuery.data.value?.top_keywords?.slice(0, props.limit) ?? [],
)

const isInitialLoading = computed(
  () => statsQuery.isPending.value && !statsQuery.data.value,
)

function onSelect(brand: string) {
  if (props.disabled || brand.trim().length < 2) {
    return
  }
  emit('select', brand)
}
</script>

<template>
  <div v-if="isInitialLoading || topBrands.length > 0" class="mt-2">
    <p v-if="isInitialLoading" class="text-xs text-gray-400">
      Loading popular brands...
    </p>

    <div v-else class="flex flex-wrap items-center gap-1.5">
      <span class="text-[11px] text-gray-400">Popular</span>
      <button
        v-for="brand in topBrands"
        :key="brand.keyword"
        type="button"
        class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs transition-colors disabled:cursor-not-allowed disabled:opacity-50"
        :class="
          activeKeyword === brand.keyword
            ? 'bg-slate-200 text-slate-800'
            : 'text-gray-500 hover:bg-slate-100 hover:text-gray-700'
        "
        :disabled="disabled || brand.keyword.length < 2"
        :title="`${brand.count.toLocaleString()} mentions`"
        @click="onSelect(brand.keyword)"
      >
        <span>{{ brand.keyword }}</span>
        <span class="text-[10px] text-gray-400">{{ brand.count }}</span>
      </button>
    </div>
  </div>
</template>
