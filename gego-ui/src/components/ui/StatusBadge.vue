<script setup lang="ts">
import { computed } from 'vue'

import { status } from '@/design/classes'

const props = defineProps<{
  connected: boolean
  label?: string
  compact?: boolean
}>()

const text = computed(() => {
  if (props.label) return props.label
  return props.connected ? 'Connected' : 'Disconnected'
})
</script>

<template>
  <div :class="connected ? status.connected : status.disconnected">
    <div :class="connected ? status.connectedDot : status.disconnectedDot" />
    <span :class="[connected ? status.connectedText : status.disconnectedText, compact ? 'text-xs' : '']">
      {{ text }}
    </span>
  </div>
</template>
