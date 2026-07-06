import { ref, watch, type MaybeRefOrGetter, toValue } from 'vue'

export function useAutocompleteList(resetOn?: MaybeRefOrGetter<unknown>) {
  const isOpen = ref(false)
  const highlightedIndex = ref(0)

  if (resetOn !== undefined) {
    watch(
      () => toValue(resetOn),
      () => {
        highlightedIndex.value = 0
      },
    )
  }

  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
  }

  function handleKeyDown(event: KeyboardEvent, itemCount: number, pick: (index: number) => void) {
    if (!isOpen.value || itemCount === 0) {
      return
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault()
      highlightedIndex.value = (highlightedIndex.value + 1) % itemCount
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      highlightedIndex.value = (highlightedIndex.value - 1 + itemCount) % itemCount
      return
    }

    if (event.key === 'Enter') {
      event.preventDefault()
      pick(highlightedIndex.value)
      return
    }

    if (event.key === 'Escape') {
      close()
    }
  }

  return {
    isOpen,
    highlightedIndex,
    open,
    close,
    handleKeyDown,
  }
}
