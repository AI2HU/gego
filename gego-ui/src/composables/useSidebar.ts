import { createSharedComposable, useMediaQuery, useScrollLock } from '@vueuse/core'
import { computed, ref, watch } from 'vue'

function _useSidebar() {
  const isOpen = ref(false)
  const isLgUp = useMediaQuery('(min-width: 1024px)')
  const scrollLock = useScrollLock(document.body)

  const isVisible = computed(() => isLgUp.value || isOpen.value)

  watch(isLgUp, (lg) => {
    if (lg) {
      isOpen.value = false
    }
  })

  watch(isOpen, (open) => {
    if (!isLgUp.value) {
      scrollLock.value = open
    }
  })

  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
  }

  function toggle() {
    isOpen.value = !isOpen.value
  }

  return {
    isOpen,
    isLgUp,
    isVisible,
    open,
    close,
    toggle,
  }
}

export const useSidebar = createSharedComposable(_useSidebar)
