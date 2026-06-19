<script setup lang="ts">
import { RouterView } from 'vue-router'

import { useSessionRefresh } from '@/composables/useSessionRefresh'

useSessionRefresh()
</script>

<template>
  <RouterView />
</template>
