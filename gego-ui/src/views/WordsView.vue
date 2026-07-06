<script setup lang="ts">
import { ref } from 'vue'

import BrandNamesTab from '@/components/words/BrandNamesTab.vue'
import WordsExclusionTab from '@/components/words/WordsExclusionTab.vue'

type WordsTab = 'exclusion' | 'brands'

const activeTab = ref<WordsTab>('exclusion')

const tabButtonClass = (tab: WordsTab) =>
  [
    'rounded-full border px-4 py-2 text-sm font-medium transition-colors',
    activeTab.value === tab
      ? 'border-slate-600 bg-slate-600 text-white'
      : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50',
  ].join(' ')
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap gap-2">
      <button type="button" :class="tabButtonClass('exclusion')" @click="activeTab = 'exclusion'">
        Words exclusion
      </button>
      <button type="button" :class="tabButtonClass('brands')" @click="activeTab = 'brands'">
        Brand names
      </button>
    </div>

    <WordsExclusionTab v-if="activeTab === 'exclusion'" />
    <BrandNamesTab v-else />
  </div>
</template>
