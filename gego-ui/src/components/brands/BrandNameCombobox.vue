<script setup lang="ts">
import { computed } from 'vue'

import BrandSuggestionsList from '@/components/brands/BrandSuggestionsList.vue'
import { useAutocompleteList } from '@/composables/useAutocompleteList'
import { input } from '@/design/classes'
import { filterBrandSuggestions } from '@/lib/brand-suggestions'
import { useBrandsQuery } from '@/queries/brands'
import type { Brand } from '@/types/brand'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  select: [brand: Brand]
}>()

const brandsQuery = useBrandsQuery()
const { isOpen, highlightedIndex, open, close, handleKeyDown } = useAutocompleteList(
  () => props.modelValue,
)

const brands = computed(() => brandsQuery.data.value ?? [])

const suggestions = computed(() => filterBrandSuggestions(brands.value, props.modelValue))

function selectBrand(brand: Brand) {
  emit('update:modelValue', brand.name)
  emit('select', brand)
  close()
}

function handleInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
  open()
}

function onKeyDown(event: KeyboardEvent) {
  handleKeyDown(event, suggestions.value.length, (index) => {
    const brand = suggestions.value[index]
    if (brand) {
      selectBrand(brand)
    }
  })
}
</script>

<template>
  <div class="relative" @mouseleave="close">
    <input
      :value="modelValue"
      type="text"
      :placeholder="placeholder"
      :disabled="disabled"
      :class="input.base"
      autocomplete="off"
      @input="handleInput"
      @focus="open"
      @keydown="onKeyDown"
      @blur="close"
    />

    <BrandSuggestionsList
      v-if="isOpen && suggestions.length > 0"
      :suggestions="suggestions"
      :highlighted-index="highlightedIndex"
      @pick="selectBrand"
    />
  </div>
</template>
