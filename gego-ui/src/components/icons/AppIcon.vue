<script setup lang="ts">
import { computed } from 'vue'

import IconBase from '@/components/icons/IconBase.vue'
import IconChartBar from '@/components/icons/IconChartBar.vue'
import IconChartLine from '@/components/icons/IconChartLine.vue'
import IconCheck from '@/components/icons/IconCheck.vue'
import IconChevronDown from '@/components/icons/IconChevronDown.vue'
import IconClock from '@/components/icons/IconClock.vue'
import IconClose from '@/components/icons/IconClose.vue'
import IconComment from '@/components/icons/IconComment.vue'
import IconDesktop from '@/components/icons/IconDesktop.vue'
import IconError from '@/components/icons/IconError.vue'
import IconExternalLink from '@/components/icons/IconExternalLink.vue'
import IconFile from '@/components/icons/IconFile.vue'
import IconGithub from '@/components/icons/IconGithub.vue'
import IconGlobe from '@/components/icons/IconGlobe.vue'
import IconJournal from '@/components/icons/IconJournal.vue'
import IconLightbulb from '@/components/icons/IconLightbulb.vue'
import IconMenu from '@/components/icons/IconMenu.vue'
import IconPlus from '@/components/icons/IconPlus.vue'
import IconRefresh from '@/components/icons/IconRefresh.vue'
import IconSearch from '@/components/icons/IconSearch.vue'
import IconServer from '@/components/icons/IconServer.vue'
import IconSettings from '@/components/icons/IconSettings.vue'
import IconSpinner from '@/components/icons/IconSpinner.vue'
import IconTag from '@/components/icons/IconTag.vue'
import IconTarget from '@/components/icons/IconTarget.vue'
import IconTrash from '@/components/icons/IconTrash.vue'
import IconUsers from '@/components/icons/IconUsers.vue'

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
  | 'target'
  | 'chart-line'
  | 'external-link'
  | 'globe'
  | 'plus'
  | 'trash'
  | 'settings'
  | 'journal'
  | 'chevron-down'
  | 'users'

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

const isFilled = computed(() => props.name === 'chevron-down')
</script>

<template>
  <IconBase
    :icon-name="name"
    :size="size"
    :spin="spin"
    :filled="isFilled"
    :view-box="isFilled ? '0 0 20 20' : '0 0 24 24'"
  >
    <IconDesktop v-if="name === 'desktop'" />
    <IconRefresh v-else-if="name === 'refresh'" />
    <IconSpinner v-else-if="name === 'spinner'" />
    <IconMenu v-else-if="name === 'menu'" />
    <IconClose v-else-if="name === 'close'" />
    <IconSearch v-else-if="name === 'search'" />
    <IconFile v-else-if="name === 'file'" />
    <IconLightbulb v-else-if="name === 'lightbulb'" />
    <IconServer v-else-if="name === 'server'" />
    <IconError v-else-if="name === 'error'" />
    <IconGithub v-else-if="name === 'github'" />
    <IconCheck v-else-if="name === 'check'" />
    <IconChartBar v-else-if="name === 'chart-bar'" />
    <IconClock v-else-if="name === 'clock'" />
    <IconComment v-else-if="name === 'comment'" />
    <IconTag v-else-if="name === 'tag'" />
    <IconTarget v-else-if="name === 'target'" />
    <IconChartLine v-else-if="name === 'chart-line'" />
    <IconExternalLink v-else-if="name === 'external-link'" />
    <IconGlobe v-else-if="name === 'globe'" />
    <IconPlus v-else-if="name === 'plus'" />
    <IconTrash v-else-if="name === 'trash'" />
    <IconSettings v-else-if="name === 'settings'" />
    <IconJournal v-else-if="name === 'journal'" />
    <IconChevronDown v-else-if="name === 'chevron-down'" />
    <IconUsers v-else-if="name === 'users'" />
  </IconBase>
</template>
