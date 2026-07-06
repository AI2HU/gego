<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

import {
  getBrandMapSelectionState,
  registerBrandMapContainer,
  unregisterBrandMapContainer,
} from '@/lib/brand-map-selection'

const containerRef = ref<HTMLElement | null>(null)
const savedMessage = ref<string | null>(null)

onMounted(() => {
  if (containerRef.value) {
    registerBrandMapContainer(containerRef.value)
  }
})

onBeforeUnmount(() => {
  if (containerRef.value) {
    unregisterBrandMapContainer(containerRef.value)
  }
})

watch(
  () => getBrandMapSelectionState().savedMessage,
  (message) => {
    const selectionState = getBrandMapSelectionState()
    if (!message || !containerRef.value) {
      return
    }
    if (selectionState.savedMessageContainer === containerRef.value) {
      savedMessage.value = message
      selectionState.savedMessage = null
      selectionState.savedMessageContainer = null
    }
  },
)
</script>

<template>
  <div>
    <div ref="containerRef" class="select-text">
      <slot />
    </div>

    <p v-if="savedMessage" class="mt-2 text-xs font-medium text-emerald-700">
      {{ savedMessage }}
    </p>
  </div>
</template>
