<script setup lang="ts">
import { computed } from 'vue'

import DetectedBrandWordSuggestionsList from '@/components/brands/DetectedBrandWordSuggestionsList.vue'
import { useAutocompleteList } from '@/composables/useAutocompleteList'
import { input } from '@/design/classes'
import { filterDetectedBrandWordSuggestions } from '@/lib/brand-suggestions'
import type { SuggestedBrandWord } from '@/types/brand'

const props = defineProps<{
  modelValue: string
  suggestions: SuggestedBrandWord[]
  placeholder?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { isOpen, highlightedIndex, open, close, handleKeyDown } = useAutocompleteList(
  () => props.modelValue,
)

const filteredSuggestions = computed(() =>
  filterDetectedBrandWordSuggestions(props.suggestions, props.modelValue),
)

function pickWord(item: SuggestedBrandWord) {
  emit('update:modelValue', item.word)
  close()
}

function handleInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
  open()
}

function onKeyDown(event: KeyboardEvent) {
  handleKeyDown(event, filteredSuggestions.value.length, (index) => {
    const item = filteredSuggestions.value[index]
    if (item) {
      pickWord(item)
    }
  })
}
</script>

<template>
  <div class="relative">
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

    <DetectedBrandWordSuggestionsList
      v-if="isOpen && filteredSuggestions.length > 0"
      :suggestions="filteredSuggestions"
      :highlighted-index="highlightedIndex"
      @pick="pickWord"
    />
  </div>
</template>
