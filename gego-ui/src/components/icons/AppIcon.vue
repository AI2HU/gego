<script setup lang="ts">
import { computed } from 'vue'

export type AppIconName =
  | 'desktop'
  | 'refresh'
  | 'spinner'
  | 'menu'
  | 'close'
  | 'search'
  | 'file'
  | 'lightbulb'
  | 'server'
  | 'error'
  | 'github'
  | 'check'
  | 'chart-bar'
  | 'clock'
  | 'comment'
  | 'tag'
  | 'chart-line'
  | 'external-link'

const props = withDefaults(
  defineProps<{
    name: AppIconName
    size?: 'sm' | 'md' | 'lg'
    spin?: boolean
  }>(),
  {
    size: 'md',
    spin: false,
  },
)

const sizeClass = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'h-4 w-4'
    case 'lg':
      return 'h-6 w-6'
    default:
      return 'h-5 w-5'
  }
})
</script>

<template>
  <svg
    :class="[sizeClass, spin ? 'animate-spin' : '']"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.75"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'desktop'">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <path d="M8 21h8" />
      <path d="M12 17v4" />
    </template>

    <template v-else-if="name === 'refresh'">
      <path d="M21 12a9 9 0 1 1-2.64-6.36" />
      <path d="M21 3v6h-6" />
    </template>

    <template v-else-if="name === 'spinner'">
      <path d="M12 3a9 9 0 1 0 9 9" opacity="0.25" />
      <path d="M12 3a9 9 0 0 1 9 9" />
    </template>

    <template v-else-if="name === 'menu'">
      <path d="M4 7h16" />
      <path d="M4 12h16" />
      <path d="M4 17h16" />
    </template>

    <template v-else-if="name === 'close'">
      <path d="M18 6 6 18" />
      <path d="m6 6 12 12" />
    </template>

    <template v-else-if="name === 'search'">
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </template>

    <template v-else-if="name === 'file'">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
    </template>

    <template v-else-if="name === 'lightbulb'">
      <path d="M9 18h6" />
      <path d="M10 22h4" />
      <path d="M12 2a7 7 0 0 0-4 12.7V17h8v-2.3A7 7 0 0 0 12 2z" />
    </template>

    <template v-else-if="name === 'server'">
      <rect x="2" y="3" width="20" height="7" rx="2" />
      <rect x="2" y="14" width="20" height="7" rx="2" />
      <circle cx="7" cy="6.5" r="1" fill="currentColor" stroke="none" />
      <circle cx="7" cy="17.5" r="1" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'error'">
      <circle cx="12" cy="12" r="10" />
      <path d="m15 9-6 6" />
      <path d="m9 9 6 6" />
    </template>

    <template v-else-if="name === 'github'">
      <path
        d="M9 19c-4.3 1.4-4.3-2.5-6-3m12 5v-3.5c0-1 .4-1.9 1.1-2.5-3.6-.4-7.4-1.8-7.4-8.1 0-2 .7-3.6 2-4.9-.2-.4-.9-2.1.2-4.4 0 0 1.6-.5 5.2 1.9 1.5-.4 3.1-.6 4.7-.6s3.2.2 4.7.6c3.6-2.4 5.2-1.9 5.2-1.9 1.1 2.3.4 4 .2 4.4 1.3 1.3 2 2.9 2 4.9 0 6.3-3.8 7.7-7.4 8.1.9.8 1.4 1.9 1.4 3.2V21"
      />
    </template>

    <template v-else-if="name === 'check'">
      <circle cx="12" cy="12" r="10" />
      <path d="m9 12 2 2 4-4" />
    </template>

    <template v-else-if="name === 'chart-bar'">
      <path d="M4 20V10" />
      <path d="M10 20V4" />
      <path d="M16 20v-6" />
      <path d="M22 20V8" />
    </template>

    <template v-else-if="name === 'clock'">
      <circle cx="12" cy="12" r="10" />
      <path d="M12 6v6l4 2" />
    </template>

    <template v-else-if="name === 'comment'">
      <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" />
    </template>

    <template v-else-if="name === 'tag'">
      <path d="M20.59 13.41 10.83 3.65a2 2 0 0 0-2.83 0L3.65 8a2 2 0 0 0 0 2.83l9.76 9.76a2 2 0 0 0 2.83 0l4.35-4.35a2 2 0 0 0 0-2.83z" />
      <circle cx="7.5" cy="7.5" r="1" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'chart-line'">
      <path d="M3 3v18h18" />
      <path d="m7 15 4-4 3 3 5-6" />
    </template>

    <template v-else-if="name === 'external-link'">
      <path d="M15 3h6v6" />
      <path d="M10 14 21 3" />
      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
    </template>
  </svg>
</template>
