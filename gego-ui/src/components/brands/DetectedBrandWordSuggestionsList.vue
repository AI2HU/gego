<script setup lang="ts">
import type { SuggestedBrandWord } from '@/types/brand'

defineProps<{
  suggestions: SuggestedBrandWord[]
  highlightedIndex: number
}>()

const emit = defineEmits<{
  pick: [item: SuggestedBrandWord]
}>()
</script>

<template>
  <ul
    class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
  >
    <li
      v-for="(item, index) in suggestions"
      :key="item.word"
    >
      <button
        type="button"
        class="flex w-full items-center justify-between gap-3 px-3 py-2 text-left text-sm transition-colors hover:bg-slate-50"
        :class="index === highlightedIndex ? 'bg-slate-100' : ''"
        @mousedown.prevent="emit('pick', item)"
      >
        <span class="min-w-0 truncate font-medium text-gray-900">{{ item.word }}</span>
        <span class="shrink-0 rounded-full bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-500">
          {{ item.count }}
        </span>
      </button>
    </li>
  </ul>
</template>
