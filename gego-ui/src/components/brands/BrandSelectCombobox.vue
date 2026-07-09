<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'

import BrandSuggestionsList from '@/components/brands/BrandSuggestionsList.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import { useAutocompleteList } from '@/composables/useAutocompleteList'
import { filterBrandSuggestions } from '@/lib/brand-suggestions'
import type { Brand } from '@/types/brand'

const props = withDefaults(
  defineProps<{
    modelValue: string | null
    brands: Brand[]
    label?: string
    placeholder?: string
    disabled?: boolean
    suggestionLimit?: number
  }>(),
  {
    label: 'Brand',
    placeholder: 'Search brands...',
    disabled: false,
    suggestionLimit: 12,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const inputId = useId()
const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
const isEditing = ref(false)

const selectedBrand = computed(
  () => props.brands.find((brand) => brand.id === props.modelValue) ?? null,
)

const { isOpen, highlightedIndex, open, close, handleKeyDown } = useAutocompleteList(query)

const suggestions = computed(() =>
  filterBrandSuggestions(props.brands, isEditing.value ? query.value : '', props.suggestionLimit),
)

const showDropdown = computed(
  () => isOpen.value && !props.disabled && (suggestions.value.length > 0 || isEditing.value),
)

const inputValue = computed(() => {
  if (isEditing.value) {
    return query.value
  }
  return selectedBrand.value?.name ?? ''
})

const hasSelection = computed(() => Boolean(props.modelValue))

watch(
  () => props.modelValue,
  () => {
    if (!isEditing.value) {
      query.value = ''
    }
  },
)

function startEditing() {
  if (props.disabled) {
    return
  }
  isEditing.value = true
  query.value = selectedBrand.value?.name ?? ''
  open()
  requestAnimationFrame(() => {
    if (document.activeElement !== inputRef.value) {
      inputRef.value?.focus()
    }
    inputRef.value?.select()
  })
}

function handleInput(event: Event) {
  isEditing.value = true
  query.value = (event.target as HTMLInputElement).value
  open()
}

function selectBrand(brand: Brand) {
  emit('update:modelValue', brand.id)
  query.value = ''
  isEditing.value = false
  close()
}

function clearSelection(event: MouseEvent) {
  event.preventDefault()
  event.stopPropagation()
  emit('update:modelValue', null)
  query.value = ''
  isEditing.value = true
  open()
  requestAnimationFrame(() => inputRef.value?.focus())
}

function onInputBlur() {
  isEditing.value = false
  query.value = ''
  close()
}

function onKeyDown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp' || event.key === 'Enter') {
    handleKeyDown(event, suggestions.value.length, (index) => {
      const brand = suggestions.value[index]
      if (brand) {
        selectBrand(brand)
      }
    })
    return
  }

  if (event.key === 'Escape') {
    isEditing.value = false
    query.value = ''
    close()
    inputRef.value?.blur()
  }
}

function toggleDropdown() {
  if (props.disabled) {
    return
  }
  if (isOpen.value) {
    isEditing.value = false
    query.value = ''
    close()
    inputRef.value?.blur()
    return
  }
  startEditing()
}
</script>

<template>
  <div class="w-full">
    <label
      v-if="label"
      :for="inputId"
      class="mb-1.5 block text-xs font-medium uppercase tracking-wider text-gray-500"
    >
      {{ label }}
    </label>

    <div class="relative">
      <div
        class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400"
      >
        <AppIcon name="search" size="sm" />
      </div>

      <input
        :id="inputId"
        ref="inputRef"
        :value="inputValue"
        type="text"
        role="combobox"
        :aria-expanded="showDropdown"
        aria-autocomplete="list"
        :aria-controls="`${inputId}-listbox`"
        :placeholder="placeholder"
        :disabled="disabled || brands.length === 0"
        class="w-full rounded-lg border border-gray-300 bg-white py-2.5 pr-20 pl-9 text-sm text-gray-900 shadow-sm transition-colors placeholder:text-gray-400 focus:border-transparent focus:ring-2 focus:ring-slate-500 focus:outline-none disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500"
        :class="showDropdown ? 'ring-2 ring-slate-500 border-transparent' : ''"
        autocomplete="off"
        @focus="startEditing"
        @input="handleInput"
        @keydown="onKeyDown"
        @blur="onInputBlur"
      />

      <div class="absolute inset-y-0 right-0 flex items-center gap-0.5 pr-2">
        <button
          v-if="hasSelection && !disabled"
          type="button"
          class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
          aria-label="Clear brand selection"
          @mousedown.prevent="clearSelection"
        >
          <AppIcon name="close" size="sm" />
        </button>

        <button
          type="button"
          class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:pointer-events-none"
          :disabled="disabled || brands.length === 0"
          aria-label="Toggle brand list"
          @mousedown.prevent="toggleDropdown"
        >
          <AppIcon
            name="chevron-down"
            size="sm"
            class="transition-transform duration-200"
            :class="showDropdown ? 'rotate-180' : ''"
          />
        </button>
      </div>

      <BrandSuggestionsList
        v-if="showDropdown && suggestions.length > 0"
        :id="`${inputId}-listbox`"
        :suggestions="suggestions"
        :highlighted-index="highlightedIndex"
        :selected-id="modelValue"
        @pick="selectBrand"
      />

      <div
        v-else-if="showDropdown && isEditing && suggestions.length === 0"
        class="absolute z-10 mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-sm text-gray-500 shadow-lg"
      >
        No brands match your search
      </div>
    </div>
  </div>
</template>
