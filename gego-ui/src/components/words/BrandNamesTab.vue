<script setup lang="ts">
import { computed, ref } from 'vue'

import BrandAliasInput from '@/components/brands/BrandAliasInput.vue'
import BrandNameCombobox from '@/components/brands/BrandNameCombobox.vue'
import DetectedBrandWordInput from '@/components/brands/DetectedBrandWordInput.vue'
import AppIcon from '@/components/icons/AppIcon.vue'
import AppAlert from '@/components/ui/AppAlert.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import TagFilter from '@/components/ui/TagFilter.vue'
import { useCreateExclusionWordMutation } from '@/queries/exclusion-words'
import {
  useBrandsQuery,
  useCreateBrandAliasMutation,
  useCreateBrandMutation,
  useDeleteBrandAliasMutation,
  useDeleteBrandMutation,
  useMapBrandMutation,
  useSuggestedBrandWordsQuery,
  useUpdateBrandMutation,
} from '@/queries/brands'
import { usePromptsQuery } from '@/queries/prompts'
import type { Brand } from '@/types/brand'

const newBrandName = ref('')
const newBrandAlias = ref('')
const newBrandCaseSensitive = ref(false)
const brandSearchQuery = ref('')
const selectedTags = ref<string[]>([])
const deletingBrandId = ref<string | null>(null)
const deletingAliasKey = ref<string | null>(null)
const editingBrandId = ref<string | null>(null)
const editingBrandName = ref('')
const showMapDialog = ref(false)
const mapAlias = ref('')
const mapCanonicalName = ref('')
const mapCaseSensitive = ref(false)
const excludingWord = ref<string | null>(null)
const addingAliasBrandId = ref<string | null>(null)
const newAliasText = ref('')
const newAliasCaseSensitive = ref(false)

const brandsQuery = useBrandsQuery()
const promptsQuery = usePromptsQuery()
const suggestionsQuery = useSuggestedBrandWordsQuery(50, selectedTags)
const createBrandMutation = useCreateBrandMutation()
const updateBrandMutation = useUpdateBrandMutation()
const deleteBrandMutation = useDeleteBrandMutation()
const createAliasMutation = useCreateBrandAliasMutation()
const deleteAliasMutation = useDeleteBrandAliasMutation()
const mapBrandMutation = useMapBrandMutation()
const excludeMutation = useCreateExclusionWordMutation()

const brands = computed(() => brandsQuery.data.value ?? [])
const suggestions = computed(() => suggestionsQuery.data.value ?? [])

const knownAliasSet = computed(() => {
  const set = new Set<string>()
  for (const brand of brands.value) {
    set.add(brand.name.toLowerCase())
    for (const alias of brand.aliases) {
      set.add(alias.alias.toLowerCase())
    }
  }
  return set
})

const filteredBrands = computed(() => {
  const query = brandSearchQuery.value.trim().toLowerCase()
  if (!query) return brands.value
  return brands.value.filter((brand) => {
    if (brand.name.toLowerCase().includes(query)) return true
    return brand.aliases.some((alias) => alias.alias.toLowerCase().includes(query))
  })
})

const allTags = computed(() => {
  const tags = new Set<string>()
  for (const prompt of promptsQuery.data.value ?? []) {
    for (const tag of prompt.tags ?? []) {
      tags.add(tag)
    }
  }
  return Array.from(tags).sort((a, b) => a.localeCompare(b))
})

const hasActiveTagFilters = computed(() => selectedTags.value.length > 0)

const visibleSuggestions = computed(() =>
  suggestions.value.filter((item) => !knownAliasSet.value.has(item.word.toLowerCase())),
)

const errorMessage = computed(() => {
  const error = brandsQuery.error.value
  if (!error) return null
  return error instanceof Error ? error.message : 'Failed to load brands'
})

const isInitialLoading = computed(() => brandsQuery.isPending.value && !brandsQuery.data.value)

const matchingBrand = computed(() => {
  const name = newBrandName.value.trim().toLowerCase()
  if (!name) return null
  return brands.value.find((brand) => brand.name.toLowerCase() === name) ?? null
})

const isSubmittingBrand = computed(
  () => createBrandMutation.isPending.value || createAliasMutation.isPending.value,
)

const canSubmitBrand = computed(() => {
  if (!newBrandName.value.trim() || isSubmittingBrand.value) {
    return false
  }
  if (matchingBrand.value && !newBrandAlias.value.trim()) {
    return false
  }
  return true
})

const aliasChipClass =
  'inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 pl-2.5 pr-1 py-1 text-xs font-medium text-slate-700'

const suggestionChipClass =
  'inline-flex items-center gap-1.5 rounded-full border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-700'

