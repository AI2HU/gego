<script setup lang="ts">
import AppIcon from '@/components/icons/AppIcon.vue'
import type { Brand } from '@/types/brand'

defineProps<{
  suggestions: Brand[]
  highlightedIndex: number
  selectedId?: string | null
  id?: string
}>()

const emit = defineEmits<{
  pick: [brand: Brand]
}>()
</script>

<template>
  <ul
    :id="id"
    role="listbox"
    class="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
  >
    <li
      v-for="(brand, index) in suggestions"
      :key="brand.id"
      role="option"
      :aria-selected="brand.id === selectedId"
    >
      <button
        type="button"
        class="flex w-full items-start gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-slate-50"
        :class="[
          index === highlightedIndex ? 'bg-slate-100' : '',
          brand.id === selectedId ? 'bg-emerald-50/80 hover:bg-emerald-50' : '',
        ]"
        @mousedown.prevent="emit('pick', brand)"
      >
        <span class="min-w-0 flex-1">
          <span class="font-medium text-gray-900">{{ brand.name }}</span>
          <span
            v-if="brand.aliases.length > 0"
            class="mt-0.5 block truncate text-xs text-gray-500"
          >
            {{ brand.aliases.map((item) => item.alias).join(', ') }}
          </span>
        </span>
        <AppIcon
          v-if="brand.id === selectedId"
          name="check"
          size="sm"
          class="mt-0.5 shrink-0 text-emerald-600"
        />
      </button>
    </li>
  </ul>
</template>
