<script setup lang="ts">
import { computed } from 'vue'

import InlineMarkdown from '@/components/search/InlineMarkdown.vue'
import { parseMarkdownBlocks } from '@/lib/response-text'

const props = defineProps<{
  text: string
  keyword: string
  caseSensitive?: boolean
}>()

const blocks = computed(() =>
  parseMarkdownBlocks(props.text, props.keyword, props.caseSensitive),
)

function headingClass(level: number): string {
  if (level <= 2) {
    return 'text-base font-semibold text-gray-900'
  }
  if (level === 3) {
    return 'text-sm font-semibold text-gray-900'
  }
  return 'text-sm font-medium text-gray-800'
}
</script>

<template>
  <div class="space-y-2 break-words">
    <template v-for="(block, index) in blocks" :key="index">
      <div v-if="block.type === 'spacer'" class="h-1" />

      <component
        :is="block.level <= 2 ? 'h2' : block.level === 3 ? 'h3' : 'h4'"
        v-else-if="block.type === 'heading'"
        :class="[headingClass(block.level), index > 0 ? 'mt-3' : '']"
      >
        <InlineMarkdown :parts="block.parts" />
      </component>

      <div
        v-else-if="block.type === 'list-item'"
        class="flex gap-2 text-sm text-gray-800"
      >
        <span class="shrink-0 text-gray-400">•</span>
        <span class="min-w-0">
          <InlineMarkdown :parts="block.parts" />
        </span>
      </div>

      <p v-else class="text-sm text-gray-800">
        <InlineMarkdown :parts="block.parts" />
      </p>
    </template>
  </div>
</template>