function toggleTag(tag: string) {
  const index = selectedTags.value.indexOf(tag)
  if (index === -1) {
    selectedTags.value = [...selectedTags.value, tag]
    return
  }
  selectedTags.value = selectedTags.value.filter((value) => value !== tag)
}

function clearTagFilters() {
  selectedTags.value = []
}

function startEditBrand(brand: Brand) {
  editingBrandId.value = brand.id
  editingBrandName.value = brand.name
}

function cancelEditBrand() {
  editingBrandId.value = null
  editingBrandName.value = ''
}

async function saveEditBrand(id: string) {
  const name = editingBrandName.value.trim()
  if (!name) return
  await updateBrandMutation.mutateAsync({ id, payload: { name } })
  cancelEditBrand()
}

async function toggleBrandTarget(brand: Brand) {
  await updateBrandMutation.mutateAsync({
    id: brand.id,
    payload: { name: brand.name, is_target: !brand.is_target },
  })
}

async function handleCreateBrand() {
  const name = newBrandName.value.trim()
  if (!name) return

  const alias = newBrandAlias.value.trim()
  const existingBrand = brands.value.find(
    (brand) => brand.name.toLowerCase() === name.toLowerCase(),
  )

  if (existingBrand) {
    if (!alias) return
    await createAliasMutation.mutateAsync({
      brandId: existingBrand.id,
      payload: { alias, case_sensitive: newBrandCaseSensitive.value },
    })
  } else {
    const aliases = []
    if (alias) {
      aliases.push({ alias, case_sensitive: newBrandCaseSensitive.value })
    }
    await createBrandMutation.mutateAsync({ name, aliases })
  }

  newBrandName.value = ''
  newBrandAlias.value = ''
  newBrandCaseSensitive.value = false
  void suggestionsQuery.refetch()
}

async function handleDeleteBrand(id: string) {
  deletingBrandId.value = id
  try {
    await deleteBrandMutation.mutateAsync(id)
    void suggestionsQuery.refetch()
  } finally {
    deletingBrandId.value = null
  }
}

function startAddAlias(brandId: string) {
  addingAliasBrandId.value = brandId
  newAliasText.value = ''
  newAliasCaseSensitive.value = false
}

function cancelAddAlias() {
  addingAliasBrandId.value = null
  newAliasText.value = ''
  newAliasCaseSensitive.value = false
}

async function handleAddAlias(brandId: string) {
  const alias = newAliasText.value.trim()
  if (!alias) return
  await createAliasMutation.mutateAsync({
    brandId,
    payload: { alias, case_sensitive: newAliasCaseSensitive.value },
  })
  cancelAddAlias()
  void suggestionsQuery.refetch()
}

async function handleDeleteAlias(brandId: string, aliasId: string) {
  deletingAliasKey.value = `${brandId}:${aliasId}`
  try {
    await deleteAliasMutation.mutateAsync({ brandId, aliasId })
    void suggestionsQuery.refetch()
  } finally {
    deletingAliasKey.value = null
  }
}

function handlePickBrandFromAlias(brand: Brand) {
  mapCanonicalName.value = brand.name
}

function openMapDialog(word: string) {
  mapAlias.value = word
  mapCanonicalName.value = ''
  mapCaseSensitive.value = false
  showMapDialog.value = true
}

function closeMapDialog() {
  showMapDialog.value = false
  mapAlias.value = ''
  mapCanonicalName.value = ''
  mapCaseSensitive.value = false
}

async function handleMapBrand() {
  const alias = mapAlias.value.trim()
  const name = mapCanonicalName.value.trim()
  if (!alias || !name) return

  await mapBrandMutation.mutateAsync({
    alias,
    name,
    case_sensitive: mapCaseSensitive.value,
  })
  closeMapDialog()
  void suggestionsQuery.refetch()
}

