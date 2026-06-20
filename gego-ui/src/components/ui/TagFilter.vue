<script setup lang="ts">
import AppButton from '@/components/ui/AppButton.vue'

defineProps<{
  tags: string[]
  selectedTags: string[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  toggle: [tag: string]
  clear: []
}>()
</script>

<template>
  <div v-if="tags.length > 0" class="space-y-2">
    <div class="flex flex-wrap items-center gap-2">
      <span class="text-xs font-medium uppercase tracking-wider text-gray-400">
        Filter by tag
      </span>
      <button
        v-for="tag in tags"
        :key="tag"
        type="button"
        class="inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium transition-colors"
        :class="
          selectedTags.includes(tag)
            ? 'border-slate-600 bg-slate-600 text-white'
            : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50'
        "
        :disabled="disabled"
        @click="emit('toggle', tag)"
      >
        {{ tag }}
      </button>
    </div>

    <div
      v-if="selectedTags.length > 0"
      class="flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600"
    >
      <p>
        Filtering by
        <span class="font-semibold text-gray-900">{{ selectedTags.join(', ') }}</span>
      </p>
      <AppButton variant="ghost" size="sm" @click="emit('clear')">
        Clear tags
      </AppButton>
    </div>
  </div>
</template>
