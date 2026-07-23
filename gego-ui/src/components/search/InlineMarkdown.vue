<script setup lang="ts">
import InlineMarkdown from '@/components/search/InlineMarkdown.vue'
import type { InlinePart } from '@/lib/response-text'

defineProps<{
  parts: InlinePart[]
}>()
</script>

<template>
  <template v-for="(part, index) in parts" :key="index">
    <span
      v-if="part.type === 'link'"
      class="inline-flex items-baseline gap-0.5"
    >
      <a
        :href="part.url"
        target="_blank"
        rel="noopener noreferrer"
        class="text-slate-700 underline decoration-slate-300 underline-offset-2 hover:text-slate-900 hover:decoration-slate-500"
      >
        <InlineMarkdown :parts="part.parts" />
      </a>
      <span
        class="inline-flex h-3.5 w-3.5 shrink-0 translate-y-[-0.5px] cursor-help items-center justify-center rounded-full bg-slate-200 text-[9px] font-bold leading-none text-slate-600"
        :title="part.url"
        aria-label="Link URL"
      >?</span>
    </span>

    <strong v-else-if="part.type === 'bold'" class="font-semibold text-gray-900">
      <InlineMarkdown :parts="part.parts" />
    </strong>

    <em v-else-if="part.type === 'italic'" class="italic">
      <InlineMarkdown :parts="part.parts" />
    </em>

    <template v-else>
      <template v-for="(segment, segmentIndex) in part.segments" :key="segmentIndex">
        <mark
          v-if="segment.highlight === 'target'"
          class="rounded bg-emerald-100 px-0.5 font-medium text-emerald-900"
          title="Target brand"
        >{{ segment.text }}</mark>
        <mark
          v-else-if="segment.highlight === 'keyword'"
          class="rounded bg-amber-100 px-0.5 font-medium text-amber-900"
          title="Search term"
        >{{ segment.text }}</mark>
        <mark
          v-else-if="segment.highlight === 'alias'"
          class="rounded bg-sky-100 px-0.5 font-medium text-sky-900"
          title="Brand alias match"
        >{{ segment.text }}</mark>
        <span v-else>{{ segment.text }}</span>
      </template>
    </template>
  </template>
</template>