async function handleExcludeSuggestion(word: string) {
  excludingWord.value = word
  try {
    await excludeMutation.mutateAsync({ word })
    void suggestionsQuery.refetch()
  } finally {
    excludingWord.value = null
  }
}
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-xl border border-gray-200/60 bg-white/60 p-6 backdrop-blur-sm">
      <div class="mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Add brand</h2>
        <p class="mt-1 text-sm text-gray-600">
          Define a canonical brand name and optionally add an alias for detected variants.
        </p>
      </div>

      <form class="space-y-3" @submit.prevent="handleCreateBrand">
        <div class="flex flex-col gap-3 sm:flex-row">
          <div class="flex-1">
            <BrandNameCombobox
              v-model="newBrandName"
              placeholder="Canonical brand name, e.g. Slip Français"
            />
          </div>
          <div class="flex-1">
            <DetectedBrandWordInput
              v-model="newBrandAlias"
              :suggestions="visibleSuggestions"
              placeholder="Optional alias, e.g. Slip Fran"
            />
            <p class="mt-1 text-xs text-gray-500">Type to match detected brand words.</p>
          </div>
        </div>
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <label class="inline-flex items-center gap-2 text-sm text-gray-700">
            <input
              v-model="newBrandCaseSensitive"
              type="checkbox"
              class="rounded border-gray-300"
            />
            Case-sensitive alias matching
          </label>
          <AppButton
            type="submit"
            :disabled="!canSubmitBrand"
          >
            {{ matchingBrand ? 'Add alias' : 'Add brand' }}
          </AppButton>
        </div>
      </form>
    </section>

    <section>
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">Brand names</h2>
          <p class="mt-1 text-sm text-gray-600">
            {{ brands.length }} brand{{ brands.length === 1 ? '' : 's' }} defined
          </p>
        </div>
        <AppInput
          v-model="brandSearchQuery"
          class="w-full sm:max-w-xs"
          placeholder="Search brands..."
        />
      </div>

      <AppAlert
        v-if="errorMessage"
        title="Unable to load brands"
        @retry="brandsQuery.refetch()"
      >
        {{ errorMessage }}
      </AppAlert>

      <LoadingState
        v-if="isInitialLoading"
        title="Loading brands"
        description="Fetching brand mappings..."
      />

      <EmptyState
        v-else-if="brands.length === 0"
        title="No brands yet"
        description="Add a brand name or map a detected word to create your first brand mapping."
        icon="tag"
      />

      <div v-else class="space-y-3">
        <div
          v-for="brand in filteredBrands"
          :key="brand.id"
          class="rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm sm:p-5"
        >
          <div class="mb-3 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0 flex-1">
              <div v-if="editingBrandId === brand.id" class="flex flex-col gap-2 sm:flex-row">
                <AppInput v-model="editingBrandName" class="flex-1" />
                <div class="flex gap-2">
                  <AppButton size="sm" @click="saveEditBrand(brand.id)">Save</AppButton>
                  <AppButton size="sm" variant="secondary" @click="cancelEditBrand">Cancel</AppButton>
                </div>
              </div>
              <div v-else class="flex items-center gap-2">
                <h3 class="text-base font-semibold text-gray-900">{{ brand.name }}</h3>
                <button
                  type="button"
                  class="inline-flex rounded-md p-1 transition-colors"
                  :class="
                    brand.is_target
                      ? 'text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700'
                      : 'text-slate-400 hover:bg-slate-100 hover:text-slate-600'
                  "
                  :title="brand.is_target ? 'Unset as target brand' : 'Set as target brand'"
                  :aria-pressed="brand.is_target"
                  :disabled="updateBrandMutation.isPending.value"
                  @click="toggleBrandTarget(brand)"
                >
                  <AppIcon name="target" size="sm" />
                </button>
                <button
                  type="button"
                  class="text-xs text-slate-500 hover:text-slate-700"
                  @click="startEditBrand(brand)"
                >
                  Edit
                </button>
              </div>
            </div>
            <AppButton
              variant="secondary"
              size="sm"
              :disabled="deletingBrandId === brand.id || deleteBrandMutation.isPending.value"
              @click="handleDeleteBrand(brand.id)"
            >
              Delete
            </AppButton>
          </div>

          <div class="flex flex-wrap gap-2">
            <span
              v-for="alias in brand.aliases"
              :key="alias.id"
              :class="aliasChipClass"
            >
              {{ alias.alias }}
              <span
                v-if="alias.case_sensitive"
                class="rounded bg-amber-100 px-1 text-[10px] font-semibold uppercase text-amber-700"
              >
                Aa
              </span>
              <button
                type="button"
                class="inline-flex rounded-full p-0.5 text-slate-400 hover:bg-slate-200 hover:text-slate-700 disabled:opacity-50"
                :disabled="deletingAliasKey === `${brand.id}:${alias.id}`"
                :title="`Remove alias ${alias.alias}`"
                @click="handleDeleteAlias(brand.id, alias.id)"
              >
                <AppIcon name="close" size="sm" />
              </button>
            </span>

            <button
              v-if="addingAliasBrandId !== brand.id"
              type="button"
              class="rounded-full border border-dashed border-slate-300 px-2.5 py-1 text-xs font-medium text-slate-500 hover:border-slate-400 hover:text-slate-700"
              @click="startAddAlias(brand.id)"
            >
              + Add alias
            </button>
          </div>

          <form
            v-if="addingAliasBrandId === brand.id"
            class="mt-3 flex flex-col gap-2 border-t border-gray-100 pt-3 sm:flex-row sm:items-center"
            @submit.prevent="handleAddAlias(brand.id)"
          >
            <div class="flex-1">
              <BrandAliasInput v-model="newAliasText" placeholder="Alias text" />
            </div>
            <label class="inline-flex items-center gap-2 text-sm text-gray-700">
              <input
                v-model="newAliasCaseSensitive"
                type="checkbox"
                class="rounded border-gray-300"
              />
              Case sensitive
            </label>
            <div class="flex gap-2">
              <AppButton size="sm" type="submit" :disabled="!newAliasText.trim()">Add</AppButton>
              <AppButton size="sm" variant="secondary" @click="cancelAddAlias">Cancel</AppButton>
            </div>
          </form>
        </div>
      </div>
    </section>

    <section>
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">Detected brand words</h2>
          <p class="mt-1 text-sm text-gray-600">
            Map incomplete detections to a canonical brand name.
            <span v-if="hasActiveTagFilters">Showing words from responses matching selected tags.</span>
          </p>
        </div>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="suggestionsQuery.isFetching.value"
          @click="suggestionsQuery.refetch()"
        >
          Refresh
        </AppButton>
      </div>

      <div
        v-if="allTags.length > 0"
        class="mb-4 rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm"
      >
        <TagFilter
          :tags="allTags"
          :selected-tags="selectedTags"
          :disabled="suggestionsQuery.isFetching.value"
          @toggle="toggleTag"
          @clear="clearTagFilters"
        />
      </div>

      <LoadingState
        v-if="suggestionsQuery.isPending.value && !suggestionsQuery.data.value"
        title="Loading suggestions"
        description="Scanning responses for detected brand words..."
      />

      <EmptyState
        v-else-if="visibleSuggestions.length === 0"
        title="No suggestions available"
        :description="
          hasActiveTagFilters
            ? 'No detected brand words found for the selected tags.'
            : 'Run prompts via the scheduler to collect responses, then map detected brand words here.'
        "
        icon="lightbulb"
      />

      <div
        v-else
        class="rounded-xl border border-gray-200/60 bg-white/60 p-4 backdrop-blur-sm sm:p-5"
      >
        <div class="flex flex-wrap gap-2">
          <div
            v-for="item in visibleSuggestions"
            :key="item.word"
            :class="suggestionChipClass"
          >
            <span>{{ item.word }}</span>
            <span class="rounded-full bg-slate-100 px-1.5 py-0.5 text-xs font-medium text-slate-500">
              {{ item.count }}
            </span>
            <AppButton
              size="sm"
              :disabled="mapBrandMutation.isPending.value"
              @click="openMapDialog(item.word)"
            >
              Map
            </AppButton>
            <AppButton
              size="sm"
              variant="secondary"
              :disabled="excludingWord === item.word || excludeMutation.isPending.value"
              @click="handleExcludeSuggestion(item.word)"
            >
              Exclude
            </AppButton>
          </div>
        </div>
      </div>
    </section>

    <div
      v-if="showMapDialog"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      @click.self="closeMapDialog"
    >
      <div class="w-full max-w-md rounded-xl border border-gray-200 bg-white p-6 shadow-xl">
        <h3 class="text-lg font-semibold text-gray-900">Map to brand</h3>
        <p class="mt-1 text-sm text-gray-600">
          Detected:
          <span class="rounded bg-amber-100 px-1 py-0.5 font-medium text-amber-900">{{ mapAlias }}</span>
        </p>

        <form class="mt-4 space-y-4" @submit.prevent="handleMapBrand">
          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600">Alias</label>
            <BrandAliasInput
              v-model="mapAlias"
              placeholder="Variant to map"
              @pick-brand="handlePickBrandFromAlias"
            />
            <p class="mt-1 text-xs text-gray-500">Type to match existing brand names.</p>
          </div>

          <div>
            <label class="mb-1 block text-xs font-medium text-gray-600">Brand name</label>
            <BrandNameCombobox
              v-model="mapCanonicalName"
              placeholder="Canonical brand name"
            />
          </div>

          <label class="inline-flex items-center gap-2 text-sm text-gray-700">
            <input
              v-model="mapCaseSensitive"
              type="checkbox"
              class="rounded border-gray-300"
            />
            Case-sensitive matching
          </label>

          <div class="flex justify-end gap-2">
            <AppButton type="button" variant="secondary" @click="closeMapDialog">Cancel</AppButton>
            <AppButton
              type="submit"
              :disabled="!mapAlias.trim() || !mapCanonicalName.trim() || mapBrandMutation.isPending.value"
            >
              Save mapping
            </AppButton>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
