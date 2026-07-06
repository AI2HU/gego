<script setup lang="ts">
import { ref, watch } from 'vue'

import AppIcon from '@/components/icons/AppIcon.vue'
import AppButton from '@/components/ui/AppButton.vue'
import { input } from '@/design/classes'
import type { PromptResponse } from '@/types/prompt'

const props = defineProps<{
  prompts: PromptResponse[]
  deletingId?: string | null
  savingTagsId?: string | null
}>()

const emit = defineEmits<{
  delete: [id: string]
  updateTags: [id: string, tags: string[]]
}>()

const editingTagsId = ref<string | null>(null)
const tagsDraft = ref('')
const confirmingDeleteId = ref<string | null>(null)

function startEditTags(prompt: PromptResponse) {
  editingTagsId.value = prompt.id
  tagsDraft.value = (prompt.tags ?? []).join(', ')
}

function cancelEditTags() {
  editingTagsId.value = null
  tagsDraft.value = ''
}

function saveTags(id: string) {
  const tags = tagsDraft.value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
  emit('updateTags', id, tags)
}

function parseTagsDraft(): string[] {
  return tagsDraft.value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
}

function requestDelete(id: string) {
  confirmingDeleteId.value = id
}

function cancelDelete() {
  confirmingDeleteId.value = null
}

function confirmDelete(id: string) {
  emit('delete', id)
  confirmingDeleteId.value = null
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

watch(
  () => props.savingTagsId,
  (current, previous) => {
    if (previous && current === null && editingTagsId.value === previous) {
      cancelEditTags()
    }
  },
)
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-gray-200/60 bg-white/80 backdrop-blur-sm shadow-sm">
    <div class="overflow-x-auto">
      <table class="min-w-full text-sm">
        <thead class="border-b border-gray-200/60 bg-slate-50/80">
          <tr class="text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
            <th class="px-4 py-3">Prompt</th>
            <th class="w-[min(280px,30%)] px-4 py-3">Tags</th>
            <th class="w-28 px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>

        <tbody class="divide-y divide-gray-100">
          <template v-for="prompt in prompts" :key="prompt.id">
            <tr class="group transition-colors hover:bg-slate-50/70">
              <td class="px-4 py-2.5 align-top min-w-[280px]">
                <p class="whitespace-pre-wrap text-gray-900 leading-snug">
                  {{ prompt.template }}
                </p>
                <p class="mt-1 text-xs text-gray-400">
                  Updated {{ formatDate(prompt.updated_at) }}
                </p>
              </td>

              <td class="px-4 py-2.5 align-top">
                <div v-if="editingTagsId === prompt.id" class="space-y-2">
                  <input
                    v-model="tagsDraft"
                    type="text"
                    placeholder="tag-one, tag-two"
                    :class="input.base"
                    class="!py-2 text-sm"
                    @keyup.enter="saveTags(prompt.id)"
                  />
                  <div class="flex flex-wrap gap-1.5">
                    <span
                      v-for="tag in parseTagsDraft()"
                      :key="tag"
                      class="inline-flex rounded-full bg-violet-50 px-2 py-0.5 text-xs font-medium text-violet-700"
                    >
                      {{ tag }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2">
                    <AppButton
                      size="sm"
                      :loading="savingTagsId === prompt.id"
                      @click="saveTags(prompt.id)"
                    >
                      Save
                    </AppButton>
                    <AppButton
                      size="sm"
                      variant="ghost"
                      :disabled="savingTagsId === prompt.id"
                      @click="cancelEditTags"
                    >
                      Cancel
                    </AppButton>
                  </div>
                </div>

                <div v-else class="flex items-start gap-2">
                  <div class="flex min-w-0 flex-1 flex-wrap gap-1">
                    <span
                      v-for="tag in prompt.tags ?? []"
                      :key="tag"
                      class="inline-flex rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700"
                    >
                      {{ tag }}
                    </span>
                    <span v-if="!(prompt.tags?.length)" class="text-xs text-gray-400 italic">
                      No tags
                    </span>
                  </div>
                  <button
                    type="button"
                    class="shrink-0 rounded-md p-1 text-gray-400 opacity-0 transition-opacity hover:bg-gray-100 hover:text-gray-700 group-hover:opacity-100"
                    title="Edit tags"
                    @click="startEditTags(prompt)"
                  >
                    <AppIcon name="tag" size="sm" />
                  </button>
                </div>
              </td>

              <td class="px-4 py-2.5 align-top text-right">
                <div v-if="confirmingDeleteId === prompt.id" class="flex items-center justify-end gap-2">
                  <span class="text-xs text-red-600">Remove?</span>
                  <AppButton
                    variant="danger"
                    size="sm"
                    :loading="deletingId === prompt.id"
                    @click="confirmDelete(prompt.id)"
                  >
                    Yes
                  </AppButton>
                  <AppButton
                    variant="ghost"
                    size="sm"
                    :disabled="deletingId === prompt.id"
                    @click="cancelDelete"
                  >
                    No
                  </AppButton>
                </div>

                <AppButton
                  v-else
                  variant="ghost"
                  size="sm"
                  :loading="deletingId === prompt.id"
                  class="!text-red-600 hover:!bg-red-50 opacity-0 group-hover:opacity-100 transition-opacity"
                  @click="requestDelete(prompt.id)"
                >
                  <AppIcon name="trash" size="sm" />
                </AppButton>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>
  </div>
</template>
