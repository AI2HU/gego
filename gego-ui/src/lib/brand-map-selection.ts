import { reactive } from 'vue'

export type SelectionAnchor = {
  top: number
  left: number
  bottom: number
  right: number
  width: number
  height: number
}

export type BrandMapSelectionState = {
  show: boolean
  selectedText: string
  alias: string
  brandName: string
  caseSensitive: boolean
  anchorRect: SelectionAnchor | null
  openScrollX: number
  openScrollY: number
  savedMessage: string | null
  savedMessageContainer: HTMLElement | null
  activeContainer: HTMLElement | null
}

const POPOVER_WIDTH = 320
const POPOVER_HEIGHT_ESTIMATE = 280

const containers = new Set<HTMLElement>()

const state = reactive<BrandMapSelectionState>({
  show: false,
  selectedText: '',
  alias: '',
  brandName: '',
  caseSensitive: false,
  anchorRect: null,
  openScrollX: 0,
  openScrollY: 0,
  savedMessage: null,
  savedMessageContainer: null,
  activeContainer: null,
})

let listenersAttached = false

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

function getAdjustedAnchor(): SelectionAnchor | null {
  if (!state.anchorRect) {
    return null
  }

  const dx = window.scrollX - state.openScrollX
  const dy = window.scrollY - state.openScrollY

  return {
    top: state.anchorRect.top - dy,
    left: state.anchorRect.left - dx,
    bottom: state.anchorRect.bottom - dy,
    right: state.anchorRect.right - dx,
    width: state.anchorRect.width,
    height: state.anchorRect.height,
  }
}

export function getBrandMapPopoverPosition(): { top: string; left: string } {
  const rect = getAdjustedAnchor()
  if (!rect) {
    return { top: '0px', left: '0px' }
  }

  const margin = 8
  const centerLeft = rect.left + rect.width / 2 - POPOVER_WIDTH / 2
  let left = clamp(centerLeft, margin, window.innerWidth - POPOVER_WIDTH - margin)

  let top = rect.bottom + margin
  if (top + POPOVER_HEIGHT_ESTIMATE > window.innerHeight - margin) {
    top = rect.top - POPOVER_HEIGHT_ESTIMATE - margin
  }
  top = clamp(top, margin, window.innerHeight - POPOVER_HEIGHT_ESTIMATE - margin)

  return {
    top: `${top}px`,
    left: `${left}px`,
  }
}

function storeAnchorFromSelection() {
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) {
    return
  }

  const range = selection.getRangeAt(0)
  const rect = range.getBoundingClientRect()
  state.anchorRect = {
    top: rect.top,
    left: rect.left,
    bottom: rect.bottom,
    right: rect.right,
    width: rect.width,
    height: rect.height,
  }
  state.openScrollX = window.scrollX
  state.openScrollY = window.scrollY
}

function updatePopoverPosition() {
  if (!state.show) {
    return
  }
  storeAnchorFromSelection()
}

function findContainerForSelection(): HTMLElement | null {
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0) {
    return null
  }

  const node = selection.getRangeAt(0).commonAncestorContainer
  const element = node instanceof Element ? node : node.parentElement
  if (!element) {
    return null
  }

  for (const container of containers) {
    if (container.contains(element)) {
      return container
    }
  }

  return null
}

export function hideBrandMapPopover() {
  state.show = false
  state.selectedText = ''
  state.alias = ''
  state.brandName = ''
  state.caseSensitive = false
  state.anchorRect = null
  state.activeContainer = null
}

function handleMouseUp() {
  const container = findContainerForSelection()
  if (!container) {
    return
  }

  state.activeContainer = container

  const selection = window.getSelection()
  if (!selection || selection.isCollapsed) {
    return
  }

  const text = selection.toString().trim()
  if (text.length < 2) {
    hideBrandMapPopover()
    return
  }

  state.selectedText = text
  state.alias = text
  state.brandName = ''
  storeAnchorFromSelection()
  state.show = true
}

function handleDocumentMouseDown(event: MouseEvent) {
  if (!state.show) {
    return
  }

  const target = event.target
  if (!(target instanceof Node)) {
    return
  }

  const popover = document.getElementById('brand-map-selection-popover')
  if (popover?.contains(target)) {
    return
  }

  if (state.activeContainer?.contains(target)) {
    return
  }

  hideBrandMapPopover()
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape' && state.show) {
    hideBrandMapPopover()
    window.getSelection()?.removeAllRanges()
  }
}

function attachGlobalListeners() {
  if (listenersAttached) {
    return
  }
  document.addEventListener('mouseup', handleMouseUp)
  document.addEventListener('mousedown', handleDocumentMouseDown)
  document.addEventListener('keydown', handleKeyDown)
  listenersAttached = true
}

function detachGlobalListeners() {
  if (containers.size > 0) {
    return
  }
  document.removeEventListener('mouseup', handleMouseUp)
  document.removeEventListener('mousedown', handleDocumentMouseDown)
  document.removeEventListener('keydown', handleKeyDown)
  listenersAttached = false
}

export function registerBrandMapContainer(element: HTMLElement) {
  containers.add(element)
  attachGlobalListeners()
}

export function unregisterBrandMapContainer(element: HTMLElement) {
  containers.delete(element)
  if (state.activeContainer === element) {
    hideBrandMapPopover()
  }
  detachGlobalListeners()
}

export function setBrandMapSavedMessage(message: string | null, container: HTMLElement | null = null) {
  state.savedMessage = message
  state.savedMessageContainer = container
}

export function getBrandMapSelectionState() {
  return state
}
