<script setup lang="ts">
import type { Brand } from '@/types/brand'

defineProps<{
  suggestions: Brand[]
  highlightedIndex: number
}>()

const emit = defineEmits<{
  pick: [brand: Brand]
}>()
</script>

<template>
  <ul
    class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
  >
    <li
      v-for="(brand, index) in suggestions"
      :key="brand.id"
    >
      <button
        type="button"
        class="flex w-full flex-col px-3 py-2 text-left text-sm transition-colors hover:bg-slate-50"
        :class="index === highlightedIndex ? 'bg-slate-100' : ''"
        @mousedown.prevent="emit('pick', brand)"
      >
        <span class="font-medium text-gray-900">{{ brand.name }}</span>
        <span
          v-if="brand.aliases.length > 0"
          class="mt-0.5 truncate text-xs text-gray-500"
        >
          {{ brand.aliases.map((item) => item.alias).join(', ') }}
        </span>
      </button>
    </li>
  </ul>
</template>
