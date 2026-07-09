<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import { useAutocompleteList } from '@/composables/useAutocompleteList'
import {
  brandCitationTargetKey,
  filterBrandCitationTargets,
  type BrandCitationTarget,
} from '@/lib/brand-citation-target'

const props = withDefaults(
  defineProps<{
    modelValue: BrandCitationTarget | null
    targets: BrandCitationTarget[]
    label?: string
    placeholder?: string
    disabled?: boolean
    suggestionLimit?: number
  }>(),
  {
    label: 'Filter by brand',
    placeholder: 'Search brands or detected words...',
    disabled: false,
    suggestionLimit: 12,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: BrandCitationTarget | null]
}>()

const inputId = useId()
const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
const isEditing = ref(false)

const selectedKey = computed(() =>
  props.modelValue ? brandCitationTargetKey(props.modelValue) : null,
)

const { isOpen, highlightedIndex, open, close, handleKeyDown } = useAutocompleteList(query)

const suggestions = computed(() =>
  filterBrandCitationTargets(
    props.targets,
    isEditing.value ? query.value : '',
    props.suggestionLimit,
  ),
)

const showDropdown = computed(
  () => isOpen.value && !props.disabled && (suggestions.value.length > 0 || isEditing.value),
)

const inputValue = computed(() => {
  if (isEditing.value) {
    return query.value
  }
  return props.modelValue?.label ?? ''
})

const hasSelection = computed(() => Boolean(props.modelValue))
const isEmpty = computed(() => props.targets.length === 0)

watch(
  () => props.modelValue,
  () => {
    if (!isEditing.value) {
      query.value = ''
    }
  },
)

function startEditing() {
  if (props.disabled || isEmpty.value) {
    return
  }
  isEditing.value = true
  query.value = props.modelValue?.label ?? ''
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

function selectTarget(target: BrandCitationTarget) {
  emit('update:modelValue', target)
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
      const target = suggestions.value[index]
      if (target) {
        selectTarget(target)
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
  if (props.disabled || isEmpty.value) {
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
        :disabled="disabled || isEmpty"
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
          aria-label="Clear selection"
          @mousedown.prevent="clearSelection"
        >
          <AppIcon name="close" size="sm" />
        </button>

        <button
          type="button"
          class="rounded-md p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:pointer-events-none"
          :disabled="disabled || isEmpty"
          aria-label="Toggle list"
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

      <ul
        v-if="showDropdown && suggestions.length > 0"
        :id="`${inputId}-listbox`"
        role="listbox"
        class="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
      >
        <li
          v-for="(target, index) in suggestions"
          :key="brandCitationTargetKey(target)"
          role="option"
          :aria-selected="brandCitationTargetKey(target) === selectedKey"
        >
          <button
            type="button"
            class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-slate-50"
            :class="[
              index === highlightedIndex ? 'bg-slate-100' : '',
              brandCitationTargetKey(target) === selectedKey ? 'bg-emerald-50/80 hover:bg-emerald-50' : '',
            ]"
            @mousedown.prevent="selectTarget(target)"
          >
            <span class="min-w-0 flex-1">
              <span class="font-medium text-gray-900">{{ target.label }}</span>
              <span
                v-if="target.kind === 'brand' && target.aliases.length > 0"
                class="mt-0.5 block truncate text-xs text-gray-500"
              >
                {{ target.aliases.join(', ') }}
              </span>
              <span
                v-else-if="target.kind === 'keyword'"
                class="mt-0.5 block text-xs text-gray-500"
              >
                Detected in responses
              </span>
            </span>
            <span
              v-if="target.kind === 'keyword'"
              class="shrink-0 rounded-full bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-500"
            >
              {{ target.count.toLocaleString() }}
            </span>
            <AppIcon
              v-if="brandCitationTargetKey(target) === selectedKey"
              name="check"
              size="sm"
              class="shrink-0 text-emerald-600"
            />
          </button>
        </li>
      </ul>

      <div
        v-else-if="showDropdown && isEditing && suggestions.length === 0"
        class="absolute z-10 mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-sm text-gray-500 shadow-lg"
      >
        No matches found
      </div>
    </div>
  </div>
</template>
