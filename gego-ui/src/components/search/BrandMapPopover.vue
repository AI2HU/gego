<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import AppButton from '@/components/ui/AppButton.vue'
import BrandAliasInput from '@/components/brands/BrandAliasInput.vue'
import BrandNameCombobox from '@/components/brands/BrandNameCombobox.vue'
import {
  getBrandMapPopoverPosition,
  getBrandMapSelectionState,
  hideBrandMapPopover,
  setBrandMapSavedMessage,
} from '@/lib/brand-map-selection'
import { useMapBrandMutation } from '@/queries/brands'
import { useAuthStore } from '@/stores/auth'
import type { Brand } from '@/types/brand'

const authStore = useAuthStore()
const mapBrandMutation = useMapBrandMutation()

const canMapBrands = computed(() => authStore.hasPermission('words'))

const state = getBrandMapSelectionState()
const popoverStyle = ref(getBrandMapPopoverPosition())

function refreshPosition() {
  popoverStyle.value = getBrandMapPopoverPosition()
}

let positionFrame = 0
function schedulePositionRefresh() {
  cancelAnimationFrame(positionFrame)
  positionFrame = requestAnimationFrame(() => {
    refreshPosition()
  })
}

onMounted(() => {
  window.addEventListener('scroll', schedulePositionRefresh, true)
  window.addEventListener('resize', schedulePositionRefresh)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(positionFrame)
  window.removeEventListener('scroll', schedulePositionRefresh, true)
  window.removeEventListener('resize', schedulePositionRefresh)
})

const showPopover = computed(() => state.show && canMapBrands.value)

function handlePickBrandFromAlias(brand: Brand) {
  state.brandName = brand.name
}

async function handleSaveMapping() {
  const alias = state.alias.trim()
  const name = state.brandName.trim()
  if (!alias || !name) {
    return
  }

  const container = state.activeContainer

  await mapBrandMutation.mutateAsync({
    alias,
    name,
    case_sensitive: state.caseSensitive,
  })

  setBrandMapSavedMessage(`Mapped "${alias}" → "${name}"`, container)
  window.getSelection()?.removeAllRanges()
  hideBrandMapPopover()
}

watch(
  () => state.show,
  (visible) => {
    if (visible) {
      schedulePositionRefresh()
    }
  },
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="showPopover"
      id="brand-map-selection-popover"
      class="fixed z-50 w-80 rounded-xl border border-gray-200 bg-white p-4 shadow-xl"
      :style="popoverStyle"
      role="dialog"
      aria-label="Map selection to brand"
      @mousedown.stop
    >
      <p class="text-xs font-semibold uppercase tracking-wider text-gray-500">Map to brand</p>
      <p class="mt-2 text-sm text-gray-700">
        Selected:
        <span class="rounded bg-amber-100 px-1 py-0.5 font-medium text-amber-900">{{ state.selectedText }}</span>
      </p>

      <form class="mt-3 space-y-3" @submit.prevent="handleSaveMapping">
        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600">Alias</label>
          <BrandAliasInput
            v-model="state.alias"
            placeholder="Variant to map, e.g. Créateurs"
            @pick-brand="handlePickBrandFromAlias"
          />
          <p class="mt-1 text-xs text-gray-500">Type to match existing brand names.</p>
        </div>

        <div>
          <label class="mb-1 block text-xs font-medium text-gray-600">Brand name</label>
          <BrandNameCombobox
            v-model="state.brandName"
            placeholder="Canonical brand name"
          />
        </div>

        <label class="inline-flex items-center gap-2 text-sm text-gray-700">
          <input
            v-model="state.caseSensitive"
            type="checkbox"
            class="rounded border-gray-300"
          />
          Case-sensitive matching
        </label>

        <div class="flex justify-end gap-2">
          <AppButton type="button" variant="secondary" size="sm" @click="hideBrandMapPopover">
            Cancel
          </AppButton>
          <AppButton
            type="submit"
            size="sm"
            :disabled="!state.alias.trim() || !state.brandName.trim() || mapBrandMutation.isPending.value"
          >
            Save mapping
          </AppButton>
        </div>
      </form>
    </div>
  </Teleport>
</template>
